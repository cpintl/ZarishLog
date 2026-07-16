-- name: ListAMCCalculations :many
SELECT * FROM amc_calculations
WHERE org_id = $1
ORDER BY calculation_date DESC
LIMIT $2 OFFSET $3;

-- name: CreateAMCCalculation :one
INSERT INTO amc_calculations (org_id, product_id, warehouse_id, calculation_date, amc_3_months, amc_6_months, amc_12_months, max_consumption, std_deviation, calculation_period_start, calculation_period_end, calculation_status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetLatestAMC :one
SELECT * FROM amc_calculations
WHERE product_id = $1 AND warehouse_id = $2
ORDER BY calculation_date DESC
LIMIT 1;

-- name: ListReorderRecommendations :many
SELECT rr.*, p.name as product_name, p.sku as product_sku
FROM reorder_recommendations rr
JOIN products p ON rr.product_id = p.id
WHERE rr.org_id = $1
ORDER BY rr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateReorderRecommendation :one
INSERT INTO reorder_recommendations (org_id, product_id, warehouse_id, recommendation_date, current_stock, reorder_point, reorder_quantity, lead_time_days, amc_used, safety_stock, recommendation_type, priority, notes, reviewed)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: MarkRecommendationReviewed :exec
UPDATE reorder_recommendations SET reviewed = true
WHERE id = $1 AND org_id = $2;

-- name: ListForecastResults :many
SELECT * FROM forecast_results
WHERE org_id = $1
ORDER BY forecast_date DESC
LIMIT $2 OFFSET $3;

-- name: CreateForecastResult :one
INSERT INTO forecast_results (org_id, product_id, warehouse_id, forecast_date, forecast_value, lower_bound, upper_bound, confidence_level, model_version, features_used)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;
