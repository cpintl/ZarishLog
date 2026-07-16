-- name: ListCategories :many
SELECT * FROM product_categories
WHERE org_id = $1
ORDER BY name;

-- name: CreateCategory :one
INSERT INTO product_categories (org_id, parent_id, name, description, unspsc, eclass, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;
