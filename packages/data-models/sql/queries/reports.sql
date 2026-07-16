-- name: ListReportDefinitions :many
SELECT * FROM report_definitions
WHERE org_id = $1
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: GetReportDefinition :one
SELECT * FROM report_definitions
WHERE id = $1 AND org_id = $2;

-- name: CreateReportDefinition :one
INSERT INTO report_definitions (org_id, code, name, description, category, sql_query, parameters, output_formats, is_system, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: ListReportSchedules :many
SELECT rs.*, rd.name as report_name
FROM report_schedules rs
JOIN report_definitions rd ON rs.report_id = rd.id
WHERE rd.org_id = $1
ORDER BY rs.next_run_at;

-- name: CreateReportSchedule :one
INSERT INTO report_schedules (report_id, schedule_cron, parameters, recipients, format, is_active, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateScheduleLastRun :exec
UPDATE report_schedules SET
    last_run_at = now(),
    next_run_at = $1,
    updated_by = $2
WHERE id = $3;
