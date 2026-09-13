-- Policy §17 — Complaints, recalls (incl. mock recall) and falsified-product triage

-- name: CreateComplaint :one
INSERT INTO complaints (
    org_id, complaint_number, received_date, received_via, severity, category,
    product_id, batch_id, description, reporter_name, reporter_contact,
    immediate_escalation, escalated, escalated_to, manufacturer_notified,
    competent_authority_notified, status, resolution_notes, capa_id, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
RETURNING *;

-- name: GetComplaint :one
SELECT * FROM complaints
WHERE id = $1 AND org_id = $2;

-- name: ListComplaints :many
SELECT * FROM complaints
WHERE org_id = $1
ORDER BY received_date DESC
LIMIT $2 OFFSET $3;

-- name: ListEscalatedComplaints :many
SELECT * FROM complaints
WHERE org_id = $1 AND (immediate_escalation = true OR escalated = true)
ORDER BY received_date DESC
LIMIT $2 OFFSET $3;

-- name: CreateRecall :one
INSERT INTO recalls (
    org_id, recall_number, recall_type, recall_date, product_id, subject_product,
    batch_numbers, reason, initiated_by, authorized_by, communications,
    disposition_plan, status, closure_evidence, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: GetRecall :one
SELECT * FROM recalls
WHERE id = $1 AND org_id = $2;

-- name: ListRecalls :many
SELECT * FROM recalls
WHERE org_id = $1
ORDER BY recall_date DESC
LIMIT $2 OFFSET $3;

-- name: ListRecallsByType :many
SELECT * FROM recalls
WHERE org_id = $1 AND recall_type = $2
ORDER BY recall_date DESC
LIMIT $3 OFFSET $4;

-- name: GetRecallLineItems :many
SELECT * FROM recall_line_items
WHERE recall_id = $1
ORDER BY recovered_at DESC;

-- name: CreateRecallLineItem :one
INSERT INTO recall_line_items (
    recall_id, recipient, location, product_id, batch_number, quantity_issued,
    quantity_recovered, quantity_outstanding, storage_disposition, recovered_at, notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;