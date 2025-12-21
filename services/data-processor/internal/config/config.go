package config

import (
	"os"
	"strconv"

	"github.com/asian-code/myapp-kubernetes/services/pkg/validation"
)

type Config struct {
	DBHost     string `validate:"required"`
	DBPort     int    `validate:"required,min=1,max=65535"`
	DBUser     string `validate:"required"`
	DBPassword string `validate:"required"`
	DBName     string `validate:"required"`
	DBSSLMode  string `validate:"required,oneof=disable require verify-ca verify-full"`
	DBMaxConns int    `validate:"required,min=1,max=100"`
	LogLevel   string `validate:"required,oneof=debug info warn error"`
}

// Load loads and validates configuration from environment variables
func Load() *Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	dbMaxConns, _ := strconv.Atoi(getEnv("DB_MAX_CONNS", "10"))

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "myhealth_user"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME", "myhealth"),
		DBSSLMode:  getEnv("DB_SSLMODE", "require"),
		DBMaxConns: dbMaxConns,
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}

	// Validate configuration and panic if invalid
	validation.MustValidate(cfg)

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
