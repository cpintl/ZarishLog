package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStockRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", testOrgID)
		c.Set("user_id", testUserID)
		c.Next()
	})
	r.POST("/api/v1/stock/transfer", CreateTransfer(db))
	r.POST("/api/v1/stock/adjust", CreateAdjustment(db))
	r.GET("/api/v1/stock/batches/:id/trail", GetBatchTrail(db))
	return r
}

func TestCreateTransfer(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupStockRouter(db)

	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	body := `{
		"transfer": {
			"from_warehouse_id":"` + testWHID + `",
			"to_warehouse_id":"` + testLoc1 + `",
			"transfer_number":"TFR-001",
			"status":"draft"
		},
		"items": [
			{"product_id":"` + testLoc1 + `","batch_id":"` + testLoc2 + `","quantity":50,"unit_cost":10.5}
		]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO stock_transfers`).
		WithArgs(testOrgID, testWHID, testLoc1, "TFR-001", "draft", testUserID, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(testLoc3, now, now))
	mock.ExpectExec(`INSERT INTO transfer_line_items`).
		WithArgs(testLoc3, testLoc1, testLoc2, float64(50), float64(10.5)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/stock/transfer", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Transfer response: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testLoc3, resp["data"].(map[string]interface{})["id"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransferValidation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupStockRouter(db)

	body := `{"transfer":{"from_warehouse_id":""},"items":[]}`

	req, _ := http.NewRequest("POST", "/api/v1/stock/transfer", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAdjustment(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupStockRouter(db)

	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	body := `{
		"adjustment": {
			"warehouse_id":"` + testWHID + `",
			"reason_code":"damage",
			"description":"Damaged goods"
		},
		"items": [
			{
				"product_id":"` + testLoc1 + `",
				"expected_quantity":100,
				"actual_quantity":95,
				"unit_cost":10.0,
				"notes":"5 units damaged"
			}
		]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO stock_adjustments`).
		WithArgs(testOrgID, testWHID, "damage", "Damaged goods", "draft", testUserID, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(testLoc4, now, now))
	mock.ExpectExec(`INSERT INTO adjustment_line_items`).
		WithArgs(testLoc4, testLoc1, nil, nil, float64(100), float64(95), float64(-5), nil, float64(10.0), "5 units damaged").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/stock/adjust", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Adjustment response: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testLoc4, resp["data"].(map[string]interface{})["id"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBatchTrail(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupStockRouter(db)

	mock.ExpectQuery(`SELECT movement_type, warehouse_id, quantity, ref_doc_type, ref_doc_id, created_by, created_at FROM stock_movements WHERE org_id = \$1 AND batch_id = \$2 ORDER BY created_at ASC`).
		WithArgs(testOrgID, testLoc1).
		WillReturnRows(sqlmock.NewRows([]string{"movement_type", "warehouse_id", "quantity", "ref_doc_type", "ref_doc_id", "created_by", "created_at"}).
			AddRow("receipt", testWHID, 100.0, "grn", "grn-1", testUserID, "2025-01-01T00:00:00Z").
			AddRow("issue", testWHID, -10.0, "srf", "srf-1", testUserID, "2025-01-02T00:00:00Z"))

	req, _ := http.NewRequest("GET", "/api/v1/stock/batches/"+testLoc1+"/trail", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].([]interface{})
	require.Len(t, data, 2)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBatchTrailNotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupStockRouter(db)

	mock.ExpectQuery(`SELECT movement_type, warehouse_id, quantity, ref_doc_type, ref_doc_id, created_by, created_at FROM stock_movements WHERE org_id = \$1 AND batch_id = \$2 ORDER BY created_at ASC`).
		WithArgs(testOrgID, "nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"movement_type", "warehouse_id", "quantity", "ref_doc_type", "ref_doc_id", "created_by", "created_at"}))

	req, _ := http.NewRequest("GET", "/api/v1/stock/batches/nonexistent/trail", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}
