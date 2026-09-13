-- Policy §7 — QMS: Deviations, CAPA, Training & Competency

-- name: CreateDeviation :one
INSERT INTO deviations (
    org_id, deviation_number, category_code, occurred_at, reported_at,
    source_document_type, source_document_id, product_id, batch_id,
    warehouse_id, location_id, description, containment_action,
    impact_assessment, root_cause, risk_rating, disposition, status,
    responsible_owner, due_date, escalated, escalated_to, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
RETURNING *;

-- name: GetDeviation :one
SELECT * FROM deviations
WHERE id = $1 AND org_id = $2;

-- name: ListDeviations :many
SELECT * FROM deviations
WHERE org_id = $1
ORDER BY occurred_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDeviationsByStatus :many
SELECT * FROM deviations
WHERE org_id = $1 AND status = $2
ORDER BY occurred_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateDeviationStatus :exec
UPDATE deviations SET
    status = $1, disposition = $2, root_cause = $3,
    escalated = $4, escalated_to = $5, updated_by = $6, updated_at = now()
WHERE id = $7 AND org_id = $8;

-- name: CreateCAPAAction :one
INSERT INTO capa_actions (
    org_id, capa_number, source_type, source_id, capa_type, title,
    root_cause, action_plan, responsible_owner, due_date, status,
    effectiveness_check, effectiveness_verified, effectiveness_verified_by,
    verified_at, closed_by, closed_at, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
RETURNING *;

-- name: GetCAPAAction :one
SELECT * FROM capa_actions
WHERE id = $1 AND org_id = $2;

-- name: ListCAPAActions :many
SELECT * FROM capa_actions
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOpenCAPAActions :many
SELECT * FROM capa_actions
WHERE org_id = $1 AND status = ANY($2::text[]) AND due_date <= $3
ORDER BY due_date ASC
LIMIT $4 OFFSET $5;

-- name: UpdateCAPAStatus :exec
UPDATE capa_actions SET
    status = $1, updated_by = $2, updated_at = now()
WHERE id = $3 AND org_id = $4;

-- name: VerifyCAPAEffectiveness :exec
UPDATE capa_actions SET
    effectiveness_verified = true,
    effectiveness_verified_by = $1,
    verified_at = now(),
    status = 'closed',
    closed_by = $1,
    closed_at = now(),
    updated_by = $1,
    updated_at = now()
WHERE id = $2 AND org_id = $3;

-- name: CreateTrainingRecord :one
INSERT INTO training_records (
    org_id, user_id, training_title, training_type, topic, training_date,
    method, trainer, assessed, assessment_result, competency_valid_until,
    certificate_url, created_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: ListTrainingRecords :many
SELECT * FROM training_records
WHERE org_id = $1
ORDER BY training_date DESC
LIMIT $2 OFFSET $3;

-- name: ListTrainingRecordsByUser :many
SELECT * FROM training_records
WHERE org_id = $1 AND user_id = $2
ORDER BY training_date DESC
LIMIT $3 OFFSET $4;