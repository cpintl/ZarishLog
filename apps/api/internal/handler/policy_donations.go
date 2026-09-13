package handler

import (
	"time"

	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
)

// Policy §11.3 — Donation registry and acceptance/rejection workflow

type donationLineItemReq struct {
	ProductID                string  `json:"product_id" validate:"required,uuid7"`
	BatchNumber              *string `json:"batch_number"`
	ExpiryDate               *string `json:"expiry_date" validate:"opt_date"`
	Quantity                 float64 `json:"quantity" validate:"required"`
	Uom                      *string `json:"uom"`
	RemainingShelfLifeMonths *int    `json:"remaining_shelf_life_months"`
	Condition                *string `json:"condition" validate:"omitempty,oneof=intact short_dated damaged unlabelled suspected_falsified"`
	ShelfLifeCompliant       *bool   `json:"shelf_life_compliant"`
	Authorized               bool    `json:"authorized"`
	Notes                    *string `json:"notes"`
}

func CreateDonation(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID                   string                `json:"org_id" validate:"required,uuid7"`
			DonationNumber          string                `json:"donation_number" validate:"required,max=100"`
			DonorName               string                `json:"donor_name" validate:"required,max=255"`
			DonorContact            *string               `json:"donor_contact"`
			OfferDate               string                `json:"offer_date" validate:"date"`
			NeedsAssessment         *string               `json:"needs_assessment"`
			ProposedRecipient       *string               `json:"proposed_recipient"`
			ApprovalReference       *string               `json:"approval_reference"`
			TransportResponsibility *string               `json:"transport_responsibility"`
			CustomsNotes            *string               `json:"customs_notes"`
			DisposalResponsibility  *string               `json:"disposal_responsibility"`
			Status                  string                `json:"status" validate:"omitempty,oneof=draft pending accepted rejected quarantined received"`
			Notes                   *string               `json:"notes"`
			CreatedBy               string                `json:"created_by" validate:"required,max=255"`
			LineItems               []donationLineItemReq `json:"line_items"`
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
		offerDate := req.OfferDate
		if offerDate == "" {
			offerDate = time.Now().UTC().Format("2006-01-02")
		}

		tx, err := db.Beginx()
		if err != nil {
			response.InternalError(c, "failed to begin transaction")
			return
		}
		defer func() { _ = tx.Rollback() }()

		var id string
		err = tx.QueryRowx(
			`INSERT INTO donations (org_id, donation_number, donor_name, donor_contact, offer_date, needs_assessment, proposed_recipient, approval_reference, transport_responsibility, customs_notes, disposal_responsibility, status, notes, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$14) RETURNING id`,
			req.OrgID, req.DonationNumber, req.DonorName, nullIfEmptyPtr(req.DonorContact), offerDate,
			nullIfEmptyPtr(req.NeedsAssessment), nullIfEmptyPtr(req.ProposedRecipient), nullIfEmptyPtr(req.ApprovalReference),
			nullIfEmptyPtr(req.TransportResponsibility), nullIfEmptyPtr(req.CustomsNotes), nullIfEmptyPtr(req.DisposalResponsibility),
			status, nullIfEmptyPtr(req.Notes), req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create donation: "+err.Error())
			return
		}

		for _, li := range req.LineItems {
			_, err = tx.Exec(
				`INSERT INTO donation_line_items (donation_id, product_id, batch_number, expiry_date, quantity, uom, remaining_shelf_life_months, condition, shelf_life_compliant, authorized, notes)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
				id, li.ProductID, nullIfEmptyPtr(li.BatchNumber), nullIfEmptyPtr(li.ExpiryDate),
				li.Quantity, nullIfEmptyPtr(li.Uom), li.RemainingShelfLifeMonths, nullIfEmptyPtr(li.Condition),
				li.ShelfLifeCompliant, li.Authorized, nullIfEmptyPtr(li.Notes),
			)
			if err != nil {
				response.InternalError(c, "failed to insert donation line item: "+err.Error())
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

func GetDonation(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		id := c.Param("id")

		var donation model.Donation
		err := db.Get(&donation, `SELECT * FROM donations WHERE id=$1`, id)
		if err != nil {
			response.NotFound(c, "donation not found")
			return
		}

		var lineItems []model.DonationLineItem
		err = db.Select(&lineItems,
			`SELECT * FROM donation_line_items WHERE donation_id=$1 ORDER BY expiry_date NULLS LAST, id`, id,
		)
		if err != nil {
			lineItems = []model.DonationLineItem{}
		}

		response.OK(c, gin.H{
			"donation":   donation,
			"line_items": lineItems,
		})
	}
}

func ListDonations(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM donations`)
		if err != nil {
			response.InternalError(c, "failed to count donations")
			return
		}

		var donations []model.Donation
		err = db.Select(&donations,
			`SELECT * FROM donations ORDER BY offer_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list donations")
			return
		}

		if donations == nil {
			donations = []model.Donation{}
		}

		response.Paginated(c, donations, total, p.Page, p.PageSize)
	}
}

// UpdateDonationDecision records the acceptance/rejection/quarantine decision
// (§11.3). An accepted donation records the certificate number/date; other
// decisions leave the certificate fields null.
func UpdateDonationDecision(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			Decision              string  `json:"decision" validate:"required,oneof=accepted rejected quarantined pending"`
			DecisionDate          string  `json:"decision_date" validate:"date"`
			DecidedBy             string  `json:"decided_by" validate:"required,max=255"`
			CertificateNumber     *string `json:"certificate_number"`
			CertificateIssuedDate *string `json:"certificate_issued_date" validate:"opt_date"`
			UpdatedBy             string  `json:"updated_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		decisionDate := req.DecisionDate
		if decisionDate == "" {
			decisionDate = time.Now().UTC().Format("2006-01-02")
		}

		_, err := db.Exec(
			`UPDATE donations SET decision=$1, status=$1, decision_date=$2, decided_by=$3, certificate_number=$4, certificate_issued_date=$5, updated_by=$6, updated_at=now() WHERE id=$7`,
			req.Decision, decisionDate, req.DecidedBy,
			nullIfEmptyPtr(req.CertificateNumber), nullIfEmptyPtr(req.CertificateIssuedDate), req.UpdatedBy, c.Param("id"),
		)
		if err != nil {
			response.InternalError(c, "failed to update donation decision: "+err.Error())
			return
		}

		response.OK(c, gin.H{"status": req.Decision})
	}
}
