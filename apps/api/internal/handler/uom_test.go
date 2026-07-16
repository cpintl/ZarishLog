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

func setupUoMRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/uoms", ListUoMs(db))
	r.GET("/api/v1/uoms/:id", GetUoM(db))
	r.POST("/api/v1/uoms", CreateUoM(db))
	r.PUT("/api/v1/uoms/:id", UpdateUoM(db))
	r.DELETE("/api/v1/uoms/:id", DeleteUoM(db))
	return r
}

func TestListUoMs(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUoMRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM units_of_measure`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT \* FROM units_of_measure ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "abbreviation", "category", "base_uom_id", "conversion_factor", "status", "created_at", "updated_at"}).
			AddRow("uom-1", "Tablet", "tab", "dosage", "base-uom-1", 1.0, "active", now, now).
			AddRow("uom-2", "Bottle", "bot", "packaging", "base-uom-2", 0.5, "active", now, now))

	req, _ := http.NewRequest("GET", "/api/v1/uoms", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("ListUoMs response: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	// PaginatedBody has Data, Total, Page, PageSize, TotalPages — no "status" field
	assert.Equal(t, float64(2), resp["total"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUoM(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUoMRouter(db)

	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT \* FROM units_of_measure WHERE id = \$1`).
		WithArgs("uom-123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "abbreviation", "category", "base_uom_id", "conversion_factor", "status", "created_at", "updated_at"}).
			AddRow("uom-123", "Tablet", "tab", "dosage", "base-uom-1", 1.0, "active", now, now))

	req, _ := http.NewRequest("GET", "/api/v1/uoms/uom-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	// response.OK returns SuccessBody{Data: ...} — no "status" field
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Tablet", data["name"])
	assert.Equal(t, "tab", data["abbreviation"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUoMNotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUoMRouter(db)

	mock.ExpectQuery(`SELECT \* FROM units_of_measure WHERE id = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sqlmock.ErrCancelled)

	req, _ := http.NewRequest("GET", "/api/v1/uoms/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUoM(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUoMRouter(db)

	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	body := `{"name":"Tablet","abbreviation":"tab","category":"dosage","status":"active"}`
	mock.ExpectQuery(`INSERT INTO units_of_measure`).
		WithArgs("Tablet", "tab", "dosage", nil, nil, "active").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("uom-new", now, now))

	req, _ := http.NewRequest("POST", "/api/v1/uoms", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "uom-new", resp["data"].(map[string]interface{})["id"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUoMValidation(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUoMRouter(db)

	body := `{"name":"","abbreviation":"","category":"invalid"}`

	req, _ := http.NewRequest("POST", "/api/v1/uoms", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUoM(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUoMRouter(db)

	body := `{"name":"Updated Tab","abbreviation":"tab","category":"dosage","status":"active"}`
	mock.ExpectExec(`UPDATE units_of_measure SET`).
		WithArgs("Updated Tab", "tab", "dosage", nil, nil, "active", "uom-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("PUT", "/api/v1/uoms/uom-123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUoM(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUoMRouter(db)

	mock.ExpectExec(`DELETE FROM units_of_measure WHERE id = \$1`).
		WithArgs("uom-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("DELETE", "/api/v1/uoms/uom-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}
