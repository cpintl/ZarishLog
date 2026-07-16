package handler

import (
	"net/http"

	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func ListCategories(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var categories []model.ProductCategory
		query := `SELECT * FROM product_categories WHERE org_id = $1 ORDER BY name`
		err := db.Select(&categories, query, c.GetString("org_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": categories})
	}
}

func CreateCategory(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cat model.ProductCategory
		if err := c.ShouldBindJSON(&cat); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": cat})
	}
}
