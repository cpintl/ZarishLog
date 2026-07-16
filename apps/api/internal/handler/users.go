package handler

import (
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID         string  `json:"id" db:"id"`
	OrgID      string  `json:"org_id" db:"org_id"`
	Email      string  `json:"email" db:"email" validate:"required,email,max=255"`
	Name       string  `json:"name" db:"name" validate:"required,max=255"`
	RoleID     *string `json:"role_id" db:"role_id" validate:"omitempty,uuid7"`
	OrgLevelID *string `json:"org_level_id" db:"org_level_id" validate:"omitempty,uuid7"`
	IsActive   bool    `json:"is_active" db:"is_active"`
	CreatedBy  string  `json:"created_by" db:"created_by"`
	UpdatedBy  string  `json:"updated_by" db:"updated_by"`
}

type Role struct {
	ID          string `json:"id" db:"id"`
	Code        string `json:"code" db:"code"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Level       int    `json:"level" db:"level"`
}

type Permission struct {
	ID          string `json:"id" db:"id"`
	Module      string `json:"module" db:"module"`
	Action      string `json:"action" db:"action"`
	Description string `json:"description" db:"description"`
}

type UserRoleAssignment struct {
	UserID     string `json:"user_id" db:"user_id"`
	RoleID     string `json:"role_id" db:"role_id" validate:"required,uuid7"`
	OrgLevelID *string `json:"org_level_id" db:"org_level_id" validate:"omitempty,uuid7"`
	GrantedBy  string `json:"granted_by" db:"granted_by"`
	RoleName   string `json:"role_name,omitempty" db:"role_name"`
}

func ListUsers(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := pagination.FromQuery(c)

		var total int
		err := db.Get(&total, `SELECT COUNT(*) FROM users`)
		if err != nil {
			response.InternalError(c, "failed to count users")
			return
		}

		var users []User
		err = db.Select(&users,
			`SELECT id, org_id, email, name, role_id, org_level_id, is_active, created_by, updated_by
			 FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			p.Limit(), p.Offset(),
		)
		if err != nil {
			response.InternalError(c, "failed to list users")
			return
		}

		if users == nil {
			users = []User{}
		}

		response.Paginated(c, users, total, p.Page, p.PageSize)
	}
}

func GetUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var u User
		err := db.Get(&u,
			`SELECT id, org_id, email, name, role_id, org_level_id, is_active, created_by, updated_by
			 FROM users WHERE id=$1`, id,
		)
		if err != nil {
			response.NotFound(c, "user not found")
			return
		}

		type Assignment struct {
			RoleID     string  `json:"role_id" db:"role_id"`
			RoleName   string  `json:"role_name" db:"role_name"`
			RoleCode   string  `json:"role_code" db:"role_code"`
			OrgLevelID *string `json:"org_level_id" db:"org_level_id"`
		}

		var assignments []Assignment
		_ = db.Select(&assignments,
			`SELECT ura.role_id, r.name AS role_name, r.code AS role_code, ura.org_level_id
			 FROM user_role_assignments ura
			 JOIN roles r ON ura.role_id = r.id
			 WHERE ura.user_id=$1`, id,
		)
		if assignments == nil {
			assignments = []Assignment{}
		}

		response.OK(c, gin.H{
			"user":        u,
			"assignments": assignments,
		})
	}
}

func CreateUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrgID      string  `json:"org_id" validate:"required,uuid7"`
			Email      string  `json:"email" validate:"required,email,max=255"`
			Name       string  `json:"name" validate:"required,max=255"`
			RoleID     *string `json:"role_id" validate:"omitempty,uuid7"`
			OrgLevelID *string `json:"org_level_id" validate:"omitempty,uuid7"`
			IsActive   *bool   `json:"is_active"`
			CreatedBy  string  `json:"created_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		active := true
		if req.IsActive != nil {
			active = *req.IsActive
		}

		var id string
		err := db.QueryRowx(
			`INSERT INTO users (org_id, email, name, role_id, org_level_id, is_active, created_by, updated_by)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$7) RETURNING id`,
			req.OrgID, req.Email, req.Name,
			nullIfEmptyPtr(req.RoleID), nullIfEmptyPtr(req.OrgLevelID),
			active, req.CreatedBy,
		).Scan(&id)
		if err != nil {
			response.InternalError(c, "failed to create user: "+err.Error())
			return
		}

		response.Created(c, gin.H{"id": id})
	}
}

func UpdateUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req struct {
			Email      string  `json:"email" validate:"required,email,max=255"`
			Name       string  `json:"name" validate:"required,max=255"`
			RoleID     *string `json:"role_id" validate:"omitempty,uuid7"`
			OrgLevelID *string `json:"org_level_id" validate:"omitempty,uuid7"`
			IsActive   *bool   `json:"is_active"`
			UpdatedBy  string  `json:"updated_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		var active bool
		if req.IsActive != nil {
			active = *req.IsActive
		}

		result, err := db.Exec(
			`UPDATE users SET email=$1, name=$2, role_id=$3, org_level_id=$4, is_active=$5, updated_by=$6, updated_at=now() WHERE id=$7`,
			req.Email, req.Name, nullIfEmptyPtr(req.RoleID), nullIfEmptyPtr(req.OrgLevelID),
			active, req.UpdatedBy, id,
		)
		if err != nil {
			response.InternalError(c, "failed to update user: "+err.Error())
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			response.NotFound(c, "user not found")
			return
		}

		response.OK(c, gin.H{"id": id})
	}
}

func DeactivateUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := db.Exec(`UPDATE users SET is_active=false, updated_at=now() WHERE id=$1`, id)
		if err != nil {
			response.InternalError(c, "failed to deactivate user")
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			response.NotFound(c, "user not found")
			return
		}

		response.OK(c, gin.H{"id": id})
	}
}

func ListRoles(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var roles []Role
		err := db.Select(&roles, `SELECT id, code, name, description, level FROM roles ORDER BY level`)
		if err != nil {
			response.InternalError(c, "failed to list roles")
			return
		}
		if roles == nil {
			roles = []Role{}
		}
		response.OK(c, roles)
	}
}

func ListPermissions(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var perms []Permission
		err := db.Select(&perms, `SELECT id, module, action, description FROM permissions ORDER BY module, action`)
		if err != nil {
			response.InternalError(c, "failed to list permissions")
			return
		}
		if perms == nil {
			perms = []Permission{}
		}
		response.OK(c, perms)
	}
}

func AssignUserRole(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var req struct {
			RoleID     string  `json:"role_id" validate:"required,uuid7"`
			OrgLevelID *string `json:"org_level_id" validate:"omitempty,uuid7"`
			GrantedBy  string  `json:"granted_by" validate:"required,max=255"`
		}
		if errs := validator.BindAndValidate(c, &req); errs != nil {
			response.Validation(c, errs)
			return
		}

		_, err := db.Exec(
			`INSERT INTO user_role_assignments (user_id, role_id, org_level_id, granted_by)
			 VALUES ($1,$2,$3,$4) ON CONFLICT (user_id, role_id, org_level_id) DO NOTHING`,
			userID, req.RoleID, nullIfEmptyPtr(req.OrgLevelID), req.GrantedBy,
		)
		if err != nil {
			response.InternalError(c, "failed to assign role: "+err.Error())
			return
		}

		response.Created(c, gin.H{"user_id": userID, "role_id": req.RoleID})
	}
}

func RemoveUserRole(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		roleID := c.Query("role_id")
		orgLevelID := c.Query("org_level_id")
		if roleID == "" {
			response.BadRequest(c, "role_id query parameter is required")
			return
		}

		if orgLevelID != "" {
			res, err := db.Exec(
				`DELETE FROM user_role_assignments WHERE user_id=$1 AND role_id=$2 AND org_level_id=$3`,
				userID, roleID, orgLevelID,
			)
			if err != nil {
				response.InternalError(c, "failed to remove role assignment")
				return
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				response.NotFound(c, "role assignment not found")
				return
			}
		} else {
			res, err := db.Exec(
				`DELETE FROM user_role_assignments WHERE user_id=$1 AND role_id=$2 AND org_level_id IS NULL`,
				userID, roleID,
			)
			if err != nil {
				response.InternalError(c, "failed to remove role assignment")
				return
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				response.NotFound(c, "role assignment not found")
				return
			}
		}

		response.OK(c, gin.H{"user_id": userID, "role_id": roleID})
	}
}
