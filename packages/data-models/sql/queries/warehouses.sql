-- name: ListWarehouses :many
SELECT * FROM warehouses
WHERE org_id = $1 AND is_active = true
ORDER BY name;

-- name: GetWarehouse :one
SELECT * FROM warehouses
WHERE id = $1 AND org_id = $2;

-- name: CreateWarehouse :one
INSERT INTO warehouses (org_id, name, code, type, address, city, country, is_active, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListLocations :many
SELECT * FROM locations
WHERE warehouse_id = $1
ORDER BY code;
