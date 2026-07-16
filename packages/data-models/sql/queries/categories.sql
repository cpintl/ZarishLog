-- name: ListCategories :many
SELECT * FROM product_categories
WHERE org_id = $1
ORDER BY name;

-- name: GetCategory :one
SELECT * FROM product_categories
WHERE id = $1 AND org_id = $2;

-- name: CreateCategory :one
INSERT INTO product_categories (org_id, parent_id, name, description, unspsc, eclass, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateCategory :exec
UPDATE product_categories SET
    name = $1, description = $2, unspsc = $3, eclass = $4,
    status = $5, updated_by = $6
WHERE id = $7 AND org_id = $8;

-- name: DeleteCategory :exec
UPDATE product_categories SET status = 'inactive', updated_by = $1
WHERE id = $2 AND org_id = $3;

-- name: ListCategoryTree :many
SELECT c.*, p.name as parent_name
FROM product_categories c
LEFT JOIN product_categories p ON c.parent_id = p.id
WHERE c.org_id = $1
ORDER BY c.name;
