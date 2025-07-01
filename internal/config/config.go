package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	DB         DBConfig
	JWTSecret  string
	WebHookUrl string
	TokenTTL   time.Duration
	RefreshTTL time.Duration
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load(log *zap.Logger) *Config {
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found, using system env")
	}

	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "auth-db"),
			Port:     getEnv("DB_PORT", "5433"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "auth"),
		},
		JWTSecret:  getEnv("JWT_SECRET", "jwt-secret"),
		WebHookUrl: getEnv("WEBHOOK_URL", "https://webhook.site/"),
		TokenTTL:   parseDuration(getEnv("TOKEN_TTL", "15m"), log),
		RefreshTTL: parseDuration(getEnv("REFRESH_TTL", "48h"), log),
	}
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}

func parseDuration(s string, log *zap.Logger) time.Duration {
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := time.ParseDuration(daysStr + "h")
		if err != nil {
			log.Warn("Ошибка парсинга TTL", zap.Error(err))
			return 0
		}
		return time.Duration(24) * days
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return duration
}
