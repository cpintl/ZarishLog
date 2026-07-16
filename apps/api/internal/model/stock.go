package model

import "time"

type StockLevel struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	ProductID   string    `json:"product_id" db:"product_id"`
	WarehouseID string    `json:"warehouse_id" db:"warehouse_id"`
	LocationID  string    `json:"location_id" db:"location_id"`
	BatchID     string    `json:"batch_id" db:"batch_id"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	ReservedQty float64   `json:"reserved_qty" db:"reserved_qty"`
	Status      string    `json:"status" db:"status"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type StockMovement struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	ProductID   string    `json:"product_id" db:"product_id"`
	WarehouseID string    `json:"warehouse_id" db:"warehouse_id"`
	LocationID  string    `json:"location_id" db:"location_id"`
	BatchID     string    `json:"batch_id" db:"batch_id"`
	MovementType string   `json:"movement_type" db:"movement_type"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	RefDocType  string    `json:"ref_doc_type" db:"ref_doc_type"`
	RefDocID    string    `json:"ref_doc_id" db:"ref_doc_id"`
	ReasonCode  string    `json:"reason_code" db:"reason_code"`
	Reference   string    `json:"reference" db:"reference"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type GRN struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	WarehouseID string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	Supplier    string    `json:"supplier" db:"supplier" validate:"required,max=255"`
	PONumber    string    `json:"po_number" db:"po_number"`
	ReceivedBy  string    `json:"received_by" db:"received_by" validate:"required,max=255"`
	Status      string    `json:"status" db:"status" validate:"omitempty,oneof=pending completed cancelled"`
	Notes       string    `json:"notes" db:"notes"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type GRNLineItem struct {
	ID           string    `json:"id" db:"id"`
	GRNID        string    `json:"grn_id" db:"grn_id" binding:"required"`
	ProductID    string    `json:"product_id" db:"product_id" binding:"required"`
	BatchNumber  string    `json:"batch_number" db:"batch_number"`
	SerialNumber string    `json:"serial_number" db:"serial_number"`
	ExpiryDate   time.Time `json:"expiry_date" db:"expiry_date"`
	Quantity     float64   `json:"quantity" db:"quantity" binding:"required"`
	UnitCost     float64   `json:"unit_cost" db:"unit_cost"`
	Status       string    `json:"status" db:"status"`
}

type StockIssue struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	WarehouseID string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	RequestedBy string    `json:"requested_by" db:"requested_by" validate:"required,max=255"`
	ApprovedBy  string    `json:"approved_by" db:"approved_by"`
	ProgramID   string    `json:"program_id" db:"program_id" validate:"omitempty,uuid7"`
	Status      string    `json:"status" db:"status" validate:"omitempty,oneof=draft submitted approved rejected"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
