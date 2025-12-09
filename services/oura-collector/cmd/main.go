package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/oura-collector/internal/client"
	"github.com/asian-code/myapp-kubernetes/services/oura-collector/internal/config"
	"github.com/asian-code/myapp-kubernetes/services/shared/database"
	"github.com/asian-code/myapp-kubernetes/services/shared/logger"
	"github.com/asian-code/myapp-kubernetes/services/shared/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log := logger.Init("oura-collector")
	cfg := config.Load()

	// Validate required environment variables
	if cfg.ProcessorURL == "" {
		log.Fatal("PROCESSOR_URL environment variable is required")
	}
	if cfg.DBPassword == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}
	if cfg.UserID == "" {
		log.Fatal("USER_ID environment variable is required")
	}

	log.Info("Starting oura-collector")

	// Initialize metrics
	m := metrics.New("oura-collector")

	// Start metrics server in background
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Info("Metrics server listening on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.WithError(err).Error("Failed to start metrics server")
		}
	}()

	// Connect to database
	ctx := context.Background()
	db, err := database.NewPool(ctx, database.Config{
		Host:     cfg.DBHost,
		Port:     parseInt(cfg.DBPort),
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
		MaxConns: 5,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Fetch OAuth token from database
	var accessToken string
	var expiresAt time.Time
	query := `SELECT access_token, expires_at FROM oauth_tokens WHERE user_id = $1 AND provider = 'oura'`
	err = db.QueryRow(ctx, query, cfg.UserID).Scan(&accessToken, &expiresAt)
	if err != nil {
		log.WithError(err).Fatal("Failed to get OAuth token from database. Please authorize the app first.")
	}

	// Check if token is expired (refresh is handled by api-service)
	if time.Now().After(expiresAt) {
		log.Fatal("OAuth token has expired. Please re-authorize the app.")
	}

	log.Info("OAuth token retrieved successfully")

	// Record collection run start
	startTime := time.Now()
	m.CollectionRunsTotal.Inc()

	// Fetch data from Oura
	ouraClient := client.New(accessToken, log)

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	date := time.Now().Format("2006-01-02")

	var dataPointsCollected int
	var hasErrors bool

	// Fetch sleep data
	sleepData, err := ouraClient.GetSleepData(ctxTimeout, date)
	if err != nil {
		log.WithError(err).Error("Failed to fetch sleep data")
		m.CollectionErrors.WithLabelValues("sleep", "fetch_failed").Inc()
		hasErrors = true
	} else {
		log.WithField("sleep_score", sleepData.Score).Info("Successfully fetched sleep data")
		dataPointsCollected++
		if err := ouraClient.SendToProcessor(ctxTimeout, cfg.ProcessorURL, sleepData); err != nil {
			log.WithError(err).Error("Failed to send sleep data to processor")
			m.CollectionErrors.WithLabelValues("sleep", "send_failed").Inc()
			hasErrors = true
		}
	}

	// Fetch activity data
	activityData, err := ouraClient.GetActivityData(ctxTimeout, date)
	if err != nil {
		log.WithError(err).Error("Failed to fetch activity data")
		m.CollectionErrors.WithLabelValues("activity", "fetch_failed").Inc()
		hasErrors = true
	} else {
		log.WithField("activity_score", activityData.Score).Info("Successfully fetched activity data")
		dataPointsCollected++
		if err := ouraClient.SendToProcessor(ctxTimeout, cfg.ProcessorURL, activityData); err != nil {
			log.WithError(err).Error("Failed to send activity data to processor")
			m.CollectionErrors.WithLabelValues("activity", "send_failed").Inc()
			hasErrors = true
		}
	}

	// Fetch readiness data
	readinessData, err := ouraClient.GetReadinessData(ctxTimeout, date)
	if err != nil {
		log.WithError(err).Error("Failed to fetch readiness data")
		m.CollectionErrors.WithLabelValues("readiness", "fetch_failed").Inc()
		hasErrors = true
	} else {
		log.WithField("readiness_score", readinessData.Score).Info("Successfully fetched readiness data")
		dataPointsCollected++
		if err := ouraClient.SendToProcessor(ctxTimeout, cfg.ProcessorURL, readinessData); err != nil {
			log.WithError(err).Error("Failed to send readiness data to processor")
			m.CollectionErrors.WithLabelValues("readiness", "send_failed").Inc()
			hasErrors = true
		}
	}

	// Record metrics
	duration := time.Since(startTime).Seconds()
	m.CollectionDuration.Observe(duration)
	m.DataPointsCollected.Add(float64(dataPointsCollected))

	if !hasErrors {
		m.LastSuccessfulRunTime.Set(float64(time.Now().Unix()))
		log.WithField("duration_seconds", duration).WithField("data_points", dataPointsCollected).Info("oura-collector completed successfully")
	} else {
		log.WithField("duration_seconds", duration).WithField("data_points", dataPointsCollected).Warn("oura-collector completed with errors")
	}
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
