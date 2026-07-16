# ZarishLog — Agent Instructions

## Project Identity
ZarishLog is an open-source, offline-first, multi-tenant humanitarian logistics, supply chain & asset management platform.

## Stack
- **Backend:** Go 1.26 + Gin 1.12 (REST API), sqlc + database/sql (type-safe SQL generation), PostgreSQL 18
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
│   ├── terraform/        # IaC modules
│   └── kubernetes/       # k3s manifests (future)
├── docs/                 # Architecture, PRD, Blueprint docs
├── config/               # CSV metadata, templates, location data
├── .github/workflows/    # CI/CD pipelines
└── scripts/              # Build, seed, utility scripts
```

## Key Conventions
- **Go:** Standard project layout, Gin handlers in `internal/handler/`, sqlc-generated code in `internal/db/`
- **SQL:** All queries in `.sql` files under `packages/data-models/queries/`, type-safe Go via sqlc generate. Generated code (`internal/db/`) is checked into the repo.
- **CRUD:** Handlers embed SQL directly via sqlx (repository layer was removed as dead code — handlers and repo duplicated the same query logic)
- **Frontend:** App Router, server components by default, client components for interactivity
- **Offline:** Dexie.js IndexedDB wrapper + Workbox service worker + Background Sync API
- **Multi-tenancy:** RLS policies on every table using `org_id`, enforced at DB level via `app.current_org_id()` session function
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

# Note: On Linux Mint, the script detects the Ubuntu codename
# for Docker repos. Run 'make db-migrate' after setup runs.
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
make db-migrate         # Run SQL migrations (in order: 001..006)
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
| `README.md` | Project overview, stack, quick start |
| `MAINTAINERS.md` | Release process, CI/CD, adding new modules |
| `config/reference_data/GLOSSARY.md` | Standardized terminology and abbreviations |
| `docs/BLUEPRINT.md` | Build phases and deliverables roadmap |
| `docs/ARCHITECTURE.md` | System architecture and API design |
| `docs/STATUS.md` | Current build status and phase tracking |
| `docs/PRODUCT_REQUIREMENTS_DOCUMENT.md` | Product requirements and PRD |

## Agent Workflow
1. Read relevant files first
2. Understand existing patterns before making changes
3. Run `go vet ./apps/api/...` after Go code changes
4. Run `go build ./apps/api/cmd/api` to verify compilation
5. If modifying SQL migrations, update the Go Product model to match
6. If adding queries, add them to `packages/data-models/sql/queries/`, then run `cd apps/api && sqlc generate`
7. Keep documentation in sync with code changes
8. Never commit secrets or API keys
9. Follow Go standard project layout for backend changes
10. Follow Next.js App Router conventions for frontend changes
11. After any migration change, run `make db-migrate` to verify
12. Update STATUS.md phase table when completing/fixing phases
13. New shared packages should be created in `internal/`:
    - `internal/response/` — structured error codes and JSON response helpers
    - `internal/validator/` — custom validators for UUIDv7, enums, dates
    - `internal/pagination/` — pagination param parsing and LIMIT/OFFSET helpers
14. When adding a new handler, always use response.JSON/response.Error* and validator.BindAndValidate
