# Desenvolvimento

Go 1.27.2 e Node 22.23.2+; PostgreSQL 17 via Docker. Dependências fixadas em go.sum e package-lock.json. Use `npm ci` no frontend.

`make test`, `make lint` e `make build`. Teste rápido do core: `cd api && go test -race -cover ./internal/security ./internal/certificate`.

Fuzzing: `go test ./internal/security -fuzz=FuzzNormalize -fuzztime=10s`. Certificados de teste são gerados em memória e servidores escutam somente loopback; nenhuma chave privada é salva no repositório.

Migrações: `DATABASE_URL=... go run ./cmd/certguard migrate`. A URL deve vir de secret/env local; não cole senhas em histórico compartilhado. O teste PostgreSQL cria schema temporário e remove somente seu schema; execute em banco descartável.

Use branches curtas e PRs com evidências. `main` representa a base revisada. Atualize ADRs para decisões de segurança. A CI exige cobertura de pelo menos 80% nos dois pacotes core.

Frontend: `npm run dev`, `npm test`, `npm run lint`, `npm run typecheck`, `npm run build`. Não há dashboard funcional nesta etapa. Implementar estado real junto à API autenticada.

## Etapa 02
Bootstrap: `certguard admin-create email` recebe senha por stdin. Rode primeiro `certguard migrate`. PUBLIC_APP_URL é obrigatória ao iniciar a API; APP_ENV=development permite HTTP somente em localhost/127.0.0.1.

Para desenvolvimento Vite, use PUBLIC_APP_URL=http://localhost:5173 (ou a porta real do Vite) e reinicie somente a API local. O proxy Vite aponta a 127.0.0.1:8081; em Compose, mantenha http://localhost:8088.

Integração HTTP/PostgreSQL injeta resultado do motor para evitar conexões externas. Testes do core continuam usando servidores TLS locais reais. A CI também executa Playwright contra Compose: login, cadastro, edição, pausa, histórico vazio, arquivamento e logout, sem verificar domínios externos.
