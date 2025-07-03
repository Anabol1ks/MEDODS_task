package main

import (
	"medods-auth/internal/auth/token"
	"medods-auth/internal/config"
	"medods-auth/internal/db"
	"medods-auth/internal/logger"
	"medods-auth/internal/router"

	"go.uber.org/zap"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}

	log := logger.L()
	log.Info("Инициализация логгера успешна")

	cfg := config.Load(log)

	db.ConnectDB(cfg, log)
	db.Migrate(log)

	token.InitJWT(cfg.JWTSecret)

	r := router.Router(cfg, db.DB, log)
	if err := r.Run(":8080"); err != nil {
		log.Error("Failed to run server", zap.Error(err))
	}
}
