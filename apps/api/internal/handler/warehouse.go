package handler

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func NotImplementedStub(_ *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.NotImplemented(c, "this endpoint is not implemented yet")
	}
}

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
		query := `SELECT * FROM warehouses WHERE org_id = $1 ORDER BY name LIMIT $2 OFFSET $3`
		err := db.Select(&warehouses, query, orgID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch warehouses")
			return
		}

		response.Paginated(c, warehouses, total, params.Page, params.PageSize)
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

		query := `
			INSERT INTO warehouses (org_id, name, code, type, address, city, country, is_active, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, created_at, updated_at`

		err := db.QueryRow(query, wh.OrgID, wh.Name, wh.Code, wh.Type, wh.Address,
			wh.City, wh.Country, wh.IsActive, wh.CreatedBy, wh.UpdatedBy,
		).Scan(&wh.ID, &wh.CreatedAt, &wh.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create warehouse")
			return
		}
		response.Created(c, wh)
	}
}
