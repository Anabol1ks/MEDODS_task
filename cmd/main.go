package main

import (
	"medods-auth/internal/config"
	"medods-auth/internal/db"
	"medods-auth/internal/logger"

	"go.uber.org/zap"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}

	log := logger.L()
	log.Info("Инициализация логгера успешна")

	cfg := config.Load(log)

	log.Info("Конфигурация загружена", zap.Any("config", cfg))

	db.ConnectDB(cfg, log)
}
