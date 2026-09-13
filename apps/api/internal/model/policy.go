package model

import "time"

// Policy models implement the controlled registers required by the
// "CPI BANGLADESH MEDICAL WAREHOUSE POLICY" (v1.0).

// ─── §8 Regulatory & Legal Compliance ───────────────────────────────────

type RegulatoryApproval struct {
	ID               string    `json:"id" db:"id"`
	OrgID            string    `json:"org_id" db:"org_id"`
	ScopeType        string    `json:"scope_type" db:"scope_type" validate:"required,oneof=warehouse product product_category supplier organization"`
	ScopeID          string    `json:"scope_id" db:"scope_id" validate:"required"`
	Authority        string    `json:"authority" db:"authority" validate:"required,oneof=DGDA DNC NBR CCI_E Customs Municipality Other"`
	LicenseType      string    `json:"license_type" db:"license_type" validate:"required,oneof=wholesale_license retail_license import_permit controlled_substance_permit premises_approval import_clearance registration other"`
	ReferenceNumber  string    `json:"reference_number" db:"reference_number" validate:"required"`
	ProductScope     *string   `json:"product_scope" db:"product_scope"`
	Conditions       *string   `json:"conditions" db:"conditions"`
	IssueDate        *string   `json:"issue_date" db:"issue_date" validate:"omitempty,date"`
	ExpiryDate       *string   `json:"expiry_date" db:"expiry_date" validate:"omitempty,date"`
	RenewalLeadDays  int       `json:"renewal_lead_days" db:"renewal_lead_days"`
	ResponsibleOwner *string   `json:"responsible_owner" db:"responsible_owner"`
	Status           string    `json:"status" db:"status" validate:"required,oneof=draft active expiring_soon expired revoked suspended"`
	EvidenceURL      *string   `json:"evidence_url" db:"evidence_url"`
	Notes            *string   `json:"notes" db:"notes"`
	CreatedBy        *string   `json:"created_by" db:"created_by"`
	UpdatedBy        *string   `json:"updated_by" db:"updated_by"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ─── §7.2 Deviations ────────────────────────────────────────────────────

type Deviation struct {
	ID                 string    `json:"id" db:"id"`
	OrgID              string    `json:"org_id" db:"org_id"`
	DeviationNumber    string    `json:"deviation_number" db:"deviation_number" validate:"required"`
	CategoryCode       *string   `json:"category_code" db:"category_code"`
	OccurredAt         time.Time `json:"occurred_at" db:"occurred_at"`
	ReportedAt         time.Time `json:"reported_at" db:"reported_at"`
	SourceDocumentType *string   `json:"source_document_type" db:"source_document_type"`
	SourceDocumentID   *string   `json:"source_document_id" db:"source_document_id"`
	ProductID          *string   `json:"product_id" db:"product_id" validate:"omitempty,uuid7"`
	BatchID            *string   `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	WarehouseID        *string   `json:"warehouse_id" db:"warehouse_id" validate:"omitempty,uuid7"`
	LocationID         *string   `json:"location_id" db:"location_id" validate:"omitempty,uuid7"`
	Description        string    `json:"description" db:"description" validate:"required"`
	ContainmentAction  *string   `json:"containment_action" db:"containment_action"`
	ImpactAssessment   *string   `json:"impact_assessment" db:"impact_assessment"`
	RootCause          *string   `json:"root_cause" db:"root_cause"`
	RiskRating         string    `json:"risk_rating" db:"risk_rating" validate:"required,oneof=low medium high critical"`
	Disposition        *string   `json:"disposition" db:"disposition"`
	Status             string    `json:"status" db:"status" validate:"required,oneof=open under_investigation contained resolved closed"`
	ResponsibleOwner   *string   `json:"responsible_owner" db:"responsible_owner"`
	DueDate            *string   `json:"due_date" db:"due_date" validate:"omitempty,date"`
	Escalated          bool      `json:"escalated" db:"escalated"`
	EscalatedTo        *string   `json:"escalated_to" db:"escalated_to"`
	CreatedBy          *string   `json:"created_by" db:"created_by"`
	UpdatedBy          *string   `json:"updated_by" db:"updated_by"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// ─── §7.3 CAPA ──────────────────────────────────────────────────────────

type CAPAAction struct {
	ID                      string     `json:"id" db:"id"`
	OrgID                   string     `json:"org_id" db:"org_id"`
	CAPANumber              string     `json:"capa_number" db:"capa_number" validate:"required"`
	SourceType              string     `json:"source_type" db:"source_type" validate:"required,oneof=deviation complaint recall audit inspection stock_discrepancy supplier trend other"`
	SourceID                *string    `json:"source_id" db:"source_id"`
	CAPAType                string     `json:"capa_type" db:"capa_type" validate:"required,oneof=corrective preventive"`
	Title                   string     `json:"title" db:"title" validate:"required"`
	RootCause               *string    `json:"root_cause" db:"root_cause"`
	ActionPlan              string     `json:"action_plan" db:"action_plan" validate:"required"`
	ResponsibleOwner        *string    `json:"responsible_owner" db:"responsible_owner"`
	DueDate                 *string    `json:"due_date" db:"due_date" validate:"omitempty,date"`
	Status                  string     `json:"status" db:"status" validate:"required,oneof=open in_progress overdue completed closed"`
	EffectivenessCheck      *string    `json:"effectiveness_check" db:"effectiveness_check"`
	EffectivenessVerified   bool       `json:"effectiveness_verified" db:"effectiveness_verified"`
	EffectivenessVerifiedBy *string    `json:"effectiveness_verified_by" db:"effectiveness_verified_by"`
	VerifiedAt              *time.Time `json:"verified_at" db:"verified_at"`
	ClosedBy                *string    `json:"closed_by" db:"closed_by"`
	ClosedAt                *time.Time `json:"closed_at" db:"closed_at"`
	CreatedBy               *string    `json:"created_by" db:"created_by"`
	UpdatedBy               *string    `json:"updated_by" db:"updated_by"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
}

// ─── §7.4 Training & Competency ─────────────────────────────────────────

type TrainingRecord struct {
	ID                   string    `json:"id" db:"id"`
	OrgID                string    `json:"org_id" db:"org_id"`
	UserID               *string   `json:"user_id" db:"user_id" validate:"omitempty,uuid7"`
	TrainingTitle        string    `json:"training_title" db:"training_title" validate:"required"`
	TrainingType         string    `json:"training_type" db:"training_type" validate:"required,oneof=initial refresher emergency contractor_safety"`
	Topic                *string   `json:"topic" db:"topic"`
	TrainingDate         string    `json:"training_date" db:"training_date" validate:"required,date"`
	Method               *string   `json:"method" db:"method"`
	Trainer              *string   `json:"trainer" db:"trainer"`
	Assessed             bool      `json:"assessed" db:"assessed"`
	AssessmentResult     *string   `json:"assessment_result" db:"assessment_result"`
	CompetencyValidUntil *string   `json:"competency_valid_until" db:"competency_valid_until" validate:"omitempty,date"`
	CertificateURL       *string   `json:"certificate_url" db:"certificate_url"`
	CreatedBy            *string   `json:"created_by" db:"created_by"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// ─── §13.2 Temperature/Humidity Monitoring ──────────────────────────────

type TemperatureMonitoringEntry struct {
	ID           string    `json:"id" db:"id"`
	OrgID        string    `json:"org_id" db:"org_id"`
	WarehouseID  string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	LocationID   *string   `json:"location_id" db:"location_id" validate:"omitempty,uuid7"`
	EquipmentID  *string   `json:"equipment_id" db:"equipment_id" validate:"omitempty,uuid7"`
	DeviceName   *string   `json:"device_name" db:"device_name"`
	MonitorType  string    `json:"monitor_type" db:"monitor_type" validate:"required,oneof=temperature humidity freeze_indicator vvm"`
	ReadingValue float64   `json:"reading_value" db:"reading_value" validate:"required"`
	MinThreshold *float64  `json:"min_threshold" db:"min_threshold"`
	MaxThreshold *float64  `json:"max_threshold" db:"max_threshold"`
	RecordedAt   time.Time `json:"recorded_at" db:"recorded_at"`
	RecordedBy   *string   `json:"recorded_by" db:"recorded_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ─── §13.3 Alarm & Excursion Response ───────────────────────────────────

type TemperatureExcursion struct {
	ID                  string     `json:"id" db:"id"`
	OrgID               string     `json:"org_id" db:"org_id"`
	WarehouseID         string     `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	LocationID          *string    `json:"location_id" db:"location_id" validate:"omitempty,uuid7"`
	EquipmentID         *string    `json:"equipment_id" db:"equipment_id" validate:"omitempty,uuid7"`
	DeviceName          *string    `json:"device_name" db:"device_name"`
	ExcursionType       string     `json:"excursion_type" db:"excursion_type" validate:"required,oneof=temperature humidity freeze power_failure equipment_failure monitoring_gap"`
	StartedAt           time.Time  `json:"started_at" db:"started_at"`
	EndedAt             *time.Time `json:"ended_at" db:"ended_at"`
	MinValue            *float64   `json:"min_value" db:"min_value"`
	MaxValue            *float64   `json:"max_value" db:"max_value"`
	ExpectedMin         *float64   `json:"expected_min" db:"expected_min"`
	ExpectedMax         *float64   `json:"expected_max" db:"expected_max"`
	AffectedProducts    *string    `json:"affected_products" db:"affected_products"`
	AffectedBatches     *string    `json:"affected_batches" db:"affected_batches"`
	QuantityAffected    *float64   `json:"quantity_affected" db:"quantity_affected"`
	Quarantined         bool       `json:"quarantined" db:"quarantined"`
	QuarantineReference *string    `json:"quarantine_reference" db:"quarantine_reference"`
	Disposition         string     `json:"disposition" db:"disposition" validate:"required,oneof=pending released restricted_use returned destroyed investigating"`
	NotifiedQuality     bool       `json:"notified_quality" db:"notified_quality"`
	NotifiedManager     bool       `json:"notified_manager" db:"notified_manager"`
	TechnicalAdvice     *string    `json:"technical_advice" db:"technical_advice"`
	RootCause           *string    `json:"root_cause" db:"root_cause"`
	CapaID              *string    `json:"capa_id" db:"capa_id" validate:"omitempty,uuid7"`
	Status              string     `json:"status" db:"status" validate:"required,oneof=open under_review dispositioned closed"`
	CreatedBy           *string    `json:"created_by" db:"created_by"`
	UpdatedBy           *string    `json:"updated_by" db:"updated_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// ─── §12.2 Stock Release ────────────────────────────────────────────────

type StockReleaseRecord struct {
	ID                  string    `json:"id" db:"id"`
	OrgID               string    `json:"org_id" db:"org_id"`
	ReleaseNumber       string    `json:"release_number" db:"release_number" validate:"required"`
	WarehouseID         string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	ProductID           string    `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchID             *string   `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	LocationID          *string   `json:"location_id" db:"location_id" validate:"omitempty,uuid7"`
	Quantity            float64   `json:"quantity" db:"quantity" validate:"required"`
	QuarantineReference *string   `json:"quarantine_reference" db:"quarantine_reference"`
	EvidenceReviewed    *string   `json:"evidence_reviewed" db:"evidence_reviewed"`
	Decision            string    `json:"decision" db:"decision" validate:"required,oneof=release restricted_release reject destruction return"`
	DecisionDate        string    `json:"decision_date" db:"decision_date" validate:"required,date"`
	DecidedBy           string    `json:"decided_by" db:"decided_by" validate:"required"`
	Notes               *string   `json:"notes" db:"notes"`
	CreatedBy           *string   `json:"created_by" db:"created_by"`
	UpdatedBy           *string   `json:"updated_by" db:"updated_by"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// ─── §14.3 Controlled-Drug Register ─────────────────────────────────────

type ControlledStockRegisterEntry struct {
	ID                  string     `json:"id" db:"id"`
	OrgID               string     `json:"org_id" db:"org_id"`
	ProductID           string     `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchID             *string    `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	WarehouseID         string     `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	RegisterDate        string     `json:"register_date" db:"register_date" validate:"required,date"`
	TransactionType     string     `json:"transaction_type" db:"transaction_type" validate:"required,oneof=receipt issue transfer_in transfer_out return disposal reconciliation opening_balance"`
	ReferenceDocument   *string    `json:"reference_document" db:"reference_document"`
	Quantity            float64    `json:"quantity" db:"quantity" validate:"required"`
	BalanceAfter        float64    `json:"balance_after" db:"balance_after" validate:"required"`
	ReceivedFrom        *string    `json:"received_from" db:"received_from"`
	IssuedTo            *string    `json:"issued_to" db:"issued_to"`
	AuthorizedRecipient *string    `json:"authorized_recipient" db:"authorized_recipient"`
	Signature           *string    `json:"signature" db:"signature"`
	Discrepancy         float64    `json:"discrepancy" db:"discrepancy"`
	ReconciledAt        *time.Time `json:"reconciled_at" db:"reconciled_at"`
	ReconciledBy        *string    `json:"reconciled_by" db:"reconciled_by"`
	CreatedBy           *string    `json:"created_by" db:"created_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
}

// ─── §11.3 Donations ────────────────────────────────────────────────────

type Donation struct {
	ID                      string    `json:"id" db:"id"`
	OrgID                   string    `json:"org_id" db:"org_id"`
	DonationNumber          string    `json:"donation_number" db:"donation_number" validate:"required"`
	DonorName               string    `json:"donor_name" db:"donor_name" validate:"required"`
	DonorContact            *string   `json:"donor_contact" db:"donor_contact"`
	OfferDate               string    `json:"offer_date" db:"offer_date" validate:"required,date"`
	Decision                *string   `json:"decision" db:"decision"`
	DecisionDate            *string   `json:"decision_date" db:"decision_date" validate:"omitempty,date"`
	DecidedBy               *string   `json:"decided_by" db:"decided_by"`
	NeedsAssessment         *string   `json:"needs_assessment" db:"needs_assessment"`
	ProposedRecipient       *string   `json:"proposed_recipient" db:"proposed_recipient"`
	ApprovalReference       *string   `json:"approval_reference" db:"approval_reference"`
	TransportResponsibility *string   `json:"transport_responsibility" db:"transport_responsibility"`
	CustomsNotes            *string   `json:"customs_notes" db:"customs_notes"`
	DisposalResponsibility  *string   `json:"disposal_responsibility" db:"disposal_responsibility"`
	CertificateNumber       *string   `json:"certificate_number" db:"certificate_number"`
	CertificateIssuedDate   *string   `json:"certificate_issued_date" db:"certificate_issued_date" validate:"omitempty,date"`
	RegistryReference       *string   `json:"registry_reference" db:"registry_reference"`
	Status                  string    `json:"status" db:"status" validate:"required,oneof=draft pending accepted rejected quarantined received"`
	Notes                   *string   `json:"notes" db:"notes"`
	CreatedBy               *string   `json:"created_by" db:"created_by"`
	UpdatedBy               *string   `json:"updated_by" db:"updated_by"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

type DonationLineItem struct {
	ID                       string  `json:"id" db:"id"`
	DonationID               string  `json:"donation_id" db:"donation_id" validate:"required,uuid7"`
	ProductID                string  `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchNumber              *string `json:"batch_number" db:"batch_number"`
	ExpiryDate               *string `json:"expiry_date" db:"expiry_date" validate:"omitempty,date"`
	Quantity                 float64 `json:"quantity" db:"quantity" validate:"required"`
	Uom                      *string `json:"uom" db:"uom"`
	RemainingShelfLifeMonths *int    `json:"remaining_shelf_life_months" db:"remaining_shelf_life_months"`
	Condition                *string `json:"condition" db:"condition"`
	ShelfLifeCompliant       *bool   `json:"shelf_life_compliant" db:"shelf_life_compliant"`
	Authorized               bool    `json:"authorized" db:"authorized"`
	Notes                    *string `json:"notes" db:"notes"`
}

// ─── §17.2 Complaints ───────────────────────────────────────────────────

type Complaint struct {
	ID                         string    `json:"id" db:"id"`
	OrgID                      string    `json:"org_id" db:"org_id"`
	ComplaintNumber            string    `json:"complaint_number" db:"complaint_number" validate:"required"`
	ReceivedDate               string    `json:"received_date" db:"received_date" validate:"required,date"`
	ReceivedVia                *string   `json:"received_via" db:"received_via"`
	Severity                   *string   `json:"severity" db:"severity"`
	Category                   *string   `json:"category" db:"category"`
	ProductID                  *string   `json:"product_id" db:"product_id" validate:"omitempty,uuid7"`
	BatchID                    *string   `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	Description                string    `json:"description" db:"description" validate:"required"`
	ReporterName               *string   `json:"reporter_name" db:"reporter_name"`
	ReporterContact            *string   `json:"reporter_contact" db:"reporter_contact"`
	ImmediateEscalation        bool      `json:"immediate_escalation" db:"immediate_escalation"`
	Escalated                  bool      `json:"escalated" db:"escalated"`
	EscalatedTo                *string   `json:"escalated_to" db:"escalated_to"`
	ManufacturerNotified       bool      `json:"manufacturer_notified" db:"manufacturer_notified"`
	CompetentAuthorityNotified bool      `json:"competent_authority_notified" db:"competent_authority_notified"`
	Status                     string    `json:"status" db:"status" validate:"required,oneof=open under_review investigating resolved closed"`
	ResolutionNotes            *string   `json:"resolution_notes" db:"resolution_notes"`
	CapaID                     *string   `json:"capa_id" db:"capa_id" validate:"omitempty,uuid7"`
	CreatedBy                  *string   `json:"created_by" db:"created_by"`
	UpdatedBy                  *string   `json:"updated_by" db:"updated_by"`
	CreatedAt                  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at" db:"updated_at"`
}

// ─── §17.3 Recalls (incl. mock recall) ──────────────────────────────────

type Recall struct {
	ID              string    `json:"id" db:"id"`
	OrgID           string    `json:"org_id" db:"org_id"`
	RecallNumber    string    `json:"recall_number" db:"recall_number" validate:"required"`
	RecallType      string    `json:"recall_type" db:"recall_type" validate:"required,oneof=ACTUAL MOCK"`
	RecallDate      string    `json:"recall_date" db:"recall_date" validate:"required,date"`
	ProductID       *string   `json:"product_id" db:"product_id" validate:"omitempty,uuid7"`
	SubjectProduct  string    `json:"subject_product" db:"subject_product" validate:"required"`
	BatchNumbers    *string   `json:"batch_numbers" db:"batch_numbers"`
	Reason          string    `json:"reason" db:"reason" validate:"required"`
	InitiatedBy     *string   `json:"initiated_by" db:"initiated_by"`
	AuthorizedBy    *string   `json:"authorized_by" db:"authorized_by"`
	Communications  *string   `json:"communications" db:"communications"`
	DispositionPlan *string   `json:"disposition_plan" db:"disposition_plan"`
	Status          string    `json:"status" db:"status" validate:"required,oneof=draft in_progress pending_disposition reconciled closed"`
	ClosureEvidence *string   `json:"closure_evidence" db:"closure_evidence"`
	CreatedBy       *string   `json:"created_by" db:"created_by"`
	UpdatedBy       *string   `json:"updated_by" db:"updated_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type RecallLineItem struct {
	ID                  string     `json:"id" db:"id"`
	RecallID            string     `json:"recall_id" db:"recall_id" validate:"required,uuid7"`
	Recipient           string     `json:"recipient" db:"recipient" validate:"required"`
	Location            *string    `json:"location" db:"location"`
	ProductID           *string    `json:"product_id" db:"product_id" validate:"omitempty,uuid7"`
	BatchNumber         *string    `json:"batch_number" db:"batch_number"`
	QuantityIssued      float64    `json:"quantity_issued" db:"quantity_issued"`
	QuantityRecovered   float64    `json:"quantity_recovered" db:"quantity_recovered"`
	QuantityOutstanding float64    `json:"quantity_outstanding" db:"quantity_outstanding"`
	StorageDisposition  *string    `json:"storage_disposition" db:"storage_disposition"`
	RecoveredAt         *time.Time `json:"recovered_at" db:"recovered_at"`
	Notes               *string    `json:"notes" db:"notes"`
}

// ─── §16.2 / §16.4 Dispatch & Delivery ──────────────────────────────────

type DispatchWaybill struct {
	ID                   string    `json:"id" db:"id"`
	OrgID                string    `json:"org_id" db:"org_id"`
	WaybillNumber        string    `json:"waybill_number" db:"waybill_number" validate:"required"`
	IssueID              *string   `json:"issue_id" db:"issue_id" validate:"omitempty,uuid7"`
	TransferID           *string   `json:"transfer_id" db:"transfer_id" validate:"omitempty,uuid7"`
	SenderWarehouseID    string    `json:"sender_warehouse_id" db:"sender_warehouse_id" validate:"required,uuid7"`
	RecipientWarehouseID *string   `json:"recipient_warehouse_id" db:"recipient_warehouse_id" validate:"omitempty,uuid7"`
	Recipient            *string   `json:"recipient" db:"recipient"`
	DispatchDate         string    `json:"dispatch_date" db:"dispatch_date" validate:"required,date"`
	Carrier              *string   `json:"carrier" db:"carrier"`
	VehicleNumber        *string   `json:"vehicle_number" db:"vehicle_number"`
	CartonCount          *int      `json:"carton_count" db:"carton_count"`
	PackingListReference *string   `json:"packing_list_reference" db:"packing_list_reference"`
	StorageRequirement   *string   `json:"storage_requirement" db:"storage_requirement"`
	TemperatureSensitive bool      `json:"temperature_sensitive" db:"temperature_sensitive"`
	Condition            *string   `json:"condition" db:"condition"`
	DocumentRefs         *string   `json:"document_refs" db:"document_refs"`
	Status               string    `json:"status" db:"status" validate:"required,oneof=planned dispatched in_transit delivered cancelled"`
	CreatedBy            *string   `json:"created_by" db:"created_by"`
	UpdatedBy            *string   `json:"updated_by" db:"updated_by"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

type DeliveryConfirmation struct {
	ID                    string    `json:"id" db:"id"`
	OrgID                 string    `json:"org_id" db:"org_id"`
	WaybillID             string    `json:"waybill_id" db:"waybill_id" validate:"required,uuid7"`
	ConfirmingWarehouseID *string   `json:"confirming_warehouse_id" db:"confirming_warehouse_id" validate:"omitempty,uuid7"`
	RecipientName         *string   `json:"recipient_name" db:"recipient_name"`
	ConfirmedDate         string    `json:"confirmed_date" db:"confirmed_date" validate:"required,date"`
	DeliveryStatus        string    `json:"delivery_status" db:"delivery_status" validate:"required,oneof=received partial rejected damaged lost failed"`
	ItemsConformed        *bool     `json:"items_conformed" db:"items_conformed"`
	QuantityConformed     *bool     `json:"quantity_conformed" db:"quantity_conformed"`
	BatchConformed        *bool     `json:"batch_conformed" db:"batch_conformed"`
	ExpiryConformed       *bool     `json:"expiry_conformed" db:"expiry_conformed"`
	ConditionConformed    *bool     `json:"condition_conformed" db:"condition_conformed"`
	TemperatureConformed  *bool     `json:"temperature_conformed" db:"temperature_conformed"`
	UnfilledQuantity      *string   `json:"unfilled_quantity" db:"unfilled_quantity"`
	Substitutions         *string   `json:"substitutions" db:"substitutions"`
	Discrepancies         *string   `json:"discrepancies" db:"discrepancies"`
	FollowUpCommitments   *string   `json:"follow_up_commitments" db:"follow_up_commitments"`
	ConfirmedBy           *string   `json:"confirmed_by" db:"confirmed_by"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// ─── §14.4 Short-Expiry Review ──────────────────────────────────────────

type ShortExpiryReview struct {
	ID                       string    `json:"id" db:"id"`
	OrgID                    string    `json:"org_id" db:"org_id"`
	ReviewDate               string    `json:"review_date" db:"review_date" validate:"required,date"`
	ProductID                string    `json:"product_id" db:"product_id" validate:"required,uuid7"`
	BatchID                  *string   `json:"batch_id" db:"batch_id" validate:"omitempty,uuid7"`
	WarehouseID              string    `json:"warehouse_id" db:"warehouse_id" validate:"required,uuid7"`
	ExpiryDate               string    `json:"expiry_date" db:"expiry_date" validate:"required,date"`
	RemainingShelfLifeMonths int       `json:"remaining_shelf_life_months" db:"remaining_shelf_life_months" validate:"required"`
	ExpectedConsumption      *string   `json:"expected_consumption" db:"expected_consumption"`
	TransferOption           *string   `json:"transfer_option" db:"transfer_option"`
	DonorCondition           *string   `json:"donor_condition" db:"donor_condition"`
	RegulatoryRestriction    *string   `json:"regulatory_restriction" db:"regulatory_restriction"`
	Decision                 string    `json:"decision" db:"decision" validate:"required,oneof=use_before_expiry transfer redistribute return donate destroy pending"`
	DecisionBy               *string   `json:"decision_by" db:"decision_by"`
	DecisionDate             *string   `json:"decision_date" db:"decision_date" validate:"omitempty,date"`
	Notes                    *string   `json:"notes" db:"notes"`
	CreatedBy                *string   `json:"created_by" db:"created_by"`
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" db:"updated_at"`
}

// ─── §20 Emergency Plans ────────────────────────────────────────────────

type EmergencyPlan struct {
	ID                     string     `json:"id" db:"id"`
	OrgID                  string     `json:"org_id" db:"org_id"`
	PlanNumber             string     `json:"plan_number" db:"plan_number" validate:"required"`
	Name                   string     `json:"name" db:"name" validate:"required"`
	CountryCode            *string    `json:"country_code" db:"country_code"`
	FacilityID             *string    `json:"facility_id" db:"facility_id" validate:"omitempty,uuid7"`
	ScenarioType           string     `json:"scenario_type" db:"scenario_type" validate:"required,oneof=outbreak displacement cyclone flood fire security scale_up power_loss cold_chain_failure supplier_interruption border_delay system_outage warehouse_loss staff_shortage other"`
	Scenario               *string    `json:"scenario" db:"scenario"`
	DemandAssumptions      *string    `json:"demand_assumptions" db:"demand_assumptions"`
	ServiceLevel           *string    `json:"service_level" db:"service_level"`
	PrepositionedStockPlan *string    `json:"prepositioned_stock_plan" db:"prepositioned_stock_plan"`
	RotationPlan           *string    `json:"rotation_plan" db:"rotation_plan"`
	AlternateSuppliers     *string    `json:"alternate_suppliers" db:"alternate_suppliers"`
	AlternateWarehouses    *string    `json:"alternate_warehouses" db:"alternate_warehouses"`
	AlternateRoutes        *string    `json:"alternate_routes" db:"alternate_routes"`
	AlternatePowerSources  *string    `json:"alternate_power_sources" db:"alternate_power_sources"`
	EmergencyAuthority     *string    `json:"emergency_authority" db:"emergency_authority"`
	PaperFallbackRecords   bool       `json:"paper_fallback_records" db:"paper_fallback_records"`
	MinimumStaffing        *string    `json:"minimum_staffing" db:"minimum_staffing"`
	ColdChainContingency   *string    `json:"cold_chain_contingency" db:"cold_chain_contingency"`
	Communications         *string    `json:"communications" db:"communications"`
	EscalationContacts     *string    `json:"escalation_contacts" db:"escalation_contacts"`
	SecurityControls       *string    `json:"security_controls" db:"security_controls"`
	ReturnToNormalPlan     *string    `json:"return_to_normal_plan" db:"return_to_normal_plan"`
	Status                 string     `json:"status" db:"status" validate:"required,oneof=draft approved active exercised superseded"`
	ApprovedBy             *string    `json:"approved_by" db:"approved_by"`
	ApprovedAt             *time.Time `json:"approved_at" db:"approved_at"`
	CreatedBy              *string    `json:"created_by" db:"created_by"`
	UpdatedBy              *string    `json:"updated_by" db:"updated_by"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

// ─── §27 Change Control ─────────────────────────────────────────────────

type ChangeControl struct {
	ID                    string    `json:"id" db:"id"`
	OrgID                 string    `json:"org_id" db:"org_id"`
	ChangeNumber          string    `json:"change_number" db:"change_number" validate:"required"`
	SubjectType           string    `json:"subject_type" db:"subject_type" validate:"required,oneof=product supplier warehouse_layout equipment software data_structure program service_level storage_condition legal_requirement sop"`
	SubjectID             *string   `json:"subject_id" db:"subject_id"`
	Description           string    `json:"description" db:"description" validate:"required"`
	ImpactAssessment      *string   `json:"impact_assessment" db:"impact_assessment"`
	RiskLevel             string    `json:"risk_level" db:"risk_level" validate:"required,oneof=low medium high critical"`
	RequiresQualityReview bool      `json:"requires_quality_review" db:"requires_quality_review"`
	AuthorizedBy          *string   `json:"authorized_by" db:"authorized_by"`
	AuthorizationDate     *string   `json:"authorization_date" db:"authorization_date" validate:"omitempty,date"`
	StartDate             *string   `json:"start_date" db:"start_date" validate:"omitempty,date"`
	EndDate               *string   `json:"end_date" db:"end_date" validate:"omitempty,date"`
	Status                string    `json:"status" db:"status" validate:"required,oneof=draft assessed approved implemented closed rejected"`
	ClosureEvidence       *string   `json:"closure_evidence" db:"closure_evidence"`
	CreatedBy             *string   `json:"created_by" db:"created_by"`
	UpdatedBy             *string   `json:"updated_by" db:"updated_by"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}
