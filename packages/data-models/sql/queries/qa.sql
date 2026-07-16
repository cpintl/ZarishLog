-- name: ListQAInspections :many
SELECT q.*, p.name as product_name, p.sku as product_sku
FROM qa_inspections q
JOIN products p ON q.product_id = p.id
WHERE q.org_id = $1
ORDER BY q.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetQAInspection :one
SELECT q.*, p.name as product_name, p.sku as product_sku
FROM qa_inspections q
JOIN products p ON q.product_id = p.id
WHERE q.id = $1 AND q.org_id = $2;

-- name: CreateQAInspection :one
INSERT INTO qa_inspections (org_id, grn_id, product_id, batch_id, inspection_date, inspector, result, notes, disposition, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateQAInspectionResult :exec
UPDATE qa_inspections SET
    result = $1, notes = $2, disposition = $3, inspector = $4, updated_by = $5
WHERE id = $6 AND org_id = $7;

-- name: ListQAChecklistTemplates :many
SELECT * FROM qa_checklist_templates
WHERE org_id = $1
ORDER BY name;

-- name: GetQAChecklistTemplate :one
SELECT * FROM qa_checklist_templates
WHERE id = $1 AND org_id = $2;

-- name: CreateQAChecklistTemplate :one
INSERT INTO qa_checklist_templates (org_id, code, name, description, category, is_mandatory, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListQAChecklistItems :many
SELECT * FROM qa_checklist_items
WHERE template_id = $1
ORDER BY item_order;

-- name: CreateQAChecklistItem :one
INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
