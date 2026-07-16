# ZarishLog — Technical Architecture

## 1. Architectural Principles

1. **Offline-first, not offline-tolerant** — every core write operation (receive, issue, transfer, adjust) works with zero connectivity and syncs later.
2. **Multi-tenant by construction** — every table carries `org_id`; PostgreSQL RLS enforces isolation at the database layer.
3. **Open-source and free-tier first** — no component chosen unless it has a genuinely free self-host or hosted path.
4. **Type-safe SQL over ORM** — sqlc generates type-safe Go code from raw SQL. No runtime ORM overhead, no magic.
5. **One data model, many surfaces** — web, mobile, and API share the same SQL schema and business logic written in Go.

## 2. System Architecture

```
                          ┌───────────────────────────┐
                          │        Users (R01–R09)     │
                          └──────────────┬─────────────┘
                                          │
       ┌──────────────────────────────────────────────────────────┐
       │                    Client Layer                           │
       │  ┌────────────┐   ┌──────────────┐   ┌────────────────┐  │
       │  │ Web PWA     │   │ Mobile (Expo)│   │ Metabase (BI)  │  │
       │  │ Next.js 15  │   │ React Native │   │                │  │
       │  │ Dexie.js    │   │ SQLite       │   │                │  │
       │  └──────┬──────┘   └──────┬───────┘   └───────┬────────┘  │
       └─────────┼─────────────────┼────────────────────┼──────────┘
                 │ REST (online)   │                    │ SQL (read)
                 │ Sync protocol   │                    │
       ┌─────────▼─────────────────▼────────────────────▼──────────┐
       │                    API Layer — Go + Gin 1.12                │
       │  Middleware: Auth (OIDC/JWT) · RBAC · Audit · Tenant       │
       │  Modules: Catalogue · Warehouse · Stock · GRN/SRF · QA ·   │
       │           Assets · Forecasting · Reporting · Notifications  │
       └─────────┬──────────────────────────────────────────────────┘
                 │
       ┌─────────▼──────────────────────────────────────────────────┐
       │                    Service Layer (Go)                        │
       │  business-logic: FEFO · AMC · FMC · Reorder · Validation   │
       │  ML Engine: Prophet forecasting · Anomaly detection         │
       └─────────┬──────────────────────────────────────────────────┘
                 │
       ┌─────────▼──────────────────────────────────────────────────┐
       │                    Data Layer                                │
       │  PostgreSQL 18 (sqlc + sqlx)                                │
       │  RLS per org_id · UUIDv7 PKs · Audit columns               │
       │  Append-only stock_movements ledger                         │
       └─────────┬──────────────────────────────────────────────────┘
                 │
       ┌─────────▼─────────┐  ┌──────────▼─────────┐  ┌───────────┐
       │  Redis 8           │  │  MinIO (S3)         │  │ Meilisearch│
       │  (cache, jobs)     │  │  (scans, photos)    │  │ (search)   │
       └───────────────────┘  └─────────────────────┘  └───────────┘
```

## 3. Technology Stack

| Layer | Technology | Rationale |
|---|---|---|
| Backend API | Go 1.26 + Gin 1.12 | Performance, single binary deploy, goroutine concurrency for sync |
| Type-safe SQL | sqlc + sqlx | Generate Go code from SQL; zero-cost abstraction, no ORM magic |
| Web Frontend | Next.js 15 + React 19 | PWA support, App Router, large ecosystem |
| Mobile | Expo/React Native | Code sharing with web via shared components |
| Database | PostgreSQL 18 | RLS, UUIDv7, JSONB, mature replication |
| Auth | Keycloak 26 | OIDC/OAuth2, RBAC, MFA, SSO |
| Offline | Dexie.js + Workbox + Background Sync | Battle-tested IndexedDB wrapper + service worker |
| Object Storage | MinIO (S3-compatible) | Photos, scanned GRNs, batch certificates |
| Search | Meilisearch | Typo-tolerant, instant search for catalogue |
| BI | Metabase CE | Point-and-click over Postgres |
| Job Queue | Redis + custom Go workers | AMC calc, expiry alerts, notifications |
| ML Engine | Go microservice (Prophet bindings) | Forecasting, anomaly detection |

## 4. API Design

```
GET    /api/v1/health
GET    /api/v1/products          # List products (paginated)
POST   /api/v1/products          # Create product
GET    /api/v1/products/:id      # Get product
PUT    /api/v1/products/:id      # Update product
DELETE /api/v1/products/:id      # Soft-delete product
POST   /api/v1/products/import   # Bulk import CSV

GET    /api/v1/categories        # List categories
POST   /api/v1/categories        # Create category

GET    /api/v1/warehouses        # List warehouses
POST   /api/v1/warehouses        # Create warehouse
GET    /api/v1/warehouses/:id/locations  # Location hierarchy

POST   /api/v1/stock/grn         # Goods Receipt Note
POST   /api/v1/stock/issue       # Stock Issue (SRF)
POST   /api/v1/stock/transfer    # Inter-warehouse transfer
POST   /api/v1/stock/adjust      # Stock adjustment
GET    /api/v1/stock/levels      # Current stock levels
GET    /api/v1/stock/movements   # Stock movement ledger

GET    /api/v1/qa/inspections    # QA inspections
POST   /api/v1/qa/inspect        # Perform inspection

GET    /api/v1/assets            # Asset register
POST   /api/v1/assets/transfer   # Asset custody transfer

GET    /api/v1/reports/stock-status
GET    /api/v1/reports/valuation
GET    /api/v1/reports/expiry

POST   /api/v1/sync/push         # Push offline events
GET    /api/v1/sync/pull         # Pull latest state
```

## 5. Data Model Overview

**Master Data:** `organizations`, `org_levels` (L1–L4), `programs`, `departments`, `product_categories`, `products`, `units_of_measure`, `warehouses`, `locations`

**Transactional Data:** `stock_levels`, `stock_movements`, `goods_receipts`, `stock_requests`, `stock_transfers`, `stock_adjustments`, `qa_inspections`, `asset_transfers`, `distributions`

**Reference Data:** status enums, UoM codes, document types

**Configuration Data:** `roles`, `permissions`, `role_permissions`, `workflows`, `notification_rules`

Key design:
- `stock_movements` is append-only — `stock_levels` is derived from it
- Every table with tenant data has `org_id` + RLS policy
- UUIDv7 for time-ordered primary keys (cluster-friendly)
- Audit columns on every table

## 6. Multi-Tenancy & Security

- RLS on every table using `app.current_org_id` session variable set per-request from JWT
- Role × Scope × Action matrix (R01–R09) enforced in Gin middleware
- Audit interceptor on every write (actor, timestamp, before/after, IP)
- Transport: TLS everywhere

## 7. Offline-First Design

1. **Local DB:** Dexie.js (IndexedDB) on web, SQLite on mobile — mirrors server schema
2. **Writes go local-first:** GRN/issue/transfer/adjustment written to local event log immediately
3. **Sync:** on reconnect, push event log → server appends to canonical ledger → pull merged state
4. **Conflict handling:** append-only movements merge cleanly; true conflicts flagged for supervisor

## 8. Deployment Topologies

| Topology | Use Case |
|---|---|
| Docker Compose (single VPS) | Pilot, single org |
| Terraform + ECS/EKS | Multi-country, HA |
| Free-tier hybrid (Vercel + Neon + R2) | Small NGO |

## 9. CI/CD Pipeline

```
PR → GitHub Actions: go vet + test, pnpm lint + typecheck + build
  → merge to main → build Docker images → push to GHCR
  → auto-deploy staging → manual approval → deploy production
```
