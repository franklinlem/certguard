# Desenvolvimento

Go 1.27.2 e Node 22.23.2+; PostgreSQL 17 via Docker. Dependências fixadas em go.sum e package-lock.json. Use `npm ci` no frontend.

`make test`, `make lint` e `make build`. Teste rápido do core: `cd api && go test -race -cover ./internal/security ./internal/certificate`.

Fuzzing: `go test ./internal/security -fuzz=FuzzNormalize -fuzztime=10s`. Certificados de teste são gerados em memória e servidores escutam somente loopback; nenhuma chave privada é salva no repositório.

Migrações: `DATABASE_URL=... go run ./cmd/certguard migrate`. A URL deve vir de secret/env local; não cole senhas em histórico compartilhado. O teste PostgreSQL cria schema temporário e remove somente seu schema; execute em banco descartável.

Use branches curtas e PRs com evidências. `main` representa a base revisada. Atualize ADRs para decisões de segurança. A CI exige cobertura de pelo menos 80% nos dois pacotes core.

Frontend: `npm run dev`, `npm test`, `npm run lint`, `npm run typecheck`, `npm run build`. Não há dashboard funcional nesta etapa. Implementar estado real junto à API autenticada.
