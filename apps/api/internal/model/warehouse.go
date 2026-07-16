package model

import "time"

type Warehouse struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Type        string    `json:"type" db:"type"`
	Address     string    `json:"address" db:"address"`
	City        string    `json:"city" db:"city"`
	Country     string    `json:"country" db:"country"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Location struct {
	ID            string    `json:"id" db:"id"`
	WarehouseID   string    `json:"warehouse_id" db:"warehouse_id"`
	ParentID      string    `json:"parent_id" db:"parent_id"`
	Code          string    `json:"code" db:"code"`
	Name          string    `json:"name" db:"name"`
	Type          string    `json:"type" db:"type"`
	IsColdChain   bool      `json:"is_cold_chain" db:"is_cold_chain"`
	IsHazardous   bool      `json:"is_hazardous" db:"is_hazardous"`
	IsSecure      bool      `json:"is_secure" db:"is_secure"`
	MaxCapacity   float64   `json:"max_capacity" db:"max_capacity"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
