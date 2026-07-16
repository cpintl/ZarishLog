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
	testOrgID   = "00000000-0000-7000-8000-000000000101"
	testUserID  = "00000000-0000-7000-8000-000000000102"
	testWH1     = "00000000-0000-7000-8000-000000000111"
	testWH2     = "00000000-0000-7000-8000-000000000112"
	testWH3     = "00000000-0000-7000-8000-000000000113"
	testBaseWH  = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
)

func setupWarehouseRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", testOrgID)
		c.Set("user_id", testUserID)
		c.Next()
	})
	r.GET("/api/v1/warehouses", ListWarehouses(db))
	r.GET("/api/v1/warehouses/:id", GetWarehouse(db))
	r.POST("/api/v1/warehouses", CreateWarehouse(db))
	r.PUT("/api/v1/warehouses/:id", UpdateWarehouse(db))
	r.DELETE("/api/v1/warehouses/:id", DeleteWarehouse(db))
	return r
}

func TestListWarehouses(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupWarehouseRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM warehouses WHERE org_id = \$1`).
		WithArgs(testOrgID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT \* FROM warehouses WHERE org_id = \$1 ORDER BY name LIMIT \$2 OFFSET \$3`).
		WithArgs(testOrgID, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "name", "code", "type", "address", "city", "country", "is_active", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testWH1, testOrgID, "Central WH", "CWH", "central", "123 Main St", "Cox's Bazar", "Bangladesh", true, testUserID, testUserID, testBaseWH, testBaseWH).
			AddRow(testWH2, testOrgID, "Sub WH", "SWH", "sub_warehouse", "456 Side St", "Cox's Bazar", "Bangladesh", true, testUserID, testUserID, testBaseWH, testBaseWH))

	req, _ := http.NewRequest("GET", "/api/v1/warehouses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp["total"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWarehouse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupWarehouseRouter(db)

	mock.ExpectQuery(`SELECT \* FROM warehouses WHERE id = \$1 AND org_id = \$2`).
		WithArgs(testWH1, testOrgID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "name", "code", "type", "address", "city", "country", "is_active", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow(testWH1, testOrgID, "Central WH", "CWH", "central", "123 Main St", "Cox's Bazar", "Bangladesh", true, testUserID, testUserID, testBaseWH, testBaseWH))

	req, _ := http.NewRequest("GET", "/api/v1/warehouses/"+testWH1, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Central WH", data["name"])
	assert.Equal(t, "CWH", data["code"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWarehouseNotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupWarehouseRouter(db)

	mock.ExpectQuery(`SELECT \* FROM warehouses WHERE id = \$1 AND org_id = \$2`).
		WithArgs("nonexistent", testOrgID).
		WillReturnError(sqlmock.ErrCancelled)

	req, _ := http.NewRequest("GET", "/api/v1/warehouses/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWarehouse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupWarehouseRouter(db)

	body := `{"name":"New WH","code":"NWH","type":"central","city":"City","country":"Country"}`
	mock.ExpectQuery(`INSERT INTO warehouses`).
		WithArgs(testOrgID, "New WH", "NWH", "central", "", "City", "Country", false, testUserID, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(testWH3, testBaseWH, testBaseWH))

	req, _ := http.NewRequest("POST", "/api/v1/warehouses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testWH3, resp["data"].(map[string]interface{})["id"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateWarehouse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupWarehouseRouter(db)

	body := `{"name":"Updated WH","code":"UWH","type":"central","city":"City","country":"Country","is_active":true}`
	mock.ExpectExec(`UPDATE warehouses SET`).
		WithArgs("Updated WH", "UWH", "central", "", "City", "Country", true, testUserID, testWH1, testOrgID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("PUT", "/api/v1/warehouses/"+testWH1, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteWarehouse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupWarehouseRouter(db)

	mock.ExpectExec(`DELETE FROM warehouses WHERE id = \$1 AND org_id = \$2`).
		WithArgs(testWH1, testOrgID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("DELETE", "/api/v1/warehouses/"+testWH1, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}
