-- Policy §20 / §27 — Emergency plans and change control

-- name: CreateEmergencyPlan :one
INSERT INTO emergency_plans (
    org_id, plan_number, name, country_code, facility_id, scenario_type, scenario,
    demand_assumptions, service_level, prepositioned_stock_plan, rotation_plan,
    alternate_suppliers, alternate_warehouses, alternate_routes,
    alternate_power_sources, emergency_authority, paper_fallback_records,
    minimum_staffing, cold_chain_contingency, communications, escalation_contacts,
    security_controls, return_to_normal_plan, status, approved_by, approved_at,
    created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
RETURNING *;

-- name: GetEmergencyPlan :one
SELECT * FROM emergency_plans
WHERE id = $1 AND org_id = $2;

-- name: ListEmergencyPlans :many
SELECT * FROM emergency_plans
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateChangeControl :one
INSERT INTO change_controls (
    org_id, change_number, subject_type, subject_id, description, impact_assessment,
    risk_level, requires_quality_review, authorized_by, authorization_date,
    start_date, end_date, status, closure_evidence, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: GetChangeControl :one
SELECT * FROM change_controls
WHERE id = $1 AND org_id = $2;

-- name: ListChangeControls :many
SELECT * FROM change_controls
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;