-- ZarishLog — Master Catalogue Extension (Phase 5)
-- PostgreSQL 18 — Adds Master Catalogue fields from SPINCO/MSF/UNSPSC standards,
-- location type 'aisle', stock statuses 'committed'/'backordered',
-- justification codes, conversion factors, and product master attributes.
-- Idempotent: uses IF NOT EXISTS / DO $$ blocks throughout.

-- ─── 1. EXTEND ENUMS ──────────────────────────────────────────────────────

-- Add 'aisle' to location_type
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'aisle' AND enumtypid = 'location_type'::regtype) THEN
    ALTER TYPE location_type ADD VALUE 'aisle' AFTER 'zone';
  END IF;
END;
$$;

-- Add 'committed' and 'backordered' to stock_status
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'committed' AND enumtypid = 'stock_status'::regtype) THEN
    ALTER TYPE stock_status ADD VALUE 'committed' AFTER 'reserved';
  END IF;
END;
$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'backordered' AND enumtypid = 'stock_status'::regtype) THEN
    ALTER TYPE stock_status ADD VALUE 'backordered' AFTER 'in_transit';
  END IF;
END;
$$;

-- ─── 2. PRODUCTS — ADDITIONAL MASTER CATALOGUE FIELDS ─────────────────────

DO $$
BEGIN
  -- Strength (e.g. "500mg", "120mg/5ml")
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='strength') THEN
    ALTER TABLE products ADD COLUMN strength text;
  END IF;

  -- Full UNSPSC commodity code (8-10 digits)
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='unspsc_commodity') THEN
    ALTER TABLE products ADD COLUMN unspsc_commodity text;
  END IF;

  -- eClass classification code
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='eclass_code') THEN
    ALTER TABLE products ADD COLUMN eclass_code text;
  END IF;

  -- Alternative codes (MSF Code, Manufacturer ID, Local Code, etc.)
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='alternate_codes') THEN
    ALTER TABLE products ADD COLUMN alternate_codes jsonb DEFAULT '{}'::jsonb;
  END IF;

  -- Is this product tracked as an individual asset?
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='is_asset') THEN
    ALTER TABLE products ADD COLUMN is_asset boolean NOT NULL DEFAULT false;
  END IF;

  -- Replenishment strategy
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='replenishment_type') THEN
    ALTER TABLE products ADD COLUMN replenishment_type text CHECK (replenishment_type IN ('min_max', 'kanban', 'jit', 'periodic', 'one_time'));
  END IF;

  -- Inventory valuation method
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='valuation_method') THEN
    ALTER TABLE products ADD COLUMN valuation_method text CHECK (valuation_method IN ('fifo', 'avco', 'standard')) DEFAULT 'fifo';
  END IF;

  -- Is this a kit/bundle of items?
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='is_kitting') THEN
    ALTER TABLE products ADD COLUMN is_kitting boolean NOT NULL DEFAULT false;
  END IF;

  -- Recommended storage temperature range
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='temp_min_c') THEN
    ALTER TABLE products ADD COLUMN temp_min_c numeric(5,2);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='temp_max_c') THEN
    ALTER TABLE products ADD COLUMN temp_max_c numeric(5,2);
  END IF;

  -- MSF/WHO Essential Medicines List flag
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='is_essential') THEN
    ALTER TABLE products ADD COLUMN is_essential boolean DEFAULT false;
  END IF;

  -- Controlled/narcotic substance flag
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='is_controlled') THEN
    ALTER TABLE products ADD COLUMN is_controlled boolean DEFAULT false;
  END IF;
END;
$$;

-- ─── 3. UNITS OF MEASURE — ADD CONVERSION FACTOR ──────────────────────────

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='units_of_measure' AND column_name='conversion_factor') THEN
    ALTER TABLE units_of_measure ADD COLUMN conversion_factor numeric(12,6) DEFAULT 1.0;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='units_of_measure' AND column_name='base_uom_id') THEN
    ALTER TABLE units_of_measure ADD COLUMN base_uom_id uuid REFERENCES units_of_measure(id);
  END IF;
END;
$$;

-- ─── 4. JUSTIFICATION CODES LOOKUP TABLE ──────────────────────────────────

CREATE TABLE IF NOT EXISTS justification_codes (
    code        text PRIMARY KEY,
    name        text NOT NULL,
    description text,
    category    text NOT NULL CHECK (category IN ('recurring', 'campaign', 'emergency', 'forecast', 'asset', 'special')),
    sort_order  int NOT NULL DEFAULT 0,
    is_active   boolean NOT NULL DEFAULT true
);

-- ─── 5. CONVERSION FACTOR ON PRODUCT PACKAGING ────────────────────────────

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='product_packaging' AND column_name='conversion_factor') THEN
    ALTER TABLE product_packaging ADD COLUMN conversion_factor numeric(12,6);
  END IF;
END;
$$;

-- ─── 6. ENTITY HIERARCHY (Function → Entity → Attribute) ───────────────────

-- Entities within a function (e.g., Warehouse under LOG-WHS, Vehicle under LOG-TRN)
CREATE TABLE IF NOT EXISTS entities (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id        uuid NOT NULL REFERENCES organizations(id),
    function_id   uuid REFERENCES functions(id),
    entity_type   text NOT NULL,
    code          text NOT NULL,
    name          text NOT NULL,
    description   text,
    parent_id     uuid REFERENCES entities(id),
    status        entity_status NOT NULL DEFAULT 'active',
    metadata      jsonb DEFAULT '{}'::jsonb,
    created_by    text,
    updated_by    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, code)
);

-- Dynamic key-value attributes for entities
CREATE TABLE IF NOT EXISTS entity_attributes (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    entity_id       uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    attribute_name  text NOT NULL,
    attribute_value text,
    attribute_type  text NOT NULL DEFAULT 'text' CHECK (attribute_type IN ('text', 'number', 'boolean', 'date', 'json')),
    uom_id          uuid REFERENCES units_of_measure(id),
    sort_order      int NOT NULL DEFAULT 0,
    is_required     boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE(entity_id, attribute_name)
);

-- ─── 7. INDEXES FOR NEW FIELDS & TABLES ───────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_products_strength ON products(strength);
CREATE INDEX IF NOT EXISTS idx_products_is_asset ON products(is_asset) WHERE is_asset = true;
CREATE INDEX IF NOT EXISTS idx_products_unspsc_commodity ON products(unspsc_commodity);
CREATE INDEX IF NOT EXISTS idx_products_is_essential ON products(is_essential) WHERE is_essential = true;
CREATE INDEX IF NOT EXISTS idx_products_is_controlled ON products(is_controlled) WHERE is_controlled = true;
CREATE INDEX IF NOT EXISTS idx_entities_org_id ON entities(org_id);
CREATE INDEX IF NOT EXISTS idx_entities_function_id ON entities(function_id);
CREATE INDEX IF NOT EXISTS idx_entity_attributes_entity_id ON entity_attributes(entity_id);

-- ─── 8. RLS POLICIES FOR NEW TABLES ───────────────────────────────────────

ALTER TABLE IF EXISTS entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS entity_attributes ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  -- entities
  IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'entities') THEN
    DROP POLICY IF EXISTS org_isolation ON entities;
    EXECUTE format('
      CREATE POLICY org_isolation ON entities
        USING (org_id = app.current_org_id()::uuid)
        WITH CHECK (org_id = app.current_org_id()::uuid)
    ');
  END IF;
  -- entity_attributes (has no org_id directly; relies on entity join — skip org_isolation)
END;
$$;
