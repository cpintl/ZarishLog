package middleware

import (
	"net/http"

	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			switch err.Type {
			case gin.ErrorTypeBind:
				response.BadRequest(c, err.Error())
			default:
				response.InternalError(c, "an unexpected error occurred")
			}
			c.Abort()
		}

		if c.Writer.Status() == http.StatusNotFound {
			response.NotFound(c, "route not found")
		}
	}
}
