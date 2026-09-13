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
	testRegApprovalID = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a60"
	testDeviationID   = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a61"
	testCapaID        = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a62"
	testTrainingID    = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a63"
	testExcursionID   = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a64"
	testDonationID    = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a65"
	testLineItemID    = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a66"
	testComplaintID   = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a67"
	testRecallID      = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a68"
	testRecallLineID  = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a69"
	testWaybillID     = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a6a"
	testDeliveryID    = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a6b"
	testEmergencyID   = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a6c"
	testChangeID      = "018f2a3b-4c5d-7e7f-8a9b-0c1d2e3f4a6d"
)

func setupPolicyRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", testOrgID)
		c.Set("user_id", testUserID)
		c.Next()
	})
	reg := r.Group("/policy/regulatory-approvals")
	{
		reg.POST("", CreateRegulatoryApproval(db))
		reg.GET("", ListRegulatoryApprovals(db))
		reg.GET("/expiring", ListExpiringRegulatoryApprovals(db))
		reg.GET("/:id", GetRegulatoryApproval(db))
	}
	dev := r.Group("/policy/deviations")
	{
		dev.POST("", CreateDeviation(db))
		dev.GET("", ListDeviations(db))
		dev.GET("/:id", GetDeviation(db))
	}
	capa := r.Group("/policy/capa")
	{
		capa.POST("", CreateCAPAAction(db))
		capa.POST("/:id/effectiveness", VerifyCAPAEffectiveness(db))
	}
	r.POST("/temperature-excursions", CreateTemperatureExcursion(db))
	r.GET("/temperature-excursions", ListTemperatureExcursions(db))
	r.POST("/temperature-excursions/:id/disposition", UpdateTemperatureExcursionDisposition(db))
	r.POST("/donations", CreateDonation(db))
	r.GET("/donations/:id", GetDonation(db))
	r.POST("/donations/:id/decision", UpdateDonationDecision(db))
	r.POST("/complaints", CreateComplaint(db))
	r.POST("/recalls", CreateRecall(db))
	r.GET("/recalls/:id", GetRecall(db))
	r.POST("/dispatch-waybills", CreateDispatchWaybill(db))
	r.POST("/delivery-confirmations", CreateDeliveryConfirmation(db))
	r.POST("/emergency-plans", CreateEmergencyPlan(db))
	r.POST("/change-controls", CreateChangeControl(db))
	return r
}

func TestCreateRegulatoryApproval(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"scope_type":"warehouse",
		"scope_id":"` + testWH1 + `",
		"authority":"DGDA",
		"license_type":"wholesale_license",
		"reference_number":"L-1001",
		"status":"active",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO regulatory_approvals`).
		WithArgs(testOrgID, "warehouse", testWH1, "DGDA", "wholesale_license", "L-1001",
			nil, nil, nil, nil, 60, nil, "active", nil, nil, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testRegApprovalID))

	req, _ := http.NewRequest("POST", "/policy/regulatory-approvals", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testRegApprovalID, resp["data"].(map[string]interface{})["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateRegulatoryApproval_ValidationError(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{"org_id":"bad","scope_type":"nope","credential":"none"}`

	req, _ := http.NewRequest("POST", "/policy/regulatory-approvals", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestListRegulatoryApprovals(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM regulatory_approvals`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM regulatory_approvals ORDER BY created_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "authority", "license_type", "reference_number", "status"}).
			AddRow(testRegApprovalID, testOrgID, "DGDA", "wholesale_license", "L-1001", "active"))

	req, _ := http.NewRequest("GET", "/policy/regulatory-approvals", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["total"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRegulatoryApproval_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	mock.ExpectQuery(`SELECT \* FROM regulatory_approvals WHERE id=\$1`).
		WithArgs(testRegApprovalID).
		WillReturnError(sqlmock.ErrCancelled)

	req, _ := http.NewRequest("GET", "/policy/regulatory-approvals/"+testRegApprovalID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateDeviation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	occurred := "2025-06-01T08:00:00Z"
	body := `{
		"org_id":"` + testOrgID + `",
		"deviation_number":"DEV-001",
		"warehouse_id":"` + testWH1 + `",
		"description":"Received pallet with torn stretch wrap",
		"risk_rating":"medium",
		"status":"open",
		"occurred_at":"` + occurred + `",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO deviations`).
		WithArgs(testOrgID, "DEV-001", nil, occurred, nil, nil, nil, nil, testWH1, nil,
			"Received pallet with torn stretch wrap", nil, nil, nil, "medium", nil, "open", nil, nil,
			false, nil, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testDeviationID))

	req, _ := http.NewRequest("POST", "/policy/deviations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListDeviations(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM deviations`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM deviations ORDER BY occurred_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "deviation_number", "description", "risk_rating", "status"}).
			AddRow(testDeviationID, testOrgID, "DEV-001", "Torn pallet wrap", "medium", "open"))

	req, _ := http.NewRequest("GET", "/policy/deviations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDeviation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	mock.ExpectQuery(`SELECT \* FROM deviations WHERE id=\$1`).
		WithArgs(testDeviationID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "deviation_number", "description", "risk_rating", "status"}).
			AddRow(testDeviationID, testOrgID, "DEV-001", "Torn pallet wrap", "medium", "open"))

	req, _ := http.NewRequest("GET", "/policy/deviations/"+testDeviationID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCAPAAction(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"capa_number":"CAPA-001",
		"source_type":"deviation",
		"source_id":"` + testDeviationID + `",
		"capa_type":"corrective",
		"title":"Update receiving SOP",
		"action_plan":"Revise pallet handling steps",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO capa_actions`).
		WithArgs(testOrgID, "CAPA-001", "deviation", testDeviationID, "corrective", "Update receiving SOP",
			nil, "Revise pallet handling steps", nil, nil, "open", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testCapaID))

	req, _ := http.NewRequest("POST", "/policy/capa", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyCAPAEffectiveness(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"effectiveness_check":"Two receiving cycles with zero repeat deviations",
		"verified_by":"QA Lead",
		"closed_by":"QA Lead"
	}`

	mock.ExpectExec(`UPDATE capa_actions SET effectiveness_check=\$1`).
		WithArgs("Two receiving cycles with zero repeat deviations", "QA Lead", "QA Lead", testCapaID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("POST", "/policy/capa/"+testCapaID+"/effectiveness", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateDonation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"donation_number":"DON-001",
		"donor_name":"Global Health Fund",
		"offer_date":"2025-06-01",
		"status":"pending",
		"created_by":"` + testUserID + `",
		"line_items":[
			{"product_id":"` + testProdID + `","batch_number":"B-LOT-1","expiry_date":"2026-01-01","quantity":500,"uom":"box","authorized":true}
		]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO donations`).
		WithArgs(testOrgID, "DON-001", "Global Health Fund", nil, "2025-06-01", nil, nil, nil, nil, nil, nil, "pending", nil, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testDonationID))
	mock.ExpectExec(`INSERT INTO donation_line_items`).
		WithArgs(testDonationID, testProdID, "B-LOT-1", "2026-01-01", float64(500), "box", nil, nil, nil, true, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/donations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDonation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	mock.ExpectQuery(`SELECT \* FROM donations WHERE id=\$1`).
		WithArgs(testDonationID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "donation_number", "donor_name", "status"}).
			AddRow(testDonationID, testOrgID, "DON-001", "Global Health Fund", "pending"))
	mock.ExpectQuery(`SELECT \* FROM donation_line_items WHERE donation_id=\$1`).
		WithArgs(testDonationID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "donation_id", "product_id", "batch_number", "quantity", "authorized"}).
			AddRow(testLineItemID, testDonationID, testProdID, "B-LOT-1", float64(500), true))

	req, _ := http.NewRequest("GET", "/donations/"+testDonationID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testLineItemID, resp["data"].(map[string]interface{})["line_items"].([]interface{})[0].(map[string]interface{})["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateDonationDecision(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"decision":"accepted",
		"decision_date":"2025-06-02",
		"decided_by":"QA Lead",
		"certificate_number":"CERT-2025-01",
		"updated_by":"` + testUserID + `"
	}`

	mock.ExpectExec(`UPDATE donations SET decision=\$1, status=\$1`).
		WithArgs("accepted", "2025-06-02", "QA Lead", "CERT-2025-01", nil, testUserID, testDonationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("POST", "/donations/"+testDonationID+"/decision", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateComplaint(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"complaint_number":"CMP-001",
		"received_date":"2025-06-01",
		"severity":"MAJOR",
		"description":"Label typo on secondary packaging",
		"category":"packaging",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO complaints`).
		WithArgs(testOrgID, "CMP-001", "2025-06-01", nil, "MAJOR", "packaging", nil, nil,
			"Label typo on secondary packaging", nil, nil, false, false, nil, false, false,
			"open", nil, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testComplaintID))

	req, _ := http.NewRequest("POST", "/complaints", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateRecall(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"recall_number":"REC-001",
		"recall_type":"MOCK",
		"recall_date":"2025-06-01",
		"subject_product":"Amoxicillin 500mg",
		"batch_numbers":"B-AMOX-1",
		"reason":"Annual mock recall exercise",
		"status":"in_progress",
		"created_by":"` + testUserID + `",
		"line_items":[
			{"recipient":"Chittagong Clinic","quantity_issued":200,"quantity_recovered":150,"storage_disposition":"quarantined"}
		]
	}`

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO recalls`).
		WithArgs(testOrgID, "REC-001", "MOCK", "2025-06-01", nil, "Amoxicillin 500mg", "B-AMOX-1",
			"Annual mock recall exercise", nil, nil, nil, nil, "in_progress", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testRecallID))
	mock.ExpectExec(`INSERT INTO recall_line_items`).
		WithArgs(testRecallID, "Chittagong Clinic", nil, nil, nil, float64(200), float64(150), float64(50), "quarantined", nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/recalls", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRecall(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	mock.ExpectQuery(`SELECT \* FROM recalls WHERE id=\$1`).
		WithArgs(testRecallID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "recall_number", "recall_type", "subject_product", "status"}).
			AddRow(testRecallID, testOrgID, "REC-001", "MOCK", "Amoxicillin 500mg", "in_progress"))
	mock.ExpectQuery(`SELECT \* FROM recall_line_items WHERE recall_id=\$1`).
		WithArgs(testRecallID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "recall_id", "recipient", "quantity_issued", "quantity_recovered", "quantity_outstanding", "storage_disposition"}).
			AddRow(testRecallLineID, testRecallID, "Chittagong Clinic", float64(200), float64(150), float64(50), "quarantined"))

	req, _ := http.NewRequest("GET", "/recalls/"+testRecallID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTemperatureExcursion(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	started := "2025-06-01T03:00:00Z"
	body := `{
		"org_id":"` + testOrgID + `",
		"warehouse_id":"` + testWH1 + `",
		"excursion_type":"freeze",
		"started_at":"` + started + `",
		"min_value":-2.0,
		"max_value":0.5,
		"expected_min":2.0,
		"expected_max":8.0,
		"quarantined":true,
		"disposition":"pending",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO temperature_excursions`).
		WithArgs(testOrgID, testWH1, nil, nil, nil, "freeze", started, nil,
			float64(-2), float64(0.5), float64(2), float64(8), nil, nil, nil,
			true, nil, "pending", false, false, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testExcursionID))

	req, _ := http.NewRequest("POST", "/temperature-excursions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTemperatureExcursionDisposition(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"disposition":"destroyed",
		"technical_advice":"Cold-chain product outside 2-8C for 6h; destroy per SOP",
		"root_cause":"compressor relay failure",
		"updated_by":"QA Lead"
	}`

	mock.ExpectExec(`UPDATE temperature_excursions SET disposition=\$1, status=\$2`).
		WithArgs("destroyed", "dispositioned", "Cold-chain product outside 2-8C for 6h; destroy per SOP",
			"compressor relay failure", nil, nil, "QA Lead", testExcursionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("POST", "/temperature-excursions/"+testExcursionID+"/disposition", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateDispatchWaybill(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"waybill_number":"WB-001",
		"sender_warehouse_id":"` + testWH1 + `",
		"dispatch_date":"2025-06-01",
		"carrier":"TransFast",
		"vehicle_number":"DHA-12-3456",
		"temperature_sensitive":true,
		"status":"dispatched",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO dispatch_waybills`).
		WithArgs(testOrgID, "WB-001", nil, nil, testWH1, nil, nil, "2025-06-01",
			"TransFast", "DHA-12-3456", nil, nil, nil, true, nil, nil, "dispatched", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testWaybillID))

	req, _ := http.NewRequest("POST", "/dispatch-waybills", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateDeliveryConfirmation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"waybill_id":"` + testWaybillID + `",
		"confirmed_date":"2025-06-02",
		"delivery_status":"received",
		"items_conformed":true,
		"quantity_conformed":true,
		"batch_conformed":true,
		"expiry_conformed":true,
		"condition_conformed":true,
		"temperature_conformed":true,
		"confirmed_by":"Storekeeper"
	}`

	mock.ExpectQuery(`INSERT INTO delivery_confirmations`).
		WithArgs(testOrgID, testWaybillID, nil, nil, "2025-06-02", "received",
			true, true, true, true, true, true, nil, nil, nil, nil, "Storekeeper").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testDeliveryID))
	mock.ExpectExec(`UPDATE dispatch_waybills SET status='delivered'`).
		WithArgs(testWaybillID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("POST", "/delivery-confirmations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateEmergencyPlan(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"plan_number":"EP-001",
		"name":"Cyclone Response Plan",
		"country_code":"BD",
		"scenario_type":"cyclone",
		"prepositioned_stock_plan":"Pre-position 30 days of ORS and antibiotics",
		"paper_fallback_records":true,
		"status":"approved",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO emergency_plans`).
		WithArgs(testOrgID, "EP-001", "Cyclone Response Plan", "BD", nil, "cyclone", nil, nil, nil,
			"Pre-position 30 days of ORS and antibiotics", nil, nil, nil, nil, nil, nil, true, nil,
			nil, nil, nil, nil, nil, "approved", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testEmergencyID))

	req, _ := http.NewRequest("POST", "/emergency-plans", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateChangeControl(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"change_number":"CC-001",
		"subject_type":"warehouse_layout",
		"description":"Reorganise cold-room racking",
		"risk_level":"low",
		"requires_quality_review":true,
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO change_controls`).
		WithArgs(testOrgID, "CC-001", "warehouse_layout", nil, "Reorganise cold-room racking",
			nil, "low", true, nil, nil, nil, nil, "draft", testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testChangeID))

	req, _ := http.NewRequest("POST", "/change-controls", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListTemperatureExcursions(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM temperature_excursions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM temperature_excursions ORDER BY started_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "warehouse_id", "excursion_type", "disposition", "status"}).
			AddRow(testExcursionID, testOrgID, testWH1, "freeze", "pending", "open"))

	req, _ := http.NewRequest("GET", "/temperature-excursions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateDonation_ValidationError(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	body := `{"org_id":123,"donation_number":"","donor_name":""}`

	req, _ := http.NewRequest("POST", "/donations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestExpiringRegulatoryApprovals(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupPolicyRouter(db)

	cutoff := time.Now().AddDate(0, 0, 60).Format("2006-01-02")

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM regulatory_approvals WHERE expiry_date IS NOT NULL AND expiry_date <= \$1`).
		WithArgs(cutoff).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM regulatory_approvals WHERE expiry_date IS NOT NULL AND expiry_date <= \$1 ORDER BY expiry_date ASC LIMIT .+ OFFSET .+`).
		WithArgs(cutoff, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "authority", "license_type", "reference_number", "status"}).
			AddRow(testRegApprovalID, testOrgID, "DGDA", "wholesale_license", "L-1001", "active"))

	req, _ := http.NewRequest("GET", "/policy/regulatory-approvals/expiring?days=60", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
