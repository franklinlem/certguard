# Validação da etapa 05

Base: etapa 04 / PR #8, ainda aberto. Primeira entrega de alertas: motor, eventos e interface; sem transporte externo.

## Local

- Go: testes com detector de corrida, vet e staticcheck aprovados. Integração PostgreSQL ignorada localmente por ausência de TEST_DATABASE_URL; aprovada na CI.
- Frontend: tipos, lint, onze testes e build aprovados.
- Casos novos: regras/limiares, fingerprint e deduplicação; lifecycle de eventos, rollback atômico, retenção e filtros; interface com observação preservada, estados vazio/erro, atualização e resolução administrativa.
- Revisão React: componentes separados, efeitos com descarte de respostas após desmontagem, dependências explícitas, conteúdo renderizado como texto, controles por labels e teclado. Nenhuma nova dependência de UI.

## CI

Código validado: `20a53c14df7ed064b3ca2854b19b611646ae1d32`.

- [Validação #38396100910](https://github.com/franklinlem/certguard/actions/runs/38396100910): PostgreSQL/race/vet/staticcheck, cobertura/fuzzing, frontend, Docker e Playwright aprovados.
- [Segurança #38396100906](https://github.com/franklinlem/certguard/actions/runs/38396100906): dependências, código e imagens aprovados.
- [PR #9](https://github.com/franklinlem/certguard/pull/9), dependente do PR #8, aberto e sem merge.

A primeira execução detectou associação imprecisa do rótulo do seletor de eventos no navegador. Corrigida com htmlFor/id e validada novamente no Chromium local (dados simulados, mobile sem overflow) e no E2E integrado da CI. A versão validada acima inclui a correção. O commit posterior apenas registra estas evidências.

Os testes integrados não realizaram inspeções TLS públicas nem enviaram mensagens externas. A verificação do PR #8 em 13/10/2026 confirmou estado aberto, sem conflitos, revisões ou comentários de alteração pendentes.

## Limites

Nenhum deploy, merge ou alteração na VPS. Nenhuma mensagem externa enviada. SMTP, webhook, canais, outbox/retries e retenção de eventos ficam para entregas seguintes, conforme plano da etapa. PRs #6–#8 continuam dependências abertas.
