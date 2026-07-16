# ZarishLog — Product Requirements Document (PRD)

**Version:** 2.0 · **Status:** Build-ready · **Stack:** Go 1.26 / Gin 1.12 / PostgreSQL 18 / Next.js 15 / React 19

---

## 1. Purpose & Vision

ZarishLog is a unified, offline-first, multi-tenant SaaS platform for humanitarian **logistics, supply chain, and asset management**. It operates at the intersection of **Disaster Management**, **Humanitarian Logistics**, and **Sustainability** (Sustainable Humanitarian Relief Logistics — SHRL principles), giving NGOs/INGOs one system instead of spreadsheets and disconnected tools.

**Vision:** Any humanitarian organization, regardless of budget, should deploy a production-grade, standards-aligned inventory and asset system in days, using only open-source software and free-tier infrastructure.

## 2. Problem Statement

1. **Data fragmentation** — item, warehouse, and program identifiers inconsistent across spreadsheets
2. **No canonical catalogue** — 22,000+ items across multiple partial catalogues with duplicates
3. **No cross-level visibility** — L2 Country Office cannot see L3/L4 stock in real time
4. **Manual, error-prone compliance** — FEFO, expiry, batch/serial tracking, QA holds on paper
5. **No offline capability** — field/camp sites cannot guarantee connectivity
6. **Weak access control** — no consistent RBAC across organizational levels
7. **No asset/inventory distinction** — durable assets vs consumable inventory need different lifecycles

## 3. Goals & Non-Goals

### Goals
- Single canonical **Master Catalogue** covering drugs, medical supplies, non-medical supplies, equipment, assets
- Full **L1→L2→L3→L4** organizational hierarchy with per-tenant data isolation
- End-to-end stock lifecycle: procurement → GRN → putaway → holding → issue/transfer → distribution → disposal
- **FEFO/FIFO enforcement**, batch/serial tracking, expiry alerts
- **QA/inspection workflow** with quarantine and disposition
- **AMC/FMC-based** replenishment and reorder-point calculation
- **Offline-first** operation with conflict-aware sync (Dexie.js + Workbox + Background Sync)
- **RBAC** matching R01–R09 roles with organizational scope
- **Reporting/BI** for stock status, movement, valuation, donor compliance
- Runs entirely on **open-source + free-tier** infrastructure

### Non-Goals (v1)
- Full financial general ledger / accounting
- Full HR/payroll
- Native donor-grant-management module
- Complex route optimization / fleet telematics

## 4. Personas & Roles

| ID | Role Name | Scope | Capability |
|---|---|---|---|
| R01 | GLOBAL_ADMIN | Global (L1) | Full system administration |
| R02 | COUNTRY_REP | Country (L2) | View-all, reporting |
| R03 | THEME_MANAGER | Country (L2), theme-scoped | View-all within program theme |
| R04 | WAREHOUSE_OFFICER | Central Warehouse (L3) | Full warehouse operations |
| R05 | WAREHOUSE_STOREKEEPER | Central Warehouse (L3) | Stock operations only |
| R06 | ADMIN_LOG_OFFICER | Project Office (L3) | Office asset/logistics management |
| R07 | DEPT_MANAGER | Department | Budget/flow approval |
| R08 | DEPT_COORDINATOR | Department (L4) | Validates stock flow |
| R09 | DEPT_OFFICER | Sub-Warehouse (L4) | Day-to-day operations |

## 5. Functional Modules

### 5.1 Master Data Management
- Canonical Item/Product Master with SKU, barcode/GTIN, classification (UNSPSC/ECLASS-aligned), UoM, packaging, tracking flags, inventory parameters
- 8 item types: Drugs, Medical Supplies, Equipment, Instruments, Materials, Vaccines, Nutrition, Lab Reagents + Assets + Consumables
- Bulk CSV/XLSX import with validation and duplicate detection
- Organization, Program, Department, Function hierarchy
- Warehouse & Location hierarchy (zone/rack/bin) with storage-condition constraints

### 5.2 Procurement & Goods Receipt
- Stock request creation with approval workflow
- GRN capturing supplier, batch/lot, expiry, quantity, condition, inspection
- Discrepancy tracking (ordered vs received vs accepted)

### 5.3 Warehouse & Inventory Operations
- Real-time stock levels per item × batch × location × warehouse
- Goods issue (SRF), inter-warehouse transfer, stock adjustment with reason codes
- FEFO-enforced picking suggestions
- Batch and serial number genealogy
- Barcode/QR generation and mobile scanning
- Physical stock count (cycle count / full count) with variance reconciliation

### 5.4 Quality Assurance
- Inspection checklist on receipt; pass/fail/quarantine disposition
- Quarantine area management with restricted access
- Expiry monitoring with configurable alert thresholds
- Corrective action / disposal record

### 5.5 Distribution & Program Allocation
- Distribution/delivery forms with program and beneficiary-count tracking
- Multi-program allocation
- Returns and disposal workflow

### 5.6 Replenishment & Forecasting
- AMC calculation from historical data (3/6/12-month rolling)
- Buffer stock and reorder-point calculation
- FMC for preparedness/contingency planning
- AI/ML forecasting via Prophet
- Low-stock/overstock/sleeping-stock alerts

### 5.7 Asset Management
- Asset register with acquisition data, custodian, location, depreciation
- Lifecycle states: in-use, in-storage, under-maintenance, disposed
- Asset transfer/custody-change workflow with digital sign-off
- Maintenance/service history

### 5.8 User & Access Management
- User CRUD with role assignment per org scope
- Permission matrix (module × action) per role
- Full activity/audit log
- MFA support

### 5.9 Offline & Mobile
- Full offline operation (stock view, GRN, issue, adjustment, scanning)
- Automatic, conflict-aware sync on reconnect; sync-status indicator
- Native mobile apps (Android/iOS) via Expo

## 6. Tech Stack (Updated)

| Layer | Choice |
|---|---|
| Backend | Go 1.26 + Gin 1.12 |
| Database | PostgreSQL 18 + sqlc + sqlx |
| Frontend | Next.js 15 + React 19 PWA |
| Mobile | Expo/React Native |
| Auth | Keycloak 26 (OIDC/OAuth2) |
| Offline | Dexie.js + Workbox + Background Sync |
| Search | Meilisearch |
| BI | Metabase CE |
| Storage | MinIO (S3-compatible) |
| Infrastructure | Docker Compose, Terraform, GitHub Actions |
| ML Engine | Go microservice (Prophet) |

## 7. Success Metrics

- 100% of active SKUs in single Master Catalogue with no duplicates
- Real-time stock visibility at L2 within <5s of L3/L4 transaction (online)
- Offline transactions sync with zero data loss
- FEFO compliance rate trackable and reportable
- 100% of stock adjustments carry reason code and user attribution
