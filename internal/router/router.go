package router

import (
	"medods-auth/internal/config"
	"medods-auth/internal/handler"
	"medods-auth/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Router(cfg *config.Config, db *gorm.DB, l *zap.Logger) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authService := service.NewAuthService(db, cfg)
	authHandler := handler.NewAuthHandler(authService, l)
	r.POST("/auth/token", authHandler.Token)
	r.POST("/auth/refresh", authHandler.Refresh)
	return r
}
