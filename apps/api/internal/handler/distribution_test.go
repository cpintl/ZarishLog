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
	testDistID = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f7a5b"
)

var testBaseDist = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func setupDistributionRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", testOrgID)
		c.Set("user_id", testUserID)
		c.Next()
	})
	r.POST("/api/v1/distributions", CreateDistribution(db))
	r.GET("/api/v1/distributions", ListDistributions(db))
	r.GET("/api/v1/distributions/:id", GetDistribution(db))
	return r
}

func TestCreateDistribution(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupDistributionRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"distribution_number":"DIST-001",
		"distribution_date":"2025-06-01",
		"location":"Camp 4",
		"created_by":"` + testUserID + `",
		"items":[],
		"beneficiaries":[]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO distributions`).
		WithArgs(testOrgID, nil, nil, "DIST-001", "2025-06-01", "Camp 4", nil, "draft", "", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testDistID))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/distributions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, testDistID, data["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateDistributionWithItems(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupDistributionRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"distribution_number":"DIST-002",
		"distribution_date":"2025-06-01",
		"created_by":"` + testUserID + `",
		"items":[
			{"product_id":"` + testProdID + `","quantity_planned":100,"quantity_distributed":0}
		],
		"beneficiaries":[
			{"beneficiary_type":"family","count":50,"criteria":"vulnerable"}
		]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO distributions`).
		WithArgs(testOrgID, nil, nil, "DIST-002", "2025-06-01", "", nil, "draft", "", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testDistID))
	mock.ExpectExec(`INSERT INTO distribution_line_items`).
		WithArgs(testDistID, testProdID, nil, float64(100), float64(0), nil, "active").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO distribution_beneficiaries`).
		WithArgs(testDistID, "family", 50, "vulnerable").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/distributions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateDistribution_ValidationError(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupDistributionRouter(db)

	body := `{"org_id":"bad","distribution_number":"","distribution_date":"bad","created_by":""}`

	req, _ := http.NewRequest("POST", "/api/v1/distributions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestListDistributions(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupDistributionRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM distributions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT (.+) FROM distributions ORDER BY created_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "org_level_id", "program_id", "distribution_number", "distribution_date", "location", "beneficiary_count", "status", "notes", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testDistID, testOrgID, nil, nil, "DIST-001", "2025-06-01", "Camp 4", nil, "draft", "", testUserID, testUserID, testBaseDist, testBaseDist))

	req, _ := http.NewRequest("GET", "/api/v1/distributions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDistribution(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupDistributionRouter(db)

	mock.ExpectQuery(`SELECT (.+) FROM distributions WHERE id=\$1`).
		WithArgs(testDistID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "org_level_id", "program_id", "distribution_number", "distribution_date", "location", "beneficiary_count", "status", "notes", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testDistID, testOrgID, nil, nil, "DIST-001", "2025-06-01", "Camp 4", nil, "draft", "", testUserID, testUserID, testBaseDist, testBaseDist))
	mock.ExpectQuery(`SELECT (.+) FROM distribution_line_items WHERE distribution_id=\$1`).
		WithArgs(testDistID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "distribution_id", "product_id", "batch_id", "quantity_planned", "quantity_distributed", "unit_cost", "status"}))
	mock.ExpectQuery(`SELECT (.+) FROM distribution_beneficiaries WHERE distribution_id=\$1`).
		WithArgs(testDistID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "distribution_id", "beneficiary_type", "count", "criteria"}))

	req, _ := http.NewRequest("GET", "/api/v1/distributions/"+testDistID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDistribution_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupDistributionRouter(db)

	mock.ExpectQuery(`SELECT (.+) FROM distributions WHERE id=\$1`).
		WithArgs(testDistID).
		WillReturnError(sqlmock.ErrCancelled)

	req, _ := http.NewRequest("GET", "/api/v1/distributions/"+testDistID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
