.PHONY: help dev build test lint setup docker-up docker-down db-up db-down \
        db-migrate db-seed validate-config clean publish version

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

help:
	@echo "╔═══════════════════════════════════════════════════════════════╗"
	@echo "║  ZarishLog Development Toolkit                               ║"
	@echo "║  Version: $(VERSION)                                         ║"
	@echo "╚═══════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  ┌─────────────────────┬─────────────────────────────────────┐"
	@echo "  │ SETUP & ENVIRONMENT │                                     │"
	@echo "  ├─────────────────────┼─────────────────────────────────────┤"
	@echo "  │ make setup          │ Run full bootstrap + install deps   │"
	@echo "  │ make docker-up      │ Start all Docker services           │"
	@echo "  │ make docker-down    │ Stop all Docker services            │"
	@echo "  │ make db-up          │ Start database services only        │"
	@echo "  │ make db-down        │ Stop database services              │"
	@echo "  ├─────────────────────┼─────────────────────────────────────┤"
	@echo "  │ BUILD & DEV         │                                     │"
	@echo "  ├─────────────────────┼─────────────────────────────────────┤"
	@echo "  │ make dev            │ Start API + Web in dev mode         │"
	@echo "  │ make build          │ Build Go binary + frontend          │
	@echo "  │ make build-go       │ Build Go binary only                │
	@echo "  │ make build-docker   │ Build Docker images                 │
	@echo "  ├─────────────────────┼─────────────────────────────────────┤
	@echo "  │ DATABASE            │                                     │
	@echo "  ├─────────────────────┼─────────────────────────────────────┤
	@echo "  │ make db-migrate     │ Run SQL migrations                  │
	@echo "  │ make db-seed        │ Seed master data                    │
	@echo "  │ make db-reset       │ Drop, recreate, migrate, seed       │
	@echo "  ├─────────────────────┼─────────────────────────────────────┤
	@echo "  │ TEST & VALIDATE     │                                     │
	@echo "  ├─────────────────────┼─────────────────────────────────────┤
	@echo "  │ make test           │ Run all tests                       │
	@echo "  │ make test-go        │ Run Go tests only                   │
	@echo "  │ make test-web       │ Run frontend tests only             │
	@echo "  │ make lint           │ Lint all code                       │
	@echo "  │ make validate       │ Validate config files               │
	@echo "  ├─────────────────────┼─────────────────────────────────────┤
	@echo "  │ PUBLISH             │                                     │
	@echo "  ├─────────────────────┼─────────────────────────────────────┤
	@echo "  │ make publish        │ Build + push Docker images          │
	@echo "  │ make release        │ Create release tag + publish        │
	@echo "  └─────────────────────┴─────────────────────────────────────┘
	@echo ""

# ─── Setup ──────────────────────────────────────────────────────────────

setup:
	@bash scripts/zarishlog-setup.sh --yes

# ─── Docker ─────────────────────────────────────────────────────────────

docker-up:
	@docker compose up -d
	@echo "✓ All services started"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis:      localhost:6379"
	@echo "  MinIO:      localhost:9000 (API) / 9001 (Console)"
	@echo "  Keycloak:   localhost:8080"
	@echo "  Meilisearch: localhost:7700"

docker-down:
	@docker compose down
	@echo "✓ All services stopped"

docker-logs:
	@docker compose logs -f

docker-ps:
	@docker compose ps

db-up:
	@docker compose up -d postgres redis
	@sleep 3
	@echo "✓ Database services ready"

db-down:
	@docker compose down postgres redis
	@echo "✓ Database services stopped"

# ─── Database ───────────────────────────────────────────────────────────

db-migrate:
	@echo "Running database migrations..."
	@psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/migrations/001_initial_schema.sql 2>/dev/null || \
	 PGPASSWORD=zarishlog_dev_password psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/migrations/001_initial_schema.sql
	@echo "✓ Migrations applied"

db-seed:
	@echo "Seeding master data..."
	@psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/seed.sql 2>/dev/null || \
	 PGPASSWORD=zarishlog_dev_password psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/seed.sql
	@echo "✓ Seed data loaded"

db-reset:
	@echo "Resetting database..."
	@psql -h localhost -U zarishlog -d zarishlog -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" 2>/dev/null || \
	 PGPASSWORD=zarishlog_dev_password psql -h localhost -U zarishlog -d zarishlog -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MAKE) db-migrate
	@$(MAKE) db-seed
	@echo "✓ Database reset complete"

db-connect:
	@psql -h localhost -U zarishlog -d zarishlog

# ─── Build ──────────────────────────────────────────────────────────────

build: build-go build-web

build-go:
	@echo "Building Go API binary..."
	@cd apps/api && CGO_ENABLED=0 go build \
		-ldflags="-X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME) -s -w" \
		-o bin/api ./cmd/api
	@echo "✓ Go binary built: apps/api/bin/api"

build-web:
	@echo "Building frontend..."
	@cd apps/web && pnpm build
	@echo "✓ Frontend built"

build-docker:
	@echo "Building Docker images..."
	@docker build -f infrastructure/docker/Dockerfile.api -t zarishlog-api:$(VERSION) .
	@docker build -f infrastructure/docker/Dockerfile.web -t zarishlog-web:$(VERSION) .
	@echo "✓ Docker images built"

# ─── Development ────────────────────────────────────────────────────────

dev:
	@echo "Starting development servers..."
	@docker compose up -d postgres redis
	@sleep 2
	@echo "Starting API server on :8080..."
	@cd apps/api && go run -ldflags="-X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME)" ./cmd/api &
	@sleep 1
	@echo "Starting Web server on :3000..."
	@cd apps/web && pnpm dev &
	@wait

# ─── Test ───────────────────────────────────────────────────────────────

test: test-go test-web

test-go:
	@echo "Running Go tests..."
	@cd apps/api && go test ./... -v -race -count=1
	@cd packages/business-logic && go test ./... -v -count=1

test-go-short:
	@echo "Running Go tests (short)..."
	@cd apps/api && go test ./... -short -count=1

test-web:
	@echo "Running frontend tests..."
	@cd apps/web && pnpm test

test-coverage:
	@echo "Running tests with coverage..."
	@cd apps/api && go test ./... -coverprofile=coverage.out -covermode=atomic -count=1
	@cd apps/api && go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: apps/api/coverage.html"

test-integration:
	@bash scripts/test.sh

# ─── Lint ───────────────────────────────────────────────────────────────

lint: lint-go lint-web

lint-go:
	@if command -v golangci-lint &>/dev/null; then \
		cd apps/api && golangci-lint run ./...; \
	else \
		cd apps/api && go vet ./...; \
	fi
	@echo "✓ Go lint passed"

lint-web:
	@cd apps/web && pnpm lint
	@echo "✓ Frontend lint passed"

# ─── Validate ───────────────────────────────────────────────────────────

validate:
	@bash scripts/validate-config.sh

# ─── Publish ────────────────────────────────────────────────────────────

publish:
	@bash scripts/build.sh --docker --publish --version $(VERSION)

release:
	@echo "Creating release $(VERSION)..."
	@git tag -a "$(VERSION)" -m "Release $(VERSION)"
	@git push origin "$(VERSION)"
	@$(MAKE) publish
	@echo "✓ Release $(VERSION) created and published"

# ─── Clean ──────────────────────────────────────────────────────────────

clean:
	@rm -rf apps/api/bin apps/api/dist apps/web/.next apps/web/out
	@rm -rf apps/web/node_modules packages/*/node_modules
	@rm -f apps/api/coverage.out apps/api/coverage.html
	@rm -f packages/business-logic/coverage.out
	@echo "✓ Clean complete"

clean-all: clean
	@rm -rf data/postgres data/redis data/minio data/meilisearch
	@docker compose down -v 2>/dev/null || true
	@echo "✓ Full clean complete (volumes removed)"

# ─── Version ────────────────────────────────────────────────────────────

version:
	@echo "$(VERSION)"
