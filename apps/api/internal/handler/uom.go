package handler

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func ListUoMs(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := pagination.FromQuery(c)

		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM units_of_measure`); err != nil {
			response.InternalError(c, "failed to count units of measure")
			return
		}

		var uoms []model.UoM
		err := db.Select(&uoms, `SELECT * FROM units_of_measure ORDER BY name LIMIT $1 OFFSET $2`, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch units of measure")
			return
		}

		response.Paginated(c, uoms, total, params.Page, params.PageSize)
	}
}

func GetUoM(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var uom model.UoM
		if err := db.Get(&uom, `SELECT * FROM units_of_measure WHERE id = $1`, id); err != nil {
			response.NotFound(c, "unit of measure not found")
			return
		}
		response.OK(c, uom)
	}
}

func CreateUoM(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var uom model.UoM
		if errs := validator.BindAndValidate(c, &uom); errs != nil {
			response.Validation(c, errs)
			return
		}

		if uom.Status == "" {
			uom.Status = "active"
		}

		err := db.QueryRow(`
			INSERT INTO units_of_measure (name, abbreviation, category, base_uom_id, conversion_factor, status)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, updated_at`,
			uom.Name, uom.Abbreviation, uom.Category,
			uom.BaseUomID, uom.ConversionFactor, uom.Status,
		).Scan(&uom.ID, &uom.CreatedAt, &uom.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create unit of measure")
			return
		}
		response.Created(c, uom)
	}
}

func UpdateUoM(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var uom model.UoM
		if errs := validator.BindAndValidate(c, &uom); errs != nil {
			response.Validation(c, errs)
			return
		}

		if uom.Status == "" {
			uom.Status = "active"
		}

		_, err := db.Exec(`
			UPDATE units_of_measure SET
				name = $1, abbreviation = $2, category = $3,
				base_uom_id = $4, conversion_factor = $5, status = $6
			WHERE id = $7`,
			uom.Name, uom.Abbreviation, uom.Category,
			uom.BaseUomID, uom.ConversionFactor, uom.Status, id,
		)
		if err != nil {
			response.InternalError(c, "failed to update unit of measure")
			return
		}
		response.OK(c, gin.H{"message": "unit of measure updated"})
	}
}

func DeleteUoM(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec(`DELETE FROM units_of_measure WHERE id = $1`, id)
		if err != nil {
			response.InternalError(c, "failed to delete unit of measure")
			return
		}
		response.OK(c, gin.H{"message": "unit of measure deleted"})
	}
}
