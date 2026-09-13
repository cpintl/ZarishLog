package handler

import (
	"time"

	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
)

// Policy §7 — QMS: Deviations, CAPA, Training & Competency

func CreateDeviation(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID              string  `json:"org_id" validate:"required,uuid7"`
			DeviationNumber    string  `json:"deviation_number" validate:"required,max=100"`
			CategoryCode       *string `json:"category_code"`
			OccurredAt         string  `json:"occurred_at"`
			SourceDocumentType *string `json:"source_document_type" validate:"omitempty,oneof=grn issue transfer adjustment disposal return temperature_excursion stock_count audit inspection complaint other"`
			SourceDocumentID   *string `json:"source_document_id"`
			ProductID          *string `json:"product_id" validate:"opt_uuid7"`
			BatchID            *string `json:"batch_id" validate:"opt_uuid7"`
			WarehouseID        *string `json:"warehouse_id" validate:"opt_uuid7"`
			LocationID         *string `json:"location_id" validate:"opt_uuid7"`
			Description        string  `json:"description" validate:"required"`
			ContainmentAction  *string `json:"containment_action"`
			ImpactAssessment   *string `json:"impact_assessment"`
			RootCause          *string `json:"root_cause"`
			RiskRating         string  `json:"risk_rating" validate:"required,oneof=low medium high critical"`
			Disposition        *string `json:"disposition" validate:"omitempty,oneof=accepted rejected quarantined destroyed returned reworked blocked other"`
			Status             string  `json:"status" validate:"omitempty,oneof=open under_investigation contained resolved closed"`
			ResponsibleOwner   *string `json:"responsible_owner"`
			DueDate            *string `json:"due_date" validate:"opt_date"`
			Escalated          bool    `json:"escalated"`
			EscalatedTo        *string `json:"escalated_to"`
			CreatedBy          string  `json:"created_by" validate:"required,max=255"`
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
		occurredAt := req.OccurredAt
		if occurredAt == "" {
			occurredAt = time.Now().UTC().Format(time.RFC3339)
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO deviations (org_id, deviation_number, category_code, occurred_at, reported_at, source_document_type, source_document_id, product_id, batch_id, warehouse_id, location_id, description, containment_action, impact_assessment, root_cause, risk_rating, disposition, status, responsible_owner, due_date, escalated, escalated_to, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,now(),$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$22) RETURNING id`,
			req.OrgID, req.DeviationNumber, nullIfEmptyPtr(req.CategoryCode), occurredAt,
			nullIfEmptyPtr(req.SourceDocumentType), nullIfEmptyPtr(req.SourceDocumentID),
			nullIfEmptyPtr(req.ProductID), nullIfEmptyPtr(req.BatchID), nullIfEmptyPtr(req.WarehouseID), nullIfEmptyPtr(req.LocationID),
			req.Description, nullIfEmptyPtr(req.ContainmentAction), nullIfEmptyPtr(req.ImpactAssessment), nullIfEmptyPtr(req.RootCause),
			req.RiskRating, nullIfEmptyPtr(req.Disposition), status, nullIfEmptyPtr(req.ResponsibleOwner), nullIfEmptyPtr(req.DueDate),
			req.Escalated, nullIfEmptyPtr(req.EscalatedTo), req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create deviation: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func ListDeviations(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM deviations`)
		if err != nil {
			response.InternalError(c, "failed to count deviations")
			return
		}

		var deviations []model.Deviation
		err = db.Select(&deviations,
			`SELECT * FROM deviations ORDER BY occurred_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list deviations")
			return
		}

		if deviations == nil {
			deviations = []model.Deviation{}
		}

		response.Paginated(c, deviations, total, p.Page, p.PageSize)
	}
}

func GetDeviation(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var deviation model.Deviation
		err := db.Get(&deviation, `SELECT * FROM deviations WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "deviation not found")
			return
		}

		response.OK(c, deviation)
	}
}

func CreateCAPAAction(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID            string  `json:"org_id" validate:"required,uuid7"`
			CAPANumber       string  `json:"capa_number" validate:"required,max=100"`
			SourceType       string  `json:"source_type" validate:"required,oneof=deviation complaint recall audit inspection stock_discrepancy supplier trend other"`
			SourceID         *string `json:"source_id"`
			CAPAType         string  `json:"capa_type" validate:"required,oneof=corrective preventive"`
			Title            string  `json:"title" validate:"required,max=500"`
			RootCause        *string `json:"root_cause"`
			ActionPlan       string  `json:"action_plan" validate:"required"`
			ResponsibleOwner *string `json:"responsible_owner"`
			DueDate          *string `json:"due_date" validate:"opt_date"`
			Status           string  `json:"status" validate:"omitempty,oneof=open in_progress overdue completed closed"`
			CreatedBy        string  `json:"created_by" validate:"required,max=255"`
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

		var id string
		err := db.QueryRowx(
			`INSERT INTO capa_actions (org_id, capa_number, source_type, source_id, capa_type, title, root_cause, action_plan, responsible_owner, due_date, status, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12) RETURNING id`,
			req.OrgID, req.CAPANumber, req.SourceType, nullIfEmptyPtr(req.SourceID), req.CAPAType, req.Title,
			nullIfEmptyPtr(req.RootCause), req.ActionPlan, nullIfEmptyPtr(req.ResponsibleOwner), nullIfEmptyPtr(req.DueDate),
			status, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create CAPA action: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func ListCAPAActions(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM capa_actions`)
		if err != nil {
			response.InternalError(c, "failed to count CAPA actions")
			return
		}

		var actions []model.CAPAAction
		err = db.Select(&actions,
			`SELECT * FROM capa_actions ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list CAPA actions")
			return
		}

		if actions == nil {
			actions = []model.CAPAAction{}
		}

		response.Paginated(c, actions, total, p.Page, p.PageSize)
	}
}

func GetCAPAAction(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var action model.CAPAAction
		err := db.Get(&action, `SELECT * FROM capa_actions WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "CAPA action not found")
			return
		}

		response.OK(c, action)
	}
}

// VerifyCAPAEffectiveness records the effectiveness check and closes the CAPA.
func VerifyCAPAEffectiveness(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			EffectivenessCheck string `json:"effectiveness_check" validate:"required"`
			VerifiedBy         string `json:"verified_by" validate:"required,max=255"`
			ClosedBy           string `json:"closed_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		_, err := db.Exec(
			`UPDATE capa_actions SET effectiveness_check=$1, effectiveness_verified=true, effectiveness_verified_by=$2, verified_at=now(), status='closed', closed_by=$3, closed_at=now(), updated_by=$3, updated_at=now() WHERE id=$4`,
			req.EffectivenessCheck, req.VerifiedBy, req.ClosedBy, c.Param("id"),
		)
		if err != nil {
			response.InternalError(c, "failed to verify CAPA effectiveness: "+err.Error())
			return
		}

		response.OK(c, gin.H{"status": "closed"})
	}
}

func CreateTrainingRecord(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID                string  `json:"org_id" validate:"required,uuid7"`
			UserID               *string `json:"user_id" validate:"opt_uuid7"`
			TrainingTitle        string  `json:"training_title" validate:"required,max=255"`
			TrainingType         string  `json:"training_type" validate:"required,oneof=initial refresher emergency contractor_safety"`
			Topic                *string `json:"topic" validate:"omitempty,oneof=sop product_identification fefo hygiene security data_integrity cold_chain controlled_products hazardous_materials emergency falsified_products recall complaints incident_reporting other"`
			TrainingDate         string  `json:"training_date" validate:"required,date"`
			Method               *string `json:"method" validate:"omitempty,oneof=classroom on_the_job online assessment other"`
			Trainer              *string `json:"trainer"`
			Assessed             bool    `json:"assessed"`
			AssessmentResult     *string `json:"assessment_result" validate:"omitempty,oneof=pass fail"`
			CompetencyValidUntil *string `json:"competency_valid_until" validate:"opt_date"`
			CertificateURL       *string `json:"certificate_url"`
			CreatedBy            string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		var id string
		err := db.QueryRowx(
			`INSERT INTO training_records (org_id, user_id, training_title, training_type, topic, training_date, method, trainer, assessed, assessment_result, competency_valid_until, certificate_url, created_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`,
			req.OrgID, nullIfEmptyPtr(req.UserID), req.TrainingTitle, req.TrainingType, nullIfEmptyPtr(req.Topic),
			req.TrainingDate, nullIfEmptyPtr(req.Method), nullIfEmptyPtr(req.Trainer), req.Assessed,
			nullIfEmptyPtr(req.AssessmentResult), nullIfEmptyPtr(req.CompetencyValidUntil), nullIfEmptyPtr(req.CertificateURL),
			req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create training record: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func ListTrainingRecords(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM training_records`)
		if err != nil {
			response.InternalError(c, "failed to count training records")
			return
		}

		var records []model.TrainingRecord
		err = db.Select(&records,
			`SELECT * FROM training_records ORDER BY training_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list training records")
			return
		}

		if records == nil {
			records = []model.TrainingRecord{}
		}

		response.Paginated(c, records, total, p.Page, p.PageSize)
	}
}
