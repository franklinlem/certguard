# Etapa 03 — agendamento e retenção

## Objetivo e dependência

Acrescentar verificações automáticas ao cadastro autenticado da etapa 02. Esta entrega parte da branch `feature/auth-endpoints-checks` (PR #6 ainda aberto), com PR separado e dependente. Não implanta nem modifica a VPS.

## Implementação planejada

1. Migração aditiva para inicializar a próxima execução dos endpoints ativos com dispersão de até 30 segundos e otimizar seleção/retenção.
2. Agendador integrado ao processo da API, habilitado explicitamente por `SCHEDULER_ENABLED=true`, com pool PostgreSQL próprio, 1–25 workers (padrão 10), timeout TLS de 1–30 segundos (padrão 10s), espera cancelável e dispersão na partida.
3. Cada worker seleciona um endpoint vencido com `FOR UPDATE SKIP LOCKED`. O lock é mantido até concluir a inspeção e persistir resultado e próxima execução na mesma transação. O check manual usa o mesmo lock e também reagenda. Sem fila ilimitada ou nova infraestrutura.
4. Agenda a partir da conclusão, intervalo configurado mais jitter de 0–10%. Não tenta recuperar todas as execuções perdidas. Falhas TLS/SSRF são resultados persistidos e seguem a mesma cadência; falhas do banco aguardam antes de tentar novamente.
5. Pausar/arquivar remove a próxima execução. Reativar ou alterar destino/intervalo agenda com dispersão inicial. Alterações de nome/limiares preservam a agenda. Cancelamento reverte a transação para permitir nova tentativa.
6. Retenção opt-in independente (`RETENTION_ENABLED=false` por padrão), janela padrão de 180 dias (mínimo 30), lote máximo configurável de 1–1000, um lote por minuto. Preserva sempre o último resultado de cada endpoint, inclusive arquivados. Exclusões são definitivas e exigem revisão da janela e backup antes de habilitar em ambiente real.
7. Interface informa configuração do agendador e próxima execução prevista, mantendo atualização explícita da lista. Não promete alertas, que pertencem ao próximo incremento. API autenticada acrescenta configuração operacional à sessão e `next_check_at` ao endpoint.

## Validação e critérios de aceite

- Testes offline do motor continuam sem alteração de proteção SSRF.
- Testes PostgreSQL: disputa entre workers e check manual, seleção somente devidos/ativos, reagendamento, pausa/retomada, rollback/cancelamento e retenção por idade/lote preservando último resultado.
- Testes do executor: limite de concorrência, configuração inválida e encerramento sem goroutines pendentes.
- Frontend: configuração e próxima execução; lint, tipos, testes e build. E2E com scheduler e retenção desabilitados para não acessar destinos externos.
- CI com PostgreSQL, imagens, E2E e análise de vulnerabilidades. Registrar resultados reais, sem apresentar testes ignorados localmente como aprovados.

## Limites e referências

Locks transacionais implicam uma conexão por check em andamento. Uma interrupção após o handshake e antes do commit pode repetir a inspeção; não há garantia de exatamente uma conexão externa por ciclo. Limite de workers é por instância; inicialmente uma réplica. Não haverá alertas, métricas distribuídas ou deploy nesta etapa.

Referências: [PostgreSQL — SELECT e SKIP LOCKED](https://www.postgresql.org/docs/17/sql-select.html), [Google Search Central](https://developers.google.com/search/docs?hl=pt-br), [cache de recursos](https://developer.chrome.com/docs/performance/insights/cache?hl=pt-br), [estabilidade de layout](https://web.dev/articles/optimize-cls?hl=pt-br). Manter painel autenticado sem indexação, API sem cache e assets versionados com cache; não adicionar imagens ou elementos que causem deslocamento de layout.
