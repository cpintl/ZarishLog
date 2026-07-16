package model

import (
	"time"
)

type Product struct {
	ID              string      `json:"id" db:"id"`
	OrgID           string      `json:"org_id" db:"org_id"`
	CategoryID      string      `json:"category_id" db:"category_id" validate:"omitempty,uuid7"`
	UomID           string      `json:"uom_id" db:"uom_id" validate:"omitempty,uuid7"`
	SKU             string      `json:"sku" db:"sku" validate:"required,max=100"`
	Name            string      `json:"name" db:"name" validate:"required,max=255"`
	Description     string      `json:"description" db:"description"`
	ItemType        string      `json:"item_type" db:"item_type" validate:"required,item_type"`
	GTIN            string      `json:"gtin" db:"gtin"`
	AlternativeCode string      `json:"alternative_code" db:"alternative_code"`
	Brand           string      `json:"brand" db:"brand"`
	Manufacturer    string      `json:"manufacturer" db:"manufacturer"`
	Strength        string      `json:"strength" db:"strength"`
	UNSPSC          string      `json:"unspsc_commodity" db:"unspsc_commodity"`
	EClass          string      `json:"eclass_code" db:"eclass_code"`
	AlternateCodes  string      `json:"alternate_codes" db:"alternate_codes"`
	IsAsset         bool        `json:"is_asset" db:"is_asset"`
	ReplenishType   string      `json:"replenishment_type" db:"replenishment_type"`
	ValuationMethod string      `json:"valuation_method" db:"valuation_method"`
	IsKitting       bool        `json:"is_kitting" db:"is_kitting"`
	TempMinC        *float64    `json:"temp_min_c" db:"temp_min_c"`
	TempMaxC        *float64    `json:"temp_max_c" db:"temp_max_c"`
	IsEssential     bool        `json:"is_essential" db:"is_essential"`
	IsControlled    bool        `json:"is_controlled" db:"is_controlled"`
	IsBatchTracked  bool        `json:"is_batch_tracked" db:"is_batch_tracked"`
	IsSerialTracked bool        `json:"is_serial_tracked" db:"is_serial_tracked"`
	IsExpiryTracked bool        `json:"is_expiry_tracked" db:"is_expiry_tracked"`
	IsHazardous     bool        `json:"is_hazardous" db:"is_hazardous"`
	IsColdChain     bool        `json:"is_cold_chain" db:"is_cold_chain"`
	DosageFormCode  string      `json:"dosage_form_code" db:"dosage_form_code"`
	GenericName     string      `json:"generic_name" db:"generic_name"`
	BrandName       string      `json:"brand_name" db:"brand_name"`
	StorageCond     string      `json:"storage_conditions" db:"storage_conditions"`
	ReferenceURLs   string      `json:"reference_urls" db:"reference_urls"`
	MinStock        float64     `json:"min_stock" db:"min_stock" validate:"omitempty,min=0"`
	MaxStock        float64     `json:"max_stock" db:"max_stock" validate:"omitempty,min=0"`
	ReorderPoint    float64     `json:"reorder_point" db:"reorder_point" validate:"omitempty,min=0"`
	LeadTimeDays    int         `json:"lead_time_days" db:"lead_time_days" validate:"omitempty,min=0"`
	UnitCost        float64     `json:"unit_cost" db:"unit_cost" validate:"omitempty,min=0"`
	SafetyStock     float64     `json:"safety_stock" db:"safety_stock" validate:"omitempty,min=0"`
	Status          string      `json:"status" db:"status" validate:"omitempty,oneof=active inactive discontinued"`
	CreatedBy       string      `json:"created_by" db:"created_by"`
	UpdatedBy       string      `json:"updated_by" db:"updated_by"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}

type ProductCategory struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	ParentID    string    `json:"parent_id" db:"parent_id" validate:"omitempty,uuid7"`
	Name        string    `json:"name" db:"name" validate:"required,max=255"`
	Description string    `json:"description" db:"description"`
	UNSPSC      string    `json:"unspsc" db:"unspsc"`
	ECLASS      string    `json:"eclass" db:"eclass"`
	Status      string    `json:"status" db:"status" validate:"omitempty,oneof=active inactive"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type UoM struct {
	ID               string     `json:"id" db:"id"`
	Name             string     `json:"name" db:"name" validate:"required,max=100"`
	Abbreviation     string     `json:"abbreviation" db:"abbreviation" validate:"required,max=10"`
	Category         string     `json:"category" db:"category" validate:"required,uom_category"`
	BaseUomID        *string    `json:"base_uom_id" db:"base_uom_id" validate:"omitempty,uuid7"`
	ConversionFactor *float64   `json:"conversion_factor" db:"conversion_factor" validate:"omitempty,min=0"`
	Status           string     `json:"status" db:"status" validate:"omitempty,oneof=active inactive"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}
