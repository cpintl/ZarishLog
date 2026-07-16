package handler

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func CreateAsset(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrgID              string   `json:"org_id" validate:"required,uuid7"`
			AssetTag           string   `json:"asset_tag" validate:"required,max=100"`
			Name               string   `json:"name" validate:"required,max=255"`
			Description        string   `json:"description"`
			ProductID          *string  `json:"product_id" validate:"omitempty,uuid7"`
			SerialNumber       string   `json:"serial_number"`
			CustodianID        *string  `json:"custodian_id" validate:"omitempty,uuid7"`
			LocationID         *string  `json:"location_id" validate:"omitempty,uuid7"`
			AcquisitionDate    *string  `json:"acquisition_date" validate:"omitempty,date"`
			PurchaseCost       *float64 `json:"purchase_cost" validate:"omitempty,min=0"`
			CurrentValue       *float64 `json:"current_value" validate:"omitempty,min=0"`
			DepreciationMethod string   `json:"depreciation_method"`
			UsefulLifeYears    *int     `json:"useful_life_years" validate:"omitempty,min=1"`
			Status             string   `json:"status" validate:"required,oneof=in_use in_storage under_maintenance disposed lost"`
			CreatedBy          string   `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO assets (org_id, asset_tag, name, description, product_id, serial_number,
			 custodian_id, location_id, acquisition_date, purchase_cost, current_value,
			 depreciation_method, useful_life_years, status, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$15) RETURNING id`,
			req.OrgID, req.AssetTag, req.Name, req.Description,
			nullIfEmptyPtr(req.ProductID), req.SerialNumber,
			nullIfEmptyPtr(req.CustodianID), nullIfEmptyPtr(req.LocationID),
			nullIfEmptyPtr(req.AcquisitionDate), req.PurchaseCost, req.CurrentValue,
			req.DepreciationMethod, req.UsefulLifeYears, req.Status, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create asset: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func ListAssets(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM assets`)
		if err != nil {
			response.InternalError(c, "failed to count assets")
			return
		}

		var assets []model.Asset
		err = db.Select(&assets,
			`SELECT id, org_id, asset_tag, name, description, product_id, serial_number,
			 custodian_id, location_id, acquisition_date, purchase_cost, current_value,
			 depreciation_method, useful_life_years, status, created_by, updated_by, created_at, updated_at
			 FROM assets ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list assets")
			return
		}

		if assets == nil {
			assets = []model.Asset{}
		}

		response.Paginated(c, assets, total, p.Page, p.PageSize)
	}
}

func GetAsset(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var asset model.Asset
		err := db.Get(&asset,
			`SELECT id, org_id, asset_tag, name, description, product_id, serial_number,
			 custodian_id, location_id, acquisition_date, purchase_cost, current_value,
			 depreciation_method, useful_life_years, status, created_by, updated_by, created_at, updated_at
			 FROM assets WHERE id=$1`, id,
		)
		if err != nil {
			response.NotFound(c, "asset not found")
			return
		}

		var custody []model.AssetCustodyChange
		_ = db.Select(&custody,
			`SELECT id, asset_id, from_user_id, to_user_id, changed_by, changed_at
			 FROM asset_custody_changes WHERE asset_id=$1 ORDER BY changed_at DESC`, id,
		)
		if custody == nil {
			custody = []model.AssetCustodyChange{}
		}

		var maintenance []model.AssetMaintenance
		_ = db.Select(&maintenance,
			`SELECT id, asset_id, maintenance_date, description, cost, performed_by, next_date, created_at
			 FROM asset_maintenance WHERE asset_id=$1 ORDER BY maintenance_date DESC`, id,
		)
		if maintenance == nil {
			maintenance = []model.AssetMaintenance{}
		}

		response.OK(c, gin.H{
			"asset":       asset,
			"custody":     custody,
			"maintenance": maintenance,
		})
	}
}

func UpdateAsset(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req struct {
			AssetTag           string   `json:"asset_tag" validate:"required,max=100"`
			Name               string   `json:"name" validate:"required,max=255"`
			Description        string   `json:"description"`
			ProductID          *string  `json:"product_id" validate:"omitempty,uuid7"`
			SerialNumber       string   `json:"serial_number"`
			CustodianID        *string  `json:"custodian_id" validate:"omitempty,uuid7"`
			LocationID         *string  `json:"location_id" validate:"omitempty,uuid7"`
			AcquisitionDate    *string  `json:"acquisition_date" validate:"omitempty,date"`
			PurchaseCost       *float64 `json:"purchase_cost" validate:"omitempty,min=0"`
			CurrentValue       *float64 `json:"current_value" validate:"omitempty,min=0"`
			DepreciationMethod string   `json:"depreciation_method"`
			UsefulLifeYears    *int     `json:"useful_life_years" validate:"omitempty,min=1"`
			Status             string   `json:"status" validate:"required,oneof=in_use in_storage under_maintenance disposed lost"`
			UpdatedBy          string   `json:"updated_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		result, err := db.Exec(
			`UPDATE assets SET asset_tag=$1, name=$2, description=$3, product_id=$4, serial_number=$5,
			 custodian_id=$6, location_id=$7, acquisition_date=$8, purchase_cost=$9, current_value=$10,
			 depreciation_method=$11, useful_life_years=$12, status=$13, updated_by=$14, updated_at=now()
			 WHERE id=$15`,
			req.AssetTag, req.Name, req.Description,
			nullIfEmptyPtr(req.ProductID), req.SerialNumber,
			nullIfEmptyPtr(req.CustodianID), nullIfEmptyPtr(req.LocationID),
			nullIfEmptyPtr(req.AcquisitionDate), req.PurchaseCost, req.CurrentValue,
			req.DepreciationMethod, req.UsefulLifeYears, req.Status, req.UpdatedBy, id,
		)
		if err != nil {
			response.InternalError(c, "failed to update asset: "+err.Error())
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			response.NotFound(c, "asset not found")
			return
		}

		response.OK(c, gin.H{"id": id})
	}
}

func DeleteAsset(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := db.Exec(`DELETE FROM assets WHERE id=$1`, id)
		if err != nil {
			response.InternalError(c, "failed to delete asset")
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			response.NotFound(c, "asset not found")
			return
		}

		response.OK(c, gin.H{"id": id})
	}
}

func TransferCustody(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		assetID := c.Param("id")

		var req struct {
			ToUserID string `json:"to_user_id" validate:"required,uuid7"`
			ChangedBy string `json:"changed_by" validate:"required,max=255"`
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

		var currentCustodian *string
		_ = tx.Get(&currentCustodian, `SELECT custodian_id FROM assets WHERE id=$1`, assetID)

		var id string
		err = tx.QueryRowx(
			`INSERT INTO asset_custody_changes (asset_id, from_user_id, to_user_id, changed_by)
			 VALUES ($1,$2,$3,$4) RETURNING id`,
			assetID, currentCustodian, req.ToUserID, req.ChangedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to record custody change: "+err.Error())
			return
		}

		_, err = tx.Exec(`UPDATE assets SET custodian_id=$1, updated_by=$2, updated_at=now() WHERE id=$3`,
			req.ToUserID, req.ChangedBy, assetID,
		)
		if err != nil {
			response.InternalError(c, "failed to update asset custodian")
			return
		}

		if err := tx.Commit(); err != nil {
			response.InternalError(c, "failed to commit")
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func CreateAssetMaintenance(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		assetID := c.Param("id")

		var req struct {
			MaintenanceDate string   `json:"maintenance_date" validate:"required,date"`
			Description     string   `json:"description"`
			Cost            *float64 `json:"cost" validate:"omitempty,min=0"`
			PerformedBy     string   `json:"performed_by"`
			NextDate        *string  `json:"next_date" validate:"omitempty,date"`
			CreatedBy       string   `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO asset_maintenance (asset_id, maintenance_date, description, cost, performed_by, next_date)
			 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
			assetID, req.MaintenanceDate, req.Description, req.Cost, req.PerformedBy,
			nullIfEmptyPtr(req.NextDate),
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create maintenance record: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func ListAssetMaintenance(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		assetID := c.Param("id")
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM asset_maintenance WHERE asset_id=$1`, assetID)
		if err != nil {
			response.InternalError(c, "failed to count maintenance records")
			return
		}

		var records []model.AssetMaintenance
		err = db.Select(&records,
			`SELECT id, asset_id, maintenance_date, description, cost, performed_by, next_date, created_at
			 FROM asset_maintenance WHERE asset_id=$1 ORDER BY maintenance_date DESC LIMIT $2 OFFSET $3`,
			assetID, p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list maintenance records")
			return
		}

		if records == nil {
			records = []model.AssetMaintenance{}
		}

		response.Paginated(c, records, total, p.Page, p.PageSize)
	}
}
