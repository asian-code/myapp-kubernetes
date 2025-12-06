package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

type Repository struct {
	db     *pgxpool.Pool
	logger *log.Entry
}

func New(db *pgxpool.Pool, logger *log.Entry) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

type SleepMetric struct {
	OuraID   string    `json:"oura_id"`
	Day      time.Time `json:"day"`
	Score    int       `json:"score"`
	Duration int       `json:"duration"`
}

type ActivityMetric struct {
	OuraID            string    `json:"oura_id"`
	Day               time.Time `json:"day"`
	Score             int       `json:"score"`
	ActiveCalories    int       `json:"active_calories"`
	Steps             int       `json:"steps"`
	MediumActivityMin int       `json:"medium_activity_minutes"`
	HighActivityMin   int       `json:"high_activity_minutes"`
}

type ReadinessMetric struct {
	OuraID string    `json:"oura_id"`
	Day    time.Time `json:"day"`
	Score  int       `json:"score"`
}

func (r *Repository) GetSleepMetrics(ctx context.Context, startDate, endDate time.Time) ([]*SleepMetric, error) {
	query := `
		SELECT oura_id, day, score, duration
		FROM sleep_metrics
		WHERE day BETWEEN $1 AND $2
		ORDER BY day DESC
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*SleepMetric
	for rows.Next() {
		var m SleepMetric
		if err := rows.Scan(&m.OuraID, &m.Day, &m.Score, &m.Duration); err != nil {
			return nil, err
		}
		metrics = append(metrics, &m)
	}

	return metrics, rows.Err()
}

func (r *Repository) GetActivityMetrics(ctx context.Context, startDate, endDate time.Time) ([]*ActivityMetric, error) {
	query := `
		SELECT oura_id, day, score, active_calories, steps, 
		       medium_activity_minutes, high_activity_minutes
		FROM activity_metrics
		WHERE day BETWEEN $1 AND $2
		ORDER BY day DESC
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*ActivityMetric
	for rows.Next() {
		var m ActivityMetric
		if err := rows.Scan(&m.OuraID, &m.Day, &m.Score, &m.ActiveCalories,
			&m.Steps, &m.MediumActivityMin, &m.HighActivityMin); err != nil {
			return nil, err
		}
		metrics = append(metrics, &m)
	}

	return metrics, rows.Err()
}

func (r *Repository) GetReadinessMetrics(ctx context.Context, startDate, endDate time.Time) ([]*ReadinessMetric, error) {
	query := `
		SELECT oura_id, day, score
		FROM readiness_metrics
		WHERE day BETWEEN $1 AND $2
		ORDER BY day DESC
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*ReadinessMetric
	for rows.Next() {
		var m ReadinessMetric
		if err := rows.Scan(&m.OuraID, &m.Day, &m.Score); err != nil {
			return nil, err
		}
		metrics = append(metrics, &m)
	}

	return metrics, rows.Err()
}
