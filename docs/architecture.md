# Arquitetura

Monorepositório: `api/` (Go), `web/` (React/TypeScript), `deploy/` (Compose) e `docs/`.

O motor independente recebe hostname e porta, normaliza, resolve A/AAAA uma única vez, valida todos os IPs e faz TCP ao primeiro IP aprovado. SNI preserva o hostname. TLS captura a cadeia; x509 verifica confiança e hostname separadamente. A CLI retorna resultado estruturado, sem persistência nesta etapa.

A API tem somente saúde e readiness. O banco possui schema inicial para endpoints e histórico; não existe caminho HTTP de monitoramento sem autenticação. `ready` testa comunicação com PostgreSQL; não certifica completude de migrações. O Compose garante migração antes da API. A página inicial identifica o estado de desenvolvimento sem métricas fictícias.

Próximas fronteiras: auth → endpoints → checks/storage → scheduler → alerts. Pacotes vazios não foram criados: cada componente será introduzido com código, testes e contrato no incremento correspondente.

A CI executa testes de integração do banco e smoke test Compose em ambiente efêmero. Deploy não faz parte da CI inicial.
