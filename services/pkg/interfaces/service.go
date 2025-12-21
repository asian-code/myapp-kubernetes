package interfaces

import (
	"context"
	"time"
)

// UserService defines business logic for user operations
type UserService interface {
	// Authentication
	Register(ctx context.Context, username, email, password string) (*UserDTO, error)
	Login(ctx context.Context, username, password string) (string, error) // Returns JWT token
	GetProfile(ctx context.Context, userID string) (*UserDTO, error)
	UpdateProfile(ctx context.Context, userID string, updates *UserUpdateDTO) error
	DeleteAccount(ctx context.Context, userID string) error
}

// OAuthService defines business logic for OAuth operations
type OAuthService interface {
	// OAuth flow
	GenerateAuthURL(ctx context.Context, userID string) (string, error)
	HandleCallback(ctx context.Context, code, state string) (*OAuthResult, error)
	RefreshAccessToken(ctx context.Context, userID, provider string) error
	RevokeToken(ctx context.Context, userID, provider string) error
}

// MetricsService defines business logic for health metrics
type MetricsService interface {
	// Data ingestion
	IngestSleepData(ctx context.Context, userID string, data *SleepDataDTO) error
	IngestActivityData(ctx context.Context, userID string, data *ActivityDataDTO) error
	IngestReadinessData(ctx context.Context, userID string, data *ReadinessDataDTO) error

	// Data retrieval
	GetSleepHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]*SleepDataDTO, error)
	GetActivityHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]*ActivityDataDTO, error)
	GetReadinessHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]*ReadinessDataDTO, error)
	GetDashboard(ctx context.Context, userID string, days int) (*DashboardDTO, error)
}

// OuraClient defines operations for interacting with Oura API
type OuraClient interface {
	// Data collection
	GetSleepData(ctx context.Context, date string) (*OuraSleepResponse, error)
	GetActivityData(ctx context.Context, date string) (*OuraActivityResponse, error)
	GetReadinessData(ctx context.Context, date string) (*OuraReadinessResponse, error)
	GetUserInfo(ctx context.Context) (*OuraUserInfo, error)

	// Token management
	ExchangeCode(ctx context.Context, code string) (*OuraTokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*OuraTokenResponse, error)
}

// Logger defines logging operations
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
	Fatal(msg string, err error, fields map[string]interface{})
	WithFields(fields map[string]interface{}) Logger
	WithContext(ctx context.Context) Logger
}

// MetricsCollector defines operations for collecting application metrics
type MetricsCollector interface {
	IncrementCounter(name string, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
	ObserveDuration(name string, duration time.Duration, labels map[string]string)
}

// DTOs (Data Transfer Objects)

type UserDTO struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	IsActive  bool       `json:"is_active"`
}

type UserUpdateDTO struct {
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type OAuthResult struct {
	UserID      string
	AccessToken string
	ExpiresAt   time.Time
}

type SleepDataDTO struct {
	OuraID   string    `json:"oura_id"`
	Day      time.Time `json:"day"`
	Score    int       `json:"score"`
	Duration int       `json:"duration"`
}

type ActivityDataDTO struct {
	OuraID            string    `json:"oura_id"`
	Day               time.Time `json:"day"`
	Score             int       `json:"score"`
	ActiveCalories    int       `json:"active_calories"`
	Steps             int       `json:"steps"`
	MediumActivityMin int       `json:"medium_activity_minutes"`
	HighActivityMin   int       `json:"high_activity_minutes"`
}

type ReadinessDataDTO struct {
	OuraID string    `json:"oura_id"`
	Day    time.Time `json:"day"`
	Score  int       `json:"score"`
}

type OAuthTokenDTO struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
	TokenType    string    `json:"token_type"`
}

type DashboardDTO struct {
	Summary       *DashboardSummaryDTO `json:"summary"`
	RecentSleep   []*SleepDataDTO      `json:"recent_sleep"`
	RecentActivity []*ActivityDataDTO   `json:"recent_activity"`
	RecentReadiness []*ReadinessDataDTO `json:"recent_readiness"`
}

type DashboardSummaryDTO struct {
	TotalDays         int     `json:"total_days"`
	AvgSleepScore     float64 `json:"avg_sleep_score"`
	AvgActivityScore  float64 `json:"avg_activity_score"`
	AvgReadinessScore float64 `json:"avg_readiness_score"`
	TotalSteps        int     `json:"total_steps"`
	AvgSleepDuration  float64 `json:"avg_sleep_duration_hours"`
}

// Oura API response types

type OuraSleepResponse struct {
	ID       string `json:"id"`
	Day      string `json:"day"`
	Score    int    `json:"score"`
	Duration int    `json:"duration"`
}

type OuraActivityResponse struct {
	ID                    string `json:"id"`
	Day                   string `json:"day"`
	Score                 int    `json:"score"`
	ActiveCalories        int    `json:"active_calories"`
	Steps                 int    `json:"steps"`
	MediumActivityMinutes int    `json:"medium_met_minutes"`
	HighActivityMinutes   int    `json:"high_met_minutes"`
}

type OuraReadinessResponse struct {
	ID    string `json:"id"`
	Day   string `json:"day"`
	Score int    `json:"score"`
}

type OuraTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type OuraUserInfo struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}
