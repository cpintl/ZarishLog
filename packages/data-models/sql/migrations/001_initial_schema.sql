-- ZarishLog — Initial Schema
-- PostgreSQL 18, Multi-tenant, UUIDv7, Audit columns, RLS

-- Enable UUIDv7 generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Helper: generate UUID v7 (time-ordered)
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS uuid
LANGUAGE plpgsql
AS $$
BEGIN
  RETURN encode(
    set_byte(
      set_byte(
        set_byte(
          set_byte(
            overlay(
              decode(lpad(to_hex(floor(extract(epoch from clock_timestamp()) * 1000)::bigint), 12, '0'), 'hex')
              placing '\x0b' from 7
            )
            , 4, 1, (get_byte(decode(lpad(to_hex(floor(extract(epoch from clock_timestamp()) * 1000)::bigint), 12, '0'), 'hex'), 4) & 0x0f) | 0x70)
          , 0, 1, (get_byte(decode(lpad(to_hex(floor(extract(epoch from clock_timestamp()) * 1000)::bigint), 12, '0'), 'hex'), 0) & 0x7f) | 0x80)
        , 7, 1, (random() * 255)::int)
      , 1, 1, (random() * 255)::int)
    , 0, 1, (random() * 255)::int)
    , 'hex'
  )::uuid;
END;
$$;

-- Set session-level org_id for RLS
CREATE OR REPLACE FUNCTION app.current_org_id() RETURNS text
LANGUAGE plpgsql STABLE PARALLEL SAFE
AS $$
BEGIN
  RETURN current_setting('app.current_org_id', true);
END;
$$;

CREATE OR REPLACE FUNCTION app.current_user_id() RETURNS text
LANGUAGE plpgsql STABLE PARALLEL SAFE
AS $$
BEGIN
  RETURN current_setting('app.current_user_id', true);
END;
$$;

-- ─── Reference Data ────────────────────────────────────────────────────

CREATE TYPE uom_category AS ENUM ('count', 'weight', 'volume', 'length', 'time', 'area');
CREATE TYPE item_type AS ENUM ('drug', 'medical_supply', 'equipment', 'instrument', 'material', 'vaccine', 'nutrition', 'lab_reagent', 'asset', 'consumable');
CREATE TYPE entity_status AS ENUM ('active', 'inactive', 'draft', 'archived');
CREATE TYPE warehouse_type AS ENUM ('central', 'sub_warehouse', 'transit', 'quarantine');
CREATE TYPE location_type AS ENUM ('zone', 'rack', 'bin', 'shelf', 'area');
CREATE TYPE movement_type AS ENUM ('receipt', 'issue', 'transfer_in', 'transfer_out', 'adjustment_add', 'adjustment_subtract', 'return', 'disposal');
CREATE TYPE doc_type AS ENUM ('grn', 'srf', 'transfer', 'adjustment', 'disposal', 'return');
CREATE TYPE stock_status AS ENUM ('on_hand', 'reserved', 'in_transit', 'on_hold', 'damaged', 'expired', 'quarantined', 'disposed');

-- ─── Master Data ───────────────────────────────────────────────────────

-- Organizations (Tenants)
CREATE TABLE organizations (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    name        text NOT NULL,
    code        text NOT NULL UNIQUE,
    status      entity_status NOT NULL DEFAULT 'active',
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Org Hierarchy Levels (L1 Global → L2 Country → L3 Project → L4 Site)
CREATE TABLE org_levels (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    parent_id   uuid REFERENCES org_levels(id),
    name        text NOT NULL,
    code        text NOT NULL,
    level       int NOT NULL CHECK (level BETWEEN 1 AND 4),
    status      entity_status NOT NULL DEFAULT 'active',
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

-- Programs (Thematic areas)
CREATE TABLE programs (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    code        text NOT NULL,
    name        text NOT NULL,
    description text,
    status      entity_status NOT NULL DEFAULT 'active',
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

-- Departments
CREATE TABLE departments (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    name        text NOT NULL,
    code        text NOT NULL,
    parent_id   uuid REFERENCES departments(id),
    status      entity_status NOT NULL DEFAULT 'active',
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

-- Roles
CREATE TABLE roles (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    code        text NOT NULL UNIQUE,
    name        text NOT NULL,
    description text,
    level       int NOT NULL DEFAULT 1
);

-- Users
CREATE TABLE users (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    email       text NOT NULL,
    name        text NOT NULL,
    role_id     uuid REFERENCES roles(id),
    org_level_id uuid REFERENCES org_levels(id),
    is_active   boolean NOT NULL DEFAULT true,
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, email)
);

-- Permissions
CREATE TABLE permissions (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    module      text NOT NULL,
    action      text NOT NULL,
    description text,
    UNIQUE(module, action)
);

CREATE TABLE role_permissions (
    role_id       uuid NOT NULL REFERENCES roles(id),
    permission_id uuid NOT NULL REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id)
);

-- ─── Product Catalogue ─────────────────────────────────────────────────

-- Units of Measure
CREATE TABLE units_of_measure (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    name         text NOT NULL,
    abbreviation text NOT NULL,
    category     uom_category NOT NULL,
    status       entity_status NOT NULL DEFAULT 'active',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- Product Categories
CREATE TABLE product_categories (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    parent_id   uuid REFERENCES product_categories(id),
    name        text NOT NULL,
    description text,
    unspsc      text,
    eclass      text,
    status      entity_status NOT NULL DEFAULT 'active',
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Products / Item Master
CREATE TABLE products (
    id                uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id            uuid NOT NULL REFERENCES organizations(id),
    category_id       uuid REFERENCES product_categories(id),
    uom_id            uuid REFERENCES units_of_measure(id),
    sku               text NOT NULL,
    name              text NOT NULL,
    description       text,
    item_type         item_type NOT NULL DEFAULT 'consumable',
    gtin              text,
    alternative_code  text,
    brand             text,
    manufacturer      text,
    is_batch_tracked  boolean NOT NULL DEFAULT false,
    is_serial_tracked boolean NOT NULL DEFAULT false,
    is_expiry_tracked boolean NOT NULL DEFAULT false,
    is_hazardous      boolean NOT NULL DEFAULT false,
    is_cold_chain     boolean NOT NULL DEFAULT false,
    min_stock         numeric(12,3) NOT NULL DEFAULT 0,
    max_stock         numeric(12,3) NOT NULL DEFAULT 0,
    reorder_point     numeric(12,3) NOT NULL DEFAULT 0,
    lead_time_days    int NOT NULL DEFAULT 0,
    unit_cost         numeric(12,2) NOT NULL DEFAULT 0,
    status            entity_status NOT NULL DEFAULT 'active',
    created_by        text,
    updated_by        text,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, sku)
);

CREATE INDEX idx_products_org_id ON products(org_id);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category ON products(category_id);

-- ─── Warehouse & Locations ─────────────────────────────────────────────

CREATE TABLE warehouses (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id      uuid NOT NULL REFERENCES organizations(id),
    name        text NOT NULL,
    code        text NOT NULL,
    type        warehouse_type NOT NULL DEFAULT 'central',
    address     text,
    city        text,
    country     text,
    is_active   boolean NOT NULL DEFAULT true,
    created_by  text,
    updated_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

CREATE TABLE locations (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    warehouse_id  uuid NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    parent_id     uuid REFERENCES locations(id),
    code          text NOT NULL,
    name          text NOT NULL,
    type          location_type NOT NULL DEFAULT 'zone',
    is_cold_chain boolean NOT NULL DEFAULT false,
    is_hazardous  boolean NOT NULL DEFAULT false,
    is_secure     boolean NOT NULL DEFAULT false,
    max_capacity  numeric(12,3),
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(warehouse_id, code)
);

-- ─── Inventory / Stock ─────────────────────────────────────────────────

-- Batches
CREATE TABLE batches (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    product_id    uuid NOT NULL REFERENCES products(id),
    batch_number  text NOT NULL,
    serial_number text,
    expiry_date   date,
    manufacturer  text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, product_id, batch_number)
);

-- Stock levels (materialized position, derived from movements)
CREATE TABLE stock_levels (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    product_id    uuid NOT NULL REFERENCES products(id),
    warehouse_id  uuid NOT NULL REFERENCES warehouses(id),
    location_id   uuid REFERENCES locations(id),
    batch_id      uuid REFERENCES batches(id),
    quantity      numeric(12,3) NOT NULL DEFAULT 0,
    reserved_qty  numeric(12,3) NOT NULL DEFAULT 0,
    status        stock_status NOT NULL DEFAULT 'on_hand',
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(product_id, warehouse_id, location_id, batch_id)
);

CREATE INDEX idx_stock_levels_org ON stock_levels(org_id);
CREATE INDEX idx_stock_levels_product ON stock_levels(product_id);
CREATE INDEX idx_stock_levels_warehouse ON stock_levels(warehouse_id);

-- Append-only stock movement ledger
CREATE TABLE stock_movements (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    product_id    uuid NOT NULL REFERENCES products(id),
    warehouse_id  uuid NOT NULL REFERENCES warehouses(id),
    location_id   uuid REFERENCES locations(id),
    batch_id      uuid REFERENCES batches(id),
    movement_type movement_type NOT NULL,
    quantity      numeric(12,3) NOT NULL,
    ref_doc_type  doc_type,
    ref_doc_id    text,
    reason_code   text,
    reference     text,
    created_by    text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_stock_movements_org ON stock_movements(org_id);
CREATE INDEX idx_stock_movements_product ON stock_movements(product_id);
CREATE INDEX idx_stock_movements_created ON stock_movements(created_at DESC);

-- ─── Transactions ──────────────────────────────────────────────────────

-- Goods Receipt Notes
CREATE TABLE goods_receipts (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    warehouse_id  uuid NOT NULL REFERENCES warehouses(id),
    grn_number    text NOT NULL,
    supplier      text,
    po_number     text,
    received_by   text,
    status        entity_status NOT NULL DEFAULT 'draft',
    notes         text,
    created_by    text,
    updated_by    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, grn_number)
);

CREATE TABLE grn_line_items (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    grn_id        uuid NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
    product_id    uuid NOT NULL REFERENCES products(id),
    batch_number  text,
    serial_number text,
    expiry_date   date,
    quantity      numeric(12,3) NOT NULL,
    unit_cost     numeric(12,2),
    status        entity_status NOT NULL DEFAULT 'active'
);

-- Stock Issues (SRF)
CREATE TABLE stock_issues (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    warehouse_id  uuid NOT NULL REFERENCES warehouses(id),
    issue_number  text NOT NULL,
    requested_by  text,
    approved_by   text,
    program_id    uuid REFERENCES programs(id),
    department_id uuid REFERENCES departments(id),
    status        entity_status NOT NULL DEFAULT 'draft',
    notes         text,
    created_by    text,
    updated_by    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, issue_number)
);

CREATE TABLE issue_line_items (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    issue_id      uuid NOT NULL REFERENCES stock_issues(id) ON DELETE CASCADE,
    product_id    uuid NOT NULL REFERENCES products(id),
    batch_id      uuid REFERENCES batches(id),
    quantity      numeric(12,3) NOT NULL,
    unit_cost     numeric(12,2)
);

-- Stock Transfers
CREATE TABLE stock_transfers (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    from_warehouse_id uuid NOT NULL REFERENCES warehouses(id),
    to_warehouse_id   uuid NOT NULL REFERENCES warehouses(id),
    transfer_number text NOT NULL,
    status          entity_status NOT NULL DEFAULT 'draft',
    created_by      text,
    updated_by      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, transfer_number)
);

-- Stock Adjustments
CREATE TABLE stock_adjustments (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    warehouse_id  uuid NOT NULL REFERENCES warehouses(id),
    reason_code   text NOT NULL,
    description   text,
    status        entity_status NOT NULL DEFAULT 'draft',
    created_by    text,
    updated_by    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- ─── Quality Assurance ─────────────────────────────────────────────────

CREATE TABLE qa_inspections (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    grn_id          uuid REFERENCES goods_receipts(id),
    product_id      uuid NOT NULL REFERENCES products(id),
    batch_id        uuid REFERENCES batches(id),
    inspection_date date NOT NULL DEFAULT CURRENT_DATE,
    inspector       text,
    result          text NOT NULL CHECK (result IN ('pass', 'fail', 'quarantine')),
    notes           text,
    disposition     text,
    created_by      text,
    updated_by      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- ─── Asset Management ──────────────────────────────────────────────────

CREATE TABLE assets (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id          uuid NOT NULL REFERENCES organizations(id),
    asset_tag       text NOT NULL,
    name            text NOT NULL,
    description     text,
    product_id      uuid REFERENCES products(id),
    serial_number   text,
    custodian_id    uuid REFERENCES users(id),
    location_id     uuid REFERENCES locations(id),
    acquisition_date date,
    purchase_cost   numeric(12,2),
    current_value   numeric(12,2),
    depreciation_method text,
    useful_life_years int,
    status          text NOT NULL DEFAULT 'in_use' CHECK (status IN ('in_use', 'in_storage', 'under_maintenance', 'disposed', 'lost')),
    created_by      text,
    updated_by      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, asset_tag)
);

CREATE TABLE asset_custody_changes (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    asset_id      uuid NOT NULL REFERENCES assets(id),
    from_user_id  uuid REFERENCES users(id),
    to_user_id    uuid REFERENCES users(id),
    changed_by    text,
    changed_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE asset_maintenance (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    asset_id      uuid NOT NULL REFERENCES assets(id),
    maintenance_date date NOT NULL,
    description   text,
    cost          numeric(12,2),
    performed_by  text,
    next_date     date,
    created_at    timestamptz NOT NULL DEFAULT now()
);

-- ─── RLS Policies ──────────────────────────────────────────────────────

-- Enable RLS on all tenant tables
ALTER TABLE org_levels ENABLE ROW LEVEL SECURITY;
ALTER TABLE programs ENABLE ROW LEVEL SECURITY;
ALTER TABLE departments ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses ENABLE ROW LEVEL SECURITY;
ALTER TABLE locations ENABLE ROW LEVEL SECURITY;
ALTER TABLE batches ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_levels ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_movements ENABLE ROW LEVEL SECURITY;
ALTER TABLE goods_receipts ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_issues ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_transfers ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_adjustments ENABLE ROW LEVEL SECURITY;
ALTER TABLE qa_inspections ENABLE ROW LEVEL SECURITY;
ALTER TABLE assets ENABLE ROW LEVEL SECURITY;

-- RLS policy: restrict to current org
CREATE OR REPLACE FUNCTION rls_org_policy()
RETURNS text
LANGUAGE plpgsql STABLE
AS $$
BEGIN
  RETURN 'org_id = app.current_org_id()';
END;
$$;

-- Apply RLS policies
DO $$
DECLARE
  tbl text;
BEGIN
  FOR tbl IN SELECT tablename FROM pg_tables WHERE tablename IN (
    'org_levels', 'programs', 'departments', 'users', 'product_categories',
    'products', 'warehouses', 'locations', 'batches', 'stock_levels',
    'stock_movements', 'goods_receipts', 'stock_issues', 'stock_transfers',
    'stock_adjustments', 'qa_inspections', 'assets'
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
