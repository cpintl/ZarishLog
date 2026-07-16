package handler

import (
	"net/http"

	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func CreateGRN(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var grn model.GRN
		if err := c.ShouldBindJSON(&grn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create GRN: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": grn})
	}
}

func CreateIssue(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var issue model.StockIssue
		if err := c.ShouldBindJSON(&issue); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create issue"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": issue})
	}
}

func CreateTransfer(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
	}
}

func CreateAdjustment(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
	}
}

func GetStockLevels(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var levels []model.StockLevel
		query := `SELECT * FROM stock_levels WHERE org_id = $1 ORDER BY product_id`
		err := db.Select(&levels, query, c.GetString("org_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stock levels"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": levels})
	}
}

func GetStockMovements(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var movements []model.StockMovement
		query := `SELECT * FROM stock_movements WHERE org_id = $1 ORDER BY created_at DESC LIMIT 100`
		err := db.Select(&movements, query, c.GetString("org_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stock movements"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": movements})
	}
}
