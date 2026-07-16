-- ZarishLog — Seed Data
-- Run after schema.sql

-- ─── Roles ─────────────────────────────────────────────────────────────
INSERT INTO roles (code, name, description, level) VALUES
  ('R01', 'GLOBAL_ADMIN', 'Full system administration, all tenants', 1),
  ('R02', 'COUNTRY_REP', 'View-all, reporting only, no ops', 2),
  ('R03', 'THEME_MANAGER', 'View-all within one program theme, reporting', 2),
  ('R04', 'WAREHOUSE_OFFICER', 'Full central warehouse operations', 3),
  ('R05', 'WAREHOUSE_STOREKEEPER', 'Stock operations only (receive, issue, count)', 3),
  ('R06', 'ADMIN_LOG_OFFICER', 'Office asset/logistics stock management', 3),
  ('R07', 'DEPT_MANAGER', 'Budget/flow approval authority', 3),
  ('R08', 'DEPT_COORDINATOR', 'Validates stock flow at sub-warehouse level', 4),
  ('R09', 'DEPT_OFFICER', 'Day-to-day sub-warehouse operations', 4)
ON CONFLICT (code) DO NOTHING;

-- ─── Permissions ──────────────────────────────────────────────────────
INSERT INTO permissions (module, action) VALUES
  ('products', 'create'), ('products', 'read'), ('products', 'update'), ('products', 'delete'),
  ('categories', 'create'), ('categories', 'read'), ('categories', 'update'), ('categories', 'delete'),
  ('warehouses', 'create'), ('warehouses', 'read'), ('warehouses', 'update'), ('warehouses', 'delete'),
  ('stock', 'receive'), ('stock', 'issue'), ('stock', 'transfer'), ('stock', 'adjust'), ('stock', 'read'),
  ('qa', 'inspect'), ('qa', 'read'),
  ('assets', 'create'), ('assets', 'read'), ('assets', 'update'), ('assets', 'transfer'),
  ('users', 'create'), ('users', 'read'), ('users', 'update'),
  ('reports', 'read')
ON CONFLICT (module, action) DO NOTHING;

-- ─── Units of Measure ─────────────────────────────────────────────────
INSERT INTO units_of_measure (name, abbreviation, category) VALUES
  ('Each', 'EA', 'count'),
  ('Box', 'BX', 'count'),
  ('Carton', 'CTN', 'count'),
  ('Pallet', 'PL', 'count'),
  ('Kilogram', 'KG', 'weight'),
  ('Gram', 'G', 'weight'),
  ('Liter', 'L', 'volume'),
  ('Milliliter', 'ML', 'volume'),
  ('Meter', 'M', 'length'),
  ('Square Meter', 'M2', 'area'),
  ('Dozen', 'DZ', 'count'),
  ('Pair', 'PR', 'count')
ON CONFLICT (abbreviation) DO NOTHING;

-- ─── Sample Organization ──────────────────────────────────────────────
INSERT INTO organizations (id, name, code) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Center for Peace and Integrity', 'CPI')
ON CONFLICT (code) DO NOTHING;

-- ─── Sample Org Levels ────────────────────────────────────────────────
INSERT INTO org_levels (org_id, name, code, level) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Global HQ', 'CPI-GHQ', 1),
  ('00000000-0000-0000-0000-000000000001', 'Country Office - Bangladesh', 'CPI-BD', 2),
  ('00000000-0000-0000-0000-000000000001', 'Project Office - Cox Bazar', 'CPI-CXB', 3),
  ('00000000-0000-0000-0000-000000000001', 'Program Site - Camp 5', 'CPI-CXB-C5', 4)
ON CONFLICT (org_id, code) DO NOTHING;

-- ─── Sample Programs ──────────────────────────────────────────────────
INSERT INTO programs (org_id, code, name) VALUES
  ('00000000-0000-0000-0000-000000000001', 'H&N', 'Health and Nutrition'),
  ('00000000-0000-0000-0000-000000000001', 'WASH', 'Water, Sanitation and Hygiene'),
  ('00000000-0000-0000-0000-000000000001', 'LVL', 'Livelihood'),
  ('00000000-0000-0000-0000-000000000001', 'EDU', 'Education'),
  ('00000000-0000-0000-0000-000000000001', 'EPRR', 'Emergency Preparedness and Response'),
  ('00000000-0000-0000-0000-000000000001', 'R&L', 'Research and Learning')
ON CONFLICT (org_id, code) DO NOTHING;

-- ─── Sample Warehouse ─────────────────────────────────────────────────
INSERT INTO warehouses (org_id, name, code, type, city, country) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Cox Bazar Central Warehouse', 'CXB-CWH', 'central', 'Cox Bazar', 'Bangladesh')
ON CONFLICT (org_id, code) DO NOTHING;

-- ─── Sample Locations ──────────────────────────────────────────────────
INSERT INTO locations (warehouse_id, code, name, type) VALUES
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'), 'RECV', 'Receiving Area', 'area'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'), 'GEN-A', 'General Storage A', 'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'), 'GEN-B', 'General Storage B', 'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'), 'COLD', 'Cold Chain Storage', 'zone'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'), 'QA', 'QA/Quarantine Area', 'area'),
  ((SELECT id FROM warehouses WHERE code = 'CXB-CWH'), 'DISP', 'Dispatch Area', 'area')
ON CONFLICT (warehouse_id, code) DO NOTHING;

-- ─── Sample Product Categories ─────────────────────────────────────────
INSERT INTO product_categories (org_id, name, description, unspsc) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Medicines & Drugs', 'Pharmaceutical products including tablets, capsules, injectables', '51000000'),
  ('00000000-0000-0000-0000-000000000001', 'Medical Supplies', 'Consumable medical supplies including dressings, gloves, syringes', '42000000'),
  ('00000000-0000-0000-0000-000000000001', 'Medical Equipment', 'Durable medical equipment including diagnostic, therapeutic devices', '42100000'),
  ('00000000-0000-0000-0000-000000000001', 'Nutrition', 'Nutritional products including therapeutic foods, supplements', '51000000'),
  ('00000000-0000-0000-0000-000000000001', 'WASH Supplies', 'Water, sanitation and hygiene supplies', '47000000'),
  ('00000000-0000-0000-0000-000000000001', 'Office & Admin', 'Office supplies, furniture, IT equipment', '44000000')
ON CONFLICT (org_id, name) DO NOTHING;

-- ─── Sample Products ───────────────────────────────────────────────────
INSERT INTO products (org_id, category_id, uom_id, sku, name, item_type, is_batch_tracked, is_expiry_tracked, is_cold_chain, status) VALUES
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'Medicines & Drugs'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
   'MED-AMOX-500', 'Amoxicillin 500mg Capsules', 'drug', true, true, false, 'active'),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'Medicines & Drugs'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
   'MED-PCM-500', 'Paracetamol 500mg Tablets', 'drug', true, true, false, 'active'),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'Medical Supplies'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
   'SUP-LTX-MED', 'Latex Examination Gloves (Box 100)', 'medical_supply', true, true, false, 'active'),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'Medical Supplies'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
   'SUP-SYG-5ML', 'Syringe 5ml (Box 100)', 'medical_supply', true, true, false, 'active'),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'Medical Supplies'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'BX'),
   'SUP-GAUZE', 'Gauze Swabs Sterile 10x10cm (Box 100)', 'medical_supply', true, true, false, 'active'),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'Nutrition'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'CTN'),
   'NUT-RUTF', 'Ready-to-Use Therapeutic Food (Carton)', 'nutrition', true, true, false, 'active'),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'WASH Supplies'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
   'WASH-SOAP-BAR', 'Soap Bar 200g', 'material', false, false, false, 'active'),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'WASH Supplies'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
   'WASH-BKT-10L', 'Plastic Bucket 10L', 'material', false, false, false, 'active'),
  ('00000000-0000-0000-0000-000000000001',
   (SELECT id FROM product_categories WHERE name = 'Medical Equipment'),
   (SELECT id FROM units_of_measure WHERE abbreviation = 'EA'),
   'EQP-BP-MANUAL', 'Manual Blood Pressure Monitor', 'equipment', false, true, false, 'active')
ON CONFLICT (org_id, sku) DO NOTHING;
