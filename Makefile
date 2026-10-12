.PHONY: test lint build dev down
GO ?= go
COMPOSE = docker compose --env-file .env --project-name certguard-local -f deploy/compose.yaml -f deploy/compose.dev.yaml
test:
	cd api && $(GO) test -race -cover ./...
	cd web && npm test
lint:
	cd api && test -z "$$($(GO) fmt ./...)" && $(GO) vet ./...
	cd web && npm run lint && npm run typecheck
build:
	cd api && $(GO) build ./cmd/certguard
	cd web && npm run build
dev:
	$(COMPOSE) up --build -d
down:
	$(COMPOSE) down
