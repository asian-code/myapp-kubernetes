package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/api-service/internal/config"
	"github.com/asian-code/myapp-kubernetes/services/api-service/internal/handler"
	"github.com/asian-code/myapp-kubernetes/services/api-service/internal/repository"
	"github.com/asian-code/myapp-kubernetes/services/shared/database"
	"github.com/asian-code/myapp-kubernetes/services/shared/logger"
	"github.com/asian-code/myapp-kubernetes/services/shared/metrics"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	log := logger.Init("api-service")
	cfg := config.Load()

	log.Info("Starting api-service")

	// Connect to database
	ctx := context.Background()
	db, err := database.NewPool(ctx, database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
		MaxConns: 25,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	log.Info("Database connection established")

	// Initialize repository
	repo := repository.New(db, log)

	// Initialize metrics
	m := metrics.New("api-service")

	// Create handler
	h := handler.New(repo, log, m, cfg.JWTSecret)

	// Setup router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/auth/login", h.Login).Methods("POST")
	router.HandleFunc("/health", h.Health).Methods("GET")
	router.HandleFunc("/metrics", h.PrometheusMetrics).Methods("GET")

	// Protected routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(h.AuthMiddleware)
	api.HandleFunc("/dashboard", h.Dashboard).Methods("GET")
	api.HandleFunc("/sleep", h.GetSleep).Methods("GET")
	api.HandleFunc("/activity", h.GetActivity).Methods("GET")
	api.HandleFunc("/readiness", h.GetReadiness).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://eric-n.com", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Setup HTTP server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      c.Handler(router),
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
