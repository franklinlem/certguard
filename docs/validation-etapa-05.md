# Validação da etapa 05

Base: etapa 04 / PR #8, ainda aberto. Primeira entrega de alertas: motor, eventos e interface; sem transporte externo.

## Local

- Go: testes com detector de corrida, vet e staticcheck aprovados. Integração PostgreSQL ignorada localmente por ausência de TEST_DATABASE_URL; será confirmada na CI.
- Frontend: tipos, lint, onze testes e build aprovados.
- Casos novos: regras/limiares, fingerprint e deduplicação; lifecycle de eventos, rollback atômico, retenção e filtros; interface com observação preservada, estados vazio/erro, atualização e resolução administrativa.
- Revisão React: componentes separados, efeitos com descarte de respostas após desmontagem, dependências explícitas, conteúdo renderizado como texto, controles por labels e teclado. Nenhuma nova dependência de UI.

## CI

Aguardando PostgreSQL, Docker, Playwright e segurança.

## Limites

Nenhum deploy, merge ou alteração na VPS. Nenhuma mensagem externa enviada. SMTP, webhook, canais, outbox/retries e retenção de eventos ficam para entregas seguintes, conforme plano da etapa. PRs #6–#8 continuam dependências abertas.
