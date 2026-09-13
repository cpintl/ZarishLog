# ZarishLog — File-by-File P0/P1 Code Audit

**Date:** 2026-09-14
**Scope:** `apps/api`, `apps/web`, `packages/data-models` (migrations/queries), offline-sync code, against the external review of `cpintl/ZarishLog`.
**Method:** read of all routes/handlers/middleware, all 8 migrations + 25 query files, Dexie/sync/service-worker, README/ARCHITECTURE/STATUS; **empirical DB verification against PostgreSQL 18** (fresh apply of migrations 001→008 + seed; `pg_policies`/`pg_triggers`/`pg_constraint` checks; RLS tested with a least-privilege role).

---

## 0.5 Implementation status (2026-09-14)

P0-1..P0-6 are **implemented and live-verified** against PostgreSQL 18 with the new least-privilege runtime role `zarishlog_app`:

| Item | Fix | Verification |
|---|---|---|
| P0-1 dynamic RLS | `008_tenant_isolation_hardening.sql` rewrites the `org_isolation` policy text on every RLS table (`USING (org_id = app.current_org_id()::uuid)`) | 86 RLS tables / 86 policies / 0 RLS-without-policy; no-context → 0 rows; org-A context → only A rows |
| P0-2 tenant context | `middleware/tenant.go` pins one pool conn per request, runs `SELECT app.set_isolation_context(...)` (session-scoped GUCs), clears on exit; `main.go` registers it after `Auth()` | live API: org-A JWT sees only A; cross-tenant body → blocked |
| P0-3 app role | migration creates `zarishlog_app` (no ownership); grants on `public`/`app` schemas + CRUD; `config.go` default DSN is the app role | isolation proofs run as `zarishlog_app`; owner-only ops (seed/migrations) still work as `zarishlog` |
| P0-4 child-table RLS | join-based policies for all 28 child tables via dynamic DO loop (`locations`/`count_variance_reconciliation` special cases) | cross-tenant child INSERT → `violates row-level security policy` |
| P0-5 body `org_id` | all `policy_*.go`, `qa.go`, `distribution.go` create handlers now overwrite `req.OrgID = c.GetString("org_id")` after bind (JWT authoritative) | build + vet + handler tests green |
| P0-6 ledger tx | `stock.go` — GRN accepts items and, in one `Beginx` tx, writes header + line items + batches + `stock_movements` + `stock_levels`; transfer/adjustment append movements + `applyStockDelta` (guarded update, refuses negative balances) | live: GRN 201 → level 50 + receipt movement; transfer of 100 against 50 → 400 `insufficient stock` with full rollback; adjustment −5 → level 45 |
| P1-2 ledger append-only | `REVOKE UPDATE, DELETE ON stock_movements FROM zarishlog_app` + CHECK `quantity <> 0`; `stock_levels` non-negative CHECKs | constraints present (3) |
| P0-7 idempotency | `client_operations` dedupe table + `UNIQUE(org_id, client_operation_id)` | table + unique + RLS policy present |

**New findings surfaced by the live tests** (not fixed; out of scope):

- **`SELECT *` model mismatch** — `handler.ListProducts` / `ListRegulatoryApprovals` etc. scan `model.Product` (`reorder_formula` etc.) that the generated `models.go` no longer matches → `sqlx: missing destination name` → `GET /products` is `500 "failed to fetch products"` even though seed data exists. All `SELECT *` reads need a model/query reconciliation pass.
- **pgcrypto `uuid_generate_v7()` emits non-RFC variant nibbles** (observed `0,6,a,f,b`; RFC 4122 requires `8..b`) → the strict `uuid7` validator regex rejects most DB-generated ids, so seed ids/FKs (and the API) disagree; real clients get `422` on compliant-looking ids.
- **`categories` table doesn't exist** — `products.category_id` FK→`product_categories`, but Phase 1 STATUS claims a `categories` table; `DELETE /categories` routes reference it. Reference-data model needs a decision.
- **`CreateIssue` still header-only** — it is a requisition document (submit/approve), not an inventory change; FEFO pickup on issue remains future work (P-future, business-logic `SortByFEFO` unused).

---

## 0. Headline verdict

The external review's central claim is **confirmed and, in two places, understated**:

> "The architecture claims are ahead of the implementation."

| Review area | Review score | Verified reality |
|---|---|---|
| Multi-tenancy concept | 🟢 8.5/10 | Concept good; **implementation fixed 2026-09-14** (P0-1..P0-4): dynamic RLS policies, per-request tenant context via a pinned connection, least-privilege `zarishlog_app` role, child-table join policies |
| Offline-first | 🟡 6/10 | Worse: **no `/sync` routes exist at all**, the Dexie outbox has **zero callers**, and `sw.ts` is **never compiled or registered**. Reads cached <> writes queue are aspirational |
| API maturity | 🟡 6/10 | Confirmed (improved): stock handlers now **write the ledger** in one tx; rediscovery of `SELECT *` model mismatch + uuid-variant validator issues (see §0.5) |
| Security | 🟢 7/10 | P0-5 fixed: body `org_id` is no longer trusted — create handlers override with the JWT org. Auth hardening (P1-12) and audit-in-tx (P1-1) remain |

Two things are **better** than the review assumed:

- **Sequence/tx discipline exists where it was added** — `CreateTransfer`/`CreateAdjustment` use `Beginx` + deferred rollback + commit (`apps/api/internal/handler/stock.go:98-143, 163-210`) exactly as the review's "atomic operation" sketch demands. The problem is what is *not* in the tx: no movement, no stock-level mutation, no availability check.
- **Domain ENUMs are real DB types** (`movement_type`, `item_type`, `warehouse_type`, `location_type`, `uom_category` — `packages/data-models/sql/migrations/001_initial_schema.sql:46-53`), matching the `validator` tags.

---

## P0 — must fix before calling this production-ready

### P0-1. RLS is dead on 52 of 57 tables (migration bug bakes `org_id = NULL`)

**What the review assumed:** RLS enforced via `app.current_org_id` set per-request. **Actual (DB-verified):** 52 tables have exactly one policy, `org_isolation`, whose `USING`/`WITH CHECK` literally evaluates `org_id = NULL::uuid → false` for every row. Live test: with the tenant GUC correctly set, `SELECT count(*) FROM products` returns **0**.

**Root cause:**
- `packages/data-models/sql/migrations/004_finer_grained_rls_isolation.sql:49` — builds the policy string with `format('org_id = %L::uuid', app.current_org_id())`, evaluating `app.current_org_id()` **at migration time** (unset → `NULL`) instead of embedding the function call.
- `004:127-138` — DO block `DROP POLICY` + `CREATE POLICY ... USING (%s)` with that baked string, iterating ~57 tables.
- `004:146-147` — `locations` re-created correctly with a subquery (dynamic); same correct dynamic pattern at `001:517-518`, `002:811-812`, `007:628-632`.
- `007_cpi_medical_policy_controls.sql:613-619` — same `format(..., current_org_id())` bug for the 16 new register tables.

**Files to change:** `004:41-75` (`app.rls_policy_expression()` — return the *text* `'org_id = app.current_org_id()::uuid'`, not an evaluated value), `007:613-619`; then a corrective migration `008` that re-runs `CREATE OR REPLACE POLICY org_isolation ... USING (org_id = app.current_org_id()::uuid)` on all 52 affected tables, and `ALTER TABLE ... FORCE ROW LEVEL SECURITY` on all tenant tables.

### P0-2. Nothing ever sets the tenant context — and `Tenant()` middleware is a no-op header setter

**What the review assumed:** "JWT → subject → memberships → tenant → `SET LOCAL` in transaction." **Actual:**
- `apps/api/internal/middleware/tenant.go:7-15` — only does `c.Request.Header.Set("X-Tenant-ID", orgID)`. That header has **no reader anywhere**.
- Grep of `apps/api` for `current_org_id|set_config|set_isolation_context` → **zero hits**.
- `app.set_isolation_context(...)` exists (`004:157-183`, uses `PERFORM set_config('app.current_org_id', p_org_id, true)`) but is **never invoked**.

**Files to change:** `apps/api/internal/middleware/tenant.go` — make it db-aware (it must run *after* `middleware.Auth`), execute `SELECT app.set_isolation_context($1::uuid, ...)` (or `SET LOCAL app.current_org_id = $1`) on the request's connection, and be registered after `Auth()` in `apps/api/cmd/api/main.go:55-57`. Because Gin pools connections, the set must happen on *every* request on the connection the handler then uses — the cleanest approach is a `db.Begin` + `SET LOCAL` held for the handler's duration (or RLS keyed on the transaction).

### P0-3. The app connects as a superuser/table owner → RLS bypassed; no role model

**Actual (DB-verified):**
- `docker-compose.yml:8` — `POSTGRES_USER: zarishlog` = superuser.
- `apps/api/internal/config/config.go:27` — default `postgresql://zarishlog:zarishlog_dev_password@...` connects as the **table owner**; `relforcerowsecurity = 0` for all 57 tables → owner bypasses RLS.
- No `GRANT`/`REVOKE`/`CREATE ROLE` anywhere in the migrations; PostgreSQL 15+ grants no `USAGE ON SCHEMA public` by default, and schema `app` has no grants (verified: a limited role gets `permission denied for schema app`).
- Any DB role can `SET app.current_org_id` (custom GUCs are unrestricted).

**Files to change:** `packages/data-models/sql/migrations/008_*.sql` — create `zarishlog_app` login role with grants on `app` schema (`GRANT USAGE ON SCHEMA app`, `EXECUTE` on `set_isolation_context`), `GRANT USAGE ON SCHEMA public`, `SELECT`/`INSERT`/`UPDATE`/`DELETE` on tenant tables, `ALTER TABLE ... FORCE ROW LEVEL SECURITY`; `docker-compose.yml` — add the app role; `apps/api/internal/config/config.go` — default `DATABASE_URL` for the app role. `ALTER ROLE ... NOSET` can restrict who sets `app.*` GUCs.

### P0-4. 28 tenant child tables have no `org_id` and no RLS (cross-tenant leakage)

**Actual:** 28 detail/child tables lack `org_id`, RLS, or both — verified as **not** in the RLS-enabled set: `grn_line_items` (001:333), `issue_line_items` (001:364), `product_packaging` (002:10), `product_substitutes` (002:28), `product_attachments` (002:39), `user_role_assignments` (002:69), `user_sessions` (002:80), `user_preferences` (002:93), `po_line_items` (002:181), `location_constraints` (002:200), `warehouse_documents` (002:215), `transfer_line_items` (002:246), `adjustment_line_items` (002:267), `qa_checklist_items` (002:302), `qa_checklist_results` (002:313), `qa_dispositions` (002:324), `distribution_line_items` (002:359), `distribution_beneficiaries` (002:371), `return_line_items` (002:411), `disposal_line_items` (002:442), `count_line_items` (002:474), `count_variance_reconciliation` (002:490), `asset_depreciation_schedule` (002:504), `asset_attachments` (002:516), `alert_recipients` (002:623), `report_schedules` (002:717), `asset_custody_changes` (001:446), `asset_maintenance` (001:455).

**Root cause:** the `004:106-149` loop only generates a policy when a table *directly* has `org_id` or `warehouse_id`; it silently skips parent-keyed children. Only `donation_line_items`/`recall_line_items` got the parent-join treatment (`007:630-632, 637-639`).

**Files to change:** migration `008` — for parent-keyed children, either add an `org_id` column with `NOT NULL DEFAULT` populated from the parent, or create join-based policies (e.g. `transfer_line_items` → `stock_transfers.org_id`), as already proven for `locations`/`donation_line_items`.

### P0-5. Client-supplied `org_id` in request bodies is trusted into INSERTs

**Actual (verified by grep):** many request structs bind `OrgID string json:"org_id" validate:"required,uuid7"` and the handler inserts `req.OrgID` **verbatim**, with no cross-check against the JWT org. Confirmed instances:

- `apps/api/internal/handler/policy_stock.go:19,48` (`stock_release_records`), `:106,136` (`controlled_stock_register`), `:195,226` (`short_expiry_reviews`)
- `apps/api/internal/handler/policy_dispatch.go:19,56` (`dispatch_waybills`), `:117,149` (`delivery_confirmations`)
- `apps/api/internal/handler/policy_regulatory.go:20,55` (`regulatory_approvals`)
- `apps/api/internal/handler/qa.go:18,45` (`qa_inspections`), `:151,182` (`qa_checklist_templates`)
- `apps/api/internal/handler/distribution.go:15,50` (`distributions`)
- `apps/api/internal/handler/policy_coldchain.go:19,45` (`temperature_monitoring_entries`), `:90,130` (`temperature_excursions`)
- `apps/api/internal/handler/policy_events.go:19,56` (`complaints`), `:129,170` (`recalls`)

**Contrast (correct pattern):** `stock.go:23,59,105,170` and the catalogue handlers (`warehouse.go:56`, `category.go:43`, `import.go:39`) **override** with `c.GetString("org_id")` (JWT). The policy/QA/distribution handlers must do the same.

**Files to change:** every `Create*` handler above — remove `OrgID` from the accepted body (or validate `req.OrgID == c.GetString("org_id")`) and set `req.OrgID` from the JWT. Models affected are in `apps/api/internal/model/` alongside each handler.

### P0-6. Stock transactions don't actually move stock

**Actual (verified):** the ledger and derived state are **never written by any handler**:
- `CreateGRN` (`stock.go:12-45`) — single `db.QueryRow` INSERT of the header only; **no `grn_line_items`, no `stock_movements`, no `stock_levels`**.
- `CreateIssue` (`stock.go:48-82`) — header only; no movement, no FEFO pickup, no level decrement. (`packages/business-logic` `SortByFEFO`/`CalculatePick` are never called by the API.)
- `CreateTransfer` (`stock.go:84-147`) — **atomic** (`Beginx`/deferred rollback/commit) for header+line items, but no `stock_movements`, no `stock_levels` delta, no `SELECT ... FOR UPDATE` availability/lock check.
- `CreateAdjustment` (`stock.go:149-214`) — same shape; computes `difference` but writes no level.
- Reads hit `stock_levels` directly (`stock.go:216-258`) — a table nothing populates, so `GET /stock/levels` is empty.

**Files to change:** `apps/api/internal/handler/stock.go` — unify all four into one ledger transaction: lock batch/level rows (`FOR UPDATE`), verify availability per movement type, `INSERT stock_movements`, apply/derive `stock_levels` delta, insert audit/outbox in the same tx. `GetBatchTrail` reads already exist (`stock.go:260-290`).

### P0-7. Offline sync is scaffolding with no endpoints and no callers

**Actual (verified):**
- No `/api/v1/sync` group exists in `apps/api/cmd/api/main.go` — `sync/push` and `sync/pull` from `docs/ARCHITECTURE.md:217-218` return **404**, they are not even stubs. (`response.NotImplemented` at `internal/response/response.go:77-78` is never called.)
- `apps/web/lib/sync.ts:3-21` `queueMutation()` — the only enqueue primitive — has **zero callers** in the repo. `processQueue()` (`sync.ts:23-67`) drains blindly: no client op id sent (`sync.ts:46-50` sends only `Content-Type`), no backoff, no dedupe, ignores response body, `retryCount >= 5` leaves a permanent poisoned row (`sync.ts:37-40`).
- `apps/web/sw.ts` (Workbox + `BackgroundSyncPlugin("zarishlog-sync", {maxRetentionTime: 1440})` at `sw.ts:8-18`) is **never built or registered**: `next.config.mjs:2-5` has no PWA plugin, no `navigator.serviceWorker.register` anywhere, no `sw.js` in build output.
- No idempotency: zero references to `client_operation_id`/`Idempotency-Key` repo-wide; no `UNIQUE(org_id, client_operation_id)` in any migration.

**Files to change (this is the "release gate" item #2):**
1. **DB:** migration `008` — `UNIQUE(org_id, client_operation_id)` on all offline-writable tables (or an `offline_ops` dedupe table).
2. **API:** new `apps/api/internal/handler/sync.go` + routes in `main.go` (`POST /api/v1/sync/push`, `GET /api/v1/sync/pull`) powered by the *already-generated* `internal/db/sync.sql.go` (`CreateSyncLog`, `CreateSyncConflict`, `GetPendingConflicts`, `ResolveSyncConflict` — currently 100% dead).
3. **Web:** `lib/sync.ts` — generate + resend a durable `client_operation_id`, send it as `Idempotency-Key`, apply backoff, treat 409 as success (already-applied). `lib/db.ts:39-48` (`OfflineMutation`) already has the shape.
4. **Web/PWA:** `next.config.mjs` + registration — wire `sw.ts` through a build plugin (`@serwist/next` or `next-pwa`) and `lib/` registration the day the API exists; until then it stays dead code.

---

## P1 — next batch

### P1-1. Audit is a fire-and-forget goroutine with swallowed errors and no before/after
`apps/api/internal/middleware/audit.go:41-47` — `go func(){ _, _ = db.Exec(...) }()`; skips GETs (`:13-16`), only writes on `status < 500` success (`:24-26`), `entity_type` = route pattern (`:36`), never fills `audit_log.changes` (`002:634-645` column exists), never records before/after, and `data_change_log` (`002:648-658`) has **no writer at all**. There is also no read API (generated `SearchAuditLogs` is unused; no route). **Fix:** write inside the handler tx (P0-6) or a dedicated worker on the same connection; capture `changes`; add `GET /audit` route; delete/gate the goroutine.

### P1-2. Ledger not append-only at the DB layer
Zero triggers (verified), direct `UPDATE stock_levels` query exists (`packages/data-models/sql/queries/stock.sql:10-13`), no CHECK on `stock_movements`.quantity sign/zero, `ref_doc_type`/`ref_doc_id` are free `text` with no FK to documents (`001:301-302`), no `DEFERRABLE` constraints. **Fix:** `008` — `REVOKE UPDATE, DELETE ON stock_movements FROM zarishlog_app;` add CHECKs, FK `stock_movements`→documents per type, migration-safe idempotency unique.

### P1-3. Hard DELETE handlers vs. soft-delete convention
`DELETE /products/:id` (`main.go:69`), `/warehouses/:id` (`:86`), `/uoms/:id` (`:104`), `/assets/:id` (`:267`), plus `DeleteLocation` (`:92`). Soft-delete exists only in *some* queries (`products.sql:49-51`, `categories.sql:21-23`); `warehouses`/`uoms`/`assets`/`locations` handlers issue destructive `DELETE` (`warehouse.go:105`). No `deleted_at` column exists. **Fix:** lifecycle fields (already present as `status`/`is_active`) + change handlers to `UPDATE`.

### P1-4. The 116 dead sqlc queries define a contract the API doesn't implement
Of 205 generated methods (`apps/api/internal/db/`), **116 are referenced by no handler**. Entirely dead query files: `alerts.sql`, `audit.sql`, `organizations.sql`, `procurement.sql`, `qa.sql`, `reports.sql`, `sync.sql`; 16/18 in `transactions.sql`. Notably dead: `CreateStockMovement`, `UpdateStockLevel`, `CreateSyncLog`, `CreateSyncConflict`, `CreateAuditLog`, `SearchStockMovements`. The review's "choose sqlc or sqlx" is more pointed: **sqlc is the orphan here** — every handler uses raw `sqlx`, and `go.mod:11` keeps both. **Decision needed:** adopt sqlc for all writes (recommended, matches `AGENTS.md`) or delete the package. Do not leave both.

### P1-5. Tenant-isolation and stock tests do not exist
`internal/middleware/` has **zero tests** (auth/tenant/audit/rbac). RLS is exercised by no test (grep `RLS|current_org_id|isolation` in tests → 0). `stock_test.go` mocks happy-path `ExpectBegin→ExpectCommit` only (`:53,61`); no rollback-path test. `packages/business-logic/amc.go` has no tests (only `fefo_test.go`). The handlers mock `c.Set("org_id", ...)` rather than validate cross-tenant behavior. **Fix:** the P0-1..P0-5 work is not done until a test asserts "tenant A cannot read B" end-to-end against a real DB, plus DB-level unit tests using `sqlmock` for the movement-write tx.

### P1-6. Reporting endpoints don't exist at all
`docs/ARCHITECTURE.md:214-216` lists `[STUB]` report endpoints; in code there is no `/reports` group and no executor — `reports.sql.go` only CRUDs the *definition* catalog. `GET /stock/expiring` (`main.go:117`) and AMC (`replenishment.go:16`) are the nearest things. **Fix:** build the reporting views (`vw_stock_on_hand`, `vw_stock_valuation`, `vw_expiry_risk`, etc.) in `008`, then thin handlers reading those views.

### P1-7. Missing composite indexes on hot paths
- `stock_movements`: needs `(org_id, batch_id)`, `(org_id, created_at DESC)`, `(org_id, product_id, created_at DESC)` for `SearchStockMovements`/`GetBatchTrail` (`queries/stock.sql:49-57`).
- `batches`: `(org_id, expiry_date)` for `GetExpiringBatches` (`stock.sql:44-47`).
- `sync_conflicts`: `(org_id, resolution, created_at)` for `GetPendingConflicts` (`queries/sync.sql:23-26` — table currently has **no** index beyond PK).

### P1-8. Health is a DB ping only
`apps/api/internal/handler/health.go:11` — `db.Ping()`; no Redis/Meilisearch/MinIO checks despite `docker-compose` shipping all three, no readiness/liveness split. **Fix:** extend with dependency checks and `/readyz`.

### P1-9. Frontend has no auth, no tenant header, no idempotency, no validation
- `app/products/page.tsx:16` and `app/status/page.tsx:14-16` hit the Go API unauthenticated (no `Authorization`, no tenant header); `request()` in `lib/config-studio/api.ts:9-22` likewise. The whole OIDC/Keycloak model is backend-only — the review's "offline locks the worker out" concern is unreachable because nothing authenticates at all yet.
- `zod` is declared (`package.json:26`) but imported nowhere; API data is unchecked casts.
- **Fix:** introduce a single `lib/api.ts` client (base URL, Bearer token from session, `X-Tenant-ID`, `Idempotency-Key`), channel all pages through it, validate with Zod schemas.

### P1-10. Docs materially overstate completion
- `README.md:7,264` ("Version 1.0.0", "All 12 planned phases of core development are complete") vs. `docs/STATUS.md:24-25` (Phases 11-12 ❌ Not started) and no sync routes. `README.md:13,39,80` ("works entirely offline... changes sync automatically") is not true of any write path.
- `docs/ARCHITECTURE.md:214-218` labels absent routes as `[STUB]`; `:241` claims RLS is set per-request (it isn't); `:243` marks audit "integration in progress" (correct).
- `docs/STATUS.md:14` "RLS on 82 tenant tables" vs `:369` "55 enforced" vs actual **57 enabled / 5 effective**.
- **Fix:** change maturity label (e.g. `0.9.x — Release Candidate`), correct ARCHITECTURE/STATUS numbers after P0-1..P0-5, add `docs/REVIEW-AUDIT.md` trail, keep `CONFIG/` and template truth in sync.

### P1-11. Delete cascades destroy audit chains
`grn_line_items→goods_receipts` CASCADE (`001:335`), `issue_line_items` (`001:366`), `transfer_line_items` (`002:248`), `adjustment_line_items` (`002:269`), `asset_depreciation_schedule`/`asset_attachments` (`002:506,518`), `warehouses→locations` (`001:242`). Deleting a parent transaction or asset silently removes the historical record. **Fix:** prohibit delete on ledger/document parents (`008`), keep cascade only on truly cosmetic children.

### P1-12. Auth hardening
`auth.go:33-35` parses HS256 with the shared `cfg.JWTSecret` and **never checks `token.Method` or the issuer/audience**; `OIDCIssuer` (`config.go:30`) is unused; `rbac.go:23-29` is a string-scan of the (client-asserted) `roles` claim. **Fix:** enforce `alg == HS256` + `iss`/`aud`, or migrate to JWKS verification against Keycloak; move roles into a server-resolved membership (the review's "Role × Scope × Action" at minimum needs the role source untrusted-from-JWT design).

---

## Release-gate mapping (from the review)

| Review gate | Status | Proof/file targets |
|---|---|---|
| 1. Tenant isolation proven | 🟢 Proven | P0-1..P0-5; migration applied + live end-to-end (org-A JWT creates own-org rows only; body `org_id` overridden by claims; DB isolation proofs as `zarishlog_app`) |
| 2. Offline sync proven | 🔴 Not | P0-7; no endpoints exist (idempotency table lands, protocol remains) |
| 3. Stock ledger transaction-safe | 🟢 Proven | P0-6 + P1-2 revokes; one tx per operation, append-only movements, guarded level deltas; live-verified with rollback |
| 4. Complete audit trail | 🟡 Partial | P1-1; table exists, writes incomplete, no read API |
| 5. Recovery + backup tested | 🔴 Not exercised | no restore/DR test, no healthy role/backup job |

## Suggested fix order

1. **008 migration** — fix P0-1 (dynamic policies + `FORCE RLS`), P0-4 (child-table RLS), P0-3 (app role + grants), plus P1-2/P1-7 (ledger revokes + indexes) and P0-7 (idempotency unique).
2. **`middleware/tenant.go` + `main.go:55-57`** — real tenant context inside the request tx (P0-2).
3. **Handlers** — stop trusting body `org_id` (P0-5); make stock writes one ledger tx (P0-6); finish audit inside the tx (P1-1).
4. **`internal/db` adoption or deletion** — one source of truth (P1-4).
5. **RLS/isolation + rollback tests** — the gate #1 proof (P1-5).
6. **Sync protocol** — `client_operation_id`, `UNIQUE(org_id, ...)`, `/sync/push|pull` handlers, wire `lib/sync.ts` and `sw.ts` (P0-7).
7. **Docs** — maturity label + status correction (P1-10).

Estimated effort: P0-1..P0-5 and P1-1/P1-2 are small, contained changes (one migration + one middleware + ~12 handler edits + tests). P0-7 (sync) and P1-4 (sqlc decision) are the two decisions
that need a call before implementation.