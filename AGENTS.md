# ZarishLog — Agent Instructions

## Project Identity
ZarishLog is an open-source, offline-first, multi-tenant humanitarian logistics, supply chain & asset management platform.

## Stack
- **Backend:** Go 1.26 + Gin 1.12 (REST API), sqlc + sqlx (type-safe SQL), PostgreSQL 18
- **Frontend:** Next.js 15 + React 19 PWA (Workbox + Dexie.js for offline-first)
- **Mobile:** Expo/React Native (shares business logic)
- **Infrastructure:** Docker Compose, Terraform, GitHub Actions
- **Auth:** Keycloak 26 (OIDC/OAuth2)
- **Search:** Meilisearch
- **Analytics:** Metabase
- **ML Engine:** Go microservice (Prophet forecasting, anomaly detection)

## Monorepo Structure
```
zarishlog/
├── apps/
│   ├── api/              # Go + Gin REST API (backend)
│   ├── web/              # Next.js 15 PWA (frontend)
│   └── mobile/           # Expo/React Native (field ops)
├── packages/
│   ├── data-models/      # SQL migrations, sqlc queries, Go types
│   ├── business-logic/   # Shared Go business rules (FEFO, AMC, etc.)
│   └── ui/               # Shared React components (design system)
├── infrastructure/
│   ├── docker/           # Dockerfiles
│   ├── terraform/         # IaC modules
│   └── kubernetes/       # k3s manifests (future)
├── docs/                 # Architecture, PRD, Blueprint docs
├── config/               # CSV metadata, templates, location data
├── .github/workflows/    # CI/CD pipelines
└── scripts/              # Build, seed, utility scripts
```

## Key Conventions
- **Go:** Standard project layout, Gin handlers in `internal/handler/`, services in `internal/service/`, repository in `internal/repository/`
- **SQL:** All queries in `.sql` files under `packages/data-models/queries/`, generated with sqlc
- **Frontend:** App Router, server components by default, client components for interactivity
- **Offline:** Dexie.js IndexedDB wrapper + Workbox service worker + Background Sync API
- **Multi-tenancy:** RLS policies on every table using `org_id`, enforced at DB level
- **Database:** UUIDv7 primary keys, audit columns (`created_by`, `updated_by`, `created_at`, `updated_at`) on every table
- **Testing:** Go `testing` package + testify, Vitest for frontend

## Sandbox Setup Script

```bash
# Full bootstrap (detects OS, installs Go/Node/Docker/psql, sets up Git hooks, VS Code)
./scripts/zarishlog-setup.sh --yes

# Flags:
#   --yes            Auto-approve all installations
#   --install-go     Force install/upgrade Go
#   --install-node   Force install/upgrade Node.js
#   --install-docker Install Docker if missing
#   --check-only     Prerequisite check only (no installs)
```

## Makefile (Development Toolkit)

```bash
make help               # Show all commands
make setup              # Run full bootstrap script
make docker-up          # Start all services (PostgreSQL, Redis, MinIO, Keycloak, Meilisearch)
make docker-down        # Stop all services
make dev                # Start API (:8080) + Web (:3000) in dev mode
make build              # Build Go binary + frontend
make build-docker       # Build Docker images
make test               # Run all tests (Go + frontend)
make test-go            # Go tests only
make test-web           # Frontend tests only
make lint               # Lint all code
make db-migrate         # Run SQL migrations
make db-seed            # Seed master data
make db-reset           # Drop, recreate, migrate, seed
make validate           # Validate config files (CSV, JSON)
make publish            # Build + push Docker images
```

## Sandbox Scripts (in `scripts/`)

| Script | Purpose |
|---|---|
| `zarishlog-setup.sh` | Bootstrap dev environment (auto-install deps) |
| `dev.sh` | Start infra + install deps + migrate + seed |
| `build.sh` | Build Go binary, frontend, Docker images |
| `test.sh` | Interactive test runner (Go, frontend, integration) |
| `validate-config.sh` | Validate CSV/JSON config files |

## Documentation

| Doc | Purpose |
|---|---|
| `SETUP.md` | Step-by-step dev environment setup |
| `CONFIGURE.md` | Configuration reference (CSV templates, env vars, Docker) |

## Agent Workflow
1. Read relevant files first
2. Understand existing patterns before making changes
3. Run lint/typecheck after code changes
4. Run tests after adding/modifying functionality
5. Keep documentation in sync with code changes
6. Never commit secrets or API keys
7. Follow Go standard project layout for backend changes
8. Follow Next.js App Router conventions for frontend changes
