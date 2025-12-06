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

// Initialize database schema
func (r *Repository) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sleep_metrics (
			id SERIAL PRIMARY KEY,
			oura_id VARCHAR(255) UNIQUE,
			day DATE NOT NULL,
			score INTEGER,
			duration INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS activity_metrics (
			id SERIAL PRIMARY KEY,
			oura_id VARCHAR(255) UNIQUE,
			day DATE NOT NULL,
			score INTEGER,
			active_calories INTEGER,
			steps INTEGER,
			medium_activity_minutes INTEGER,
			high_activity_minutes INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS readiness_metrics (
			id SERIAL PRIMARY KEY,
			oura_id VARCHAR(255) UNIQUE,
			day DATE NOT NULL,
			score INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sleep_day ON sleep_metrics(day)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_day ON activity_metrics(day)`,
		`CREATE INDEX IF NOT EXISTS idx_readiness_day ON readiness_metrics(day)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(ctx, query); err != nil {
			return err
		}
	}

	r.logger.Info("Database schema initialized")
	return nil
}

type SleepMetric struct {
	OuraID   string
	Day      time.Time
	Score    int
	Duration int
}

type ActivityMetric struct {
	OuraID              string
	Day                 time.Time
	Score               int
	ActiveCalories      int
	Steps               int
	MediumActivityMin   int
	HighActivityMin     int
}

type ReadinessMetric struct {
	OuraID string
	Day    time.Time
	Score  int
}

func (r *Repository) SaveSleepMetric(ctx context.Context, metric *SleepMetric) error {
	query := `
		INSERT INTO sleep_metrics (oura_id, day, score, duration)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (oura_id) 
		DO UPDATE SET score = $3, duration = $4, updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(ctx, query, metric.OuraID, metric.Day, metric.Score, metric.Duration)
	return err
}

func (r *Repository) SaveActivityMetric(ctx context.Context, metric *ActivityMetric) error {
	query := `
		INSERT INTO activity_metrics (
			oura_id, day, score, active_calories, steps, 
			medium_activity_minutes, high_activity_minutes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (oura_id)
		DO UPDATE SET 
			score = $3, 
			active_calories = $4, 
			steps = $5, 
			medium_activity_minutes = $6, 
			high_activity_minutes = $7,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(ctx, query,
		metric.OuraID, metric.Day, metric.Score, metric.ActiveCalories,
		metric.Steps, metric.MediumActivityMin, metric.HighActivityMin,
	)
	return err
}

func (r *Repository) SaveReadinessMetric(ctx context.Context, metric *ReadinessMetric) error {
	query := `
		INSERT INTO readiness_metrics (oura_id, day, score)
		VALUES ($1, $2, $3)
		ON CONFLICT (oura_id)
		DO UPDATE SET score = $3, updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(ctx, query, metric.OuraID, metric.Day, metric.Score)
	return err
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
