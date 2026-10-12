# Validação da etapa 03

Esta etapa depende da etapa 02 (PR #6) e não modifica a VPS.

## Ambiente local

- Go: testes com detector de corrida e vet aprovados. Testes PostgreSQL são ignorados localmente pela ausência de `TEST_DATABASE_URL`; a validação real será feita na CI.
- Frontend: tipos, lint, três testes Vitest e build aprovados.
- Casos novos: configuração, concorrência limitada, cancelamento, backoff, retenção periódica, estado/agenda no painel; integração SQL cobre locks, reagendamento, migração de dados existentes e retenção por lote/idade.

## CI

Aguardando execução da branch. Docker/PostgreSQL/Playwright não foram executados localmente. Agendador e retenção ficam desabilitados no E2E para não acessar destinos externos; testes SQL usam inspeção controlada.

## Limites

Nenhum deploy ou alteração na VPS. A indicação do painel representa configuração, não saúde do executor. Limites de workers são por instância; operar inicialmente uma réplica. Retenção é opt-in e preserva o último registro mesmo acima da idade configurada. Alertas e dashboard agregado ficam para o próximo incremento.
