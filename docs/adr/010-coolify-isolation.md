# ADR-010 — Compose explícito para homologação no Coolify

Data: 13/10/2026. Estado: aceito para configuração inicial, sujeito à validação de implantação.

## Contexto

O Compose gerado pelo modo gerenciado do Coolify 4.3.23 incluiu db, migrate e API na rede do recurso acessível ao Traefik. A exigência do projeto é conectar somente o frontend ao proxy. O inventário confirmou rede externa `coolify`, entrypoints `http`/`https` e resolver `letsencrypt` existentes.

## Decisão

Usar `compose.coolify.yaml` exclusivo da homologação, na raiz, com modo Raw (deploy file as-is). Banco/migração somente na rede interna database; API em database/app; frontend em app/proxy. Somente web recebe labels Traefik, sem publicar portas do host ou alterar configuração global. Contextos de build relativos à raiz eliminam dependência de realocação do arquivo pelo painel.

Domínio, flags desabilitadas e ambiente seguro ficam fixos nesse arquivo para impedir uso acidental em produção. Senha do banco é gerada pelo mecanismo SERVICE_PASSWORD_64_DB do Coolify e fornecida exclusivamente pelo ambiente do recurso, sem trânsito por logs ou Git. Volume e redes privadas usam o projeto Compose exclusivo selecionado pelo Coolify. Nenhum acesso ao servidor de backup é necessário nesta etapa.

## Validação e consequências

CI renderiza Compose e verifica redes por serviço, ausência de portas/binds/privilégios, isolamento do volume, limites, flags e labels. Builds e E2E do produto continuam na stack efêmera existente. A primeira implantação deve comprovar também o isolamento efetivo com inspeções Docker filtradas. Configuração do proxy passa a ser responsabilidade deste arquivo; alterações futuras exigem conferir o inventário novamente.

Referência: [documentação oficial de Compose/Raw do Coolify](https://coolify.io/docs/applications/builds/docker-compose). O modo Raw preserva a definição e exige labels e redes completos. O Compose local permanece em deploy/compose.yaml.
