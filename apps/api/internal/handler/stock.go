package handler

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func CreateGRN(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var grn model.GRN
		if errs := validator.BindAndValidate(c, &grn); errs != nil {
			response.Validation(c, errs)
			return
		}

		grn.OrgID = c.GetString("org_id")
		grn.CreatedBy = c.GetString("user_id")
		grn.UpdatedBy = c.GetString("user_id")

		query := `
			INSERT INTO goods_receipts (org_id, warehouse_id, supplier, po_number, received_by, status, notes, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, created_at, updated_at`

		err := db.QueryRow(query, grn.OrgID, grn.WarehouseID, grn.Supplier, grn.PONumber,
			grn.ReceivedBy, grn.Status, grn.Notes, grn.CreatedBy, grn.UpdatedBy,
		).Scan(&grn.ID, &grn.CreatedAt, &grn.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create GRN")
			return
		}

		response.Created(c, grn)
	}
}

func CreateIssue(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var issue model.StockIssue
		if errs := validator.BindAndValidate(c, &issue); errs != nil {
			response.Validation(c, errs)
			return
		}

		issue.OrgID = c.GetString("org_id")

		query := `
			INSERT INTO stock_issues (org_id, warehouse_id, requested_by, approved_by, program_id, status)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at`

		err := db.QueryRow(query, issue.OrgID, issue.WarehouseID, issue.RequestedBy,
			issue.ApprovedBy, issue.ProgramID, issue.Status,
		).Scan(&issue.ID, &issue.CreatedAt)

		if err != nil {
			response.InternalError(c, "failed to create issue")
			return
		}

		response.Created(c, issue)
	}
}

func CreateTransfer(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.NotImplemented(c, "inter-warehouse transfer not implemented yet")
	}
}

func CreateAdjustment(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.NotImplemented(c, "stock adjustment not implemented yet")
	}
}

func GetStockLevels(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		params := pagination.FromQuery(c)

		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM stock_levels WHERE org_id = $1`, orgID); err != nil {
			response.InternalError(c, "failed to count stock levels")
			return
		}

		var levels []model.StockLevel
		query := `SELECT * FROM stock_levels WHERE org_id = $1 ORDER BY product_id LIMIT $2 OFFSET $3`
		err := db.Select(&levels, query, orgID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch stock levels")
			return
		}

		response.Paginated(c, levels, total, params.Page, params.PageSize)
	}
}

func GetStockMovements(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		params := pagination.FromQuery(c)

		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM stock_movements WHERE org_id = $1`, orgID); err != nil {
			response.InternalError(c, "failed to count stock movements")
			return
		}

		var movements []model.StockMovement
		query := `SELECT * FROM stock_movements WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		err := db.Select(&movements, query, orgID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch stock movements")
			return
		}

		response.Paginated(c, movements, total, params.Page, params.PageSize)
	}
}
