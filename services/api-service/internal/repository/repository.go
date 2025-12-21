package repository

import (
	"context"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/pkg/interfaces"
	"github.com/jackc/pgx/v5"
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

// User repository methods

func (r *Repository) CreateUser(ctx context.Context, username, email, passwordHash string) (string, error) {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var userID string
	err := r.db.QueryRow(ctx, query, username, email, passwordHash).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*interfaces.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, last_login
		FROM users
		WHERE id = $1
	`

	var user interfaces.User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*interfaces.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, last_login
		FROM users
		WHERE username = $1
	`

	var user interfaces.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*interfaces.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, last_login
		FROM users
		WHERE email = $1
	`

	var user interfaces.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID string, loginTime time.Time) error {
	query := `
		UPDATE users
		SET last_login = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, loginTime, time.Now(), userID)
	return err
}

func (r *Repository) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	// Build dynamic update query based on provided fields
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE users SET "
	args := []interface{}{}
	argNum := 1

	for field, value := range updates {
		if argNum > 1 {
			query += ", "
		}
		query += field + " = $" + string(rune(argNum+'0'))
		args = append(args, value)
		argNum++
	}

	// Always update updated_at
	query += ", updated_at = $" + string(rune(argNum+'0'))
	args = append(args, time.Now())
	argNum++

	query += " WHERE id = $" + string(rune(argNum+'0'))
	args = append(args, userID)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// OAuth repository methods

func (r *Repository) SaveToken(ctx context.Context, token *interfaces.OAuthToken) error {
	query := `
		INSERT INTO oauth_tokens (user_id, access_token, refresh_token, expires_at, scope, provider)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, provider) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			scope = EXCLUDED.scope,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		token.UserID,
		token.AccessToken,
		token.RefreshToken,
		token.ExpiresAt,
		token.Scope,
		token.Provider,
	)
	return err
}

func (r *Repository) GetToken(ctx context.Context, userID, provider string) (*interfaces.OAuthToken, error) {
	query := `
		SELECT user_id, access_token, refresh_token, expires_at, scope, provider, created_at, updated_at
		FROM oauth_tokens
		WHERE user_id = $1 AND provider = $2
	`

	var token interfaces.OAuthToken
	err := r.db.QueryRow(ctx, query, userID, provider).Scan(
		&token.UserID,
		&token.AccessToken,
		&token.RefreshToken,
		&token.ExpiresAt,
		&token.Scope,
		&token.Provider,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *Repository) UpdateToken(ctx context.Context, token *interfaces.OAuthToken) error {
	query := `
		UPDATE oauth_tokens
		SET access_token = $1, refresh_token = $2, expires_at = $3, 
		    scope = $4, provider = $5, updated_at = NOW()
		WHERE user_id = $6
	`

	result, err := r.db.Exec(ctx, query,
		token.AccessToken,
		token.RefreshToken,
		token.ExpiresAt,
		token.Scope,
		token.Provider,
		token.UserID,
	)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *Repository) DeleteToken(ctx context.Context, userID, provider string) error {
	query := `DELETE FROM oauth_tokens WHERE user_id = $1 AND provider = $2`
	_, err := r.db.Exec(ctx, query, userID, provider)
	return err
}

func (r *Repository) RefreshToken(ctx context.Context, userID, provider string, newAccessToken, newRefreshToken string, expiresAt time.Time) error {
	query := `
		UPDATE oauth_tokens
		SET access_token = $1, refresh_token = $2, expires_at = $3, updated_at = NOW()
		WHERE user_id = $4 AND provider = $5
	`

	result, err := r.db.Exec(ctx, query, newAccessToken, newRefreshToken, expiresAt, userID, provider)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
