package handler

import (
	"time"

	"github.com/cpintl/ZarishLog/apps/api/internal/model"
	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/cpintl/ZarishLog/apps/api/internal/validator"
	"github.com/gin-gonic/gin"
)

// Policy §16 — Dispatch waybills and delivery confirmation

func CreateDispatchWaybill(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID                string  `json:"org_id" validate:"required,uuid7"`
			WaybillNumber        string  `json:"waybill_number" validate:"required,max=100"`
			IssueID              *string `json:"issue_id" validate:"opt_uuid7"`
			TransferID           *string `json:"transfer_id" validate:"opt_uuid7"`
			SenderWarehouseID    string  `json:"sender_warehouse_id" validate:"required,uuid7"`
			RecipientWarehouseID *string `json:"recipient_warehouse_id" validate:"opt_uuid7"`
			Recipient            *string `json:"recipient"`
			DispatchDate         string  `json:"dispatch_date" validate:"date"`
			Carrier              *string `json:"carrier"`
			VehicleNumber        *string `json:"vehicle_number"`
			CartonCount          *int    `json:"carton_count"`
			PackingListReference *string `json:"packing_list_reference"`
			StorageRequirement   *string `json:"storage_requirement"`
			TemperatureSensitive bool    `json:"temperature_sensitive"`
			Condition            *string `json:"condition" validate:"omitempty,oneof=intact damaged sealed tampered"`
			DocumentRefs         *string `json:"document_refs"`
			Status               string  `json:"status" validate:"omitempty,oneof=planned dispatched in_transit delivered cancelled"`
			CreatedBy            string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		status := req.Status
		if status == "" {
			status = "planned"
		}
		dispatchDate := req.DispatchDate
		if dispatchDate == "" {
			dispatchDate = time.Now().UTC().Format("2006-01-02")
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO dispatch_waybills (org_id, waybill_number, issue_id, transfer_id, sender_warehouse_id, recipient_warehouse_id, recipient, dispatch_date, carrier, vehicle_number, carton_count, packing_list_reference, storage_requirement, temperature_sensitive, condition, document_refs, status, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$18) RETURNING id`,
			req.OrgID, req.WaybillNumber, nullIfEmptyPtr(req.IssueID), nullIfEmptyPtr(req.TransferID),
			req.SenderWarehouseID, nullIfEmptyPtr(req.RecipientWarehouseID), nullIfEmptyPtr(req.Recipient), dispatchDate,
			nullIfEmptyPtr(req.Carrier), nullIfEmptyPtr(req.VehicleNumber), req.CartonCount,
			nullIfEmptyPtr(req.PackingListReference), nullIfEmptyPtr(req.StorageRequirement),
			req.TemperatureSensitive, nullIfEmptyPtr(req.Condition), nullIfEmptyPtr(req.DocumentRefs),
			status, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create waybill: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func GetDispatchWaybill(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var waybill model.DispatchWaybill
		err := db.Get(&waybill, `SELECT * FROM dispatch_waybills WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "waybill not found")
			return
		}

		response.OK(c, waybill)
	}
}

func ListDispatchWaybills(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM dispatch_waybills`)
		if err != nil {
			response.InternalError(c, "failed to count waybills")
			return
		}

		var waybills []model.DispatchWaybill
		err = db.Select(&waybills,
			`SELECT * FROM dispatch_waybills ORDER BY dispatch_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list waybills")
			return
		}

		if waybills == nil {
			waybills = []model.DispatchWaybill{}
		}

		response.Paginated(c, waybills, total, p.Page, p.PageSize)
	}
}

func CreateDeliveryConfirmation(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var req struct {
			OrgID                 string  `json:"org_id" validate:"required,uuid7"`
			WaybillID             string  `json:"waybill_id" validate:"required,uuid7"`
			ConfirmingWarehouseID *string `json:"confirming_warehouse_id" validate:"opt_uuid7"`
			RecipientName         *string `json:"recipient_name"`
			ConfirmedDate         string  `json:"confirmed_date" validate:"date"`
			DeliveryStatus        string  `json:"delivery_status" validate:"required,oneof=received partial rejected damaged lost failed"`
			ItemsConformed        *bool   `json:"items_conformed"`
			QuantityConformed     *bool   `json:"quantity_conformed"`
			BatchConformed        *bool   `json:"batch_conformed"`
			ExpiryConformed       *bool   `json:"expiry_conformed"`
			ConditionConformed    *bool   `json:"condition_conformed"`
			TemperatureConformed  *bool   `json:"temperature_conformed"`
			UnfilledQuantity      *string `json:"unfilled_quantity"`
			Substitutions         *string `json:"substitutions"`
			Discrepancies         *string `json:"discrepancies"`
			FollowUpCommitments   *string `json:"follow_up_commitments"`
			ConfirmedBy           *string `json:"confirmed_by"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}
		req.OrgID = c.GetString("org_id")

		confirmedDate := req.ConfirmedDate
		if confirmedDate == "" {
			confirmedDate = time.Now().UTC().Format("2006-01-02")
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO delivery_confirmations (org_id, waybill_id, confirming_warehouse_id, recipient_name, confirmed_date, delivery_status, items_conformed, quantity_conformed, batch_conformed, expiry_conformed, condition_conformed, temperature_conformed, unfilled_quantity, substitutions, discrepancies, follow_up_commitments, confirmed_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) RETURNING id`,
			req.OrgID, req.WaybillID, nullIfEmptyPtr(req.ConfirmingWarehouseID), nullIfEmptyPtr(req.RecipientName),
			confirmedDate, req.DeliveryStatus, req.ItemsConformed, req.QuantityConformed, req.BatchConformed,
			req.ExpiryConformed, req.ConditionConformed, req.TemperatureConformed,
			nullIfEmptyPtr(req.UnfilledQuantity), nullIfEmptyPtr(req.Substitutions), nullIfEmptyPtr(req.Discrepancies),
			nullIfEmptyPtr(req.FollowUpCommitments), nullIfEmptyPtr(req.ConfirmedBy),
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create delivery confirmation: "+err.Error())
			return
		}

		_, _ = db.Exec(`UPDATE dispatch_waybills SET status='delivered', updated_at=now() WHERE id=$1`, req.WaybillID)

		response.Created(c, gin.H{"id": id})
	}
}

func GetDeliveryConfirmation(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		var confirmation model.DeliveryConfirmation
		err := db.Get(&confirmation, `SELECT * FROM delivery_confirmations WHERE id=$1`, c.Param("id"))
		if err != nil {
			response.NotFound(c, "delivery confirmation not found")
			return
		}

		response.OK(c, confirmation)
	}
}

func ListDeliveryConfirmations(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM delivery_confirmations`)
		if err != nil {
			response.InternalError(c, "failed to count delivery confirmations")
			return
		}

		var confirmations []model.DeliveryConfirmation
		err = db.Select(&confirmations,
			`SELECT * FROM delivery_confirmations ORDER BY confirmed_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list delivery confirmations")
			return
		}

		if confirmations == nil {
			confirmations = []model.DeliveryConfirmation{}
		}

		response.Paginated(c, confirmations, total, p.Page, p.PageSize)
	}
}
