-- Policy §8 — Regulatory & Legal Compliance (Bangladesh DGDA/DNC matrix)

-- name: CreateRegulatoryApproval :one
INSERT INTO regulatory_approvals (
    org_id, scope_type, scope_id, authority, license_type, reference_number,
    product_scope, conditions, issue_date, expiry_date, renewal_lead_days,
    responsible_owner, status, evidence_url, notes, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: GetRegulatoryApproval :one
SELECT * FROM regulatory_approvals
WHERE id = $1 AND org_id = $2;

-- name: ListRegulatoryApprovals :many
SELECT * FROM regulatory_approvals
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateRegulatoryApprovalStatus :exec
UPDATE regulatory_approvals SET
    status = $1, updated_by = $2, updated_at = now()
WHERE id = $3 AND org_id = $4;

-- name: ListExpiringRegulatoryApprovals :many
SELECT * FROM regulatory_approvals
WHERE org_id = $1 AND expiry_date IS NOT NULL
  AND expiry_date <= $2
ORDER BY expiry_date ASC
LIMIT $3 OFFSET $4;