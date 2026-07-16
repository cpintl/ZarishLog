package handler

import (
	"net/http"

	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func ListWarehouses(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var warehouses []model.Warehouse
		query := `SELECT * FROM warehouses WHERE org_id = $1 ORDER BY name`
		err := db.Select(&warehouses, query, c.GetString("org_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch warehouses"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": warehouses})
	}
}

func CreateWarehouse(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var wh model.Warehouse
		if err := c.ShouldBindJSON(&wh); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		wh.OrgID = c.GetString("org_id")
		wh.CreatedBy = c.GetString("user_id")
		wh.UpdatedBy = c.GetString("user_id")

		query := `
			INSERT INTO warehouses (org_id, name, code, type, address, city, country, is_active, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, created_at, updated_at`

		err := db.QueryRow(query, wh.OrgID, wh.Name, wh.Code, wh.Type, wh.Address,
			wh.City, wh.Country, wh.IsActive, wh.CreatedBy, wh.UpdatedBy,
		).Scan(&wh.ID, &wh.CreatedAt, &wh.UpdatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create warehouse"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": wh})
	}
}
