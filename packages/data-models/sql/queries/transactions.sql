-- name: ListGRNs :many
SELECT g.*, w.name as warehouse_name
FROM goods_receipts g
LEFT JOIN warehouses w ON g.warehouse_id = w.id
WHERE g.org_id = $1
ORDER BY g.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetGRN :one
SELECT g.*, w.name as warehouse_name
FROM goods_receipts g
LEFT JOIN warehouses w ON g.warehouse_id = w.id
WHERE g.id = $1 AND g.org_id = $2;

-- name: ListGRNLineItems :many
SELECT gli.*, p.name as product_name, p.sku as product_sku
FROM grn_line_items gli
JOIN products p ON gli.product_id = p.id
WHERE gli.grn_id = $1
ORDER BY gli.id;

-- name: ListStockIssues :many
SELECT si.*, w.name as warehouse_name
FROM stock_issues si
LEFT JOIN warehouses w ON si.warehouse_id = w.id
WHERE si.org_id = $1
ORDER BY si.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetStockIssue :one
SELECT si.*, w.name as warehouse_name
FROM stock_issues si
LEFT JOIN warehouses w ON si.warehouse_id = w.id
WHERE si.id = $1 AND si.org_id = $2;

-- name: ListIssueLineItems :many
SELECT ili.*, p.name as product_name, p.sku as product_sku
FROM issue_line_items ili
JOIN products p ON ili.product_id = p.id
WHERE ili.issue_id = $1
ORDER BY ili.id;

-- name: CreateStockIssue :one
INSERT INTO stock_issues (org_id, warehouse_id, issue_number, requested_by, approved_by, program_id, department_id, status, notes, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: CreateIssueLineItem :exec
INSERT INTO issue_line_items (issue_id, product_id, batch_id, quantity, unit_cost)
VALUES ($1, $2, $3, $4, $5);

-- name: ListTransfers :many
SELECT t.*, fw.name as from_warehouse_name, tw.name as to_warehouse_name
FROM stock_transfers t
LEFT JOIN warehouses fw ON t.from_warehouse_id = fw.id
LEFT JOIN warehouses tw ON t.to_warehouse_id = tw.id
WHERE t.org_id = $1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTransfer :one
SELECT t.*, fw.name as from_warehouse_name, tw.name as to_warehouse_name
FROM stock_transfers t
LEFT JOIN warehouses fw ON t.from_warehouse_id = fw.id
LEFT JOIN warehouses tw ON t.to_warehouse_id = tw.id
WHERE t.id = $1 AND t.org_id = $2;

-- name: CreateTransfer :one
INSERT INTO stock_transfers (org_id, from_warehouse_id, to_warehouse_id, transfer_number, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListTransferLineItems :many
SELECT tli.*, p.name as product_name, p.sku as product_sku
FROM transfer_line_items tli
JOIN products p ON tli.product_id = p.id
WHERE tli.transfer_id = $1
ORDER BY tli.id;

-- name: CreateTransferLineItem :exec
INSERT INTO transfer_line_items (transfer_id, product_id, batch_id, quantity, unit_cost, status)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListAdjustments :many
SELECT sa.*, w.name as warehouse_name
FROM stock_adjustments sa
LEFT JOIN warehouses w ON sa.warehouse_id = w.id
WHERE sa.org_id = $1
ORDER BY sa.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAdjustment :one
SELECT sa.*, w.name as warehouse_name
FROM stock_adjustments sa
LEFT JOIN warehouses w ON sa.warehouse_id = w.id
WHERE sa.id = $1 AND sa.org_id = $2;

-- name: CreateAdjustment :one
INSERT INTO stock_adjustments (org_id, warehouse_id, reason_code, description, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: CreateAdjustmentLineItem :exec
INSERT INTO adjustment_line_items (adjustment_id, product_id, batch_id, location_id, expected_quantity, actual_quantity, difference, reason_code, unit_cost, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: ListAdjustmentReasonCodes :many
SELECT * FROM adjustment_reason_codes
WHERE is_active = true
ORDER BY category, code;
