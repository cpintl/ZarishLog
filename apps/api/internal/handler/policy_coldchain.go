package handler

import (
	"time"

	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
)

// Policy §13 — Cold Chain & Environmental Control

func CreateTemperatureMonitoringEntry(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID        string   `json:"org_id" validate:"required,uuid7"`
			WarehouseID  string   `json:"warehouse_id" validate:"required,uuid7"`
			LocationID   *string  `json:"location_id" validate:"opt_uuid7"`
			EquipmentID  *string  `json:"equipment_id" validate:"opt_uuid7"`
			DeviceName   *string  `json:"device_name"`
			MonitorType  string   `json:"monitor_type" validate:"required,oneof=temperature humidity freeze_indicator vvm"`
			ReadingValue float64  `json:"reading_value" validate:"required"`
			MinThreshold *float64 `json:"min_threshold"`
			MaxThreshold *float64 `json:"max_threshold"`
			RecordedAt   string   `json:"recorded_at"`
			RecordedBy   *string  `json:"recorded_by"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		recordedAt := req.RecordedAt
		if recordedAt == "" {
			recordedAt = time.Now().UTC().Format(time.RFC3339)
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO temperature_monitoring_entries (org_id, warehouse_id, location_id, equipment_id, device_name, monitor_type, reading_value, min_threshold, max_threshold, recorded_at, recorded_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`,
			req.OrgID, req.WarehouseID, nullIfEmptyPtr(req.LocationID), nullIfEmptyPtr(req.EquipmentID),
			nullIfEmptyPtr(req.DeviceName), req.MonitorType, req.ReadingValue, req.MinThreshold, req.MaxThreshold,
			recordedAt, nullIfEmptyPtr(req.RecordedBy),
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create monitoring entry: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func ListTemperatureMonitoringEntries(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM temperature_monitoring_entries`)
		if err != nil {
			response.InternalError(c, "failed to count monitoring entries")
			return
		}

		var entries []model.TemperatureMonitoringEntry
		err = db.Select(&entries,
			`SELECT * FROM temperature_monitoring_entries ORDER BY recorded_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list monitoring entries")
			return
		}

		if entries == nil {
			entries = []model.TemperatureMonitoringEntry{}
		}

		response.Paginated(c, entries, total, p.Page, p.PageSize)
	}
}

func CreateTemperatureExcursion(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID               string   `json:"org_id" validate:"required,uuid7"`
			WarehouseID         string   `json:"warehouse_id" validate:"required,uuid7"`
			LocationID          *string  `json:"location_id" validate:"opt_uuid7"`
			EquipmentID         *string  `json:"equipment_id" validate:"opt_uuid7"`
			DeviceName          *string  `json:"device_name"`
			ExcursionType       string   `json:"excursion_type" validate:"required,oneof=temperature humidity freeze power_failure equipment_failure monitoring_gap"`
			StartedAt           string   `json:"started_at"`
			EndedAt             *string  `json:"ended_at"`
			MinValue            *float64 `json:"min_value"`
			MaxValue            *float64 `json:"max_value"`
			ExpectedMin         *float64 `json:"expected_min"`
			ExpectedMax         *float64 `json:"expected_max"`
			AffectedProducts    *string  `json:"affected_products"`
			AffectedBatches     *string  `json:"affected_batches"`
			QuantityAffected    *float64 `json:"quantity_affected"`
			Quarantined         bool     `json:"quarantined"`
			QuarantineReference *string  `json:"quarantine_reference"`
			Disposition         string   `json:"disposition" validate:"omitempty,oneof=pending released restricted_use returned destroyed investigating"`
			NotifiedQuality     bool     `json:"notified_quality"`
			NotifiedManager     bool     `json:"notified_manager"`
			CreatedBy           string   `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		startedAt := req.StartedAt
		if startedAt == "" {
			startedAt = time.Now().UTC().Format(time.RFC3339)
		}
		disposition := req.Disposition
		if disposition == "" {
			disposition = "pending"
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO temperature_excursions (org_id, warehouse_id, location_id, equipment_id, device_name, excursion_type, started_at, ended_at, min_value, max_value, expected_min, expected_max, affected_products, affected_batches, quantity_affected, quarantined, quarantine_reference, disposition, notified_quality, notified_manager, status, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,'open',$21,$21) RETURNING id`,
			req.OrgID, req.WarehouseID, nullIfEmptyPtr(req.LocationID), nullIfEmptyPtr(req.EquipmentID),
			nullIfEmptyPtr(req.DeviceName), req.ExcursionType, startedAt, nullIfEmptyPtr(req.EndedAt),
			req.MinValue, req.MaxValue, req.ExpectedMin, req.ExpectedMax,
			nullIfEmptyPtr(req.AffectedProducts), nullIfEmptyPtr(req.AffectedBatches), req.QuantityAffected,
			req.Quarantined, nullIfEmptyPtr(req.QuarantineReference), disposition,
			req.NotifiedQuality, req.NotifiedManager, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create temperature excursion: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func GetTemperatureExcursion(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var excursion model.TemperatureExcursion
		err := db.Get(&excursion, `SELECT * FROM temperature_excursions WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "temperature excursion not found")
			return
		}

		response.OK(c, excursion)
	}
}

func ListTemperatureExcursions(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM temperature_excursions`)
		if err != nil {
			response.InternalError(c, "failed to count temperature excursions")
			return
		}

		var excursions []model.TemperatureExcursion
		err = db.Select(&excursions,
			`SELECT * FROM temperature_excursions ORDER BY started_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list temperature excursions")
			return
		}

		if excursions == nil {
			excursions = []model.TemperatureExcursion{}
		}

		response.Paginated(c, excursions, total, p.Page, p.PageSize)
	}
}

// UpdateTemperatureExcursionDisposition records the documented disposition
// (§13.3): release / restricted use / destruction / return, plus technical
// advice, root cause, linked CAPA, and end-of-excursion time.
func UpdateTemperatureExcursionDisposition(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			Disposition     string  `json:"disposition" validate:"required,oneof=pending released restricted_use returned destroyed investigating"`
			Status          string  `json:"status" validate:"omitempty,oneof=open under_review dispositioned closed"`
			TechnicalAdvice *string `json:"technical_advice"`
			RootCause       *string `json:"root_cause"`
			CapaID          *string `json:"capa_id" validate:"opt_uuid7"`
			EndedAt         *string `json:"ended_at"`
			UpdatedBy       string  `json:"updated_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		status := req.Status
		if status == "" {
			status = "dispositioned"
		}

		_, err := db.Exec(
			`UPDATE temperature_excursions SET disposition=$1, status=$2, technical_advice=$3, root_cause=$4, capa_id=$5, ended_at=$6, updated_by=$7, updated_at=now() WHERE id=$8`,
			req.Disposition, status, nullIfEmptyPtr(req.TechnicalAdvice), nullIfEmptyPtr(req.RootCause),
			nullIfEmptyPtr(req.CapaID), nullIfEmptyPtr(req.EndedAt), req.UpdatedBy, c.Param("id"),
		)
		if err != nil {
			response.InternalError(c, "failed to update excursion disposition: "+err.Error())
			return
		}

		response.OK(c, gin.H{"status": status})
	}
}
