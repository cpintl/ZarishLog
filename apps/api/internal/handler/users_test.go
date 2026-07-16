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

func setupUsersRouter(db *sqlx.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", testOrgID)
		c.Set("user_id", testUserID)
		c.Next()
	})
	r.GET("/api/v1/users", ListUsers(db))
	r.POST("/api/v1/users", CreateUser(db))
	r.GET("/api/v1/users/:id", GetUser(db))
	r.PUT("/api/v1/users/:id", UpdateUser(db))
	r.DELETE("/api/v1/users/:id", DeactivateUser(db))
	r.POST("/api/v1/users/:id/roles", AssignUserRole(db))
	r.DELETE("/api/v1/users/:id/roles", RemoveUserRole(db))
	r.GET("/api/v1/roles", ListRoles(db))
	r.GET("/api/v1/permissions", ListPermissions(db))
	return r
}

func TestListUsers(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT (.+) FROM users ORDER BY created_at DESC LIMIT .+ OFFSET .+`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "email", "name", "role_id", "org_level_id", "is_active", "created_by", "updated_by"}).
			AddRow(testInspID, testOrgID, "user@example.com", "Test User", nil, nil, true, testUserID, testUserID))

	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	body := `{
		"org_id":"` + testOrgID + `",
		"email":"user@example.com",
		"name":"Test User",
		"created_by":"` + testUserID + `"
	}`

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(testOrgID, "user@example.com", "Test User", nil, nil, true, testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testInspID))

	req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(body))
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

func TestCreateUser_ValidationError(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	body := `{"org_id":"bad","email":"not-email","name":"","created_by":""}`

	req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	mock.ExpectQuery(`SELECT (.+) FROM users WHERE id=\$1`).
		WithArgs(testInspID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "email", "name", "role_id", "org_level_id", "is_active", "created_by", "updated_by"}).
			AddRow(testInspID, testOrgID, "user@example.com", "Test User", nil, nil, true, testUserID, testUserID))
	mock.ExpectQuery(`SELECT ura\.role_id, r\.name AS role_name, r\.code AS role_code, ura\.org_level_id FROM user_role_assignments ura JOIN roles r ON ura\.role_id = r\.id WHERE ura\.user_id=\$1`).
		WithArgs(testInspID).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "role_name", "role_code", "org_level_id"}))

	req, _ := http.NewRequest("GET", "/api/v1/users/"+testInspID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	body := `{
		"email":"updated@example.com",
		"name":"Updated User",
		"updated_by":"` + testUserID + `"
	}`

	mock.ExpectExec(`UPDATE users SET`).
		WithArgs("updated@example.com", "Updated User", nil, nil, false, testUserID, testInspID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("PUT", "/api/v1/users/"+testInspID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeactivateUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	mock.ExpectExec(`UPDATE users SET is_active=false`).
		WithArgs(testInspID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("DELETE", "/api/v1/users/"+testInspID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAssignUserRole(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	body := `{"role_id":"` + testProdID + `","granted_by":"` + testUserID + `"}`

	mock.ExpectExec(`INSERT INTO user_role_assignments`).
		WithArgs(testInspID, testProdID, nil, testUserID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("POST", "/api/v1/users/"+testInspID+"/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRemoveUserRole(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	mock.ExpectExec(`DELETE FROM user_role_assignments WHERE user_id=\$1 AND role_id=\$2 AND org_level_id IS NULL`).
		WithArgs(testInspID, testProdID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("DELETE", "/api/v1/users/"+testInspID+"/roles?role_id="+testProdID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListRoles(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	mock.ExpectQuery(`SELECT id, code, name, description, level FROM roles ORDER BY level`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "level"}).
			AddRow(testInspID, "R01", "Global Admin", "Full access", 1))

	req, _ := http.NewRequest("GET", "/api/v1/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListPermissions(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	router := setupUsersRouter(db)

	mock.ExpectQuery(`SELECT id, module, action, description FROM permissions ORDER BY module, action`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "module", "action", "description"}).
			AddRow(testInspID, "products", "read", "View products"))

	req, _ := http.NewRequest("GET", "/api/v1/permissions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
