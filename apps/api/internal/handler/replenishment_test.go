package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReplenishmentRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", testOrgID)
		c.Set("user_id", testUserID)
		c.Next()
	})
	r.POST("/api/v1/replenishment/amc", CalculateAMC(db))
	r.GET("/api/v1/replenishment/amc", ListAMCCalculations(db))
	r.GET("/api/v1/replenishment/amc/latest", GetLatestAMC(db))
	r.GET("/api/v1/replenishment/recommendations", ListReorderRecommendations(db))
	r.POST("/api/v1/replenishment/recommendations", CreateReorderRecommendation(db))
	r.PUT("/api/v1/replenishment/recommendations/:id/review", MarkRecommendationReviewed(db))
	r.GET("/api/v1/replenishment/forecasts", ListForecastResults(db))
	r.POST("/api/v1/replenishment/forecasts", CreateForecastResult(db))
	return r
}

func TestCalculateAMC(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupReplenishmentRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"product_id":"` + testProdID + `",
		"warehouse_id":"` + testWHID + `",
		"created_by":"` + testUserID + `"
	}`

	for i := 0; i < 4; i++ {
		mock.ExpectQuery(`SELECT to_char`).
			WithArgs(testOrgID, testProdID, testWHID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"month", "qty"}).
				AddRow("2025-01", 100.0).AddRow("2025-02", 95.0))
	}
	mock.ExpectQuery(`INSERT INTO amc_calculations`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testInspID))

	req, _ := http.NewRequest("POST", "/api/v1/replenishment/amc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListAMCCalculations(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupReplenishmentRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM amc_calculations`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT (.+) FROM amc_calculations ORDER BY created_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "product_id", "warehouse_id", "calculation_date", "amc_3_months", "amc_6_months", "amc_12_months", "max_consumption", "std_deviation", "calculation_period_start", "calculation_period_end", "calculation_status", "created_at"}).
			AddRow(testInspID, testOrgID, testProdID, testWHID, "2025-06-01", 100.0, 95.0, 90.0, 120.0, 15.0, "2025-03-01", "2025-06-01", "completed", testBaseQA))

	req, _ := http.NewRequest("GET", "/api/v1/replenishment/amc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLatestAMC_MissingParams(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupReplenishmentRouter(db)

	req, _ := http.NewRequest("GET", "/api/v1/replenishment/amc/latest", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateReorderRecommendation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupReplenishmentRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"product_id":"` + testProdID + `",
		"warehouse_id":"` + testWHID + `",
		"current_stock":50,
		"reorder_point":100,
		"reorder_quantity":200
	}`

	mock.ExpectQuery(`INSERT INTO reorder_recommendations`).
		WithArgs(testOrgID, testProdID, testWHID, float64(50), float64(100), float64(200), nil, nil, nil, "reorder", "high", "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testInspID))

	req, _ := http.NewRequest("POST", "/api/v1/replenishment/recommendations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, testInspID, data["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListReorderRecommendations(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupReplenishmentRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM reorder_recommendations`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT rr.\*, p\.name AS product_name, p\.sku AS product_sku FROM reorder_recommendations rr JOIN products p ON rr\.product_id = p\.id ORDER BY rr\.created_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "product_id", "warehouse_id", "recommendation_date", "current_stock", "reorder_point", "reorder_quantity", "lead_time_days", "amc_used", "safety_stock", "recommendation_type", "priority", "notes", "reviewed", "created_at", "product_name", "product_sku"}).
			AddRow(testInspID, testOrgID, testProdID, testWHID, "2025-06-01", float64(50), float64(100), float64(200), nil, nil, nil, "reorder", "high", "", false, testBaseQA, "Paracetamol", "MED-PAR-500"))

	req, _ := http.NewRequest("GET", "/api/v1/replenishment/recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkRecommendationReviewed(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupReplenishmentRouter(db)

	mock.ExpectExec(`UPDATE reorder_recommendations SET reviewed=true WHERE id=\$1`).
		WithArgs(testInspID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("PUT", "/api/v1/replenishment/recommendations/"+testInspID+"/review", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateForecastResult(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupReplenishmentRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"product_id":"` + testProdID + `",
		"warehouse_id":"` + testWHID + `",
		"forecast_date":"2025-07-01",
		"forecast_value":850.5,
		"confidence_level":95
	}`

	mock.ExpectQuery(`INSERT INTO forecast_results`).
		WithArgs(testOrgID, testProdID, testWHID, "2025-07-01", float64(850.5), nil, nil, float64(95), "", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testInspID))

	req, _ := http.NewRequest("POST", "/api/v1/replenishment/forecasts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListForecastResults(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupReplenishmentRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM forecast_results`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT (.+) FROM forecast_results ORDER BY forecast_date DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "product_id", "warehouse_id", "forecast_date", "forecast_value", "lower_bound", "upper_bound", "confidence_level", "model_version", "features_used", "created_at"}).
			AddRow(testInspID, testOrgID, testProdID, testWHID, "2025-07-01", float64(850.5), nil, nil, float64(95.0), "", nil, testBaseQA))

	req, _ := http.NewRequest("GET", "/api/v1/replenishment/forecasts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
