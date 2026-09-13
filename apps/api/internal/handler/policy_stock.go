package handler

import (
	"time"

	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
)

// Policy §12 / §14 — Quality release and controlled-product registers

func CreateStockReleaseRecord(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID               string  `json:"org_id" validate:"required,uuid7"`
			ReleaseNumber       string  `json:"release_number" validate:"required,max=100"`
			WarehouseID         string  `json:"warehouse_id" validate:"required,uuid7"`
			ProductID           string  `json:"product_id" validate:"required,uuid7"`
			BatchID             *string `json:"batch_id" validate:"opt_uuid7"`
			LocationID          *string `json:"location_id" validate:"opt_uuid7"`
			Quantity            float64 `json:"quantity" validate:"required"`
			QuarantineReference *string `json:"quarantine_reference"`
			EvidenceReviewed    *string `json:"evidence_reviewed"`
			Decision            string  `json:"decision" validate:"required,oneof=release restricted_release reject destruction return"`
			DecisionDate        string  `json:"decision_date" validate:"date"`
			DecidedBy           string  `json:"decided_by" validate:"required,max=255"`
			Notes               *string `json:"notes"`
			CreatedBy           string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		decisionDate := req.DecisionDate
		if decisionDate == "" {
			decisionDate = time.Now().UTC().Format("2006-01-02")
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO stock_release_records (org_id, release_number, warehouse_id, product_id, batch_id, location_id, quantity, quarantine_reference, evidence_reviewed, decision, decision_date, decided_by, notes, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$14) RETURNING id`,
			req.OrgID, req.ReleaseNumber, req.WarehouseID, req.ProductID, nullIfEmptyPtr(req.BatchID), nullIfEmptyPtr(req.LocationID),
			req.Quantity, nullIfEmptyPtr(req.QuarantineReference), nullIfEmptyPtr(req.EvidenceReviewed),
			req.Decision, decisionDate, req.DecidedBy, nullIfEmptyPtr(req.Notes), req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create stock release record: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func GetStockReleaseRecord(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var record model.StockReleaseRecord
		err := db.Get(&record, `SELECT * FROM stock_release_records WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "stock release record not found")
			return
		}

		response.OK(c, record)
	}
}

func ListStockReleaseRecords(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM stock_release_records`)
		if err != nil {
			response.InternalError(c, "failed to count stock release records")
			return
		}

		var records []model.StockReleaseRecord
		err = db.Select(&records,
			`SELECT * FROM stock_release_records ORDER BY decision_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list stock release records")
			return
		}

		if records == nil {
			records = []model.StockReleaseRecord{}
		}

		response.Paginated(c, records, total, p.Page, p.PageSize)
	}
}

func CreateControlledStockRegisterEntry(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID               string  `json:"org_id" validate:"required,uuid7"`
			ProductID           string  `json:"product_id" validate:"required,uuid7"`
			BatchID             *string `json:"batch_id" validate:"opt_uuid7"`
			WarehouseID         string  `json:"warehouse_id" validate:"required,uuid7"`
			RegisterDate        string  `json:"register_date" validate:"date"`
			TransactionType     string  `json:"transaction_type" validate:"required,oneof=receipt issue transfer_in transfer_out return disposal reconciliation opening_balance"`
			ReferenceDocument   *string `json:"reference_document"`
			Quantity            float64 `json:"quantity" validate:"required"`
			BalanceAfter        float64 `json:"balance_after" validate:"required"`
			ReceivedFrom        *string `json:"received_from"`
			IssuedTo            *string `json:"issued_to"`
			AuthorizedRecipient *string `json:"authorized_recipient"`
			Signature           *string `json:"signature"`
			Discrepancy         float64 `json:"discrepancy"`
			CreatedBy           string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		registerDate := req.RegisterDate
		if registerDate == "" {
			registerDate = time.Now().UTC().Format("2006-01-02")
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO controlled_stock_register (org_id, product_id, batch_id, warehouse_id, register_date, transaction_type, reference_document, quantity, balance_after, received_from, issued_to, authorized_recipient, signature, discrepancy, created_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id`,
			req.OrgID, req.ProductID, nullIfEmptyPtr(req.BatchID), req.WarehouseID, registerDate, req.TransactionType,
			nullIfEmptyPtr(req.ReferenceDocument), req.Quantity, req.BalanceAfter,
			nullIfEmptyPtr(req.ReceivedFrom), nullIfEmptyPtr(req.IssuedTo), nullIfEmptyPtr(req.AuthorizedRecipient),
			nullIfEmptyPtr(req.Signature), req.Discrepancy, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create controlled stock register entry: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func GetControlledStockRegisterEntry(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var entry model.ControlledStockRegisterEntry
		err := db.Get(&entry, `SELECT * FROM controlled_stock_register WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "controlled stock register entry not found")
			return
		}

		response.OK(c, entry)
	}
}

func ListControlledStockRegister(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM controlled_stock_register`)
		if err != nil {
			response.InternalError(c, "failed to count controlled stock entries")
			return
		}

		var entries []model.ControlledStockRegisterEntry
		err = db.Select(&entries,
			`SELECT * FROM controlled_stock_register ORDER BY register_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list controlled stock entries")
			return
		}

		if entries == nil {
			entries = []model.ControlledStockRegisterEntry{}
		}

		response.Paginated(c, entries, total, p.Page, p.PageSize)
	}
}

func CreateShortExpiryReview(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID                    string  `json:"org_id" validate:"required,uuid7"`
			ReviewDate               string  `json:"review_date" validate:"date"`
			ProductID                string  `json:"product_id" validate:"required,uuid7"`
			BatchID                  *string `json:"batch_id" validate:"opt_uuid7"`
			WarehouseID              string  `json:"warehouse_id" validate:"required,uuid7"`
			ExpiryDate               string  `json:"expiry_date" validate:"required,date"`
			RemainingShelfLifeMonths int     `json:"remaining_shelf_life_months" validate:"required"`
			ExpectedConsumption      *string `json:"expected_consumption"`
			TransferOption           *string `json:"transfer_option"`
			DonorCondition           *string `json:"donor_condition"`
			RegulatoryRestriction    *string `json:"regulatory_restriction"`
			Decision                 string  `json:"decision" validate:"required,oneof=use_before_expiry transfer redistribute return donate destroy pending"`
			DecisionBy               *string `json:"decision_by"`
			DecisionDate             *string `json:"decision_date" validate:"opt_date"`
			Notes                    *string `json:"notes"`
			CreatedBy                string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		reviewDate := req.ReviewDate
		if reviewDate == "" {
			reviewDate = time.Now().UTC().Format("2006-01-02")
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO short_expiry_reviews (org_id, review_date, product_id, batch_id, warehouse_id, expiry_date, remaining_shelf_life_months, expected_consumption, transfer_option, donor_condition, regulatory_restriction, decision, decision_by, decision_date, notes, created_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING id`,
			req.OrgID, reviewDate, req.ProductID, nullIfEmptyPtr(req.BatchID), req.WarehouseID, req.ExpiryDate,
			req.RemainingShelfLifeMonths, nullIfEmptyPtr(req.ExpectedConsumption), nullIfEmptyPtr(req.TransferOption),
			nullIfEmptyPtr(req.DonorCondition), nullIfEmptyPtr(req.RegulatoryRestriction), req.Decision,
			nullIfEmptyPtr(req.DecisionBy), nullIfEmptyPtr(req.DecisionDate), nullIfEmptyPtr(req.Notes), req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create short-expiry review: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func ListShortExpiryReviews(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM short_expiry_reviews`)
		if err != nil {
			response.InternalError(c, "failed to count short-expiry reviews")
			return
		}

		var reviews []model.ShortExpiryReview
		err = db.Select(&reviews,
			`SELECT * FROM short_expiry_reviews ORDER BY expiry_date ASC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list short-expiry reviews")
			return
		}

		if reviews == nil {
			reviews = []model.ShortExpiryReview{}
		}

		response.Paginated(c, reviews, total, p.Page, p.PageSize)
	}
}
