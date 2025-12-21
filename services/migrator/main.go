package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// Build database URL from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "myhealth_user")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := getEnv("DB_NAME", "myhealth")
	sslMode := getEnv("DB_SSLMODE", "require")

	if dbPassword == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, sslMode,
	)

	// Get migrations directory
	migrationsPath := getEnv("MIGRATIONS_PATH", "file://migrations")

	log.WithFields(log.Fields{
		"db_host":         dbHost,
		"db_port":         dbPort,
		"db_name":         dbName,
		"migrations_path": migrationsPath,
	}).Info("Starting database migration")

	// Create migrate instance
	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to create migrate instance")
	}
	defer m.Close()

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.WithError(err).Warn("Failed to get current version")
	} else if !errors.Is(err, migrate.ErrNilVersion) {
		log.WithFields(log.Fields{
			"version": version,
			"dirty":   dirty,
		}).Info("Current migration version")
	}

	// Run migrations
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("No migrations to apply - database is up to date")
			os.Exit(0)
		}
		log.WithError(err).Fatal("Migration failed")
	}

	// Get new version
	version, dirty, err = m.Version()
	if err != nil {
		log.WithError(err).Warn("Failed to get final version")
	} else {
		log.WithFields(log.Fields{
			"version": version,
			"dirty":   dirty,
		}).Info("Migration completed successfully")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
