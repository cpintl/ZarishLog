-- Policy §13 — Cold Chain & Environmental Control

-- name: CreateTemperatureMonitoringEntry :one
INSERT INTO temperature_monitoring_entries (
    org_id, warehouse_id, location_id, equipment_id, device_name,
    monitor_type, reading_value, min_threshold, max_threshold,
    recorded_at, recorded_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: ListTemperatureMonitoringEntries :many
SELECT * FROM temperature_monitoring_entries
WHERE org_id = $1
ORDER BY recorded_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTemperatureMonitoringEntriesByWarehouse :many
SELECT * FROM temperature_monitoring_entries
WHERE org_id = $1 AND warehouse_id = $2 AND monitor_type = $3
  AND recorded_at >= $4 AND recorded_at <= $5
ORDER BY recorded_at ASC
LIMIT $6 OFFSET $7;

-- name: CreateTemperatureExcursion :one
INSERT INTO temperature_excursions (
    org_id, warehouse_id, location_id, equipment_id, device_name,
    excursion_type, started_at, ended_at, min_value, max_value,
    expected_min, expected_max, affected_products, affected_batches,
    quantity_affected, quarantined, quarantine_reference, disposition,
    notified_quality, notified_manager, technical_advice, root_cause,
    capa_id, status, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
RETURNING *;

-- name: GetTemperatureExcursion :one
SELECT * FROM temperature_excursions
WHERE id = $1 AND org_id = $2;

-- name: ListTemperatureExcursions :many
SELECT * FROM temperature_excursions
WHERE org_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOpenTemperatureExcursions :many
SELECT * FROM temperature_excursions
WHERE org_id = $1 AND status <> 'closed'
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateTemperatureExcursionDisposition :exec
UPDATE temperature_excursions SET
    disposition = $1, status = $2, technical_advice = $3,
    root_cause = $4, capa_id = $5, ended_at = $6,
    updated_by = $7, updated_at = now()
WHERE id = $8 AND org_id = $9;