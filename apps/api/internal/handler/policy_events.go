package handler

import (
	"time"

	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
)

// Policy §17 — Complaints, recalls (incl. mock recall) and falsified-product triage

func CreateComplaint(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID                      string  `json:"org_id" validate:"required,uuid7"`
			ComplaintNumber            string  `json:"complaint_number" validate:"required,max=100"`
			ReceivedDate               string  `json:"received_date" validate:"date"`
			ReceivedVia                *string `json:"received_via" validate:"omitempty,oneof=field_visit phone email form third_party other"`
			Severity                   *string `json:"severity"`
			Category                   *string `json:"category" validate:"omitempty,oneof=suspected_harm falsification contamination wrong_product controlled_diversion quality_widespread packaging other"`
			ProductID                  *string `json:"product_id" validate:"opt_uuid7"`
			BatchID                    *string `json:"batch_id" validate:"opt_uuid7"`
			Description                string  `json:"description" validate:"required"`
			ReporterName               *string `json:"reporter_name"`
			ReporterContact            *string `json:"reporter_contact"`
			ImmediateEscalation        bool    `json:"immediate_escalation"`
			EscalatedTo                *string `json:"escalated_to"`
			ManufacturerNotified       bool    `json:"manufacturer_notified"`
			CompetentAuthorityNotified bool    `json:"competent_authority_notified"`
			Status                     string  `json:"status" validate:"omitempty,oneof=open under_review investigating resolved closed"`
			ResolutionNotes            *string `json:"resolution_notes"`
			CreatedBy                  string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		status := req.Status
		if status == "" {
			status = "open"
		}
		receivedDate := req.ReceivedDate
		if receivedDate == "" {
			receivedDate = time.Now().UTC().Format("2006-01-02")
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO complaints (org_id, complaint_number, received_date, received_via, severity, category, product_id, batch_id, description, reporter_name, reporter_contact, immediate_escalation, escalated, escalated_to, manufacturer_notified, competent_authority_notified, status, resolution_notes, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$19) RETURNING id`,
			req.OrgID, req.ComplaintNumber, receivedDate, nullIfEmptyPtr(req.ReceivedVia),
			nullIfEmptyPtr(req.Severity), nullIfEmptyPtr(req.Category), nullIfEmptyPtr(req.ProductID), nullIfEmptyPtr(req.BatchID),
			req.Description, nullIfEmptyPtr(req.ReporterName), nullIfEmptyPtr(req.ReporterContact),
			req.ImmediateEscalation, req.ImmediateEscalation, nullIfEmptyPtr(req.EscalatedTo),
			req.ManufacturerNotified, req.CompetentAuthorityNotified,
			status, nullIfEmptyPtr(req.ResolutionNotes), req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create complaint: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func GetComplaint(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var complaint model.Complaint
		err := db.Get(&complaint, `SELECT * FROM complaints WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "complaint not found")
			return
		}

		response.OK(c, complaint)
	}
}

func ListComplaints(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM complaints`)
		if err != nil {
			response.InternalError(c, "failed to count complaints")
			return
		}

		var complaints []model.Complaint
		err = db.Select(&complaints,
			`SELECT * FROM complaints ORDER BY received_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list complaints")
			return
		}

		if complaints == nil {
			complaints = []model.Complaint{}
		}

		response.Paginated(c, complaints, total, p.Page, p.PageSize)
	}
}

type recallLineItemReq struct {
	Recipient          string  `json:"recipient" validate:"required,max=255"`
	Location           *string `json:"location"`
	ProductID          *string `json:"product_id" validate:"opt_uuid7"`
	BatchNumber        *string `json:"batch_number"`
	QuantityIssued     float64 `json:"quantity_issued"`
	QuantityRecovered  float64 `json:"quantity_recovered"`
	StorageDisposition *string `json:"storage_disposition" validate:"omitempty,oneof=quarantined returned destroyed secured"`
	RecoveredAt        *string `json:"recovered_at"`
	Notes              *string `json:"notes"`
}

func CreateRecall(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID           string              `json:"org_id" validate:"required,uuid7"`
			RecallNumber    string              `json:"recall_number" validate:"required,max=100"`
			RecallType      string              `json:"recall_type" validate:"required,oneof=ACTUAL MOCK"`
			RecallDate      string              `json:"recall_date" validate:"date"`
			ProductID       *string             `json:"product_id" validate:"opt_uuid7"`
			SubjectProduct  string              `json:"subject_product" validate:"required,max=255"`
			BatchNumbers    *string             `json:"batch_numbers"`
			Reason          string              `json:"reason" validate:"required"`
			InitiatedBy     *string             `json:"initiated_by"`
			AuthorizedBy    *string             `json:"authorized_by"`
			Communications  *string             `json:"communications"`
			DispositionPlan *string             `json:"disposition_plan"`
			Status          string              `json:"status" validate:"omitempty,oneof=draft in_progress pending_disposition reconciled closed"`
			LineItems       []recallLineItemReq `json:"line_items"`
			CreatedBy       string              `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		status := req.Status
		if status == "" {
			status = "draft"
		}
		recallDate := req.RecallDate
		if recallDate == "" {
			recallDate = time.Now().UTC().Format("2006-01-02")
		}

		tx, err := db.Beginx()
		if err != nil {
			response.InternalError(c, "failed to begin transaction")
			return
		}
		defer func() { _ = tx.Rollback() }()

		var id string
		err = tx.QueryRowx(
			`INSERT INTO recalls (org_id, recall_number, recall_type, recall_date, product_id, subject_product, batch_numbers, reason, initiated_by, authorized_by, communications, disposition_plan, status, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$14) RETURNING id`,
			req.OrgID, req.RecallNumber, req.RecallType, recallDate, nullIfEmptyPtr(req.ProductID),
			req.SubjectProduct, nullIfEmptyPtr(req.BatchNumbers), req.Reason,
			nullIfEmptyPtr(req.InitiatedBy), nullIfEmptyPtr(req.AuthorizedBy),
			nullIfEmptyPtr(req.Communications), nullIfEmptyPtr(req.DispositionPlan),
			status, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create recall: "+err.Error())
			return
		}

		for _, li := range req.LineItems {
			outstanding := li.QuantityIssued - li.QuantityRecovered
			_, err = tx.Exec(
				`INSERT INTO recall_line_items (recall_id, recipient, location, product_id, batch_number, quantity_issued, quantity_recovered, quantity_outstanding, storage_disposition, recovered_at, notes)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
				id, li.Recipient, nullIfEmptyPtr(li.Location), nullIfEmptyPtr(li.ProductID), nullIfEmptyPtr(li.BatchNumber),
				li.QuantityIssued, li.QuantityRecovered, outstanding, nullIfEmptyPtr(li.StorageDisposition),
				nullIfEmptyPtr(li.RecoveredAt), nullIfEmptyPtr(li.Notes),
			)
			if err != nil {
				response.InternalError(c, "failed to insert recall line item: "+err.Error())
				return
			}
		}

		if err := tx.Commit(); err != nil {
			response.InternalError(c, "failed to commit")
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func GetRecall(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		id := c.Param("id")

		var recall model.Recall
		err := db.Get(&recall, `SELECT * FROM recalls WHERE id=$1`, id)
		if err != nil {
			response.NotFound(c, "recall not found")
			return
		}

		var lineItems []model.RecallLineItem
		err = db.Select(&lineItems,
			`SELECT * FROM recall_line_items WHERE recall_id=$1 ORDER BY recovered_at DESC`, id,
		)
		if err != nil {
			lineItems = []model.RecallLineItem{}
		}

		response.OK(c, gin.H{
			"recall":     recall,
			"line_items": lineItems,
		})
	}
}

func ListRecalls(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM recalls`)
		if err != nil {
			response.InternalError(c, "failed to count recalls")
			return
		}

		var recalls []model.Recall
		err = db.Select(&recalls,
			`SELECT * FROM recalls ORDER BY recall_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list recalls")
			return
		}

		if recalls == nil {
			recalls = []model.Recall{}
		}

		response.Paginated(c, recalls, total, p.Page, p.PageSize)
	}
}
