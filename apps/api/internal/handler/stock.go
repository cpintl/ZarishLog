package handler

import (
	"errors"

	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type InsufficientStockError struct {
	msg string
}

func (e *InsufficientStockError) Error() string { return e.msg }

// applyStockDelta adjusts the derived stock_levels row for a key atomically
// within the caller's transaction. The guarded UPDATE serializes concurrent
// adjustments on the same key and refuses negative balances.
func applyStockDelta(tx *sqlx.Tx, orgID, productID, warehouseID string, locationID, batchID interface{}, delta float64) error {
	target := `id = (SELECT id FROM stock_levels WHERE org_id=$2 AND product_id=$3 AND warehouse_id=$4
		AND location_id IS NOT DISTINCT FROM $5 AND batch_id IS NOT DISTINCT FROM $6
		ORDER BY updated_at DESC, id LIMIT 1)`
	res, err := tx.Exec(
		`UPDATE stock_levels SET quantity = quantity + $1, updated_at = now() WHERE `+target+` AND quantity + $1 >= 0`,
		delta, orgID, productID, warehouseID, locationID, batchID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 1 {
		return nil
	}

	var count int
	err = tx.Get(&count,
		`SELECT COUNT(*) FROM stock_levels WHERE org_id=$1 AND product_id=$2 AND warehouse_id=$3
		 AND location_id IS NOT DISTINCT FROM $4 AND batch_id IS NOT DISTINCT FROM $5`,
		orgID, productID, warehouseID, locationID, batchID,
	)
	if err != nil {
		return err
	}
	if count > 0 {
		return &InsufficientStockError{msg: "insufficient stock for movement"}
	}

	if _, err := tx.Exec(
		`INSERT INTO stock_levels (org_id, product_id, warehouse_id, location_id, batch_id, quantity, status)
		 VALUES ($1,$2,$3,$4,$5,0,'on_hand') ON CONFLICT DO NOTHING`,
		orgID, productID, warehouseID, locationID, batchID,
	); err != nil {
		return err
	}
	res, err = tx.Exec(
		`UPDATE stock_levels SET quantity = quantity + $1, updated_at = now() WHERE `+target+` AND quantity + $1 >= 0`,
		delta, orgID, productID, warehouseID, locationID, batchID,
	)
	if err != nil {
		return err
	}
	if n, _ = res.RowsAffected(); n == 1 {
		return nil
	}
	return &InsufficientStockError{msg: "insufficient stock for movement"}
}

// insertMovement writes an append-only stock_movements ledger entry. Sign
// convention: positive quantities into stock, negative out of stock.
func insertMovement(tx *sqlx.Tx, orgID, productID, warehouseID string, locationID, batchID interface{}, movementType string, quantity float64, refDocType, refDocID, reference, userID string) error {
	_, err := tx.Exec(
		`INSERT INTO stock_movements (org_id, product_id, warehouse_id, location_id, batch_id, movement_type, quantity, ref_doc_type, ref_doc_id, reference, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		orgID, productID, warehouseID, locationID, batchID, movementType, quantity, refDocType, refDocID, reference, userID,
	)
	return err
}

func CreateGRN(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		userID := c.GetString("user_id")

		var grn model.GRN
		if errs := validator.BindAndValidate(c, &grn); errs != nil {
			response.Validation(c, errs)
			return
		}

		grn.OrgID = orgID
		grn.CreatedBy = userID
		grn.UpdatedBy = userID
		if grn.Status == "" {
			grn.Status = "draft"
		}

		if len(grn.Items) == 0 {
			err := db.QueryRow(`
				INSERT INTO goods_receipts (org_id, warehouse_id, grn_number, supplier, po_number, received_by, status, notes, created_by, updated_by)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
				RETURNING id, created_at, updated_at`,
				grn.OrgID, grn.WarehouseID, grn.GRNNumber,
				nullIfEmpty(grn.Supplier), nullIfEmpty(grn.PONumber), nullIfEmpty(grn.ReceivedBy),
				grn.Status, nullIfEmpty(grn.Notes), grn.CreatedBy, grn.UpdatedBy,
			).Scan(&grn.ID, &grn.CreatedAt, &grn.UpdatedAt)
			if err != nil {
				response.InternalError(c, "failed to create GRN")
				return
			}
			response.Created(c, grn)
			return
		}

		tx, err := db.Beginx()
		if err != nil {
			response.InternalError(c, "failed to begin transaction")
			return
		}
		defer func() { _ = tx.Rollback() }()

		err = tx.QueryRow(`
			INSERT INTO goods_receipts (org_id, warehouse_id, grn_number, supplier, po_number, received_by, status, notes, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, created_at, updated_at`,
			grn.OrgID, grn.WarehouseID, grn.GRNNumber,
			nullIfEmpty(grn.Supplier), nullIfEmpty(grn.PONumber), nullIfEmpty(grn.ReceivedBy),
			grn.Status, nullIfEmpty(grn.Notes), grn.CreatedBy, grn.UpdatedBy,
		).Scan(&grn.ID, &grn.CreatedAt, &grn.UpdatedAt)
		if err != nil {
			response.InternalError(c, "failed to create GRN")
			return
		}

		for _, item := range grn.Items {
			var batchID interface{} = nil
			if item.BatchNumber != "" {
				err = tx.QueryRowx(
					`INSERT INTO batches (org_id, product_id, batch_number, serial_number, expiry_date)
					 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
					orgID, item.ProductID, item.BatchNumber,
					nullIfEmpty(item.SerialNumber), nullIfEmptyPtr(item.ExpiryDate),
				).Scan(&batchID)
				if err != nil {
					response.InternalError(c, "failed to create batch: "+err.Error())
					return
				}
			}

			_, err = tx.Exec(
				`INSERT INTO grn_line_items (grn_id, product_id, batch_number, serial_number, expiry_date, quantity, unit_cost, status)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,'active')`,
				grn.ID, item.ProductID, nullIfEmpty(item.BatchNumber), nullIfEmpty(item.SerialNumber),
				nullIfEmptyPtr(item.ExpiryDate), item.Quantity, item.UnitCost,
			)
			if err != nil {
				response.InternalError(c, "failed to create GRN line item")
				return
			}

			if err := insertMovement(tx, orgID, item.ProductID, grn.WarehouseID, nil, batchID, "receipt", item.Quantity, "grn", grn.ID, grn.GRNNumber, userID); err != nil {
				response.InternalError(c, "failed to record stock movement: "+err.Error())
				return
			}
			if err := applyStockDelta(tx, orgID, item.ProductID, grn.WarehouseID, nil, batchID, item.Quantity); err != nil {
				response.InternalError(c, err.Error())
				return
			}
		}

		if err := tx.Commit(); err != nil {
			response.InternalError(c, "failed to commit GRN")
			return
		}

		response.Created(c, grn)
	}
}

func CreateIssue(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		userID := c.GetString("user_id")

		var issue model.StockIssue
		if errs := validator.BindAndValidate(c, &issue); errs != nil {
			response.Validation(c, errs)
			return
		}

		issue.OrgID = orgID
		issue.CreatedBy = userID
		issue.UpdatedBy = userID
		if issue.Status == "" {
			issue.Status = "draft"
		}

		err := db.QueryRow(`
			INSERT INTO stock_issues (org_id, warehouse_id, issue_number, requested_by, approved_by, program_id, department_id, status, notes, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id, created_at, updated_at`,
			issue.OrgID, issue.WarehouseID, issue.IssueNumber, issue.RequestedBy,
			issue.ApprovedBy, nullIfEmpty(issue.ProgramID), nullIfEmpty(issue.DepartmentID),
			issue.Status, issue.Notes, issue.CreatedBy, issue.UpdatedBy,
		).Scan(&issue.ID, &issue.CreatedAt, &issue.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create stock issue")
			return
		}

		response.Created(c, issue)
	}
}

func CreateTransfer(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		userID := c.GetString("user_id")

		var req struct {
			Transfer model.StockTransfer      `json:"transfer"`
			Items    []model.TransferLineItem `json:"items" validate:"required,min=1"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		tx, err := db.Beginx()
		if err != nil {
			response.InternalError(c, "failed to begin transaction")
			return
		}
		defer func() { _ = tx.Rollback() }()

		req.Transfer.OrgID = orgID
		req.Transfer.CreatedBy = userID
		req.Transfer.UpdatedBy = userID
		if req.Transfer.Status == "" {
			req.Transfer.Status = "draft"
		}

		err = tx.QueryRow(`
			INSERT INTO stock_transfers (org_id, from_warehouse_id, to_warehouse_id, transfer_number, status, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at`,
			req.Transfer.OrgID, req.Transfer.FromWarehouseID, req.Transfer.ToWarehouseID,
			req.Transfer.TransferNumber, req.Transfer.Status,
			req.Transfer.CreatedBy, req.Transfer.UpdatedBy,
		).Scan(&req.Transfer.ID, &req.Transfer.CreatedAt, &req.Transfer.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create stock transfer")
			return
		}

		for _, item := range req.Items {
			item.TransferID = req.Transfer.ID
			_, err = tx.Exec(`
				INSERT INTO transfer_line_items (transfer_id, product_id, batch_id, quantity, unit_cost, status)
				VALUES ($1, $2, $3, $4, $5, 'active')`,
				item.TransferID, item.ProductID, nullIfEmpty(item.BatchID),
				item.Quantity, item.UnitCost,
			)
			if err != nil {
				response.InternalError(c, "failed to create transfer line item")
				return
			}

			batchID := nullIfEmpty(item.BatchID)
			if err := applyStockDelta(tx, orgID, item.ProductID, req.Transfer.FromWarehouseID, nil, batchID, -item.Quantity); err != nil {
				var ise *InsufficientStockError
				if errors.As(err, &ise) {
					response.BadRequest(c, ise.Error())
					return
				}
				response.InternalError(c, "failed to release transfer stock: "+err.Error())
				return
			}
			if err := insertMovement(tx, orgID, item.ProductID, req.Transfer.FromWarehouseID, nil, batchID, "transfer_out", -item.Quantity, "transfer", req.Transfer.ID, req.Transfer.TransferNumber, userID); err != nil {
				response.InternalError(c, "failed to record transfer_out movement: "+err.Error())
				return
			}
			if err := applyStockDelta(tx, orgID, item.ProductID, req.Transfer.ToWarehouseID, nil, batchID, item.Quantity); err != nil {
				response.InternalError(c, "failed to receive transfer stock: "+err.Error())
				return
			}
			if err := insertMovement(tx, orgID, item.ProductID, req.Transfer.ToWarehouseID, nil, batchID, "transfer_in", item.Quantity, "transfer", req.Transfer.ID, req.Transfer.TransferNumber, userID); err != nil {
				response.InternalError(c, "failed to record transfer_in movement: "+err.Error())
				return
			}
		}

		if err := tx.Commit(); err != nil {
			response.InternalError(c, "failed to commit transfer")
			return
		}

		response.Created(c, req.Transfer)
	}
}

func CreateAdjustment(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		userID := c.GetString("user_id")

		var req struct {
			Adjustment model.StockAdjustment      `json:"adjustment"`
			Items      []model.AdjustmentLineItem `json:"items" validate:"required,min=1"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		tx, err := db.Beginx()
		if err != nil {
			response.InternalError(c, "failed to begin transaction")
			return
		}
		defer func() { _ = tx.Rollback() }()

		req.Adjustment.OrgID = orgID
		req.Adjustment.CreatedBy = userID
		req.Adjustment.UpdatedBy = userID
		if req.Adjustment.Status == "" {
			req.Adjustment.Status = "draft"
		}

		err = tx.QueryRow(`
			INSERT INTO stock_adjustments (org_id, warehouse_id, reason_code, description, status, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at`,
			req.Adjustment.OrgID, req.Adjustment.WarehouseID, req.Adjustment.ReasonCode,
			req.Adjustment.Description, req.Adjustment.Status,
			req.Adjustment.CreatedBy, req.Adjustment.UpdatedBy,
		).Scan(&req.Adjustment.ID, &req.Adjustment.CreatedAt, &req.Adjustment.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create stock adjustment")
			return
		}

		for _, item := range req.Items {
			item.AdjustmentID = req.Adjustment.ID
			item.Difference = item.ActualQuantity - item.ExpectedQuantity
			_, err = tx.Exec(`
				INSERT INTO adjustment_line_items (adjustment_id, product_id, batch_id, location_id, expected_quantity, actual_quantity, difference, reason_code, unit_cost, notes)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
				item.AdjustmentID, item.ProductID, nullIfEmpty(item.BatchID),
				nullIfEmpty(item.LocationID), item.ExpectedQuantity, item.ActualQuantity,
				item.Difference, nullIfEmpty(item.ReasonCode), item.UnitCost, nullIfEmpty(item.Notes),
			)
			if err != nil {
				response.InternalError(c, "failed to create adjustment line item")
				return
			}

			if item.Difference == 0 {
				continue
			}
			movementType := "adjustment_add"
			if item.Difference < 0 {
				movementType = "adjustment_subtract"
			}
			locationID := nullIfEmpty(item.LocationID)
			batchID := nullIfEmpty(item.BatchID)
			if err := applyStockDelta(tx, orgID, item.ProductID, req.Adjustment.WarehouseID, locationID, batchID, item.Difference); err != nil {
				var ise *InsufficientStockError
				if errors.As(err, &ise) {
					response.BadRequest(c, ise.Error())
					return
				}
				response.InternalError(c, "failed to apply adjustment to stock: "+err.Error())
				return
			}
			if err := insertMovement(tx, orgID, item.ProductID, req.Adjustment.WarehouseID, locationID, batchID, movementType, item.Difference, "adjustment", req.Adjustment.ID, req.Adjustment.Description, userID); err != nil {
				response.InternalError(c, "failed to record adjustment movement: "+err.Error())
				return
			}
		}

		if err := tx.Commit(); err != nil {
			response.InternalError(c, "failed to commit adjustment")
			return
		}

		response.Created(c, req.Adjustment)
	}
}

func GetStockLevels(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		params := pagination.FromQuery(c)

		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM stock_levels WHERE org_id = $1`, orgID); err != nil {
			response.InternalError(c, "failed to count stock levels")
			return
		}

		var levels []model.StockLevel
		err := db.Select(&levels, `SELECT * FROM stock_levels WHERE org_id = $1 ORDER BY product_id LIMIT $2 OFFSET $3`, orgID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch stock levels")
			return
		}

		if levels == nil {
			levels = []model.StockLevel{}
		}

		response.Paginated(c, levels, total, params.Page, params.PageSize)
	}
}

func GetStockMovements(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		params := pagination.FromQuery(c)

		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM stock_movements WHERE org_id = $1`, orgID); err != nil {
			response.InternalError(c, "failed to count stock movements")
			return
		}

		var movements []model.StockMovement
		err := db.Select(&movements, `SELECT * FROM stock_movements WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, orgID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch stock movements")
			return
		}

		if movements == nil {
			movements = []model.StockMovement{}
		}

		response.Paginated(c, movements, total, params.Page, params.PageSize)
	}
}

func GetBatchTrail(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		batchID := c.Param("id")

		type BatchTrailEntry struct {
			MovementType string  `json:"movement_type" db:"movement_type"`
			WarehouseID  *string `json:"warehouse_id" db:"warehouse_id"`
			Quantity     float64 `json:"quantity" db:"quantity"`
			RefDocType   string  `json:"ref_doc_type" db:"ref_doc_type"`
			RefDocID     string  `json:"ref_doc_id" db:"ref_doc_id"`
			CreatedBy    string  `json:"created_by" db:"created_by"`
			CreatedAt    string  `json:"created_at" db:"created_at"`
		}

		var trail []BatchTrailEntry
		query := `SELECT movement_type, warehouse_id, quantity, ref_doc_type, ref_doc_id, created_by, created_at FROM stock_movements WHERE org_id = $1 AND batch_id = $2 ORDER BY created_at ASC`
		err := db.Select(&trail, query, orgID, batchID)
		if err != nil {
			response.InternalError(c, "failed to fetch batch trail")
			return
		}

		if len(trail) == 0 {
			response.NotFound(c, "batch trail not found")
			return
		}

		response.OK(c, trail)
	}
}
