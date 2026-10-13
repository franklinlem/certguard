# ADR-005 — agendamento transacional e retenção opt-in

Status: proposto na etapa 03, dependente do ADR-004.

O volume inicial permite manter o lock de uma linha e uma conexão PostgreSQL durante cada inspeção TLS. Escolhemos workers no processo da API, com pool de banco próprio, sem broker ou serviço adicional. `FOR UPDATE SKIP LOCKED` distribui endpoints devidos; o check manual usa `NOWAIT` sobre a mesma linha. Persistência do resultado e próxima execução ocorrem no mesmo commit.

O banco calcula timestamps e jitter: até 30s na primeira agenda/retomada, intervalo mais 0–10% após cada conclusão. Cada worker também aguarda 0–30s na partida, e no mínimo 250ms entre trabalhos. Sem trabalho ou com erro de persistência, aguarda 5s. Timeout externo do trabalho = timeout TLS + 5s para banco; cancelamento reverte a transação. Não há recuperação de todas as execuções perdidas. Falhas TLS são resultados válidos persistidos.

O limite é por instância, com 10 workers por padrão e máximo 25; quatro checks manuais adicionais usam o pool da API. Operar inicialmente uma réplica e revisar o orçamento de conexões antes de escalar. Locks evitam simultaneidade por endpoint, mas uma queda após handshake e antes do commit pode causar repetição. Isso é aceitável para inspeção sem payload HTTP. A interface informa configuração e previsão, sem prometer execução exata ou saúde operacional.

A retenção é uma opção independente, desabilitada por padrão: 180 dias, lote de 500 por minuto, máximo 1000, operação com deadline de 10s. Exclui somente registros anteriores à janela e que tenham um resultado posterior do mesmo endpoint. Datas iguais usam UUID para desempate. Preserva o último registro inclusive de endpoints arquivados; portanto, não é uma garantia de apagar todos os dados acima da idade definida. `SKIP LOCKED` permite lotes concorrentes sem espera por linhas ocupadas. Não altera endpoints ou sessões.

Migração 003 inicializa agendas existentes e adiciona índices; não executa retenção. Habilitar remoção em ambiente real requer janela revisada, backup e autorização quando houver risco de perda de dados, conforme a solicitação do usuário. Nenhum ambiente real foi alterado nesta etapa.

Referência: [PostgreSQL 17 — locking clauses](https://www.postgresql.org/docs/17/sql-select.html#SQL-FOR-UPDATE-SHARE).
