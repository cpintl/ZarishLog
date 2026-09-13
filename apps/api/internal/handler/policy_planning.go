package handler

import (
	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
)

// Policy §20 — Emergency medical supply plans

func CreateEmergencyPlan(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID                  string  `json:"org_id" validate:"required,uuid7"`
			PlanNumber             string  `json:"plan_number" validate:"required,max=100"`
			Name                   string  `json:"name" validate:"required,max=255"`
			CountryCode            *string `json:"country_code"`
			FacilityID             *string `json:"facility_id" validate:"opt_uuid7"`
			ScenarioType           string  `json:"scenario_type" validate:"required,oneof=outbreak displacement cyclone flood fire security scale_up power_loss cold_chain_failure supplier_interruption border_delay system_outage warehouse_loss staff_shortage other"`
			Scenario               *string `json:"scenario"`
			DemandAssumptions      *string `json:"demand_assumptions"`
			ServiceLevel           *string `json:"service_level"`
			PrepositionedStockPlan *string `json:"prepositioned_stock_plan"`
			RotationPlan           *string `json:"rotation_plan"`
			AlternateSuppliers     *string `json:"alternate_suppliers"`
			AlternateWarehouses    *string `json:"alternate_warehouses"`
			AlternateRoutes        *string `json:"alternate_routes"`
			AlternatePowerSources  *string `json:"alternate_power_sources"`
			EmergencyAuthority     *string `json:"emergency_authority"`
			PaperFallbackRecords   bool    `json:"paper_fallback_records"`
			MinimumStaffing        *string `json:"minimum_staffing"`
			ColdChainContingency   *string `json:"cold_chain_contingency"`
			Communications         *string `json:"communications"`
			EscalationContacts     *string `json:"escalation_contacts"`
			SecurityControls       *string `json:"security_controls"`
			ReturnToNormalPlan     *string `json:"return_to_normal_plan"`
			Status                 string  `json:"status" validate:"omitempty,oneof=draft approved active exercised superseded"`
			CreatedBy              string  `json:"created_by" validate:"required,max=255"`
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

		var id string
		err := db.QueryRowx(
			`INSERT INTO emergency_plans (org_id, plan_number, name, country_code, facility_id, scenario_type, scenario, demand_assumptions, service_level, prepositioned_stock_plan, rotation_plan, alternate_suppliers, alternate_warehouses, alternate_routes, alternate_power_sources, emergency_authority, paper_fallback_records, minimum_staffing, cold_chain_contingency, communications, escalation_contacts, security_controls, return_to_normal_plan, status, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$25) RETURNING id`,
			req.OrgID, req.PlanNumber, req.Name, nullIfEmptyPtr(req.CountryCode), nullIfEmptyPtr(req.FacilityID),
			req.ScenarioType, nullIfEmptyPtr(req.Scenario), nullIfEmptyPtr(req.DemandAssumptions), nullIfEmptyPtr(req.ServiceLevel),
			nullIfEmptyPtr(req.PrepositionedStockPlan), nullIfEmptyPtr(req.RotationPlan),
			nullIfEmptyPtr(req.AlternateSuppliers), nullIfEmptyPtr(req.AlternateWarehouses), nullIfEmptyPtr(req.AlternateRoutes), nullIfEmptyPtr(req.AlternatePowerSources),
			nullIfEmptyPtr(req.EmergencyAuthority), req.PaperFallbackRecords, nullIfEmptyPtr(req.MinimumStaffing),
			nullIfEmptyPtr(req.ColdChainContingency), nullIfEmptyPtr(req.Communications), nullIfEmptyPtr(req.EscalationContacts),
			nullIfEmptyPtr(req.SecurityControls), nullIfEmptyPtr(req.ReturnToNormalPlan),
			status, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create emergency plan: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func GetEmergencyPlan(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var plan model.EmergencyPlan
		err := db.Get(&plan, `SELECT * FROM emergency_plans WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "emergency plan not found")
			return
		}

		response.OK(c, plan)
	}
}

func ListEmergencyPlans(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM emergency_plans`)
		if err != nil {
			response.InternalError(c, "failed to count emergency plans")
			return
		}

		var plans []model.EmergencyPlan
		err = db.Select(&plans,
			`SELECT * FROM emergency_plans ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list emergency plans")
			return
		}

		if plans == nil {
			plans = []model.EmergencyPlan{}
		}

		response.Paginated(c, plans, total, p.Page, p.PageSize)
	}
}

// Policy §27 — Change control

func CreateChangeControl(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID                 string  `json:"org_id" validate:"required,uuid7"`
			ChangeNumber          string  `json:"change_number" validate:"required,max=100"`
			SubjectType           string  `json:"subject_type" validate:"required,oneof=product supplier warehouse_layout equipment software data_structure program service_level storage_condition legal_requirement sop"`
			SubjectID             *string `json:"subject_id"`
			Description           string  `json:"description" validate:"required"`
			ImpactAssessment      *string `json:"impact_assessment"`
			RiskLevel             string  `json:"risk_level" validate:"required,oneof=low medium high critical"`
			RequiresQualityReview bool    `json:"requires_quality_review"`
			AuthorizedBy          *string `json:"authorized_by"`
			AuthorizationDate     *string `json:"authorization_date" validate:"opt_date"`
			StartDate             *string `json:"start_date" validate:"opt_date"`
			EndDate               *string `json:"end_date" validate:"opt_date"`
			Status                string  `json:"status" validate:"omitempty,oneof=draft assessed approved implemented closed rejected"`
			CreatedBy             string  `json:"created_by" validate:"required,max=255"`
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

		var id string
		err := db.QueryRowx(
			`INSERT INTO change_controls (org_id, change_number, subject_type, subject_id, description, impact_assessment, risk_level, requires_quality_review, authorized_by, authorization_date, start_date, end_date, status, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$14) RETURNING id`,
			req.OrgID, req.ChangeNumber, req.SubjectType, nullIfEmptyPtr(req.SubjectID), req.Description,
			nullIfEmptyPtr(req.ImpactAssessment), req.RiskLevel, req.RequiresQualityReview,
			nullIfEmptyPtr(req.AuthorizedBy), nullIfEmptyPtr(req.AuthorizationDate),
			nullIfEmptyPtr(req.StartDate), nullIfEmptyPtr(req.EndDate),
			status, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create change control: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func GetChangeControl(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var change model.ChangeControl
		err := db.Get(&change, `SELECT * FROM change_controls WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "change control not found")
			return
		}

		response.OK(c, change)
	}
}

func ListChangeControls(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM change_controls`)
		if err != nil {
			response.InternalError(c, "failed to count change controls")
			return
		}

		var changes []model.ChangeControl
		err = db.Select(&changes,
			`SELECT * FROM change_controls ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list change controls")
			return
		}

		if changes == nil {
			changes = []model.ChangeControl{}
		}

		response.Paginated(c, changes, total, p.Page, p.PageSize)
	}
}
