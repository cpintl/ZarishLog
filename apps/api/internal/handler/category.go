package handler

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func ListCategories(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		params := pagination.FromQuery(c)

		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM product_categories WHERE org_id = $1`, orgID); err != nil {
			response.InternalError(c, "failed to count categories")
			return
		}

		var categories []model.ProductCategory
		query := `SELECT * FROM product_categories WHERE org_id = $1 ORDER BY name LIMIT $2 OFFSET $3`
		err := db.Select(&categories, query, orgID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch categories")
			return
		}

		response.Paginated(c, categories, total, params.Page, params.PageSize)
	}
}

func CreateCategory(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cat model.ProductCategory
		if errs := validator.BindAndValidate(c, &cat); errs != nil {
			response.Validation(c, errs)
			return
		}

		cat.OrgID = c.GetString("org_id")
		cat.CreatedBy = c.GetString("user_id")
		cat.UpdatedBy = c.GetString("user_id")

		query := `
			INSERT INTO product_categories (org_id, parent_id, name, description, unspsc, eclass, status, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, created_at, updated_at`

		err := db.QueryRow(query, cat.OrgID, cat.ParentID, cat.Name, cat.Description,
			cat.UNSPSC, cat.ECLASS, cat.Status, cat.CreatedBy, cat.UpdatedBy,
		).Scan(&cat.ID, &cat.CreatedAt, &cat.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create category")
			return
		}
		response.Created(c, cat)
	}
}
