-- name: ListProducts :many
SELECT * FROM products
WHERE org_id = $1 AND status = 'active'
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: GetProduct :one
SELECT * FROM products
WHERE id = $1 AND org_id = $2;

-- name: CreateProduct :one
INSERT INTO products (
    org_id, category_id, uom_id, sku, name, description, item_type,
    gtin, alternative_code, brand, manufacturer,
    is_batch_tracked, is_serial_tracked, is_expiry_tracked,
    is_hazardous, is_cold_chain,
    min_stock, max_stock, reorder_point, lead_time_days, unit_cost,
    status, created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14,
    $15, $16,
    $17, $18, $19, $20, $21,
    $22, $23, $24
) RETURNING *;

-- name: UpdateProduct :exec
UPDATE products SET
    name = $1, description = $2, category_id = $3, uom_id = $4,
    gtin = $5, brand = $6, manufacturer = $7,
    is_batch_tracked = $8, is_serial_tracked = $9, is_expiry_tracked = $10,
    is_hazardous = $11, is_cold_chain = $12,
    min_stock = $13, max_stock = $14, reorder_point = $15,
    lead_time_days = $16, unit_cost = $17,
    status = $18, updated_by = $19
WHERE id = $20 AND org_id = $21;

-- name: DeleteProduct :exec
UPDATE products SET status = 'inactive', updated_by = $1
WHERE id = $2 AND org_id = $3;

-- name: SearchProducts :many
SELECT * FROM products
WHERE org_id = $1
  AND (to_tsvector('simple', name || ' ' || sku || ' ' || COALESCE(description, '')) @@ plainto_tsquery('simple', $2) OR $2 = '')
ORDER BY name
LIMIT $3 OFFSET $4;
