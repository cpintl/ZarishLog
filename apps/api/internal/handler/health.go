package handler

import (
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Health(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := db.Ping()
		if err != nil {
			response.Error(c, 503, response.ErrInternal, "database disconnected")
			return
		}
		response.OK(c, gin.H{
			"status": "healthy",
			"db":     "connected",
			"org_id": c.GetString("org_id"),
		})
	}
}
