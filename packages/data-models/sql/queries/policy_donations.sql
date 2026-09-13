-- Policy §11.3 — Donation registry and acceptance/rejection workflow

-- name: CreateDonation :one
INSERT INTO donations (
    org_id, donation_number, donor_name, donor_contact, offer_date, decision,
    decision_date, decided_by, needs_assessment, proposed_recipient,
    approval_reference, transport_responsibility, customs_notes,
    disposal_responsibility, certificate_number, certificate_issued_date,
    registry_reference, status, notes, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
RETURNING *;

-- name: GetDonation :one
SELECT * FROM donations
WHERE id = $1 AND org_id = $2;

-- name: ListDonations :many
SELECT * FROM donations
WHERE org_id = $1
ORDER BY offer_date DESC
LIMIT $2 OFFSET $3;

-- name: ListDonationsByStatus :many
SELECT * FROM donations
WHERE org_id = $1 AND status = $2
ORDER BY offer_date DESC
LIMIT $3 OFFSET $4;

-- name: GetDonationLineItems :many
SELECT * FROM donation_line_items
WHERE donation_id = $1
ORDER BY expiry_date NULLS LAST, id;

-- name: CreateDonationLineItem :one
INSERT INTO donation_line_items (
    donation_id, product_id, batch_number, expiry_date, quantity, uom,
    remaining_shelf_life_months, condition, shelf_life_compliant, authorized, notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateDonationDecision :exec
UPDATE donations SET
    decision = $1, status = $2, decision_date = $3, decided_by = $4,
    certificate_number = $5, certificate_issued_date = $6,
    updated_by = $7, updated_at = now()
WHERE id = $8 AND org_id = $9;