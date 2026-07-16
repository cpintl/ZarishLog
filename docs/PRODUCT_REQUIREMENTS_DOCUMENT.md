# ZarishLog — Product Requirements Document (PRD)

**Version:** 1.0 · **Status:** Draft for build · **Owner:** ZarishLog Program

---

## 1. Purpose & Vision

ZarishLog is a unified, offline-first, multi-tenant SaaS platform for humanitarian **logistics, supply chain, and asset management**. It operates at the intersection of **Disaster Management**, **Humanitarian Logistics**, and **Sustainability** (Sustainable Humanitarian Relief Logistics — SHRL principles), giving NGOs/INGOs one system, instead of spreadsheets and disconnected tools, to run procurement-to-distribution operations across a multi-level (L1–L4) organizational structure.

**Vision statement:** Any humanitarian organization, regardless of budget, should be able to deploy a production-grade, standards-aligned inventory and asset system in days, using only open-source software and free-tier infrastructure.

## 2. Problem Statement

Consolidated from all source material reviewed (ZarishStock TODOs, ZarishLog architectural drafts, Master Catalogue Definition, Product/Item Master research, role & warehouse data):

1. **Data fragmentation** — item, warehouse, and program identifiers are inconsistent across spreadsheets (e.g., `MASTER ASSETS INVENTORY.xlsx` mixes size/color/condition into one free-text description field).
2. **No canonical catalogue** — 22,000+ items exist across multiple partial catalogues with overlapping/duplicate SKUs and inconsistent categories.
3. **No cross-level visibility** — L2 Country Office cannot see L3/L4 stock in real time.
4. **Manual, error-prone compliance** — FEFO, expiry, batch/serial tracking, and QA holds are tracked on paper or in disconnected sheets.
5. **No offline capability** — most commercial WMS/ERP tools require constant connectivity, which field/camp sites cannot guarantee.
6. **Weak access control** — no consistent role-based permission model across organizational levels (roles R01–R09 exist informally but aren't enforced by any system).
7. **No asset/inventory distinction** — durable assets (equipment, vehicles, IT) and consumable inventory (drugs, supplies) are managed the same way, when they need different lifecycles (depreciation vs. consumption).

## 3. Goals & Non-Goals

### Goals
- Single canonical **Master Catalogue** (items, categories, UoM, statuses) covering drugs, medical supplies, non-medical supplies, equipment, and assets.
- Full **L1→L2→L3→L4** organizational hierarchy with per-tenant data isolation.
- End-to-end stock lifecycle: procurement request → GRN → putaway → stock holding → issue/transfer/adjustment → distribution → returns/disposal.
- **FEFO/FIFO enforcement**, batch/serial tracking, expiry alerts.
- **QA/inspection workflow** with quarantine and disposition.
- **AMC/FMC-based** replenishment and reorder-point calculation.
- **Offline-first** operation at warehouse and field level, with conflict-aware sync.
- **RBAC** matching the 9 defined roles (R01–R09) and organizational scope.
- **Reporting/BI** for stock status, movement history, valuation, utilization, and donor compliance.
- Runs entirely on **open-source + free-tier** infrastructure, self-hostable or cloud-hosted.

### Non-Goals (v1)
- Full financial general ledger / accounting (integrate with existing Finance systems via API instead of rebuilding).
- Full HR/payroll.
- Native donor-grant-management module (MEAL integration point only, not full MEAL suite).
- Complex route optimization/fleet telematics (fleet *records* are in scope; live GPS tracking is a future phase).

## 4. Personas & Roles

Derived directly from the organization's existing role table; ZarishLog formalizes and enforces these digitally.

| Role ID | Role Name | Scope | Capability |
|---|---|---|---|
| R01 | GLOBAL_ADMIN | Global (L1) | Full system administration, all tenants |
| R02 | COUNTRY_REP | Country (L2) | View-all, reporting only, no ops |
| R03 | THEME_MANAGER | Country (L2), theme-scoped | View-all within one program theme, reporting |
| R04 | WAREHOUSE_OFFICER | Central Warehouse (L3) | Full central warehouse operations |
| R05 | WAREHOUSE_STOREKEEPER | Central Warehouse (L3) | Stock operations only (receive, issue, count) |
| R06 | ADMIN_LOG_OFFICER | Project Office (L3) | Office asset/logistics stock management |
| R07 | DEPT_MANAGER | Department, all levels | Budget/flow approval authority |
| R08 | DEPT_COORDINATOR | Department (L4) | Validates stock flow at sub-warehouse level |
| R09 | DEPT_OFFICER | Sub-Warehouse (L4) | Day-to-day sub-warehouse operations |

Each role maps to a concrete permission set (module × action matrix) enforced at the API layer via row-level security + RBAC claims in the JWT, not just hidden UI — this closes the "weak access control" gap identified above.

## 5. Organizational Hierarchy (Master Data)

```
Organization (L1) — root tenant, e.g. CPI, YPSA-CPI
└── Country Office (L2)
    └── Project Office (L3)
        ├── Central Warehouse (CWH)
        │   ├── Receiving Area
        │   ├── General Storage (by category)
        │   ├── Program Storage (by program)
        │   ├── QA / Quarantine Area
        │   ├── Dispatch Preparation Area
        │   └── Dispatch Area
        └── Program Site (L4)
            ├── Sub-Warehouse (SWH, by program)
            └── Program units: H&N, WASH&CE, Livelihood, Education, EPRR, R&L
```

Every entity in this hierarchy is a first-class, addressable master-data record with a stable ID (`L1-ZS-HQ`, `L3-CXB-CW`, etc., following the pattern already established in source data), not a free-text field.

## 6. Functional Requirements by Module

### 6.1 Master Data Management
- FR-1.1: Maintain canonical **Item/Product Master** with identification (SKU, barcode/GTIN, alt codes), classification (category/sub-category, UNSPSC/ECLASS-aligned where applicable), UoM & packaging, tracking flags (batch/serial/expiry/hazardous/cold-chain), and inventory parameters (lead time, min/max, reorder point).
- FR-1.2: Support 8 item types: Drugs, Medical Supplies, Equipment, Instruments, Materials, Vaccines, Nutrition, Lab Reagents — plus non-medical Assets and generic Consumables.
- FR-1.3: Bulk import/export of catalogue via CSV/XLSX with validation and duplicate detection (resolves the fragmentation problem directly).
- FR-1.4: Maintain Organization, Program, Department, Function hierarchy as versioned master data.
- FR-1.5: Maintain Warehouse & Location hierarchy (zone/rack/bin) with storage-condition constraints (ambient, cold chain, hazardous, secure).

### 6.2 Procurement & Goods Receipt
- FR-2.1: Purchase/stock request creation with approval workflow (maps to Dept Manager/Coordinator roles).
- FR-2.2: Goods Receipt Note (GRN) capturing supplier, batch/lot, expiry, quantity, condition, and inspection outcome.
- FR-2.3: Discrepancy tracking (ordered vs. received vs. accepted).

### 6.3 Warehouse & Inventory Operations
- FR-3.1: Real-time stock levels per item × batch × location × warehouse.
- FR-3.2: Goods issue (SRF), inter-warehouse transfer, and stock adjustment with mandatory reason codes and approval thresholds.
- FR-3.3: FEFO-enforced picking suggestions; system blocks/warns on FEFO violations.
- FR-3.4: Batch and serial number tracking with full genealogy (receipt → storage → issue → destination).
- FR-3.5: Barcode/QR code generation and mobile scanning for items, locations, and batches.
- FR-3.6: Physical stock count (cycle count / full count) with variance reconciliation.

### 6.4 Quality Assurance
- FR-4.1: Inspection checklist on receipt; pass/fail/quarantine disposition.
- FR-4.2: Quarantine area management with restricted access.
- FR-4.3: Expiry monitoring with configurable alert thresholds (e.g., 90/60/30 days).
- FR-4.4: Corrective action / disposal record linked to affected batches.

### 6.5 Distribution & Program Allocation
- FR-5.1: Distribution/delivery forms with program and (aggregate, non-PII) beneficiary-count tracking.
- FR-5.2: Multi-program allocation across H&N, SD, WASH&CE, Livelihood, Education, EPRR, R&L.
- FR-5.3: Returns and disposal workflow with financial/quantity adjustment.

### 6.6 Replenishment & Forecasting
- FR-6.1: AMC (Average Monthly Consumption) calculation from historical issue data (3/6/12-month rolling window).
- FR-6.2: Buffer/security stock and reorder-point calculation from lead time + AMC.
- FR-6.3: FMC (Forecasted Monthly Consumption) for preparedness/contingency planning.
- FR-6.4: Automated low-stock, overstock, and "sleeping stock" (no movement >6 months) alerts.

### 6.7 Asset Management (distinct from consumable inventory)
- FR-7.1: Asset register with acquisition data, custodian, location, and depreciation schedule.
- FR-7.2: Asset lifecycle states (in-use, in-storage, under-maintenance, disposed).
- FR-7.3: Asset transfer/custody-change workflow with digital sign-off.
- FR-7.4: Maintenance/service history log.

### 6.8 Reporting & Analytics
- FR-8.1: Inventory status, stock valuation, movement history, and audit-trail reports.
- FR-8.2: Warehouse utilization and stock-turnover analysis.
- FR-8.3: Low-stock/overstock/expiry dashboards.
- FR-8.4: Export to PDF/CSV; scheduled email delivery of key reports.
- FR-8.5: Donor/compliance report templates.

### 6.9 User & Access Management
- FR-9.1: User CRUD with role assignment per organizational scope (not just global role).
- FR-9.2: Permission management (module × action matrix) per role.
- FR-9.3: Full activity/audit log (who changed what, when, from where).
- FR-9.4: Self-service password reset, MFA support.

### 6.10 Notifications
- FR-10.1: In-app and email notifications for low stock, expiry, QA failure, unauthorized adjustment, pending approvals.
- FR-10.2: Configurable notification rules per role/scope.

### 6.11 Offline & Mobile
- FR-11.1: Full offline operation (stock view, GRN, issue, adjustment, scanning) at warehouse and field level.
- FR-11.2: Automatic, conflict-aware sync when connectivity resumes; visible sync-status indicator.
- FR-11.3: Native-feeling mobile apps (Android/iOS) via React Native/Expo, sharing business logic with web.

## 7. Non-Functional Requirements

| Category | Requirement |
|---|---|
| Performance | Sub-200ms API response for standard queries at 22,000+ item catalogue scale; list views paginated/virtualized |
| Availability | Self-hosted target 99.5%+; graceful degradation to offline mode on network loss |
| Security | OIDC/JWT auth, RBAC + row-level tenant isolation, encryption at rest and in transit, audit logging of all writes |
| Multi-tenancy | Complete logical data isolation per Organization; no cross-tenant data leakage |
| Portability | Runs on Linux/macOS/Windows dev machines; deployable self-host, cloud, or hybrid (including Google Workspace-adjacent hosting) |
| Localization | GMT+6 default timezone; `DD MMMM YYYY` date format, 12-hour AM/PM time format; extensible to other locales |
| Accessibility | WCAG 2.1 AA target for web UI |
| Cost | Zero licensing cost; infrastructure runnable entirely within free tiers for pilot/small deployments |
| Scalability | Horizontal scale path from single Docker Compose host → k3s cluster without re-architecture |

## 8. Data Classification

| Type | Examples | Characteristics |
|---|---|---|
| **Master data** | Items, Warehouses, Locations, Organizations, Users | Relatively static, versioned, system of record |
| **Transactional data** | GRNs, Issues, Transfers, Adjustments, Stock Movements | High volume, append-mostly, immutable once posted |
| **Reference data** | UoM codes, statuses, category lists, role definitions | Small, controlled vocabularies |
| **Configuration data** | Approval workflows, notification rules, permission matrices | Tenant-configurable, drives system behavior |

Each field in the schema is further tagged **mandatory/optional**, **static/dynamic**, and **system-controlled/user-entered** — full detail in the database schema (`ARCHITECTURE.md` §4 and the underlying `SCHEMA_DESIGN.md` working document).

## 9. Identified Gaps & Resolutions

| Gap found in source material | Resolution adopted in this PRD |
|---|---|
| Multiple, overlapping master catalogues (ZarishLog Master Catalogue, Master Catalogue Definition, 1_Product_Item_Master) with inconsistent field sets | Single reconciled Item Master schema (§6.1) is the system of record; all prior catalogues are treated as import sources, deduplicated by name+category fuzzy match at import time |
| Free-text `ITEM DESCRIPTION` fields mixing size/color/condition | Deconstructed into discrete attributes (Specification, Brand, Condition, etc.) enforced at entry |
| No enforced RBAC despite defined roles | Roles R01–R09 formally mapped to a permission matrix enforced server-side |
| Assets and consumables modeled identically in prior drafts | Split into distinct Asset and Inventory-Item domains with different lifecycles (§6.7 vs §6.1–6.3) |
| Undefined conflict resolution for offline sync | Last-write-wins for reference data; append-only event log + server-side reconciliation for stock transactions (no silent overwrite of quantities) |
| Ambiguous "Medicine" vs "Drug" terminology | Standardized to "Drug" per ZarishLog terminology standard |
| No defined notification delivery mechanism | Email (via free-tier transactional email, e.g., Resend/Postmark free tier) + in-app center |
| Title-only source documents with no content (Stock Management Manual, Asset Management Guideline) | Schema designed to be extensible so these policies can be encoded later without structural rework |

## 10. Success Metrics

- 100% of active SKUs represented in the single Master Catalogue with no duplicate codes.
- Real-time stock visibility available at L2 within <5 seconds of an L3/L4 transaction (when online).
- Offline transactions sync with zero data loss on reconnect.
- FEFO compliance rate (issues matching earliest-expiry batch) trackable and reportable.
- 100% of stock adjustments carry a reason code and are attributable to a user.

## 11. Related Documents
- `BLUEPRINT.md` — phased delivery plan and monorepo build order
- `ARCHITECTURE.md` — technical architecture and infrastructure decisions
- `README.md` — project overview and quick start
