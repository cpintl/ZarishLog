package model

import "time"

type Warehouse struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	Name      string    `json:"name" db:"name" validate:"required,max=255"`
	Code      string    `json:"code" db:"code" validate:"required,max=50"`
	Type      string    `json:"type" db:"type" validate:"required,wh_type"`
	Address   string    `json:"address" db:"address"`
	City      string    `json:"city" db:"city" validate:"required,max=100"`
	Country   string    `json:"country" db:"country" validate:"required,max=100"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	UpdatedBy string    `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Location struct {
	ID          string     `json:"id" db:"id"`
	WarehouseID string     `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	ParentID    *string    `json:"parent_id" db:"parent_id" validate:"omitempty,uuid7"`
	Code        string     `json:"code" db:"code" validate:"required,max=50"`
	Name        string     `json:"name" db:"name" validate:"required,max=255"`
	Type        string     `json:"type" db:"type" validate:"required,loc_type"`
	IsColdChain bool       `json:"is_cold_chain" db:"is_cold_chain"`
	IsHazardous bool       `json:"is_hazardous" db:"is_hazardous"`
	IsSecure    bool       `json:"is_secure" db:"is_secure"`
	MaxCapacity *float64   `json:"max_capacity" db:"max_capacity" validate:"omitempty,min=0"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type LocationConstraint struct {
	LocationID        string   `json:"location_id" db:"location_id" validate:"required,uuid7"`
	MinTemperature    *float64 `json:"min_temperature" db:"min_temperature"`
	MaxTemperature    *float64 `json:"max_temperature" db:"max_temperature"`
	MinHumidity       *float64 `json:"min_humidity" db:"min_humidity"`
	MaxHumidity       *float64 `json:"max_humidity" db:"max_humidity"`
	IsHazardousAllowed bool    `json:"is_hazardous_allowed" db:"is_hazardous_allowed"`
	IsFoodGrade       bool     `json:"is_food_grade" db:"is_food_grade"`
	IsPharmaGrade     bool     `json:"is_pharma_grade" db:"is_pharma_grade"`
	MaxWeightCapacity *float64 `json:"max_weight_capacity" db:"max_weight_capacity" validate:"omitempty,min=0"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
