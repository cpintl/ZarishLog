package handler

import (
	"fmt"
	"time"

	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Policy §8 — Regulatory & Legal Compliance (Bangladesh DGDA/DNC matrix)

func CreateRegulatoryApproval(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrgID            string  `json:"org_id" validate:"required,uuid7"`
			ScopeType        string  `json:"scope_type" validate:"required,oneof=warehouse product product_category supplier organization"`
			ScopeID          string  `json:"scope_id" validate:"required,max=255"`
			Authority        string  `json:"authority" validate:"required,oneof=DGDA DNC NBR CCI_E Customs Municipality Other"`
			LicenseType      string  `json:"license_type" validate:"required,oneof=wholesale_license retail_license import_permit controlled_substance_permit premises_approval import_clearance registration other"`
			ReferenceNumber  string  `json:"reference_number" validate:"required,max=255"`
			ProductScope     *string `json:"product_scope"`
			Conditions       *string `json:"conditions"`
			IssueDate        *string `json:"issue_date" validate:"opt_date"`
			ExpiryDate       *string `json:"expiry_date" validate:"opt_date"`
			RenewalLeadDays  int     `json:"renewal_lead_days"`
			ResponsibleOwner *string `json:"responsible_owner"`
			Status           string  `json:"status" validate:"omitempty,oneof=draft active expiring_soon expired revoked suspended"`
			EvidenceURL      *string `json:"evidence_url"`
			Notes            *string `json:"notes"`
			CreatedBy        string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		status := req.Status
		if status == "" {
			status = "draft"
		}
		renewalLead := req.RenewalLeadDays
		if renewalLead == 0 {
			renewalLead = 60
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO regulatory_approvals (org_id, scope_type, scope_id, authority, license_type, reference_number, product_scope, conditions, issue_date, expiry_date, renewal_lead_days, responsible_owner, status, evidence_url, notes, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$16) RETURNING id`,
			req.OrgID, req.ScopeType, req.ScopeID, req.Authority, req.LicenseType, req.ReferenceNumber,
			nullIfEmptyPtr(req.ProductScope), nullIfEmptyPtr(req.Conditions), nullIfEmptyPtr(req.IssueDate), nullIfEmptyPtr(req.ExpiryDate),
			renewalLead, nullIfEmptyPtr(req.ResponsibleOwner), status, nullIfEmptyPtr(req.EvidenceURL), nullIfEmptyPtr(req.Notes),
			req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create regulatory approval: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func ListRegulatoryApprovals(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM regulatory_approvals`)
		if err != nil {
			response.InternalError(c, "failed to count regulatory approvals")
			return
		}

		var approvals []model.RegulatoryApproval
		err = db.Select(&approvals,
			`SELECT * FROM regulatory_approvals ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list regulatory approvals")
			return
		}

		if approvals == nil {
			approvals = []model.RegulatoryApproval{}
		}

		response.Paginated(c, approvals, total, p.Page, p.PageSize)
	}
}

func GetRegulatoryApproval(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var approval model.RegulatoryApproval
		err := db.Get(&approval, `SELECT * FROM regulatory_approvals WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "regulatory approval not found")
			return
		}

		response.OK(c, approval)
	}
}

func ListExpiringRegulatoryApprovals(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.DefaultQuery("days", "60")
		var days int
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil || days < 1 {
			days = 60
		}
		p := pagination.FromQuery(c)

		cutoff := time.Now().AddDate(0, 0, days)

		var total int
		err := db.Get(&total,
			`SELECT COUNT(*) FROM regulatory_approvals WHERE expiry_date IS NOT NULL AND expiry_date <= $1`,
			cutoff.Format("2006-01-02"),
		)
		if err != nil {
			response.InternalError(c, "failed to count expiring approvals")
			return
		}

		var approvals []model.RegulatoryApproval
		err = db.Select(&approvals,
			`SELECT * FROM regulatory_approvals WHERE expiry_date IS NOT NULL AND expiry_date <= $1 ORDER BY expiry_date ASC LIMIT $2 OFFSET $3`,
			cutoff.Format("2006-01-02"), p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list expiring approvals")
			return
		}

		if approvals == nil {
			approvals = []model.RegulatoryApproval{}
		}

		response.Paginated(c, approvals, total, p.Page, p.PageSize)
	}
}
