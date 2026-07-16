-- name: CreateSyncLog :one
INSERT INTO sync_log (org_id, device_id, user_id, sync_type, started_at, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateSyncLog :exec
UPDATE sync_log SET
    completed_at = $1, status = $2,
    records_pushed = $3, records_pulled = $4,
    errors_count = $5, error_message = $6
WHERE id = $7;

-- name: CreateSyncConflict :one
INSERT INTO sync_conflicts (org_id, table_name, record_id, device_id, server_version, client_version, conflict_type, resolution)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ResolveSyncConflict :exec
UPDATE sync_conflicts SET
    resolution = $1, resolved_by = $2, resolved_at = now()
WHERE id = $3;

-- name: GetPendingConflicts :many
SELECT * FROM sync_conflicts
WHERE org_id = $1 AND resolution = 'pending'
ORDER BY created_at ASC;
