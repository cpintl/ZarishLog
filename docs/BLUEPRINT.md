# ZarishLog — Build Blueprint

**Tech Stack:** Go 1.26 + Gin 1.12 · Next.js 15 + React 19 · PostgreSQL 18 + sqlc + sqlx · Keycloak 26 · Dexie.js + Workbox · Docker Compose · Terraform · GitHub Actions

---

## Phase 0 — Foundation (Week 1)

- [x] Master Catalogue design (1912 products across 65+ categories, 12 domains)
- [x] Architecture decisions (Go/Gin over NestJS, sqlc over Prisma)
- [x] Monorepo structure scaffold
- [x] AGENTS.md and sandbox environment
- [x] Docker Compose (PostgreSQL 18, Redis 8, MinIO, Keycloak 26, Meilisearch)
- [x] .env / .env.example with all services
- [x] GitHub CI workflow (go vet + test, pnpm lint + typecheck + build)
- [x] Basic Makefile for local dev

## Phase 1 — Database & Data Models (Week 2)

- [x] SQL schema — 76 tables across 16 domains with:
  - UUIDv7 primary keys
  - `org_id` for multi-tenancy
  - Audit columns (`created_by`, `updated_by`, `created_at`, `updated_at`)
  - RLS policies on every tenant table
- [x] sqlc configuration (137 queries, 16 query files) — type-safe Go query generation
- [x] Master data seeding (23 UoM, 13 categories, 12 roles, 63 permissions, 40 products)
- [x] Sample org hierarchy seed (L1–L4 from CSV, 8 org levels)
- [x] Sample product catalogue seed (40 items from master_product_list.csv)
- [x] Indexes for performance-critical queries (stock_levels, movements, batches, alerts)
- [x] Expanded pharmaceutical catalogue (1912 products across 65+ categories from Bangladesh national drug database, medical supplies, equipment, and assets)

## Phase 2 — Go API Core (Week 3)

- [x] Go module init (`github.com/cpintl/zarishlog-api`) — `apps/api/`
- [x] Configuration layer (viper/envconfig) — `internal/config/`
- [x] Database connection pool (sqlx)
- [x] Middleware: auth (OIDC/JWT validation), RBAC (`RequireRole` on all groups), tenant context — `internal/middleware/`
- [x] Health check endpoint — `GET /api/v1/health`
- [x] Error handling middleware (structured JSON errors) — `internal/response/` error types + `internal/middleware/error.go`
- [x] Request validation (bindings) — `internal/validator/` with custom UUIDv7/enum/date validators via go-playground/validator
- [x] Pagination helpers — `internal/pagination/` (page/page_size query params, LIMIT/OFFSET, COUNT queries)
- [x] Audit logging middleware — `internal/middleware/audit.go` (async inserts to audit_log table)

## Phase 3 — Product/Catalogue Module (Week 4)

- [x] Category CRUD (List/Create) — handlers in `internal/handler/category.go`
- [x] Product CRUD (List/Get/Create/Update/Delete) — `internal/handler/product.go`
- [x] UoM CRUD (List/Get/Create/Update/Delete) — `internal/handler/uom.go`
- [x] Bulk CSV import with validation — `internal/handler/import.go`
- [x] Search endpoint (PostgreSQL ILIKE) — `internal/handler/import.go`
- [x] Unit tests (testify + sqlmock)
- [ ] Integration tests (testcontainers-go)

## Phase 4 — Warehouse & Location Module (Week 5)

- [x] Warehouse CRUD (List/Get/Create/Update/Delete) — `internal/handler/warehouse.go`
- [x] Location hierarchy (zone/rack/bin) CRUD with parent tree — `internal/handler/location.go`
- [x] Location constraints (temp/humidity/pharma/hazardous) — `internal/handler/location.go`
- [x] `loc_type` validator + model validation tags
- [ ] Storage condition validation
- [ ] Warehouse association with org levels

## Phase 5 — Stock & Inventory Module (Week 6)

- [x] GRN (Goods Receipt Note) — create with header — `internal/handler/stock.go`
- [x] Stock Issue (SRF) — create with header — `internal/handler/stock.go`
- [x] Inter-warehouse transfer with line items (transactional) — `internal/handler/stock.go`
- [x] Stock adjustment with reason codes (computes difference) — `internal/handler/stock.go`
- [x] Stock ledger (append-only movements) — List with filters
- [x] Stock levels — queried from warehouse × product
- [x] Batch/serial genealogy trail — `GET /stock/batches/:id/trail`
- [ ] Barcode/QR generation

## Phase 6 — Quality Assurance Module (Week 7)

- [ ] QA inspection on receipt
- [ ] Pass/fail/quarantine disposition
- [ ] Quarantine area management (RLS enforced)
- [ ] Expiry monitoring + alert configuration
- [ ] Corrective action / disposal records
- [ ] QA checklist templates

## Phase 7 — Distribution & Asset Management (Week 8)

- [ ] Distribution/delivery forms with program tracking
- [ ] Multi-program allocation
- [ ] Returns and disposal workflow
- [ ] Asset register with lifecycle states
- [ ] Asset transfer/custody-change workflow
- [ ] Depreciation schedule calculation
- [ ] Maintenance/service history

## Phase 8 — Replenishment & Forecasting (Week 9)

- [ ] AMC calculation (3/6/12-month rolling window)
- [ ] Buffer stock and reorder point calculation
- [ ] FMC (Forecasted Monthly Consumption)
- [ ] Low-stock/overstock/sleeping-stock alerts
- [ ] ML forecasting microservice (Prophet)

## Phase 9 — User & Access Management (Week 10)

- [ ] User CRUD with role assignment per org scope
- [ ] Permission matrix (module × action) enforcement
- [ ] Self-service password reset
- [ ] MFA support
- [ ] Full activity/audit log

## Phase 10 — Offline-First & PWA (Week 11)

- [ ] Dexie.js IndexedDB schema matching server tables
- [ ] Workbox service worker with cache strategies
- [ ] Background Sync API for offline writes
- [ ] Conflict resolution (append-only event log)
- [ ] Sync status indicator UI
- [ ] Mobile offline (Expo + SQLite)

## Phase 11 — Reporting & Analytics (Week 12)

- [ ] Metabase dashboards (stock status, movement, valuation)
- [ ] Stock turnover analysis
- [ ] Expiry dashboard
- [ ] Donor/compliance report templates
- [ ] Export to PDF/CSV

## Phase 12 — Deployment & Infrastructure (Week 13-14)

- [ ] Docker multi-stage builds (Go binary, Next.js static)
- [ ] Terraform: VPC, RDS, ElastiCache, ECS/EKS
- [ ] GitHub Actions: build → push → deploy (staging + production)
- [ ] GitHub Environments with manual approval gates
- [ ] Monitoring setup (Uptime Kuma, OpenObserve)
- [ ] Load testing (k6)

---

## Key Milestones

| Milestone | Phase | What it looks like |
|---|---|---|
| **Catalogue Online** | P1–P3 | API serves products; web shows product table |
| **Warehouse Operational** | P4–P5 | GRN, issue, transfer work end-to-end |
| **QA Active** | P6 | Inspections, quarantine, expiry alerts |
| **Assets Tracked** | P7 | Asset lifecycle from acquisition to disposal |
| **AI Forecasting** | P8 | AMC/FMC calculations drive reorder suggestions |
| **Offline Capable** | P10 | Field ops work fully offline, sync on reconnect |
| **Production Ready** | P12 | CI/CD, monitoring, IaC, load tested |
