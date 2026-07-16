-- name: ListAlerts :many
SELECT a.*, p.name as product_name, w.name as warehouse_name
FROM alerts a
LEFT JOIN products p ON a.product_id = p.id
LEFT JOIN warehouses w ON a.warehouse_id = w.id
WHERE a.org_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetActiveAlerts :many
SELECT a.*, p.name as product_name, w.name as warehouse_name
FROM alerts a
LEFT JOIN products p ON a.product_id = p.id
LEFT JOIN warehouses w ON a.warehouse_id = w.id
WHERE a.org_id = $1 AND a.resolved_at IS NULL
ORDER BY a.created_at DESC;

-- name: CreateAlert :one
INSERT INTO alerts (org_id, alert_config_id, alert_type, severity, title, message, product_id, warehouse_id, batch_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: AcknowledgeAlert :exec
UPDATE alerts SET
    is_acknowledged = true, acknowledged_by = $1, acknowledged_at = now()
WHERE id = $2 AND org_id = $3;

-- name: ResolveAlert :exec
UPDATE alerts SET
    resolved_at = now()
WHERE id = $1 AND org_id = $2;

-- name: ListAlertConfigurations :many
SELECT * FROM alert_configurations
WHERE org_id = $1
ORDER BY name;

-- name: CreateAlertConfiguration :one
INSERT INTO alert_configurations (org_id, alert_type, name, description, threshold_type, threshold_value, enabled, notification_channels, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;
