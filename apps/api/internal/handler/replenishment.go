package handler

import (
	"fmt"
	"math"
	"time"

	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func CalculateAMC(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrgID       string `json:"org_id" validate:"required,uuid7"`
			ProductID   string `json:"product_id" validate:"required,uuid7"`
			WarehouseID string `json:"warehouse_id" validate:"required,uuid7"`
			CreatedBy   string `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		now := time.Now()
		periodEnd := now.Format("2006-01-02")
		periodStart3 := now.AddDate(0, -3, 0).Format("2006-01-02")
		periodStart6 := now.AddDate(0, -6, 0).Format("2006-01-02")
		periodStart12 := now.AddDate(0, -12, 0).Format("2006-01-02")

		type Monthly struct {
			Month string  `db:"month"`
			Qty   float64 `db:"qty"`
		}

		avg := func(periodStart string) (float64, float64, int) {
			var rows []Monthly
			err := db.Select(&rows,
				`SELECT to_char(movement_date, 'YYYY-MM') AS month, SUM(quantity) AS qty
				 FROM stock_movements
				 WHERE org_id=$1 AND product_id=$2 AND warehouse_id=$3
				 AND movement_type IN ('issue','transfer_out','adjustment_out')
				 AND movement_date >= $4 AND movement_date <= $5
				 GROUP BY month ORDER BY month`,
				req.OrgID, req.ProductID, req.WarehouseID, periodStart, periodEnd,
			)
			if err != nil || len(rows) == 0 {
				return 0, 0, 0
			}
			var total float64
			var maxVal float64
			for _, r := range rows {
				total += r.Qty
				if r.Qty > maxVal {
					maxVal = r.Qty
				}
			}
			n := len(rows)
			mean := total / float64(n)
			return mean, maxVal, n
		}

		amc3, max3, n3 := avg(periodStart3)
		amc6, max6, n6 := avg(periodStart6)
		amc12, max12, n12 := avg(periodStart12)

		var maxConsumption float64
		if max3 >= max6 && max3 >= max12 {
			maxConsumption = max3
		} else if max6 >= max12 {
			maxConsumption = max6
		} else {
			maxConsumption = max12
		}

		var stdDev float64
		if n12 > 1 {
			var rows []Monthly
			_ = db.Select(&rows,
				`SELECT to_char(movement_date, 'YYYY-MM') AS month, SUM(quantity) AS qty
				 FROM stock_movements
				 WHERE org_id=$1 AND product_id=$2 AND warehouse_id=$3
				 AND movement_type IN ('issue','transfer_out','adjustment_out')
				 AND movement_date >= $4 AND movement_date <= $5
				 GROUP BY month ORDER BY month`,
				req.OrgID, req.ProductID, req.WarehouseID, periodStart12, periodEnd,
			)
			if len(rows) > 0 {
				var total float64
				for _, r := range rows {
					total += r.Qty
				}
				mean := total / float64(len(rows))
				var sumSq float64
				for _, r := range rows {
					d := r.Qty - mean
					sumSq += d * d
				}
				stdDev = math.Sqrt(sumSq / float64(len(rows)))
			}
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO amc_calculations (org_id, product_id, warehouse_id, calculation_date,
			 amc_3_months, amc_6_months, amc_12_months, max_consumption, std_deviation,
			 calculation_period_start, calculation_period_end, calculation_status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
			req.OrgID, req.ProductID, req.WarehouseID, periodEnd,
			nullIfZero(amc3, n3), nullIfZero(amc6, n6), nullIfZero(amc12, n12),
			nullIfZero(maxConsumption, n12), nullIfZero(stdDev, n12),
			periodStart12, periodEnd, "completed",
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to store AMC: "+err.Error())
			return
		}

		response.Created(c, gin.H{
			"id": id,
			"amc_3_months":  nullIfZero(amc3, n3),
			"amc_6_months":  nullIfZero(amc6, n6),
			"amc_12_months": nullIfZero(amc12, n12),
			"max_consumption": nullIfZero(maxConsumption, n12),
			"std_deviation": nullIfZero(stdDev, n12),
			"months_3_data": n3,
			"months_6_data": n6,
			"months_12_data": n12,
		})
	}
}

func nullIfZero(val float64, n int) interface{} {
	if n == 0 || val == 0 {
		return nil
	}
	return val
}

func ListAMCCalculations(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM amc_calculations`)
		if err != nil {
			response.InternalError(c, "failed to count AMC records")
			return
		}

		var records []model.AMCCalculation
		err = db.Select(&records,
			`SELECT id, org_id, product_id, warehouse_id, calculation_date,
			 amc_3_months, amc_6_months, amc_12_months, max_consumption, std_deviation,
			 calculation_period_start, calculation_period_end, calculation_status, created_at
			 FROM amc_calculations ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list AMC records")
			return
		}

		if records == nil {
			records = []model.AMCCalculation{}
		}

		response.Paginated(c, records, total, p.Page, p.PageSize)
	}
}

func GetLatestAMC(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Query("product_id")
		warehouseID := c.Query("warehouse_id")
		if productID == "" || warehouseID == "" {
			response.BadRequest(c, "product_id and warehouse_id are required")
			return
		}

		var rec model.AMCCalculation
		err := db.Get(&rec,
			`SELECT id, org_id, product_id, warehouse_id, calculation_date,
			 amc_3_months, amc_6_months, amc_12_months, max_consumption, std_deviation,
			 calculation_period_start, calculation_period_end, calculation_status, created_at
			 FROM amc_calculations
			 WHERE product_id=$1 AND warehouse_id=$2
			 ORDER BY calculation_date DESC LIMIT 1`,
			productID, warehouseID,
		)
		if err != nil {
			response.NotFound(c, "AMC not found for this product/warehouse")
			return
		}

		response.OK(c, rec)
	}
}

func ListReorderRecommendations(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM reorder_recommendations`)
		if err != nil {
			response.InternalError(c, "failed to count recommendations")
			return
		}

		type RecWithProduct struct {
			model.ReorderRecommendation
			ProductName string `json:"product_name" db:"product_name"`
			ProductSKU  string `json:"product_sku" db:"product_sku"`
		}

		var recs []RecWithProduct
		err = db.Select(&recs,
			`SELECT rr.*, p.name AS product_name, p.sku AS product_sku
			 FROM reorder_recommendations rr
			 JOIN products p ON rr.product_id = p.id
			 ORDER BY rr.created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list recommendations")
			return
		}

		if recs == nil {
			recs = []RecWithProduct{}
		}

		response.Paginated(c, recs, total, p.Page, p.PageSize)
	}
}

func CreateReorderRecommendation(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrgID              string   `json:"org_id" validate:"required,uuid7"`
			ProductID          string   `json:"product_id" validate:"required,uuid7"`
			WarehouseID        string   `json:"warehouse_id" validate:"required,uuid7"`
			CurrentStock       float64  `json:"current_stock" validate:"min=0"`
			ReorderPoint       float64  `json:"reorder_point" validate:"min=0"`
			ReorderQuantity    float64  `json:"reorder_quantity" validate:"min=0"`
			LeadTimeDays       *int     `json:"lead_time_days" validate:"omitempty,min=0"`
			AMCUsed            *float64 `json:"amc_used"`
			SafetyStock        *float64 `json:"safety_stock"`
			RecommendationType string   `json:"recommendation_type" validate:"omitempty,oneof=reorder excess normal critical"`
			Priority           string   `json:"priority" validate:"omitempty,oneof=low medium high critical"`
			Notes              string   `json:"notes"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		recType := req.RecommendationType
		if recType == "" {
			if req.CurrentStock <= req.ReorderPoint {
				recType = "reorder"
			} else if req.CurrentStock > req.ReorderPoint*3 {
				recType = "excess"
			} else {
				recType = "normal"
			}
		}

		priority := req.Priority
		if priority == "" {
			switch recType {
			case "critical":
				priority = "critical"
			case "reorder":
				priority = "high"
			case "excess":
				priority = "low"
			default:
				priority = "medium"
			}
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO reorder_recommendations (org_id, product_id, warehouse_id,
			 current_stock, reorder_point, reorder_quantity, lead_time_days,
			 amc_used, safety_stock, recommendation_type, priority, notes)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
			req.OrgID, req.ProductID, req.WarehouseID,
			req.CurrentStock, req.ReorderPoint, req.ReorderQuantity,
			req.LeadTimeDays, req.AMCUsed, req.SafetyStock,
			recType, priority, req.Notes,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create recommendation: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func MarkRecommendationReviewed(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := db.Exec(`UPDATE reorder_recommendations SET reviewed=true WHERE id=$1`, id)
		if err != nil {
			response.InternalError(c, "failed to update recommendation")
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			response.NotFound(c, "recommendation not found")
			return
		}

		response.OK(c, gin.H{"id": id})
	}
}

func ListForecastResults(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM forecast_results`)
		if err != nil {
			response.InternalError(c, "failed to count forecasts")
			return
		}

		var results []model.ForecastResult
		err = db.Select(&results,
			`SELECT id, org_id, product_id, warehouse_id, forecast_date, forecast_value,
			 lower_bound, upper_bound, confidence_level, model_version, features_used, created_at
			 FROM forecast_results ORDER BY forecast_date DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list forecasts")
			return
		}

		if results == nil {
			results = []model.ForecastResult{}
		}

		response.Paginated(c, results, total, p.Page, p.PageSize)
	}
}

func CreateForecastResult(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.ForecastResult
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO forecast_results (org_id, product_id, warehouse_id, forecast_date,
			 forecast_value, lower_bound, upper_bound, confidence_level, model_version, features_used)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
			req.OrgID, req.ProductID, req.WarehouseID, req.ForecastDate,
			req.ForecastValue, req.LowerBound, req.UpperBound,
			req.ConfidenceLevel, req.ModelVersion, req.FeaturesUsed,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create forecast: "+err.Error())
			_ = fmt.Sprintf("unique constraint violation: %v", err)
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}
