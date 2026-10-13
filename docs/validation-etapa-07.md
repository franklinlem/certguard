# Validação da etapa 07

Base: PR #10. Revisão dos PRs registrada em revisao-prs-etapa-07.md, sem merges. Nenhuma alteração na VPS ou mensagem real.

## Local

- Go race, vet e staticcheck aprovados.
- Métricas: auth/fail-closed, no-store, erros genéricos, formato e cardinalidade fixa. SQL aguarda CI.
- Frontend: tipos/lint/quatorze testes/build aprovados. Script de backup passou na análise sintática; não foi executado localmente. Não há Docker/PostgreSQL locais.

## CI

Aguardando PostgreSQL, smoke concorrente, benchmark de 1000 endpoints/3000 checks, dump/restore descartável, E2E desktop/mobile e segurança. Registrar execuções e números reais aqui após conclusão.

## Limites

Benchmark de leituras não mede capacidade de escrita, TLS externo, SMTP ou VPS. Contadores globais por banco e sem histórico anterior já removido. Homologação real não executada; roteiro depende de inventário/acesso e autorização para alterações com risco. Retenção de eventos/entregas e rotação assistida de chave ainda pendentes.
