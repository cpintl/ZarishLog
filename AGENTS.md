# ZarishLog — AI Agent Guide

This repo is a monorepo for an offline-first, multi-tenant humanitarian logistics platform. Keep changes aligned with the architecture and the developer docs below.

## High-signal docs

Start with these when context is needed:

- [README.md](README.md)
- [SETUP.md](SETUP.md)
- [CONFIGURE.md](CONFIGURE.md)
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- [docs/STATUS.md](docs/STATUS.md)
- [MAINTAINERS.md](MAINTAINERS.md)
- [config/reference_data/GLOSSARY.md](config/reference_data/GLOSSARY.md)

Do not duplicate project documentation in AGENTS instructions; link to the canonical file instead.

## Repo layout

- [apps/api](apps/api): Go REST API using Gin, sqlc, and PostgreSQL.
- [apps/web](apps/web): Next.js 16 PWA frontend.
- [apps/mobile](apps/mobile): Expo/React Native client (scaffold only).
- [packages/data-models](packages/data-models): SQL migrations, queries, and shared schema artifacts.
- [packages/business-logic](packages/business-logic): shared Go business rules (FEFO, AMC).
- [packages/ui](packages/ui): reusable frontend UI components (scaffold only).
- [scripts](scripts): local bootstrap, sandbox, validation, and build helpers.

## Toolchain versions

| Tool     | Version                   | Notes                                          |
| -------- | ------------------------- | ---------------------------------------------- |
| Go       | 1.26 (CI pins 1.26.4)    | `go.work` ties `apps/api` + `packages/business-logic` |
| Node.js  | 24 (`.nvmrc`)             | CI uses 24 — `apps/web/package.json` requires `>=24` |
| pnpm     | 12.0 (packageManager)     | `corepack enable && corepack prepare pnpm@12 --activate` |
| Postgres | 18 (Docker)               | Local dev password: `zarishlog_dev_password`   |
| Docker Compose | v2.32+               | See `docker-compose.yml` for all services       |

## Go workspace

`go.work` at the repo root links two Go modules:
- `apps/api` (Gin + sqlc + sqlx + go-playground/validator)
- `packages/business-logic` (pure business rules, no HTTP)

Run Go commands from the module directory, not the repo root (e.g., `cd apps/api && go test ./...`).

## sqlc workflow (critical)

SQL queries live in `packages/data-models/sql/queries/*.sql`. sqlc generates typed Go code into `apps/api/internal/db/`.

After editing any `.sql` query file:

```bash
cd apps/api && sqlc generate
```

Config: `apps/api/sqlc.yaml`. Generated files (`*.sql.go`, `models.go`, `querier.go`) are checked in — commit them.

Schema migrations live in `packages/data-models/sql/migrations/` and are numbered sequentially (001–006). Apply in filename order.

## API conventions

### Response helpers (`internal/response/`)

Use the `response` package for all JSON responses:
- `response.OK(c, data)` — 200 with `{"data": ...}`
- `response.Created(c, data)` — 201 with `{"data": ...}`
- `response.Paginated(c, data, total, page, pageSize)` — 200 with `{"data": [...], "total": N, "page": N, ...}`
- `response.NotFound(c, msg)`, `response.BadRequest(c, msg)`, `response.InternalError(c, msg)` — error responses with `{"code": "...", "message": "..."}`
- `response.Validation(c, details)` — 422 with field errors

### Validation (`internal/validator/`)

Use `validator.BindAndValidate(c, &req)` to bind JSON and validate in one step. Custom tags:
- `uuid7` — UUIDv7 format (required for all entity IDs)
- `date` — `YYYY-MM-DD` format
- `item_type`, `movement_type`, `wh_type`, `loc_type`, `uom_category` — domain enums
- `opt_uuid7`, `opt_date` — optional versions of the above

### Route structure

All API routes live under `/api/v1`. Protected routes require OIDC JWT via `middleware.Auth()`. Role-based access is enforced per route group via `middleware.RequireRole(...)`.

### Middleware stack (applied in order)

1. `gin.Recovery()` + `gin.Logger()`
2. `middleware.ErrorHandler()`
3. `middleware.CORS()`
4. `middleware.Tenant()` — extracts org from JWT, sets RLS context
5. `middleware.Auth(cfg)` — on protected groups only
6. `middleware.Audit(db)` — on protected groups only

### Multi-tenancy

Every tenant-scoped table uses `org_id` with Row-Level Security. Do not bypass this pattern. The `middleware.Tenant()` handler sets the PostgreSQL session variable that RLS policies read.

## Frontend conventions

- **Framework**: Next.js 16 App Router, React 19, TypeScript 5.9, Tailwind CSS 4.
- **PWA**: Service worker via Workbox (`sw.ts`), offline support via Dexie.js + IndexedDB (`lib/db.ts`).
- **Tests**: Vitest + jsdom. Test files live in `lib/**/*.test.ts(x)` and `hooks/**/*.test.ts(x)` — these paths are configured in `vitest.config.ts`.
- **Lint**: ESLint 9 flat config extending `next/core-web-vitals`.
- **Formatting**: Prettier with double quotes, trailing commas, 100 char width, LF line endings (`.prettierrc`).
- **Path alias**: `@/` maps to the web app root (configured in `vitest.config.ts` resolve alias).

## Testing

### Go tests

```bash
cd apps/api && go test ./... -v -race -count=1          # full suite
cd apps/api && go test ./... -short -count=1             # fast (skip integration)
cd packages/business-logic && go test ./... -v -count=1  # business rules only
```

Tests are colocated with source files (e.g., `handler/stock_test.go`). CI runs with a live Postgres service container — some tests may need a database.

### Frontend tests

```bash
cd apps/web && pnpm test          # vitest run
cd apps/web && pnpm lint          # eslint
cd apps/web && pnpm typecheck     # tsc --noEmit
```

### CI validation order (`.github/workflows/ci.yml`)

Backend: `go vet` → `go test -race` → `go build`
Frontend: `pnpm lint` → `pnpm typecheck` → `pnpm test` → `pnpm build`

Run `make lint` or `make test` from the repo root to validate everything.

## Pre-commit hook

`.githooks/pre-commit` runs `go vet` on `apps/api` and checks for merge conflict markers in `.go`, `.ts`, `.tsx`, `.md`, `.sql` files. Git hooks are not auto-installed — run `git config core.hooksPath .githooks` to enable.

## Change workflow

1. Read the relevant code and docs before editing.
2. Match the existing patterns in the affected module.
3. Prefer small, focused changes over broad refactors.
4. If changing SQL or schema, update the migration file and any matching Go types or queries. Regenerate with `sqlc generate`.
5. Run the smallest relevant validation command after the change.
6. If the work affects phase status or architecture assumptions, keep [docs/STATUS.md](docs/STATUS.md) and the linked docs accurate.

## Guardrails

- Do not add ad hoc repositories or duplication when the repo already uses direct SQL and typed query patterns.
- Do not create fake app structure or new frameworks that are inconsistent with the existing monorepo.
- Do not treat config files or serialized templates as disposable; they are part of the project's operational model.
- Prefer root-cause fixes and minimal edits.
- Never commit secrets, tokens, or environment files. `.env` is gitignored.

## Good examples to follow

- API handlers and validation patterns in [apps/api/internal/handler](apps/api/internal/handler)
- Response and error conventions in [apps/api/internal/response/response.go](apps/api/internal/response/response.go)
- Database and migration conventions in [packages/data-models/sql](packages/data-models/sql)
- Business logic patterns in [packages/business-logic](packages/business-logic)
- Product/module status and roadmap in [docs/STATUS.md](docs/STATUS.md)
- Local environment setup in [SETUP.md](SETUP.md)
