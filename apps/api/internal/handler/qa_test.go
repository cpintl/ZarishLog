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

var testBaseQA = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

const (
	testProdID = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a5b"
	testInspID = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a5c"
	testTplID  = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a5d"
	testItemID = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a5e"
	testDispID = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a5f"
)

func setupQARouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", testOrgID)
		c.Set("user_id", testUserID)
		c.Next()
	})
	r.POST("/api/v1/qa/inspections", CreateInspection(db))
	r.GET("/api/v1/qa/inspections", ListInspections(db))
	r.GET("/api/v1/qa/inspections/:id", GetInspection(db))
	r.POST("/api/v1/qa/inspections/:id/disposition", CreateDisposition(db))
	r.POST("/api/v1/qa/checklists", CreateChecklistTemplate(db))
	r.GET("/api/v1/qa/checklists", ListChecklistTemplates(db))
	r.GET("/api/v1/qa/checklists/:id", GetChecklistTemplate(db))
	r.GET("/api/v1/stock/expiring", GetExpiringStock(db))
	return r
}

func TestCreateInspection(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"product_id":"` + testProdID + `",
		"inspection_date":"2025-06-01",
		"inspector":"Dr. Smith",
		"result":"pass",
		"notes":"All good",
		"created_by":"` + testUserID + `",
		"results":[]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO qa_inspections`).
		WithArgs(testOrgID, nil, testProdID, nil, "2025-06-01", "Dr. Smith", "pass", "All good", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testInspID))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/qa/inspections", strings.NewReader(body))
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

func TestCreateInspectionWithResults(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"product_id":"` + testProdID + `",
		"inspection_date":"2025-06-01",
		"inspector":"Dr. Smith",
		"result":"pass",
		"notes":"",
		"created_by":"` + testUserID + `",
		"results":[
			{"checklist_item_id":"` + testItemID + `","answer":"yes","score":100,"notes":""}
		]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO qa_inspections`).
		WithArgs(testOrgID, nil, testProdID, nil, "2025-06-01", "Dr. Smith", "pass", "", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testInspID))
	mock.ExpectExec(`INSERT INTO qa_checklist_results`).
		WithArgs(testInspID, testItemID, "yes", float64(100), "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/qa/inspections", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateInspection_ValidationError(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	body := `{"org_id":"bad","product_id":"` + testProdID + `","inspection_date":"2025-06-01","inspector":"X","result":"invalid","created_by":"` + testUserID + `"}`

	req, _ := http.NewRequest("POST", "/api/v1/qa/inspections", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestListInspections(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM qa_inspections`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT (.+) FROM qa_inspections ORDER BY created_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "grn_id", "product_id", "batch_id", "inspection_date", "inspector", "result", "notes", "disposition", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testInspID, testOrgID, nil, testProdID, nil, "2025-06-01", "Dr. Smith", "pass", "", nil, testUserID, testUserID, testBaseQA, testBaseQA))

	req, _ := http.NewRequest("GET", "/api/v1/qa/inspections", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["total"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetInspection(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	mock.ExpectQuery(`SELECT (.+) FROM qa_inspections WHERE id=\$1`).
		WithArgs(testInspID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "grn_id", "product_id", "batch_id", "inspection_date", "inspector", "result", "notes", "disposition", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testInspID, testOrgID, nil, testProdID, nil, "2025-06-01", "Dr. Smith", "pass", "", nil, testUserID, testUserID, testBaseQA, testBaseQA))
	mock.ExpectQuery(`SELECT (.+) FROM qa_checklist_results WHERE inspection_id=\$1`).
		WithArgs(testInspID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "inspection_id", "checklist_item_id", "answer", "score", "notes", "created_at"}))
	mock.ExpectQuery(`SELECT (.+) FROM qa_dispositions WHERE inspection_id=\$1`).
		WithArgs(testInspID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "inspection_id", "disposition_type", "disposition_date", "approved_by", "destination_location_id", "notes", "created_by", "created_at", "updated_at"}))

	req, _ := http.NewRequest("GET", "/api/v1/qa/inspections/"+testInspID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetInspection_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	mock.ExpectQuery(`SELECT (.+) FROM qa_inspections WHERE id=\$1`).
		WithArgs(testInspID).
		WillReturnError(sqlmock.ErrCancelled)

	req, _ := http.NewRequest("GET", "/api/v1/qa/inspections/"+testInspID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateDisposition(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	body := `{
		"disposition_type":"pass",
		"disposition_date":"2025-06-02",
		"approved_by":"Dr. Smith",
		"notes":"Approved",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO qa_dispositions`).
		WithArgs(testInspID, "pass", "2025-06-02", "Dr. Smith", nil, "Approved", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testDispID))
	mock.ExpectExec(`UPDATE qa_inspections SET disposition`).
		WithArgs("pass", testUserID, testInspID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("POST", "/api/v1/qa/inspections/"+testInspID+"/disposition", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, testDispID, data["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateChecklistTemplate(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"code":"VISUAL-001",
		"name":"Visual Inspection",
		"category":"quality",
		"created_by":"` + testUserID + `",
		"items":[
			{"item_order":1,"question":"Is packaging intact?","expected_answer":"yes","is_critical":true,"weight":2.0}
		]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO qa_checklist_templates`).
		WithArgs(testOrgID, "VISUAL-001", "Visual Inspection", "", "quality", false, "active", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testTplID))
	mock.ExpectExec(`INSERT INTO qa_checklist_items`).
		WithArgs(testTplID, 1, "Is packaging intact?", "yes", true, float64(2.0)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/qa/checklists", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, testTplID, data["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListChecklistTemplates(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM qa_checklist_templates`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT (.+) FROM qa_checklist_templates ORDER BY created_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "code", "name", "description", "category", "is_mandatory", "status", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testTplID, testOrgID, "VISUAL-001", "Visual Inspection", "", "quality", false, "active", testUserID, testUserID, testBaseQA, testBaseQA))

	req, _ := http.NewRequest("GET", "/api/v1/qa/checklists", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetChecklistTemplate(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	mock.ExpectQuery(`SELECT (.+) FROM qa_checklist_templates WHERE id=\$1`).
		WithArgs(testTplID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "code", "name", "description", "category", "is_mandatory", "status", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testTplID, testOrgID, "VISUAL-001", "Visual Inspection", "", "quality", false, "active", testUserID, testUserID, testBaseQA, testBaseQA))
	mock.ExpectQuery(`SELECT (.+) FROM qa_checklist_items WHERE template_id=\$1 ORDER BY item_order`).
		WithArgs(testTplID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "template_id", "item_order", "question", "expected_answer", "is_critical", "weight"}).
			AddRow(testItemID, testTplID, 1, "Is packaging intact?", "yes", true, float64(2.0)))

	req, _ := http.NewRequest("GET", "/api/v1/qa/checklists/"+testTplID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExpiringStock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupQARouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM batches WHERE expiry_date IS NOT NULL AND expiry_date.+`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT b.id AS batch_id, b.product_id, b.batch_ref, b.expiry_date, COALESCE\(sl.quantity,0\) AS quantity FROM batches b.+`).
		WillReturnRows(sqlmock.NewRows([]string{"batch_id", "product_id", "batch_ref", "expiry_date", "quantity"}).
			AddRow(testLoc1, testProdID, "BATCH-001", "2025-07-01", float64(100)))

	req, _ := http.NewRequest("GET", "/api/v1/stock/expiring?days=30", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
