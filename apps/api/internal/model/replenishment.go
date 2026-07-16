package model

import "time"

type AMCCalculation struct {
	ID                   string    `json:"id" db:"id"`
	OrgID                string    `json:"org_id" db:"org_id"`
	ProductID            string    `json:"product_id" db:"product_id" validate:"required,uuid7"`
	WarehouseID          string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	CalculationDate      string    `json:"calculation_date" db:"calculation_date"`
	AMC3Months           *float64  `json:"amc_3_months" db:"amc_3_months"`
	AMC6Months           *float64  `json:"amc_6_months" db:"amc_6_months"`
	AMC12Months          *float64  `json:"amc_12_months" db:"amc_12_months"`
	MaxConsumption       *float64  `json:"max_consumption" db:"max_consumption"`
	StdDeviation         *float64  `json:"std_deviation" db:"std_deviation"`
	PeriodStart          *string   `json:"calculation_period_start" db:"calculation_period_start"`
	PeriodEnd            *string   `json:"calculation_period_end" db:"calculation_period_end"`
	CalculationStatus    string    `json:"calculation_status" db:"calculation_status"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

type ReorderRecommendation struct {
	ID                 string    `json:"id" db:"id"`
	OrgID              string    `json:"org_id" db:"org_id"`
	ProductID          string    `json:"product_id" db:"product_id" validate:"required,uuid7"`
	WarehouseID        string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	RecommendationDate string    `json:"recommendation_date" db:"recommendation_date"`
	CurrentStock       float64   `json:"current_stock" db:"current_stock" validate:"min=0"`
	ReorderPoint       float64   `json:"reorder_point" db:"reorder_point" validate:"min=0"`
	ReorderQuantity    float64   `json:"reorder_quantity" db:"reorder_quantity" validate:"min=0"`
	LeadTimeDays       *int      `json:"lead_time_days" db:"lead_time_days" validate:"omitempty,min=0"`
	AMCUsed            *float64  `json:"amc_used" db:"amc_used"`
	SafetyStock        *float64  `json:"safety_stock" db:"safety_stock"`
	RecommendationType string    `json:"recommendation_type" db:"recommendation_type" validate:"omitempty,oneof=reorder excess normal critical"`
	Priority           string    `json:"priority" db:"priority" validate:"omitempty,oneof=low medium high critical"`
	Notes              string    `json:"notes" db:"notes"`
	Reviewed           bool      `json:"reviewed" db:"reviewed"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

type ForecastResult struct {
	ID              string    `json:"id" db:"id"`
	OrgID           string    `json:"org_id" db:"org_id"`
	ProductID       string    `json:"product_id" db:"product_id" validate:"required,uuid7"`
	WarehouseID     string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	ForecastDate    string    `json:"forecast_date" db:"forecast_date" validate:"required,date"`
	ForecastValue   float64   `json:"forecast_value" db:"forecast_value" validate:"required,min=0"`
	LowerBound      *float64  `json:"lower_bound" db:"lower_bound"`
	UpperBound      *float64  `json:"upper_bound" db:"upper_bound"`
	ConfidenceLevel float64   `json:"confidence_level" db:"confidence_level"`
	ModelVersion    string    `json:"model_version" db:"model_version"`
	FeaturesUsed    *string   `json:"features_used" db:"features_used"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
