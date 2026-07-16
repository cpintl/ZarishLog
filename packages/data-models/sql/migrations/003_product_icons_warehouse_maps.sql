-- ZarishLog — Product Icons, Warehouse Maps & Extended Fields
-- PostgreSQL 18 — Adds dosage forms (icon lookup), product metadata,
-- warehouse geo/contact fields, user profile fields, organization branding,
-- stock movement audit fields, and QA inspection scoring.

-- ─── 1. DOSAGE FORMS LOOKUP TABLE ────────────────────────────────────────

CREATE TABLE IF NOT EXISTS dosage_forms (
    code                text PRIMARY KEY,
    name                text NOT NULL,
    icon                text NOT NULL,
    category            text NOT NULL CHECK (category IN ('solid', 'liquid', 'semi_solid', 'gas', 'device')),
    is_parenteral       boolean DEFAULT false,
    requires_cold_chain boolean DEFAULT false,
    sort_order          int DEFAULT 0
);

INSERT INTO dosage_forms AS d (code, name, icon, category, is_parenteral, requires_cold_chain, sort_order)
VALUES
    ('TAB',     'Tablet',             '💊', 'solid',      false, false, 1),
    ('CAP',     'Capsule',            '💊', 'solid',      false, false, 2),
    ('INJ',     'Injection',          '💉', 'liquid',     true,  false, 3),
    ('SYR',     'Syrup',              '🧪', 'liquid',     false, false, 4),
    ('SUS',     'Suspension',         '🧪', 'liquid',     false, false, 5),
    ('CRM',     'Cream',              '🧴', 'semi_solid', false, false, 6),
    ('ONT',     'Ointment',           '🧴', 'semi_solid', false, false, 7),
    ('EYE',     'Eye Drops',          '🧪', 'liquid',     false, false, 8),
    ('EAR',     'Ear Drops',          '🧪', 'liquid',     false, false, 9),
    ('NAS',     'Nasal Spray',        '💨', 'liquid',     false, false, 10),
    ('INH',     'Inhaler',            '💨', 'gas',        false, false, 11),
    ('SUP',     'Suppository',        '💊', 'semi_solid', false, false, 12),
    ('PAT',     'Patch',              '🩹', 'solid',      false, false, 13),
    ('SOL',     'Oral Solution',      '🧪', 'liquid',     false, false, 14),
    ('PWD',     'Powder',             '📦', 'solid',      false, false, 15),
    ('GRN',     'Granules',           '📦', 'solid',      false, false, 16),
    ('SPR',     'Spray',              '💨', 'liquid',     false, false, 17),
    ('LOT',     'Lotion',             '🧴', 'liquid',     false, false, 18),
    ('GEL',     'Gel',                '🧴', 'semi_solid', false, false, 19),
    ('WAF',     'Wafer',              '💊', 'solid',      false, false, 20),
    ('IMP',     'Implant',            '💉', 'solid',      true,  false, 21),
    ('TAB-EFF', 'Effervescent Tablet','💊', 'solid',      false, false, 22)
ON CONFLICT (code) DO UPDATE SET
    name                = EXCLUDED.name,
    icon                = EXCLUDED.icon,
    category            = EXCLUDED.category,
    is_parenteral       = EXCLUDED.is_parenteral,
    requires_cold_chain = EXCLUDED.requires_cold_chain,
    sort_order          = EXCLUDED.sort_order;

-- ─── 2. PRODUCTS — EXTENDED FIELDS ───────────────────────────────────────

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='dosage_form_code') THEN
        ALTER TABLE products ADD COLUMN dosage_form_code text REFERENCES dosage_forms(code);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='generic_name') THEN
        ALTER TABLE products ADD COLUMN generic_name text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='brand_name') THEN
        ALTER TABLE products ADD COLUMN brand_name text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='reference_urls') THEN
        ALTER TABLE products ADD COLUMN reference_urls jsonb DEFAULT '[]'::jsonb;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='storage_conditions') THEN
        ALTER TABLE products ADD COLUMN storage_conditions text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='reorder_formula') THEN
        ALTER TABLE products ADD COLUMN reorder_formula jsonb DEFAULT '{}'::jsonb;
    END IF;
END;
$$;

-- ─── 3. WAREHOUSES — LOCATION & CONTACT FIELDS ───────────────────────────

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='latitude') THEN
        ALTER TABLE warehouses ADD COLUMN latitude numeric(10,7);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='longitude') THEN
        ALTER TABLE warehouses ADD COLUMN longitude numeric(10,7);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='google_maps_url') THEN
        ALTER TABLE warehouses ADD COLUMN google_maps_url text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='contact_phone') THEN
        ALTER TABLE warehouses ADD COLUMN contact_phone text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='operating_hours') THEN
        ALTER TABLE warehouses ADD COLUMN operating_hours text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='has_generator') THEN
        ALTER TABLE warehouses ADD COLUMN has_generator boolean DEFAULT false;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='has_cctv') THEN
        ALTER TABLE warehouses ADD COLUMN has_cctv boolean DEFAULT false;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='has_fire_system') THEN
        ALTER TABLE warehouses ADD COLUMN has_fire_system boolean DEFAULT false;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='warehouses' AND column_name='security_guard') THEN
        ALTER TABLE warehouses ADD COLUMN security_guard boolean DEFAULT false;
    END IF;
END;
$$;

-- ─── 4. USERS — PROFILE FIELDS ───────────────────────────────────────────

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='phone') THEN
        ALTER TABLE users ADD COLUMN phone text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='job_title') THEN
        ALTER TABLE users ADD COLUMN job_title text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='department_id') THEN
        ALTER TABLE users ADD COLUMN department_id uuid REFERENCES departments(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='avatar_url') THEN
        ALTER TABLE users ADD COLUMN avatar_url text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='signature_url') THEN
        ALTER TABLE users ADD COLUMN signature_url text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='last_login_at') THEN
        ALTER TABLE users ADD COLUMN last_login_at timestamptz;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='password_changed_at') THEN
        ALTER TABLE users ADD COLUMN password_changed_at timestamptz;
    END IF;
END;
$$;

-- ─── 5. ORGANIZATIONS — BRANDING FIELDS ──────────────────────────────────

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='organizations' AND column_name='logo_url') THEN
        ALTER TABLE organizations ADD COLUMN logo_url text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='organizations' AND column_name='website') THEN
        ALTER TABLE organizations ADD COLUMN website text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='organizations' AND column_name='country') THEN
        ALTER TABLE organizations ADD COLUMN country text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='organizations' AND column_name='default_currency') THEN
        ALTER TABLE organizations ADD COLUMN default_currency text DEFAULT 'USD';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='organizations' AND column_name='timezone') THEN
        ALTER TABLE organizations ADD COLUMN timezone text DEFAULT 'UTC';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='organizations' AND column_name='date_format') THEN
        ALTER TABLE organizations ADD COLUMN date_format text DEFAULT 'YYYY-MM-DD';
    END IF;
END;
$$;

-- ─── 6. STOCK MOVEMENTS — AUDIT FIELDS ───────────────────────────────────

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='stock_movements' AND column_name='source_document_url') THEN
        ALTER TABLE stock_movements ADD COLUMN source_document_url text;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='stock_movements' AND column_name='is_offline_sync') THEN
        ALTER TABLE stock_movements ADD COLUMN is_offline_sync boolean DEFAULT false;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='stock_movements' AND column_name='device_id') THEN
        ALTER TABLE stock_movements ADD COLUMN device_id text;
    END IF;
END;
$$;

-- ─── 7. QA INSPECTIONS — SCORING FIELDS ──────────────────────────────────

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='qa_inspections' AND column_name='checklist_template_id') THEN
        ALTER TABLE qa_inspections ADD COLUMN checklist_template_id uuid REFERENCES qa_checklist_templates(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='qa_inspections' AND column_name='overall_score') THEN
        ALTER TABLE qa_inspections ADD COLUMN overall_score numeric(5,2);
    END IF;
END;
$$;

-- ─── 8. INDEXES ──────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_products_generic_name ON products(generic_name);
CREATE INDEX IF NOT EXISTS idx_products_dosage_form_code ON products(dosage_form_code);
CREATE INDEX IF NOT EXISTS idx_warehouses_lat_lng ON warehouses(latitude, longitude);
