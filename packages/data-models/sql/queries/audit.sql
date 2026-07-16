-- name: CreateAuditLog :exec
INSERT INTO audit_log (org_id, user_id, action, entity_type, entity_id, changes, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListAuditLogs :many
SELECT al.*, u.name as user_name
FROM audit_log al
LEFT JOIN users u ON al.user_id = u.id
WHERE al.org_id = $1
ORDER BY al.timestamp DESC
LIMIT $2 OFFSET $3;

-- name: SearchAuditLogs :many
SELECT al.*, u.name as user_name
FROM audit_log al
LEFT JOIN users u ON al.user_id = u.id
WHERE al.org_id = $1
  AND ($2 = '' OR al.action = $2)
  AND ($3 = '' OR al.entity_type = $3)
  AND ($4::date IS NULL OR al.timestamp >= $4::date)
  AND ($5::date IS NULL OR al.timestamp <= $5::date + interval '1 day')
ORDER BY al.timestamp DESC
LIMIT $6 OFFSET $7;

-- name: CreateDataChangeLog :exec
INSERT INTO data_change_log (org_id, table_name, record_id, operation, old_values, new_values, changed_by)
VALUES ($1, $2, $3, $4, $5, $6, $7);
