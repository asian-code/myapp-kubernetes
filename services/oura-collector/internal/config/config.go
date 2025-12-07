package config

import (
	"os"
)

type Config struct {
	ProcessorURL string
	LogLevel     string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	UserID       string // The user ID to fetch data for
}

func Load() *Config {
	return &Config{
		ProcessorURL: os.Getenv("PROCESSOR_URL"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "myhealth_user"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       getEnv("DB_NAME", "myhealth"),
		DBSSLMode:    getEnv("DB_SSLMODE", "require"),
		UserID:       os.Getenv("USER_ID"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
