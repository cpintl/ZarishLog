package handler

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func ListWarehouses(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		params := pagination.FromQuery(c)

		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM warehouses WHERE org_id = $1`, orgID); err != nil {
			response.InternalError(c, "failed to count warehouses")
			return
		}

		var warehouses []model.Warehouse
		err := db.Select(&warehouses, `SELECT * FROM warehouses WHERE org_id = $1 ORDER BY name LIMIT $2 OFFSET $3`, orgID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch warehouses")
			return
		}

		response.Paginated(c, warehouses, total, params.Page, params.PageSize)
	}
}

func GetWarehouse(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		id := c.Param("id")
		var wh model.Warehouse
		err := db.Get(&wh, `SELECT * FROM warehouses WHERE id = $1 AND org_id = $2`, id, orgID)
		if err != nil {
			response.NotFound(c, "warehouse not found")
			return
		}
		response.OK(c, wh)
	}
}

func CreateWarehouse(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var wh model.Warehouse
		if errs := validator.BindAndValidate(c, &wh); errs != nil {
			response.Validation(c, errs)
			return
		}

		wh.OrgID = c.GetString("org_id")
		wh.CreatedBy = c.GetString("user_id")
		wh.UpdatedBy = c.GetString("user_id")

		err := db.QueryRow(`
			INSERT INTO warehouses (org_id, name, code, type, address, city, country, is_active, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, created_at, updated_at`,
			wh.OrgID, wh.Name, wh.Code, wh.Type, wh.Address, wh.City, wh.Country, wh.IsActive, wh.CreatedBy, wh.UpdatedBy,
		).Scan(&wh.ID, &wh.CreatedAt, &wh.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create warehouse")
			return
		}
		response.Created(c, wh)
	}
}

func UpdateWarehouse(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		id := c.Param("id")

		var wh model.Warehouse
		if errs := validator.BindAndValidate(c, &wh); errs != nil {
			response.Validation(c, errs)
			return
		}

		wh.UpdatedBy = c.GetString("user_id")

		_, err := db.Exec(`
			UPDATE warehouses SET name=$1, code=$2, type=$3, address=$4, city=$5, country=$6, is_active=$7, updated_by=$8
			WHERE id=$9 AND org_id=$10`,
			wh.Name, wh.Code, wh.Type, wh.Address, wh.City, wh.Country, wh.IsActive, wh.UpdatedBy, id, orgID,
		)
		if err != nil {
			response.InternalError(c, "failed to update warehouse")
			return
		}
		response.OK(c, gin.H{"message": "warehouse updated"})
	}
}

func DeleteWarehouse(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		id := c.Param("id")
		_, err := db.Exec(`DELETE FROM warehouses WHERE id = $1 AND org_id = $2`, id, orgID)
		if err != nil {
			response.InternalError(c, "failed to delete warehouse")
			return
		}
		response.OK(c, gin.H{"message": "warehouse deleted"})
	}
}
