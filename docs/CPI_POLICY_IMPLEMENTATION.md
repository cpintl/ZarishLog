# CPI Bangladesh Medical Warehouse — Policy Implementation

Maps the compliance registers required by **"CPI BANGLADESH MEDICAL WAREHOUSE POLICY" (v1.0)** to the
ZarishLog schema, API, and tests. The policy governs warehouse, cold-chain, quality, and
distribution operations for Bangladesh-registered humanitarian medical supplies (DGDA / DNC).

Status: **Phase 1 implemented** — schema, sqlc queries, Go models/handlers, routes, and handler
tests are complete. Batch/temperature monitoring intake from connected devices and BI reporting are
planned follow-ups.

---

## 1. Policy Section → Implementation Traceability

| Policy § | Requirement | Table(s) | API routes |
|---|---|---|---|
| §7 | Warehouse & storage standards; physical inventory | (existing `warehouses`, `locations`) | existing |
| §8 | Regulatory & legal compliance — DGDA / DNC certificates, licenses, permits | `regulatory_approvals` | `/policy/regulatory-approvals` (+ `/expiring`) |
| §11.3 | Temperature monitoring & excursions | `temperature_monitoring_entries`, `temperature_excursions` | `/temperature-monitoring`, `/temperature-excursions` |
| §12.2 | Controlled drugs — register, reconciliation | `controlled_stock_register` | `/controlled-stock` |
| §13 | Donations — offer, line items, acceptance decision | `donations`, `donation_line_items` | `/donations` (+ `/:id/decision`) |
| §14.3 | Out-of-spec / returned goods — deviations, CAPA, training | `deviations`, `capa_actions`, `training_records` | `/policy/deviations`, `/policy/capa`, `/policy/training` |
| §14.4 | Mock recalls (annual traceability exercise) | `recalls`, `recall_line_items` | `/recalls` |
| §16 | Dispensing controlled products; short-expiry management | `stock_release_records`, `short_expiry_reviews` | `/stock-releases`, `/short-expiry-reviews` |
| §17.2 | Complaint handling | `complaints`, `complaint_severities` | `/complaints` |
| §17.3 | Recalls — actual + mock exercises | `recalls`, `recall_line_items`, `recall_types` | `/recalls` |
| §20 | Emergency preparedness / continuity | `emergency_plans` | `/emergency-plans` |
| §27 | Change control | `change_controls` | `/change-controls` |
| §10 / distribution | Dispatch waybills + proof of delivery | `dispatch_waybills`, `delivery_confirmations` | `/dispatch-waybills`, `/delivery-confirmations` |

---

## 2. Schema — Migration 007

File: `packages/data-models/sql/migrations/007_cpi_medical_policy_controls.sql`

### 2.1 Policy reference dictionaries (3, seeded)

| Table | Purpose | Seed values |
|---|---|---|
| `deviation_categories` | Deviations taxonomy with default risk level | DAMAGED_GOODS, TEMPERATURE_EXCURSION, MISSING_RECORDS, STOCK_DISCREPANCY, UNAUTHORIZED_ACCESS, FALSIFICATION, INCORRECT_PICKING, LATE_REPORTING, QUALIFICATION_FAILURE, CONTROLLED_SUBSTANCE, OTHER |
| `complaint_severities` | Complaint severity + immediate-escalation flag (§17.2) | CRITICAL, MAJOR, MINOR |
| `recall_types` | Actual vs mock recall exercises (§17.3) | ACTUAL, MOCK |

### 2.2 Policy registers (18)

| Table | Key fields & rules |
|---|---|
| `regulatory_approvals` | Authority (DGDA/DNC/NBR/CCI_E/Customs/…), license_type, issue/expiry dates, `renewal_lead_days`, status (draft/active/expiring_soon/expired/revoked/suspended) |
| `deviations` | Deviation number (auto `DEV-YYYY-####`), category, severity, occurred/reported time, containment, root cause, risk rating, escalation, status |
| `capa_actions` | Source (deviation/complaint/audit/excursion), corrective/preventive, root cause, action plan, effectiveness verification → close |
| `training_records` | Staff qualification/training log with validity period |
| `temperature_monitoring_entries` | Periodic temp/humidity readings (C / RH%) |
| `temperature_excursions` | Freeze/heat/transport/humidity events, min/max vs expected range, affected products & quantities, quarantine, disposition, quality/manager notification |
| `stock_release_records` | Released product (quarantine/regulatory/recall), evidence URL, decision, decided by |
| `controlled_stock_register` | Controlled-substance register: transaction type (receive/store/issue/return/dispose/reconcile), ref document, quantity & balance-after, received-from / issued-to, recipient signature, discrepancy |
| `donations` | Donor, offer/needs/recipient, approval, transport/customs/disposal flags, status, decision + certificate (§13) |
| `donation_line_items` | Product, batch, expiry (FEFO-sorted), quantity, UoM, shelf-life compliance, authorized |
| `complaints` | Severity (§17.2), category (product/process/packaging/washout/controlled/…), immediate escalation (§17.1), manufacturer + competent authority notification, status |
| `recalls` | Type ACTUAL/MOCK, product scope, batch numbers, reason, initiation/authorization, communications, disposition plan, status |
| `recall_line_items` | Recipient/location, issued vs recovered, **outstanding = issued − recovered** (computed at insert) |
| `dispatch_waybills` | Dispatch header: sender/recipient warehouses, carrier, vehicle, carton count, storage requirement, temperature-sensitive, condition, doc refs |
| `delivery_confirmations` | Proof of delivery: conformance checks (items/quantity/batch/expiry/condition/temp), discrepancies → auto-marks waybill `delivered` |
| `short_expiry_reviews` | Reviews for short-dated stock (3 / 6 / 12 month bands, action code, approved by) |
| `emergency_plans` | Scenario plans (cyclone/flood/fire/power/cold-chain failure — §20): demand assumptions, service level, prepositioned stock & rotation plans, alternate suppliers/warehouses/routes/power, paper-fallback records, staffing, communications, escalation, security |
| `change_controls` | Subject type (process/equipment/layout/software/supplier/contract/…) with impact assessment, risk level, quality-review requirement, approval, start/end dates, status |

### 2.3 Indexes & RLS

- Indexes on every register: `(org_id)` and either `(org_id, created_at)` or natural sort keys
  (e.g. `expiry_date`, `occurred_at`, `started_at`, `register_date`, `decision_date`).
- **RLS:** 16 registers via `app.rls_policy_expression()` (org-level); the two child line-item tables
  (`donation_line_items`, `recall_line_items`) resolve org through their parent via an `IN (SELECT …)`
  policy — RLS is not duplicated on the children.

---

## 3. API — `/api/v1` endpoints (all JWT-protected)

| Group | Routes | Roles |
|---|---|---|
| `/policy/regulatory-approvals` | POST, GET, GET `/expiring?days=`, GET `/:id` | admin, warehouse_manager, pharmacist, quality_officer |
| `/policy/deviations` | POST, GET, GET `/:id` | admin, warehouse_manager, pharmacist, quality_officer |
| `/policy/capa` | POST, GET, GET `/:id`, POST `/:id/effectiveness` | admin, warehouse_manager, pharmacist, quality_officer |
| `/policy/training` | POST, GET | admin, warehouse_manager, pharmacist, quality_officer |
| `/temperature-monitoring` | POST, GET | + logistics_officer |
| `/temperature-excursions` | POST, GET, GET `/:id`, POST `/:id/disposition` | admin, warehouse_manager, pharmacist, quality_officer |
| `/stock-releases` | POST, GET, GET `/:id` | admin, warehouse_manager, pharmacist, quality_officer |
| `/controlled-stock` | POST, GET, GET `/:id` | admin, warehouse_manager, pharmacist |
| `/donations` | POST (nested line items), GET, GET `/:id`, POST `/:id/decision` | admin, warehouse_manager, pharmacist, quality_officer |
| `/complaints` | POST, GET, GET `/:id` | admin, warehouse_manager, pharmacist, quality_officer |
| `/recalls` | POST (nested line items), GET, GET `/:id` | admin, warehouse_manager, pharmacist, quality_officer |
| `/dispatch-waybills` | POST, GET, GET `/:id` | admin, warehouse_manager, logistics_officer |
| `/delivery-confirmations` | POST, GET, GET `/:id` | admin, warehouse_manager, logistics_officer |
| `/short-expiry-reviews` | POST, GET | admin, warehouse_manager, pharmacist |
| `/emergency-plans` | POST, GET, GET `/:id` | admin, warehouse_manager |
| `/change-controls` | POST, GET, GET `/:id` | admin, warehouse_manager, quality_officer |

### Behavioral rules encoded in handlers

- **Donations / recalls** create header + nested `line_items` in a single transaction.
  Recall lines compute `quantity_outstanding = quantity_issued − quantity_recovered` at insert.
- **Donation decision** (`POST /donations/:id/decision`) sets decision + status + certificate fields.
- **Delivery confirmation** also flips its waybill to `delivered`.
- **CAPA effectiveness** (`POST /policy/capa/:id/effectiveness`), **excursion disposition**
  (`POST /temperature-excursions/:id/disposition`) close their parent record.
- **Expiring approvals** compares `expiry_date <= now()+60d` (configurable `?days=`, default 60).
- Org isolation is enforced by RLS, not by application-level org filters: every select carries
  `org_id` implicitly through the tenant session variable set by `middleware.Tenant()`.
- `created_by` / `updated_by` are required request fields (matching the existing handlers' style).

---

## 4. sqlc queries

`packages/data-models/sql/queries/policy_{regulatory,quality,coldchain,stock_controls,donations,events,dispatch,planning}.sql` (8 files).
Regenerate typed Go into `apps/api/internal/db/` with:

```bash
cd apps/api && sqlc generate
```

## 5. Tests

`apps/api/internal/handler/policy_test.go` — 19 sqlmock tests covering:

- Create/List/Get trios across the policy domains (regulatory, deviations, CAPA, excursions,
  donations, recalls, waybills, emergency plans, change controls).
- Transactional nested line items (donations, recalls) with `ExpectBegin`/`ExpectCommit`.
- Workflow endpoints: `/expiring`, `/effectiveness`, `/:id/disposition`, `/:id/decision`,
  and delivery-confirmation → waybill update.
- Validation-error (422) and not-found cases.

Run:

```bash
cd apps/api && go test ./internal/handler/ -count=1   # subset
cd apps/api && go test ./... -race -count=1            # full suite
```

---

## 6. Planned follow-ups

- Automated temperature-log ingestion from connected devices into `temperature_monitoring_entries`.
- Workflow endpoints for controlled-stock reconciliation and recall line-item recovery updates.
- BI reports: regulatory expiry matrix, mock-recall coverage, excursion/cold-chain dashboards.