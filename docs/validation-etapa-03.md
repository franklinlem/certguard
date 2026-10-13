# Validação da etapa 03

Esta etapa depende da etapa 02 (PR #6) e não modifica a VPS.

## Ambiente local

- Go: testes com detector de corrida, vet e staticcheck aprovados. Testes PostgreSQL são ignorados localmente pela ausência de `TEST_DATABASE_URL`; a integração real foi aprovada na CI.
- Frontend: tipos, lint, três testes Vitest e build aprovados.
- Casos novos: configuração, concorrência limitada, cancelamento, backoff, retenção periódica, estado/agenda no painel; integração SQL cobre locks, reagendamento, migração de dados existentes e retenção por lote/idade.

## CI

Código validado: `a31393ae2ed4d961644d36ed3562b2544ef5d11f`.

- [Validação #38296862965](https://github.com/franklinlem/certguard/actions/runs/38296862965): backend, frontend e Docker aprovados, incluindo PostgreSQL 17, migração de instalação existente, testes com detector de corrida, cobertura core, fuzzing e Playwright.
- [Segurança #38296862985](https://github.com/franklinlem/certguard/actions/runs/38296862985): dependências, código e imagens aprovados.
- [PR #7](https://github.com/franklinlem/certguard/pull/7) aberto sobre a branch da etapa 02. Não houve merge.

Docker/PostgreSQL/Playwright não foram executados localmente. A atualização posterior deste documento não altera o código validado. Agendador e retenção ficam desabilitados no E2E para não acessar destinos externos; testes SQL usam inspeção controlada.

## Limites

Nenhum deploy ou alteração na VPS. A indicação do painel representa configuração, não saúde do executor. Limites de workers são por instância; operar inicialmente uma réplica. Retenção é opt-in e preserva o último registro mesmo acima da idade configurada. Alertas e dashboard agregado ficam para o próximo incremento.
