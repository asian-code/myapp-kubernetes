package metrics

import (
	"context"
	"time"

	"github.com/asian-code/myapp-kubernetes/services/pkg/errors"
	"github.com/asian-code/myapp-kubernetes/services/pkg/interfaces"
	"github.com/asian-code/myapp-kubernetes/services/pkg/validation"
)

type service struct {
	repo   interfaces.MetricsRepository
	logger interfaces.Logger
}

// NewService creates a new metrics service instance
func NewService(repo interfaces.MetricsRepository, logger interfaces.Logger) interfaces.MetricsService {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// IngestSleepData validates and saves sleep metrics to the database
func (s *service) IngestSleepData(ctx context.Context, userID string, data *interfaces.SleepDataDTO) error {
	if userID == "" {
		return errors.New(errors.ErrCodeBadRequest, "user ID is required")
	}

	// Validate input
	if err := validation.Validate(data); err != nil {
		return errors.Wrap(err, errors.ErrCodeBadRequest, "invalid sleep data")
	}

	// Additional business validation
	if data.Score < 0 || data.Score > 100 {
		return errors.New(errors.ErrCodeBadRequest, "sleep score must be between 0 and 100")
	}

	if data.Duration < 0 {
		return errors.New(errors.ErrCodeBadRequest, "sleep duration cannot be negative")
	}

	// Convert DTO to entity
	metric := &interfaces.SleepMetric{
		UserID:   userID,
		OuraID:   data.OuraID,
		Day:      data.Day,
		Score:    data.Score,
		Duration: data.Duration,
	}

	// Save to database
	if err := s.repo.SaveSleepMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to save sleep metric", err, map[string]interface{}{
			"user_id": userID,
			"day":     data.Day,
		})
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to save sleep data")
	}

	s.logger.Info("Sleep metric ingested successfully", map[string]interface{}{
		"user_id": userID,
		"day":     data.Day,
		"score":   data.Score,
	})

	return nil
}

// IngestActivityData validates and saves activity metrics to the database
func (s *service) IngestActivityData(ctx context.Context, userID string, data *interfaces.ActivityDataDTO) error {
	if userID == "" {
		return errors.New(errors.ErrCodeBadRequest, "user ID is required")
	}

	// Validate input
	if err := validation.Validate(data); err != nil {
		return errors.Wrap(err, errors.ErrCodeBadRequest, "invalid activity data")
	}

	// Additional business validation
	if data.Score < 0 || data.Score > 100 {
		return errors.New(errors.ErrCodeBadRequest, "activity score must be between 0 and 100")
	}

	if data.ActiveCalories < 0 || data.Steps < 0 {
		return errors.New(errors.ErrCodeBadRequest, "calories and steps cannot be negative")
	}

	if data.MediumActivityMin < 0 || data.HighActivityMin < 0 {
		return errors.New(errors.ErrCodeBadRequest, "activity minutes cannot be negative")
	}

	// Convert DTO to entity
	metric := &interfaces.ActivityMetric{
		UserID:            userID,
		OuraID:            data.OuraID,
		Day:               data.Day,
		Score:             data.Score,
		ActiveCalories:    data.ActiveCalories,
		Steps:             data.Steps,
		MediumActivityMin: data.MediumActivityMin,
		HighActivityMin:   data.HighActivityMin,
	}

	// Save to database
	if err := s.repo.SaveActivityMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to save activity metric", err, map[string]interface{}{
			"user_id": userID,
			"day":     data.Day,
		})
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to save activity data")
	}

	s.logger.Info("Activity metric ingested successfully", map[string]interface{}{
		"user_id": userID,
		"day":     data.Day,
		"score":   data.Score,
		"steps":   data.Steps,
	})

	return nil
}

// IngestReadinessData validates and saves readiness metrics to the database
func (s *service) IngestReadinessData(ctx context.Context, userID string, data *interfaces.ReadinessDataDTO) error {
	if userID == "" {
		return errors.New(errors.ErrCodeBadRequest, "user ID is required")
	}

	// Validate input
	if err := validation.Validate(data); err != nil {
		return errors.Wrap(err, errors.ErrCodeBadRequest, "invalid readiness data")
	}

	// Additional business validation
	if data.Score < 0 || data.Score > 100 {
		return errors.New(errors.ErrCodeBadRequest, "readiness score must be between 0 and 100")
	}

	// Convert DTO to entity
	metric := &interfaces.ReadinessMetric{
		UserID: userID,
		OuraID: data.OuraID,
		Day:    data.Day,
		Score:  data.Score,
	}

	// Save to database
	if err := s.repo.SaveReadinessMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to save readiness metric", err, map[string]interface{}{
			"user_id": userID,
			"day":     data.Day,
		})
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to save readiness data")
	}

	s.logger.Info("Readiness metric ingested successfully", map[string]interface{}{
		"user_id": userID,
		"day":     data.Day,
		"score":   data.Score,
	})

	return nil
}

// GetSleepHistory retrieves sleep metrics for a date range
func (s *service) GetSleepHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]*interfaces.SleepDataDTO, error) {
	if userID == "" {
		return nil, errors.New(errors.ErrCodeBadRequest, "user ID is required")
	}

	if endDate.Before(startDate) {
		return nil, errors.New(errors.ErrCodeBadRequest, "end date must be after start date")
	}

	// Limit date range to prevent excessive queries
	if endDate.Sub(startDate) > 365*24*time.Hour {
		return nil, errors.New(errors.ErrCodeBadRequest, "date range cannot exceed 365 days")
	}

	metrics, err := s.repo.GetSleepMetrics(ctx, userID, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to retrieve sleep metrics", err, map[string]interface{}{
			"user_id":    userID,
			"start_date": startDate,
			"end_date":   endDate,
		})
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to retrieve sleep history")
	}

	// Convert entities to DTOs
	dtos := make([]*interfaces.SleepDataDTO, 0, len(metrics))
	for _, m := range metrics {
		dtos = append(dtos, &interfaces.SleepDataDTO{
			OuraID:   m.OuraID,
			Day:      m.Day,
			Score:    m.Score,
			Duration: m.Duration,
		})
	}

	return dtos, nil
}

// GetActivityHistory retrieves activity metrics for a date range
func (s *service) GetActivityHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]*interfaces.ActivityDataDTO, error) {
	if userID == "" {
		return nil, errors.New(errors.ErrCodeBadRequest, "user ID is required")
	}

	if endDate.Before(startDate) {
		return nil, errors.New(errors.ErrCodeBadRequest, "end date must be after start date")
	}

	// Limit date range to prevent excessive queries
	if endDate.Sub(startDate) > 365*24*time.Hour {
		return nil, errors.New(errors.ErrCodeBadRequest, "date range cannot exceed 365 days")
	}

	metrics, err := s.repo.GetActivityMetrics(ctx, userID, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to retrieve activity metrics", err, map[string]interface{}{
			"user_id":    userID,
			"start_date": startDate,
			"end_date":   endDate,
		})
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to retrieve activity history")
	}

	// Convert entities to DTOs
	dtos := make([]*interfaces.ActivityDataDTO, 0, len(metrics))
	for _, m := range metrics {
		dtos = append(dtos, &interfaces.ActivityDataDTO{
			OuraID:            m.OuraID,
			Day:               m.Day,
			Score:             m.Score,
			ActiveCalories:    m.ActiveCalories,
			Steps:             m.Steps,
			MediumActivityMin: m.MediumActivityMin,
			HighActivityMin:   m.HighActivityMin,
		})
	}

	return dtos, nil
}

// GetReadinessHistory retrieves readiness metrics for a date range
func (s *service) GetReadinessHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]*interfaces.ReadinessDataDTO, error) {
	if userID == "" {
		return nil, errors.New(errors.ErrCodeBadRequest, "user ID is required")
	}

	if endDate.Before(startDate) {
		return nil, errors.New(errors.ErrCodeBadRequest, "end date must be after start date")
	}

	// Limit date range to prevent excessive queries
	if endDate.Sub(startDate) > 365*24*time.Hour {
		return nil, errors.New(errors.ErrCodeBadRequest, "date range cannot exceed 365 days")
	}

	metrics, err := s.repo.GetReadinessMetrics(ctx, userID, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to retrieve readiness metrics", err, map[string]interface{}{
			"user_id":    userID,
			"start_date": startDate,
			"end_date":   endDate,
		})
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to retrieve readiness history")
	}

	// Convert entities to DTOs
	dtos := make([]*interfaces.ReadinessDataDTO, 0, len(metrics))
	for _, m := range metrics {
		dtos = append(dtos, &interfaces.ReadinessDataDTO{
			OuraID: m.OuraID,
			Day:    m.Day,
			Score:  m.Score,
		})
	}

	return dtos, nil
}

// GetDashboard retrieves aggregated dashboard data for the specified number of days
func (s *service) GetDashboard(ctx context.Context, userID string, days int) (*interfaces.DashboardDTO, error) {
	if userID == "" {
		return nil, errors.New(errors.ErrCodeBadRequest, "user ID is required")
	}

	if days <= 0 {
		return nil, errors.New(errors.ErrCodeBadRequest, "days must be greater than 0")
	}

	// Limit to prevent excessive queries
	if days > 365 {
		return nil, errors.New(errors.ErrCodeBadRequest, "days cannot exceed 365")
	}

	// Get aggregated summary
	summary, err := s.repo.GetDashboardSummary(ctx, userID, days)
	if err != nil {
		s.logger.Error("Failed to retrieve dashboard summary", err, map[string]interface{}{
			"user_id": userID,
			"days":    days,
		})
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to retrieve dashboard data")
	}

	// Calculate date range for recent metrics
	endDate := time.Now().UTC()

	// Get recent metrics (last 7 days regardless of request)
	recentDays := 7
	if days < 7 {
		recentDays = days
	}
	recentStartDate := endDate.AddDate(0, 0, -recentDays)

	// Fetch recent data in parallel would be ideal, but keeping it simple for now
	recentSleep, err := s.GetSleepHistory(ctx, userID, recentStartDate, endDate)
	if err != nil {
		// Log but don't fail the entire request
		s.logger.Warn("Failed to retrieve recent sleep data for dashboard", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		recentSleep = []*interfaces.SleepDataDTO{}
	}

	recentActivity, err := s.GetActivityHistory(ctx, userID, recentStartDate, endDate)
	if err != nil {
		s.logger.Warn("Failed to retrieve recent activity data for dashboard", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		recentActivity = []*interfaces.ActivityDataDTO{}
	}

	recentReadiness, err := s.GetReadinessHistory(ctx, userID, recentStartDate, endDate)
	if err != nil {
		s.logger.Warn("Failed to retrieve recent readiness data for dashboard", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		recentReadiness = []*interfaces.ReadinessDataDTO{}
	}

	// Build dashboard response
	dashboard := &interfaces.DashboardDTO{
		Summary: &interfaces.DashboardSummaryDTO{
			TotalDays:         summary.TotalDays,
			AvgSleepScore:     summary.AvgSleepScore,
			AvgActivityScore:  summary.AvgActivityScore,
			AvgReadinessScore: summary.AvgReadinessScore,
			TotalSteps:        summary.TotalSteps,
			AvgSleepDuration:  summary.AvgSleepDuration,
		},
		RecentSleep:     recentSleep,
		RecentActivity:  recentActivity,
		RecentReadiness: recentReadiness,
	}

	s.logger.Info("Dashboard data retrieved successfully", map[string]interface{}{
		"user_id":           userID,
		"days":              days,
		"recent_sleep":      len(recentSleep),
		"recent_activity":   len(recentActivity),
		"recent_readiness":  len(recentReadiness),
	})

	return dashboard, nil
}
