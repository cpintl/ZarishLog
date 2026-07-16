package handler

import (
	"fmt"
	"time"

	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func CreateInspection(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrgID          string                  `json:"org_id" validate:"required,uuid7"`
			GRNID          *string                 `json:"grn_id" validate:"omitempty,uuid7"`
			ProductID      string                  `json:"product_id" validate:"required,uuid7"`
			BatchID        *string                 `json:"batch_id" validate:"omitempty,uuid7"`
			InspectionDate string                  `json:"inspection_date" validate:"required,date"`
			Inspector      string                  `json:"inspector" validate:"required,max=255"`
			Result         string                  `json:"result" validate:"required,oneof=pass fail quarantine"`
			Notes          string                  `json:"notes"`
			CreatedBy      string                  `json:"created_by" validate:"required,max=255"`
			Results        []model.QAChecklistResult `json:"results"`
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
		defer tx.Rollback()

		var id string
		err = tx.QueryRowx(
			`INSERT INTO qa_inspections (org_id, grn_id, product_id, batch_id, inspection_date, inspector, result, notes, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9) RETURNING id`,
			req.OrgID, nullIfEmptyPtr(req.GRNID), req.ProductID, nullIfEmptyPtr(req.BatchID),
			req.InspectionDate, req.Inspector, req.Result, req.Notes, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create inspection: "+err.Error())
			return
		}

		for _, r := range req.Results {
			_, err = tx.Exec(
				`INSERT INTO qa_checklist_results (inspection_id, checklist_item_id, answer, score, notes)
				 VALUES ($1,$2,$3,$4,$5)`,
				id, r.ChecklistItemID, r.Answer, r.Score, r.Notes,
			)
			if err != nil {
				response.InternalError(c, "failed to insert checklist result: "+err.Error())
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

func ListInspections(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM qa_inspections`)
		if err != nil {
			response.InternalError(c, "failed to count inspections")
			return
		}

		var inspections []model.QAInspection
		err = db.Select(&inspections,
			`SELECT id, org_id, grn_id, product_id, batch_id, inspection_date, inspector, result, notes, disposition,
			        created_by, updated_by, created_at, updated_at
			 FROM qa_inspections ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list inspections")
			return
		}

		if inspections == nil {
			inspections = []model.QAInspection{}
		}

		response.Paginated(c, inspections, total, p.Page, p.PageSize)
	}
}

func GetInspection(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var insp model.QAInspection
		err := db.Get(&insp,
			`SELECT id, org_id, grn_id, product_id, batch_id, inspection_date, inspector, result, notes, disposition,
			        created_by, updated_by, created_at, updated_at
			 FROM qa_inspections WHERE id=$1`, id,
		)
		if err != nil {
			response.NotFound(c, "inspection not found")
			return
		}

		var results []model.QAChecklistResult
		err = db.Select(&results,
			`SELECT id, inspection_id, checklist_item_id, answer, score, notes, created_at
			 FROM qa_checklist_results WHERE inspection_id=$1`, id,
		)
		if err != nil {
			results = []model.QAChecklistResult{}
		}

		var disposition *model.QADisposition
		var disp model.QADisposition
		err = db.Get(&disp,
			`SELECT id, inspection_id, disposition_type, disposition_date, approved_by,
			        destination_location_id, notes, created_by, created_at, updated_at
			 FROM qa_dispositions WHERE inspection_id=$1`, id,
		)
		if err == nil {
			disposition = &disp
		}

		response.OK(c, gin.H{
			"inspection":  insp,
			"results":     results,
			"disposition": disposition,
		})
	}
}

func CreateChecklistTemplate(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrgID       string               `json:"org_id" validate:"required,uuid7"`
			Code        string               `json:"code" validate:"required,max=50"`
			Name        string               `json:"name" validate:"required,max=255"`
			Description string               `json:"description"`
			Category    string               `json:"category" validate:"required,max=100"`
			IsMandatory bool                 `json:"is_mandatory"`
			Status      string               `json:"status" validate:"omitempty,oneof=active inactive draft archived"`
			CreatedBy   string               `json:"created_by" validate:"required,max=255"`
			Items       []model.QAChecklistItem `json:"items"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		status := req.Status
		if status == "" {
			status = "active"
		}

		tx, err := db.Beginx()
		if err != nil {
			response.InternalError(c, "failed to begin transaction")
			return
		}
		defer tx.Rollback()

		var templateID string
		err = tx.QueryRowx(
			`INSERT INTO qa_checklist_templates (org_id, code, name, description, category, is_mandatory, status, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8) RETURNING id`,
			req.OrgID, req.Code, req.Name, req.Description, req.Category, req.IsMandatory, status, req.CreatedBy,
		).Scan(&templateID)
		if err != nil {
			response.InternalError(c, "failed to create template: "+err.Error())
			return
		}

		for _, item := range req.Items {
			_, err = tx.Exec(
				`INSERT INTO qa_checklist_items (template_id, item_order, question, expected_answer, is_critical, weight)
				 VALUES ($1,$2,$3,$4,$5,$6)`,
				templateID, item.ItemOrder, item.Question, item.ExpectedAnswer, item.IsCritical, item.Weight,
			)
			if err != nil {
				response.InternalError(c, "failed to insert checklist item: "+err.Error())
				return
			}
		}

		if err := tx.Commit(); err != nil {
			response.InternalError(c, "failed to commit")
			return
		}

		response.Created(c, gin.H{"id": templateID})
	}
}

func ListChecklistTemplates(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM qa_checklist_templates`)
		if err != nil {
			response.InternalError(c, "failed to count templates")
			return
		}

		var templates []model.QAChecklistTemplate
		err = db.Select(&templates,
			`SELECT id, org_id, code, name, description, category, is_mandatory, status, created_by, updated_by, created_at, updated_at
			 FROM qa_checklist_templates ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list templates")
			return
		}

		if templates == nil {
			templates = []model.QAChecklistTemplate{}
		}

		response.Paginated(c, templates, total, p.Page, p.PageSize)
	}
}

func GetChecklistTemplate(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var tpl model.QAChecklistTemplate
		err := db.Get(&tpl,
			`SELECT id, org_id, code, name, description, category, is_mandatory, status, created_by, updated_by, created_at, updated_at
			 FROM qa_checklist_templates WHERE id=$1`, id,
		)
		if err != nil {
			response.NotFound(c, "checklist template not found")
			return
		}

		var items []model.QAChecklistItem
		err = db.Select(&items,
			`SELECT id, template_id, item_order, question, expected_answer, is_critical, weight
			 FROM qa_checklist_items WHERE template_id=$1 ORDER BY item_order`, id,
		)
		if err != nil {
			items = []model.QAChecklistItem{}
		}

		response.OK(c, gin.H{
			"template": tpl,
			"items":    items,
		})
	}
}

func CreateDisposition(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		inspectionID := c.Param("id")

		var req struct {
			DispositionType       string  `json:"disposition_type" validate:"required,oneof=pass fail quarantine rework partial"`
			DispositionDate       string  `json:"disposition_date" validate:"required,date"`
			ApprovedBy            string  `json:"approved_by" validate:"required,max=255"`
			DestinationLocationID *string `json:"destination_location_id" validate:"omitempty,uuid7"`
			Notes                 string  `json:"notes"`
			CreatedBy             string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO qa_dispositions (inspection_id, disposition_type, disposition_date, approved_by, destination_location_id, notes, created_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
			inspectionID, req.DispositionType, req.DispositionDate, req.ApprovedBy,
			nullIfEmptyPtr(req.DestinationLocationID), req.Notes, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create disposition: "+err.Error())
			return
		}

		_, _ = db.Exec(`UPDATE qa_inspections SET disposition=$1, updated_by=$2, updated_at=now() WHERE id=$3`,
			req.DispositionType, req.CreatedBy, inspectionID,
		)

		response.Created(c, gin.H{"id": id})
	}
}

func GetExpiringStock(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.DefaultQuery("days", "30")
		var days int
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil || days < 1 {
			days = 30
		}
		p := pagination.FromQuery(c)

		cutoff := time.Now().AddDate(0, 0, days)

		var total int
		err := db.Get(&total,
			`SELECT COUNT(*) FROM batches WHERE expiry_date IS NOT NULL AND expiry_date <= $1 AND expiry_date >= CURRENT_DATE`,
			cutoff.Format("2006-01-02"),
		)
		if err != nil {
			response.InternalError(c, "failed to count expiring batches")
			return
		}

		type ExpiringBatch struct {
			BatchID    string  `json:"batch_id" db:"batch_id"`
			ProductID  string  `json:"product_id" db:"product_id"`
			BatchRef   string  `json:"batch_ref" db:"batch_ref"`
			ExpiryDate string  `json:"expiry_date" db:"expiry_date"`
			DaysUntil  int     `json:"days_until"`
			Quantity   float64 `json:"quantity" db:"quantity"`
		}

		var batches []ExpiringBatch
		err = db.Select(&batches,
			`SELECT b.id AS batch_id, b.product_id, b.batch_ref, b.expiry_date, COALESCE(sl.quantity,0) AS quantity
			 FROM batches b
			 LEFT JOIN stock_levels sl ON sl.batch_id = b.id
			 WHERE b.expiry_date IS NOT NULL AND b.expiry_date <= $1 AND b.expiry_date >= CURRENT_DATE
			 ORDER BY b.expiry_date ASC LIMIT $2 OFFSET $3`,
			cutoff.Format("2006-01-02"), p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list expiring batches")
			return
		}

		for i := range batches {
			ed, err := time.Parse("2006-01-02", batches[i].ExpiryDate)
			if err == nil {
				batches[i].DaysUntil = int(time.Until(ed).Hours() / 24)
			}
		}

		if batches == nil {
			batches = []ExpiringBatch{}
		}

		response.Paginated(c, batches, total, p.Page, p.PageSize)
	}
}
