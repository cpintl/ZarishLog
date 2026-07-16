-- ZarishLog — Comprehensive Seed Data
-- Run after schema (001_initial_schema.sql + 002_extended_schema.sql)
-- Idempotent: uses ON CONFLICT throughout for safe re-runs

BEGIN;

-- ═══════════════════════════════════════════════════════════════════════════
-- 1. ROLES
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO roles (code, name, description, level) VALUES
  ('R01', 'GLOBAL_ADMIN',        'Full system administration, all tenants', 1),
  ('R02', 'COUNTRY_REP',         'Country representative — read-only, reporting, no ops', 2),
  ('R03', 'THEME_MANAGER',       'Thematic program manager — read + reports within scope', 2),
  ('R04', 'WAREHOUSE_OFFICER',   'Full central warehouse operations — procure, QA, count', 3),
  ('R05', 'WAREHOUSE_STOREKEEPER', 'Stock operations — receive, issue, count, read', 3),
  ('R06', 'ADMIN_LOG_OFFICER',   'Office asset & logistics stock management', 3),
  ('R07', 'DEPT_MANAGER',        'Department head — stock read, reports, approval authority', 3),
  ('R08', 'DEPT_COORDINATOR',    'Sub-warehouse stock flow validation', 4),
  ('R09', 'DEPT_OFFICER',        'Day-to-day sub-warehouse operations', 4),
  ('R10', 'FIELD_WORKER',        'Field distribution staff — distributions CR', 4),
  ('R11', 'QUALITY_OFFICER',     'QA inspection staff — full QA, stock read', 3),
  ('R12', 'AUDITOR',             'Read-only audit access — read + audit + reports', 2)
ON CONFLICT (code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 2. PERMISSIONS
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO permissions (module, action) VALUES
  -- Products & catalogue
  ('products', 'create'), ('products', 'read'), ('products', 'update'), ('products', 'delete'),
  ('categories', 'create'), ('categories', 'read'), ('categories', 'update'), ('categories', 'delete'),
  -- Warehouses & locations
  ('warehouses', 'create'), ('warehouses', 'read'), ('warehouses', 'update'), ('warehouses', 'delete'),
  -- Stock operations
  ('stock', 'receive'), ('stock', 'issue'), ('stock', 'transfer'), ('stock', 'adjust'), ('stock', 'read'),
  -- Quality assurance
  ('qa', 'inspect'), ('qa', 'read'),
  -- Assets
  ('assets', 'create'), ('assets', 'read'), ('assets', 'update'), ('assets', 'transfer'),
  -- User management
  ('users', 'create'), ('users', 'read'), ('users', 'update'),
  -- Reports (existing + extended)
  ('reports', 'read'), ('reports', 'create'), ('reports', 'update'), ('reports', 'delete'), ('reports', 'schedule'),
  -- Suppliers
  ('suppliers', 'create'), ('suppliers', 'read'), ('suppliers', 'update'), ('suppliers', 'delete'),
  -- Procurement / Purchase orders
  ('procurement', 'create'), ('procurement', 'read'), ('procurement', 'update'), ('procurement', 'delete'),
  -- Distributions
  ('distributions', 'create'), ('distributions', 'read'), ('distributions', 'update'), ('distributions', 'delete'),
  -- Returns
  ('returns', 'create'), ('returns', 'read'), ('returns', 'update'), ('returns', 'delete'),
  -- Disposals
  ('disposals', 'create'), ('disposals', 'read'), ('disposals', 'update'), ('disposals', 'delete'),
  -- Stock counts
  ('stock_counts', 'create'), ('stock_counts', 'read'), ('stock_counts', 'update'), ('stock_counts', 'delete'),
  -- Alerts
  ('alerts', 'read'), ('alerts', 'acknowledge'),
  -- Sync
  ('sync', 'read'), ('sync', 'manage'),
  -- Audit
  ('audit', 'read'), ('audit', 'export')
ON CONFLICT (module, action) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 3. ROLE-PERMISSION ASSIGNMENTS
-- ═══════════════════════════════════════════════════════════════════════════

-- Helper: assign a single permission to a role
DO $$
DECLARE
  r RECORD;
  p RECORD;
  role_id_val    uuid;
  perm_id_val    uuid;
BEGIN
  -- R01 — GLOBAL_ADMIN gets ALL permissions
  SELECT id INTO role_id_val FROM roles WHERE code = 'R01';
  FOR p IN SELECT id FROM permissions LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R02 — COUNTRY_REP: read + reports (read/schedule)
  SELECT id INTO role_id_val FROM roles WHERE code = 'R02';
  FOR p IN SELECT id FROM permissions WHERE module IN ('categories','products','warehouses','stock','qa','assets','users','suppliers','procurement','distributions','returns','disposals','stock_counts','alerts','sync','audit') AND action = 'read'
    UNION ALL SELECT id FROM permissions WHERE module = 'reports' AND action IN ('read','schedule') LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R03 — THEME_MANAGER: read + reports
  SELECT id INTO role_id_val FROM roles WHERE code = 'R03';
  FOR p IN SELECT id FROM permissions WHERE module IN ('categories','products','warehouses','stock','qa','assets','suppliers','distributions','alerts') AND action = 'read'
    UNION ALL SELECT id FROM permissions WHERE module = 'reports' AND action IN ('read','create','schedule') LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R04 — WAREHOUSE_OFFICER: full warehouse ops + procure + QA + counts
  SELECT id INTO role_id_val FROM roles WHERE code = 'R04';
  FOR p IN SELECT id FROM permissions WHERE (
    module IN ('products','categories','warehouses','locations') AND action IN ('create','read','update')
  ) OR (
    module = 'stock' AND action IN ('receive','issue','transfer','adjust','read')
  ) OR (
    module IN ('suppliers','procurement') AND action IN ('create','read','update')
  ) OR (
    module = 'qa' AND action IN ('inspect','read')
  ) OR (
    module IN ('stock_counts','returns','disposals','alerts') AND action IN ('create','read','update')
  ) OR (
    module = 'reports' AND action IN ('read','create')
  ) LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R05 — WAREHOUSE_STOREKEEPER: stock receive/issue/read + counts
  SELECT id INTO role_id_val FROM roles WHERE code = 'R05';
  FOR p IN SELECT id FROM permissions WHERE (
    module = 'stock' AND action IN ('receive','issue','read')
  ) OR (
    module = 'stock_counts' AND action IN ('create','read')
  ) OR (
    module IN ('products','warehouses','locations') AND action = 'read'
  ) OR (
    module = 'alerts' AND action IN ('read','acknowledge')
  ) LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R06 — ADMIN_LOG_OFFICER: assets CRUD + stock read
  SELECT id INTO role_id_val FROM roles WHERE code = 'R06';
  FOR p IN SELECT id FROM permissions WHERE (
    module = 'assets' AND action IN ('create','read','update','transfer')
  ) OR (
    module = 'stock' AND action = 'read'
  ) OR (
    module IN ('products','categories','warehouses') AND action = 'read'
  ) OR (
    module = 'reports' AND action IN ('read','create')
  ) LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R07 — DEPT_MANAGER: stock read + reports + approve
  SELECT id INTO role_id_val FROM roles WHERE code = 'R07';
  FOR p IN SELECT id FROM permissions WHERE (
    module IN ('stock','products','categories','warehouses') AND action = 'read'
  ) OR (
    module = 'reports' AND action IN ('read','create','schedule')
  ) OR (
    module = 'procurement' AND action IN ('read','update')
  ) OR (
    module = 'alerts' AND action = 'read'
  ) LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R08 — DEPT_COORDINATOR: stock read + issue + transfers
  SELECT id INTO role_id_val FROM roles WHERE code = 'R08';
  FOR p IN SELECT id FROM permissions WHERE (
    module = 'stock' AND action IN ('read','issue','transfer')
  ) OR (
    module IN ('products','warehouses','locations') AND action = 'read'
  ) OR (
    module = 'alerts' AND action = 'read'
  ) LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R09 — DEPT_OFFICER: stock read within scope
  SELECT id INTO role_id_val FROM roles WHERE code = 'R09';
  FOR p IN SELECT id FROM permissions WHERE (
    module = 'stock' AND action = 'read'
  ) OR (
    module IN ('products','warehouses','locations') AND action = 'read'
  ) LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R10 — FIELD_WORKER: distributions create/read
  SELECT id INTO role_id_val FROM roles WHERE code = 'R10';
  FOR p IN SELECT id FROM permissions WHERE (
    module = 'distributions' AND action IN ('create','read')
  ) OR (
    module IN ('products','stock') AND action = 'read'
  ) OR (
    module = 'alerts' AND action IN ('read','acknowledge')
  ) LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R11 — QUALITY_OFFICER: QA all + stock read
  SELECT id INTO role_id_val FROM roles WHERE code = 'R11';
  FOR p IN SELECT id FROM permissions WHERE (
    module = 'qa' AND action IN ('inspect','read')
  ) OR (
    module IN ('stock','products','categories','warehouses') AND action = 'read'
  ) OR (
    module = 'reports' AND action IN ('read','create')
  ) OR (
    module = 'alerts' AND action IN ('read','acknowledge')
  ) LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;

  -- R12 — AUDITOR: read all + audit/export + reports read
  SELECT id INTO role_id_val FROM roles WHERE code = 'R12';
  FOR p IN SELECT id FROM permissions WHERE action = 'read'
    UNION ALL SELECT id FROM permissions WHERE module = 'audit' AND action IN ('read','export')
    UNION ALL SELECT id FROM permissions WHERE module = 'reports' AND action IN ('read','schedule')
    UNION ALL SELECT id FROM permissions WHERE module = 'sync' AND action = 'read' LOOP
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (role_id_val, p.id)
    ON CONFLICT DO NOTHING;
  END LOOP;
END $$;

-- ═══════════════════════════════════════════════════════════════════════════
-- 4. UNITS OF MEASURE
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO units_of_measure (name, abbreviation, category) VALUES
  -- Existing (12)
  ('Each',          'EA',   'count'),
  ('Box',           'BX',   'count'),
  ('Carton',        'CTN',  'count'),
  ('Pallet',        'PL',   'count'),
  ('Kilogram',      'KG',   'weight'),
  ('Gram',          'G',    'weight'),
  ('Liter',         'L',    'volume'),
  ('Milliliter',    'ML',   'volume'),
  ('Meter',         'M',    'length'),
  ('Square Meter',  'M2',   'area'),
  ('Dozen',         'DZ',   'count'),
  ('Pair',          'PR',   'count'),
  -- From CSV (3 new)
  ('Set',           'ST',   'count'),
  ('Packet',        'PK',   'count'),
  ('Drum',          'DR',   'volume'),
  -- Additional pharmaceutical / medical (8)
  ('Tube',          'TU',   'count'),
  ('Bottle',        'BT',   'count'),
  ('Vial',          'VL',   'count'),
  ('Ampoule',       'AMP',  'count'),
  ('Sachet',        'SA',   'count'),
  ('Strip',         'SR',   'count'),
  ('Tablet',        'TAB',  'count'),
  ('Capsule',       'CAP',  'count')
ON CONFLICT (abbreviation) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 5. ORGANIZATION (CPI hierarchy from organization.csv)
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO organizations (id, name, code) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Center for Peace and Integrity', 'CPI')
ON CONFLICT (code) DO NOTHING;

-- Org levels with parent references (all levels from CSV)
INSERT INTO org_levels (org_id, parent_id, name, code, level) VALUES
  ('00000000-0000-0000-0000-000000000001', NULL, 'Global HQ',                        'CPI-GHQ',   1),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-GHQ'),
   'Country Office - Bangladesh', 'CPI-BD', 2),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-GHQ'),
   'Country Office - Myanmar',    'CPI-MM', 2),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-BD'),
   'Project Office - Cox Bazar',  'CPI-CXB', 3),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-BD'),
   'Project Office - Dhaka',      'CPI-DHK', 3),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-MM'),
   'Project Office - Yangon',     'CPI-YGN', 3),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-CXB'),
   'Camp 5 Health Post',          'CPI-CXB-C5',  4),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-CXB'),
   'Camp 4W Distribution Point',  'CPI-CXB-C4W', 4)
ON CONFLICT (org_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 6. PROGRAMS
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO programs (org_id, code, name, description) VALUES
  ('00000000-0000-0000-0000-000000000001', 'H&N',  'Health and Nutrition',                'Medical and nutrition programs including primary healthcare'),
  ('00000000-0000-0000-0000-000000000001', 'WASH', 'Water, Sanitation and Hygiene',       'Clean water and sanitation infrastructure and promotion'),
  ('00000000-0000-0000-0000-000000000001', 'LVL',  'Livelihood',                          'Economic empowerment and livelihood support programs'),
  ('00000000-0000-0000-0000-000000000001', 'EDU',  'Education',                           'Formal and non-formal education programs'),
  ('00000000-0000-0000-0000-000000000001', 'EPRR', 'Emergency Preparedness and Response', 'Disaster response readiness and contingency planning'),
  ('00000000-0000-0000-0000-000000000001', 'R&L',  'Research and Learning',               'Monitoring, evaluation and research activities'),
  ('00000000-0000-0000-0000-000000000001', 'PROT', 'Protection',                          'Child protection and gender-based violence prevention'),
  ('00000000-0000-0000-0000-000000000001', 'SHELTER', 'Emergency Shelter',                'Emergency shelter and temporary housing solutions'),
  ('00000000-0000-0000-0000-000000000001', 'NFI',  'Non-Food Items',                      'Essential non-food item distributions'),
  ('00000000-0000-0000-0000-000000000001', 'LOGS', 'Logistics Support',                   'Cross-cutting logistics and supply chain support')
ON CONFLICT (org_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 7. DEPARTMENTS
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO departments (org_id, name, code) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Health & Nutrition',                     'H&N'),
  ('00000000-0000-0000-0000-000000000001', 'WASH',                                   'WASH'),
  ('00000000-0000-0000-0000-000000000001', 'Logistics',                              'LOG'),
  ('00000000-0000-0000-0000-000000000001', 'Procurement',                            'PROC'),
  ('00000000-0000-0000-0000-000000000001', 'Finance',                                'FIN'),
  ('00000000-0000-0000-0000-000000000001', 'HR & Admin',                             'HR'),
  ('00000000-0000-0000-0000-000000000001', 'Programs',                               'PROG'),
  ('00000000-0000-0000-0000-000000000001', 'MEAL (Monitoring, Evaluation, Accountability, Learning)', 'MEAL')
ON CONFLICT (org_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 8. USERS
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO users (org_id, email, name, role_id, org_level_id) VALUES
  ('00000000-0000-0000-0000-000000000001', 'admin@zarishlog.org',         'System Admin',
   (SELECT id FROM roles WHERE code = 'R01'),
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-GHQ')),
  ('00000000-0000-0000-0000-000000000001', 'countryrep.bd@zarishlog.org', 'Fatima Rahman',
   (SELECT id FROM roles WHERE code = 'R02'),
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-BD')),
  ('00000000-0000-0000-0000-000000000001', 'wh.manager@zarishlog.org',    'Mohammad Ali',
   (SELECT id FROM roles WHERE code = 'R04'),
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-CXB')),
  ('00000000-0000-0000-0000-000000000001', 'storekeeper@zarishlog.org',   'Ayesha Begum',
   (SELECT id FROM roles WHERE code = 'R05'),
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-CXB')),
  ('00000000-0000-0000-0000-000000000001', 'dept.manager@zarishlog.org',  'Kamal Hossain',
   (SELECT id FROM roles WHERE code = 'R07'),
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-CXB')),
  ('00000000-0000-0000-0000-000000000001', 'field.worker@zarishlog.org',  'Nurul Islam',
   (SELECT id FROM roles WHERE code = 'R10'),
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-CXB-C5')),
  ('00000000-0000-0000-0000-000000000001', 'qa.officer@zarishlog.org',    'Shahina Akhter',
   (SELECT id FROM roles WHERE code = 'R11'),
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-CXB')),
  ('00000000-0000-0000-0000-000000000001', 'logistics.admin@zarishlog.org', 'Tariq Mahmud',
   (SELECT id FROM roles WHERE code = 'R06'),
   (SELECT id FROM org_levels WHERE org_id = '00000000-0000-0000-0000-000000000001' AND code = 'CPI-BD'))
ON CONFLICT (org_id, email) DO NOTHING;

-- User role assignments (for the extended schema user_role_assignments table)
INSERT INTO user_role_assignments (user_id, role_id, org_level_id)
SELECT u.id, r.id, u.org_level_id
FROM users u
JOIN roles r ON r.id = u.role_id
ON CONFLICT (user_id, role_id, org_level_id) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 9. WAREHOUSES
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO warehouses (org_id, name, code, type, city, country) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Cox Bazar Central Warehouse', 'CXB-CWH', 'central',        'Cox Bazar', 'Bangladesh'),
  ('00000000-0000-0000-0000-000000000001', 'Sub-Warehouse Camp 4',        'CXB-SWH-1', 'sub_warehouse', 'Cox Bazar', 'Bangladesh'),
  ('00000000-0000-0000-0000-000000000001', 'Transit Hub Dhaka',           'CXB-TRN',   'transit',        'Dhaka',     'Bangladesh'),
  ('00000000-0000-0000-0000-000000000001', 'Cold Chain Hub',              'CXB-COLD',  'central',        'Cox Bazar', 'Bangladesh')
ON CONFLICT (org_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 10. LOCATIONS
-- ═══════════════════════════════════════════════════════════════════════════
-- Existing CXB-CWH locations
INSERT INTO locations (warehouse_id, code, name, type) VALUES
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'),  'RECV',  'Receiving Area',      'area'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'),  'GEN-A', 'General Storage A',   'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'),  'GEN-B', 'General Storage B',   'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'),  'COLD',  'Cold Chain Storage',  'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'),  'QA',    'QA/Quarantine Area',  'area'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'),  'DISP',  'Dispatch Area',       'area')
ON CONFLICT (warehouse_id, code) DO NOTHING;

-- CXB-SWH-1 (Sub-Warehouse Camp 4)
INSERT INTO locations (warehouse_id, code, name, type) VALUES
  ((SELECT id FROM warehouses WHERE code = 'CXB-SWH-1'), 'RECV',  'Receiving Area',     'area'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-SWH-1'), 'GEN',   'General Storage',    'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-SWH-1'), 'DISP',  'Dispatch Area',      'area')
ON CONFLICT (warehouse_id, code) DO NOTHING;

-- CXB-TRN (Transit Hub Dhaka)
INSERT INTO locations (warehouse_id, code, name, type) VALUES
  ((SELECT id FROM warehouses WHERE code = 'CXB-TRN'),  'RECV',   'Receiving Area',     'area'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-TRN'),  'STG-A',  'Storage Zone A',     'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-TRN'),  'STG-B',  'Storage Zone B',     'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-TRN'),  'DISP',   'Dispatch Area',      'area')
ON CONFLICT (warehouse_id, code) DO NOTHING;

-- CXB-COLD (Cold Chain Hub)
INSERT INTO locations (warehouse_id, code, name, type, is_cold_chain) VALUES
  ((SELECT id FROM warehouses WHERE code = 'CXB-COLD'), 'COLD-A',   'Deep Freeze -20C',   'zone', true),
  ((SELECT id FROM warehouses WHERE code = 'CXB-COLD'), 'COLD-B',   'Chilled 2-8C',       'zone', true),
  ((SELECT id FROM warehouses WHERE code = 'CXB-COLD'), 'COLD-C',   'Ambient 15-25C',     'zone', true),
  ((SELECT id FROM warehouses WHERE code = 'CXB-COLD'), 'RECV-COL', 'Cold Receiving Area','area', true)
ON CONFLICT (warehouse_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 11. PRODUCT CATEGORIES
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO product_categories (org_id, name, description, unspsc) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Medicines & Drugs',        'Pharmaceutical products including tablets, capsules, injectables', '51000000'),
  ('00000000-0000-0000-0000-000000000001', 'Medical Supplies',         'Consumable medical supplies including dressings, gloves, syringes', '42000000'),
  ('00000000-0000-0000-0000-000000000001', 'Medical Equipment',        'Durable medical equipment including diagnostic, therapeutic devices', '42100000'),
  ('00000000-0000-0000-0000-000000000001', 'Nutrition',                'Nutritional products including therapeutic foods, supplements', '51000000'),
  ('00000000-0000-0000-0000-000000000001', 'WASH Supplies',            'Water, sanitation and hygiene supplies', '47000000'),
  ('00000000-0000-0000-0000-000000000001', 'Office & Admin',           'Office supplies, furniture, IT equipment', '44000000'),
  ('00000000-0000-0000-0000-000000000001', 'Laboratory & Diagnostics', 'Lab reagents, test kits, diagnostic consumables', '41000000'),
  ('00000000-0000-0000-0000-000000000001', 'Cold Chain Supplies',      'Vaccine carriers, cold boxes, temperature monitoring', '42172000'),
  ('00000000-0000-0000-0000-000000000001', 'Shelter & NFI',            'Emergency shelter materials and non-food items', '46000000'),
  ('00000000-0000-0000-0000-000000000001', 'Communications & IT',      'Radio, satellite, IT networking equipment', '43000000'),
  ('00000000-0000-0000-0000-000000000001', 'Vehicles & Transport',     'Vehicles, motorcycles, bicycles, and transport parts', '48000000'),
  ('00000000-0000-0000-0000-000000000001', 'Protective Gear (PPE)',    'Personal protective equipment', '42000000'),
  ('00000000-0000-0000-0000-000000000001', 'Cleaning & Hygiene',       'Cleaning agents, disinfectants, hygiene supplies', '47000000')
ON CONFLICT (org_id, name) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 12. PRODUCTS (from master_product_list.csv — all 40 rows)
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO products (org_id, category_id, uom_id, sku, name, item_type, description, is_batch_tracked, is_expiry_tracked, is_cold_chain, is_hazardous, status)
VALUES
-- Medicines & Drugs (5)
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medicines & Drugs'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'MED-AMOX-500', 'Amoxicillin 500mg Capsules', 'drug', 'Broad-spectrum antibiotic', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medicines & Drugs'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'MED-PCM-500', 'Paracetamol 500mg Tablets', 'drug', 'Analgesic and antipyretic', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medicines & Drugs'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'MED-ART-100', 'Artemether 20mg+Lumefantrine 120mg', 'drug', 'Antimalarial combination therapy', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medicines & Drugs'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'PK'),
 'MED-ORS', 'WHO Oral Rehydration Salts', 'drug', 'For diarrhea management', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medicines & Drugs'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'MED-VIT-A', 'Vitamin A Capsules 200000 IU', 'drug', 'High-dose vitamin A supplement', true, true, false, false, 'active'),
-- Medical Supplies (6)
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
 'SUP-LTX-MED', 'Latex Examination Gloves (Box 100)', 'medical_supply', 'Sterile latex examination gloves', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
 'SUP-SYG-5ML', 'Syringe 5ml (Box 100)', 'medical_supply', 'Disposable syringe 5ml', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
 'SUP-GAUZE', 'Gauze Swabs Sterile 10x10cm (Box 100)', 'medical_supply', 'Sterile gauze swabs', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
 'SUP-MASK-SRG', 'Surgical Mask (Box 50)', 'medical_supply', 'Disposable surgical face masks', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'SUP-BANDAGE', 'Crepe Bandage 10cm x 4.5m', 'medical_supply', 'Elastic crepe bandage', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'SUP-CATH-FOLEY', 'Foley Catheter 16FR', 'medical_supply', 'Silicone foley catheter', false, true, false, false, 'active'),
-- Medical Equipment (5)
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Equipment'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'EQP-BP-MANUAL', 'Manual Blood Pressure Monitor', 'equipment', 'Aneroid sphygmomanometer', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Equipment'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'EQP-STETH', 'Stethoscope Dual Head', 'equipment', 'Dual-head acoustic stethoscope', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Equipment'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'EQP-THERM-IR', 'Infrared Thermometer', 'equipment', 'Non-contact infrared thermometer', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Equipment'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'EQP-OX-PULSE', 'Pulse Oximeter', 'equipment', 'Finger pulse oximeter', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Equipment'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'EQP-REF-COLD', 'Solar-Powered Refrigerator', 'equipment', 'For vaccine cold chain storage', false, false, true, false, 'active'),
-- Nutrition (3)
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Nutrition'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'CTN'),
 'NUT-RUTF', 'RUTF Sachet 92g (Carton)', 'nutrition', 'Ready-to-Use Therapeutic Food', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Nutrition'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
 'NUT-F100', 'F100 Therapeutic Milk (Box 12)', 'nutrition', 'Therapeutic milk for severe malnutrition', true, true, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Nutrition'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'CTN'),
 'NUT-LNS', 'LNS (Lipid-Based Nutrient Supplement)', 'nutrition', 'Lipid-based nutrient supplement', true, true, false, false, 'active'),
-- WASH Supplies (6)
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'WASH Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'WASH-SOAP-BAR', 'Soap Bar 200g', 'material', 'Laundry and handwashing soap', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'WASH Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'WASH-BKT-10L', 'Plastic Bucket 10L', 'material', 'Collapsible water bucket with tap', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'WASH Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'WASH-JERRY-20L', 'Jerry Can 20L', 'material', 'Plastic jerry can for water transport', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'WASH Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
 'WASH-TAB-CHLOR', 'Chlorine Tablets (Bottle 300)', 'material', 'Water purification tablets', true, false, false, true, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'WASH Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
 'WASH-AQUA-TAB', 'AquaTab 33mg (Box 1000)', 'material', 'Water purification tablets 33mg', true, false, false, true, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'WASH Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'WASH-MOSQ-NET', 'Insecticide Mosquito Net', 'material', 'Long-lasting insecticidal net', false, false, false, false, 'active'),
-- Office & Admin — misc under "Office & Admin" category (12)
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'WASH-TARP', 'Tarpaulin 4x6m', 'material', 'Weather-resistant shelter tarp', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'ST'),
 'GEN-BLAN-KIT', 'Blanket Kit (Emergency)', 'material', 'Thermal emergency blankets kit', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'ST'),
 'GEN-KITCHEN-SET', 'Kitchen Set Family', 'material', 'Family kitchen set 4 persons', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'ST'),
 'GEN-DIGNITY-KIT', 'Dignity Kit Women', 'material', 'Female hygiene and dignity kit', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'GEN-SOLAR-LAMP', 'Solar Lamp with USB', 'equipment', 'Portable solar powered LED lamp', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
 'ADMN-PAPER-A4', 'Office Paper A4 Box', 'consumable', 'Printer paper A4 5000 sheets', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'ADMN-TONER', 'Toner Cartridge HP 85A', 'consumable', 'Laser printer toner cartridge', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'ADMN-DESK', 'Office Desk 120x60cm', 'asset', 'Standard office desk', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'ADMN-CHAIR', 'Office Chair Ergonomic', 'asset', 'Ergonomic office chair', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'ADMN-LAPTOP', 'Laptop Computer 15.6 inch', 'asset', 'Business laptop with accessories', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'ADMN-TABLET', 'Tablet 10 inch', 'asset', 'Android tablet for field data collection', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'ADMN-VHF', 'VHF Radio Handheld', 'asset', 'Handheld VHF radio for field communication', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Office & Admin'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'ADMN-POWER-BANK', 'Power Bank 20000mAh', 'equipment', 'Solar-compatible power bank', false, false, false, false, 'active'),
-- Instruments (2)
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Equipment'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'INSTR-SCALE-DIG', 'Digital Weighing Scale 150kg', 'instrument', 'Digital scale for nutrition screening', false, false, false, false, 'active'),
('00000000-0000-0000-0000-000000000001',
 (SELECT id FROM product_categories WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'Medical Supplies'),
 (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
 'INSTR-MUAC', 'MUAC Measuring Tape', 'instrument', 'Mid-upper arm circumference tape', false, false, false, false, 'active')
ON CONFLICT (org_id, sku) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 13. PRODUCT PACKAGING
-- ═══════════════════════════════════════════════════════════════════════════
-- MED-AMOX-500: Each → Box of 100 → Carton of 10 boxes
INSERT INTO product_packaging (product_id, uom_id, package_level, quantity, package_description)
SELECT p.id, uom.id, 1, 100, 'Box of 100 capsules'
FROM products p, units_of_measure uom
WHERE p.sku = 'MED-AMOX-500' AND uom.abbreviation = 'BX'
  AND NOT EXISTS (SELECT 1 FROM product_packaging WHERE product_id = p.id AND package_level = 1);

INSERT INTO product_packaging (product_id, uom_id, package_level, quantity, package_description)
SELECT p.id, uom.id, 2, 10, 'Carton of 10 boxes (1000 capsules)'
FROM products p, units_of_measure uom
WHERE p.sku = 'MED-AMOX-500' AND uom.abbreviation = 'CTN'
  AND NOT EXISTS (SELECT 1 FROM product_packaging WHERE product_id = p.id AND package_level = 2);

-- SUP-LTX-MED: Each → Box of 100 → Carton of 10 boxes
INSERT INTO product_packaging (product_id, uom_id, package_level, quantity, package_description)
SELECT p.id, uom.id, 1, 100, 'Box of 100 gloves'
FROM products p, units_of_measure uom
WHERE p.sku = 'SUP-LTX-MED' AND uom.abbreviation = 'BX'
  AND NOT EXISTS (SELECT 1 FROM product_packaging WHERE product_id = p.id AND package_level = 1);

INSERT INTO product_packaging (product_id, uom_id, package_level, quantity, package_description)
SELECT p.id, uom.id, 2, 10, 'Carton of 10 boxes (1000 gloves)'
FROM products p, units_of_measure uom
WHERE p.sku = 'SUP-LTX-MED' AND uom.abbreviation = 'CTN'
  AND NOT EXISTS (SELECT 1 FROM product_packaging WHERE product_id = p.id AND package_level = 2);

-- WASH-SOAP-BAR: Each → Carton of 50
INSERT INTO product_packaging (product_id, uom_id, package_level, quantity, package_description)
SELECT p.id, uom.id, 1, 50, 'Carton of 50 soap bars'
FROM products p, units_of_measure uom
WHERE p.sku = 'WASH-SOAP-BAR' AND uom.abbreviation = 'CTN'
  AND NOT EXISTS (SELECT 1 FROM product_packaging WHERE product_id = p.id AND package_level = 1);

-- NUT-RUTF: Sachet → Carton of 150
INSERT INTO product_packaging (product_id, uom_id, package_level, quantity, package_description)
SELECT p.id, uom.id, 1, 150, 'Carton of 150 sachets'
FROM products p, units_of_measure uom
WHERE p.sku = 'NUT-RUTF' AND uom.abbreviation = 'CTN'
  AND NOT EXISTS (SELECT 1 FROM product_packaging WHERE product_id = p.id AND package_level = 1);

INSERT INTO product_packaging (product_id, uom_id, package_level, quantity, package_description)
SELECT p.id, uom.id, 1, 1, 'Individual sachet 92g'
FROM products p, units_of_measure uom
WHERE p.sku = 'NUT-RUTF' AND uom.abbreviation = 'SA'
  AND NOT EXISTS (SELECT 1 FROM product_packaging WHERE product_id = p.id AND package_level = 1 AND uom_id = uom.id);

-- ═══════════════════════════════════════════════════════════════════════════
-- 14. ADJUSTMENT REASON CODES
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO adjustment_reason_codes (code, category, description, requires_approval) VALUES
  ('DAMAGE',          'damage',  'Damaged goods',                                 true),
  ('EXPIRY',          'expiry',  'Expired goods',                                 true),
  ('THEFT',           'theft',   'Theft or unexplained loss',                     true),
  ('COUNT_ERROR',     'count',   'Count variance found during physical count',    false),
  ('BREAKAGE',        'damage',  'Breakage during handling or transport',         false),
  ('DONATION',        'donation','Donated to program or external organization',   false),
  ('SAMPLE',          'quality', 'Quality sample taken for testing',              false),
  ('RETURN_TO_SUPPLIER', 'return', 'Returned to supplier or manufacturer',        true)
ON CONFLICT (code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 15. DISPOSAL METHODS
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO disposal_methods (code, name, description, requires_witness, environmental_impact) VALUES
  ('INCINERATION', 'Controlled Incineration',  'High-temperature controlled incineration',  true,  'Air emissions, ash residue'),
  ('BURIAL',       'Sanitary Burial',          'Sanitary burial in designated landfill',     false, 'Soil impact'),
  ('CHEMICAL',     'Chemical Treatment',       'Chemical neutralization or treatment',       false, 'Chemical waste'),
  ('RECYCLING',    'Material Recycling',       'Material recovery and recycling',             false, 'Positive — reduces waste'),
  ('LANDFILL',     'Sanitary Landfill',        'Disposal in authorized sanitary landfill',   false, 'Long-term land use'),
  ('RETURN',       'Return to Supplier',       'Return to manufacturer or supplier',         false, 'No local impact')
ON CONFLICT (code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 16. QA CHECKLIST TEMPLATES & ITEMS
-- ═══════════════════════════════════════════════════════════════════════════

-- Template 1: GOODS_RECEIPT_INSPECTION
INSERT INTO qa_checklist_templates (org_id, code, name, description, category, is_mandatory) VALUES
  ('00000000-0000-0000-0000-000000000001', 'GOODS_RECEIPT_INSPECTION', 'Goods Receipt QA Inspection',
   'Standard goods receipt quality assurance checklist', 'receiving', true)
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 1, 'Is packaging intact and undamaged?',            'yes', true,  2.0
FROM qa_checklist_templates WHERE code = 'GOODS_RECEIPT_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 1);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 2, 'Does product match PO description?',            'yes', true,  2.0
FROM qa_checklist_templates WHERE code = 'GOODS_RECEIPT_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 2);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 3, 'Is batch/lot number visible?',                  'yes', false, 1.0
FROM qa_checklist_templates WHERE code = 'GOODS_RECEIPT_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 3);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 4, 'Is expiry date > 6 months from today?',         'yes', true,  2.0
FROM qa_checklist_templates WHERE code = 'GOODS_RECEIPT_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 4);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 5, 'Is cold chain maintained? (if applicable)',      'yes', false, 1.5
FROM qa_checklist_templates WHERE code = 'GOODS_RECEIPT_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 5);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 6, 'Are quantity and UoM correct?',                 'yes', true,  2.0
FROM qa_checklist_templates WHERE code = 'GOODS_RECEIPT_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 6);

-- Template 2: STORAGE_INSPECTION
INSERT INTO qa_checklist_templates (org_id, code, name, description, category, is_mandatory) VALUES
  ('00000000-0000-0000-0000-000000000001', 'STORAGE_INSPECTION', 'Storage Area Inspection',
   'Standard storage area inspection checklist', 'storage', false)
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 1, 'Are temperature logs maintained?',               'yes', false, 1.0
FROM qa_checklist_templates WHERE code = 'STORAGE_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 1);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 2, 'Are FEFO rules followed?',                      'yes', true,  2.0
FROM qa_checklist_templates WHERE code = 'STORAGE_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 2);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 3, 'Are storage areas clean and organized?',         'yes', false, 1.0
FROM qa_checklist_templates WHERE code = 'STORAGE_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 3);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 4, 'Are hazardous materials properly segregated?',   'yes', true,  2.0
FROM qa_checklist_templates WHERE code = 'STORAGE_INSPECTION' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 4);

-- Template 3: DISTRIBUTION_READINESS
INSERT INTO qa_checklist_templates (org_id, code, name, description, category, is_mandatory) VALUES
  ('00000000-0000-0000-0000-000000000001', 'DISTRIBUTION_READINESS', 'Pre-Distribution Readiness',
   'Pre-distribution checklist for field distribution readiness', 'distribution', true)
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 1, 'Items sorted by batch/expiry?',                  'yes', true,  2.0
FROM qa_checklist_templates WHERE code = 'DISTRIBUTION_READINESS' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 1);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 2, 'Distribution area secured?',                     'yes', false, 1.0
FROM qa_checklist_templates WHERE code = 'DISTRIBUTION_READINESS' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 2);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 3, 'Beneficiary list verified?',                     'yes', false, 1.5
FROM qa_checklist_templates WHERE code = 'DISTRIBUTION_READINESS' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 3);

INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
SELECT id, 4, 'Items properly packed for transport?',            'yes', false, 1.0
FROM qa_checklist_templates WHERE code = 'DISTRIBUTION_READINESS' AND org_id = '00000000-0000-0000-0000-000000000001'
AND NOT EXISTS (SELECT 1 FROM qa_checklist_items WHERE template_id = qa_checklist_templates.id AND item_order = 4);

-- ═══════════════════════════════════════════════════════════════════════════
-- 17. ALERT CONFIGURATIONS
-- ═══════════════════════════════════════════════════════════════════════════
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM alert_configurations WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'EXPIRY_SOON') THEN
    INSERT INTO alert_configurations (org_id, alert_type, name, description, threshold_type, threshold_value, enabled, notification_channels)
    VALUES ('00000000-0000-0000-0000-000000000001', 'expiry',  'EXPIRY_SOON',     'Warn when product expiry is within 90 days',  'days', '90',  true, 'email,in_app');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM alert_configurations WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'EXPIRY_IMMINENT') THEN
    INSERT INTO alert_configurations (org_id, alert_type, name, description, threshold_type, threshold_value, enabled, notification_channels)
    VALUES ('00000000-0000-0000-0000-000000000001', 'expiry',  'EXPIRY_IMMINENT',  'Warn when product expiry is within 30 days',  'days', '30',  true, 'email,in_app,sms');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM alert_configurations WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'LOW_STOCK_WARNING') THEN
    INSERT INTO alert_configurations (org_id, alert_type, name, description, threshold_type, threshold_value, enabled, notification_channels)
    VALUES ('00000000-0000-0000-0000-000000000001', 'low_stock','LOW_STOCK_WARNING', 'Alert when stock falls below reorder point',  'percentage', '0', true, 'email,in_app');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM alert_configurations WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'CRITICAL_STOCK') THEN
    INSERT INTO alert_configurations (org_id, alert_type, name, description, threshold_type, threshold_value, enabled, notification_channels)
    VALUES ('00000000-0000-0000-0000-000000000001', 'low_stock','CRITICAL_STOCK',   'Alert when stock falls below safety stock',  'percentage', '0', true, 'email,in_app,sms');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM alert_configurations WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'OVERSTOCK') THEN
    INSERT INTO alert_configurations (org_id, alert_type, name, description, threshold_type, threshold_value, enabled, notification_channels)
    VALUES ('00000000-0000-0000-0000-000000000001', 'overstock','OVERSTOCK',        'Alert when stock exceeds maximum level',    'percentage', '0', true, 'email,in_app');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM alert_configurations WHERE org_id = '00000000-0000-0000-0000-000000000001' AND name = 'SLEEPING_STOCK') THEN
    INSERT INTO alert_configurations (org_id, alert_type, name, description, threshold_type, threshold_value, enabled, notification_channels)
    VALUES ('00000000-0000-0000-0000-000000000001', 'sleeping_stock','SLEEPING_STOCK','Alert when product has had no movement for 180 days','days', '180', true, 'email,in_app');
  END IF;
END $$;

-- ═══════════════════════════════════════════════════════════════════════════
-- 18. SAMPLE STOCK DATA — Batches & Stock Levels
-- ═══════════════════════════════════════════════════════════════════════════

-- Create batches for batch-tracked products
INSERT INTO batches (org_id, product_id, batch_number, expiry_date, manufacturer)
SELECT '00000000-0000-0000-0000-000000000001', p.id, 'AMOX-2026-A', CURRENT_DATE + INTERVAL '18 months', 'PharmaCorp Ltd'
FROM products p WHERE p.sku = 'MED-AMOX-500'
ON CONFLICT (org_id, product_id, batch_number) DO NOTHING;

INSERT INTO batches (org_id, product_id, batch_number, expiry_date, manufacturer)
SELECT '00000000-0000-0000-0000-000000000001', p.id, 'AMOX-2026-B', CURRENT_DATE + INTERVAL '24 months', 'PharmaCorp Ltd'
FROM products p WHERE p.sku = 'MED-AMOX-500'
ON CONFLICT (org_id, product_id, batch_number) DO NOTHING;

INSERT INTO batches (org_id, product_id, batch_number, expiry_date, manufacturer)
SELECT '00000000-0000-0000-0000-000000000001', p.id, 'LTX-2026-A', CURRENT_DATE + INTERVAL '24 months', 'MedSupply Intl'
FROM products p WHERE p.sku = 'SUP-LTX-MED'
ON CONFLICT (org_id, product_id, batch_number) DO NOTHING;

INSERT INTO batches (org_id, product_id, batch_number, expiry_date, manufacturer)
SELECT '00000000-0000-0000-0000-000000000001', p.id, 'GAU-2026-A', CURRENT_DATE + INTERVAL '36 months', 'MedSupply Intl'
FROM products p WHERE p.sku = 'SUP-GAUZE'
ON CONFLICT (org_id, product_id, batch_number) DO NOTHING;

INSERT INTO batches (org_id, product_id, batch_number, expiry_date, manufacturer)
SELECT '00000000-0000-0000-0000-000000000001', p.id, 'RUTF-2026-A', CURRENT_DATE + INTERVAL '18 months', 'NutriAid Global'
FROM products p WHERE p.sku = 'NUT-RUTF'
ON CONFLICT (org_id, product_id, batch_number) DO NOTHING;

-- Additional batches for products included in stock levels
INSERT INTO batches (org_id, product_id, batch_number, expiry_date, manufacturer)
SELECT '00000000-0000-0000-0000-000000000001', p.id, 'PCM-2026-A', CURRENT_DATE + INTERVAL '24 months', 'PharmaCorp Ltd'
FROM products p WHERE p.sku = 'MED-PCM-500'
ON CONFLICT (org_id, product_id, batch_number) DO NOTHING;

INSERT INTO batches (org_id, product_id, batch_number, expiry_date, manufacturer)
SELECT '00000000-0000-0000-0000-000000000001', p.id, 'SYG-2026-A', CURRENT_DATE + INTERVAL '36 months', 'MedSupply Intl'
FROM products p WHERE p.sku = 'SUP-SYG-5ML'
ON CONFLICT (org_id, product_id, batch_number) DO NOTHING;

INSERT INTO batches (org_id, product_id, batch_number, expiry_date, manufacturer)
SELECT '00000000-0000-0000-0000-000000000001', p.id, 'CHLOR-2026-A', NULL, 'ChemCorp Ltd'
FROM products p WHERE p.sku = 'WASH-TAB-CHLOR'
ON CONFLICT (org_id, product_id, batch_number) DO NOTHING;

-- Create stock levels for CXB-CWH warehouse
-- MED-AMOX-500: 5000 units in GEN-A (batch AMOX-2026-A)
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 5000, 'on_hand'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'MED-AMOX-500' AND w.code = 'CXB-CWH' AND l.code = 'GEN-A' AND b.batch_number = 'AMOX-2026-A'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- MED-AMOX-500: 2000 units in GEN-A (batch AMOX-2026-B)
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 2000, 'on_hand'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'MED-AMOX-500' AND w.code = 'CXB-CWH' AND l.code = 'GEN-A' AND b.batch_number = 'AMOX-2026-B'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- MED-PCM-500: 10000 units in GEN-A
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 10000, 'on_hand'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'MED-PCM-500' AND w.code = 'CXB-CWH' AND l.code = 'GEN-A' AND b.batch_number = 'PCM-2026-A'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- SUP-LTX-MED: 200 boxes in GEN-B
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 200, 'on_hand'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'SUP-LTX-MED' AND w.code = 'CXB-CWH' AND l.code = 'GEN-B' AND b.batch_number = 'LTX-2026-A'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- SUP-SYG-5ML: 150 boxes in GEN-B
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 150, 'on_hand'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'SUP-SYG-5ML' AND w.code = 'CXB-CWH' AND l.code = 'GEN-B' AND b.batch_number = 'SYG-2026-A'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- SUP-GAUZE: 300 boxes in GEN-B
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 300, 'on_hand'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'SUP-GAUZE' AND w.code = 'CXB-CWH' AND l.code = 'GEN-B' AND b.batch_number = 'GAU-2026-A'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- NUT-RUTF: 50 cartons in GEN-A
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 50, 'on_hand'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'NUT-RUTF' AND w.code = 'CXB-CWH' AND l.code = 'GEN-A' AND b.batch_number = 'RUTF-2026-A'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- WASH-SOAP-BAR: 2000 units in GEN-B (no batch)
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, NULL, 2000, 'on_hand'
FROM products p, warehouses w, locations l
WHERE p.sku = 'WASH-SOAP-BAR' AND w.code = 'CXB-CWH' AND l.code = 'GEN-B'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- EQP-BP-MANUAL: 25 units in GEN-A (no batch)
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, NULL, 25, 'on_hand'
FROM products p, warehouses w, locations l
WHERE p.sku = 'EQP-BP-MANUAL' AND w.code = 'CXB-CWH' AND l.code = 'GEN-A'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- WASH-BKT-10L: 500 units in GEN-B (no batch)
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, NULL, 500, 'on_hand'
FROM products p, warehouses w, locations l
WHERE p.sku = 'WASH-BKT-10L' AND w.code = 'CXB-CWH' AND l.code = 'GEN-B'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- WASH-TAB-CHLOR: 100 bottles in QA area (quarantined)
INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 100, 'quarantined'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'WASH-TAB-CHLOR' AND w.code = 'CXB-CWH' AND l.code = 'QA' AND b.batch_number = 'CHLOR-2026-A'
ON CONFLICT (product_id, warehouse_id, location_id, batch_id) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 19. SAMPLE STOCK MOVEMENTS
-- ═══════════════════════════════════════════════════════════════════════════

-- Goods receipt: GRN-2026-001 (Jan 15)
INSERT INTO goods_receipts (org_id, warehouse_id, grn_number, supplier, po_number, received_by, status, notes)
SELECT '00000000-0000-0000-0000-000000000001', w.id, 'GRN-2026-001', 'PharmaCorp Ltd', 'PO-2026-001', 'wh.manager@zarishlog.org', 'active', 'Initial stock receipt'
FROM warehouses w WHERE w.code = 'CXB-CWH'
ON CONFLICT (org_id, grn_number) DO NOTHING;

-- GRN line items
INSERT INTO grn_line_items (grn_id, product_id, batch_number, expiry_date, quantity, unit_cost)
SELECT grn.id, p.id, 'AMOX-2026-A', (CURRENT_DATE + INTERVAL '18 months')::date, 5000, 0.15
FROM goods_receipts grn, products p
WHERE grn.grn_number = 'GRN-2026-001' AND p.sku = 'MED-AMOX-500'
  AND NOT EXISTS (SELECT 1 FROM grn_line_items WHERE grn_id = grn.id AND product_id = p.id AND batch_number = 'AMOX-2026-A');

INSERT INTO grn_line_items (grn_id, product_id, batch_number, expiry_date, quantity, unit_cost)
SELECT grn.id, p.id, 'LTX-2026-A', (CURRENT_DATE + INTERVAL '24 months')::date, 200, 3.50
FROM goods_receipts grn, products p
WHERE grn.grn_number = 'GRN-2026-001' AND p.sku = 'SUP-LTX-MED'
  AND NOT EXISTS (SELECT 1 FROM grn_line_items WHERE grn_id = grn.id AND product_id = p.id AND batch_number = 'LTX-2026-A');

-- Stock movements for GRN-2026-001
INSERT INTO stock_movements (org_id, product_id, warehouse_id, location_id, batch_id, movement_type, quantity, ref_doc_type, ref_doc_id, reason_code, created_by, created_at)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 'receipt', 5000, 'grn', 'GRN-2026-001', 'initial_stock', 'wh.manager@zarishlog.org', CURRENT_DATE - INTERVAL '6 months'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'MED-AMOX-500' AND w.code = 'CXB-CWH' AND l.code = 'RECV' AND b.batch_number = 'AMOX-2026-A';

INSERT INTO stock_movements (org_id, product_id, warehouse_id, location_id, batch_id, movement_type, quantity, ref_doc_type, ref_doc_id, reason_code, created_by, created_at)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 'receipt', 200, 'grn', 'GRN-2026-001', 'initial_stock', 'wh.manager@zarishlog.org', CURRENT_DATE - INTERVAL '6 months'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'SUP-LTX-MED' AND w.code = 'CXB-CWH' AND l.code = 'RECV' AND b.batch_number = 'LTX-2026-A';

-- Stock issue: ISS-2026-001 (Feb 1) — Issued 500 Amoxicillin to Health Program
INSERT INTO stock_issues (org_id, warehouse_id, issue_number, requested_by, approved_by, program_id, department_id, status, notes)
SELECT '00000000-0000-0000-0000-000000000001', w.id, 'ISS-2026-001', 'Kamal Hossain', 'Mohammad Ali',
       prog.id, dept.id, 'active', 'Routine issue to Health & Nutrition program'
FROM warehouses w, programs prog, departments dept
WHERE w.code = 'CXB-CWH' AND prog.code = 'H&N' AND dept.code = 'H&N'
ON CONFLICT (org_id, issue_number) DO NOTHING;

INSERT INTO issue_line_items (issue_id, product_id, batch_id, quantity, unit_cost)
SELECT iss.id, p.id, b.id, 500, 0.15
FROM stock_issues iss, products p, batches b
WHERE iss.issue_number = 'ISS-2026-001' AND p.sku = 'MED-AMOX-500' AND b.batch_number = 'AMOX-2026-A'
  AND NOT EXISTS (SELECT 1 FROM issue_line_items WHERE issue_id = iss.id AND product_id = p.id);

INSERT INTO stock_movements (org_id, product_id, warehouse_id, location_id, batch_id, movement_type, quantity, ref_doc_type, ref_doc_id, reason_code, created_by, created_at)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, b.id, 'issue', -500, 'srf', 'ISS-2026-001', 'program_issue', 'wh.manager@zarishlog.org', CURRENT_DATE - INTERVAL '5 months 15 days'
FROM products p, warehouses w, locations l, batches b
WHERE p.sku = 'MED-AMOX-500' AND w.code = 'CXB-CWH' AND l.code = 'GEN-A' AND b.batch_number = 'AMOX-2026-A';

-- Stock issue: ISS-2026-002 (Feb 1) — Issued 200 Soap Bars to WASH Program
INSERT INTO stock_issues (org_id, warehouse_id, issue_number, requested_by, approved_by, program_id, department_id, status, notes)
SELECT '00000000-0000-0000-0000-000000000001', w.id, 'ISS-2026-002', 'Kamal Hossain', 'Mohammad Ali',
       prog.id, dept.id, 'active', 'WASH program distribution supplies'
FROM warehouses w, programs prog, departments dept
WHERE w.code = 'CXB-CWH' AND prog.code = 'WASH' AND dept.code = 'WASH'
ON CONFLICT (org_id, issue_number) DO NOTHING;

INSERT INTO issue_line_items (issue_id, product_id, batch_id, quantity, unit_cost)
SELECT iss.id, p.id, NULL, 200, 0.50
FROM stock_issues iss, products p
WHERE iss.issue_number = 'ISS-2026-002' AND p.sku = 'WASH-SOAP-BAR'
  AND NOT EXISTS (SELECT 1 FROM issue_line_items WHERE issue_id = iss.id AND product_id = p.id);

INSERT INTO stock_movements (org_id, product_id, warehouse_id, location_id, batch_id, movement_type, quantity, ref_doc_type, ref_doc_id, reason_code, created_by, created_at)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, l.id, NULL, 'issue', -200, 'srf', 'ISS-2026-002', 'program_issue', 'wh.manager@zarishlog.org', CURRENT_DATE - INTERVAL '5 months 15 days'
FROM products p, warehouses w, locations l
WHERE p.sku = 'WASH-SOAP-BAR' AND w.code = 'CXB-CWH' AND l.code = 'GEN-B';

-- ═══════════════════════════════════════════════════════════════════════════
-- 20. SAMPLE REORDER RECOMMENDATIONS
-- ═══════════════════════════════════════════════════════════════════════════

-- Update products with realistic stock thresholds
UPDATE products SET
  min_stock = 100,
  max_stock = 1000,
  reorder_point = 200,
  lead_time_days = 30,
  unit_cost = 4.50
WHERE sku = 'NUT-RUTF' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET
  min_stock = 50,
  max_stock = 500,
  reorder_point = 100,
  lead_time_days = 45,
  unit_cost = 3.50
WHERE sku = 'SUP-LTX-MED' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET
  min_stock = 200,
  max_stock = 5000,
  reorder_point = 500,
  lead_time_days = 60,
  unit_cost = 0.15
WHERE sku = 'MED-AMOX-500' AND org_id = '00000000-0000-0000-0000-000000000001';

-- Reorder recommendation: NUT-RUTF (below reorder point, high priority)
INSERT INTO reorder_recommendations (org_id, product_id, warehouse_id, recommendation_date, current_stock, reorder_point, reorder_quantity, lead_time_days, amc_used, safety_stock, recommendation_type, priority, notes)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, CURRENT_DATE, sl.quantity, p.reorder_point, 300, p.lead_time_days, 25, p.safety_stock, 'reorder', 'high', 'RUTF stock critically low — reorder urgently for nutrition program'
FROM products p, warehouses w, stock_levels sl
WHERE p.sku = 'NUT-RUTF' AND w.code = 'CXB-CWH' AND sl.product_id = p.id AND sl.warehouse_id = w.id
  AND NOT EXISTS (SELECT 1 FROM reorder_recommendations WHERE product_id = p.id AND warehouse_id = w.id AND reviewed = false)
LIMIT 1;

-- Reorder recommendation: SUP-LTX-MED (at reorder point, medium priority)
INSERT INTO reorder_recommendations (org_id, product_id, warehouse_id, recommendation_date, current_stock, reorder_point, reorder_quantity, lead_time_days, amc_used, safety_stock, recommendation_type, priority, notes)
SELECT '00000000-0000-0000-0000-000000000001', p.id, w.id, CURRENT_DATE, sl.quantity, p.reorder_point, 200, p.lead_time_days, 15, p.safety_stock, 'reorder', 'medium', 'Gloves stock at reorder point — schedule routine replenishment'
FROM products p, warehouses w, stock_levels sl
WHERE p.sku = 'SUP-LTX-MED' AND w.code = 'CXB-CWH' AND sl.product_id = p.id AND sl.warehouse_id = w.id
  AND NOT EXISTS (SELECT 1 FROM reorder_recommendations WHERE product_id = p.id AND warehouse_id = w.id AND reviewed = false)
LIMIT 1;

-- ─── Product Enhancements: Dosage Forms, Generic Names, Drug DB Ref URLs ──

-- Drugs
UPDATE products SET dosage_form_code = 'CAP', generic_name = 'Amoxicillin trihydrate 500mg', brand_name = 'Amoxil, Moxilin', storage_conditions = 'Store below 30°C', reference_urls = '[{"label":"medex.bd","url":"https://medex.bd/drugs/amoxicillin"},{"label":"WHO EML","url":"https://list.essentialmeds.org/medicines/6"}]'
WHERE sku = 'MED-AMOX-500' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET dosage_form_code = 'TAB', generic_name = 'Paracetamol 500mg', brand_name = 'Panadol, Tylenol', storage_conditions = 'Store below 30°C', reference_urls = '[{"label":"medex.bd","url":"https://medex.bd/drugs/paracetamol"},{"label":"WHO EML","url":"https://list.essentialmeds.org/medicines/7"}]'
WHERE sku = 'MED-PCM-500' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET dosage_form_code = 'TAB', generic_name = 'Artemether 20mg + Lumefantrine 120mg', brand_name = 'Coartem, Riamet', storage_conditions = 'Store below 30°C', reference_urls = '[{"label":"medex.bd","url":"https://medex.bd/drugs/artemether-lumefantrine"}]'
WHERE sku = 'MED-ART-100' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET dosage_form_code = 'PWD', generic_name = 'Oral Rehydration Salts (WHO formulation)', brand_name = 'ORS, Pedialyte', storage_conditions = 'Store in cool dry place', reference_urls = '[{"label":"WHO ORS","url":"https://www.who.int/news-room/fact-sheets/detail/diarrhoeal-disease"}]'
WHERE sku = 'MED-ORS' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET dosage_form_code = 'CAP', generic_name = 'Vitamin A (Retinol) 200000 IU', brand_name = 'A-Vital', storage_conditions = 'Store below 25°C, protect from light', reference_urls = '[{"label":"WHO Vit A","url":"https://www.who.int/publications/i/item/9789241501767"}]'
WHERE sku = 'MED-VIT-A' AND org_id = '00000000-0000-0000-0000-000000000001';

-- Medical Supplies (dosage_form not applicable for supplies, leave null)
UPDATE products SET generic_name = 'Latex examination gloves, sterile', brand_name = 'Sempermed, Ansell', storage_conditions = 'Store in cool dry place'
WHERE sku = 'SUP-LTX-MED' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET generic_name = 'Disposable syringe 5ml Luer Lock', storage_conditions = 'Store in sterile packaging, below 40°C'
WHERE sku = 'SUP-SYG-5ML' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET generic_name = 'Sterile gauze swabs 10x10cm 8-ply', storage_conditions = 'Store in sterile packaging'
WHERE sku = 'SUP-GAUZE' AND org_id = '00000000-0000-0000-0000-000000000001';

-- Nutrition
UPDATE products SET dosage_form_code = 'PWD', generic_name = 'Ready-to-Use Therapeutic Food (peanut-based)', brand_name = 'Plumpy''Nut, Plumpy''Sup', storage_conditions = 'Store below 30°C, no refrigeration needed', reference_urls = '[{"label":"WHO RUTF","url":"https://www.who.int/publications/i/item/9789241512435"}]'
WHERE sku = 'NUT-RUTF' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET dosage_form_code = 'PWD', generic_name = 'F100 Therapeutic Milk formulation', storage_conditions = 'Mix with water, use within 24hrs of reconstitution'
WHERE sku = 'NUT-F100' AND org_id = '00000000-0000-0000-0000-000000000001';

-- WASH supplies
UPDATE products SET generic_name = 'Chlorine-based water purification tablets (Sodium Dichloroisocyanurate)', storage_conditions = 'Store in airtight container, below 30°C'
WHERE sku = 'WASH-TAB-CHLOR' AND org_id = '00000000-0000-0000-0000-000000000001';

-- Equipment
UPDATE products SET generic_name = 'Aneroid sphygmomanometer adult cuff', storage_conditions = 'Store in protective case'
WHERE sku = 'EQP-BP-MANUAL' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET generic_name = 'Acoustic stethoscope dual-head', storage_conditions = 'Store clean, avoid extreme temperatures'
WHERE sku = 'EQP-STETH' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE products SET generic_name = 'Non-contact infrared thermometer', storage_conditions = 'Store at 10-40°C'
WHERE sku = 'EQP-THERM-IR' AND org_id = '00000000-0000-0000-0000-000000000001';

-- ─── Warehouse Geographic Coordinates (for map view / route planning) ─────

UPDATE warehouses SET
  latitude = 21.4272,
  longitude = 91.9781,
  google_maps_url = 'https://maps.google.com/?q=21.4272,91.9781',
  contact_phone = '+8801712345601',
  operating_hours = 'Sun-Thu 8:00-17:00, Sat 9:00-13:00',
  has_generator = TRUE,
  has_cctv = TRUE,
  has_fire_system = TRUE,
  security_guard = TRUE
WHERE code = 'CXB-CWH' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE warehouses SET
  latitude = 21.4100,
  longitude = 91.9600,
  google_maps_url = 'https://maps.google.com/?q=21.4100,91.9600',
  contact_phone = '+8801712345602',
  operating_hours = 'Sun-Thu 8:00-17:00',
  has_generator = TRUE,
  has_cctv = FALSE,
  has_fire_system = TRUE,
  security_guard = FALSE
WHERE code = 'CXB-SWH-1' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE warehouses SET
  latitude = 23.8103,
  longitude = 90.4125,
  google_maps_url = 'https://maps.google.com/?q=23.8103,90.4125',
  contact_phone = '+8801712345603',
  operating_hours = '24 hours (shift-based)',
  has_generator = TRUE,
  has_cctv = TRUE,
  has_fire_system = TRUE,
  security_guard = TRUE
WHERE code = 'CXB-TRN' AND org_id = '00000000-0000-0000-0000-000000000001';

UPDATE warehouses SET
  latitude = 21.4250,
  longitude = 91.9760,
  google_maps_url = 'https://maps.google.com/?q=21.4250,91.9760',
  contact_phone = '+8801712345604',
  operating_hours = 'Sun-Thu 6:00-20:00',
  has_generator = TRUE,
  has_cctv = TRUE,
  has_fire_system = TRUE,
  security_guard = TRUE
WHERE code = 'CXB-COLD' AND org_id = '00000000-0000-0000-0000-000000000001';

-- ─── Organization Branding ───────────────────────────────────────────────

UPDATE organizations SET
  logo_url = 'https://example.org/logo.png',
  website = 'https://example.org',
  country = 'Bangladesh',
  default_currency = 'USD',
  timezone = 'Asia/Dhaka',
  date_format = 'YYYY-MM-DD'
WHERE id = '00000000-0000-0000-0000-000000000001';

-- ═══════════════════════════════════════════════════════════════════════════
-- 21. ADDITIONAL ORGANIZATIONS (Master Catalogue ORG_001–ORG_007)
-- ═══════════════════════════════════════════════════════════════════════════

INSERT INTO organizations (id, name, code, country, default_currency) VALUES
  ('00000000-0000-0000-0000-000000000010', 'United Nations',                     'UN',    'Global', 'USD'),
  ('00000000-0000-0000-0000-000000000020', 'World Health Organization',          'WHO',   'Global', 'USD'),
  ('00000000-0000-0000-0000-000000000030', 'Médecins Sans Frontières',           'MSF',   'Global', 'USD'),
  ('00000000-0000-0000-0000-000000000040', 'IFRC',                               'IFRC',  'Global', 'USD'),
  ('00000000-0000-0000-0000-000000000050', 'World Food Programme',               'WFP',   'Global', 'USD'),
  ('00000000-0000-0000-0000-000000000060', 'UNICEF',                             'UNICEF','Global', 'USD'),
  ('00000000-0000-0000-0000-000000000070', 'International Committee of Red Cross','ICRC',  'Global', 'USD')
ON CONFLICT (code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 22. STANDARD PROGRAMS (Master Catalogue — PRG-* codes)
-- ═══════════════════════════════════════════════════════════════════════════

INSERT INTO programs (org_id, code, name, description) VALUES
  ('00000000-0000-0000-0000-000000000001', 'PRG-HLT', 'Health',              'Health program covering general medicine, NCDs, SRH, mental health'),
  ('00000000-0000-0000-0000-000000000001', 'PRG-WASH','Water Sanitation & Hygiene', 'WASH infrastructure, hygiene promotion, and supplies'),
  ('00000000-0000-0000-0000-000000000001', 'PRG-NUT', 'Nutrition',           'Therapeutic feeding, supplementation, and nutrition surveillance'),
  ('00000000-0000-0000-0000-000000000001', 'PRG-PRO', 'Protection',          'Child protection, GBV prevention, psychosocial support'),
  ('00000000-0000-0000-0000-000000000001', 'PRG-SHL', 'Shelter',             'Emergency shelter, transitional housing, and NFI distribution'),
  ('00000000-0000-0000-0000-000000000001', 'PRG-LOG', 'Logistics',           'Cross-cutting logistics coordination and supply chain support'),
  ('00000000-0000-0000-0000-000000000001', 'PRG-SUP', 'Supply Chain',        'Supply chain management, procurement, and fleet operations')
ON CONFLICT (org_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 23. STANDARD DEPARTMENTS (Master Catalogue — hierarchical codes)
-- ═══════════════════════════════════════════════════════════════════════════

INSERT INTO departments (org_id, name, code) VALUES
  -- Health cluster
  ('00000000-0000-0000-0000-000000000001', 'General Medicine',                          'HLT-GEN'),
  ('00000000-0000-0000-0000-000000000001', 'Non-Communicable Diseases',                 'HLT-NCD'),
  ('00000000-0000-0000-0000-000000000001', 'Sexual & Reproductive Health',              'HLT-SRH'),
  ('00000000-0000-0000-0000-000000000001', 'Mental Health',                             'HLT-MH'),
  ('00000000-0000-0000-0000-000000000001', 'Pharmaceutical',                            'HLT-PHA'),
  ('00000000-0000-0000-0000-000000000001', 'Laboratory',                                'HLT-LAB'),
  -- Logistics cluster
  ('00000000-0000-0000-0000-000000000001', 'Logistics — Procurement',                   'LOG-PRC'),
  ('00000000-0000-0000-0000-000000000001', 'Logistics — Warehousing',                   'LOG-WHS'),
  ('00000000-0000-0000-0000-000000000001', 'Logistics — Transport & Fleet',             'LOG-TRN'),
  ('00000000-0000-0000-0000-000000000001', 'Logistics — Fuel',                          'LOG-FLT')
ON CONFLICT (org_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 24. FUNCTIONS (Master Catalogue — process-level codes)
-- ═══════════════════════════════════════════════════════════════════════════

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'WHS-REC', 'Receiving', 'Goods receipt and inbound inspection'
FROM departments d WHERE d.code = 'LOG-WHS' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'WHS-PUT', 'Put-Away', 'Put-away and bin-to-stock allocation'
FROM departments d WHERE d.code = 'LOG-WHS' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'WHS-PIC', 'Picking', 'Order picking and staging'
FROM departments d WHERE d.code = 'LOG-WHS' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'WHS-PAC', 'Packing', 'Packing and dispatch preparation'
FROM departments d WHERE d.code = 'LOG-WHS' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'WHS-DSP', 'Dispatch', 'Final dispatch and outbound shipping'
FROM departments d WHERE d.code = 'LOG-WHS' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'PHA-SEL', 'Selection', 'Pharmaceutical product selection and formulary management'
FROM departments d WHERE d.code = 'HLT-PHA' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'PHA-STK', 'Stock Management', 'Pharmacy inventory control and stock management'
FROM departments d WHERE d.code = 'HLT-PHA' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'PHA-ORD', 'Ordering', 'Pharmaceutical ordering and procurement coordination'
FROM departments d WHERE d.code = 'HLT-PHA' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'PHA-DIS', 'Dispensing', 'Medicine dispensing and patient counseling'
FROM departments d WHERE d.code = 'HLT-PHA' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'PHA-SUP', 'Supplier Management', 'Pharmaceutical supplier evaluation and contracting'
FROM departments d WHERE d.code = 'HLT-PHA' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'PRC-ORD', 'Order Management', 'Purchase order creation, approval, and tracking'
FROM departments d WHERE d.code = 'LOG-PRC' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'PRC-CON', 'Contract Management', 'Supplier contract administration and compliance'
FROM departments d WHERE d.code = 'LOG-PRC' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'TRN-DSP', 'Fleet Dispatch', 'Vehicle scheduling, routing, and dispatch'
FROM departments d WHERE d.code = 'LOG-TRN' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'TRN-MNT', 'Fleet Maintenance', 'Vehicle maintenance and repair scheduling'
FROM departments d WHERE d.code = 'LOG-TRN' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'FLT-STK', 'Fuel Stock Management', 'Fuel receipt, storage, and dispensing'
FROM departments d WHERE d.code = 'LOG-FLT' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'LAB-TST', 'Lab Testing', 'Sample collection, testing, and results reporting'
FROM departments d WHERE d.code = 'HLT-LAB' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'LAB-STK', 'Lab Stock Management', 'Lab reagent and consumable inventory management'
FROM departments d WHERE d.code = 'HLT-LAB' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO functions (org_id, department_id, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', d.id, 'GEN-CNS', 'General Consultation', 'Outpatient consultation and primary care'
FROM departments d WHERE d.code = 'HLT-GEN' AND d.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 25. ENTITIES (Master Catalogue — objects within functions)
-- ═══════════════════════════════════════════════════════════════════════════

INSERT INTO entities (org_id, function_id, entity_type, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', f.id, 'warehouse', 'WH-CXB-CWH', 'Cox Bazar Central Warehouse', 'Primary central warehouse for CPI Bangladesh operations'
FROM functions f WHERE f.code = 'WHS-REC' AND f.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO entities (org_id, function_id, entity_type, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', f.id, 'warehouse', 'WH-CXB-SWH1', 'Sub-Warehouse Camp 4', 'Satellite warehouse at Camp 4 distribution point'
FROM functions f WHERE f.code = 'WHS-REC' AND f.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO entities (org_id, function_id, entity_type, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', f.id, 'warehouse', 'WH-CXB-COLD', 'Cold Chain Hub', 'Dedicated cold chain storage for vaccines and temperature-sensitive items'
FROM functions f WHERE f.code = 'WHS-REC' AND f.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO entities (org_id, function_id, entity_type, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', f.id, 'vehicle', 'VH-CXB-001', 'Toyota Land Cruiser (CXB-001)', 'Field operations vehicle based at Cox Bazar'
FROM functions f WHERE f.code = 'TRN-DSP' AND f.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

INSERT INTO entities (org_id, function_id, entity_type, code, name, description)
SELECT '00000000-0000-0000-0000-000000000001', f.id, 'vehicle', 'VH-CXB-002', 'Truck Isuzu 3-ton (CXB-002)', 'Cargo transport truck for supply runs'
FROM functions f WHERE f.code = 'TRN-DSP' AND f.org_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (org_id, code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 26. ENTITY ATTRIBUTES (dynamic properties of entities)
-- ═══════════════════════════════════════════════════════════════════════════

-- Warehouse capacity attributes
INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'total_capacity_m2', '1200', 'number', 1, true
FROM entities e WHERE e.code = 'WH-CXB-CWH'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'cold_chain_capable', 'true', 'boolean', 2, true
FROM entities e WHERE e.code = 'WH-CXB-CWH'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'has_generator', 'true', 'boolean', 3, false
FROM entities e WHERE e.code = 'WH-CXB-CWH'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'security_guard', 'true', 'boolean', 4, false
FROM entities e WHERE e.code = 'WH-CXB-CWH'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'operating_hours', 'Sun-Thu 8:00-17:00, Sat 9:00-13:00', 'text', 5, false
FROM entities e WHERE e.code = 'WH-CXB-CWH'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

-- Vehicle attributes
INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'registration_number', 'CXB-001 / DHAKA-METRO-A-1234', 'text', 1, true
FROM entities e WHERE e.code = 'VH-CXB-001'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'fuel_type', 'diesel', 'text', 2, true
FROM entities e WHERE e.code = 'VH-CXB-001'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'capacity_kg', '2500', 'number', 3, true
FROM entities e WHERE e.code = 'VH-CXB-001'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

INSERT INTO entity_attributes (entity_id, attribute_name, attribute_value, attribute_type, sort_order, is_required)
SELECT e.id, 'last_maintenance_date', '2026-06-15', 'date', 4, false
FROM entities e WHERE e.code = 'VH-CXB-001'
ON CONFLICT (entity_id, attribute_name) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 27. JUSTIFICATION CODES (MSF ordering justifications)
-- ═══════════════════════════════════════════════════════════════════════════

INSERT INTO justification_codes (code, name, description, category, sort_order) VALUES
  ('P', 'Recurring / Resupply',  'Recurring / Routine Resupply — scheduled orders based on AMC',        'recurring', 1),
  ('M', 'Campaign',              'Campaign — time-bound mass campaign (e.g., vaccination, distribution)','campaign',  2),
  ('E', 'Emergency',             'Emergency — acute shortage, disaster, or unexpected demand surge',     'emergency', 3),
  ('F', 'Forecast',              'Forecast — seasonal or projected increase based on trend analysis',    'forecast',  4),
  ('A', 'Asset',                 'Asset — procurement of capital equipment or durable assets',           'asset',     5),
  ('S', 'Special',               'Special — one-off, pilot, or research-related procurement',            'special',   6)
ON CONFLICT (code) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- 28. UOM CONVERSION FACTORS
-- ═══════════════════════════════════════════════════════════════════════════

-- Set base UoMs and conversion factors within each category
UPDATE units_of_measure SET conversion_factor = 1.0,   base_uom_id = id WHERE abbreviation = 'EA';   -- Each = base for count
UPDATE units_of_measure SET conversion_factor = 100.0, base_uom_id = (SELECT id FROM units_of_measure WHERE abbreviation = 'EA') WHERE abbreviation = 'BX';   -- 100 per box
UPDATE units_of_measure SET conversion_factor = 1000.0,base_uom_id = (SELECT id FROM units_of_measure WHERE abbreviation = 'EA') WHERE abbreviation = 'CTN';  -- 1000 per carton
UPDATE units_of_measure SET conversion_factor = 1.0,   base_uom_id = id WHERE abbreviation = 'KG';   -- Kg = base for weight
UPDATE units_of_measure SET conversion_factor = 0.001, base_uom_id = (SELECT id FROM units_of_measure WHERE abbreviation = 'KG') WHERE abbreviation = 'G';
UPDATE units_of_measure SET conversion_factor = 1.0,   base_uom_id = id WHERE abbreviation = 'L';    -- Liter = base for volume
UPDATE units_of_measure SET conversion_factor = 0.001, base_uom_id = (SELECT id FROM units_of_measure WHERE abbreviation = 'L') WHERE abbreviation = 'ML';
UPDATE units_of_measure SET conversion_factor = 12.0,  base_uom_id = (SELECT id FROM units_of_measure WHERE abbreviation = 'EA') WHERE abbreviation = 'DZ';   -- 12 per dozen
UPDATE units_of_measure SET conversion_factor = 2.0,   base_uom_id = (SELECT id FROM units_of_measure WHERE abbreviation = 'EA') WHERE abbreviation = 'PR';   -- 2 per pair

COMMIT;
