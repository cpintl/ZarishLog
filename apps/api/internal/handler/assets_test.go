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

const (
	testAssetID = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f6a5b"
	testMaintID = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f6a5c"
)

var testBaseAsset = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func setupAssetRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", testOrgID)
		c.Set("user_id", testUserID)
		c.Next()
	})
	r.POST("/api/v1/assets", CreateAsset(db))
	r.GET("/api/v1/assets", ListAssets(db))
	r.GET("/api/v1/assets/:id", GetAsset(db))
	r.PUT("/api/v1/assets/:id", UpdateAsset(db))
	r.DELETE("/api/v1/assets/:id", DeleteAsset(db))
	r.POST("/api/v1/assets/:id/custody", TransferCustody(db))
	r.POST("/api/v1/assets/:id/maintenance", CreateAssetMaintenance(db))
	r.GET("/api/v1/assets/:id/maintenance", ListAssetMaintenance(db))
	return r
}

func TestCreateAsset(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"asset_tag":"AST-001",
		"name":"Refrigerator",
		"status":"in_use",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO assets`).
		WithArgs(testOrgID, "AST-001", "Refrigerator", "", nil, "", nil, nil, nil, nil, nil, "", nil, "in_use", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testAssetID))

	req, _ := http.NewRequest("POST", "/api/v1/assets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, testAssetID, data["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAsset_ValidationError(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	body := `{"org_id":"bad","asset_tag":"","name":"","status":"invalid","created_by":""}`

	req, _ := http.NewRequest("POST", "/api/v1/assets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestListAssets(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM assets`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT (.+) FROM assets ORDER BY created_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "asset_tag", "name", "description", "product_id", "serial_number", "custodian_id", "location_id", "acquisition_date", "purchase_cost", "current_value", "depreciation_method", "useful_life_years", "status", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testAssetID, testOrgID, "AST-001", "Refrigerator", "", nil, "", nil, nil, nil, nil, nil, "", nil, "in_use", testUserID, testUserID, testBaseAsset, testBaseAsset))

	req, _ := http.NewRequest("GET", "/api/v1/assets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAsset(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	mock.ExpectQuery(`SELECT (.+) FROM assets WHERE id=\$1`).
		WithArgs(testAssetID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "asset_tag", "name", "description", "product_id", "serial_number", "custodian_id", "location_id", "acquisition_date", "purchase_cost", "current_value", "depreciation_method", "useful_life_years", "status", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testAssetID, testOrgID, "AST-001", "Refrigerator", "", nil, "", nil, nil, nil, nil, nil, "", nil, "in_use", testUserID, testUserID, testBaseAsset, testBaseAsset))
	mock.ExpectQuery(`SELECT (.+) FROM asset_custody_changes WHERE asset_id=\$1 ORDER BY changed_at DESC`).
		WithArgs(testAssetID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "asset_id", "from_user_id", "to_user_id", "changed_by", "changed_at"}))
	mock.ExpectQuery(`SELECT (.+) FROM asset_maintenance WHERE asset_id=\$1 ORDER BY maintenance_date DESC`).
		WithArgs(testAssetID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "asset_id", "maintenance_date", "description", "cost", "performed_by", "next_date", "created_at"}))

	req, _ := http.NewRequest("GET", "/api/v1/assets/"+testAssetID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAsset_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	mock.ExpectQuery(`SELECT (.+) FROM assets WHERE id=\$1`).
		WithArgs(testAssetID).
		WillReturnError(sqlmock.ErrCancelled)

	req, _ := http.NewRequest("GET", "/api/v1/assets/"+testAssetID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateAsset(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	body := `{
		"asset_tag":"AST-001",
		"name":"Refrigerator Updated",
		"status":"under_maintenance",
		"updated_by":"` + testUserID + `"
	}`

	mock.ExpectExec(`UPDATE assets SET`).
		WithArgs("AST-001", "Refrigerator Updated", "", nil, "", nil, nil, nil, nil, nil, "", nil, "under_maintenance", testUserID, testAssetID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("PUT", "/api/v1/assets/"+testAssetID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAsset(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	mock.ExpectExec(`DELETE FROM assets WHERE id=\$1`).
		WithArgs(testAssetID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("DELETE", "/api/v1/assets/"+testAssetID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAsset_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	mock.ExpectExec(`DELETE FROM assets WHERE id=\$1`).
		WithArgs(testAssetID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req, _ := http.NewRequest("DELETE", "/api/v1/assets/"+testAssetID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTransferCustody(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	body := `{"to_user_id":"` + testUserID + `","changed_by":"` + testUserID + `"}`

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT custodian_id FROM assets WHERE`).
		WithArgs(testAssetID).
		WillReturnRows(sqlmock.NewRows([]string{"custodian_id"}).AddRow(nil))
	mock.ExpectQuery(`INSERT INTO asset_custody_changes`).
		WithArgs(testAssetID, nil, testUserID, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testInspID))
	mock.ExpectExec(`UPDATE assets SET custodian_id`).
		WithArgs(testUserID, testUserID, testAssetID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/assets/"+testAssetID+"/custody", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAssetMaintenance(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	body := `{
		"maintenance_date":"2025-06-15",
		"description":"Annual service",
		"performed_by":"Tech A",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO asset_maintenance`).
		WithArgs(testAssetID, "2025-06-15", "Annual service", nil, "Tech A", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testMaintID))

	req, _ := http.NewRequest("POST", "/api/v1/assets/"+testAssetID+"/maintenance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, testMaintID, data["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListAssetMaintenance(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupAssetRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM asset_maintenance WHERE asset_id=\$1`).
		WithArgs(testAssetID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT (.+) FROM asset_maintenance WHERE asset_id=\$1 ORDER BY maintenance_date DESC LIMIT .+ OFFSET .+`).
		WithArgs(testAssetID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "asset_id", "maintenance_date", "description", "cost", "performed_by", "next_date", "created_at"}).
			AddRow(testMaintID, testAssetID, "2025-06-15", "Annual service", nil, "Tech A", nil, testBaseAsset))

	req, _ := http.NewRequest("GET", "/api/v1/assets/"+testAssetID+"/maintenance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
