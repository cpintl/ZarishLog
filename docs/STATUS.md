# ZarishLog — Project Status Dashboard

**Generated:** 2026-07-16
**Stack:** Go 1.26 + Gin 1.12 · PostgreSQL 18 + sqlc + sqlx · Next.js 15 + React 19 PWA · Keycloak 26 · Docker Compose · Terraform
**Repository:** `github.com/cpintl/zarishlog` (monorepo)

---

## 1. Phase Completion Table

| Phase | Status | Key Deliverables |
|---|---|---|
| Phase 0 — Foundation | ✅ Complete | Monorepo scaffold, Docker Compose (PostgreSQL 18, Redis 8, MinIO, Keycloak 26, Meilisearch), Makefile, GitHub CI, AGENTS.md, .env |
| Phase 1 — Database & Data Models | ✅ Complete | 73 tables, 137 sqlc queries, 3 migrations, 1161-line seed data, RLS on 37 tenant tables, config-as-CSV/JSON framework |
| Phase 2 — Go API Core | ❌ Not started | Go module init, config layer, DB pool, middleware (auth/RBAC/audit/tenant), health check, error handling, pagination |
| Phase 3 — Product/Catalogue Module | ❌ Not started | Category CRUD, Product CRUD, UoM CRUD, bulk CSV import, Meilisearch search, unit/integration tests |
| Phase 4 — Warehouse & Location Module | ❌ Not started | Warehouse CRUD, location hierarchy, location constraints, storage validation |
| Phase 5 — Stock & Inventory Module | ❌ Not started | GRN, SRF (FEFO), inter-warehouse transfer, adjustments, stock ledger, batch genealogy, barcode/QR |
| Phase 6 — Quality Assurance | ❌ Not started | QA inspection on receipt, pass/fail/quarantine, expiry monitoring, corrective action |
| Phase 7 — Distribution & Asset Management | ❌ Not started | Distribution forms, multi-program allocation, returns/disposal, asset lifecycle, depreciation |
| Phase 8 — Replenishment & Forecasting | ❌ Not started | AMC (3/6/12-month), buffer stock, reorder points, ML forecasting microservice (Prophet) |
| Phase 9 — User & Access Management | ❌ Not started | User CRUD + role assignment, permission matrix, password reset, MFA, audit log |
| Phase 10 — Offline-First & PWA | ❌ Not started | Dexie.js IndexedDB, Workbox service worker, Background Sync, conflict resolution |
| Phase 11 — Reporting & Analytics | ❌ Not started | Metabase dashboards, stock turnover, expiry dashboard, donor compliance reports |
| Phase 12 — Deployment & Infrastructure | ❌ Not started | Docker multi-stage, Terraform (VPC/RDS/ECS), GitHub Actions deploy, monitoring, load testing |

---

## 2. Phase 1 Detailed Delivery Report

### 2.1 Migration Summary

| Migration | Tables | Lines | Purpose |
|---|---|---|---|
| `001_initial_schema.sql` | 26 | 520 | Core schema: org, master data, warehouse, stock, transactions, QA, assets, RLS framework |
| `002_extended_schema.sql` | 46 | 816 | Extended: packaging, procurement, QA checklists, distribution, returns, disposals, physical count, replenishment, alerts, audit, sync, reports |
| `003_product_icons_warehouse_maps.sql` | 1 | 195 | Dosage forms lookup (22 icon types), extended product/warehouse/user/org fields, QA scoring, stock movement audit fields |
| **Total** | **73** | **1531** | — |

### 2.2 Seed Data Inventory

| Entity | Count | Details |
|---|---|---|
| Organizations | 1 | CPI (Center for Peace and Integrity) |
| Org Levels (L1-L4) | 8 | Global HQ, 2 Country Offices, 3 Project Offices, 2 Program Sites |
| Programs | 10 | H&N, WASH, LVL, EDU, EPRR, R&L, PROT, SHELTER, NFI, LOGS |
| Departments | 8 | H&N, WASH, LOG, PROC, FIN, HR, PROG, MEAL |
| Roles | 12 | R01 Global Admin through R12 Auditor |
| Permissions | 17 modules | 63 module/action pairs across products, stock, QA, assets, users, reports, suppliers, procurement, distributions, returns, disposals, stock counts, alerts, sync, audit |
| Users | 8 | System Admin, Country Rep, Warehouse Manager, Storekeeper, Dept Manager, Field Worker, QA Officer, Logistics Admin |
| Warehouses | 4 | CXB-CWH (central), CXB-SWH-1 (sub), CXB-TRN (transit), CXB-COLD (cold chain) |
| Locations | 17 | 6 CXB-CWH + 3 CXB-SWH-1 + 4 CXB-TRN + 4 CXB-COLD |
| Product Categories | 13 | Medicines & Drugs, Medical Supplies, Medical Equipment, Nutrition, WASH, Office & Admin, Lab & Diagnostics, Cold Chain, Shelter & NFI, Comms & IT, Vehicles, PPE, Cleaning & Hygiene |
| Products | 40 | Drugs (5), Medical Supplies (6), Equipment (5), Nutrition (3), WASH (6), Office/Admin/Shelter (15) |
| Units of Measure | 23 | EA, BX, CTN, KG, L, PK, RL, TU, BT, VL, AMP, SA, SR, TAB, CAP, etc. |
| Batches | 8 | Sample batches with expiry dates across products |
| Stock Levels | 11 | Initial stock positions across warehouses |
| Stock Movements | 4 | Sample receipt/issue/adjustment ledger entries |
| GRNs | 1 | Goods receipt with 2 line items |
| Stock Issues | 2 | Stock requests with multiple line items |
| Adjustment Reason Codes | 8 | Damage, expiry, theft, count variance, breakage, donation, quality failure, transfer loss |
| Disposal Methods | 6 | Incineration, landfill, recycling, chemical treatment, secure disposal, return to supplier |
| QA Checklist Templates | 3 | General inspection, cold chain, hazardous materials |
| QA Checklist Items | 14 | 7 general + 4 cold chain + 3 hazardous items |
| Alert Configurations | 6 | Expiry (30/60/90d), low stock, overstock, sleeping stock, QA hold, approval pending |
| Reorder Recommendations | 2 | Sample recommendations |
| **Total seed SQL** | **1161 lines** | 74 INSERT statements across 28 tables |

### 2.3 sqlc Query Inventory

| Query File | Queries | Key Operations |
|---|---|---|
| `transactions.sql` | 18 | GRN + SRF + transfer + adjustment CRUD, line items, document lookups |
| `organizations.sql` | 11 | Org, org_levels, programs, departments, function CRUD with hierarchy |
| `stock.sql` | 11 | Stock levels (by product/warehouse/batch), movements (by type/date/ref), FEFO peek |
| `procurement.sql` | 10 | Suppliers, contracts, purchase orders, PO line items CRUD |
| `products.sql` | 10 | Product CRUD, SKU lookup, category listing, substitutes, attachments |
| `users.sql` | 9 | User CRUD, email lookup, role assignment, preferences, sessions |
| `qa.sql` | 9 | Inspections CRUD, checklist templates/items, results, dispositions |
| `assets.sql` | 9 | Asset CRUD, custody changes, maintenance, depreciation schedule |
| `replenishment.sql` | 8 | AMC calculations, reorder recommendations, forecast results |
| `warehouses.sql` | 8 | Warehouse CRUD, location hierarchy, constraints, documents |
| `alerts.sql` | 7 | Alert configs CRUD, generated alerts, recipients, acknowledge |
| `categories.sql` | 6 | Category CRUD with parent hierarchy, subtree listing |
| `distribution.sql` | 6 | Distribution CRUD, line items, beneficiary records |
| `reports.sql` | 6 | Report definitions, schedules, parameter CRUD |
| `sync.sql` | 5 | Sync log, sync conflicts, conflict resolution |
| `audit.sql` | 4 | Audit log query, data change log, export |
| **Total** | **137** | 16 files, typed Go generation via sqlc |

### 2.4 Config Templates Inventory

| Category | Count | Files |
|---|---|---|
| CSV Import Templates | 12 | `asset_registration`, `disposal_form`, `distribution`, `goods_receipt`, `inspection_checklist`, `organization_import`, `product_import`, `stock_adjustment`, `stock_count`, `stock_issue`, `stock_transfer`, `user_import` |
| JSON Form Templates | 8 | `disposal_form`, `distribution_form`, `goods_receipt_form`, `product_registration_form`, `stock_count_form`, `stock_issue_form`, `user_registration_form`, `warehouse_registration_form` |
| README | 1 | Form field reference documentation |
| **Total** | **21** | `config/templates/` |

### 2.5 Config Metadata Files

| File | Format | Content |
|---|---|---|
| `master_product_list.csv` | CSV | 40-product master catalogue with SKU, category, UoM, type, tracking flags |
| `organization.csv` | CSV | 8-entry L1-L4 org hierarchy |
| `programs.csv` | CSV | 6 thematic program areas |
| `uom.csv` | CSV | 23 units of measure across 6 categories |
| `warehouse.json` | JSON | 4-warehouse config with 17 locations, lat/lng, maps URLs, facility metadata |
| `roles.md` | Markdown | Human-readable R01-R09 role definitions + permission matrix |

### 2.6 Extended Schema Features

| Feature | Count | Details |
|---|---|---|
| Product Dosage Forms | 22 types | TAB, CAP, INJ, SYR, CRM, ONT, EYE, EAR, NAS, INH, SUP, PAT, SOL, PWD, GRN, SPR, LOT, GEL, WAF, IMP, TAB-EFF, SUS |
| Product Extended Fields | 6 | `dosage_form_code`, `generic_name`, `brand_name`, `reference_urls` (JSONB), `storage_conditions`, `reorder_formula` (JSONB) |
| Warehouse Geo/Contact | 9 fields | `latitude`, `longitude`, `google_maps_url`, `contact_phone`, `operating_hours`, `has_generator`, `has_cctv`, `has_fire_system`, `security_guard` |
| Organization Branding | 6 fields | `logo_url`, `website`, `country`, `default_currency`, `timezone`, `date_format` |
| User Profile | 7 fields | `phone`, `job_title`, `department_id`, `avatar_url`, `signature_url`, `last_login_at`, `password_changed_at` |
| Stock Movement Audit | 3 fields | `source_document_url`, `is_offline_sync`, `device_id` |
| QA Scoring | 2 fields | `checklist_template_id`, `overall_score` |
| RLS-Protected Tables | 37 | All tenant-scoped tables have `org_id` isolation policy |

### 2.7 Database Indexes

| Category | Count | Key Indexes |
|---|---|---|
| Primary/B-tree | ~30 | `idx_products_org_id`, `idx_stock_levels_product_warehouse`, `idx_batches_expiry_date`, `idx_stock_movements_created`, `idx_alerts_org_type`, etc. |
| Composite (org_id + created_at) | 10 | Purchase orders, distributions, returns, disposals, stock counts, snapshots, suppliers, audit log, data change log, sync log |
| GIN (JSONB) | 2 | `idx_audit_log_changes`, `idx_audit_log_changes_path` |
| Spatial | 1 | `idx_warehouses_lat_lng` |

---

## 3. All Database Tables (73 tables, 16 domains)

### Domain 1: Organization & Users (12 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 1 | `organizations` | Tenant root | Top-level tenant (NGO/INGO) |
| 2 | `org_levels` | Tenant | L1 Global -> L2 Country -> L3 Project -> L4 Site hierarchy |
| 3 | `programs` | Tenant | Thematic program areas (H&N, WASH, EDU, etc.) |
| 4 | `departments` | Tenant | Organizational departments with parent hierarchy |
| 5 | `functions` | Tenant | Roles/functions within a department |
| 6 | `users` | Tenant | User accounts with role + org level assignment |
| 7 | `roles` | Reference | R01-R12 role definitions with access level |
| 8 | `permissions` | Reference | Module/action permission pairs |
| 9 | `role_permissions` | Junction | Many-to-many role <-> permission assignments |
| 10 | `user_role_assignments` | Tenant | Granular user-role scoping to org levels |
| 11 | `user_sessions` | Tenant | Auth session tokens with expiry |
| 12 | `user_preferences` | Tenant | Language, theme, timezone, notification prefs |

### Domain 2: Master Data & Product Catalogue (7 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 13 | `units_of_measure` | Reference | UoM catalog (EA, BX, KG, L, etc.) by category |
| 14 | `product_categories` | Tenant | Hierarchical product categories with UNSPSC/ECLASS codes |
| 15 | `products` | Tenant | Item master (SKU, GTIN, batch/serial tracking, storage flags) |
| 16 | `product_packaging` | Tenant | Multi-level packaging hierarchy with barcode, dimensions, weight |
| 17 | `product_substitutes` | Junction | Therapeutic/generic/brand/equivalency substitutions |
| 18 | `product_attachments` | Tenant | SOPs, photos, MSDS documents per product |
| 19 | `dosage_forms` | Reference | 22 drug dosage forms with emoji icons and storage rules |

### Domain 3: Procurement & Suppliers (4 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 20 | `suppliers` | Tenant | Vendor/supplier registry with contact and tax info |
| 21 | `supplier_contracts` | Tenant | Contract terms, discounts, currency, exclusivity |
| 22 | `purchase_orders` | Tenant | PO header with approval workflow (pending/approved/rejected) |
| 23 | `po_line_items` | Tenant | PO detail (product, qty ordered/received, unit price, line total) |

### Domain 4: Warehouse & Locations (4 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 24 | `warehouses` | Tenant | Central, sub-warehouse, transit, quarantine types with geo/contact |
| 25 | `locations` | Tenant | Hierarchical zone/rack/bin/shelf/area with cold/hazard/secure flags |
| 26 | `location_constraints` | Tenant | Temperature, humidity, capacity, pharma/food grade constraints |
| 27 | `warehouse_documents` | Tenant | Licenses, permits, certificates with expiry tracking |

### Domain 5: Stock & Batches (4 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 28 | `batches` | Tenant | Batch/lot/serial records with expiry and manufacturer |
| 29 | `stock_levels` | Tenant | Current positions (product x warehouse x location x batch) |
| 30 | `stock_movements` | Tenant | Append-only ledger (receipt, issue, transfer, adjustment, return, disposal) |
| 31 | `stock_snapshots` | Tenant | Periodic snapshots for AMC calculation (daily/weekly/monthly) |

### Domain 6: Goods Receipt & Issue (9 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 32 | `goods_receipts` | Tenant | GRN header (supplier, PO ref, received by) |
| 33 | `grn_line_items` | Tenant | GRN detail (product, batch, expiry, quantity, unit cost) |
| 34 | `stock_issues` | Tenant | SRF header (requested by, approved by, program, department) |
| 35 | `issue_line_items` | Tenant | SRF detail (product, batch, quantity, unit cost) |
| 36 | `stock_transfers` | Tenant | Inter-warehouse transfer header |
| 37 | `transfer_line_items` | Tenant | Transfer detail (product, batch, quantity) |
| 38 | `stock_adjustments` | Tenant | Adjustment header (reason code, status) |
| 39 | `adjustment_reason_codes` | Reference | Reason catalog (damage, expiry, theft, breakage, etc.) |
| 40 | `adjustment_line_items` | Tenant | Adjustment detail (expected vs actual, variance, reason) |

### Domain 7: Quality Assurance (5 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 41 | `qa_inspections` | Tenant | Inspection record (product, batch, result, score) |
| 42 | `qa_checklist_templates` | Tenant | Inspection checklist definitions by category |
| 43 | `qa_checklist_items` | Tenant | Individual checklist questions with criticality and weight |
| 44 | `qa_checklist_results` | Tenant | Actual inspection answers with scores |
| 45 | `qa_dispositions` | Tenant | Pass/fail/quarantine/rework/partial outcomes with destination |

### Domain 8: Distribution (3 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 46 | `distributions` | Tenant | Distribution header (program, org level, beneficiary count) |
| 47 | `distribution_line_items` | Tenant | Distribution detail (product, batch, planned vs actual qty) |
| 48 | `distribution_beneficiaries` | Tenant | Beneficiary type/count records per distribution |

### Domain 9: Returns & Disposal (5 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 49 | `stock_returns` | Tenant | Return header (beneficiary/program/supplier/internal) |
| 50 | `return_line_items` | Tenant | Return detail (condition: unopened/damaged/expired) |
| 51 | `disposals` | Tenant | Disposal header (method, authorized by, witness) |
| 52 | `disposal_line_items` | Tenant | Disposal detail (product, batch, quantity) |
| 53 | `disposal_methods` | Reference | Incineration, landfill, recycling, chemical treatment, etc. |

### Domain 10: Physical Count (3 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 54 | `stock_counts` | Tenant | Physical count header (full/cycle/spot/annual) |
| 55 | `count_line_items` | Tenant | Count detail (expected vs counted, variance %, status) |
| 56 | `count_variance_reconciliation` | Tenant | Reconciliation: adjustment, recount, write-off, justified |

### Domain 11: Asset Management (5 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 57 | `assets` | Tenant | Asset register (tag, custodian, value, depreciation, lifecycle status) |
| 58 | `asset_custody_changes` | Tenant | Custody transfer history (from/to/changed by) |
| 59 | `asset_maintenance` | Tenant | Maintenance records (date, description, cost, next date) |
| 60 | `asset_depreciation_schedule` | Tenant | Period-by-period depreciation (method, book values) |
| 61 | `asset_attachments` | Tenant | Asset documents (purchase invoices, photos, certificates) |

### Domain 12: Replenishment & Forecasting (3 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 62 | `amc_calculations` | Tenant | 3/6/12-month rolling AMC with max consumption, std dev |
| 63 | `reorder_recommendations` | Tenant | Reorder/excess/critical recommendations with priority |
| 64 | `forecast_results` | Tenant | ML engine forecasts (date, value, confidence bounds, model version) |

### Domain 13: Alerts & Notifications (3 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 65 | `alert_configurations` | Tenant | Configurable alert rules (expiry, low/over/sleeping stock, QA, approval) |
| 66 | `alerts` | Tenant | Generated alerts (severity, title, message, ack/resolution tracking) |
| 67 | `alert_recipients` | Tenant | Per-config user notification channel assignments |

### Domain 14: Audit & Compliance (2 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 68 | `audit_log` | Tenant | Application-level audit trail (action, entity, changes JSONB, IP, UA) |
| 69 | `data_change_log` | Tenant | Row-level data change log for sync conflict resolution |

### Domain 15: Offline Sync (2 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 70 | `sync_log` | Tenant | Push/pull/full sync history (records, errors, status) |
| 71 | `sync_conflicts` | Tenant | Conflict records (server vs client versions, resolution strategy) |

### Domain 16: Reports & Scheduling (2 tables)

| # | Table | Type | Description |
|---|---|---|---|
| 72 | `report_definitions` | Tenant | Report catalog (SQL query, parameters, output formats, scheduling) |
| 73 | `report_schedules` | Tenant | Cron-based report schedules with recipients and format |

### Summary

| Metric | Count |
|---|---|
| Tables | 73 (26 initial + 46 extended + 1 dosage forms) |
| Tenant-scoped (RLS) | 37 with `org_id` enforced |
| Reference/lookup | 8 shared across tenants |
| Junction/child | 28 inherited via FK |
| Rows of seed data | 1,161 lines of SQL, 74 INSERT statements |
| Custom ENUM types | 8 (`uom_category`, `item_type`, `entity_status`, `warehouse_type`, `location_type`, `movement_type`, `doc_type`, `stock_status`) |
| Database functions | 3 (`uuid_generate_v7`, `app.current_org_id`, `app.current_user_id`, `rls_org_policy`) |

---

## 4. File Inventory

### `packages/data-models/sql/migrations/` (3 files)

| File | Tables | Lines |
|---|---|---|
| `001_initial_schema.sql` | 26 | 520 |
| `002_extended_schema.sql` | 46 | 816 |
| `003_product_icons_warehouse_maps.sql` | 1 (alterations) | 195 |
| **Total** | **73** | **1,531** |

### `packages/data-models/sql/queries/` (16 files, 137 queries)

| File | Queries | Lines |
|---|---|---|
| `transactions.sql` | 18 | 110 |
| `organizations.sql` | 11 | 55 |
| `stock.sql` | 11 | 57 |
| `products.sql` | 10 | 69 |
| `procurement.sql` | 10 | 62 |
| `users.sql` | 9 | 46 |
| `assets.sql` | 9 | 54 |
| `qa.sql` | 9 | 47 |
| `warehouses.sql` | 8 | 44 |
| `replenishment.sql` | 8 | 44 |
| `alerts.sql` | 7 | 41 |
| `categories.sql` | 6 | 30 |
| `distribution.sql` | 6 | 37 |
| `reports.sql` | 6 | 33 |
| `sync.sql` | 5 | 26 |
| `audit.sql` | 4 | 27 |
| **Total** | **137** | **822** |

### `config/templates/` (21 files)

**CSV Templates (12):**
- `asset_registration_template.xlsx.csv`
- `disposal_form_template.xlsx.csv`
- `distribution_template.xlsx.csv`
- `goods_receipt_template.xlsx.csv`
- `inspection_checklist_template.xlsx.csv`
- `organization_import_template.xlsx.csv`
- `product_import_template.xlsx.csv`
- `stock_adjustment_template.xlsx.csv`
- `stock_count_template.xlsx.csv`
- `stock_issue_template.xlsx.csv`
- `stock_transfer_template.xlsx.csv`
- `user_import_template.xlsx.csv`

**JSON Form Templates (8):**
- `disposal_form.json`
- `distribution_form.json`
- `goods_receipt_form.json`
- `product_registration_form.json`
- `stock_count_form.json`
- `stock_issue_form.json`
- `user_registration_form.json`
- `warehouse_registration_form.json`

**Documentation (1):**
- `README.md` (284 lines — form field reference)

### `config/metadata/` (4 CSV files)

| File | Rows | Content |
|---|---|---|
| `master_product_list.csv` | 40 products | SKU, name, category, UoM, item type, tracking flags |
| `organization.csv` | 8 org levels | L1-L4 hierarchy with parent codes |
| `programs.csv` | 6 programs | Thematic area codes and descriptions |
| `uom.csv` | 23 units | Name, abbreviation, category |

### `config/location/` (1 JSON file)

| File | Lines | Content |
|---|---|---|
| `warehouse.json` | 145 | 4 warehouses with 17 locations, geo-coordinates, maps URLs |

---

## 5. Next Steps — Phase 2 Roadmap

Phase 2 will build the Go API core that connects the database layer to HTTP endpoints:

- **Go module init** — `apps/api/` with `go.mod`, `cmd/server/main.go`, standard project layout (`internal/handler/`, `internal/service/`, `internal/repository/`)
- **Configuration layer** — Envconfig/viper to load `.env`, DB DSN, Keycloak URLs, Meilisearch host
- **Database pool** — sqlx connection pool with health check, migration runner integration
- **Middleware stack** — OIDC/JWT validation (Keycloak), RBAC enforcement (check org_id + role), tenant context injection, audit logging, structured JSON error handling
- **Health endpoint** — `GET /health` with DB/ping, Redis/ping, uptime
- **Request validation** — Gin bindings with custom validators for UUIDv7, ISO dates, enum constraints
- **Pagination** — Cursor-based pagination helper for list endpoints
- **Base CRUD pattern** — Establish the handler -> service -> repository three-layer pattern using all 16 query files

Ready to begin when the Go development environment is configured.

---

*Document automatically generated from source code analysis. Update as the project progresses through Phases 2-12.*
