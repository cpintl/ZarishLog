-- name: GetStockLevels :many
SELECT * FROM stock_levels
WHERE org_id = $1
ORDER BY product_id;

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
