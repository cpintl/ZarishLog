package model

import "time"

type Distribution struct {
	ID                 string    `json:"id" db:"id"`
	OrgID              string    `json:"org_id" db:"org_id"`
	OrgLevelID         *string   `json:"org_level_id" db:"org_level_id" validate:"omitempty,uuid7"`
	ProgramID          *string   `json:"program_id" db:"program_id" validate:"omitempty,uuid7"`
	DistributionNumber string    `json:"distribution_number" db:"distribution_number" validate:"required,max=100"`
	DistributionDate   string    `json:"distribution_date" db:"distribution_date" validate:"required,date"`
	Location           string    `json:"location" db:"location"`
	BeneficiaryCount   *int      `json:"beneficiary_count" db:"beneficiary_count" validate:"omitempty,min=0"`
	Status             string    `json:"status" db:"status" validate:"omitempty,oneof=draft active completed cancelled"`
	Notes              string    `json:"notes" db:"notes"`
	CreatedBy          string    `json:"created_by" db:"created_by"`
	UpdatedBy          string    `json:"updated_by" db:"updated_by"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type DistributionLineItem struct {
	ID                 string   `json:"id" db:"id"`
	DistributionID     string   `json:"distribution_id" db:"distribution_id" validate:"required,uuid7"`
	ProductID          string   `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchID            *string  `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	QuantityPlanned    float64  `json:"quantity_planned" db:"quantity_planned" validate:"required,min=0"`
	QuantityDistributed float64 `json:"quantity_distributed" db:"quantity_distributed" validate:"min=0"`
	UnitCost           *float64 `json:"unit_cost" db:"unit_cost" validate:"omitempty,min=0"`
	Status             string   `json:"status" db:"status" validate:"omitempty,oneof=active inactive"`
}

type DistributionBeneficiary struct {
	ID             string `json:"id" db:"id"`
	DistributionID string `json:"distribution_id" db:"distribution_id" validate:"required,uuid7"`
	BeneficiaryType string `json:"beneficiary_type" db:"beneficiary_type" validate:"required,max=100"`
	Count          int    `json:"count" db:"count" validate:"required,min=1"`
	Criteria       string `json:"criteria" db:"criteria"`
}
