package middleware

import (
	"context"
	"net/http"

	"github.com/cpintl/ZarishLog/apps/api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Tenant pins a dedicated connection for the request, sets the tenant
// isolation context on it (session-scoped), and hands it to the handler chain.
//
// The API connects to PostgreSQL as zarishlog_app - a non-owner role, so RLS
// is enforced. Tenant GUCs are session-scoped (set_config is_local=false), so
// they survive across individual queries and stay visible inside any explicit
// transactions the handler opens on this same connection. The connection is
// returned to the pool only after the tenant context has been cleared, so a
// later request can never observe another tenant's context.
func Tenant(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetString("org_id")
		if orgID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing tenant context"})
			return
		}

		ctx := c.Request.Context()

		conn, err := db.Connx(ctx)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to allocate tenant connection"})
			return
		}

		_, err = conn.ExecContext(ctx,
			`SELECT app.set_isolation_context($1, $2, $3, $4, $5)`,
			orgID,
			c.GetString("program_id"),
			c.GetString("org_level"),
			c.GetString("department_id"),
			c.GetString("user_id"),
		)
		if err != nil {
			_ = conn.Close()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to set tenant context"})
			return
		}

		c.Set(handler.TenantConnKey, conn)

		defer func() {
			_, _ = conn.ExecContext(context.Background(), `SELECT app.set_isolation_context()`)
			_ = conn.Close()
		}()

		c.Next()
	}
}
