package config

import (
	"os"

	"github.com/asian-code/myapp-kubernetes/services/pkg/validation"
)

type Config struct {
	ProcessorURL string `validate:"required,url"`
	LogLevel     string `validate:"required,oneof=debug info warn error"`
	DBHost       string `validate:"required"`
	DBPort       string `validate:"required"`
	DBUser       string `validate:"required"`
	DBPassword   string `validate:"required"`
	DBName       string `validate:"required"`
	DBSSLMode    string `validate:"required,oneof=disable require verify-ca verify-full"`
	UserID       string `validate:"required"` // The user ID to fetch data for
}

// Load loads and validates configuration from environment variables
func Load() *Config {
	cfg := &Config{
		ProcessorURL: os.Getenv("PROCESSOR_URL"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "myhealth_user"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       getEnv("DB_NAME", "myhealth"),
		DBSSLMode:    getEnv("DB_SSLMODE", "require"),
		UserID:       os.Getenv("USER_ID"),
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
