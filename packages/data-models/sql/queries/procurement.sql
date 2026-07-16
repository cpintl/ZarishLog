-- name: ListSuppliers :many
SELECT * FROM suppliers
WHERE org_id = $1
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: GetSupplier :one
SELECT * FROM suppliers
WHERE id = $1 AND org_id = $2;

-- name: CreateSupplier :one
INSERT INTO suppliers (org_id, code, name, type, contact_person, email, phone, address, city, country, tax_id, payment_terms, is_active, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: UpdateSupplier :exec
UPDATE suppliers SET
    code = $1, name = $2, type = $3, contact_person = $4,
    email = $5, phone = $6, address = $7, city = $8, country = $9,
    tax_id = $10, payment_terms = $11, is_active = $12,
    status = $13::entity_status, updated_by = $14
WHERE id = $15 AND org_id = $16;

-- name: ListPurchaseOrders :many
SELECT po.*, s.name as supplier_name
FROM purchase_orders po
LEFT JOIN suppliers s ON po.supplier_id = s.id
WHERE po.org_id = $1
ORDER BY po.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetPurchaseOrder :one
SELECT po.*, s.name as supplier_name
FROM purchase_orders po
LEFT JOIN suppliers s ON po.supplier_id = s.id
WHERE po.id = $1 AND po.org_id = $2;

-- name: CreatePurchaseOrder :one
INSERT INTO purchase_orders (org_id, po_number, supplier_id, warehouse_id, program_id, department_id, order_date, expected_delivery_date, delivery_address, currency, subtotal, tax_amount, total_amount, status, approval_status, notes, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
RETURNING *;

-- name: UpdatePurchaseOrderStatus :exec
UPDATE purchase_orders SET
    status = $1::entity_status,
    approval_status = $2,
    approved_by = $3,
    approved_at = CASE WHEN $2 = 'approved' THEN now() ELSE approved_at END,
    updated_by = $4
WHERE id = $5 AND org_id = $6;

-- name: AddPOLineItem :one
INSERT INTO po_line_items (po_id, product_id, line_number, quantity_ordered, quantity_received, unit_price, line_total, discount_percent, tax_percent, scheduled_date, status, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: ListPOLineItems :many
SELECT pli.*, p.name as product_name, p.sku as product_sku
FROM po_line_items pli
JOIN products p ON pli.product_id = p.id
WHERE pli.po_id = $1
ORDER BY pli.line_number;
