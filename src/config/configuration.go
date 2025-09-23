package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT string
	Env  string
}

func LoadConfig() *Config {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	if os.Getenv("RUNNING_IN_DOCKER") == "" {
		if err := godotenv.Load(".env." + env); err != nil {
			logger.Error("Failed to load .env file: ", err, "")
		}
	}

	return &Config{
		PORT: getEnv("PORT", "8080"),
		Env:  env,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
