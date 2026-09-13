-- Policy §16 — Dispatch waybills and delivery confirmation

-- name: CreateDispatchWaybill :one
INSERT INTO dispatch_waybills (
    org_id, waybill_number, issue_id, transfer_id, sender_warehouse_id,
    recipient_warehouse_id, recipient, dispatch_date, carrier, vehicle_number,
    carton_count, packing_list_reference, storage_requirement,
    temperature_sensitive, condition, document_refs, status, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
RETURNING *;

-- name: GetDispatchWaybill :one
SELECT * FROM dispatch_waybills
WHERE id = $1 AND org_id = $2;

-- name: ListDispatchWaybills :many
SELECT * FROM dispatch_waybills
WHERE org_id = $1
ORDER BY dispatch_date DESC
LIMIT $2 OFFSET $3;

-- name: ListDispatchWaybillsByStatus :many
SELECT * FROM dispatch_waybills
WHERE org_id = $1 AND status = $2
ORDER BY dispatch_date DESC
LIMIT $3 OFFSET $4;

-- name: CreateDeliveryConfirmation :one
INSERT INTO delivery_confirmations (
    org_id, waybill_id, confirming_warehouse_id, recipient_name, confirmed_date,
    delivery_status, items_conformed, quantity_conformed, batch_conformed,
    expiry_conformed, condition_conformed, temperature_conformed,
    unfilled_quantity, substitutions, discrepancies, follow_up_commitments, confirmed_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: GetDeliveryConfirmation :one
SELECT * FROM delivery_confirmations
WHERE id = $1 AND org_id = $2;

-- name: ListDeliveryConfirmations :many
SELECT * FROM delivery_confirmations
WHERE org_id = $1
ORDER BY confirmed_date DESC
LIMIT $2 OFFSET $3;

-- name: ListDeliveryConfirmationsByWaybill :many
SELECT * FROM delivery_confirmations
WHERE org_id = $1 AND waybill_id = $2
ORDER BY confirmed_date DESC
LIMIT $3 OFFSET $4;