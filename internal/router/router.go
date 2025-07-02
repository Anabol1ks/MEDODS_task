package router

import (
	"medods-auth/internal/config"

	"github.com/gin-gonic/gin"
)

func Router(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}
