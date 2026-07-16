package model

import "time"

type Asset struct {
	ID                 string    `json:"id" db:"id"`
	OrgID              string    `json:"org_id" db:"org_id"`
	AssetTag           string    `json:"asset_tag" db:"asset_tag" validate:"required,max=100"`
	Name               string    `json:"name" db:"name" validate:"required,max=255"`
	Description        string    `json:"description" db:"description"`
	ProductID          *string   `json:"product_id" db:"product_id" validate:"omitempty,uuid7"`
	SerialNumber       string    `json:"serial_number" db:"serial_number"`
	CustodianID        *string   `json:"custodian_id" db:"custodian_id" validate:"omitempty,uuid7"`
	LocationID         *string   `json:"location_id" db:"location_id" validate:"omitempty,uuid7"`
	AcquisitionDate    *string   `json:"acquisition_date" db:"acquisition_date" validate:"omitempty,date"`
	PurchaseCost       *float64  `json:"purchase_cost" db:"purchase_cost" validate:"omitempty,min=0"`
	CurrentValue       *float64  `json:"current_value" db:"current_value" validate:"omitempty,min=0"`
	DepreciationMethod string    `json:"depreciation_method" db:"depreciation_method"`
	UsefulLifeYears    *int      `json:"useful_life_years" db:"useful_life_years" validate:"omitempty,min=1"`
	Status             string    `json:"status" db:"status" validate:"required,oneof=in_use in_storage under_maintenance disposed lost"`
	CreatedBy          string    `json:"created_by" db:"created_by"`
	UpdatedBy          string    `json:"updated_by" db:"updated_by"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type AssetCustodyChange struct {
	ID         string    `json:"id" db:"id"`
	AssetID    string    `json:"asset_id" db:"asset_id" validate:"required,uuid7"`
	FromUserID *string   `json:"from_user_id" db:"from_user_id" validate:"omitempty,uuid7"`
	ToUserID   *string   `json:"to_user_id" db:"to_user_id" validate:"omitempty,uuid7"`
	ChangedBy  string    `json:"changed_by" db:"changed_by"`
	ChangedAt  time.Time `json:"changed_at" db:"changed_at"`
}

type AssetMaintenance struct {
	ID             string    `json:"id" db:"id"`
	AssetID        string    `json:"asset_id" db:"asset_id" validate:"required,uuid7"`
	MaintenanceDate string   `json:"maintenance_date" db:"maintenance_date" validate:"required,date"`
	Description    string    `json:"description" db:"description"`
	Cost           *float64  `json:"cost" db:"cost" validate:"omitempty,min=0"`
	PerformedBy    string    `json:"performed_by" db:"performed_by"`
	NextDate       *string   `json:"next_date" db:"next_date" validate:"omitempty,date"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
