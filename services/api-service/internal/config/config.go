package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost           string
	DBPort           int
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	DBMaxConns       int
	JWTSecret        string
	LogLevel         string
	OuraClientID     string
	OuraClientSecret string
	OuraRedirectURI  string
}

func Load() *Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	dbMaxConns, _ := strconv.Atoi(getEnv("DB_MAX_CONNS", "10"))

	return &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           dbPort,
		DBUser:           getEnv("DB_USER", "myhealth_user"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBName:           getEnv("DB_NAME", "myhealth"),
		DBSSLMode:        getEnv("DB_SSLMODE", "require"),
		DBMaxConns:       dbMaxConns,
		JWTSecret:        os.Getenv("JWT_SECRET"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		OuraClientID:     os.Getenv("OURA_CLIENT_ID"),
		OuraClientSecret: os.Getenv("OURA_CLIENT_SECRET"),
		OuraRedirectURI:  getEnv("OURA_REDIRECT_URI", "https://myhealth.eric-n.com/api/callback"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
