package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Audit(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		c.Next()

		if c.Request.Method == "OPTIONS" {
			return
		}

		if c.Writer.Status() >= 500 {
			return
		}

		userID := c.GetString("user_id")
		orgID := c.GetString("org_id")
		if userID == "" {
			return
		}

		id := uuid.New().String()
		action := c.Request.Method
		entityType := c.FullPath()
		entityID := c.Param("id")
		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		go func() {
			db.Exec(
				`INSERT INTO audit_log (id, org_id, user_id, action, entity_type, entity_id, ip_address, user_agent, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
				id, orgID, userID, action, entityType, entityID, ip, ua, time.Now(),
			)
		}()
	}
}
