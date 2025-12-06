package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/data-processor/internal/repository"
	"github.com/asian-code/myapp-kubernetes/services/shared/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	repo    *repository.Repository
	logger  *log.Entry
	metrics *metrics.Metrics
}

func New(repo *repository.Repository, logger *log.Entry, m *metrics.Metrics) *Handler {
	return &Handler{
		repo:    repo,
		logger:  logger,
		metrics: m,
	}
}

type IngestRequest struct {
	Type string          `json:"type"` // "sleep", "activity", "readiness"
	Data json.RawMessage `json:"data"`
}

type SleepData struct {
	ID       string `json:"id"`
	Day      string `json:"day"`
	Score    int    `json:"score"`
	Duration int    `json:"duration"`
}

type ActivityData struct {
	ID                string `json:"id"`
	Day               string `json:"day"`
	Score             int    `json:"score"`
	ActiveCalories    int    `json:"active_calories"`
	Steps             int    `json:"steps"`
	MediumActivityMin int    `json:"medium_activity_minutes"`
	HighActivityMin   int    `json:"high_activity_minutes"`
}

type ReadinessData struct {
	ID    string `json:"id"`
	Day   string `json:"day"`
	Score int    `json:"score"`
}

func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metrics.HTTPRequestDuration.Observe(time.Since(start).Seconds())
		h.metrics.HTTPRequestsTotal.Inc()
	}()

	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	switch req.Type {
	case "sleep":
		var data SleepData
		if err := json.Unmarshal(req.Data, &data); err != nil {
			http.Error(w, "Invalid sleep data", http.StatusBadRequest)
			return
		}

		day, _ := time.Parse("2006-01-02", data.Day)
		metric := &repository.SleepMetric{
			OuraID:   data.ID,
			Day:      day,
			Score:    data.Score,
			Duration: data.Duration,
		}

		if err := h.repo.SaveSleepMetric(ctx, metric); err != nil {
			h.logger.WithError(err).Error("Failed to save sleep metric")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

	case "activity":
		var data ActivityData
		if err := json.Unmarshal(req.Data, &data); err != nil {
			http.Error(w, "Invalid activity data", http.StatusBadRequest)
			return
		}

		day, _ := time.Parse("2006-01-02", data.Day)
		metric := &repository.ActivityMetric{
			OuraID:            data.ID,
			Day:               day,
			Score:             data.Score,
			ActiveCalories:    data.ActiveCalories,
			Steps:             data.Steps,
			MediumActivityMin: data.MediumActivityMin,
			HighActivityMin:   data.HighActivityMin,
		}

		if err := h.repo.SaveActivityMetric(ctx, metric); err != nil {
			h.logger.WithError(err).Error("Failed to save activity metric")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

	case "readiness":
		var data ReadinessData
		if err := json.Unmarshal(req.Data, &data); err != nil {
			http.Error(w, "Invalid readiness data", http.StatusBadRequest)
			return
		}

		day, _ := time.Parse("2006-01-02", data.Day)
		metric := &repository.ReadinessMetric{
			OuraID: data.ID,
			Day:    day,
			Score:  data.Score,
		}

		if err := h.repo.SaveReadinessMetric(ctx, metric); err != nil {
			h.logger.WithError(err).Error("Failed to save readiness metric")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricType := vars["type"]

	// Parse date range from query params (default: last 7 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	if start := r.URL.Query().Get("start"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = t
		}
	}
	if end := r.URL.Query().Get("end"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endDate = t
		}
	}

	ctx := r.Context()

	switch metricType {
	case "sleep":
		metrics, err := h.repo.GetSleepMetrics(ctx, startDate, endDate)
		if err != nil {
			h.logger.WithError(err).Error("Failed to get sleep metrics")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(metrics)

	case "activity":
		metrics, err := h.repo.GetActivityMetrics(ctx, startDate, endDate)
		if err != nil {
			h.logger.WithError(err).Error("Failed to get activity metrics")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(metrics)

	case "readiness":
		metrics, err := h.repo.GetReadinessMetrics(ctx, startDate, endDate)
		if err != nil {
			h.logger.WithError(err).Error("Failed to get readiness metrics")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(metrics)

	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *Handler) PrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}
