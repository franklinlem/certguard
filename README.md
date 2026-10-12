# CertGuard

Monitor self-hosted de certificados TLS, com validação de destinos e proteção SSRF.

**Estado: autenticação, endpoints e verificações manuais persistidas. Ainda não é o MVP completo nem uma versão pronta para produção.**

## Entregue nesta etapa
- Go: inspeção TLS via CLI, validação de cadeia/hostname, SANs, fingerprint, validade e classificação.
- Bloqueio SSRF, DNS fixado no IP aprovado, timeout total e testes TLS isolados.
- API `/health` e `/ready`, PostgreSQL e migração transacional inicial.
- Login, cadastro/edição/pausa/arquivamento de endpoints e histórico na interface React.
- Sessões persistidas, CSRF, rate limiting e verificação manual autenticada com cooldown.
- Docker Compose e GitHub Actions com integração PostgreSQL e E2E.
- Planejamento, decisões técnicas, modelo de ameaças e roteiro de homologação.

Scheduler, alertas, recuperação de senha e dashboard agregado pertencem aos próximos incrementos. As verificações manuais estão restritas ao administrador autenticado.

## Desenvolvimento
Requisitos: Go 1.27.2, Node.js 22.23.2+ e Docker com Compose v2.

```sh
cp .env.example .env
# Defina POSTGRES_PASSWORD com senha aleatória alfanumérica de pelo menos 32 caracteres.
make dev
```

Crie o primeiro administrador sem colocar a senha em argumentos ou arquivos:

```sh
read -r -s -p 'Senha inicial (12–72 bytes): ' CERTGUARD_BOOTSTRAP_PASSWORD
printf '%s' "$CERTGUARD_BOOTSTRAP_PASSWORD" | docker compose --env-file .env --project-name certguard-local -f deploy/compose.yaml exec -T api certguard admin-create seu-email@exemplo.com
unset CERTGUARD_BOOTSTRAP_PASSWORD
```

O comando recusa uma segunda inicialização; não altera administradores existentes. Não há senha padrão.

Abra http://localhost:8088. API local: http://localhost:8081/health e `/ready`. PUBLIC_APP_URL deve corresponder exatamente à origem usada no navegador. Em homologação/produção use APP_ENV=production e a URL HTTPS do ambiente.
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
- [Plano da etapa 02](docs/plano-etapa-02.md) e [fundação](docs/plano-etapa-01.md) e [planejamento completo](docs/planejamento-completo.md)
- [Arquitetura](docs/architecture.md), [segurança](docs/security.md) e [ameaças](docs/threat-model.md)
- [Desenvolvimento](docs/development.md), [API](docs/api.md) e [OpenAPI](docs/openapi.yaml)
- [Implantação](docs/deployment.md) e [decisões](docs/decisions/)
- [Validação da etapa 02](docs/validation-etapa-02.md) e [fundação](docs/validation.md)

## Ambientes planejados
Homologação: `certguard-dev.franklem.uk`. Produção: `certguard.franklem.uk`.
Projetos Coolify, redes, credenciais e volumes distintos. Sem deploy automático nesta etapa.

## Licença
Apache-2.0. Copyright 2026 Franklin Martins.
