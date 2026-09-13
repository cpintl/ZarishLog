-- ZarishLog — CPI Bangladesh Medical Warehouse Policy Controls
-- PostgreSQL 18, Multi-tenant, UUIDv7, Audit columns, RLS
-- Implements the controlled registers and workflows required by the
-- "CPI BANGLADESH MEDICAL WAREHOUSE POLICY" (v1.0):
--   §7  QMS: deviations, CAPA, training/competency
--   §8  Regulatory & legal compliance in Bangladesh (DGDA/DNC matrix)
--   §11 Donations acceptance/rejection registry
--   §12 Quality release decisions for quarantined stock
--   §13 Temperature/humidity monitoring, alarm & excursion handling, calibration
--   §14 Controlled-drug register and short-expiry (near-expiry) review
--   §16 Dispatch waybills and delivery confirmation
--   §17 Complaints, recalls (incl. mock recall), falsified-product triage
--   §20 Emergency medical supply plans
--   §27 Exceptions and change control
-- Idempotent: uses IF NOT EXISTS / DO $$ blocks throughout.

-- ─── 1. EXTEND DOC TYPE ENUM ──────────────────────────────────────────
-- Donations flow through goods receiving; the stock ledger records the event.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'donation' AND enumtypid = 'doc_type'::regtype) THEN
    ALTER TYPE doc_type ADD VALUE 'donation' AFTER 'return';
  END IF;
END;
$$;

-- ─── 2. POLICY REFERENCE DICTIONARIES ──────────────────────────────────

-- Deviation / nonconformance categories (§7.2)
CREATE TABLE IF NOT EXISTS deviation_categories (
    code               text PRIMARY KEY,
    name               text NOT NULL,
    description        text,
    default_risk_level text NOT NULL DEFAULT 'medium'
                       CHECK (default_risk_level IN ('low', 'medium', 'high', 'critical')),
    sort_order         int NOT NULL DEFAULT 0,
    is_active          boolean NOT NULL DEFAULT true
);

INSERT INTO deviation_categories (code, name, description, default_risk_level, sort_order) VALUES
    ('DAMAGED_GOODS',       'Damaged or suspect goods',          'Damaged, contaminated, or suspect incoming or stored product', 'medium', 1),
    ('TEMPERATURE_EXCURSION','Temperature excursion',            'Out-of-range temperature or freeze event during storage/transport', 'high', 2),
    ('MISSING_RECORDS',     'Missing or incomplete records',     'Missing GRN, stock card, batch/expiry, or transaction records', 'low', 3),
    ('STOCK_DISCREPANCY',   'Stock discrepancy',                 'Physical vs theoretical stock variance', 'medium', 4),
    ('UNAUTHORIZED_ACCESS', 'Unauthorized access',               'Access to products, data, or premises without authorization', 'high', 5),
    ('FALSIFICATION',       'Falsification concern',             'Suspected falsified, counterfeit, or tampered product', 'critical', 6),
    ('INCORRECT_PICKING',   'Incorrect picking',                 'Wrong product, strength, batch, expiry, serial, or quantity issued', 'medium', 7),
    ('LATE_REPORTING',      'Late or missing reporting',         'Missed daily/weekly/monthly reporting obligations', 'low', 8),
    ('QUALIFICATION_FAILURE','Qualification or validation failure','Equipment, mapping, or validation activity failed', 'medium', 9),
    ('CONTROLLED_SUBSTANCE','Controlled-substance deviation',    'Controlled-drug register, reconciliation, or custody event', 'critical', 10),
    ('OTHER',               'Other',                             'Any material departure from an approved process or requirement', 'medium', 99)
ON CONFLICT (code) DO NOTHING;

-- Complaint severities (§17.2)
CREATE TABLE IF NOT EXISTS complaint_severities (
    code            text PRIMARY KEY,
    name            text NOT NULL,
    description     text,
    immediate_escalation boolean NOT NULL DEFAULT false,
    sort_order      int NOT NULL DEFAULT 0,
    is_active       boolean NOT NULL DEFAULT true
);

INSERT INTO complaint_severities (code, name, description, immediate_escalation, sort_order) VALUES
    ('CRITICAL',  'Critical',  'Suspected harm, falsification, contamination, wrong product, controlled-product diversion, or widespread quality risk', true, 1),
    ('MAJOR',     'Major',     'Material quality, labeling, or traceability failure without immediate harm', false, 2),
    ('MINOR',     'Minor',     'Localized or cosmetic issue with no material product-risk', false, 3)
ON CONFLICT (code) DO NOTHING;

-- Recall types (§17.3) — actual recalls and annual mock recall exercises
CREATE TABLE IF NOT EXISTS recall_types (
    code        text PRIMARY KEY,
    name        text NOT NULL,
    description text,
    sort_order  int NOT NULL DEFAULT 0,
    is_active   boolean NOT NULL DEFAULT true
);

INSERT INTO recall_types (code, name, description, sort_order) VALUES
    ('ACTUAL', 'Actual recall', 'Recall of distributed product from the market or field', 1),
    ('MOCK',   'Mock recall',   'Scheduled exercise to test recall traceability and procedure', 2)
ON CONFLICT (code) DO NOTHING;

-- ─── 3. REGULATORY & LEGAL COMPLIANCE (§8) ─────────────────────────────

CREATE TABLE IF NOT EXISTS regulatory_approvals (
    id                uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id            uuid NOT NULL REFERENCES organizations(id),
    scope_type        text NOT NULL CHECK (scope_type IN ('warehouse', 'product', 'product_category', 'supplier', 'organization')),
    scope_id          text NOT NULL,
    authority         text NOT NULL CHECK (authority IN ('DGDA', 'DNC', 'NBR', 'CCI_E', 'Customs', 'Municipality', 'Other')),
    license_type      text NOT NULL CHECK (license_type IN ('wholesale_license', 'retail_license', 'import_permit', 'controlled_substance_permit', 'premises_approval', 'import_clearance', 'registration', 'other')),
    reference_number  text NOT NULL,
    product_scope     text,
    conditions        text,
    issue_date        date,
    expiry_date       date,
    renewal_lead_days int NOT NULL DEFAULT 60,
    responsible_owner text,
    status            text NOT NULL DEFAULT 'draft'
                      CHECK (status IN ('draft', 'active', 'expiring_soon', 'expired', 'revoked', 'suspended')),
    evidence_url      text,
    notes             text,
    created_by        text,
    updated_by        text,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, authority, reference_number)
);

CREATE INDEX IF NOT EXISTS idx_regulatory_approvals_org_expiry ON regulatory_approvals(org_id, expiry_date);

-- ─── 4. QUALITY MANAGEMENT SYSTEM — DEVIATIONS & CAPA (§7.2 / §7.3) ─────

CREATE TABLE IF NOT EXISTS deviations (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                uuid NOT NULL REFERENCES organizations(id),
    deviation_number      text NOT NULL,
    category_code         text REFERENCES deviation_categories(code),
    occurred_at           timestamptz NOT NULL DEFAULT now(),
    reported_at           timestamptz NOT NULL DEFAULT now(),
    source_document_type  text CHECK (source_document_type IN ('grn', 'issue', 'transfer', 'adjustment', 'disposal', 'return', 'temperature_excursion', 'stock_count', 'audit', 'inspection', 'complaint', 'other')),
    source_document_id    text,
    product_id            uuid REFERENCES products(id),
    batch_id              uuid REFERENCES batches(id),
    warehouse_id          uuid REFERENCES warehouses(id),
    location_id           uuid REFERENCES locations(id),
    description           text NOT NULL,
    containment_action    text,
    impact_assessment     text,
    root_cause            text,
    risk_rating           text NOT NULL DEFAULT 'medium' CHECK (risk_rating IN ('low', 'medium', 'high', 'critical')),
    disposition           text CHECK (disposition IN ('accepted', 'rejected', 'quarantined', 'destroyed', 'returned', 'reworked', 'blocked', 'other')),
    status                text NOT NULL DEFAULT 'open'
                          CHECK (status IN ('open', 'under_investigation', 'contained', 'resolved', 'closed')),
    responsible_owner     text,
    due_date              date,
    escalated             boolean NOT NULL DEFAULT false,
    escalated_to          text,
    created_by            text,
    updated_by            text,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, deviation_number)
);

CREATE INDEX IF NOT EXISTS idx_deviations_org_status ON deviations(org_id, status);
CREATE INDEX IF NOT EXISTS idx_deviations_org_product ON deviations(org_id, product_id);

-- CAPA register (corrective & preventive actions)
CREATE TABLE IF NOT EXISTS capa_actions (
    id                        uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                    uuid NOT NULL REFERENCES organizations(id),
    capa_number               text NOT NULL,
    source_type               text NOT NULL CHECK (source_type IN ('deviation', 'complaint', 'recall', 'audit', 'inspection', 'stock_discrepancy', 'supplier', 'trend', 'other')),
    source_id                 text,
    capa_type                 text NOT NULL CHECK (capa_type IN ('corrective', 'preventive')),
    title                     text NOT NULL,
    root_cause                text,
    action_plan               text NOT NULL,
    responsible_owner         text,
    due_date                  date,
    status                    text NOT NULL DEFAULT 'open'
                              CHECK (status IN ('open', 'in_progress', 'overdue', 'completed', 'closed')),
    effectiveness_check       text,
    effectiveness_verified    boolean NOT NULL DEFAULT false,
    effectiveness_verified_by text,
    verified_at               timestamptz,
    closed_by                 text,
    closed_at                 timestamptz,
    created_by                text,
    updated_by                text,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, capa_number)
);

CREATE INDEX IF NOT EXISTS idx_capa_actions_org_status ON capa_actions(org_id, status);

-- Training & competency records (§7.4)
CREATE TABLE IF NOT EXISTS training_records (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                uuid NOT NULL REFERENCES organizations(id),
    user_id               uuid REFERENCES users(id),
    training_title        text NOT NULL,
    training_type         text NOT NULL DEFAULT 'initial' CHECK (training_type IN ('initial', 'refresher', 'emergency', 'contractor_safety')),
    topic                 text CHECK (topic IN ('sop', 'product_identification', 'fefo', 'hygiene', 'security', 'data_integrity', 'cold_chain', 'controlled_products', 'hazardous_materials', 'emergency', 'falsified_products', 'recall', 'complaints', 'incident_reporting', 'other')),
    training_date         date NOT NULL DEFAULT CURRENT_DATE,
    method                text CHECK (method IN ('classroom', 'on_the_job', 'online', 'assessment', 'other')),
    trainer               text,
    assessed              boolean NOT NULL DEFAULT false,
    assessment_result     text CHECK (assessment_result IN ('pass', 'fail')),
    competency_valid_until date,
    certificate_url       text,
    created_by            text,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_training_records_org_user ON training_records(org_id, user_id);

-- ─── 5. COLD CHAIN & ENVIRONMENTAL CONTROL (§13) ───────────────────────

-- Continuous/periodic temperature & humidity observations
CREATE TABLE IF NOT EXISTS temperature_monitoring_entries (
    id             uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id         uuid NOT NULL REFERENCES organizations(id),
    warehouse_id   uuid NOT NULL REFERENCES warehouses(id),
    location_id    uuid REFERENCES locations(id),
    equipment_id   uuid REFERENCES entities(id),
    device_name    text,
    monitor_type   text NOT NULL CHECK (monitor_type IN ('temperature', 'humidity', 'freeze_indicator', 'vvm')),
    reading_value  numeric(8,2) NOT NULL,
    min_threshold  numeric(8,2),
    max_threshold  numeric(8,2),
    recorded_at    timestamptz NOT NULL DEFAULT now(),
    recorded_by    text,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_temp_mon_entries_org_recorded ON temperature_monitoring_entries(org_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_temp_mon_entries_location ON temperature_monitoring_entries(warehouse_id, location_id);

-- Alarm & excursion handling (§13.3): every out-of-range event triggers
-- quarantine ("DO NOT USE"), preservation of data, notification, assessment,
-- documented disposition, and CAPA.
CREATE TABLE IF NOT EXISTS temperature_excursions (
    id                 uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id             uuid NOT NULL REFERENCES organizations(id),
    warehouse_id       uuid NOT NULL REFERENCES warehouses(id),
    location_id        uuid REFERENCES locations(id),
    equipment_id       uuid REFERENCES entities(id),
    device_name        text,
    excursion_type     text NOT NULL CHECK (excursion_type IN ('temperature', 'humidity', 'freeze', 'power_failure', 'equipment_failure', 'monitoring_gap')),
    started_at         timestamptz NOT NULL DEFAULT now(),
    ended_at           timestamptz,
    min_value          numeric(8,2),
    max_value          numeric(8,2),
    expected_min       numeric(8,2),
    expected_max       numeric(8,2),
    affected_products  text,
    affected_batches   text,
    quantity_affected  numeric(12,3),
    quarantined        boolean NOT NULL DEFAULT true,
    quarantine_reference text,
    disposition        text NOT NULL DEFAULT 'pending'
                       CHECK (disposition IN ('pending', 'released', 'restricted_use', 'returned', 'destroyed', 'investigating')),
    notified_quality   boolean NOT NULL DEFAULT false,
    notified_manager   boolean NOT NULL DEFAULT false,
    technical_advice   text,
    root_cause         text,
    capa_id            uuid REFERENCES capa_actions(id),
    status             text NOT NULL DEFAULT 'open'
                       CHECK (status IN ('open', 'under_review', 'dispositioned', 'closed')),
    created_by         text,
    updated_by         text,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_temp_excursions_org_status ON temperature_excursions(org_id, status);
CREATE INDEX IF NOT EXISTS idx_temp_excursions_org_started ON temperature_excursions(org_id, started_at DESC);

-- ─── 6. QUALITY RELEASE & CONTROLLED PRODUCTS (§12.2 / §14.3) ──────────

-- Documented authorization to move stock from quarantine into usable stock,
-- or an alternative disposition decision, traceable to product/batch/evidence.
CREATE TABLE IF NOT EXISTS stock_release_records (
    id                   uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id               uuid NOT NULL REFERENCES organizations(id),
    release_number       text NOT NULL,
    warehouse_id         uuid NOT NULL REFERENCES warehouses(id),
    product_id           uuid NOT NULL REFERENCES products(id),
    batch_id             uuid REFERENCES batches(id),
    location_id          uuid REFERENCES locations(id),
    quantity             numeric(12,3) NOT NULL,
    quarantine_reference text,
    evidence_reviewed    text,
    decision             text NOT NULL CHECK (decision IN ('release', 'restricted_release', 'reject', 'destruction', 'return')),
    decision_date        date NOT NULL DEFAULT CURRENT_DATE,
    decided_by           text NOT NULL,
    notes                text,
    created_by           text,
    updated_by           text,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, release_number)
);

CREATE INDEX IF NOT EXISTS idx_stock_release_records_org ON stock_release_records(org_id);

-- Controlled-drug register (§14.3): locked, access-controlled products;
-- ledger reconciles at every receipt/issue and periodic frequency.
CREATE TABLE IF NOT EXISTS controlled_stock_register (
    id                   uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id               uuid NOT NULL REFERENCES organizations(id),
    product_id           uuid NOT NULL REFERENCES products(id),
    batch_id             uuid REFERENCES batches(id),
    warehouse_id         uuid NOT NULL REFERENCES warehouses(id),
    register_date        date NOT NULL DEFAULT CURRENT_DATE,
    transaction_type     text NOT NULL CHECK (transaction_type IN ('receipt', 'issue', 'transfer_in', 'transfer_out', 'return', 'disposal', 'reconciliation', 'opening_balance')),
    reference_document   text,
    quantity             numeric(12,3) NOT NULL,
    balance_after        numeric(12,3) NOT NULL,
    received_from        text,
    issued_to            text,
    authorized_recipient text,
    signature            text,
    discrepancy          numeric(12,3) NOT NULL DEFAULT 0,
    reconciled_at        timestamptz,
    reconciled_by        text,
    created_by           text,
    created_at           timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_controlled_register_org_product ON controlled_stock_register(org_id, product_id);
CREATE INDEX IF NOT EXISTS idx_controlled_register_org_date ON controlled_stock_register(org_id, register_date);

-- ─── 7. DONATIONS (§11.3) ───────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS donations (
    id                       uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                   uuid NOT NULL REFERENCES organizations(id),
    donation_number          text NOT NULL,
    donor_name               text NOT NULL,
    donor_contact            text,
    offer_date               date NOT NULL DEFAULT CURRENT_DATE,
    decision                 text CHECK (decision IN ('accepted', 'rejected', 'quarantined', 'pending')),
    decision_date            date,
    decided_by               text,
    needs_assessment         text,
    proposed_recipient       text,
    approval_reference       text,
    transport_responsibility text,
    customs_notes            text,
    disposal_responsibility  text,
    certificate_number       text,
    certificate_issued_date  date,
    registry_reference       text,
    status                   text NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft', 'pending', 'accepted', 'rejected', 'quarantined', 'received')),
    notes                    text,
    created_by               text,
    updated_by               text,
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, donation_number)
);

CREATE TABLE IF NOT EXISTS donation_line_items (
    id                         uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    donation_id                uuid NOT NULL REFERENCES donations(id) ON DELETE CASCADE,
    product_id                 uuid NOT NULL REFERENCES products(id),
    batch_number               text,
    expiry_date                date,
    quantity                   numeric(12,3) NOT NULL,
    uom                        text,
    remaining_shelf_life_months int,
    condition                  text CHECK (condition IN ('intact', 'short_dated', 'damaged', 'unlabelled', 'suspected_falsified')),
    shelf_life_compliant       boolean,
    authorized                 boolean NOT NULL DEFAULT false,
    notes                      text
);

CREATE INDEX IF NOT EXISTS idx_donations_org_status ON donations(org_id, status);
CREATE INDEX IF NOT EXISTS idx_donation_line_items_donation ON donation_line_items(donation_id);

-- ─── 8. COMPLAINTS, RECALLS & FALSIFIED PRODUCTS (§17) ─────────────────

CREATE TABLE IF NOT EXISTS complaints (
    id                         uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                     uuid NOT NULL REFERENCES organizations(id),
    complaint_number           text NOT NULL,
    received_date              date NOT NULL DEFAULT CURRENT_DATE,
    received_via               text CHECK (received_via IN ('field_visit', 'phone', 'email', 'form', 'third_party', 'other')),
    severity                   text REFERENCES complaint_severities(code),
    category                   text CHECK (category IN ('suspected_harm', 'falsification', 'contamination', 'wrong_product', 'controlled_diversion', 'quality_widespread', 'packaging', 'other')),
    product_id                 uuid REFERENCES products(id),
    batch_id                   uuid REFERENCES batches(id),
    description                text NOT NULL,
    reporter_name              text,
    reporter_contact           text,
    immediate_escalation       boolean NOT NULL DEFAULT false,
    escalated                  boolean NOT NULL DEFAULT false,
    escalated_to               text,
    manufacturer_notified      boolean NOT NULL DEFAULT false,
    competent_authority_notified boolean NOT NULL DEFAULT false,
    status                     text NOT NULL DEFAULT 'open'
                               CHECK (status IN ('open', 'under_review', 'investigating', 'resolved', 'closed')),
    resolution_notes           text,
    capa_id                    uuid REFERENCES capa_actions(id),
    created_by                 text,
    updated_by                 text,
    created_at                 timestamptz NOT NULL DEFAULT now(),
    updated_at                 timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, complaint_number)
);

CREATE INDEX IF NOT EXISTS idx_complaints_org_severity ON complaints(org_id, severity);
CREATE INDEX IF NOT EXISTS idx_complaints_org_status ON complaints(org_id, status);

CREATE TABLE IF NOT EXISTS recalls (
    id                   uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id               uuid NOT NULL REFERENCES organizations(id),
    recall_number        text NOT NULL,
    recall_type          text NOT NULL REFERENCES recall_types(code),
    recall_date          date NOT NULL DEFAULT CURRENT_DATE,
    product_id           uuid REFERENCES products(id),
    subject_product      text NOT NULL,
    batch_numbers        text,
    reason               text NOT NULL,
    initiated_by         text,
    authorized_by        text,
    communications       text,
    disposition_plan     text,
    status               text NOT NULL DEFAULT 'draft'
                         CHECK (status IN ('draft', 'in_progress', 'pending_disposition', 'reconciled', 'closed')),
    closure_evidence     text,
    created_by           text,
    updated_by           text,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, recall_number)
);

CREATE TABLE IF NOT EXISTS recall_line_items (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    recall_id             uuid NOT NULL REFERENCES recalls(id) ON DELETE CASCADE,
    recipient             text NOT NULL,
    location              text,
    product_id            uuid REFERENCES products(id),
    batch_number          text,
    quantity_issued       numeric(12,3) NOT NULL DEFAULT 0,
    quantity_recovered    numeric(12,3) NOT NULL DEFAULT 0,
    quantity_outstanding  numeric(12,3) NOT NULL DEFAULT 0,
    storage_disposition   text CHECK (storage_disposition IN ('quarantined', 'returned', 'destroyed', 'secured')),
    recovered_at          timestamptz,
    notes                 text
);

CREATE INDEX IF NOT EXISTS idx_recalls_org_status ON recalls(org_id, status);
CREATE INDEX IF NOT EXISTS idx_recall_line_items_recall ON recall_line_items(recall_id);

-- ─── 9. DISPATCH & DELIVERY CONFIRMATION (§16) ─────────────────────────

-- Dispatch documents: GIN/waybill/packing list at dispatch.
CREATE TABLE IF NOT EXISTS dispatch_waybills (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                uuid NOT NULL REFERENCES organizations(id),
    waybill_number        text NOT NULL,
    issue_id              uuid REFERENCES stock_issues(id),
    transfer_id           uuid REFERENCES stock_transfers(id),
    sender_warehouse_id   uuid NOT NULL REFERENCES warehouses(id),
    recipient_warehouse_id uuid REFERENCES warehouses(id),
    recipient             text,
    dispatch_date         date NOT NULL DEFAULT CURRENT_DATE,
    carrier               text,
    vehicle_number        text,
    carton_count          int,
    packing_list_reference text,
    storage_requirement   text,
    temperature_sensitive boolean NOT NULL DEFAULT false,
    condition             text CHECK (condition IN ('intact', 'damaged', 'sealed', 'tampered')),
    document_refs         text,
    status                text NOT NULL DEFAULT 'planned'
                          CHECK (status IN ('planned', 'dispatched', 'in_transit', 'delivered', 'cancelled')),
    created_by            text,
    updated_by            text,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, waybill_number)
);

CREATE INDEX IF NOT EXISTS idx_dispatch_waybills_org_status ON dispatch_waybills(org_id, status);

-- Delivery confirmation by the receiving unit (§16.4).
CREATE TABLE IF NOT EXISTS delivery_confirmations (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                uuid NOT NULL REFERENCES organizations(id),
    waybill_id            uuid NOT NULL REFERENCES dispatch_waybills(id),
    confirming_warehouse_id uuid REFERENCES warehouses(id),
    recipient_name        text,
    confirmed_date        date NOT NULL DEFAULT CURRENT_DATE,
    delivery_status       text NOT NULL CHECK (delivery_status IN ('received', 'partial', 'rejected', 'damaged', 'lost', 'failed')),
    items_conformed       boolean,
    quantity_conformed    boolean,
    batch_conformed       boolean,
    expiry_conformed      boolean,
    condition_conformed   boolean,
    temperature_conformed boolean,
    unfilled_quantity     text,
    substitutions         text,
    discrepancies         text,
    follow_up_commitments text,
    confirmed_by          text,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_delivery_confirmations_org_waybill ON delivery_confirmations(org_id, waybill_id);

-- ─── 10. SHORT-EXPIRY (NEAR-EXPIRY) REVIEW (§14.4) ─────────────────────

CREATE TABLE IF NOT EXISTS short_expiry_reviews (
    id                         uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                     uuid NOT NULL REFERENCES organizations(id),
    review_date                date NOT NULL DEFAULT CURRENT_DATE,
    product_id                 uuid NOT NULL REFERENCES products(id),
    batch_id                   uuid REFERENCES batches(id),
    warehouse_id               uuid NOT NULL REFERENCES warehouses(id),
    expiry_date                date NOT NULL,
    remaining_shelf_life_months int NOT NULL,
    expected_consumption       text,
    transfer_option            text,
    donor_condition            text,
    regulatory_restriction     text,
    decision                   text NOT NULL CHECK (decision IN ('use_before_expiry', 'transfer', 'redistribute', 'return', 'donate', 'destroy', 'pending')),
    decision_by                text,
    decision_date              date,
    notes                      text,
    created_by                 text,
    created_at                 timestamptz NOT NULL DEFAULT now(),
    updated_at                 timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_short_expiry_reviews_org_expiry ON short_expiry_reviews(org_id, expiry_date);

-- ─── 11. EMERGENCY PLANS & CHANGE CONTROL (§20 / §27) ──────────────────

CREATE TABLE IF NOT EXISTS emergency_plans (
    id                          uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                      uuid NOT NULL REFERENCES organizations(id),
    plan_number                 text NOT NULL,
    name                        text NOT NULL,
    country_code                text,
    facility_id                 uuid REFERENCES warehouses(id),
    scenario_type               text NOT NULL CHECK (scenario_type IN ('outbreak', 'displacement', 'cyclone', 'flood', 'fire', 'security', 'scale_up', 'power_loss', 'cold_chain_failure', 'supplier_interruption', 'border_delay', 'system_outage', 'warehouse_loss', 'staff_shortage', 'other')),
    scenario                    text,
    demand_assumptions          text,
    service_level               text,
    prepositioned_stock_plan    text,
    rotation_plan               text,
    alternate_suppliers         text,
    alternate_warehouses        text,
    alternate_routes            text,
    alternate_power_sources     text,
    emergency_authority         text,
    paper_fallback_records      boolean NOT NULL DEFAULT true,
    minimum_staffing            text,
    cold_chain_contingency      text,
    communications              text,
    escalation_contacts         text,
    security_controls           text,
    return_to_normal_plan       text,
    status                      text NOT NULL DEFAULT 'draft'
                                CHECK (status IN ('draft', 'approved', 'active', 'exercised', 'superseded')),
    approved_by                 text,
    approved_at                 timestamptz,
    created_by                  text,
    updated_by                  text,
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, plan_number)
);

CREATE INDEX IF NOT EXISTS idx_emergency_plans_org_status ON emergency_plans(org_id, status);

CREATE TABLE IF NOT EXISTS change_controls (
    id                      uuid PRIMARY KEY DEFAULT uuid_generate_v7(),
    org_id                  uuid NOT NULL REFERENCES organizations(id),
    change_number           text NOT NULL,
    subject_type            text NOT NULL CHECK (subject_type IN ('product', 'supplier', 'warehouse_layout', 'equipment', 'software', 'data_structure', 'program', 'service_level', 'storage_condition', 'legal_requirement', 'sop')),
    subject_id              text,
    description             text NOT NULL,
    impact_assessment       text,
    risk_level              text NOT NULL DEFAULT 'medium' CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    requires_quality_review boolean NOT NULL DEFAULT true,
    authorized_by           text,
    authorization_date      date,
    start_date              date,
    end_date                date,
    status                  text NOT NULL DEFAULT 'draft'
                            CHECK (status IN ('draft', 'assessed', 'approved', 'implemented', 'closed', 'rejected')),
    closure_evidence        text,
    created_by              text,
    updated_by              text,
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    UNIQUE(org_id, change_number)
);

CREATE INDEX IF NOT EXISTS idx_change_controls_org_status ON change_controls(org_id, status);

-- ─── 12. RLS POLICIES ──────────────────────────────────────────────────

-- Enable RLS on all new tenant-scoped tables and apply org isolation,
-- reusing the finer-grained policy expression when session dimensions are set.
DO $$
DECLARE
  tbl text;
  tables text[] := ARRAY[
    'regulatory_approvals', 'deviations', 'capa_actions', 'training_records',
    'temperature_monitoring_entries', 'temperature_excursions',
    'stock_release_records', 'controlled_stock_register',
    'donations', 'complaints', 'recalls',
    'dispatch_waybills', 'delivery_confirmations',
    'short_expiry_reviews', 'emergency_plans', 'change_controls'
  ];
BEGIN
  FOREACH tbl IN ARRAY tables
  LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', tbl);
    EXECUTE format('DROP POLICY IF EXISTS org_isolation ON %I', tbl);
    EXECUTE format('
      CREATE POLICY org_isolation ON %I
        USING (%s)
        WITH CHECK (%s)
    ', tbl, app.rls_policy_expression(), app.rls_policy_expression());
  END LOOP;
END;
$$;

-- Child line-item tables resolve org through their parent register.
DO $$
BEGIN
  -- donation_line_items -> donations
  ALTER TABLE donation_line_items ENABLE ROW LEVEL SECURITY;
  DROP POLICY IF EXISTS org_isolation ON donation_line_items;
  CREATE POLICY org_isolation ON donation_line_items
    USING (donation_id IN (SELECT d.id FROM donations d WHERE d.org_id = app.current_org_id()::uuid))
    WITH CHECK (donation_id IN (SELECT d.id FROM donations d WHERE d.org_id = app.current_org_id()::uuid));

  -- recall_line_items -> recalls
  ALTER TABLE recall_line_items ENABLE ROW LEVEL SECURITY;
  DROP POLICY IF EXISTS org_isolation ON recall_line_items;
  CREATE POLICY org_isolation ON recall_line_items
    USING (recall_id IN (SELECT r.id FROM recalls r WHERE r.org_id = app.current_org_id()::uuid))
    WITH CHECK (recall_id IN (SELECT r.id FROM recalls r WHERE r.org_id = app.current_org_id()::uuid));
END;
$$;