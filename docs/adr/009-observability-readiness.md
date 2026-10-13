# ADR-009 — observabilidade durável e readiness de schema

Status: proposto na etapa 07, dependente do ADR-008.

Exportador `/metrics` somente na API, habilitado por token separado das sessões administrativas. Token vazio desabilita com 404; 32–256 caracteres ASCII sem espaços. Bearer comparado por SHA-256/constant time. Sem query token, cookies ou acesso pelo proxy web; erro de coleta retorna 503 genérico, consulta com timeout de 3s e no-store. Não instalar Prometheus ou expor portas automaticamente.

Contadores representam operações persistidas no banco, não conexões externas físicas: checks, falhas (code != success), histograma cumulativo de duração e tentativas/aceites/falhas de entrega. Cada incremento pertence à transação do trabalho, após evento/fila e demais locks; rollback não incrementa. Não usar contagem direta de histórico como counter, pois retenção a reduziria. Migração 006 semeia a partir do histórico ainda existente; remoções anteriores não são reconstruíveis.

Histograma usa duração de checks em segundos, buckets 0.1/0.5/1/2.5/5/10/30/+Inf. Gauges mostram último resultado, vencimentos, agenda e fila em snapshot de leitura. Não descrevem saúde em tempo real. Etiquetas somente le e status enumerados; 30 séries fixas, sem hostname, URL, IDs, e-mail ou códigos remotos.

Uma linha de contadores duráveis simplifica transações no volume inicial; representa ponto de serialização na persistência. Benchmark e smoke avaliam leituras, sem afirmar capacidade de escrita da VPS. Reavaliar sharding/agregação em maior escala. Métricas são globais por banco/ambiente: coletar uma instância por instalação para evitar duplicar counters ao somar réplicas do mesmo banco.

Readiness verifica Ping e conjunto exato de versões do schema (1–6). Liveness não depende de schema/banco. Um binário novo não sinaliza pronto antes de migrar; schema de versão futura exige revisão de compatibilidade. Rollback não apaga migrações automaticamente, podendo exigir binário compatível. Índices aditivos suportam consulta do último destino e histórico global de entregas.

CI adiciona benchmark de leituras, smoke concorrente, testes de privacidade/auth/rollback/retenção, E2E em desktop/mobile e dump/restore num banco novo do projeto efêmero. O ensaio não valida infraestrutura Contabo, provedor SMTP ou restauração de chave real. Nenhum recurso real é criado ou modificado nesta etapa.

Referências: [formato Prometheus](https://prometheus.io/docs/instrumenting/exposition_formats/), [nomes/cardinalidade](https://prometheus.io/docs/practices/naming/), [pg_dump](https://www.postgresql.org/docs/17/app-pgdump.html), [pg_restore](https://www.postgresql.org/docs/17/app-pgrestore.html).
