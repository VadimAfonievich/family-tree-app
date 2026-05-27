package config

import (
	"os"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	TelegramBotToken string
	ServerPort       string
	LogLevel         string
}

func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://familytree:familytree@localhost:5432/familytree?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "super-secret-change-me"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
