# Histórico

## Integração e inventário — 13/10/2026
- PRs 1–11 integrados em main; dependências testadas sobre o produto completo.
- Inventário somente leitura da Contabo/Coolify, DNS e pendências de HTTPS/backup registrados.
- Sem implantação; homologação e produção ainda não validadas.

## Etapa 07 — integrada em 13/10/2026
- Métricas bearer opt-in com contadores/histograma duráveis e gauges sem identificadores.
- Readiness de schema e proxy restrito para rotas operacionais.
- Benchmark/smoke PostgreSQL, E2E desktop/mobile e ensaio dump/restore descartável.
- Revisão de PRs e roteiro de homologação; sem merge ou alteração da VPS.

## Etapa 06 — integrada em 13/10/2026
- Canais SMTP/webhook cifrados, pausados no cadastro e ativação explícita.
- Outbox transacional, retries limitados, cancelamento e histórico de estado.
- HTTPS/HMAC, SSRF, SMTP TLS obrigatório e testes locais isolados.
- Executor desabilitado por padrão; sem mensagens reais ou deploy.

## Etapa 05 — integrada em 13/10/2026
- Regras de alerta, deduplicação por episódio e recuperação confirmada.
- Eventos transacionais com snapshot preservado após retenção dos checks.
- Filtros, indicadores e histórico no painel autenticado.
- Sem SMTP/webhook, notificações externas ou deploy.

## Etapa 04 — integrada em 13/10/2026
- Dashboard autenticado com contadores globais e último resultado por destino.
- Detalhes de certificado/TLS e gráfico temporal com tabela acessível.
- Estados de ausência, carregamento e falha; atualização explícita.
- Testes SQL, frontend e E2E; sem deploy.

## Etapa 03 — integrada em 13/10/2026
- Agendamento opt-in com workers limitados, jitter e locks compartilhados com checks manuais.
- Próxima execução persistida e exibida no painel.
- Retenção opt-in em lotes, preservando o último resultado por endpoint.
- Testes de concorrência, cancelamento, reprogramação e limpeza; sem deploy.

## Etapa 02 — integrada em 13/10/2026
- Administrador inicial por CLI, login/logout, sessão persistida e CSRF.
- Cadastro, edição, pausa e arquivamento de endpoints.
- Check manual com lock/cooldown e histórico transacional.
- Interface real e testes de integração/E2E; sem deploy.

## Fundação — 12/10/2026
- Fundação do monorepositório e motor TLS com proteção SSRF.
- Testes de certificados, DNS/rebinding, timeout, migração e saúde.
- Base web, Docker Compose, CI e decisões técnicas.
- Nenhum deploy; autenticação e fluxo completo de monitoramento pendentes.
