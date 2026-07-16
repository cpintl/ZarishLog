package handler

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func CreateDistribution(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrgID              string                      `json:"org_id" validate:"required,uuid7"`
			OrgLevelID         *string                     `json:"org_level_id" validate:"omitempty,uuid7"`
			ProgramID          *string                     `json:"program_id" validate:"omitempty,uuid7"`
			DistributionNumber string                      `json:"distribution_number" validate:"required,max=100"`
			DistributionDate   string                      `json:"distribution_date" validate:"required,date"`
			Location           string                      `json:"location"`
			BeneficiaryCount   *int                        `json:"beneficiary_count" validate:"omitempty,min=0"`
			Status             string                      `json:"status" validate:"omitempty,oneof=draft active completed cancelled"`
			Notes              string                      `json:"notes"`
			CreatedBy          string                      `json:"created_by" validate:"required,max=255"`
			Items              []model.DistributionLineItem `json:"items"`
			Beneficiaries      []model.DistributionBeneficiary `json:"beneficiaries"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		status := req.Status
		if status == "" {
			status = "draft"
		}

		tx, err := db.Beginx()
		if err != nil {
			response.InternalError(c, "failed to begin transaction")
			return
		}
		defer tx.Rollback()

		var distID string
		err = tx.QueryRowx(
			`INSERT INTO distributions (org_id, org_level_id, program_id, distribution_number,
			 distribution_date, location, beneficiary_count, status, notes, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10) RETURNING id`,
			req.OrgID, nullIfEmptyPtr(req.OrgLevelID), nullIfEmptyPtr(req.ProgramID),
			req.DistributionNumber, req.DistributionDate, req.Location,
			req.BeneficiaryCount, status, req.Notes, req.CreatedBy,
		).Scan(&distID)
		if err != nil {
			response.InternalError(c, "failed to create distribution: "+err.Error())
			return
		}

		for _, item := range req.Items {
			itemStatus := item.Status
			if itemStatus == "" {
				itemStatus = "active"
			}
			_, err = tx.Exec(
				`INSERT INTO distribution_line_items (distribution_id, product_id, batch_id,
				 quantity_planned, quantity_distributed, unit_cost, status)
				 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				distID, item.ProductID, nullIfEmptyPtr(item.BatchID),
				item.QuantityPlanned, item.QuantityDistributed, item.UnitCost, itemStatus,
			)
			if err != nil {
				response.InternalError(c, "failed to insert line item: "+err.Error())
				return
			}
		}

		for _, b := range req.Beneficiaries {
			_, err = tx.Exec(
				`INSERT INTO distribution_beneficiaries (distribution_id, beneficiary_type, count, criteria)
				 VALUES ($1,$2,$3,$4)`,
				distID, b.BeneficiaryType, b.Count, b.Criteria,
			)
			if err != nil {
				response.InternalError(c, "failed to insert beneficiary: "+err.Error())
				return
			}
		}

		if err := tx.Commit(); err != nil {
			response.InternalError(c, "failed to commit")
			return
		}

		response.Created(c, gin.H{"id": distID})
	}
}

func ListDistributions(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM distributions`)
		if err != nil {
			response.InternalError(c, "failed to count distributions")
			return
		}

		var dists []model.Distribution
		err = db.Select(&dists,
			`SELECT id, org_id, org_level_id, program_id, distribution_number, distribution_date,
			 location, beneficiary_count, status, notes, created_by, updated_by, created_at, updated_at
			 FROM distributions ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list distributions")
			return
		}

		if dists == nil {
			dists = []model.Distribution{}
		}

		response.Paginated(c, dists, total, p.Page, p.PageSize)
	}
}

func GetDistribution(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var dist model.Distribution
		err := db.Get(&dist,
			`SELECT id, org_id, org_level_id, program_id, distribution_number, distribution_date,
			 location, beneficiary_count, status, notes, created_by, updated_by, created_at, updated_at
			 FROM distributions WHERE id=$1`, id,
		)
		if err != nil {
			response.NotFound(c, "distribution not found")
			return
		}

		var items []model.DistributionLineItem
		err = db.Select(&items,
			`SELECT id, distribution_id, product_id, batch_id, quantity_planned,
			 quantity_distributed, unit_cost, status
			 FROM distribution_line_items WHERE distribution_id=$1`, id,
		)
		if err != nil {
			items = []model.DistributionLineItem{}
		}

		var beneficiaries []model.DistributionBeneficiary
		err = db.Select(&beneficiaries,
			`SELECT id, distribution_id, beneficiary_type, count, criteria
			 FROM distribution_beneficiaries WHERE distribution_id=$1`, id,
		)
		if err != nil {
			beneficiaries = []model.DistributionBeneficiary{}
		}

		response.OK(c, gin.H{
			"distribution":  dist,
			"items":         items,
			"beneficiaries": beneficiaries,
		})
	}
}
