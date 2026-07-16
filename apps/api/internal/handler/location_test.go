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

var (
	testWHID = "00000000-0000-7000-8000-000000000001"
	testLoc1 = "00000000-0000-7000-8000-000000000011"
	testLoc2 = "00000000-0000-7000-8000-000000000012"
	testLoc3 = "00000000-0000-7000-8000-000000000013"
	testLoc4 = "00000000-0000-7000-8000-000000000014"
	testBase = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
)

func setupLocationRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/warehouses/:warehouse_id/locations", ListLocations(db))
	r.GET("/api/v1/warehouses/:warehouse_id/locations/tree", ListLocationTree(db))
	r.GET("/api/v1/warehouses/:warehouse_id/locations/:id", GetLocation(db))
	r.POST("/api/v1/warehouses/:warehouse_id/locations", CreateLocation(db))
	r.PUT("/api/v1/warehouses/:warehouse_id/locations/:id", UpdateLocation(db))
	r.DELETE("/api/v1/warehouses/:warehouse_id/locations/:id", DeleteLocation(db))
	r.GET("/api/v1/warehouses/:warehouse_id/locations/:id/constraints", GetLocationConstraints(db))
	r.PUT("/api/v1/warehouses/:warehouse_id/locations/:id/constraints", UpsertLocationConstraints(db))
	return r
}

func TestListLocations(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM locations WHERE warehouse_id = \$1`).
		WithArgs(testWHID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT \* FROM locations WHERE warehouse_id = \$1 ORDER BY code LIMIT \$2 OFFSET \$3`).
		WithArgs(testWHID, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "warehouse_id", "parent_id", "code", "name", "type", "is_cold_chain", "is_hazardous", "is_secure", "max_capacity", "is_active", "created_at", "updated_at"}).
			AddRow(testLoc1, testWHID, nil, "Z-01", "Zone A", "zone", false, false, false, nil, true, testBase, testBase).
			AddRow(testLoc2, testWHID, nil, "Z-02", "Zone B", "zone", true, false, true, 500.0, true, testBase, testBase))

	req, _ := http.NewRequest("GET", "/api/v1/warehouses/"+testWHID+"/locations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp["total"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLocation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	mock.ExpectQuery(`SELECT \* FROM locations WHERE id = \$1`).
		WithArgs(testLoc1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "warehouse_id", "parent_id", "code", "name", "type", "is_cold_chain", "is_hazardous", "is_secure", "max_capacity", "is_active", "created_at", "updated_at"}).
			AddRow(testLoc1, testWHID, nil, "Z-01", "Zone A", "zone", false, false, false, nil, true, testBase, testBase))

	req, _ := http.NewRequest("GET", "/api/v1/warehouses/"+testWHID+"/locations/"+testLoc1, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Zone A", data["name"])
	assert.Equal(t, "Z-01", data["code"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLocationNotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	mock.ExpectQuery(`SELECT \* FROM locations WHERE id = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sqlmock.ErrCancelled)

	req, _ := http.NewRequest("GET", "/api/v1/warehouses/"+testWHID+"/locations/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateLocation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	body := `{"warehouse_id":"` + testWHID + `","code":"Z-01","name":"Zone A","type":"zone","is_active":true}`
	mock.ExpectQuery(`INSERT INTO locations`).
		WithArgs(testWHID, sqlmock.AnyArg(), "Z-01", "Zone A", "zone", false, false, false, sqlmock.AnyArg(), true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(testLoc1, testBase, testBase))

	req, _ := http.NewRequest("POST", "/api/v1/warehouses/"+testWHID+"/locations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("CreateLocation response: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testLoc1, resp["data"].(map[string]interface{})["id"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateLocation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	body := `{"warehouse_id":"` + testWHID + `","code":"Z-01","name":"Zone A Updated","type":"zone","is_cold_chain":true,"is_active":true}`
	mock.ExpectExec(`UPDATE locations SET`).
		WithArgs(sqlmock.AnyArg(), "Z-01", "Zone A Updated", "zone", true, false, false, sqlmock.AnyArg(), true, testLoc1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("PUT", "/api/v1/warehouses/"+testWHID+"/locations/"+testLoc1, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteLocation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	mock.ExpectExec(`DELETE FROM locations WHERE id = \$1`).
		WithArgs(testLoc1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("DELETE", "/api/v1/warehouses/"+testWHID+"/locations/"+testLoc1, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLocationConstraints(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	mock.ExpectQuery(`SELECT \* FROM location_constraints WHERE location_id = \$1`).
		WithArgs(testLoc1).
		WillReturnRows(sqlmock.NewRows([]string{"location_id", "min_temperature", "max_temperature", "min_humidity", "max_humidity", "is_hazardous_allowed", "is_food_grade", "is_pharma_grade", "max_weight_capacity", "created_at", "updated_at"}).
			AddRow(testLoc1, 2.0, 8.0, nil, nil, false, false, true, nil, testBase, testBase))

	req, _ := http.NewRequest("GET", "/api/v1/warehouses/"+testWHID+"/locations/"+testLoc1+"/constraints", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["is_pharma_grade"])
	assert.Equal(t, 2.0, data["min_temperature"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertLocationConstraints(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	body := `{"min_temperature":2,"max_temperature":8,"is_pharma_grade":true}`
	mock.ExpectExec(`INSERT INTO location_constraints`).
		WithArgs(testLoc1, 2.0, 8.0, sqlmock.AnyArg(), sqlmock.AnyArg(), false, false, true, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("PUT", "/api/v1/warehouses/"+testWHID+"/locations/"+testLoc1+"/constraints", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListLocationTree(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupLocationRouter(db)

	mock.ExpectQuery(`SELECT \* FROM locations WHERE warehouse_id = \$1 ORDER BY type, code`).
		WithArgs(testWHID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "warehouse_id", "parent_id", "code", "name", "type", "is_cold_chain", "is_hazardous", "is_secure", "max_capacity", "is_active", "created_at", "updated_at"}).
			AddRow(testLoc1, testWHID, nil, "Z-01", "Zone A", "zone", false, false, false, nil, true, testBase, testBase).
			AddRow(testLoc2, testWHID, &testLoc1, "R-01", "Rack 1", "rack", false, false, false, nil, true, testBase, testBase))

	req, _ := http.NewRequest("GET", "/api/v1/warehouses/"+testWHID+"/locations/tree", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].([]interface{})
	require.Len(t, data, 1)
	zone := data[0].(map[string]interface{})
	assert.Equal(t, "Zone A", zone["name"])
	children := zone["children"].([]interface{})
	require.Len(t, children, 1)
	child := children[0].(map[string]interface{})
	assert.Equal(t, "Rack 1", child["name"])

	require.NoError(t, mock.ExpectationsWereMet())
}
