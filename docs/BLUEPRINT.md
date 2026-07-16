# ZarishLog — Delivery Blueprint

This blueprint turns the PRD into a buildable, phased plan. It assumes a **non-coder founder working with an AI pair-programmer** (Claude Code / Claude Cowork / GitHub Copilot / Gemini / Codex / OpenCode, etc.) rather than a hired engineering team, so each phase is scoped to be deliverable through specification + AI-assisted implementation + review, not manual coding.

## 0. Working Model

1. **You specify** — describe the module/behavior in plain language, referencing this PRD.
2. **AI Coding Agents builds** — generates schema, API, and UI code directly in the monorepo, running tests as it goes.
3. **You review** — check the running app (screenshots, demo data), not the code.
4. **Iterate** — refine via conversation, not manual edits.
5. **Ship** — GitHub Actions handles build/test/deploy on merge to `main`, gated by a manual "Approve" click in GitHub Environments (satisfies the GUI-driven CI/CD requirement).

Each phase below ends with a **demoable milestone** — something you can click through, not just a checklist.

## Phase 1 — Foundation (Repo, Schema, Infra)

**Goal:** Empty-but-real system running locally with the full data model in place.

- [ ] Initialize Turborepo monorepo with `apps/web`, `apps/api`, `apps/mobile`, `packages/*` per README structure
- [ ] Provision PostgreSQL (Docker Compose locally; Supabase/Neon free tier for shared dev)
- [ ] Implement full schema: Organizations, Org Levels (L1–L4), Programs, Departments, Item Master, Categories, UoM, Warehouses, Locations, Stock Levels, Stock Movements, Batches, Users, Roles, Permissions, Audit Log
- [ ] Set up Prisma/Drizzle migrations + seed scripts
- [ ] Import master catalogue from `master_product_list.csv` + reconciled catalogue docs into the new Item Master (dedupe pass)
- [ ] Stand up Keycloak (or Supabase Auth) with the 9 roles pre-configured
- [ ] Docker Compose for local dev: Postgres, Redis, MinIO, Keycloak
- [ ] GitHub repo + Actions skeleton (lint, typecheck, test on PR)

**Milestone:** `docker compose up` boots the full stack; you can log in as GLOBAL_ADMIN and see the imported catalogue in a raw admin table.

## Phase 2 — Core API & Business Logic

- [ ] Product/Item CRUD + search/filter endpoints
- [ ] Warehouse & location management endpoints
- [ ] Stock level query endpoints (by item/warehouse/batch)
- [ ] Goods Receipt (GRN) procedure with batch/expiry capture
- [ ] Goods Issue (SRF) procedure with FEFO-suggested picking
- [ ] Transfer and Adjustment procedures with reason codes and approval gating
- [ ] AMC/FMC calculation service + reorder-point engine
- [ ] CSV import/export procedures
- [ ] RBAC middleware enforcing the role×scope×action matrix on every endpoint
- [ ] Audit logging middleware (auto-captures actor, before/after state)

**Milestone:** Full CRUD + core stock lifecycle testable via API client (Postman/Bruno), with permission checks demonstrably blocking out-of-scope actions.

## Phase 3 — Web Console (Dashboard, Products, Warehouses)

- [ ] Next.js app shell, role-based navigation
- [ ] Dashboard: inventory summary, low-stock widget, recent activity, key charts
- [ ] Product list/detail/add/edit/delete + CSV import/export UI
- [ ] Warehouse list/detail, location hierarchy view, capacity visualization

**Milestone:** A warehouse officer can log in, browse the catalogue, and manage warehouse structure end-to-end in the browser.

## Phase 4 — Inventory Operations UI

- [ ] Stock level view (filterable by warehouse/item/batch/status)
- [ ] GRN, SRF, Transfer, Adjustment forms
- [ ] Batch/serial tracking interface
- [ ] Barcode/QR scanning interface (web + mobile camera)

**Milestone:** A full goods-receipt-to-issue cycle can be performed through the UI, and stock levels update correctly with FEFO applied.

## Phase 5 — QA, Distribution & Asset Management

- [ ] QA inspection form + quarantine management
- [ ] Distribution/delivery form with program allocation
- [ ] Returns & disposal workflow
- [ ] Asset register UI (separate from consumable inventory), custody transfer, maintenance log

**Milestone:** Items can be received, QA-approved, distributed to a program, and a durable asset can be tracked through its custody lifecycle.

## Phase 6 — Reporting, Notifications & User Management

- [ ] Inventory status, movement, valuation, audit-trail reports
- [ ] Low-stock, expiry, utilization dashboards; PDF export
- [ ] Notification rules engine + email delivery (free-tier transactional email) + in-app center
- [ ] User management UI: add/edit, role assignment, activity log, password/MFA management

**Milestone:** An admin can configure alerts, and a country rep can pull a compliance report without touching the database.

## Phase 7 — Offline-First & Mobile

- [ ] Local-first data layer (RxDB/PowerSync + IndexedDB on web, SQLite on mobile)
- [ ] Service worker + PWA manifest for installable web app
- [ ] Sync engine with conflict resolution (append-only ledger for stock transactions)
- [ ] Offline mode indicator + sync status monitoring
- [ ] Expo mobile app sharing business-logic package with web; barcode scanning via device camera

**Milestone:** A storekeeper can complete a full receive/issue cycle on a phone with airplane mode on, then watch it sync cleanly when reconnected.

## Phase 8 — Hardening, Testing & Deployment

- [ ] Unit tests for business logic (FEFO, AMC/FMC, permission matrix)
- [ ] Integration tests for core workflows (GRN → stock → issue → report)
- [ ] Load testing at 22,000+ item / multi-year transaction volume
- [ ] Security review (RBAC boundaries, tenant isolation, dependency audit)
- [ ] Terraform modules for self-host / cloud / hybrid deployment
- [ ] Production GitHub Actions pipeline with manual approval gate
- [ ] Monitoring (Beszel/Uptime Kuma) and backup automation (pg_dump to object storage on schedule)

**Milestone:** One-click (Actions "Run workflow") deploy to a staging environment, promotable to production with a single approval click.

## Phase 9 — Documentation & Handover

- [ ] API documentation (OpenAPI/Swagger, auto-generated from NestJS)
- [ ] User guide (per role) and Admin guide
- [ ] Deployment guide (self-host, cloud, hybrid)
- [ ] Training materials (short screen-recordings preferred over long PDFs)
- [ ] Final packaged release (tagged GitHub release + Docker images)

## Always-Free Resource Map

| Need | Free-tier option |
|---|---|
| Postgres hosting | Supabase (500MB free) / Neon (0.5GB free, branching) / self-host on a free-tier VPS |
| App/API hosting | Vercel (frontend, free hobby tier) / Railway or Fly.io (small free allowance) / self-host |
| Object storage | Cloudflare R2 (10GB free) / self-hosted MinIO |
| Auth | Supabase Auth free tier / self-hosted Keycloak / OAuth2 (google) / Firebase Auth (no usage limits, just compute) |
| Email | Resend (3,000 emails/month free) |
| CI/CD | GitHub Actions (2,000 free minutes/month on private repos, unlimited on public) |
| Monitoring | Beszel / Uptime Kuma (self-hosted, free) |
| BI dashboards | Metabase CE (self-hosted, free) |
| DNS/CDN/TLS | Cloudflare free tier |

This keeps the entire pilot deployment (single organization, a handful of warehouses) runnable at **$0–5/month** (a small VPS is typically the only unavoidable cost if not using PaaS free tiers).

## Related Documents
- `README.md` — quick start and stack summary
- `PRODUCT_REQUIREMENTS_DOCUMENT.md` — full functional/non-functional scope
- `ARCHITECTURE.md` — technical design detail
