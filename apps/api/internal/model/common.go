package model

import "time"

type Organization struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Code      string    `json:"code" db:"code"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type OrgLevel struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	ParentID  string    `json:"parent_id" db:"parent_id"`
	Name      string    `json:"name" db:"name"`
	Level     int       `json:"level" db:"level"`
	Code      string    `json:"code" db:"code"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID          string    `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	Email       string    `json:"email" db:"email"`
	Name        string    `json:"name" db:"name"`
	RoleID      string    `json:"role_id" db:"role_id"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Role struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Code        string `json:"code" db:"code"`
	Description string `json:"description" db:"description"`
}

type PaginationParams struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
