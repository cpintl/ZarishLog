package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrBadRequest     ErrorCode = "BAD_REQUEST"
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrForbidden      ErrorCode = "FORBIDDEN"
	ErrInternal       ErrorCode = "INTERNAL_ERROR"
	ErrValidation     ErrorCode = "VALIDATION_ERROR"
	ErrNotImplemented ErrorCode = "NOT_IMPLEMENTED"
	ErrConflict       ErrorCode = "CONFLICT"
)

type ErrorBody struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type SuccessBody struct {
	Data interface{} `json:"data"`
}

type PaginatedBody struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessBody{Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessBody{Data: data})
}

func Error(c *gin.Context, status int, code ErrorCode, msg string, details ...interface{}) {
	body := ErrorBody{Code: code, Message: msg}
	if len(details) > 0 {
		body.Details = details[0]
	}
	c.JSON(status, body)
	c.Abort()
}

func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, ErrNotFound, msg)
}

func BadRequest(c *gin.Context, msg string, details ...interface{}) {
	Error(c, http.StatusBadRequest, ErrBadRequest, msg, details...)
}

func InternalError(c *gin.Context, msg string) {
	Error(c, http.StatusInternalServerError, ErrInternal, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, ErrUnauthorized, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, ErrForbidden, msg)
}

func NotImplemented(c *gin.Context, msg string) {
	Error(c, http.StatusNotImplemented, ErrNotImplemented, msg)
}

func Conflict(c *gin.Context, msg string) {
	Error(c, http.StatusConflict, ErrConflict, msg)
}

func Validation(c *gin.Context, details interface{}) {
	Error(c, http.StatusUnprocessableEntity, ErrValidation, "validation failed", details)
}

func Paginated(c *gin.Context, data interface{}, total int, page int, pageSize int) {
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	c.JSON(http.StatusOK, PaginatedBody{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}
