package model

import "time"

type QAInspection struct {
	ID             string     `json:"id" db:"id"`
	OrgID          string     `json:"org_id" db:"org_id"`
	GRNID          *string    `json:"grn_id" db:"grn_id" validate:"omitempty,uuid7"`
	ProductID      string     `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchID        *string    `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	InspectionDate string     `json:"inspection_date" db:"inspection_date" validate:"required,date"`
	Inspector      string     `json:"inspector" db:"inspector" validate:"required,max=255"`
	Result         string     `json:"result" db:"result" validate:"required,oneof=pass fail quarantine"`
	Notes          string     `json:"notes" db:"notes"`
	Disposition    *string    `json:"disposition" db:"disposition"`
	CreatedBy      string     `json:"created_by" db:"created_by"`
	UpdatedBy      string     `json:"updated_by" db:"updated_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type QAChecklistTemplate struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	Code        string    `json:"code" db:"code" validate:"required,max=50"`
	Name        string    `json:"name" db:"name" validate:"required,max=255"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category" validate:"required,max=100"`
	IsMandatory bool      `json:"is_mandatory" db:"is_mandatory"`
	Status      string    `json:"status" db:"status" validate:"omitempty,oneof=active inactive draft archived"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type QAChecklistItem struct {
	ID              string  `json:"id" db:"id"`
	TemplateID      string  `json:"template_id" db:"template_id" validate:"required,uuid7"`
	ItemOrder       int     `json:"item_order" db:"item_order" validate:"min=1"`
	Question        string  `json:"question" db:"question" validate:"required"`
	ExpectedAnswer  string  `json:"expected_answer" db:"expected_answer" validate:"required"`
	IsCritical      bool    `json:"is_critical" db:"is_critical"`
	Weight          float64 `json:"weight" db:"weight" validate:"min=0,max=5"`
}

type QAChecklistResult struct {
	ID              string    `json:"id" db:"id"`
	InspectionID    string    `json:"inspection_id" db:"inspection_id" validate:"required,uuid7"`
	ChecklistItemID string    `json:"checklist_item_id" db:"checklist_item_id" validate:"required,uuid7"`
	Answer          string    `json:"answer" db:"answer"`
	Score           *float64  `json:"score" db:"score" validate:"omitempty,min=0,max=100"`
	Notes           string    `json:"notes" db:"notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type QADisposition struct {
	ID                      string    `json:"id" db:"id"`
	InspectionID            string    `json:"inspection_id" db:"inspection_id" validate:"required,uuid7"`
	DispositionType         string    `json:"disposition_type" db:"disposition_type" validate:"required,oneof=pass fail quarantine rework partial"`
	DispositionDate         string    `json:"disposition_date" db:"disposition_date" validate:"required,date"`
	ApprovedBy              string    `json:"approved_by" db:"approved_by"`
	DestinationLocationID   *string   `json:"destination_location_id" db:"destination_location_id" validate:"omitempty,uuid7"`
	Notes                   string    `json:"notes" db:"notes"`
	CreatedBy               string    `json:"created_by" db:"created_by"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}
