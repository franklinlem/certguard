# ADR-001 — Incremento inicial
Status: aceito. Data: 12/10/2026.

O pedido prioriza fundação e motor TLS. Entregamos esse núcleo antes de autenticação, CRUD e scheduler; não declaramos v0.1/v0.2 completas. React é apenas uma base após testes do motor. Net/http reduz dependências; pgx/v5 faz acesso ao PostgreSQL.

Migração inicial SQL embutida em Go, com transação, versão e advisory lock. Foi escolhido um runner mínimo para uma migração; antes de ampliar o schema, avaliar goose e migrar o histórico sem executar novamente a versão 1. Não existe rollback destrutivo automático.

Metadados de certificados ficam previstos em JSONB no histórico, preservando cadeia e SANs; campos consultáveis (status, endpoint, data, IP) são colunas. Normalização dos demais campos será uma decisão medida conforme consultas reais.

Licença Apache-2.0 conforme proposta do planejamento. Repositório privado; licença não implica publicação pública.
