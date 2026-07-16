package handler

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

func Version(version, commitHash, buildTime string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":     version,
			"commit_hash": commitHash,
			"build_time":  buildTime,
			"go_version":  runtime.Version(),
			"go_arch":     runtime.GOARCH,
			"go_os":       runtime.GOOS,
		})
	}
}
