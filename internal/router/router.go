package router

import (
	"medods-auth/internal/config"
	"medods-auth/internal/handler"
	"medods-auth/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Router(cfg *config.Config, db *gorm.DB, l *zap.Logger) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authService := service.NewAuthService(db, cfg)
	authHandler := handler.NewAuthHandler(authService, l)

	r.POST("/auth/token", authHandler.Token)
	r.POST("/auth/refresh", authHandler.Refresh)
	r.POST("/auth/logout", authHandler.Logout)
	r.GET("/me", authHandler.Me)

	return r
}
