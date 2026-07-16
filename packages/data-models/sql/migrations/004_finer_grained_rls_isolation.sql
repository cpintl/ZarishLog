-- ZarishLog — Finer-Grained RLS: Program, Org-Level & Department Isolation
-- PostgreSQL 18 — Adds session-level dimension scoping on top of org_id RLS
-- Migration 004: Idempotent, production-quality

-- ─── 0. ENSURE APP SCHEMA ─────────────────────────────────────────────────

CREATE SCHEMA IF NOT EXISTS app;

-- ─── 1. NEW SESSION-LEVEL SETTING FUNCTIONS ───────────────────────────────

CREATE OR REPLACE FUNCTION app.current_program_id() RETURNS text
LANGUAGE plpgsql STABLE PARALLEL SAFE
AS $$
BEGIN
  RETURN NULLIF(current_setting('app.current_program_id', true), '');
END;
$$;

CREATE OR REPLACE FUNCTION app.current_org_level_id() RETURNS text
LANGUAGE plpgsql STABLE PARALLEL SAFE
AS $$
BEGIN
  RETURN NULLIF(current_setting('app.current_org_level_id', true), '');
END;
$$;

CREATE OR REPLACE FUNCTION app.current_department_id() RETURNS text
LANGUAGE plpgsql STABLE PARALLEL SAFE
AS $$
BEGIN
  RETURN NULLIF(current_setting('app.current_department_id', true), '');
END;
$$;

-- Drop the unused unqualified helper from 001 if it exists; we replace it
-- with a fully-qualified, multi-dimensional alternative.
DROP FUNCTION IF EXISTS rls_org_policy();

-- ─── 2. COMPREHENSIVE POLICY EXPRESSION FUNCTION ──────────────────────────

CREATE OR REPLACE FUNCTION app.rls_policy_expression()
RETURNS text
LANGUAGE plpgsql STABLE
AS $$
DECLARE
  conditions text[] := '{}';
  result text;
BEGIN
  conditions := array_append(conditions, format('org_id = %L::uuid', app.current_org_id()));

  IF app.current_program_id() IS NOT NULL THEN
    conditions := array_append(conditions, format(
      '(%L IS NULL OR program_id IS NULL OR program_id::text = %L)',
      app.current_program_id(), app.current_program_id()
    ));
  END IF;

  IF app.current_org_level_id() IS NOT NULL THEN
    conditions := array_append(conditions, format(
      '(%L IS NULL OR org_level_id IS NULL OR org_level_id::text = %L)',
      app.current_org_level_id(), app.current_org_level_id()
    ));
  END IF;

  IF app.current_department_id() IS NOT NULL THEN
    conditions := array_append(conditions, format(
      '(%L IS NULL OR department_id IS NULL OR department_id::text = %L)',
      app.current_department_id(), app.current_department_id()
    ));
  END IF;

  result := array_to_string(conditions, ' AND ');
  RETURN result;
END;
$$;

-- ─── 3. RECREATE RLS POLICIES FOR ALL TENANT TABLES ──────────────────────

-- Enable RLS on all tenant-scoped tables (idempotent; repeated ALTERs are no-ops).
-- Includes tables from 001 and 002 migrations plus the new tables in 003.
-- Child/detail tables without an org_id column are skipped gracefully.

DO $$
DECLARE
  tbl text;
  has_org_id boolean;
  tables text[] := ARRAY[
    'org_levels', 'programs', 'departments', 'functions', 'users',
    'product_categories', 'products', 'product_packaging', 'product_substitutes',
    'suppliers', 'supplier_contracts', 'purchase_orders',
    'warehouses', 'locations', 'location_constraints',
    'batches', 'stock_levels', 'stock_movements', 'stock_snapshots',
    'goods_receipts', 'stock_issues', 'stock_transfers', 'stock_adjustments',
    'transfer_line_items', 'adjustment_line_items',
    'qa_inspections', 'qa_checklist_templates', 'qa_dispositions',
    'distributions', 'stock_returns', 'disposals',
    'stock_counts', 'count_line_items',
    'assets', 'asset_custody_changes', 'asset_maintenance',
    'amc_calculations', 'reorder_recommendations', 'forecast_results',
    'alert_configurations', 'alerts',
    'audit_log', 'data_change_log',
    'sync_log', 'sync_conflicts',
    'report_definitions', 'report_schedules'
  ];
BEGIN
  FOREACH tbl IN ARRAY tables
  LOOP
    -- Only proceed if the table exists
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = tbl AND schemaname = 'public') THEN
      -- Check whether the table has an org_id column
      SELECT EXISTS (
        SELECT 1 FROM pg_attribute
        WHERE attrelid = (tbl::regclass)
          AND attname = 'org_id'
          AND NOT attisdropped
      ) INTO has_org_id;

      -- Drop any stale policy regardless (idempotent)
      EXECUTE format('DROP POLICY IF EXISTS org_isolation ON %I', tbl);

      -- Only create the policy if the table is tenant-scoped (has org_id)
      IF has_org_id THEN
        EXECUTE format('
          ALTER TABLE %I ENABLE ROW LEVEL SECURITY
        ', tbl);
        EXECUTE format('
          CREATE POLICY org_isolation ON %I
            USING (%s)
            WITH CHECK (%s)
        ', tbl, app.rls_policy_expression(), app.rls_policy_expression());
      END IF;
    END IF;
  END LOOP;
END;
$$;

-- ─── 4. HELPER FUNCTION FOR APPLICATION MIDDLEWARE ───────────────────────

CREATE OR REPLACE FUNCTION app.set_isolation_context(
  p_org_id text DEFAULT NULL,
  p_program_id text DEFAULT NULL,
  p_org_level_id text DEFAULT NULL,
  p_department_id text DEFAULT NULL,
  p_user_id text DEFAULT NULL
) RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
  IF p_org_id IS NOT NULL THEN
    PERFORM set_config('app.current_org_id', p_org_id, true);
  END IF;
  IF p_program_id IS NOT NULL THEN
    PERFORM set_config('app.current_program_id', p_program_id, true);
  END IF;
  IF p_org_level_id IS NOT NULL THEN
    PERFORM set_config('app.current_org_level_id', p_org_level_id, true);
  END IF;
  IF p_department_id IS NOT NULL THEN
    PERFORM set_config('app.current_department_id', p_department_id, true);
  END IF;
  IF p_user_id IS NOT NULL THEN
    PERFORM set_config('app.current_user_id', p_user_id, true);
  END IF;
END;
$$;
