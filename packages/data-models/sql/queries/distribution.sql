-- name: ListDistributions :many
SELECT d.*, ol.name as org_level_name, p.name as program_name
FROM distributions d
LEFT JOIN org_levels ol ON d.org_level_id = ol.id
LEFT JOIN programs p ON d.program_id = p.id
WHERE d.org_id = $1
ORDER BY d.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetDistribution :one
SELECT d.*, ol.name as org_level_name, p.name as program_name
FROM distributions d
LEFT JOIN org_levels ol ON d.org_level_id = ol.id
LEFT JOIN programs p ON d.program_id = p.id
WHERE d.id = $1 AND d.org_id = $2;

-- name: CreateDistribution :one
INSERT INTO distributions (org_id, org_level_id, program_id, distribution_number, distribution_date, location, beneficiary_count, status, notes, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateDistributionStatus :exec
UPDATE distributions SET
    status = $1::entity_status, updated_by = $2
WHERE id = $3 AND org_id = $4;

-- name: ListDistributionLineItems :many
SELECT dli.*, p.name as product_name, p.sku as product_sku
FROM distribution_line_items dli
JOIN products p ON dli.product_id = p.id
WHERE dli.distribution_id = $1
ORDER BY dli.id;

-- name: AddDistributionLineItem :one
INSERT INTO distribution_line_items (distribution_id, product_id, batch_id, quantity_planned, quantity_distributed, unit_cost, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
