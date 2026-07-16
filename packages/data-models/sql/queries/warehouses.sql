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

-- name: CreateLocation :one
INSERT INTO locations (warehouse_id, parent_id, code, name, type, is_cold_chain, is_hazardous, is_secure, max_capacity, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateLocation :exec
UPDATE locations SET
    code = $1, name = $2, type = $3::location_type,
    is_cold_chain = $4, is_hazardous = $5, is_secure = $6,
    max_capacity = $7, is_active = $8
WHERE id = $9;

-- name: GetWarehouseWithLocations :many
SELECT w.*, l.id as location_id, l.code as location_code, l.name as location_name,
       l.type as location_type, l.is_cold_chain, l.is_hazardous, l.is_secure,
       l.max_capacity, l.is_active as location_active
FROM warehouses w
LEFT JOIN locations l ON l.warehouse_id = w.id
WHERE w.id = $1 AND w.org_id = $2
ORDER BY l.code;

-- name: ListWarehousesByType :many
SELECT * FROM warehouses
WHERE org_id = $1 AND type = $2::warehouse_type
ORDER BY name;
