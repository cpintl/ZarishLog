package model

import "time"

type StockLevel struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	ProductID   string    `json:"product_id" db:"product_id" validate:"required,uuid7"`
	WarehouseID string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	LocationID  *string   `json:"location_id" db:"location_id" validate:"omitempty,uuid7"`
	BatchID     *string   `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	Quantity    float64   `json:"quantity" db:"quantity" validate:"min=0"`
	ReservedQty float64   `json:"reserved_qty" db:"reserved_qty" validate:"min=0"`
	Status      string    `json:"status" db:"status" validate:"omitempty,oneof=on_hand reserved committed in_transit backordered"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type StockMovement struct {
	ID           string    `json:"id" db:"id"`
	OrgID        string    `json:"org_id" db:"org_id"`
	ProductID    string    `json:"product_id" db:"product_id" validate:"required,uuid7"`
	WarehouseID  string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	LocationID   *string   `json:"location_id" db:"location_id" validate:"omitempty,uuid7"`
	BatchID      *string   `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	MovementType string    `json:"movement_type" db:"movement_type" validate:"required,movement_type"`
	Quantity     float64   `json:"quantity" db:"quantity"`
	RefDocType   string    `json:"ref_doc_type" db:"ref_doc_type"`
	RefDocID     string    `json:"ref_doc_id" db:"ref_doc_id"`
	ReasonCode   string    `json:"reason_code" db:"reason_code"`
	Reference    string    `json:"reference" db:"reference"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type GRN struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	WarehouseID string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	GRNNumber   string    `json:"grn_number" db:"grn_number" validate:"required,max=100"`
	Supplier    string    `json:"supplier" db:"supplier" validate:"required,max=255"`
	PONumber    string    `json:"po_number" db:"po_number"`
	ReceivedBy  string    `json:"received_by" db:"received_by" validate:"required,max=255"`
	Status      string    `json:"status" db:"status" validate:"omitempty,oneof=draft pending completed cancelled"`
	Notes       string    `json:"notes" db:"notes"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type GRNLineItem struct {
	ID           string    `json:"id" db:"id"`
	GRNID        string    `json:"grn_id" db:"grn_id" validate:"required,uuid7"`
	ProductID    string    `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchNumber  string    `json:"batch_number" db:"batch_number"`
	SerialNumber string    `json:"serial_number" db:"serial_number"`
	ExpiryDate   *string   `json:"expiry_date" db:"expiry_date" validate:"omitempty,date"`
	Quantity     float64   `json:"quantity" db:"quantity" validate:"required,min=0"`
	UnitCost     float64   `json:"unit_cost" db:"unit_cost" validate:"min=0"`
	Status       string    `json:"status" db:"status"`
}

type StockIssue struct {
	ID           string    `json:"id" db:"id"`
	OrgID        string    `json:"org_id" db:"org_id"`
	WarehouseID  string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	IssueNumber  string    `json:"issue_number" db:"issue_number" validate:"required,max=100"`
	RequestedBy  string    `json:"requested_by" db:"requested_by" validate:"required,max=255"`
	ApprovedBy   string    `json:"approved_by" db:"approved_by"`
	ProgramID    string    `json:"program_id" db:"program_id" validate:"omitempty,uuid7"`
	DepartmentID string    `json:"department_id" db:"department_id" validate:"omitempty,uuid7"`
	Status       string    `json:"status" db:"status" validate:"omitempty,oneof=draft submitted approved rejected"`
	Notes        string    `json:"notes" db:"notes"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	UpdatedBy    string    `json:"updated_by" db:"updated_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type IssueLineItem struct {
	ID        string  `json:"id" db:"id"`
	IssueID   string  `json:"issue_id" db:"issue_id" validate:"required,uuid7"`
	ProductID string  `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchID   string  `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	Quantity  float64 `json:"quantity" db:"quantity" validate:"required,min=0"`
	UnitCost  float64 `json:"unit_cost" db:"unit_cost" validate:"min=0"`
}

type StockTransfer struct {
	ID              string    `json:"id" db:"id"`
	OrgID           string    `json:"org_id" db:"org_id"`
	FromWarehouseID string    `json:"from_warehouse_id" db:"from_warehouse_id" validate:"required,uuid7"`
	ToWarehouseID   string    `json:"to_warehouse_id" db:"to_warehouse_id" validate:"required,uuid7"`
	TransferNumber  string    `json:"transfer_number" db:"transfer_number" validate:"required,max=100"`
	Status          string    `json:"status" db:"status" validate:"omitempty,oneof=draft submitted completed cancelled"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	UpdatedBy       string    `json:"updated_by" db:"updated_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type TransferLineItem struct {
	ID         string  `json:"id" db:"id"`
	TransferID string  `json:"transfer_id" db:"transfer_id" validate:"required,uuid7"`
	ProductID  string  `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchID    string  `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	Quantity   float64 `json:"quantity" db:"quantity" validate:"required,min=0"`
	UnitCost   float64 `json:"unit_cost" db:"unit_cost" validate:"min=0"`
	Status     string  `json:"status" db:"status"`
}

type StockAdjustment struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	WarehouseID string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	ReasonCode  string    `json:"reason_code" db:"reason_code" validate:"required,max=50"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status" validate:"omitempty,oneof=draft submitted approved cancelled"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AdjustmentLineItem struct {
	ID               string  `json:"id" db:"id"`
	AdjustmentID     string  `json:"adjustment_id" db:"adjustment_id" validate:"required,uuid7"`
	ProductID        string  `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchID          string  `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	LocationID       string  `json:"location_id" db:"location_id" validate:"omitempty,uuid7"`
	ExpectedQuantity float64 `json:"expected_quantity" db:"expected_quantity" validate:"min=0"`
	ActualQuantity   float64 `json:"actual_quantity" db:"actual_quantity" validate:"min=0"`
	Difference       float64 `json:"difference" db:"difference"`
	ReasonCode       string  `json:"reason_code" db:"reason_code"`
	UnitCost         float64 `json:"unit_cost" db:"unit_cost" validate:"min=0"`
	Notes            string  `json:"notes" db:"notes"`
}
