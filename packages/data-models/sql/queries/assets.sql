-- name: ListAssets :many
SELECT a.*, u.name as custodian_name, l.code as location_code
FROM assets a
LEFT JOIN users u ON a.custodian_id = u.id
LEFT JOIN locations l ON a.location_id = l.id
WHERE a.org_id = $1
ORDER BY a.name
LIMIT $2 OFFSET $3;

-- name: GetAsset :one
SELECT a.*, u.name as custodian_name, l.code as location_code
FROM assets a
LEFT JOIN users u ON a.custodian_id = u.id
LEFT JOIN locations l ON a.location_id = l.id
WHERE a.id = $1 AND a.org_id = $2;

-- name: CreateAsset :one
INSERT INTO assets (org_id, asset_tag, name, description, product_id, serial_number, custodian_id, location_id, acquisition_date, purchase_cost, current_value, depreciation_method, useful_life_years, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: UpdateAsset :exec
UPDATE assets SET
    name = $1, description = $2, custodian_id = $3, location_id = $4,
    current_value = $5, status = $6, updated_by = $7
WHERE id = $8 AND org_id = $9;

-- name: ListAssetCustodyChanges :many
SELECT acc.*, fu.name as from_user_name, tu.name as to_user_name
FROM asset_custody_changes acc
LEFT JOIN users fu ON acc.from_user_id = fu.id
LEFT JOIN users tu ON acc.to_user_id = tu.id
WHERE acc.asset_id = $1
ORDER BY acc.changed_at DESC;

-- name: AddCustodyChange :one
INSERT INTO asset_custody_changes (asset_id, from_user_id, to_user_id, changed_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListAssetMaintenance :many
SELECT * FROM asset_maintenance
WHERE asset_id = $1
ORDER BY maintenance_date DESC;

-- name: AddMaintenanceRecord :one
INSERT INTO asset_maintenance (asset_id, maintenance_date, description, cost, performed_by, next_date)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListAssetDepreciation :many
SELECT * FROM asset_depreciation_schedule
WHERE asset_id = $1
ORDER BY period_date DESC;
