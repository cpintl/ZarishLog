-- name: ListOrganizations :many
SELECT * FROM organizations
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: GetOrganization :one
SELECT * FROM organizations
WHERE id = $1;

-- name: CreateOrganization :one
INSERT INTO organizations (name, code, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateOrganization :exec
UPDATE organizations SET
    name = $1, status = $2::entity_status, updated_by = $3
WHERE id = $4;

-- name: ListOrgLevels :many
SELECT * FROM org_levels
WHERE org_id = $1
ORDER BY level, name;

-- name: CreateOrgLevel :one
INSERT INTO org_levels (org_id, parent_id, name, code, level, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetOrgTree :many
SELECT c.*, p.name as parent_name
FROM org_levels c
LEFT JOIN org_levels p ON c.parent_id = p.id
WHERE c.org_id = $1
ORDER BY c.level, c.name;

-- name: ListPrograms :many
SELECT * FROM programs
WHERE org_id = $1
ORDER BY name;

-- name: CreateProgram :one
INSERT INTO programs (org_id, code, name, description, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListDepartments :many
SELECT * FROM departments
WHERE org_id = $1
ORDER BY name;

-- name: CreateDepartment :one
INSERT INTO departments (org_id, name, code, parent_id, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
