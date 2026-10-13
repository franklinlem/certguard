# CertGuard

Monitor self-hosted de certificados TLS, com validação de destinos e proteção SSRF.

**Estado: autenticação, endpoints, verificações manuais, agendamento automático opt-in, dashboard, eventos e entregas de alerta opt-in. Ainda não é o MVP completo nem uma versão pronta para produção.**

## Entregue nesta etapa
- Go: inspeção TLS via CLI, validação de cadeia/hostname, SANs, fingerprint, validade e classificação.
- Bloqueio SSRF, DNS fixado no IP aprovado, timeout total e testes TLS isolados.
- API `/health` e `/ready`, PostgreSQL e migração transacional inicial.
- Login, cadastro/edição/pausa/arquivamento de endpoints e histórico na interface React.
- Sessões persistidas, CSRF, rate limiting e verificação manual autenticada com cooldown.
- Agendador com workers limitados, jitter, bloqueio por endpoint e próxima execução persistida.
- Retenção opt-in em lotes, preservando o último resultado de cada endpoint.
- Dashboard global, status por endpoint, detalhes de certificado e gráfico acessível de histórico.
- Motor de alertas com deduplicação, resolução e histórico de eventos observados.
- Métricas protegidas, contadores duráveis, readiness de schema e ensaio CI de backup/restauração.
- Docker Compose e GitHub Actions com integração PostgreSQL e E2E.
- Planejamento, decisões técnicas, modelo de ameaças e roteiro de homologação.

Entregas por SMTP/webhook estão implementadas e desabilitadas por padrão; recuperação de senha, retenção de eventos/entregas e rotação assistida de chave permanecem pendentes. As verificações manuais estão restritas ao administrador autenticado.

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

## Agendamento e retenção (etapa 03)

Consulte [a operação do agendador](docs/scheduler.md) antes de habilitar `SCHEDULER_ENABLED=true`. O padrão permanece desabilitado. Somente endpoints ativos cadastrados pelo administrador são elegíveis; cada inspeção mantém a proteção SSRF. A próxima execução é uma previsão, sujeita a capacidade e disponibilidade.

A limpeza de histórico é independente e desabilitada por padrão. `RETENTION_ENABLED=true` remove definitivamente resultados antigos, mantendo o último de cada endpoint. Revise a janela e o backup antes de habilitar em ambiente real. Esta entrega não altera a VPS.

Planejamento: [etapa 03](docs/plano-etapa-03.md). Decisão: [ADR-005](docs/adr/005-scheduler-retention.md). Evidências: [validação](docs/validation-etapa-03.md).

## Dashboard (etapa 04)

Indicadores incluem todos os endpoints não arquivados, mesmo fora da página atual. O status é o último observado, com data explícita. Detalhes e gráfico mostram dados reais persistidos; erros sem certificado não viram dias fictícios. Use **Atualizar painel** para obter um novo snapshot. O gráfico exibe até 20 checks da página, com tabela acessível.

[Planejamento](docs/plano-etapa-04.md), [ADR-006](docs/adr/006-dashboard-snapshot.md) e [validação](docs/validation-etapa-04.md).

## Eventos de alerta (etapa 05)

Novos checks registram problemas e recuperações, deduplicando repetições. Consulte **Alertas e eventos** no painel e expanda a observação original. A etapa 06 acrescenta canais e envio opt-in por e-mail/webhook; consulte a operação antes de habilitar.

[Planejamento](docs/plano-etapa-05.md), [regras e operação](docs/alerts.md), [ADR-007](docs/adr/007-alert-episodes.md) e [validação](docs/validation-etapa-05.md).

## Canais e entregas (etapa 06)

Canais novos começam pausados. Configurações são cifradas com chave externa; API não retorna URL, segredo, senha ou destinatário. `DELIVERY_ENABLED=false` mantém fila preservada; habilitar processa pendências. Use somente destinos autorizados e valide em homologação.

[Operação e backup da chave](docs/notifications.md), [plano](docs/plano-etapa-06.md), [ADR-008](docs/adr/008-delivery-outbox.md) e [validação](docs/validation-etapa-06.md).

## Hardening e homologação (etapa 07)

`/ready` verifica banco e migrações esperadas. `/metrics` é opt-in por bearer separado, somente na API interna; não é publicado pelo proxy web. Contadores sobrevivem à retenção dos checks. Consulte [observabilidade](docs/observability.md).

[Revisão dos PRs](docs/revisao-prs-etapa-07.md), [roteiro de homologação](docs/homologacao.md), [plano](docs/plano-etapa-07.md), [ADR-009](docs/adr/009-observability-readiness.md) e [validação](docs/validation-etapa-07.md). Homologação/produção não executadas; CI e benchmark não substituem validação da VPS.
