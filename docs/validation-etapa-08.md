# Etapa 08 — implantação inicial da homologação

Data: 13/10/2026. Estado: **homologação online, validação inicial aprovada; aceite completo pendente**. Produção não implantada. Backup externo ainda pendente. Criação do administrador confirmada pelo responsável em 13/10/2026.

## Revisão implantada e CI

Commit fixado no Coolify: `66db5902c15927611d436fa0a630b491a1ad28eb`, branch main. Migrações esperadas 1–6.

- PR [12](https://github.com/franklinlem/certguard/pull/12): Compose específico e política de isolamento. Validação 38414754028 e Segurança 38414753372 aprovadas após correção da representação numérica de memória no JSON do Compose.
- PR [13](https://github.com/franklinlem/certguard/pull/13): labels compatíveis com Raw e build sem segredo. Validação 38415849733 e Segurança 38415849694 aprovadas.
- PR [14](https://github.com/franklinlem/certguard/pull/14): identidade de projeto consistente entre build/runtime. Validação 38436924948 e Segurança 38436925044 aprovadas.
- Conjunto final em main: [Validação 38437775184](https://github.com/franklinlem/certguard/actions/runs/38437775184) e [Segurança 38437775141](https://github.com/franklinlem/certguard/actions/runs/38437775141), ambas aprovadas.

CI verifica sintaxe e política do Compose tanto com nome local de homologação quanto com UUID simulado do Coolify. Confere redes por serviço, ausência de portas do host/binds/privilégios, limite de memória, volume exclusivo, flags desabilitadas e labels do web. Permanecem testes completos do produto, PostgreSQL, concorrência, fuzzing, builds, auditoria, E2E desktop/mobile e restauração descartável. A primeira falha da nova checagem não foi apresentada como aprovação.

## Configuração no Coolify

Projeto CertGuard (`e9tmgu7akme29lrctms1ktoe`), ambiente `homologacao` (`6lujazqxlsryjw0301a2qerd`), recurso `certguard-dev` (`rk48swz6mdqduoyyhfnqhylp`), applicationId 9, servidor localhost/Contabo.

- Integração GitHub já existente `frank-coolify-gh-app`, sem ampliação de permissões.
- Base `/`, arquivo `/compose.coolify.yaml`, modo Raw, deploy automático desligado.
- Comandos de build/start nativos, sem comando customizado. Nome Compose `${COOLIFY_RESOURCE_UUID:-certguard-dev}` mantém a mesma identidade nas duas fases.
- Somente web na rede externa existente `coolify`, com labels para `certguard-dev.franklem.uk`, entrypoints http/https e resolver letsencrypt; redes app/database próprias, database interna.
- APP_ENV=production e origem HTTPS fixa. Scheduler/retenção/entregas falsos; métricas e chave de canais vazias.
- Senha exclusiva gerada pelo Coolify em SERVICE_PASSWORD_64_DB, protegida e indisponível durante build. Nenhum valor de credencial foi registrado nos documentos.
- Não foi criado recurso de produção, nem alterada configuração global de Traefik, firewall, DNS ou serviços anteriores.

No modo Raw, o painel pode mostrar “No public domain configured”, pois a rota é definida nos labels versionados do Compose. A evidência externa abaixo comprova o roteamento. Não acrescentar uma segunda rota de produção para resolver essa indicação visual.

## Tentativas e resultado da inicialização

1. `v5jlzkp70qihaughldp1xp7s`, 11:41: falha antes da criação de containers por mistura de labels em mapa/lista. Corrigida no PR 13.
2. `jqkkolln8ndzfll1p3keh6o9`, 11:49: build aprovado; início buscou imagens com prefixo diferente e tentou reconstruir fora do checkout.
3. `bz55ywqf40elvypqkkium8oi`, 14:10: tentativa de comando customizado não encontrou .env do host dentro do helper. Comando removido; solução versionada no PR 14.
4. [Deploy final ol0vcpnrd4uskkgdkggzgwbo](https://coolify.franklem.uk/project/e9tmgu7akme29lrctms1ktoe/environment/6lujazqxlsryjw0301a2qerd/application/rk48swz6mdqduoyyhfnqhylp/deployment/ol0vcpnrd4uskkgdkggzgwbo): redes e volume novos criados às 14:21:46; banco saudável às 14:21:57; migração encerrou e liberou API; API e web iniciados às 14:21:58–59. Log confirma “New container started”.

Recursos criados: redes `rk48swz6mdqduoyyhfnqhylp_app` e `_database`, volume `rk48swz6mdqduoyyhfnqhylp_postgres_data`, containers com sufixos db-1, migrate-1, api-1 e web-1. Migrate é tarefa de execução única; seu encerramento após sucesso é esperado.

## Verificações externas comprovadas

Executadas às **14:26:53 -03:00**, com verificação normal de certificado/cadeia/hostname, sem bypass TLS. Resolução normal local já funcionava; uma consulta anterior também usou o IPv4 confirmado com SNI correto.

| Verificação | Resultado |
|---|---|
| HTTPS / | 200, título CertGuard — Monitoramento TLS |
| TLS | TLS 1.3, Let's Encrypt YR2, SAN certguard-dev.franklem.uk |
| Validade do certificado | Até 11/01/2027 16:23:52 UTC |
| HTTP / | 301 para HTTPS do mesmo domínio |
| /api/health | 200, status ok |
| /api/ready | 200, status ready: banco acessível e schema 1–6 reconhecido |
| /api/v1/auth/me | 401 sem sessão |
| /api/v1/endpoints | 401 sem sessão |
| /api/v1/dashboard | 401 sem sessão |
| /metrics e /api/metrics | 404 pelo web |
| Indexação | X-Robots-Tag noindex, nofollow |
| Cache | HTML no-cache, API no-store, assets com hash max-age=31536000 |
| Assets JS/CSS | 200, tipos corretos |

## Evidências internas e limites

Antes da implantação, às 11:37:11: 4850 MiB disponíveis, 78G livres, 25 containers anteriores. Assinatura SHA-256 da lista ordenada de IDs/nomes desses containers: `6cc9a131cdb81e3964da7f2f1c6b623c2ba6a78d1cd2e41ba2843cd35a7cdbb6`.

Inspeção filtrada do arquivo persistido da primeira tentativa confirmou: db/migrate somente database, API database/app, web app/proxy; nenhum env_file acrescentado aos serviços. Essa é evidência de configuração, não substitui inspeção Docker após a subida.

O controle do navegador desconectou após a implantação. A inspeção final de saúde, portas e redes dos containers, comparação da assinatura dos serviços existentes e captura visual **não foram concluídas**. Acesso SSH local por chave continua rejeitado; não houve extração de credenciais ou alteração de autenticação para contornar isso. Verificações externas e logs acima permanecem confirmados. Não afirmar aceite completo ou ausência de impacto medida sem concluir essa comparação.

## Pendências de aceite

1. Recuperar sessão de controle do Coolify e inspecionar redes/portas/limites/saúde efetivos. Comparar os serviços anteriores e registrar recursos após a subida.
2. Administrador `franklinlem@gmail.com` criado pelo responsável, que confirmou sucesso após corrigir o tamanho da senha. Nenhuma senha foi recebida ou registrada. Validar o fluxo autenticado completo em homologação.
3. Validar login/logout, cookie/CSRF e fluxos autenticados com destino de teste autorizado. Não sondar destinos internos reais para testar SSRF.
4. Backup escolhido: `franklin@frank-note2`, servidor físico Linux Mint XFCE com Docker/Hermes/outros backups/Uptime Kuma, segundo o responsável. Acesso, diretório, criptografia, retenção, conectividade e restore serão verificados depois, conforme pedido. Nenhum acesso ou alteração nesse servidor foi feito.
5. Não ativar scheduler/retention/delivery nem usar dados reais até concluir os respectivos gates. Produção continua bloqueada até aceite documentado da homologação.

Rollback: somente recurso CertGuard, preservando volume e sem downgrade automático de schema. Não usar versões anteriores incompatíveis com schema 6. Não executar prune/down -v/DROP ou alterações de recursos compartilhados.
