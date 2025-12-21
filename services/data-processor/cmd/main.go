package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/data-processor/internal/config"
	"github.com/asian-code/myapp-kubernetes/services/data-processor/internal/handler"
	"github.com/asian-code/myapp-kubernetes/services/data-processor/internal/repository"
	"github.com/asian-code/myapp-kubernetes/services/shared/database"
	"github.com/asian-code/myapp-kubernetes/services/shared/logger"
	"github.com/asian-code/myapp-kubernetes/services/shared/metrics"
	"github.com/gorilla/mux"
)

func main() {
	log := logger.Init("data-processor")
	cfg := config.Load()

	log.Info("Starting data-processor service")

	// Connect to database
	ctx := context.Background()
	db, err := database.NewPool(ctx, database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
		MaxConns: cfg.DBMaxConns,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	log.Info("Database connection established")

	// Initialize repository
	repo := repository.New(db, log)

	// Initialize metrics
	m := metrics.New("data-processor")

	// Create handler
	h := handler.New(repo, log, m)

	// Setup router
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/ingest", h.Ingest).Methods("POST")
	router.HandleFunc("/api/v1/metrics/{type}", h.GetMetrics).Methods("GET")
	router.HandleFunc("/health", h.Health).Methods("GET")
	router.HandleFunc("/metrics", h.PrometheusMetrics).Methods("GET")

	// Setup HTTP server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Starting HTTP server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server forced to shutdown")
	}

	log.Info("Server exited")
}
