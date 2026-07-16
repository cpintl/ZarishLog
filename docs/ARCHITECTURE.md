# ZarishLog — Technical Architecture

## 1. Architectural Principles

1. **Offline-first, not offline-tolerant** — every core write operation (receive, issue, transfer, adjust) must work with zero connectivity and sync later.
2. **Multi-tenant by construction** — every table below the Organization level carries `organization_id`; Postgres Row-Level Security (RLS) enforces isolation at the database layer, not just the application layer.
3. **Open-source and free-tier first** — no component is chosen unless it has a genuinely free (not trial) self-host or hosted path.
4. **One data model, many surfaces** — web, mobile, and API share the same TypeScript types and validation (Zod) via a shared package, so business rules (FEFO, AMC) are written once.
5. **Boring where it counts** — the stock ledger, auth, and audit trail use the most conservative, battle-tested technology (Postgres, OIDC). Novelty is reserved for offline sync, where it's actually needed.

## 2. High-Level System Diagram

```
                         ┌───────────────────────────┐
                         │        Users (roles)       │
                         │ R01 Global Admin … R09 Ops  │
                         └──────────────┬──────────────┘
                                         │
      ┌──────────────────────────────────────────────────────────────┐
      │                        Client Layer                           │
      │  ┌───────────────┐   ┌────────────────┐   ┌─────────────────┐│
      │  │  Web PWA        │   │ Mobile (Expo)   │   │ Admin/BI (Metabase)│
      │  │  Next.js 15     │   │ React Native    │   │                 ││
      │  │  + RxDB/PGlite  │   │ + SQLite        │   │                 ││
      │  └───────┬─────────┘   └────────┬────────┘   └────────┬────────┘│
      └──────────┼───────────────────────┼──────────────────────┼────────┘
                 │  REST/GraphQL (online)│                     │ SQL (read-only)
                 │  Sync protocol (offline→online)              │
      ┌──────────▼───────────────────────▼──────────────────────▼────────┐
      │                     API Layer — NestJS                            │
      │  Auth Guard (OIDC/JWT) │ RBAC Middleware │ Audit Interceptor       │
      │  Modules: Catalogue · Warehouse · Stock · GRN/SRF · QA · Assets ·  │
      │           Forecasting (AMC/FMC) · Reporting · Notifications        │
      └──────────┬───────────────────────────────┬─────────────────────────┘
                 │                               │
      ┌──────────▼─────────┐          ┌──────────▼─────────┐
      │  PostgreSQL 17       │          │  Redis + BullMQ      │
      │  (RLS multi-tenant)  │          │  (jobs, notifications)│
      └──────────┬───────────┘          └───────────────────────┘
                 │
      ┌──────────▼─────────┐
      │  MinIO / R2 (S3)      │  photos, GRN scans, batch certs
      └─────────────────────┘
```

## 3. Technology Stack & Rationale

| Concern | Technology | Rationale |
|---|---|---|
| Monorepo tooling | Turborepo (pnpm workspaces) | Incremental caching, shared packages across web/mobile/api; free, actively maintained |
| Web frontend | Next.js 15 + React 19 | Mature PWA support, App Router server components reduce client bundle, huge free-tier hosting availability (Vercel/Cloudflare Pages) |
| UI system | Tailwind CSS + shadcn/ui | No proprietary component licensing; fully customizable |
| Mobile | React Native via Expo | Single codebase for Android/iOS, shares `packages/business-logic` and `packages/data-models` with web; strong offline/camera/barcode libraries |
| API framework | NestJS | Backend-first, modular, dependency-injected structure that stays maintainable as modules grow (Catalogue, Warehouse, Stock, QA, Assets, Forecasting, Reporting) — this project's ~10 modules and long lifespan favor NestJS's discipline over a route-handler-only approach; Next.js API routes remain a valid fast-prototype path for very early phases |
| Database | PostgreSQL 17 | Relational integrity for a 22,000+ item catalogue and an audit-critical transactional ledger; native RLS for multi-tenancy; JSONB for flexible/legacy attributes; free self-host or free-tier managed (Supabase/Neon) |
| ORM | Prisma (or Drizzle for perf-sensitive paths) | Type-safe schema shared with `packages/data-models`; migration history as code |
| Auth | Keycloak (self-host) or Supabase Auth (managed) | OIDC/JWT, RBAC, MFA, SSO federation with donor/partner IdPs if needed later; both are genuinely free |
| Offline sync | RxDB or PowerSync, backed by PGlite (web, Postgres-in-WASM) / SQLite (mobile) | Full SQL locally, proven CRDT/log-based sync to server Postgres; avoids hand-rolled conflict resolution |
| Background jobs | BullMQ + Redis | AMC/FMC recalculation, expiry-alert sweeps, notification dispatch, scheduled exports |
| Object storage | MinIO (self-host, S3 API) or Cloudflare R2 (10GB free) | S3-compatible; interchangeable without code changes |
| CI/CD | GitHub Actions | Free minutes cover this project's scale; GitHub Environments gives the required GUI approval step |
| IaC | Terraform | Modules for `vpc`, `database`, `storage`, `compute`; identical config pattern across self-host/GCP/AWS/Azure |
| Container runtime | Docker Compose (single host) → k3s (lightweight Kubernetes) when scaling beyond one VPS | Avoids operating full Kubernetes complexity until genuinely needed |
| Reverse proxy | Caddy or Traefik | Automatic free TLS (Let's Encrypt), zero-config virtual hosts |
| Monitoring | Beszel or Uptime Kuma + OpenObserve | Lightweight footprint appropriate for a small self-hosted deployment; free |
| BI/reporting | Metabase CE | Point-and-click dashboards directly on Postgres; avoids building a custom BI layer for v1 |
| Email | Resend / Postmark free tier | Transactional email for notifications and reports |

### Why not X?
- **Odoo/OpenBoxes/OpenLMIS as a base** — evaluated (all appear in prior research) but rejected as the *foundation*: they solve adjacent problems (general ERP, medical logistics ordering) but impose their own data model and UI conventions that fight the multi-level L1–L4 + offline-first requirements. They remain valuable as **integration targets** (e.g., Finance via Odoo) and as **reference implementations** to mine for business-rule ideas (AMC formula, FEFO logic), not as the platform itself.
- **Firebase/Firestore** — rejected in favor of Postgres: document model is weaker for the relational integrity a 22,000-item catalogue with batch/serial genealogy needs, and Postgres's RLS gives cleaner multi-tenancy than Firestore security rules at this data complexity.
- **MongoDB** — same reasoning; a normalized relational ledger is safer for financial/audit-grade stock transactions than a document store.

## 4. Data Model Overview

Full field-level schema lives in the working `SCHEMA_DESIGN.md`; the canonical layers are:

**Master Data:** `organizations`, `org_levels` (L1–L4), `programs`, `departments`, `product_categories`, `products` (Item Master), `units_of_measure`, `warehouses`, `locations` (zone/rack/bin).

**Transactional Data:** `stock_levels`, `stock_movements`, `goods_receipts` (GRN), `stock_requests` (SRF), `stock_transfers`, `stock_adjustments`, `qa_inspections`, `asset_transfers`, `distributions`.

**Reference Data:** status enums (Active/Inactive, On Hand/Reserved/In Transit/On Hold/Damaged/Expired/Backordered), UoM codes, document types (GRN/SRF/PR/Transfer/Disposal).

**Configuration Data:** `roles`, `permissions`, `role_permissions`, `workflows` (approval chains), `notification_rules`.

Key relationships:
- `products` → `product_categories` (many-to-one), `products` → `units_of_measure` (many-to-one)
- `stock_levels` is a materialized position keyed on (`product_id`, `warehouse_id`, `location_id`, `batch_id`)
- `stock_movements` is the append-only ledger — `stock_levels` is derived/recomputed from it, never edited directly, which is what makes offline sync safe (movements merge as an event log; positions are recalculated)
- `asset_transfers` tracks custody, separate from `stock_movements`, because assets are individually serialized and depreciated rather than consumed

## 5. Multi-Tenancy & Security Model

- Every tenant-scoped table carries `organization_id`; Postgres RLS policies restrict every query to `current_setting('app.current_org')`, set per-request from the authenticated JWT.
- Role × Scope × Action permission matrix (R01–R09, §4 of PRD) is enforced in a NestJS guard on every endpoint — UI hiding is a convenience, not a security boundary.
- All writes pass through an audit interceptor capturing actor, timestamp, before/after state, and IP — satisfies the audit-trail requirement without per-module custom logging.
- Secrets managed via environment variables + `.env` locally, GitHub Actions secrets in CI, and a proper secrets manager (Doppler free tier or cloud-native equivalent) in production.
- Transport: TLS everywhere via Caddy/Traefik automatic certificates.

## 6. Offline-First Design

1. **Local database:** each client (web via PGlite/RxDB, mobile via SQLite) holds a scoped replica of the data relevant to the user's warehouse(s).
2. **Writes go local-first:** every GRN/issue/transfer/adjustment is written to the local event log immediately and reflected in the UI instantly.
3. **Sync:** on reconnect, the local event log is pushed to the API, which appends to the canonical `stock_movements` ledger; the client then pulls the latest merged position.
4. **Conflict handling:** because stock movements are append-only events (not in-place quantity edits), most "conflicts" are simply concurrent, order-independent events that merge cleanly. True conflicts (e.g., two adjustments on the same batch with contradictory reason codes) are flagged for a supervisor to resolve, never silently overwritten.
5. **Indicators:** UI always shows sync state (Synced / Pending N changes / Syncing) so field staff know their data's status.

## 7. Deployment Topologies

| Topology | Use case | Stack |
|---|---|---|
| **Single VPS (Docker Compose)** | Pilot, single organization | All services on one $5–10/month VPS; Caddy for TLS |
| **Managed free-tier hybrid** | Small NGO, no ops capacity | Vercel (web) + Supabase (Postgres/Auth) + Cloudflare R2 (storage) — entirely free tier |
| **Self-hosted k3s cluster** | Multi-country, higher availability need | k3s on 3+ nodes, Terraform-provisioned, Traefik ingress |
| **Cloud (GCP/AWS/Azure free tier)** | Organization already has cloud credits | Terraform modules per provider; Cloud SQL/RDS for Postgres, GCS/S3 for storage |

All four topologies run the **same containers**, differing only in the Terraform module and environment variables — this is what makes the system genuinely "multi-environment suitable."

## 8. CI/CD Pipeline

```
PR opened → GitHub Actions: lint, typecheck, unit tests, build
  → merge to main → Actions: build Docker images, push to GHCR
  → auto-deploy to staging
  → manual "Approve" click in GitHub Environments (GUI-driven, per requirement)
  → deploy to production (docker compose pull && up -d, or k3s rollout)
```

No custom deployment UI is required — GitHub's native Environments/Actions UI satisfies the "GUI-based, drop-down selection" requirement without building bespoke tooling.

## 9. Localization Defaults

- Timezone: **GMT+6** stored in config, applied at the API layer for all computed fields (e.g., "days until expiry"); raw timestamps stored in UTC in Postgres, converted at the edge.
- Date display: `DD MMMM YYYY` (e.g., "15 July 2026").
- Time display: 12-hour `hh:mm AM/PM`.
- Implemented via a single shared `packages/business-logic/datetime.ts` utility used by both web and mobile — one place to change if requirements shift.

## 10. Related Documents
- `README.md` — overview, stack summary, quick start
- `PRODUCT_REQUIREMENTS_DOCUMENT.md` — functional and non-functional requirements
- `BLUEPRINT.md` — phased delivery plan
