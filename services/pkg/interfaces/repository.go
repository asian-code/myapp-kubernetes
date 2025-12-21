package interfaces

import (
	"context"
	"time"
)

// UserRepository defines operations for user management
type UserRepository interface {
	// User CRUD operations
	CreateUser(ctx context.Context, username, email, passwordHash string) (string, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
}

// OAuthRepository defines operations for OAuth token management
type OAuthRepository interface {
	// OAuth token operations
	SaveToken(ctx context.Context, token *OAuthToken) error
	GetToken(ctx context.Context, userID, provider string) (*OAuthToken, error)
	UpdateToken(ctx context.Context, token *OAuthToken) error
	DeleteToken(ctx context.Context, userID, provider string) error
	RefreshToken(ctx context.Context, userID, provider string, newAccessToken, newRefreshToken string, expiresAt time.Time) error
}

// MetricsRepository defines operations for health metrics
type MetricsRepository interface {
	// Sleep metrics
	SaveSleepMetric(ctx context.Context, metric *SleepMetric) error
	GetSleepMetrics(ctx context.Context, userID string, startDate, endDate time.Time) ([]*SleepMetric, error)
	GetSleepMetricByID(ctx context.Context, metricID string) (*SleepMetric, error)

	// Activity metrics
	SaveActivityMetric(ctx context.Context, metric *ActivityMetric) error
	GetActivityMetrics(ctx context.Context, userID string, startDate, endDate time.Time) ([]*ActivityMetric, error)
	GetActivityMetricByID(ctx context.Context, metricID string) (*ActivityMetric, error)

	// Readiness metrics
	SaveReadinessMetric(ctx context.Context, metric *ReadinessMetric) error
	GetReadinessMetrics(ctx context.Context, userID string, startDate, endDate time.Time) ([]*ReadinessMetric, error)
	GetReadinessMetricByID(ctx context.Context, metricID string) (*ReadinessMetric, error)

	// Dashboard aggregations
	GetDashboardSummary(ctx context.Context, userID string, days int) (*DashboardSummary, error)
}

// User represents a user entity
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLogin    *time.Time
	IsActive     bool
}

// OAuthToken represents an OAuth token entity
type OAuthToken struct {
	ID           string
	UserID       string
	Provider     string
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	Scope        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SleepMetric represents sleep data
type SleepMetric struct {
	ID        string
	UserID    string
	OuraID    string
	Day       time.Time
	Score     int
	Duration  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ActivityMetric represents activity data
type ActivityMetric struct {
	ID                string
	UserID            string
	OuraID            string
	Day               time.Time
	Score             int
	ActiveCalories    int
	Steps             int
	MediumActivityMin int
	HighActivityMin   int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ReadinessMetric represents readiness data
type ReadinessMetric struct {
	ID        string
	UserID    string
	OuraID    string
	Day       time.Time
	Score     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DashboardSummary aggregates metrics for dashboard display
type DashboardSummary struct {
	TotalDays         int
	AvgSleepScore     float64
	AvgActivityScore  float64
	AvgReadinessScore float64
	TotalSteps        int
	AvgSleepDuration  float64
}
