package model

import "time"

type Product struct {
	ID              string    `json:"id" db:"id"`
	OrgID           string    `json:"org_id" db:"org_id"`
	CategoryID      string    `json:"category_id" db:"category_id"`
	UomID           string    `json:"uom_id" db:"uom_id"`
	SKU             string    `json:"sku" db:"sku"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	ItemType        string    `json:"item_type" db:"item_type"`
	GTIN            string    `json:"gtin" db:"gtin"`
	AlternativeCode string    `json:"alternative_code" db:"alternative_code"`
	Brand           string    `json:"brand" db:"brand"`
	Manufacturer    string    `json:"manufacturer" db:"manufacturer"`
	IsBatchTracked  bool      `json:"is_batch_tracked" db:"is_batch_tracked"`
	IsSerialTracked bool      `json:"is_serial_tracked" db:"is_serial_tracked"`
	IsExpiryTracked bool      `json:"is_expiry_tracked" db:"is_expiry_tracked"`
	IsHazardous     bool      `json:"is_hazardous" db:"is_hazardous"`
	IsColdChain     bool      `json:"is_cold_chain" db:"is_cold_chain"`
	MinStock        float64   `json:"min_stock" db:"min_stock"`
	MaxStock        float64   `json:"max_stock" db:"max_stock"`
	ReorderPoint    float64   `json:"reorder_point" db:"reorder_point"`
	LeadTimeDays    int       `json:"lead_time_days" db:"lead_time_days"`
	UnitCost        float64   `json:"unit_cost" db:"unit_cost"`
	SafetyStock     float64   `json:"safety_stock" db:"safety_stock"`
	Status          string    `json:"status" db:"status"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	UpdatedBy       string    `json:"updated_by" db:"updated_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type ProductCategory struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	ParentID    string    `json:"parent_id" db:"parent_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	UNSPSC      string    `json:"unspsc" db:"unspsc"`
	ECLASS      string    `json:"eclass" db:"eclass"`
	Status      string    `json:"status" db:"status"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type UoM struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Abbreviation string `json:"abbreviation" db:"abbreviation"`
	Category    string `json:"category" db:"category"`
}
