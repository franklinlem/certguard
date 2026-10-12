# CertGuard

Monitor self-hosted de certificados TLS, com validação de destinos e proteção SSRF.

**Estado: fundação e TLS Check Engine. Ainda não é o MVP e não está pronto para produção.**

## Entregue nesta etapa
- Go: inspeção TLS via CLI, validação de cadeia/hostname, SANs, fingerprint, validade e classificação.
- Bloqueio SSRF, DNS fixado no IP aprovado, timeout total e testes TLS isolados.
- API `/health` e `/ready`, PostgreSQL e migração transacional inicial.
- Base React + TypeScript em português, Docker Compose e GitHub Actions.
- Planejamento, decisões técnicas, modelo de ameaças e roteiro de homologação.

Autenticação, CRUD, persistência dos resultados do motor, scheduler, dashboard de monitoramento e alertas pertencem aos próximos incrementos. Não há endpoint público para disparar verificações.

## Desenvolvimento
Requisitos: Go 1.27.2, Node.js 22.23.2+ e Docker com Compose v2.

```sh
cp .env.example .env
# Defina POSTGRES_PASSWORD com senha aleatória alfanumérica de pelo menos 32 caracteres.
make dev
```

Abra http://localhost:8088. API local: http://localhost:8081/health e `/ready`.
A senha deve ser alfanumérica nesta configuração, pois é interpolada em uma URL PostgreSQL.
O banco não publica portas no host. A migração é um serviço de execução única antes da API.

```sh
cd api
go run ./cmd/certguard check example.com
# Use somente domínios que você está autorizado a monitorar.
```

O comando gera JSON; retorna código de saída diferente de zero se a verificação falhar. `success` pode ter severidade `warning` ou `critical` por proximidade do vencimento.

```sh
cd web && npm ci && cd ..
make test
make lint
make build
```

Testes PostgreSQL: configure `TEST_DATABASE_URL` apontando para banco descartável e rode `go test ./migrations` em `api/`. Sem essa variável o teste é explicitamente pulado; a CI provisiona PostgreSQL 17.

## Documentação
- [Plano desta etapa](docs/plano-etapa-01.md) e [planejamento completo](docs/planejamento-completo.md)
- [Arquitetura](docs/architecture.md), [segurança](docs/security.md) e [ameaças](docs/threat-model.md)
- [Desenvolvimento](docs/development.md), [API](docs/api.md) e [OpenAPI](docs/openapi.yaml)
- [Implantação](docs/deployment.md) e [decisões](docs/decisions/)
- [Validação da entrega](docs/validation.md)

## Ambientes planejados
Homologação: `certguard-dev.franklem.uk`. Produção: `certguard.franklem.uk`.
Projetos Coolify, redes, credenciais e volumes distintos. Sem deploy automático nesta etapa.

## Licença
Apache-2.0. Copyright 2026 Franklin Martins.
