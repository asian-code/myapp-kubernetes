package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set required environment variables
	os.Setenv("DB_PASSWORD", "test_password_1234567890123456")
	os.Setenv("OURA_CLIENT_ID", "test_client_id")
	os.Setenv("OURA_CLIENT_SECRET", "test_client_secret")
	os.Setenv("JWT_SECRET", "test_jwt_secret_1234567890123456")
	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("OURA_CLIENT_ID")
		os.Unsetenv("OURA_CLIENT_SECRET")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := Load()

	if cfg.DBHost != "localhost" {
		t.Errorf("expected DBHost to be localhost, got %s", cfg.DBHost)
	}

	if cfg.DBPort != 5432 {
		t.Errorf("expected DBPort to be 5432, got %d", cfg.DBPort)
	}

	if cfg.DBPassword != "test_password_1234567890123456" {
		t.Errorf("expected DBPassword to be set")
	}

	if cfg.JWTSecret != "test_jwt_secret_1234567890123456" {
		t.Errorf("expected JWTSecret to be set")
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel to be info, got %s", cfg.LogLevel)
	}
}

func TestLoad_MissingRequiredField(t *testing.T) {
	// Clear all env vars
	os.Clearenv()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Load to panic on missing required fields")
		}
	}()

	Load()
}

func TestLoad_CustomValues(t *testing.T) {
	// Set custom values
	os.Setenv("DB_HOST", "custom-host")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "custom_user")
	os.Setenv("DB_PASSWORD", "custom_password_1234567890123456")
	os.Setenv("DB_NAME", "custom_db")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("OURA_CLIENT_ID", "custom_client")
	os.Setenv("OURA_CLIENT_SECRET", "custom_secret")
	os.Setenv("JWT_SECRET", "custom_jwt_secret_12345678901234567890")

	defer func() {
		os.Clearenv()
	}()

	cfg := Load()

	if cfg.DBHost != "custom-host" {
		t.Errorf("expected DBHost to be custom-host, got %s", cfg.DBHost)
	}

	if cfg.DBPort != 3306 {
		t.Errorf("expected DBPort to be 3306, got %d", cfg.DBPort)
	}

	if cfg.DBUser != "custom_user" {
		t.Errorf("expected DBUser to be custom_user, got %s", cfg.DBUser)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel to be debug, got %s", cfg.LogLevel)
	}

	if cfg.DBSSLMode != "disable" {
		t.Errorf("expected DBSSLMode to be disable, got %s", cfg.DBSSLMode)
	}
}
