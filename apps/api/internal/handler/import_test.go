package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupImportRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", "org-123")
		c.Set("user_id", "user-456")
		c.Next()
	})
	r.POST("/api/v1/products/import", ImportProducts(db))
	r.GET("/api/v1/products/search", SearchProducts(db))
	return r
}

func TestImportProducts_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupImportRouter(db)

	csvContent := "sku,name,item_type\nABC-123,Aspirin Tablets,consumable\nDEF-456,Paracetamol,consumable\n"
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "products.csv")
	part.Write([]byte(csvContent))
	writer.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products WHERE`).
		WithArgs("ABC-123", "org-123").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO products`).
		WithArgs("org-123", nil, nil, "ABC-123", "Aspirin Tablets", "", "consumable", "", "", "user-456", "user-456").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products WHERE`).
		WithArgs("DEF-456", "org-123").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO products`).
		WithArgs("org-123", nil, nil, "DEF-456", "Paracetamol", "", "consumable", "", "", "user-456", "user-456").
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/v1/products/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp["data"].(map[string]interface{})["imported"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportProducts_MissingFile(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	router := setupImportRouter(sqlxDB)

	req, _ := http.NewRequest("POST", "/api/v1/products/import", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportProducts_DuplicateSKU(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupImportRouter(db)

	csvContent := "sku,name\nDUP-001,Duplicate Item\n"
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "products.csv")
	part.Write([]byte(csvContent))
	writer.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products WHERE`).
		WithArgs("DUP-001", "org-123").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	req, _ := http.NewRequest("POST", "/api/v1/products/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupImportRouter(db)

	likeQ := "%aspirin%"

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products WHERE`).
		WithArgs("org-123", likeQ).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, sku, name, item_type, brand, status FROM products WHERE`).
		WithArgs("org-123", likeQ, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sku", "name", "item_type", "brand", "status"}).
			AddRow("prod-1", "ASP-001", "Aspirin 500mg", "consumable", "Bayer", "active"))

	req, _ := http.NewRequest("GET", "/api/v1/products/search?q=aspirin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["total"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_NoQuery(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	router := setupImportRouter(sqlxDB)

	req, _ := http.NewRequest("GET", "/api/v1/products/search", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
