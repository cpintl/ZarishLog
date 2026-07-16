-- name: GetStockLevels :many
SELECT * FROM stock_levels
WHERE org_id = $1
ORDER BY product_id;

-- name: GetStockLevel :one
SELECT * FROM stock_levels
WHERE product_id = $1 AND warehouse_id = $2 AND location_id = $3 AND batch_id = $4;

-- name: UpdateStockLevel :exec
UPDATE stock_levels SET
    quantity = $1, reserved_qty = $2
WHERE product_id = $3 AND warehouse_id = $4 AND location_id = $5 AND batch_id = $6;

-- name: GetStockMovements :many
SELECT * FROM stock_movements
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: CreateGRN :one
INSERT INTO goods_receipts (org_id, warehouse_id, grn_number, supplier, po_number, received_by, status, notes, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CreateGRNLineItem :exec
INSERT INTO grn_line_items (grn_id, product_id, batch_number, serial_number, expiry_date, quantity, unit_cost, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: CreateStockMovement :exec
INSERT INTO stock_movements (org_id, product_id, warehouse_id, location_id, batch_id, movement_type, quantity, ref_doc_type, ref_doc_id, reason_code, reference, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: CreateBatch :one
INSERT INTO batches (org_id, product_id, batch_number, serial_number, expiry_date, manufacturer)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetBatchesByProduct :many
SELECT * FROM batches
WHERE product_id = $1 AND org_id = $2
ORDER BY created_at DESC;

-- name: GetExpiringBatches :many
SELECT * FROM batches
WHERE expiry_date BETWEEN $1 AND $2 AND org_id = $3
ORDER BY expiry_date;

-- name: SearchStockMovements :many
SELECT * FROM stock_movements
WHERE org_id = $1
  AND ($2 = '' OR movement_type = $2::movement_type)
  AND ($3::date IS NULL OR created_at >= $3::date)
  AND ($4::date IS NULL OR created_at <= $4::date + interval '1 day')
  AND ($5 = '' OR product_id = $5::uuid)
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;
