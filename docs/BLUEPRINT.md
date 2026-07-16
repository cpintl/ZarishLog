# ZarishLog — Build Blueprint

**Tech Stack:** Go 1.26 + Gin 1.12 · Next.js 15 + React 19 · PostgreSQL 18 + sqlc + sqlx · Keycloak 26 · Dexie.js + Workbox · Docker Compose · Terraform · GitHub Actions

---

## Phase 0 — Foundation (Week 1)

- [x] Master Catalogue design (200+ entities, 12 domains)
- [x] Architecture decisions (Go/Gin over NestJS, sqlc over Prisma)
- [x] Monorepo structure scaffold
- [x] AGENTS.md and sandbox environment
- [x] Docker Compose (PostgreSQL 18, Redis 8, MinIO, Keycloak 26, Meilisearch)
- [x] .env / .env.example with all services
- [x] GitHub CI workflow (go vet + test, pnpm lint + typecheck + build)
- [x] Basic Makefile for local dev

## Phase 1 — Database & Data Models (Week 2)

- [ ] SQL schema — all 200+ tables with:
  - UUIDv7 primary keys
  - `org_id` for multi-tenancy
  - Audit columns (`created_by`, `updated_by`, `created_at`, `updated_at`)
  - RLS policies on every tenant table
- [ ] sqlc configuration + type-safe Go query generation
- [ ] Master data seeding (UoM, categories, statuses, roles, permissions)
- [ ] Sample org hierarchy seed (L1–L4)
- [ ] Sample product catalogue seed (18+ items from CSV)
- [ ] Indexes for performance-critical queries (stock_levels, movements)

## Phase 2 — Go API Core (Week 3)

- [ ] Go module init (`github.com/cpintl/zarishlog-api`)
- [ ] Configuration layer (viper/envconfig)
- [ ] Database connection pool (sqlx)
- [ ] Middleware: auth (OIDC/JWT validation), RBAC, audit logging, tenant context
- [ ] Health check endpoint
- [ ] Error handling middleware (structured JSON errors)
- [ ] Request validation (bindings)
- [ ] Pagination helpers

## Phase 3 — Product/Catalogue Module (Week 4)

- [ ] Category CRUD (handler → service → repository)
- [ ] Product CRUD (handler → service → repository)
- [ ] UoM CRUD
- [ ] Bulk import CSV/XLSX with validation
- [ ] Search endpoint (with Meilisearch integration)
- [ ] Unit tests (testify + sqlmock)
- [ ] Integration tests (testcontainers-go)

## Phase 4 — Warehouse & Location Module (Week 5)

- [ ] Warehouse CRUD
- [ ] Location hierarchy (zone/rack/bin) CRUD
- [ ] Location constraints (ambient, cold chain, hazardous, secure)
- [ ] Storage condition validation
- [ ] Warehouse association with org levels

## Phase 5 — Stock & Inventory Module (Week 6)

- [ ] GRN (Goods Receipt Note) — receive stock with batch/expiry
- [ ] SRF (Stock Request Form) — issue stock with FEFO enforcement
- [ ] Inter-warehouse transfer
- [ ] Stock adjustment with reason codes
- [ ] Stock ledger (append-only movements)
- [ ] Stock levels (materialized view)
- [ ] Batch/serial genealogy tracking
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
