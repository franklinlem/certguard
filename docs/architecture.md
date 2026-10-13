# Arquitetura — etapa 03

Monorepositório: api/ (Go), web/ (React/TypeScript), deploy/ (Compose) e docs/.

O motor independente normaliza hostname, resolve A/AAAA uma vez, valida todos os IPs e faz TCP ao primeiro IP aprovado. SNI preserva o hostname. TLS captura a cadeia e x509 verifica confiança/hostname separadamente. A CLI de inspeção continua disponível sem persistência.

A API usa net/http. internal/auth contém primitives de credenciais e sessão; internal/storage concentra SQL parametrizado; internal/httpapi aplica autenticação, CSRF e limites. Check manual chama o motor dentro de transação com lock de endpoint, persistindo resultado e timestamp atomicamente.

PostgreSQL mantém administrador, sessões (hash de token), endpoints e histórico JSONB. Migrações 001, 002 e 003 são versionadas e transacionais sob advisory lock. Readiness testa Ping do banco; não certifica o schema. Compose aplica migrações antes de iniciar API.

A interface React usa cookies same-origin e token CSRF em memória. Login, cadastro, edição, pausa, histórico, checks manuais e arquivamento estão implementados. Nginx serve assets e encaminha /api/v1 sem remover o prefixo.

A CI usa PostgreSQL e Compose efêmeros. Integração HTTP testa o fluxo com resultado de inspeção controlado; testes core usam servidores TLS locais. Playwright valida o fluxo administrativo sem conexões TLS externas. Próximos componentes: alertas e visão agregada. Não há deploy automático.

O agendador integra o ciclo de vida de `serve`, com pool PostgreSQL separado da API. Cada worker usa `SKIP LOCKED` e mantém a transação durante o check; resultado, último check e próxima execução são atômicos. Manual e automático compartilham o lock. A retenção tem uma goroutine independente, executa um lote por minuto e preserva o último resultado. Ambos exigem habilitação explícita. Veja ADR-005.
