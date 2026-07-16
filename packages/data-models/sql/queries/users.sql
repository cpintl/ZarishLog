-- name: ListUsers :many
SELECT u.*, r.name as role_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.org_id = $1
ORDER BY u.name
LIMIT $2 OFFSET $3;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 AND org_id = $2;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE org_id = $1 AND email = $2;

-- name: CreateUser :one
INSERT INTO users (org_id, email, name, role_id, org_level_id, is_active, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users SET
    email = $1, name = $2, role_id = $3, org_level_id = $4,
    is_active = $5, updated_by = $6
WHERE id = $7 AND org_id = $8;

-- name: ListRoles :many
SELECT * FROM roles
ORDER BY level, name;

-- name: CreateRole :one
INSERT INTO roles (code, name, description, level)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPermissions :many
SELECT * FROM permissions
ORDER BY module, action;

-- name: GetUserPermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN user_role_assignments ura ON rp.role_id = ura.role_id
WHERE ura.user_id = $1
ORDER BY p.module, p.action;
