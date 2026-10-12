# CertGuard

Monitor self-hosted de certificados TLS/HTTPS

## 1. Visão geral

O **CertGuard** é uma aplicação self-hosted para monitorar certificados TLS de sites e serviços autorizados. Seu objetivo principal é detectar certificados próximos do vencimento, expirados, inválidos, com hostname incompatível ou com mudanças inesperadas.

O projeto será composto por:

- API backend em Go
- Banco de dados PostgreSQL
- Worker de verificações TLS
- Dashboard web em React + TypeScript
- Sistema de alertas
- Métricas Prometheus
- Docker Compose para desenvolvimento e produção inicial
- Testes automatizados unitários, integração, segurança e E2E

A primeira versão deve permanecer focada em **monitoramento TLS**. Não será um scanner de vulnerabilidades, pentest, crawler ou ferramenta ofensiva.

---

## 2. Objetivos do produto

### 2.1 Objetivo principal

Permitir que uma pessoa ou equipe cadastre domínios/endpoints TLS e acompanhe:

- validade do certificado;
- emissor;
- Subject e SANs;
- hostname match;
- início e fim da validade;
- cadeia de confiança;
- fingerprint;
- mudanças entre verificações;
- erros de conexão;
- expiração próxima;
- histórico de verificações.

### 2.2 Objetivos secundários

- Fornecer alertas automáticos.
- Expor uma API REST documentada.
- Possuir interface responsiva.
- Operar via Docker Compose.
- Disponibilizar métricas Prometheus.
- Ser adequado para publicação open source no GitHub.
- Permitir evolução futura para SaaS/Cloud.

### 2.3 Não objetivos do MVP

Não implementar no MVP:

- scanner de portas;
- exploração de vulnerabilidades;
- brute force;
- auditoria de cipher suites extensa;
- varredura HTTP completa;
- crawling de páginas;
- certificados cliente;
- ACME/renovação automática;
- gerenciamento de chaves privadas;
- alteração de configuração remota;
- descoberta automática de domínios;
- multi-region;
- alta disponibilidade;
- billing;
- integração com gateways de pagamento.

---

## 3. Público-alvo

- desenvolvedores;
- administradores de sistemas;
- DevOps/SRE;
- pequenas empresas;
- mantenedores de sites;
- pessoas que operam VPSs e múltiplos subdomínios;
- equipes que precisam evitar expiração inesperada de certificados.

---

## 4. Casos de uso principais

1. Cadastrar `example.com:443`.
2. O CertGuard realiza uma conexão TLS segura.
3. O backend extrai os dados do certificado.
4. O resultado é salvo no PostgreSQL.
5. O dashboard mostra o estado atual.
6. O sistema calcula quantos dias faltam para expiração.
7. Se o certificado entrar em faixa de alerta, gera um evento.
8. O alerta pode ser enviado por e-mail ou webhook.
9. O usuário consulta o histórico.
10. Caso o certificado mude, o CertGuard destaca a alteração.

---

## 5. Escopo do MVP

### 5.1 Funcionalidades obrigatórias

- autenticação;
- cadastro de usuário administrador;
- cadastro de endpoints TLS;
- verificações periódicas;
- verificação manual;
- status do endpoint;
- detalhes do certificado;
- hostname validation;
- cadeia de confiança;
- dias restantes;
- histórico;
- dashboard;
- filtros;
- alertas configuráveis;
- e-mail;
- webhook;
- API REST;
- OpenAPI;
- Docker Compose;
- métricas Prometheus;
- logs estruturados;
- testes automatizados.

### 5.2 Status possíveis

#### Endpoint

- `healthy`
- `warning`
- `critical`
- `error`
- `disabled`

#### Check

- `success`
- `dns_error`
- `connection_timeout`
- `connection_refused`
- `tls_error`
- `certificate_expired`
- `certificate_not_yet_valid`
- `hostname_mismatch`
- `untrusted_chain`
- `blocked_target`
- `internal_error`

### 5.3 Regras sugeridas

- **Healthy:** mais de 30 dias até expiração.
- **Warning:** 30 dias ou menos.
- **Critical:** 7 dias ou menos.
- **Expired:** certificado expirado.
- Limiares configuráveis por endpoint posteriormente; no MVP podem existir defaults globais editáveis.

---

# 6. Stack tecnológica

## 6.1 Backend

- **Go 1.24+**
- `net/http` ou framework leve como **Chi**
- `crypto/tls`
- `crypto/x509`
- `net`
- `context`
- `pgx/v5`
- `sqlc` opcional e recomendado
- `goose` ou `golang-migrate`
- `slog` para logs estruturados
- OpenAPI com `oapi-codegen` ou documentação equivalente
- Prometheus client para métricas

### Justificativa

Go é adequado para:

- concorrência;
- networking;
- TLS;
- baixo consumo;
- fácil distribuição;
- aplicações self-hosted;
- workers.

---

## 6.2 Frontend

- **React 19**
- **TypeScript**
- **Vite**
- React Router
- TanStack Query
- React Hook Form
- Zod
- Recharts ou Apache ECharts
- Tailwind CSS ou CSS Modules
- Vitest
- Playwright

### Requisitos de UI

- responsivo;
- dark/light mode;
- acessível;
- desktop-first, mas funcional em mobile;
- sem dependência pesada desnecessária.

---

## 6.3 Banco de dados

- **PostgreSQL 17+**

### Motivos

- confiabilidade;
- JSONB quando necessário;
- constraints;
- índices;
- migrations;
- histórico temporal;
- fácil operação em Docker.

---

## 6.4 Infraestrutura

### MVP

- Docker
- Docker Compose
- PostgreSQL
- backend
- frontend
- reverse proxy opcional
- Prometheus opcional no compose de desenvolvimento
- Grafana opcional em perfil separado

### Produção

Pode ser publicado atrás de:

- Traefik
- Caddy
- Nginx

---

# 7. Arquitetura

```text
┌─────────────────────────────┐
│           Browser           │
│ React + TypeScript          │
└──────────────┬──────────────┘
               │ HTTPS / REST
               ▼
┌─────────────────────────────┐
│          Go API             │
│ Auth / Endpoints / Checks   │
│ Alerts / Metrics / OpenAPI  │
└──────┬─────────────┬────────┘
       │             │
       │             ▼
       │   ┌──────────────────┐
       │   │  Check Scheduler │
       │   │  Worker Pool     │
       │   └────────┬─────────┘
       │            │ TLS
       │            ▼
       │     Internet targets
       │
       ▼
┌─────────────────────────────┐
│         PostgreSQL          │
└─────────────────────────────┘
```

---

# 8. Componentes

## 8.1 API

Responsável por:

- autenticação;
- CRUD de endpoints;
- leitura de checks;
- histórico;
- alertas;
- configurações;
- health checks;
- métricas;
- API tokens futura.

## 8.2 Scheduler

Responsável por:

- identificar endpoints vencidos para nova execução;
- distribuir verificações;
- evitar verificações duplicadas;
- aplicar jitter;
- respeitar limites de concorrência.

## 8.3 Worker TLS

Responsável por:

1. validar target;
2. resolver DNS;
3. aplicar proteção SSRF;
4. conectar TCP;
5. realizar handshake TLS;
6. capturar cadeia;
7. validar certificado;
8. calcular fingerprint;
9. salvar resultado.

## 8.4 Alert Engine

Responsável por:

- detectar mudança de estado;
- detectar aproximação de validade;
- evitar alertas repetidos;
- criar eventos;
- disparar canais.

---

# 9. Segurança

Segurança é um requisito central.

## 9.1 SSRF

O CertGuard recebe hosts informados por usuários. Isso cria risco de SSRF.

### Bloquear por padrão

- `127.0.0.0/8`
- `::1`
- `10.0.0.0/8`
- `172.16.0.0/12`
- `192.168.0.0/16`
- link-local
- multicast
- IPv6 ULA
- ranges reservados
- metadata services Cloud
- endereços não roteáveis

### Proteções

- resolução DNS controlada;
- validação de todos os IPs resolvidos;
- revalidação antes da conexão;
- bloquear redirects não é relevante para TLS puro, pois não haverá HTTP follow no MVP;
- limitar portas;
- default apenas `443`;
- se portas customizadas forem permitidas, limitar a portas explicitamente aceitas;
- timeout curto;
- tamanho limitado de dados;
- sem proxy arbitrário;
- sem HTTP fetch.

### Modo self-hosted avançado

Versão futura pode permitir redes privadas mediante configuração administrativa explícita.

---

## 9.2 Autenticação

MVP:

- e-mail + senha;
- hash com Argon2id ou bcrypt bem configurado;
- sessão via cookie HttpOnly, Secure, SameSite;
- CSRF protection se usar cookie auth;
- logout global opcional;
- reset de senha via token;
- rate limiting em login.

Futuro:

- TOTP;
- WebAuthn;
- SSO.

---

## 9.3 Headers

Aplicar:

- Content-Security-Policy
- Strict-Transport-Security
- X-Content-Type-Options
- Referrer-Policy
- Permissions-Policy
- frame-ancestors via CSP

---

## 9.4 Logs

Nunca registrar:

- senhas;
- tokens;
- cookies;
- bodies sensíveis;
- credenciais SMTP;
- segredos de webhook.

---

# 10. Modelo de dados

## 10.1 users

```text
id UUID PK
email VARCHAR UNIQUE NOT NULL
password_hash TEXT NOT NULL
display_name VARCHAR
role VARCHAR
created_at TIMESTAMPTZ
updated_at TIMESTAMPTZ
last_login_at TIMESTAMPTZ NULL
```

## 10.2 monitored_endpoints

```text
id UUID PK
name VARCHAR NOT NULL
hostname VARCHAR NOT NULL
port INTEGER NOT NULL DEFAULT 443
enabled BOOLEAN DEFAULT TRUE
interval_minutes INTEGER DEFAULT 360
warning_days INTEGER DEFAULT 30
critical_days INTEGER DEFAULT 7
last_check_at TIMESTAMPTZ NULL
next_check_at TIMESTAMPTZ NULL
created_at TIMESTAMPTZ
updated_at TIMESTAMPTZ
```

Constraints:

- porta entre 1 e 65535;
- hostname normalizado;
- `warning_days > critical_days >= 0`;
- interval mínimo, por exemplo 5 minutos.

## 10.3 certificate_checks

```text
id UUID PK
endpoint_id UUID FK
checked_at TIMESTAMPTZ
status VARCHAR
duration_ms INTEGER
resolved_ip INET NULL
tls_version VARCHAR NULL
cipher_suite VARCHAR NULL
subject TEXT NULL
issuer TEXT NULL
serial_number TEXT NULL
not_before TIMESTAMPTZ NULL
not_after TIMESTAMPTZ NULL
days_remaining INTEGER NULL
fingerprint_sha256 TEXT NULL
hostname_valid BOOLEAN NULL
chain_valid BOOLEAN NULL
error_code VARCHAR NULL
error_message TEXT NULL
raw_metadata JSONB NULL
```

## 10.4 certificate_sans

```text
id UUID PK
check_id UUID FK
san_type VARCHAR
san_value TEXT
```

## 10.5 alert_channels

```text
id UUID PK
type VARCHAR
name VARCHAR
enabled BOOLEAN
config_encrypted BYTEA
created_at TIMESTAMPTZ
updated_at TIMESTAMPTZ
```

## 10.6 alert_events

```text
id UUID PK
endpoint_id UUID FK
check_id UUID FK
event_type VARCHAR
severity VARCHAR
dedupe_key VARCHAR
created_at TIMESTAMPTZ
resolved_at TIMESTAMPTZ NULL
```

## 10.7 alert_deliveries

```text
id UUID PK
alert_event_id UUID FK
channel_id UUID FK
status VARCHAR
attempt_count INTEGER
last_error TEXT NULL
next_attempt_at TIMESTAMPTZ NULL
created_at TIMESTAMPTZ
updated_at TIMESTAMPTZ
```

---

# 11. Índices

Criar índices em:

- `monitored_endpoints(next_check_at)`
- `certificate_checks(endpoint_id, checked_at DESC)`
- `certificate_checks(status, checked_at DESC)`
- `alert_events(endpoint_id, created_at DESC)`
- `alert_deliveries(status, next_attempt_at)`
- `users(email)`

Avaliar retenção de checks antigos para evitar crescimento ilimitado.

---

# 12. API REST

Prefixo:

```text
/api/v1
```

## Auth

```text
POST   /auth/login
POST   /auth/logout
GET    /auth/me
POST   /auth/forgot-password
POST   /auth/reset-password
```

## Endpoints

```text
GET    /endpoints
POST   /endpoints
GET    /endpoints/{id}
PATCH  /endpoints/{id}
DELETE /endpoints/{id}
POST   /endpoints/{id}/check
POST   /endpoints/{id}/enable
POST   /endpoints/{id}/disable
```

## Checks

```text
GET /endpoints/{id}/checks
GET /checks/{checkId}
```

## Dashboard

```text
GET /dashboard/summary
GET /dashboard/expiring
GET /dashboard/recent-changes
```

## Alerts

```text
GET    /alert-channels
POST   /alert-channels
PATCH  /alert-channels/{id}
DELETE /alert-channels/{id}
POST   /alert-channels/{id}/test
GET    /alerts
```

## System

```text
GET /health
GET /ready
GET /metrics
```

---

# 13. Exemplo de resposta

```json
{
  "id": "uuid",
  "endpoint": {
    "hostname": "example.com",
    "port": 443
  },
  "status": "healthy",
  "checked_at": "2026-10-12T12:00:00Z",
  "certificate": {
    "issuer": "Let's Encrypt",
    "subject": "example.com",
    "not_before": "2026-09-01T00:00:00Z",
    "not_after": "2026-11-30T23:59:59Z",
    "days_remaining": 49,
    "hostname_valid": true,
    "chain_valid": true,
    "fingerprint_sha256": "..."
  }
}
```

---

# 14. Scheduler

## Algoritmo

A cada ciclo:

1. buscar endpoints:
   - enabled;
   - `next_check_at <= now()`;
2. aplicar lock;
3. enviar para worker pool;
4. worker realiza check;
5. salvar resultado;
6. calcular novo status;
7. atualizar:
   - last_check_at;
   - next_check_at;
8. avaliar alertas.

### Concorrência

Config:

```text
CHECK_WORKERS=10
CHECK_TIMEOUT=10s
```

Aplicar jitter para impedir thundering herd após restart.

---

# 15. Processo TLS

Pseudo-fluxo:

```text
normalize hostname
↓
resolve DNS
↓
validate all resolved IPs
↓
choose allowed target
↓
dial TCP with timeout
↓
TLS handshake
↓
capture ConnectionState
↓
verify hostname
↓
verify chain
↓
extract certificate metadata
↓
calculate SHA-256 fingerprint
↓
persist result
```

Não desativar validação TLS apenas para facilitar coleta.

Para capturar certificados inválidos, pode ser necessário um fluxo separado e cuidadosamente documentado usando `InsecureSkipVerify` apenas para inspeção, seguido de validação manual com `x509.Verify`. Nunca tratar a conexão como confiável apenas porque o handshake foi concluído.

---

# 16. Detecção de mudança

Comparar com último check bem-sucedido:

- fingerprint;
- serial;
- issuer;
- not_after;
- SANs;
- chain;
- TLS version opcional.

Eventos:

```text
certificate_changed
issuer_changed
expiry_changed
hostname_became_invalid
chain_became_invalid
endpoint_recovered
endpoint_failed
expiry_warning
expiry_critical
certificate_expired
```

---

# 17. Alertas

## 17.1 Canais MVP

- webhook;
- SMTP/e-mail.

## 17.2 Regras

Exemplo:

- 30 dias: warning;
- 7 dias: critical;
- 3 dias: critical reminder;
- expirado: critical;
- hostname mismatch: critical;
- cadeia inválida: critical;
- endpoint recuperado: info.

## 17.3 Deduplicação

Não enviar o mesmo alerta a cada check.

Exemplo:

```text
dedupe_key =
endpoint_id:event_type:threshold
```

Reenviar apenas:

- ao cruzar limiar;
- após período configurado;
- quando estado mudar.

---

# 18. Webhooks

Payload:

```json
{
  "event": "certificate_expiry_warning",
  "severity": "warning",
  "timestamp": "2026-10-12T12:00:00Z",
  "endpoint": {
    "hostname": "example.com",
    "port": 443
  },
  "certificate": {
    "days_remaining": 15,
    "not_after": "2026-10-27T23:59:59Z"
  }
}
```

Assinar com:

```text
HMAC-SHA256
```

Headers:

```text
X-CertGuard-Event
X-CertGuard-Timestamp
X-CertGuard-Signature
```

Proteção contra SSRF também é necessária nos webhooks de saída.

---

# 19. Dashboard

## 19.1 Página inicial

Cards:

- total monitorado;
- healthy;
- warning;
- critical;
- errors;
- vencendo em 30 dias.

Tabela:

```text
Endpoint | Status | Expira em | Último check | Próximo check
```

## 19.2 Endpoint

Mostrar:

- hostname;
- IP resolvido;
- porta;
- status;
- issuer;
- subject;
- SANs;
- serial;
- fingerprint;
- validade;
- TLS version;
- cipher;
- chain status;
- histórico;
- gráfico de dias restantes;
- eventos.

## 19.3 Alertas

- ativos;
- resolvidos;
- por severidade;
- histórico de entrega.

---

# 20. UX/UI

## Design

Visual inspirado em ferramentas DevOps modernas.

### Componentes

- cards;
- tables;
- badges;
- timeline;
- charts;
- drawers;
- modal de confirmação;
- toast notifications;
- skeleton loading.

### Status

Nunca comunicar apenas por cor.

Exemplo:

```text
✓ Healthy
! Warning
!! Critical
× Error
```

### Responsividade

Desktop:

- sidebar;
- tabela completa;
- gráficos.

Mobile:

- bottom nav ou sidebar colapsada;
- cards;
- tabelas transformadas em cards.

---

# 21. SEO

O dashboard autenticado não precisa de SEO.

Criar site público/landing:

```text
/
 /features
 /self-hosted-tls-monitoring
 /ssl-certificate-expiration-monitor
 /docs
 /privacy
 /security
```

Implementar:

- sitemap.xml;
- robots.txt;
- canonical;
- Open Graph;
- Schema.org SoftwareApplication;
- documentação indexável.

---

# 22. Observabilidade

## Logs

Formato JSON:

```json
{
  "level": "info",
  "event": "certificate_check",
  "endpoint_id": "...",
  "hostname": "example.com",
  "duration_ms": 421,
  "status": "healthy"
}
```

## Métricas Prometheus

```text
certguard_checks_total
certguard_check_duration_seconds
certguard_check_failures_total
certguard_endpoints_total
certguard_certificates_expiring_total
certguard_alert_deliveries_total
certguard_alert_delivery_failures_total
```

## Health

```text
/health
/ready
```

`ready` deve testar pelo menos banco.

---

# 23. Retenção

Defaults sugeridos:

- checks detalhados: 180 dias;
- agregações futuras: mais longas;
- alert events: 365 dias;
- audit logs: 365 dias.

Criar job periódico de limpeza.

---

# 24. Configuração

Exemplo `.env.example`:

```dotenv
APP_ENV=production
APP_PORT=8080
DATABASE_URL=postgres://certguard:password@db:5432/certguard

SESSION_SECRET=change-me
ENCRYPTION_KEY=change-me

CHECK_WORKERS=10
CHECK_TIMEOUT_SECONDS=10
DEFAULT_CHECK_INTERVAL_MINUTES=360

DEFAULT_WARNING_DAYS=30
DEFAULT_CRITICAL_DAYS=7

SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=

PUBLIC_APP_URL=https://certguard.example.com
```

Nenhum segredo real deve ser commitado.

---

# 25. Estrutura do repositório

```text
certguard/
├── api/
│   ├── cmd/
│   │   └── certguard/
│   ├── internal/
│   │   ├── auth/
│   │   ├── certificate/
│   │   ├── endpoints/
│   │   ├── checks/
│   │   ├── scheduler/
│   │   ├── alerts/
│   │   ├── security/
│   │   ├── storage/
│   │   └── httpapi/
│   ├── migrations/
│   ├── queries/
│   ├── go.mod
│   └── go.sum
├── web/
│   ├── src/
│   ├── public/
│   ├── tests/
│   └── package.json
├── docs/
│   ├── architecture.md
│   ├── security.md
│   ├── api.md
│   ├── deployment.md
│   ├── threat-model.md
│   └── development.md
├── deploy/
│   ├── compose.yaml
│   └── compose.dev.yaml
├── .github/
│   └── workflows/
├── .env.example
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
├── SECURITY.md
├── LICENSE
└── Makefile
```

---

# 26. Git workflow

Branches:

```text
main
develop
feature/*
fix/*
release/*
```

Sugestão simplificada para projeto individual:

- `main` protegida;
- branches curtas;
- Pull Requests;
- squash merge.

Commits:

```text
feat:
fix:
docs:
test:
refactor:
chore:
ci:
```

---

# 27. CI

GitHub Actions:

## Backend

- gofmt;
- go vet;
- staticcheck;
- unit tests;
- race detector;
- build.

## Frontend

- npm ci;
- lint;
- TypeScript check;
- unit tests;
- build.

## Security

- dependency scanning;
- CodeQL;
- Trivy para imagens Docker;
- secret scanning.

## PR

Todo PR deve passar por todos os checks.

---

# 28. Testes

## 28.1 Unitários backend

Testar:

- parsing hostname;
- bloqueio SSRF;
- classificação de IP;
- cálculo de dias;
- fingerprint;
- status;
- thresholds;
- dedupe de alertas;
- payload webhook;
- validação.

Meta inicial:

```text
>= 80% cobertura nos pacotes core
```

Cobertura não deve substituir testes úteis.

---

## 28.2 TLS integration tests

Criar certificados de teste:

- válido;
- expirado;
- not-yet-valid;
- hostname incorreto;
- CA não confiável;
- wildcard;
- SAN múltiplo.

Subir servidores TLS locais isolados em testes.

---

## 28.3 PostgreSQL integration tests

Testar:

- migrations;
- constraints;
- histórico;
- concorrência;
- scheduler locks;
- idempotência;
- limpeza.

Usar Testcontainers quando apropriado.

---

## 28.4 Frontend

Vitest:

- componentes;
- forms;
- status badges;
- date formatting;
- filters.

Playwright:

- login;
- criar endpoint;
- rodar check;
- abrir detalhes;
- alterar endpoint;
- criar alert channel;
- testar filtros.

---

## 28.5 Segurança

Testes específicos:

- tentativa de cadastrar localhost;
- IPv4 privado;
- IPv6 loopback;
- IPv6 ULA;
- hostname resolvendo para IP privado;
- DNS rebind simulado;
- portas inválidas;
- hostname malformado;
- brute force login;
- CSRF;
- webhook para destino privado;
- payloads enormes;
- headers maliciosos.

---

## 28.6 Performance

Testar:

- 100 endpoints;
- 1.000 endpoints;
- worker pools 5/10/25;
- latência;
- consumo de RAM;
- crescimento do banco.

MVP deve suportar pelo menos:

```text
1.000 endpoints
intervalo >= 5 min
em uma VPS pequena
```

sem garantia comercial de escala.

---

# 29. Threat model

Documentar em `docs/threat-model.md`.

Ameaças principais:

- SSRF;
- DNS rebinding;
- brute force;
- session hijacking;
- CSRF;
- XSS;
- SQL injection;
- vazamento de SMTP secrets;
- webhook abuse;
- DoS por endpoints lentos;
- storage exhaustion;
- alert storm;
- supply-chain compromise.

Para cada ameaça:

```text
Threat
Asset
Attack path
Mitigation
Residual risk
```

---

# 30. Backup

Backup necessário:

- PostgreSQL;
- configuração;
- encryption key separadamente.

Não existe necessidade de backup de certificados, pois podem ser reconsultados, mas o histórico está no banco.

Procedimento documentado:

```text
pg_dump
restore
integrity check
```

Testar restauração periodicamente.

---

# 31. Docker

Serviços:

```yaml
services:
  api:
  web:
  db:
```

Perfil opcional:

```yaml
profiles:
  observability:
    prometheus:
    grafana:
```

Aplicar:

- non-root user;
- read-only filesystem onde possível;
- cap_drop;
- healthcheck;
- resource limits;
- volumes mínimos.

---

# 32. Roadmap

## v0.1.0 — Fundação

- estrutura;
- DB;
- migrations;
- auth;
- CRUD endpoints;
- Docker;
- CI.

## v0.2.0 — TLS engine

- DNS;
- SSRF protection;
- TLS checks;
- persistence;
- manual checks.

## v0.3.0 — Scheduler

- periodic checks;
- worker pool;
- jitter;
- locks;
- retention.

## v0.4.0 — Dashboard

- overview;
- endpoints;
- details;
- history;
- charts.

## v0.5.0 — Alerts

- alert engine;
- SMTP;
- webhook;
- dedupe;
- retries.

## v0.6.0 — Hardening

- metrics;
- threat model;
- security tests;
- load tests;
- UX/mobile;
- docs.

## v1.0.0

Release estável do MVP.

---

# 33. Critérios de aceite do MVP

O projeto só deve ser considerado MVP concluído se:

- [ ] usuário consegue autenticar;
- [ ] consegue cadastrar hostname;
- [ ] SSRF bloqueia targets proibidos;
- [ ] consegue executar check manual;
- [ ] scheduler executa checks periódicos;
- [ ] certificado é validado;
- [ ] hostname é validado;
- [ ] cadeia é validada;
- [ ] fingerprint é calculado;
- [ ] histórico é salvo;
- [ ] dashboard mostra status;
- [ ] alertas de expiração funcionam;
- [ ] webhook funciona;
- [ ] SMTP funciona;
- [ ] alertas não são duplicados indiscriminadamente;
- [ ] OpenAPI está disponível;
- [ ] `/health` funciona;
- [ ] `/ready` funciona;
- [ ] `/metrics` funciona;
- [ ] Docker Compose sobe o projeto do zero;
- [ ] migrations rodam automaticamente ou por comando documentado;
- [ ] frontend funciona em desktop e celular;
- [ ] testes unitários passam;
- [ ] testes integration passam;
- [ ] testes E2E passam;
- [ ] testes SSRF passam;
- [ ] documentação de instalação está pronta;
- [ ] nenhum segredo está commitado.

---

# 34. README do GitHub

O README deve incluir:

1. logo/nome;
2. screenshots;
3. descrição curta;
4. features;
5. arquitetura;
6. quick start;
7. Docker Compose;
8. configuração;
9. segurança;
10. API docs;
11. métricas;
12. roadmap;
13. contribuição;
14. licença.

Exemplo de headline:

> **CertGuard — Self-hosted TLS certificate monitoring without surprises.**

---

# 35. Licença

Sugestão inicial:

**Apache License 2.0**

Motivos:

- permissiva;
- adequada a software de infraestrutura;
- proteção relacionada a patentes;
- comum em projetos comerciais/open source.

MIT também é uma alternativa válida se simplicidade for prioridade.

---

# 36. Futuras versões

## v1.x

- MFA/TOTP;
- API tokens;
- Telegram;
- Slack;
- Discord;
- Telegram/WhatsApp via adapters;
- páginas públicas de status;
- importação em massa;
- tags;
- grupos;
- maintenance windows.

## v2

- organizações multi-user;
- RBAC avançado;
- múltiplos projetos;
- certificados internos configuráveis;
- suporte a SMTP TLS;
- IMAP TLS;
- PostgreSQL TLS;
- MQTT TLS;
- APIs Cloud.

## SaaS

Caso vire produto:

- multi-tenant;
- Stripe/Mercado Pago;
- planos;
- quotas;
- workers distribuídos;
- region isolation;
- e-mail transacional;
- terms/privacy;
- LGPD;
- SLA;
- backups automatizados.

---

# 37. Ordem prática de implementação

Recomendada:

```text
1. Repo
2. Go API
3. PostgreSQL
4. migrations
5. endpoint CRUD
6. SSRF validator
7. TLS checker
8. manual check
9. integration tests
10. scheduler
11. React dashboard
12. auth
13. history
14. alert engine
15. webhook
16. SMTP
17. metrics
18. E2E
19. security hardening
20. docs
21. v1.0.0
```

Não começar pelo dashboard.

O coração do CertGuard é:

```text
safe target validation
        +
correct TLS inspection
        +
reliable scheduling
        +
useful alerting
```

---

# 38. Definição de pronto

Uma tarefa só está pronta quando:

- código implementado;
- lint passa;
- testes passam;
- tratamento de erro existe;
- segurança foi considerada;
- documentação foi atualizada;
- nenhum TODO crítico permanece;
- funciona no Docker;
- critério de aceite foi demonstrado.

---

# 39. Primeira milestone recomendada

## Milestone: `TLS Check Engine`

Escopo:

- parser de hostname;
- IP validation;
- DNS resolution;
- proteção SSRF;
- conexão TCP;
- handshake TLS;
- captura da cadeia;
- validação de hostname;
- validação da chain;
- fingerprint;
- dias restantes;
- resultado estruturado;
- testes unitários;
- servidores TLS locais para testes.

Não desenvolver frontend antes dessa milestone estar estável.

Esse componente é o núcleo técnico do produto e pode ser testado de maneira independente.

---

# 40. Resultado esperado

Ao final da versão 1.0, o usuário deverá poder:

1. abrir o CertGuard;
2. cadastrar seus domínios;
3. acompanhar o estado dos certificados;
4. receber alertas antes da expiração;
5. identificar certificados inválidos;
6. consultar o histórico;
7. rodar tudo em sua própria infraestrutura com Docker.

O projeto deverá ser pequeno o bastante para uma instalação self-hosted simples, mas organizado de forma que possa futuramente evoluir para um serviço Cloud.

---

# 41. Resumo técnico

```text
Nome: CertGuard
Backend: Go
Frontend: React + TypeScript
Database: PostgreSQL
Infra: Docker Compose
API: REST/OpenAPI
Metrics: Prometheus
TLS: crypto/tls + crypto/x509
Testing: Go testing + Testcontainers + Vitest + Playwright
CI/CD: GitHub Actions
License: Apache-2.0
MVP target: v1.0.0
```

---

# 42. Princípios do projeto

1. Segurança antes de conveniência.
2. Sem scanner ofensivo.
3. Sem chaves privadas.
4. SSRF bloqueado por padrão.
5. Histórico auditável.
6. Alertas úteis, não barulhentos.
7. Dependências mínimas.
8. Self-hosted primeiro.
9. API documentada.
10. Testes como parte da implementação.
