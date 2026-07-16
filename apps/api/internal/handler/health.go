package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Health(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := db.Ping()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"db":     "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"db":     "connected",
			"org_id": c.GetString("org_id"),
		})
	}
}
