package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/pkg/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricsRepository implements interfaces.MetricsRepository
type MockMetricsRepository struct {
	mock.Mock
}

func (m *MockMetricsRepository) SaveSleepMetric(ctx context.Context, metric *interfaces.SleepMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockMetricsRepository) GetSleepMetrics(ctx context.Context, userID string, startDate, endDate time.Time) ([]*interfaces.SleepMetric, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*interfaces.SleepMetric), args.Error(1)
}

func (m *MockMetricsRepository) GetSleepMetricByID(ctx context.Context, metricID string) (*interfaces.SleepMetric, error) {
	args := m.Called(ctx, metricID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.SleepMetric), args.Error(1)
}

func (m *MockMetricsRepository) SaveActivityMetric(ctx context.Context, metric *interfaces.ActivityMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockMetricsRepository) GetActivityMetrics(ctx context.Context, userID string, startDate, endDate time.Time) ([]*interfaces.ActivityMetric, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*interfaces.ActivityMetric), args.Error(1)
}

func (m *MockMetricsRepository) GetActivityMetricByID(ctx context.Context, metricID string) (*interfaces.ActivityMetric, error) {
	args := m.Called(ctx, metricID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.ActivityMetric), args.Error(1)
}

func (m *MockMetricsRepository) SaveReadinessMetric(ctx context.Context, metric *interfaces.ReadinessMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockMetricsRepository) GetReadinessMetrics(ctx context.Context, userID string, startDate, endDate time.Time) ([]*interfaces.ReadinessMetric, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*interfaces.ReadinessMetric), args.Error(1)
}

func (m *MockMetricsRepository) GetReadinessMetricByID(ctx context.Context, metricID string) (*interfaces.ReadinessMetric, error) {
	args := m.Called(ctx, metricID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.ReadinessMetric), args.Error(1)
}

func (m *MockMetricsRepository) GetDashboardSummary(ctx context.Context, userID string, days int) (*interfaces.DashboardSummary, error) {
	args := m.Called(ctx, userID, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.DashboardSummary), args.Error(1)
}

// MockLogger implements interfaces.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, err error, fields map[string]interface{}) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) Fatal(msg string, err error, fields map[string]interface{}) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) interfaces.Logger {
	args := m.Called(fields)
	return args.Get(0).(interfaces.Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) interfaces.Logger {
	args := m.Called(ctx)
	return args.Get(0).(interfaces.Logger)
}

// Test IngestSleepData
func TestMetricsService_IngestSleepData_Success(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	sleepData := &interfaces.SleepDataDTO{
		OuraID:   "oura-sleep-123",
		Day:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Score:    85,
		Duration: 28800, // 8 hours in seconds
	}

	mockRepo.On("SaveSleepMetric", ctx, mock.AnythingOfType("*interfaces.SleepMetric")).Return(nil)
	mockLogger.On("Info", "Sleep metric ingested successfully", mock.Anything).Return()

	err := service.IngestSleepData(ctx, userID, sleepData)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestMetricsService_IngestSleepData_MissingUserID(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	sleepData := &interfaces.SleepDataDTO{
		OuraID:   "oura-sleep-123",
		Day:      time.Now(),
		Score:    85,
		Duration: 28800,
	}

	err := service.IngestSleepData(ctx, "", sleepData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestMetricsService_IngestSleepData_InvalidScore(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	sleepData := &interfaces.SleepDataDTO{
		OuraID:   "oura-sleep-123",
		Day:      time.Now(),
		Score:    150, // Invalid: exceeds 100
		Duration: 28800,
	}

	err := service.IngestSleepData(ctx, userID, sleepData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sleep score must be between 0 and 100")
}

func TestMetricsService_IngestSleepData_NegativeDuration(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	sleepData := &interfaces.SleepDataDTO{
		OuraID:   "oura-sleep-123",
		Day:      time.Now(),
		Score:    85,
		Duration: -100, // Invalid: negative
	}

	err := service.IngestSleepData(ctx, userID, sleepData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sleep duration cannot be negative")
}

// Test IngestActivityData
func TestMetricsService_IngestActivityData_Success(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	activityData := &interfaces.ActivityDataDTO{
		OuraID:            "oura-activity-123",
		Day:               time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Score:             90,
		ActiveCalories:    500,
		Steps:             10000,
		MediumActivityMin: 30,
		HighActivityMin:   20,
	}

	mockRepo.On("SaveActivityMetric", ctx, mock.AnythingOfType("*interfaces.ActivityMetric")).Return(nil)
	mockLogger.On("Info", "Activity metric ingested successfully", mock.Anything).Return()

	err := service.IngestActivityData(ctx, userID, activityData)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestMetricsService_IngestActivityData_InvalidScore(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	activityData := &interfaces.ActivityDataDTO{
		OuraID:            "oura-activity-123",
		Day:               time.Now(),
		Score:             -10, // Invalid: negative
		ActiveCalories:    500,
		Steps:             10000,
		MediumActivityMin: 30,
		HighActivityMin:   20,
	}

	err := service.IngestActivityData(ctx, userID, activityData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "activity score must be between 0 and 100")
}

func TestMetricsService_IngestActivityData_NegativeSteps(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	activityData := &interfaces.ActivityDataDTO{
		OuraID:            "oura-activity-123",
		Day:               time.Now(),
		Score:             90,
		ActiveCalories:    500,
		Steps:             -1000, // Invalid: negative
		MediumActivityMin: 30,
		HighActivityMin:   20,
	}

	err := service.IngestActivityData(ctx, userID, activityData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "calories and steps cannot be negative")
}

// Test IngestReadinessData
func TestMetricsService_IngestReadinessData_Success(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	readinessData := &interfaces.ReadinessDataDTO{
		OuraID: "oura-readiness-123",
		Day:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Score:  75,
	}

	mockRepo.On("SaveReadinessMetric", ctx, mock.AnythingOfType("*interfaces.ReadinessMetric")).Return(nil)
	mockLogger.On("Info", "Readiness metric ingested successfully", mock.Anything).Return()

	err := service.IngestReadinessData(ctx, userID, readinessData)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestMetricsService_IngestReadinessData_InvalidScore(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	readinessData := &interfaces.ReadinessDataDTO{
		OuraID: "oura-readiness-123",
		Day:    time.Now(),
		Score:  101, // Invalid: exceeds 100
	}

	err := service.IngestReadinessData(ctx, userID, readinessData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readiness score must be between 0 and 100")
}

// Test GetSleepHistory
func TestMetricsService_GetSleepHistory_Success(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	mockMetrics := []*interfaces.SleepMetric{
		{
			ID:       "metric-1",
			UserID:   userID,
			OuraID:   "oura-1",
			Day:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Score:    85,
			Duration: 28800,
		},
	}

	mockRepo.On("GetSleepMetrics", ctx, userID, startDate, endDate).Return(mockMetrics, nil)

	result, err := service.GetSleepHistory(ctx, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "oura-1", result[0].OuraID)
	assert.Equal(t, 85, result[0].Score)
	mockRepo.AssertExpectations(t)
}

func TestMetricsService_GetSleepHistory_InvalidDateRange(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	startDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // End before start

	result, err := service.GetSleepHistory(ctx, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "end date must be after start date")
}

func TestMetricsService_GetSleepHistory_DateRangeTooLarge(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // > 365 days

	result, err := service.GetSleepHistory(ctx, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "date range cannot exceed 365 days")
}

// Test GetDashboard
func TestMetricsService_GetDashboard_Success(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	days := 30

	mockSummary := &interfaces.DashboardSummary{
		TotalDays:         30,
		AvgSleepScore:     85.5,
		AvgActivityScore:  90.0,
		AvgReadinessScore: 80.0,
		TotalSteps:        300000,
		AvgSleepDuration:  8.2,
	}

	mockRepo.On("GetDashboardSummary", ctx, userID, days).Return(mockSummary, nil)
	mockRepo.On("GetSleepMetrics", ctx, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return([]*interfaces.SleepMetric{}, nil)
	mockRepo.On("GetActivityMetrics", ctx, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return([]*interfaces.ActivityMetric{}, nil)
	mockRepo.On("GetReadinessMetrics", ctx, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return([]*interfaces.ReadinessMetric{}, nil)
	mockLogger.On("Info", "Dashboard data retrieved successfully", mock.Anything).Return()

	result, err := service.GetDashboard(ctx, userID, days)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Summary)
	assert.Equal(t, 30, result.Summary.TotalDays)
	assert.Equal(t, 85.5, result.Summary.AvgSleepScore)
	assert.Equal(t, 90.0, result.Summary.AvgActivityScore)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestMetricsService_GetDashboard_InvalidDays(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	days := 0 // Invalid: must be > 0

	result, err := service.GetDashboard(ctx, userID, days)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "days must be greater than 0")
}

func TestMetricsService_GetDashboard_DaysExceedsLimit(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	mockLogger := new(MockLogger)
	service := NewService(mockRepo, mockLogger)

	ctx := context.Background()
	userID := "user-123"
	days := 500 // Invalid: exceeds 365

	result, err := service.GetDashboard(ctx, userID, days)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "days cannot exceed 365")
}
