-- Policy §12 / §14 — Quality release and controlled-product registers

-- name: CreateStockReleaseRecord :one
INSERT INTO stock_release_records (
    org_id, release_number, warehouse_id, product_id, batch_id, location_id,
    quantity, quarantine_reference, evidence_reviewed, decision, decision_date,
    decided_by, notes, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetStockReleaseRecord :one
SELECT * FROM stock_release_records
WHERE id = $1 AND org_id = $2;

-- name: ListStockReleaseRecords :many
SELECT * FROM stock_release_records
WHERE org_id = $1
ORDER BY decision_date DESC
LIMIT $2 OFFSET $3;

-- name: CreateControlledStockRegisterEntry :one
INSERT INTO controlled_stock_register (
    org_id, product_id, batch_id, warehouse_id, register_date, transaction_type,
    reference_document, quantity, balance_after, received_from, issued_to,
    authorized_recipient, signature, discrepancy, reconciled_at, reconciled_by, created_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: GetControlledStockRegisterEntry :one
SELECT * FROM controlled_stock_register
WHERE id = $1 AND org_id = $2;

-- name: ListControlledStockRegister :many
SELECT * FROM controlled_stock_register
WHERE org_id = $1
ORDER BY register_date DESC
LIMIT $2 OFFSET $3;

-- name: ListControlledStockRegisterByProduct :many
SELECT * FROM controlled_stock_register
WHERE org_id = $1 AND product_id = $2
ORDER BY register_date DESC
LIMIT $3 OFFSET $4;

-- name: CreateShortExpiryReview :one
INSERT INTO short_expiry_reviews (
    org_id, review_date, product_id, batch_id, warehouse_id, expiry_date,
    remaining_shelf_life_months, expected_consumption, transfer_option,
    donor_condition, regulatory_restriction, decision, decision_by,
    decision_date, notes, created_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: ListShortExpiryReviews :many
SELECT * FROM short_expiry_reviews
WHERE org_id = $1
ORDER BY expiry_date ASC
LIMIT $2 OFFSET $3;

-- name: ListOpenShortExpiryReviews :many
SELECT * FROM short_expiry_reviews
WHERE org_id = $1 AND decision = 'pending'
ORDER BY expiry_date ASC
LIMIT $2 OFFSET $3;