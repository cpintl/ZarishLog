package handler

import (
	"fmt"

	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func ListProducts(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		params := pagination.FromQuery(c)

		var total int
		countQuery := `SELECT COUNT(*) FROM products WHERE org_id = $1`
		if err := db.Get(&total, countQuery, orgID); err != nil {
			response.InternalError(c, "failed to count products")
			return
		}

		var products []model.Product
		query := `SELECT * FROM products WHERE org_id = $1 ORDER BY name LIMIT $2 OFFSET $3`
		err := db.Select(&products, query, orgID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch products")
			return
		}

		response.Paginated(c, products, total, params.Page, params.PageSize)
	}
}

func GetProduct(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		orgID := c.GetString("org_id")

		var product model.Product
		query := `SELECT * FROM products WHERE id = $1 AND org_id = $2`
		err := db.Get(&product, query, id, orgID)
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}

		response.OK(c, product)
	}
}

func CreateProduct(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var product model.Product
		if errs := validator.BindAndValidate(c, &product); errs != nil {
			response.Validation(c, errs)
			return
		}

		product.OrgID = c.GetString("org_id")
		product.CreatedBy = c.GetString("user_id")
		product.UpdatedBy = c.GetString("user_id")

		query := fmt.Sprintf(`
			INSERT INTO products (org_id, category_id, uom_id, sku, name, description, item_type,
				gtin, alternative_code, brand, manufacturer, is_batch_tracked, is_serial_tracked,
				is_expiry_tracked, is_hazardous, is_cold_chain, min_stock, max_stock, reorder_point,
				lead_time_days, unit_cost, safety_stock, status, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
				$17, $18, $19, $20, $21, $22, $23, $24, $25)
			RETURNING id, created_at, updated_at`)

		err := db.QueryRow(query,
			product.OrgID, product.CategoryID, product.UomID, product.SKU, product.Name,
			product.Description, product.ItemType, product.GTIN, product.AlternativeCode,
			product.Brand, product.Manufacturer, product.IsBatchTracked, product.IsSerialTracked,
			product.IsExpiryTracked, product.IsHazardous, product.IsColdChain,
			product.MinStock, product.MaxStock, product.ReorderPoint, product.LeadTimeDays,
			product.UnitCost, product.SafetyStock, product.Status, product.CreatedBy, product.UpdatedBy,
		).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

		if err != nil {
			response.InternalError(c, fmt.Sprintf("failed to create product: %v", err))
			return
		}

		response.Created(c, product)
	}
}

func UpdateProduct(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var product model.Product
		if errs := validator.BindAndValidate(c, &product); errs != nil {
			response.Validation(c, errs)
			return
		}

		product.UpdatedBy = c.GetString("user_id")

		query := fmt.Sprintf(`
			UPDATE products SET
				name = $1, description = $2, category_id = $3, uom_id = $4,
				gtin = $5, brand = $6, manufacturer = $7, is_batch_tracked = $8,
				is_serial_tracked = $9, is_expiry_tracked = $10, is_hazardous = $11,
				is_cold_chain = $12, min_stock = $13, max_stock = $14, reorder_point = $15,
				lead_time_days = $16, unit_cost = $17, safety_stock = $18, status = $19,
				updated_by = $20
			WHERE id = $21 AND org_id = $22`)

		_, err := db.Exec(query,
			product.Name, product.Description, product.CategoryID, product.UomID,
			product.GTIN, product.Brand, product.Manufacturer, product.IsBatchTracked,
			product.IsSerialTracked, product.IsExpiryTracked, product.IsHazardous,
			product.IsColdChain, product.MinStock, product.MaxStock, product.ReorderPoint,
			product.LeadTimeDays, product.UnitCost, product.SafetyStock, product.Status,
			product.UpdatedBy, id, c.GetString("org_id"),
		)
		if err != nil {
			response.InternalError(c, "failed to update product")
			return
		}

		response.OK(c, gin.H{"message": "product updated"})
	}
}

func DeleteProduct(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		query := `UPDATE products SET status = 'inactive', updated_by = $1 WHERE id = $2 AND org_id = $3`
		_, err := db.Exec(query, c.GetString("user_id"), id, c.GetString("org_id"))
		if err != nil {
			response.InternalError(c, "failed to delete product")
			return
		}

		response.OK(c, gin.H{"message": "product deleted"})
	}
}
