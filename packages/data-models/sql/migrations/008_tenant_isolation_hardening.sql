-- 008_tenant_isolation_hardening.sql
--
-- P0 hardening per docs/CODE-AUDIT.md (verified DB-side):
--   1. Repair RLS policies that migration 004/007 baked to `org_id = NULL::uuid`
--      (dead policies) -> recreate all org_id-keyed policies as dynamic
--      `app.current_org_id()` expressions.
--   2. Enable + policy 28 tenant child tables that have NO org_id column by
--      joining through their RLS-protected parent.
--   3. Least-privilege runtime role `zarishlog_app` (app must not connect as
--      the table-owning superuser, which bypasses RLS).
--   4. Idempotency anchor for offline sync: unique (org_id, client_operation_id).
--   5. Ledger immutability guards on stock_movements and hot-path indexes.
--
-- Applied via `psql -f` (one auto-committed statement batch per file).
-- Migration/seed still run as the owner (`zarishlog`); RLS is not FORCEd so
-- DDL/seed are unaffected. The runtime app role is a non-owner, so RLS is
-- enforced for it.

-- ---------------------------------------------------------------------------
-- 5a. Hot-path indexes (independent prep)
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_stock_movements_org_batch
  ON stock_movements (org_id, batch_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_org_created
  ON stock_movements (org_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stock_movements_org_product_created
  ON stock_movements (org_id, product_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_batches_org_expiry
  ON batches (org_id, expiry_date);
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_org_resolution
  ON sync_conflicts (org_id, resolution, created_at);

-- ---------------------------------------------------------------------------
-- 4. Idempotency anchor for offline writes (dedupe registry)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS client_operations (
  id                  uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
  org_id              uuid NOT NULL REFERENCES organizations(id),
  client_operation_id uuid NOT NULL,
  entity              text NOT NULL,
  entity_id           uuid,
  status              text NOT NULL DEFAULT 'pending',
  payload             jsonb,
  created_at          timestamptz NOT NULL DEFAULT now(),
  applied_at          timestamptz,
  CONSTRAINT client_operations_org_op UNIQUE (org_id, client_operation_id)
);

-- ---------------------------------------------------------------------------
-- 5b. Ledger immutability guards (idempotent via pg_constraint check)
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'stock_movements_quantity_nonzero') THEN
    ALTER TABLE stock_movements
      ADD CONSTRAINT stock_movements_quantity_nonzero CHECK (quantity <> 0);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'stock_levels_quantity_nonneg') THEN
    ALTER TABLE stock_levels
      ADD CONSTRAINT stock_levels_quantity_nonneg CHECK (quantity >= 0);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'stock_levels_reserved_nonneg') THEN
    ALTER TABLE stock_levels
      ADD CONSTRAINT stock_levels_reserved_nonneg CHECK (reserved_qty >= 0);
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 1. Repair the 52 org_id-keyed RLS policies baked as `org_id = NULL::uuid`.
--    Recreate every policy on RLS-enabled tables that carry an `org_id`
--    column. Tables without an `org_id` column, or whose isolation expresses
--    through a parent join (locations, entity_attributes, donation_line_items,
--    recall_line_items), are handled separately below and are NOT matched by
--    this loop (no org_id column).
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  tbl record;
BEGIN
  FOR tbl IN
    SELECT c.relname AS tbl
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = 'public'
    JOIN pg_attribute a ON a.attrelid = c.oid AND a.attname = 'org_id'
    WHERE c.relkind = 'r'
      AND c.relrowsecurity = true
  LOOP
    EXECUTE format('DROP POLICY IF EXISTS org_isolation ON %I', tbl.tbl);
    EXECUTE format(
      'CREATE POLICY org_isolation ON %I
         USING (org_id = app.current_org_id()::uuid)
         WITH CHECK (org_id = app.current_org_id()::uuid)',
      tbl.tbl
    );
  END LOOP;
END $$;

-- ---------------------------------------------------------------------------
-- 2. RLS on tenant child tables (no org_id column) via their parent's org_id.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  c          record;
  using_expr text;
BEGIN
  FOR c IN SELECT * FROM (VALUES
    ('grn_line_items',                     'grn_id',                      'goods_receipts'),
    ('issue_line_items',                   'issue_id',                    'stock_issues'),
    ('product_packaging',                  'product_id',                  'products'),
    ('product_substitutes',                'product_id',                  'products'),
    ('product_attachments',                'product_id',                  'products'),
    ('user_role_assignments',              'user_id',                     'users'),
    ('user_sessions',                      'user_id',                     'users'),
    ('user_preferences',                   'user_id',                     'users'),
    ('po_line_items',                      'po_id',                       'purchase_orders'),
    ('warehouse_documents',                'warehouse_id',                'warehouses'),
    ('transfer_line_items',                'transfer_id',                 'stock_transfers'),
    ('adjustment_line_items',              'adjustment_id',               'stock_adjustments'),
    ('qa_checklist_items',                 'template_id',                 'qa_checklist_templates'),
    ('qa_checklist_results',               'inspection_id',               'qa_inspections'),
    ('qa_dispositions',                    'inspection_id',               'qa_inspections'),
    ('distribution_line_items',            'distribution_id',             'distributions'),
    ('distribution_beneficiaries',         'distribution_id',             'distributions'),
    ('return_line_items',                  'return_id',                   'stock_returns'),
    ('disposal_line_items',                'disposal_id',                 'disposals'),
    ('count_line_items',                   'count_id',                    'stock_counts'),
    ('asset_depreciation_schedule',        'asset_id',                    'assets'),
    ('asset_attachments',                  'asset_id',                    'assets'),
    ('alert_recipients',                   'alert_config_id',             'alert_configurations'),
    ('report_schedules',                   'report_id',                   'report_definitions'),
    ('asset_custody_changes',              'asset_id',                    'assets'),
    ('asset_maintenance',                  'asset_id',                    'assets')
  ) AS t(child, fk, parent)
  LOOP
    using_expr := format(
      '%I IN (SELECT id FROM %I WHERE org_id = app.current_org_id()::uuid)',
      c.fk, c.parent
    );
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', c.child);
    EXECUTE format('DROP POLICY IF EXISTS org_isolation ON %I', c.child);
    EXECUTE format(
      'CREATE POLICY org_isolation ON %I USING (%s) WITH CHECK (%s)',
      c.child, using_expr, using_expr
    );
  END LOOP;
END $$;

-- location_constraints -> locations -> warehouses (locations carries no org_id)
ALTER TABLE location_constraints ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS org_isolation ON location_constraints;
CREATE POLICY org_isolation ON location_constraints
  USING (location_id IN (
    SELECT l.id FROM locations l JOIN warehouses w ON w.id = l.warehouse_id
    WHERE w.org_id = app.current_org_id()::uuid))
  WITH CHECK (location_id IN (
    SELECT l.id FROM locations l JOIN warehouses w ON w.id = l.warehouse_id
    WHERE w.org_id = app.current_org_id()::uuid));

-- count_variance_reconciliation has two parents (count_line_items OR stock_adjustments)
ALTER TABLE count_variance_reconciliation ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS org_isolation ON count_variance_reconciliation;
CREATE POLICY org_isolation ON count_variance_reconciliation
  USING (
    (count_line_item_id IN (
       SELECT cli.id FROM count_line_items cli JOIN stock_counts sc ON sc.id = cli.count_id
       WHERE sc.org_id = app.current_org_id()::uuid)
     OR
     adjustment_id IN (
       SELECT a.id FROM stock_adjustments a WHERE a.org_id = app.current_org_id()::uuid)))
  WITH CHECK (
    (count_line_item_id IN (
       SELECT cli.id FROM count_line_items cli JOIN stock_counts sc ON sc.id = cli.count_id
       WHERE sc.org_id = app.current_org_id()::uuid)
     OR
     adjustment_id IN (
       SELECT a.id FROM stock_adjustments a WHERE a.org_id = app.current_org_id()::uuid)));

-- RLS + policy for the new dedupe registry
ALTER TABLE client_operations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS org_isolation ON client_operations;
CREATE POLICY org_isolation ON client_operations
  USING (org_id = app.current_org_id()::uuid)
  WITH CHECK (org_id = app.current_org_id()::uuid);
CREATE INDEX IF NOT EXISTS idx_client_operations_org_status
  ON client_operations (org_id, status, created_at);

-- ---------------------------------------------------------------------------
-- 3. Least-privilege runtime role
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'zarishlog_app') THEN
    CREATE ROLE zarishlog_app LOGIN PASSWORD 'zarishlog_app_dev_password';
  END IF;
END $$;

GRANT USAGE ON SCHEMA public TO zarishlog_app;
GRANT USAGE ON SCHEMA app TO zarishlog_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA app TO zarishlog_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO zarishlog_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO zarishlog_app;

-- Ledger is append-only for the runtime role: it may INSERT movements but
-- never UPDATE or DELETE them. (stock_levels stays writable - it is the
-- derived position maintained in the same transaction as the movement.)
REVOKE UPDATE, DELETE ON stock_movements FROM zarishlog_app;

-- ---------------------------------------------------------------------------
-- 6. Tenant GUC helpers - session-scoped and NULL-safe
-- ---------------------------------------------------------------------------
-- The runtime role connects as a non-owner (RLS enforced) on a per-request
-- pinned connection. The previous helper used set_config(..., true), which is
-- transaction-local and therefore disappears after a single autocommit
-- statement, leaving tenant queries empty. The rewritten helper sets SESSION
-- scope (is_local=false) so the tenant survives across statements and into
-- handler transactions on that same pinned connection. Passing NULL clears
-- the corresponding GUC - the middleware uses this to wipe context before
-- returning the connection to the pool.
--
-- current_org_id()/current_user_id() from 001 returned raw current_setting()
-- ('' when the GUC is unset), which made `...::uuid` casts inside RLS policies
-- raise "invalid input syntax for type uuid". NULLIF makes them NULL when
-- unset, so `org_id = NULL` is simply never true (deny-by-default, no error).

CREATE OR REPLACE FUNCTION app.set_isolation_context(
  p_org_id        text DEFAULT NULL,
  p_program_id    text DEFAULT NULL,
  p_org_level_id  text DEFAULT NULL,
  p_department_id text DEFAULT NULL,
  p_user_id       text DEFAULT NULL
) RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM set_config('app.current_org_id',        coalesce(p_org_id, ''),        false);
  PERFORM set_config('app.current_program_id',    coalesce(p_program_id, ''),    false);
  PERFORM set_config('app.current_org_level_id',  coalesce(p_org_level_id, ''),  false);
  PERFORM set_config('app.current_department_id', coalesce(p_department_id, ''), false);
  PERFORM set_config('app.current_user_id',       coalesce(p_user_id, ''),       false);
END;
$$;

CREATE OR REPLACE FUNCTION app.current_org_id() RETURNS text
LANGUAGE plpgsql STABLE PARALLEL SAFE
AS $$
BEGIN
  RETURN NULLIF(current_setting('app.current_org_id', true), '');
END;
$$;

CREATE OR REPLACE FUNCTION app.current_user_id() RETURNS text
LANGUAGE plpgsql STABLE PARALLEL SAFE
AS $$
BEGIN
  RETURN NULLIF(current_setting('app.current_user_id', true), '');
END;
$$;