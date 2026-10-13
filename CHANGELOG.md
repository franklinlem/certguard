# Histórico

## Etapa 06 — em revisão
- Canais SMTP/webhook cifrados, pausados no cadastro e ativação explícita.
- Outbox transacional, retries limitados, cancelamento e histórico de estado.
- HTTPS/HMAC, SSRF, SMTP TLS obrigatório e testes locais isolados.
- Executor desabilitado por padrão; sem mensagens reais ou deploy.

## Etapa 05 — em revisão
- Regras de alerta, deduplicação por episódio e recuperação confirmada.
- Eventos transacionais com snapshot preservado após retenção dos checks.
- Filtros, indicadores e histórico no painel autenticado.
- Sem SMTP/webhook, notificações externas ou deploy.

## Etapa 04 — em revisão
- Dashboard autenticado com contadores globais e último resultado por destino.
- Detalhes de certificado/TLS e gráfico temporal com tabela acessível.
- Estados de ausência, carregamento e falha; atualização explícita.
- Testes SQL, frontend e E2E; sem deploy.

## Etapa 03 — em revisão
- Agendamento opt-in com workers limitados, jitter e locks compartilhados com checks manuais.
- Próxima execução persistida e exibida no painel.
- Retenção opt-in em lotes, preservando o último resultado por endpoint.
- Testes de concorrência, cancelamento, reprogramação e limpeza; sem deploy.

## Etapa 02 — em revisão
- Administrador inicial por CLI, login/logout, sessão persistida e CSRF.
- Cadastro, edição, pausa e arquivamento de endpoints.
- Check manual com lock/cooldown e histórico transacional.
- Interface real e testes de integração/E2E; sem deploy.

## Fundação — 12/10/2026
- Fundação do monorepositório e motor TLS com proteção SSRF.
- Testes de certificados, DNS/rebinding, timeout, migração e saúde.
- Base web, Docker Compose, CI e decisões técnicas.
- Nenhum deploy; autenticação e fluxo completo de monitoramento pendentes.
