# Inventário para homologação — 13/10/2026

Inspeção somente leitura pelo painel autenticado [Coolify](https://coolify.franklem.uk/) e seu terminal. Nenhuma configuração salva, implantação, reinício, instalação, alteração de DNS/firewall ou remoção realizada. Consultas não incluíram ambientes dos containers, credenciais ou conteúdo de backups.

## Identificação e capacidade observada

Servidor `localhost` no Coolify, hostname `vmi3570746`; interface eth0 confirmou o IPv4 `158.220.120.2` informado pelo responsável. Acesso SSH direto com a chave local foi rejeitado; inventário executado pelo terminal existente do painel.

| Item | Evidência |
|---|---|
| Data da leitura de recursos | 2026-10-13 10:41:56 -03:00 |
| Sistema | Ubuntu 24.04.5 LTS, x86_64, kernel 6.8.0-139-generic |
| CPU | 4 núcleos; carga 0,82 / 0,88 / 0,83 |
| Memória | 7941 MiB totais; 2967 MiB usados; 4974 MiB disponíveis |
| Swap | 2047 MiB totais; 223 MiB usados |
| Disco de / e /data | Mesmo /dev/sda1: 96G totais, 19G usados, 78G disponíveis, 20% ocupado |
| Uptime efetivo | 21h55 na leitura; prevalece sobre o campo de início antigo exibido no resumo do painel |
| Coolify | 4.3.23 |
| Docker / Compose | 29.9.0 / 5.6.0, conforme resumo do servidor no painel |
| Proxy em execução | `coolify-proxy`, imagem `traefik:v3.7`, saudável |

Há margem aparente para o orçamento de execução do CertGuard (db 512 MiB, API 256 MiB, web 128 MiB), mas isso não mede pico de build/migração nem garante capacidade sob carga. A amostra de consumo dos containers mostrou Coolify em aproximadamente 362 MiB e proxy em 45 MiB. Não foi feito teste de carga na VPS.

## Serviços existentes a preservar

O painel lista quatro projetos: Serviço Fácil, Teste VPS, WebhookLab e omnisearch (dois ambientes). Docker apresentou 25 containers, todos em execução na consulta. A lista contém infraestrutura Coolify, serviços web/API/PostgreSQL/Redis de projetos existentes, aplicação de porta interna 3000 e Nginx. Nenhum nome de container continha `certguard`, inclusive na consulta de containers parados. Não há projeto CertGuard na listagem do painel.

Infraestrutura: `coolify`, `coolify-db` (PostgreSQL 15), `coolify-redis` (Redis 7), `coolify-realtime` 1.0.19, `coolify-proxy` e `coolify-sentinel`.

| Identificador do recurso | Grupos observados |
|---|---|
| `ldqo0wyojrlns1v56su8k2qh` | web, API, PostgreSQL 18.4 |
| `lg8czchlclxn7oxobrtcj1y6` | frontend, API, PostgreSQL 17, Redis |
| `vrbhd9vn1wuhtvjla5dyanpq` | frontend, API, PostgreSQL 17, Redis |
| `3d1rsbhzhmdmbcwk8e4gxehn` | frontend, backend, banco (omnisearch) |
| `xk1dq9xkgvuc6edif5swluxr` | frontend, backend, banco (omnisearch) |
| `4er3fhkwgi8ew34rlgga4hea` / `esgvf3zgq0vnxr3taou283qz` | aplicação e Nginx existentes |

Identificadores são apresentados para evitar reutilização acidental; a associação nominal dos três primeiros grupos aos projetos não foi inferida. A amostra de estado mostrou serviços em execução; alguns frontends/Nginx não têm indicação de healthcheck, o que não equivale a falha.

## Redes, portas e proxy

Redes bridge existentes: `bridge`, `coolify`, `ldqo0wyojrlns1v56su8k2qh`, `lg8czchlclxn7oxobrtcj1y6`, `vrbhd9vn1wuhtvjla5dyanpq`, `3d1rsbhzhmdmbcwk8e4gxehn`, `xk1dq9xkgvuc6edif5swluxr`; também `host` e `none`. O proxy participa de `coolify` e das cinco redes de recursos. Redes de aplicação usam faixas 10.0.1.0/24 a 10.0.6.0/24; a rede `coolify` também possui IPv6 privado.

Configuração efetiva observada do Traefik:

- entrypoints `http` em :80 e `https` em :443, HTTP/3 habilitado;
- provider Docker habilitado, `exposedbydefault=false`;
- provider de arquivos em `/traefik/dynamic/`, com observação de alterações;
- resolver `letsencrypt`, desafio HTTP no entrypoint `http`, armazenamento `/traefik/acme.json`;
- API insegura desabilitada.

Portas TCP em escuta no host: 22, 80, 443, 3000, 8000, 8080, 6001 e 6002, além de DNS local em 127.0.0.53/54. Isso descreve binds locais, não comprova exposição através do firewall do provedor. Docker publica 80/443/8080 do proxy, 8000 do Coolify e 6001–6002 do realtime em IPv4/IPv6; 443/UDP também está publicado. API e bancos existentes consultados mostram portas internas. A atribuição da escuta 3000 a um processo não foi aprofundada.

CertGuard deve usar a integração nativa do Coolify com esse proxy, criar rede e volume próprios e publicar somente web:8080 internamente. Não criar segundo proxy, disputar portas do host ou conectar API/banco a redes de outros projetos. O nome da futura rede depende do recurso a criar e ainda não existe.

UFW está ativo e lista permissões TCP 22/80/443 para qualquer origem em IPv4/IPv6, além de SSH para redes internas Docker. Não foram auditadas regras completas de Docker/iptables nem firewall externo da Contabo; UFW isoladamente não comprova bloqueio das portas publicadas pelo Docker. Nenhuma regra foi alterada.

Volumes existentes a preservar: `coolify-db`, `coolify-redis`, volumes `postgres-data` dos grupos ldqo/lg8/vrbh, `redis-data` dos grupos lg8/vrbh, e `hf-cache`, `pgdata`, `uploads-data` de cada ambiente omnisearch; há também um volume anônimo. Nenhum volume foi lido, removido ou reaproveitado.

## DNS, HTTPS e backups

Na primeira consulta de 13/10/2026, o resolvedor local não resolveu `certguard-dev.franklem.uk` nem `certguard.franklem.uk`. Consulta independente DNS-over-HTTPS do Google retornou NOERROR sem respostas A/AAAA/CNAME para ambos, com SOA Cloudflare (`keanu.ns.cloudflare.com`). HTTPS não pôde ser validado sem resolução. Isso não prova ausência de configuração em propagação ou de acesso privado.

Após criação pelo responsável, nova consulta DNS-over-HTTPS às **10:50:21 -03:00** confirmou registros A de ambos os domínios para **158.220.120.2**, TTL 300. A captura fornecida mostra ambos em **Somente DNS**, TTL automático. O resolvedor local ainda retornava ausência de resolução nessa leitura. A criação de DNS não é implantação nem autorização de produção. Homologação permanece em `certguard-dev.franklem.uk`.

Consulta HTTPS com resolução fixada exclusivamente para o IPv4 confirmado e SNI `certguard-dev.franklem.uk` falhou na validação por **certificado autoassinado**. Não foi usado bypass de validação. O certificado confiável/roteamento do domínio de homologação ainda precisa ser provisionado na futura configuração da aplicação; o registro DNS, sozinho, não estabelece HTTPS válido.

O painel S3 Storage informa **0 destinos de backup cadastrados** na equipe consultada. Não há evidência de política, cópia externa, retenção ou ensaio de restauração do CertGuard (a aplicação ainda não foi implantada). Ausência no painel não exclui backups manuais ou snapshots do provedor. O ensaio automatizado de dump/restore da CI usa banco descartável e não substitui backup da VPS.

O diretório `/data/coolify/backups` existe e contém 22 arquivos até profundidade 3. A amostra mais recente tem data 13/10/2026 03:16 e tamanhos 1270, 1531 e 2821 bytes. Foram consultados somente metadados: conteúdo, integridade, cobertura dos bancos e possibilidade de restauração **não foram validados**. Esses arquivos no mesmo servidor não comprovam cópia externa ou backup futuro do CertGuard.

## Pendências antes da implantação

1. Confirmar propagação do DNS de homologação e caminho do HTTPS/Cloudflare; definir proteção de acesso ao ambiente.
2. Definir backup específico do banco com destino externo, retenção, cópia separada da chave de canais e restauração em banco novo descartável.
3. Criar projeto/recurso isolado `certguard-dev` com auto-deploy desligado, commit explícito e migrações 1–6; revisar configuração gerada antes da primeira subida.
4. Manter scheduler, retenção e entregas desabilitados inicialmente; bootstrap administrativo e receptores/destinos de teste somente autorizados.
5. Medir fluxo real, recursos, readiness, cookies/CSRF, TLS, pause/retomada e rollback compatível com schema antes do aceite.

Inventário não é aceite de homologação nem implantação. Não há aprovação para alterar serviços compartilhados, firewall, proxy global ou apagar dados. Produção permanece bloqueada até validação documentada de homologação.
