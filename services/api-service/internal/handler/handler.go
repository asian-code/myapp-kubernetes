package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/api-service/internal/auth"
	"github.com/asian-code/myapp-kubernetes/services/api-service/internal/repository"
	"github.com/asian-code/myapp-kubernetes/services/shared/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	repo      *repository.Repository
	logger    *log.Entry
	metrics   *metrics.Metrics
	jwtSecret string
}

func New(repo *repository.Repository, logger *log.Entry, m *metrics.Metrics, jwtSecret string) *Handler {
	return &Handler{
		repo:      repo,
		logger:    logger,
		metrics:   m,
		jwtSecret: jwtSecret,
	}
}

type DashboardResponse struct {
	LatestSleep     *repository.SleepMetric     `json:"latest_sleep"`
	LatestActivity  *repository.ActivityMetric  `json:"latest_activity"`
	LatestReadiness *repository.ReadinessMetric `json:"latest_readiness"`
	WeeklySummary   *WeeklySummary              `json:"weekly_summary"`
}

type WeeklySummary struct {
	AvgSleepScore     float64 `json:"avg_sleep_score"`
	AvgActivityScore  float64 `json:"avg_activity_score"`
	AvgReadinessScore float64 `json:"avg_readiness_score"`
	TotalSteps        int     `json:"total_steps"`
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metrics.HTTPRequestDuration.WithLabelValues(r.Method, "/dashboard").Observe(time.Since(start).Seconds())
		h.metrics.HTTPRequestsTotal.WithLabelValues(r.Method, "/dashboard", "200").Inc()
	}()

	ctx := r.Context()
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	// Get latest metrics
	sleepMetrics, _ := h.repo.GetSleepMetrics(ctx, startDate, endDate)
	activityMetrics, _ := h.repo.GetActivityMetrics(ctx, startDate, endDate)
	readinessMetrics, _ := h.repo.GetReadinessMetrics(ctx, startDate, endDate)

	response := &DashboardResponse{
		WeeklySummary: &WeeklySummary{},
	}

	// Get latest of each
	if len(sleepMetrics) > 0 {
		response.LatestSleep = sleepMetrics[0]
		// Calculate average
		total := 0
		for _, m := range sleepMetrics {
			total += m.Score
		}
		response.WeeklySummary.AvgSleepScore = float64(total) / float64(len(sleepMetrics))
	}

	if len(activityMetrics) > 0 {
		response.LatestActivity = activityMetrics[0]
		totalScore := 0
		totalSteps := 0
		for _, m := range activityMetrics {
			totalScore += m.Score
			totalSteps += m.Steps
		}
		response.WeeklySummary.AvgActivityScore = float64(totalScore) / float64(len(activityMetrics))
		response.WeeklySummary.TotalSteps = totalSteps
	}

	if len(readinessMetrics) > 0 {
		response.LatestReadiness = readinessMetrics[0]
		total := 0
		for _, m := range readinessMetrics {
			total += m.Score
		}
		response.WeeklySummary.AvgReadinessScore = float64(total) / float64(len(readinessMetrics))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetSleep(w http.ResponseWriter, r *http.Request) {
	h.getMetrics(w, r, "sleep")
}

func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	h.getMetrics(w, r, "activity")
}

func (h *Handler) GetReadiness(w http.ResponseWriter, r *http.Request) {
	h.getMetrics(w, r, "readiness")
}

func (h *Handler) getMetrics(w http.ResponseWriter, r *http.Request, metricType string) {
	start := time.Now()
	defer func() {
		endpoint := fmt.Sprintf("/metrics/%s", metricType)
		h.metrics.HTTPRequestDuration.WithLabelValues(r.Method, endpoint).Observe(time.Since(start).Seconds())
		h.metrics.HTTPRequestsTotal.WithLabelValues(r.Method, endpoint, "200").Inc()
	}()

	// Parse date range from query params (default: last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startParam := r.URL.Query().Get("start"); startParam != "" {
		if t, err := time.Parse("2006-01-02", startParam); err == nil {
			startDate = t
		}
	}
	if endParam := r.URL.Query().Get("end"); endParam != "" {
		if t, err := time.Parse("2006-01-02", endParam); err == nil {
			endDate = t
		}
	}

	ctx := r.Context()

	var result interface{}
	var err error

	switch metricType {
	case "sleep":
		result, err = h.repo.GetSleepMetrics(ctx, startDate, endDate)
	case "activity":
		result, err = h.repo.GetActivityMetrics(ctx, startDate, endDate)
	case "readiness":
		result, err = h.repo.GetReadinessMetrics(ctx, startDate, endDate)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to get metrics")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// TODO: Implement proper user authentication
	// For now, just generate a token for any user
	token, err := auth.GenerateToken(req.Username, h.jwtSecret)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *Handler) PrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

// AuthMiddleware validates JWT tokens
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := auth.ValidateToken(token, h.jwtSecret)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
