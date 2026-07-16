package middleware

import (
	"github.com/gin-gonic/gin"
)

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		if orgID != "" {
			c.Request.Header.Set("X-Tenant-ID", orgID)
		}
		c.Next()
	}
}
