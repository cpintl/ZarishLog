-- ZarishLog — Extended Schema (Phase 2)
-- PostgreSQL 18, Multi-tenant, UUIDv7, Audit columns, RLS
-- Adds packaging, procurement, warehouse, stock, QA, distribution,
-- returns/disposal, physical count, asset management, replenishment,
-- notifications, audit, offline sync, and report tables.

-- ─── 1. MASTER DATA ENHANCEMENTS ────────────────────────────────────────

-- Product packaging / hierarchy levels
CREATE TABLE product_packaging (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    product_id          uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    uom_id              uuid NOT NULL REFERENCES units_of_measure(id),
    package_level       int NOT NULL CHECK (package_level >= 1),
    quantity            numeric(12,3) NOT NULL,
    package_description text,
    barcode             text,
    length              numeric(10,2),
    width               numeric(10,2),
    height              numeric(10,2),
    weight              numeric(10,2),
    status              entity_status NOT NULL DEFAULT 'active',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- Product substitutes / alternatives
CREATE TABLE product_substitutes (
    product_id          uuid NOT NULL REFERENCES products(id),
    substitute_product_id uuid NOT NULL REFERENCES products(id),
    substitution_type   text NOT NULL CHECK (substitution_type IN ('therapeutic', 'generic', 'brand', 'equivalent')),
    priority            int NOT NULL DEFAULT 1,
    is_active           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (product_id, substitute_product_id)
);

-- Product attachments (SOPs, photos, MSDS)
CREATE TABLE product_attachments (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    product_id  uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    file_name   text NOT NULL,
    file_type   text,
    file_url    text NOT NULL,
    description text,
    uploaded_by text,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- ─── 2. ENHANCED ORGANIZATION & USERS ───────────────────────────────────

-- Functions / roles within a department
CREATE TABLE functions (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    department_id uuid NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    code          text NOT NULL,
    name          text NOT NULL,
    description   text,
    status        entity_status NOT NULL DEFAULT 'active',
    created_by    text,
    updated_by    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

-- User-role assignments with org-level scope (replaces single role_id on users)
CREATE TABLE user_role_assignments (
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     uuid NOT NULL REFERENCES roles(id),
    org_level_id uuid REFERENCES org_levels(id),
    granted_by  text,
    granted_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz,
    PRIMARY KEY (user_id, role_id, org_level_id)
);

-- User sessions
CREATE TABLE user_sessions (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token           text NOT NULL UNIQUE,
    refresh_token   text,
    ip_address      text,
    user_agent      text,
    expires_at      timestamptz NOT NULL,
    last_activity_at timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- User preferences
CREATE TABLE user_preferences (
    user_id             uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    language            text NOT NULL DEFAULT 'en',
    theme               text NOT NULL DEFAULT 'light',
    timezone            text NOT NULL DEFAULT 'UTC',
    date_format         text NOT NULL DEFAULT 'YYYY-MM-DD',
    notification_email  boolean NOT NULL DEFAULT true,
    notification_push   boolean NOT NULL DEFAULT true,
    dashboard_config    jsonb,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- ─── 3. SUPPLIERS & PROCUREMENT ─────────────────────────────────────────

-- Suppliers / vendors
CREATE TABLE suppliers (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    code            text NOT NULL,
    name            text NOT NULL,
    type            text CHECK (type IN ('manufacturer', 'distributor', 'wholesaler', 'local', 'international')),
    contact_person  text,
    email           text,
    phone           text,
    address         text,
    city            text,
    country         text,
    tax_id          text,
    payment_terms   text,
    is_active       boolean NOT NULL DEFAULT true,
    status          entity_status NOT NULL DEFAULT 'active',
    created_by      text,
    updated_by      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

-- Supplier contracts
CREATE TABLE supplier_contracts (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    supplier_id     uuid NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    contract_number text NOT NULL,
    contract_date   date NOT NULL,
    expiry_date     date,
    terms           text,
    discount_percent numeric(5,2) NOT NULL DEFAULT 0,
    currency        text NOT NULL DEFAULT 'USD',
    is_exclusive    boolean NOT NULL DEFAULT false,
    status          entity_status NOT NULL DEFAULT 'active',
    created_by      text,
    updated_by      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, contract_number)
);

-- Purchase orders
CREATE TABLE purchase_orders (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                uuid NOT NULL REFERENCES organizations(id),
    po_number             text NOT NULL,
    supplier_id           uuid REFERENCES suppliers(id),
    warehouse_id          uuid REFERENCES warehouses(id),
    program_id            uuid REFERENCES programs(id),
    department_id         uuid REFERENCES departments(id),
    order_date            date NOT NULL DEFAULT CURRENT_DATE,
    expected_delivery_date date,
    delivery_address      text,
    currency              text NOT NULL DEFAULT 'USD',
    subtotal              numeric(14,2) NOT NULL DEFAULT 0,
    tax_amount            numeric(14,2) NOT NULL DEFAULT 0,
    total_amount          numeric(14,2) NOT NULL DEFAULT 0,
    status                entity_status NOT NULL DEFAULT 'draft',
    approval_status       text NOT NULL DEFAULT 'pending' CHECK (approval_status IN ('pending', 'approved', 'rejected')),
    approved_by           text,
    approved_at           timestamptz,
    notes                 text,
    created_by            text,
    updated_by            text,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, po_number)
);

-- Purchase order line items
CREATE TABLE po_line_items (
    id                uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    po_id             uuid NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id        uuid NOT NULL REFERENCES products(id),
    line_number       int NOT NULL,
    quantity_ordered  numeric(12,3) NOT NULL,
    quantity_received numeric(12,3) NOT NULL DEFAULT 0,
    unit_price        numeric(12,2) NOT NULL,
    line_total        numeric(14,2),
    discount_percent  numeric(5,2) NOT NULL DEFAULT 0,
    tax_percent       numeric(5,2) NOT NULL DEFAULT 0,
    scheduled_date    date,
    status            entity_status NOT NULL DEFAULT 'active',
    notes             text
);

-- ─── 4. WAREHOUSE ENHANCEMENTS ──────────────────────────────────────────

-- Per-location environmental constraints
CREATE TABLE location_constraints (
    location_id         uuid PRIMARY KEY REFERENCES locations(id) ON DELETE CASCADE,
    min_temperature     numeric(5,2),
    max_temperature     numeric(5,2),
    min_humidity        numeric(5,2),
    max_humidity        numeric(5,2),
    is_hazardous_allowed boolean NOT NULL DEFAULT false,
    is_food_grade       boolean NOT NULL DEFAULT false,
    is_pharma_grade     boolean NOT NULL DEFAULT false,
    max_weight_capacity numeric(12,2),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- Warehouse documents (licenses, permits, certificates)
CREATE TABLE warehouse_documents (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    warehouse_id  uuid NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    document_type text NOT NULL,
    document_name text NOT NULL,
    file_url      text NOT NULL,
    expiry_date   date,
    uploaded_by   text,
    created_at    timestamptz NOT NULL DEFAULT now()
);

-- ─── 5. STOCK ENHANCEMENTS ──────────────────────────────────────────────

-- Periodic stock snapshots for AMC calculation
CREATE TABLE stock_snapshots (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    product_id      uuid NOT NULL REFERENCES products(id),
    warehouse_id    uuid NOT NULL REFERENCES warehouses(id),
    location_id     uuid REFERENCES locations(id),
    batch_id        uuid REFERENCES batches(id),
    quantity_on_hand numeric(12,3) NOT NULL,
    quantity_reserved numeric(12,3) NOT NULL DEFAULT 0,
    snapshot_date   date NOT NULL DEFAULT CURRENT_DATE,
    snapshot_type   text NOT NULL DEFAULT 'daily' CHECK (snapshot_type IN ('daily', 'weekly', 'monthly', 'manual')),
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- ─── 6. TRANSFER ENHANCEMENTS ───────────────────────────────────────────

-- Transfer line items (replaces inline line items)
CREATE TABLE transfer_line_items (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    transfer_id uuid NOT NULL REFERENCES stock_transfers(id) ON DELETE CASCADE,
    product_id  uuid NOT NULL REFERENCES products(id),
    batch_id    uuid REFERENCES batches(id),
    quantity    numeric(12,3) NOT NULL,
    unit_cost   numeric(12,2),
    status      entity_status NOT NULL DEFAULT 'active'
);

-- Adjustment reason code catalog
CREATE TABLE adjustment_reason_codes (
    code              text PRIMARY KEY,
    category          text NOT NULL,
    description       text NOT NULL,
    requires_approval boolean NOT NULL DEFAULT false,
    is_active         boolean NOT NULL DEFAULT true,
    created_at        timestamptz NOT NULL DEFAULT now()
);

-- Adjustment line items (replaces inline)
CREATE TABLE adjustment_line_items (
    id                uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    adjustment_id     uuid NOT NULL REFERENCES stock_adjustments(id) ON DELETE CASCADE,
    product_id        uuid NOT NULL REFERENCES products(id),
    batch_id          uuid REFERENCES batches(id),
    location_id       uuid REFERENCES locations(id),
    expected_quantity numeric(12,3) NOT NULL,
    actual_quantity   numeric(12,3) NOT NULL,
    difference        numeric(12,3) NOT NULL,
    reason_code       text REFERENCES adjustment_reason_codes(code),
    unit_cost         numeric(12,2),
    notes             text,
    created_at        timestamptz NOT NULL DEFAULT now()
);

-- ─── 7. QA ENHANCEMENTS ─────────────────────────────────────────────────

-- QA checklist templates
CREATE TABLE qa_checklist_templates (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    code        text NOT NULL,
    name        text NOT NULL,
    description text,
    category    text NOT NULL,
    is_mandatory boolean NOT NULL DEFAULT false,
    status      entity_status NOT NULL DEFAULT 'active',
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

-- QA checklist items (questions)
CREATE TABLE qa_checklist_items (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    template_id     uuid NOT NULL REFERENCES qa_checklist_templates(id) ON DELETE CASCADE,
    item_order      int NOT NULL,
    question        text NOT NULL,
    expected_answer text NOT NULL,
    is_critical     boolean NOT NULL DEFAULT false,
    weight          numeric(3,2) NOT NULL DEFAULT 1.0
);

-- QA checklist results (actual inspection answers)
CREATE TABLE qa_checklist_results (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    inspection_id   uuid NOT NULL REFERENCES qa_inspections(id) ON DELETE CASCADE,
    checklist_item_id uuid NOT NULL REFERENCES qa_checklist_items(id),
    answer          text,
    score           numeric(3,2),
    notes           text,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- QA dispositions (approve/reject outcome of inspection)
CREATE TABLE qa_dispositions (
    id                      uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    inspection_id           uuid NOT NULL REFERENCES qa_inspections(id) ON DELETE CASCADE,
    disposition_type        text NOT NULL CHECK (disposition_type IN ('pass', 'fail', 'quarantine', 'rework', 'partial')),
    disposition_date        date NOT NULL DEFAULT CURRENT_DATE,
    approved_by             text,
    destination_location_id uuid REFERENCES locations(id),
    notes                   text,
    created_by              text,
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now()
);

-- ─── 8. DISTRIBUTION ────────────────────────────────────────────────────

-- Distributions to beneficiaries
CREATE TABLE distributions (
    id                uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id            uuid NOT NULL REFERENCES organizations(id),
    org_level_id      uuid REFERENCES org_levels(id),
    program_id        uuid REFERENCES programs(id),
    distribution_number text NOT NULL,
    distribution_date date NOT NULL DEFAULT CURRENT_DATE,
    location          text,
    beneficiary_count int,
    status            entity_status NOT NULL DEFAULT 'draft',
    notes             text,
    created_by        text,
    updated_by        text,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, distribution_number)
);

-- Distribution line items
CREATE TABLE distribution_line_items (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    distribution_id     uuid NOT NULL REFERENCES distributions(id) ON DELETE CASCADE,
    product_id          uuid NOT NULL REFERENCES products(id),
    batch_id            uuid REFERENCES batches(id),
    quantity_planned    numeric(12,3) NOT NULL,
    quantity_distributed numeric(12,3) NOT NULL DEFAULT 0,
    unit_cost           numeric(12,2),
    status              entity_status NOT NULL DEFAULT 'active'
);

-- Distribution beneficiary records
CREATE TABLE distribution_beneficiaries (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    distribution_id uuid NOT NULL REFERENCES distributions(id) ON DELETE CASCADE,
    beneficiary_type text NOT NULL,
    count           int NOT NULL,
    criteria        text
);

-- ─── 9. RETURNS & DISPOSAL ──────────────────────────────────────────────

-- Disposal method catalog
CREATE TABLE disposal_methods (
    code                text PRIMARY KEY,
    name                text NOT NULL,
    description         text,
    requires_witness    boolean NOT NULL DEFAULT false,
    environmental_impact text,
    is_active           boolean NOT NULL DEFAULT true
);

-- Stock returns
CREATE TABLE stock_returns (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    warehouse_id  uuid REFERENCES warehouses(id),
    return_number text NOT NULL,
    returned_from text,
    return_date   date NOT NULL DEFAULT CURRENT_DATE,
    return_type   text CHECK (return_type IN ('beneficiary', 'program', 'supplier', 'internal')),
    reason        text,
    status        entity_status NOT NULL DEFAULT 'draft',
    notes         text,
    created_by    text,
    updated_by    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, return_number)
);

-- Return line items
CREATE TABLE return_line_items (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    return_id   uuid NOT NULL REFERENCES stock_returns(id) ON DELETE CASCADE,
    product_id  uuid NOT NULL REFERENCES products(id),
    batch_id    uuid REFERENCES batches(id),
    quantity    numeric(12,3) NOT NULL,
    condition   text CHECK (condition IN ('unopened', 'damaged', 'expired', 'near_expiry')),
    disposition text,
    unit_cost   numeric(12,2)
);

-- Disposals
CREATE TABLE disposals (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id              uuid NOT NULL REFERENCES organizations(id),
    warehouse_id        uuid REFERENCES warehouses(id),
    disposal_number     text NOT NULL,
    disposal_date       date NOT NULL DEFAULT CURRENT_DATE,
    disposal_method_code text REFERENCES disposal_methods(code),
    authorized_by       text,
    witness_by          text,
    reason              text,
    status              entity_status NOT NULL DEFAULT 'draft',
    created_by          text,
    updated_by          text,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, disposal_number)
);

-- Disposal line items
CREATE TABLE disposal_line_items (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    disposal_id uuid NOT NULL REFERENCES disposals(id) ON DELETE CASCADE,
    product_id  uuid NOT NULL REFERENCES products(id),
    batch_id    uuid REFERENCES batches(id),
    quantity    numeric(12,3) NOT NULL,
    unit_cost   numeric(12,2),
    notes       text
);

-- ─── 10. PHYSICAL COUNT ─────────────────────────────────────────────────

-- Physical stock counts
CREATE TABLE stock_counts (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    warehouse_id uuid NOT NULL REFERENCES warehouses(id),
    count_number text NOT NULL,
    count_date  date NOT NULL DEFAULT CURRENT_DATE,
    count_type  text CHECK (count_type IN ('full', 'cycle', 'spot', 'annual')),
    status      entity_status NOT NULL DEFAULT 'draft',
    counted_by  text,
    verified_by text,
    notes       text,
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, count_number)
);

-- Count line items
CREATE TABLE count_line_items (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    count_id            uuid NOT NULL REFERENCES stock_counts(id) ON DELETE CASCADE,
    product_id          uuid NOT NULL REFERENCES products(id),
    location_id         uuid REFERENCES locations(id),
    batch_id            uuid REFERENCES batches(id),
    expected_quantity   numeric(12,3) NOT NULL,
    counted_quantity    numeric(12,3) NOT NULL,
    variance            numeric(12,3) NOT NULL,
    variance_percent    numeric(7,4),
    status              entity_status NOT NULL DEFAULT 'active',
    reviewed_by         text,
    notes               text
);

-- Count variance reconciliation
CREATE TABLE count_variance_reconciliation (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    count_line_item_id  uuid NOT NULL REFERENCES count_line_items(id) ON DELETE CASCADE,
    reconciliation_type text CHECK (reconciliation_type IN ('adjustment', 'recount', 'write_off', 'justified')),
    adjustment_id       uuid REFERENCES stock_adjustments(id),
    reason              text NOT NULL,
    approved_by         text,
    approved_at         timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now()
);

-- ─── 11. ASSET MANAGEMENT ENHANCEMENTS ──────────────────────────────────

-- Asset depreciation schedule
CREATE TABLE asset_depreciation_schedule (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    asset_id            uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    period_date         date NOT NULL,
    depreciation_amount numeric(12,2) NOT NULL,
    book_value_before   numeric(12,2) NOT NULL,
    book_value_after    numeric(12,2) NOT NULL,
    method              text NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now()
);

-- Asset attachments
CREATE TABLE asset_attachments (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    asset_id    uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    file_name   text NOT NULL,
    file_type   text,
    file_url    text NOT NULL,
    description text,
    uploaded_at timestamptz NOT NULL DEFAULT now()
);

-- ─── 12. REPLENISHMENT & FORECASTING ────────────────────────────────────

-- Average Monthly Consumption calculations
CREATE TABLE amc_calculations (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                uuid NOT NULL REFERENCES organizations(id),
    product_id            uuid NOT NULL REFERENCES products(id),
    warehouse_id          uuid NOT NULL REFERENCES warehouses(id),
    calculation_date      date NOT NULL DEFAULT CURRENT_DATE,
    amc_3_months          numeric(12,3),
    amc_6_months          numeric(12,3),
    amc_12_months         numeric(12,3),
    max_consumption       numeric(12,3),
    std_deviation         numeric(12,3),
    calculation_period_start date,
    calculation_period_end   date,
    calculation_status    text NOT NULL DEFAULT 'completed' CHECK (calculation_status IN ('pending', 'completed', 'failed')),
    created_at            timestamptz NOT NULL DEFAULT now()
);

-- Reorder recommendations
CREATE TABLE reorder_recommendations (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id              uuid NOT NULL REFERENCES organizations(id),
    product_id          uuid NOT NULL REFERENCES products(id),
    warehouse_id        uuid NOT NULL REFERENCES warehouses(id),
    recommendation_date date NOT NULL DEFAULT CURRENT_DATE,
    current_stock       numeric(12,3) NOT NULL,
    reorder_point       numeric(12,3) NOT NULL,
    reorder_quantity    numeric(12,3) NOT NULL,
    lead_time_days      int,
    amc_used            numeric(12,3),
    safety_stock        numeric(12,3),
    recommendation_type text CHECK (recommendation_type IN ('reorder', 'excess', 'normal', 'critical')),
    priority            text CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    notes               text,
    reviewed            boolean NOT NULL DEFAULT false,
    created_at          timestamptz NOT NULL DEFAULT now()
);

-- ML engine forecast results
CREATE TABLE forecast_results (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    product_id      uuid NOT NULL REFERENCES products(id),
    warehouse_id    uuid NOT NULL REFERENCES warehouses(id),
    forecast_date   date NOT NULL,
    forecast_value  numeric(12,3) NOT NULL,
    lower_bound     numeric(12,3),
    upper_bound     numeric(12,3),
    confidence_level numeric(5,2) NOT NULL DEFAULT 95.0,
    model_version   text,
    features_used   jsonb,
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, product_id, warehouse_id, forecast_date)
);

-- ─── 13. NOTIFICATIONS & ALERTS ─────────────────────────────────────────

-- Alert configurations
CREATE TABLE alert_configurations (
    id                     uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                 uuid NOT NULL REFERENCES organizations(id),
    alert_type             text NOT NULL CHECK (alert_type IN ('expiry', 'low_stock', 'overstock', 'sleeping_stock', 'qa_hold', 'approval')),
    name                   text NOT NULL,
    description            text,
    threshold_type         text CHECK (threshold_type IN ('percentage', 'absolute', 'days', 'boolean')),
    threshold_value        text NOT NULL,
    enabled                boolean NOT NULL DEFAULT true,
    notification_channels  text NOT NULL DEFAULT 'email',
    created_by             text,
    updated_by             text,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now()
);

-- Generated alerts
CREATE TABLE alerts (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    alert_config_id uuid REFERENCES alert_configurations(id),
    alert_type      text NOT NULL,
    severity        text CHECK (severity IN ('info', 'warning', 'critical')),
    title           text NOT NULL,
    message         text NOT NULL,
    product_id      uuid REFERENCES products(id),
    warehouse_id    uuid REFERENCES warehouses(id),
    batch_id        uuid REFERENCES batches(id),
    is_read         boolean NOT NULL DEFAULT false,
    is_acknowledged boolean NOT NULL DEFAULT false,
    acknowledged_by text,
    acknowledged_at timestamptz,
    resolved_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- Alert recipients
CREATE TABLE alert_recipients (
    id                   uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    alert_config_id      uuid NOT NULL REFERENCES alert_configurations(id) ON DELETE CASCADE,
    user_id              uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_channel text NOT NULL DEFAULT 'email',
    is_active            boolean NOT NULL DEFAULT true
);

-- ─── 14. AUDIT & COMPLIANCE ─────────────────────────────────────────────

-- Audit log (application-level)
CREATE TABLE audit_log (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    user_id     uuid REFERENCES users(id),
    action      text NOT NULL,
    entity_type text NOT NULL,
    entity_id   text,
    changes     jsonb,
    ip_address  text,
    user_agent  text,
    timestamp   timestamptz NOT NULL DEFAULT now()
);

-- Data change log (for offline sync conflict resolution)
CREATE TABLE data_change_log (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    table_name      text NOT NULL,
    record_id       uuid NOT NULL,
    operation       text NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values      jsonb,
    new_values      jsonb,
    changed_by      text NOT NULL,
    change_timestamp timestamptz NOT NULL DEFAULT now()
);

-- ─── 15. OFFLINE SYNC ───────────────────────────────────────────────────

-- Sync log
CREATE TABLE sync_log (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    device_id       text NOT NULL,
    user_id         uuid REFERENCES users(id),
    sync_type       text CHECK (sync_type IN ('push', 'pull', 'full')),
    started_at      timestamptz NOT NULL DEFAULT now(),
    completed_at    timestamptz,
    status          text CHECK (status IN ('in_progress', 'completed', 'failed', 'partial')),
    records_pushed  int NOT NULL DEFAULT 0,
    records_pulled  int NOT NULL DEFAULT 0,
    errors_count    int NOT NULL DEFAULT 0,
    error_message   text
);

-- Sync conflicts
CREATE TABLE sync_conflicts (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    table_name      text NOT NULL,
    record_id       uuid NOT NULL,
    device_id       text NOT NULL,
    server_version  jsonb,
    client_version  jsonb,
    conflict_type   text CHECK (conflict_type IN ('update_update', 'update_delete', 'delete_update')),
    resolution      text NOT NULL DEFAULT 'pending' CHECK (resolution IN ('server_wins', 'client_wins', 'manual', 'pending')),
    resolved_by     text,
    resolved_at     timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- ─── 16. REPORTS ────────────────────────────────────────────────────────

-- Report definitions
CREATE TABLE report_definitions (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    code          text NOT NULL,
    name          text NOT NULL,
    description   text,
    category      text NOT NULL,
    sql_query     text,
    parameters    jsonb,
    output_formats text NOT NULL DEFAULT 'pdf,csv,xlsx',
    is_system     boolean NOT NULL DEFAULT false,
    status        entity_status NOT NULL DEFAULT 'active',
    created_by    text,
    updated_by    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

-- Report schedules
CREATE TABLE report_schedules (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    report_id     uuid NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
    schedule_cron text NOT NULL,
    parameters    jsonb,
    recipients    text,
    format        text NOT NULL DEFAULT 'pdf',
    is_active     boolean NOT NULL DEFAULT true,
    last_run_at   timestamptz,
    next_run_at   timestamptz,
    created_by    text,
    updated_by    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- ─── INDEXES ────────────────────────────────────────────────────────────

-- Composite indexes on (org_id, created_at DESC) for major transaction tables
CREATE INDEX idx_purchase_orders_org_created ON purchase_orders(org_id, created_at DESC);
CREATE INDEX idx_distributions_org_created ON distributions(org_id, created_at DESC);
CREATE INDEX idx_stock_returns_org_created ON stock_returns(org_id, created_at DESC);
CREATE INDEX idx_disposals_org_created ON disposals(org_id, created_at DESC);
CREATE INDEX idx_stock_counts_org_created ON stock_counts(org_id, created_at DESC);
CREATE INDEX idx_stock_snapshots_org_created ON stock_snapshots(org_id, created_at DESC);
CREATE INDEX idx_suppliers_org_created ON suppliers(org_id, created_at DESC);
CREATE INDEX idx_audit_log_org_timestamp ON audit_log(org_id, timestamp DESC);
CREATE INDEX idx_data_change_log_org_timestamp ON data_change_log(org_id, change_timestamp DESC);
CREATE INDEX idx_sync_log_org_created ON sync_log(org_id, created_at DESC);

-- Index on (product_id, warehouse_id) for stock_levels
CREATE INDEX idx_stock_levels_product_warehouse ON stock_levels(product_id, warehouse_id);

-- Index on expiry_date for batches (FEFO)
CREATE INDEX idx_batches_expiry_date ON batches(expiry_date);

-- Index on movement_type for stock_movements
CREATE INDEX idx_stock_movements_movement_type ON stock_movements(movement_type);

-- Index on (ref_doc_type, ref_doc_id) for stock_movements
CREATE INDEX idx_stock_movements_ref_doc ON stock_movements(ref_doc_type, ref_doc_id);

-- Index on (org_id, alert_type) for alerts
CREATE INDEX idx_alerts_org_type ON alerts(org_id, alert_type);

-- Index on (org_id, product_id) for reorder_recommendations
CREATE INDEX idx_reorder_recommendations_org_product ON reorder_recommendations(org_id, product_id);

-- GIN index on audit_log(changes) for JSONB queries
CREATE INDEX idx_audit_log_changes ON audit_log USING GIN (changes);

-- GIN index with jsonb_path_ops for audit_log
CREATE INDEX idx_audit_log_changes_path ON audit_log USING GIN (changes jsonb_path_ops);

-- ─── RLS POLICIES ──────────────────────────────────────────────────────

-- Enable RLS on all new tenant-scoped tables
ALTER TABLE functions ENABLE ROW LEVEL SECURITY;
ALTER TABLE suppliers ENABLE ROW LEVEL SECURITY;
ALTER TABLE supplier_contracts ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE qa_checklist_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE distributions ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_returns ENABLE ROW LEVEL SECURITY;
ALTER TABLE disposals ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_counts ENABLE ROW LEVEL SECURITY;
ALTER TABLE amc_calculations ENABLE ROW LEVEL SECURITY;
ALTER TABLE reorder_recommendations ENABLE ROW LEVEL SECURITY;
ALTER TABLE forecast_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_configurations ENABLE ROW LEVEL SECURITY;
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE data_change_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE sync_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE sync_conflicts ENABLE ROW LEVEL SECURITY;
ALTER TABLE report_definitions ENABLE ROW LEVEL SECURITY;

-- Apply RLS policies for all new tenant-scoped tables
DO $$
DECLARE
  tbl text;
BEGIN
  FOR tbl IN SELECT tablename FROM pg_tables WHERE tablename IN (
    'functions', 'suppliers', 'supplier_contracts', 'purchase_orders',
    'stock_snapshots', 'qa_checklist_templates', 'distributions',
    'stock_returns', 'disposals', 'stock_counts',
    'amc_calculations', 'reorder_recommendations', 'forecast_results',
    'alert_configurations', 'alerts', 'audit_log', 'data_change_log',
    'sync_log', 'sync_conflicts', 'report_definitions'
  ) LOOP
    EXECUTE format('DROP POLICY IF EXISTS org_isolation ON %I', tbl);
    EXECUTE format('
      CREATE POLICY org_isolation ON %I
        USING (org_id = app.current_org_id()::uuid)
        WITH CHECK (org_id = app.current_org_id()::uuid)
    ', tbl);
  END LOOP;
END;
$$;
