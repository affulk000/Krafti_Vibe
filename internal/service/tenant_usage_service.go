package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Usage tracking errors
var (
	// ErrAPIRateLimitExceeded is returned when API rate limit is exceeded
	ErrAPIRateLimitExceeded = errors.New("API rate limit exceeded")

	// ErrInvalidFeature is returned when an unknown feature is tracked
	ErrInvalidFeature = errors.New("invalid feature name")
)

// TenantUsageService defines the interface for tenant usage tracking operations
type TenantUsageService interface {
	// Daily Usage Operations
	GetDailyUsage(ctx context.Context, tenantID uuid.UUID, date time.Time) (*dto.UsageTrackingResponse, error)
	GetUsageHistory(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.UsageHistoryResponse, error)

	// API Usage
	IncrementAPIUsage(ctx context.Context, tenantID uuid.UUID) error
	CheckAPIRateLimit(ctx context.Context, tenantID uuid.UUID) (bool, error)
	GetAPIUsageStats(ctx context.Context, tenantID uuid.UUID, days int) (*dto.UsageSummary, error)

	// Feature Usage Tracking
	TrackFeatureUsage(ctx context.Context, req *dto.TrackFeatureUsageRequest, tenantID uuid.UUID) error
	TrackBookingCreated(ctx context.Context, tenantID uuid.UUID) error
	TrackProjectCreated(ctx context.Context, tenantID uuid.UUID) error
	TrackSMSSent(ctx context.Context, tenantID uuid.UUID, count int) error
	TrackEmailSent(ctx context.Context, tenantID uuid.UUID, count int) error
	TrackStorageUsage(ctx context.Context, tenantID uuid.UUID, bytesUsed int64) error
	TrackBandwidthUsage(ctx context.Context, tenantID uuid.UUID, bytesUsed int64) error

	// Usage Analytics
	GetPeakUsage(ctx context.Context, tenantID uuid.UUID, days int) (*dto.UsageTrackingResponse, error)
	GetAverageUsage(ctx context.Context, tenantID uuid.UUID, days int) (*dto.UsageSummary, error)

	// Cleanup Operations
	DeleteOldUsageRecords(ctx context.Context, olderThanDays int) (int64, error)
}

// tenantUsageService implements TenantUsageService
type tenantUsageService struct {
	repos  *repository.Repositories
	logger *zap.Logger
}

// NewTenantUsageService creates a new tenant usage service
func NewTenantUsageService(
	repos *repository.Repositories,
	logger *zap.Logger,
) TenantUsageService {
	return &tenantUsageService{
		repos:  repos,
		logger: logger,
	}
}

// ============================================================================
// Daily Usage Operations
// ============================================================================

// GetDailyUsage retrieves usage tracking for a specific date
func (s *tenantUsageService) GetDailyUsage(ctx context.Context, tenantID uuid.UUID, date time.Time) (*dto.UsageTrackingResponse, error) {
	normalizedDate := date.Truncate(24 * time.Hour)

	usage, err := s.repos.TenantUsageTracking.FindByTenantAndDate(ctx, tenantID, normalizedDate)
	if err != nil {
		// Return empty usage if none exists
		apiLimit := s.getAPILimitForTenant(ctx, tenantID)
		return &dto.UsageTrackingResponse{
			TenantID:        tenantID,
			Date:            normalizedDate,
			APICallsCount:   0,
			APICallsLimit:   apiLimit,
			APIUsagePercent: 0,
		}, nil
	}

	return dto.ToUsageTrackingResponse(usage), nil
}

// GetUsageHistory retrieves usage history for a date range with summary
func (s *tenantUsageService) GetUsageHistory(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.UsageHistoryResponse, error) {
	// Normalize dates
	startDate = startDate.Truncate(24 * time.Hour)
	endDate = endDate.Truncate(24 * time.Hour)

	// Validate date range
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Limit range to 365 days
	maxRange := 365 * 24 * time.Hour
	if endDate.Sub(startDate) > maxRange {
		return nil, fmt.Errorf("date range cannot exceed 365 days")
	}

	usages, err := s.repos.TenantUsageTracking.FindByTenantAndDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get usage history", zap.Error(err))
		return nil, fmt.Errorf("failed to get usage history: %w", err)
	}

	// Convert to response DTOs
	dailyUsage := make([]*dto.UsageTrackingResponse, 0, len(usages))
	for _, usage := range usages {
		dailyUsage = append(dailyUsage, dto.ToUsageTrackingResponse(usage))
	}

	// Calculate summary
	summary := s.calculateUsageSummary(usages)

	return &dto.UsageHistoryResponse{
		TenantID:   tenantID,
		StartDate:  startDate,
		EndDate:    endDate,
		DailyUsage: dailyUsage,
		Summary:    summary,
	}, nil
}

// ============================================================================
// API Usage
// ============================================================================

// IncrementAPIUsage increments the API call count for today
func (s *tenantUsageService) IncrementAPIUsage(ctx context.Context, tenantID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)

	usage, err := s.repos.TenantUsageTracking.FindByTenantAndDate(ctx, tenantID, today)
	if err != nil {
		// Create new usage record for today
		apiLimit := s.getAPILimitForTenant(ctx, tenantID)
		usage = &models.TenantUsageTracking{
			TenantID:      tenantID,
			Date:          today,
			APICallsCount: 0,
			APICallsLimit: apiLimit,
		}
		if err := s.repos.TenantUsageTracking.Create(ctx, usage); err != nil {
			s.logger.Error("failed to create usage record", zap.Error(err))
			return fmt.Errorf("failed to create usage record: %w", err)
		}
	}

	// Check if we can make the API call
	if !usage.CanMakeAPICall() {
		s.logger.Warn("API rate limit exceeded",
			zap.String("tenant_id", tenantID.String()),
			zap.Int64("current", usage.APICallsCount),
			zap.Int64("limit", usage.APICallsLimit),
		)
		return ErrAPIRateLimitExceeded
	}

	// Increment API call
	if err := usage.IncrementAPICall(); err != nil {
		return ErrAPIRateLimitExceeded
	}

	if err := s.repos.TenantUsageTracking.Update(ctx, usage); err != nil {
		s.logger.Error("failed to update usage", zap.Error(err))
		return fmt.Errorf("failed to update usage: %w", err)
	}

	return nil
}

// CheckAPIRateLimit checks if the tenant has exceeded their API rate limit
func (s *tenantUsageService) CheckAPIRateLimit(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	today := time.Now().Truncate(24 * time.Hour)

	usage, err := s.repos.TenantUsageTracking.FindByTenantAndDate(ctx, tenantID, today)
	if err != nil {
		// No usage record means they haven't made any calls today
		return true, nil
	}

	return usage.CanMakeAPICall(), nil
}

// GetAPIUsageStats gets API usage statistics for the specified number of days
func (s *tenantUsageService) GetAPIUsageStats(ctx context.Context, tenantID uuid.UUID, days int) (*dto.UsageSummary, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	endDate := time.Now().Truncate(24 * time.Hour)

	// Get total API calls
	totalCalls, err := s.repos.TenantUsageTracking.GetTotalAPICallsForPeriod(ctx, tenantID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get total API calls", zap.Error(err))
		return nil, fmt.Errorf("failed to get API stats: %w", err)
	}

	// Get average
	average, err := s.repos.TenantUsageTracking.GetAverageAPICallsPerDay(ctx, tenantID, days)
	if err != nil {
		s.logger.Error("failed to get average API calls", zap.Error(err))
		return nil, fmt.Errorf("failed to get API stats: %w", err)
	}

	// Get peak usage
	peakUsage, err := s.repos.TenantUsageTracking.GetPeakUsageDay(ctx, tenantID, days)
	if err != nil {
		s.logger.Error("failed to get peak usage", zap.Error(err))
		// Continue without peak data
	}

	summary := &dto.UsageSummary{
		TotalAPICalls:   totalCalls,
		AverageAPICalls: average,
	}

	if peakUsage != nil {
		summary.PeakAPICalls = peakUsage.APICallsCount
		summary.PeakAPICallsDate = peakUsage.Date.Format("2006-01-02")
	}

	return summary, nil
}

// ============================================================================
// Feature Usage Tracking
// ============================================================================

// TrackFeatureUsage tracks usage of specific features
func (s *tenantUsageService) TrackFeatureUsage(ctx context.Context, req *dto.TrackFeatureUsageRequest, tenantID uuid.UUID) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	today := time.Now().Truncate(24 * time.Hour)

	usage, err := s.getOrCreateUsageRecord(ctx, tenantID, today)
	if err != nil {
		return err
	}

	// Update specific counters based on feature
	switch req.Feature {
	case "booking":
		usage.BookingsCreated += req.Count
	case "project":
		usage.ProjectsCreated += req.Count
	case "sms":
		usage.SMSSent += req.Count
	case "email":
		usage.EmailsSent += req.Count
	case "storage":
		usage.StorageUsedGB += int64(req.Count)
	case "bandwidth":
		usage.BandwidthUsedGB += int64(req.Count)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidFeature, req.Feature)
	}

	if err := s.repos.TenantUsageTracking.Update(ctx, usage); err != nil {
		s.logger.Error("failed to update usage", zap.Error(err))
		return fmt.Errorf("failed to update usage: %w", err)
	}

	return nil
}

// TrackBookingCreated tracks when a booking is created
func (s *tenantUsageService) TrackBookingCreated(ctx context.Context, tenantID uuid.UUID) error {
	return s.incrementFeatureCounter(ctx, tenantID, "booking", 1)
}

// TrackProjectCreated tracks when a project is created
func (s *tenantUsageService) TrackProjectCreated(ctx context.Context, tenantID uuid.UUID) error {
	return s.incrementFeatureCounter(ctx, tenantID, "project", 1)
}

// TrackSMSSent tracks SMS messages sent
func (s *tenantUsageService) TrackSMSSent(ctx context.Context, tenantID uuid.UUID, count int) error {
	if count <= 0 {
		count = 1
	}
	return s.incrementFeatureCounter(ctx, tenantID, "sms", count)
}

// TrackEmailSent tracks emails sent
func (s *tenantUsageService) TrackEmailSent(ctx context.Context, tenantID uuid.UUID, count int) error {
	if count <= 0 {
		count = 1
	}
	return s.incrementFeatureCounter(ctx, tenantID, "email", count)
}

// TrackStorageUsage updates storage usage (in bytes, converted to GB)
func (s *tenantUsageService) TrackStorageUsage(ctx context.Context, tenantID uuid.UUID, bytesUsed int64) error {
	today := time.Now().Truncate(24 * time.Hour)

	usage, err := s.getOrCreateUsageRecord(ctx, tenantID, today)
	if err != nil {
		return err
	}

	// Convert bytes to GB
	gbUsed := bytesUsed / (1024 * 1024 * 1024)
	usage.StorageUsedGB = gbUsed

	if err := s.repos.TenantUsageTracking.Update(ctx, usage); err != nil {
		s.logger.Error("failed to update storage usage", zap.Error(err))
		return fmt.Errorf("failed to update storage usage: %w", err)
	}

	return nil
}

// TrackBandwidthUsage tracks bandwidth usage (in bytes, converted to GB)
func (s *tenantUsageService) TrackBandwidthUsage(ctx context.Context, tenantID uuid.UUID, bytesUsed int64) error {
	today := time.Now().Truncate(24 * time.Hour)

	usage, err := s.getOrCreateUsageRecord(ctx, tenantID, today)
	if err != nil {
		return err
	}

	// Convert bytes to GB and add to existing
	gbUsed := bytesUsed / (1024 * 1024 * 1024)
	usage.BandwidthUsedGB += gbUsed

	if err := s.repos.TenantUsageTracking.Update(ctx, usage); err != nil {
		s.logger.Error("failed to update bandwidth usage", zap.Error(err))
		return fmt.Errorf("failed to update bandwidth usage: %w", err)
	}

	return nil
}

// ============================================================================
// Usage Analytics
// ============================================================================

// GetPeakUsage returns the day with highest API usage
func (s *tenantUsageService) GetPeakUsage(ctx context.Context, tenantID uuid.UUID, days int) (*dto.UsageTrackingResponse, error) {
	if days <= 0 {
		days = 30
	}

	peakUsage, err := s.repos.TenantUsageTracking.GetPeakUsageDay(ctx, tenantID, days)
	if err != nil {
		s.logger.Error("failed to get peak usage", zap.Error(err))
		return nil, fmt.Errorf("failed to get peak usage: %w", err)
	}

	if peakUsage == nil {
		return nil, nil
	}

	return dto.ToUsageTrackingResponse(peakUsage), nil
}

// GetAverageUsage calculates average usage metrics
func (s *tenantUsageService) GetAverageUsage(ctx context.Context, tenantID uuid.UUID, days int) (*dto.UsageSummary, error) {
	if days <= 0 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	endDate := time.Now().Truncate(24 * time.Hour)

	usages, err := s.repos.TenantUsageTracking.FindByTenantAndDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get usage records", zap.Error(err))
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	return s.calculateUsageSummary(usages), nil
}

// ============================================================================
// Cleanup Operations
// ============================================================================

// DeleteOldUsageRecords deletes usage records older than the specified number of days
func (s *tenantUsageService) DeleteOldUsageRecords(ctx context.Context, olderThanDays int) (int64, error) {
	if olderThanDays < 30 {
		olderThanDays = 30 // Minimum retention of 30 days
	}

	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	s.logger.Info("deleting old usage records",
		zap.Int("older_than_days", olderThanDays),
		zap.Time("cutoff_date", cutoffDate),
	)

	count, err := s.repos.TenantUsageTracking.DeleteOldRecords(ctx, cutoffDate)
	if err != nil {
		s.logger.Error("failed to delete old usage records", zap.Error(err))
		return 0, fmt.Errorf("failed to delete old records: %w", err)
	}

	s.logger.Info("old usage records deleted", zap.Int64("count", count))
	return count, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// getOrCreateUsageRecord gets existing usage record or creates a new one for today
func (s *tenantUsageService) getOrCreateUsageRecord(ctx context.Context, tenantID uuid.UUID, date time.Time) (*models.TenantUsageTracking, error) {
	usage, err := s.repos.TenantUsageTracking.FindByTenantAndDate(ctx, tenantID, date)
	if err == nil {
		return usage, nil
	}

	// Create new usage record
	apiLimit := s.getAPILimitForTenant(ctx, tenantID)
	usage = &models.TenantUsageTracking{
		TenantID:      tenantID,
		Date:          date,
		APICallsLimit: apiLimit,
	}

	if err := s.repos.TenantUsageTracking.Create(ctx, usage); err != nil {
		s.logger.Error("failed to create usage record", zap.Error(err))
		return nil, fmt.Errorf("failed to create usage record: %w", err)
	}

	return usage, nil
}

// incrementFeatureCounter increments a feature counter
func (s *tenantUsageService) incrementFeatureCounter(ctx context.Context, tenantID uuid.UUID, feature string, count int) error {
	today := time.Now().Truncate(24 * time.Hour)

	if err := s.repos.TenantUsageTracking.IncrementFeatureUsage(ctx, tenantID, today, feature, count); err != nil {
		s.logger.Error("failed to increment feature counter",
			zap.String("feature", feature),
			zap.Error(err),
		)
		return fmt.Errorf("failed to track %s usage: %w", feature, err)
	}

	return nil
}

// getAPILimitForTenant returns the API rate limit based on tenant plan
func (s *tenantUsageService) getAPILimitForTenant(ctx context.Context, tenantID uuid.UUID) int64 {
	tenant, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		return 1000 // Default limit
	}

	// Define limits per plan
	limits := map[models.TenantPlan]int64{
		models.TenantPlanSolo:        1000,
		models.TenantPlanSmall:       5000,
		models.TenantPlanCorporation: 25000,
		models.TenantPlanEnterprise:  100000,
	}

	if limit, ok := limits[tenant.Plan]; ok {
		return limit
	}

	return 1000 // Default
}

// calculateUsageSummary calculates aggregate statistics from usage records
func (s *tenantUsageService) calculateUsageSummary(usages []*models.TenantUsageTracking) *dto.UsageSummary {
	summary := &dto.UsageSummary{}

	if len(usages) == 0 {
		return summary
	}

	var totalAPICalls int64
	var totalBookings, totalProjects, totalSMS, totalEmails int
	var totalStorage, totalBandwidth int64
	var peakAPICalls int64
	var peakDate string

	for _, usage := range usages {
		totalAPICalls += usage.APICallsCount
		totalBookings += usage.BookingsCreated
		totalProjects += usage.ProjectsCreated
		totalSMS += usage.SMSSent
		totalEmails += usage.EmailsSent
		totalStorage += usage.StorageUsedGB
		totalBandwidth += usage.BandwidthUsedGB

		if usage.APICallsCount > peakAPICalls {
			peakAPICalls = usage.APICallsCount
			peakDate = usage.Date.Format("2006-01-02")
		}
	}

	summary.TotalAPICalls = totalAPICalls
	summary.TotalBookings = totalBookings
	summary.TotalProjects = totalProjects
	summary.TotalSMS = totalSMS
	summary.TotalEmails = totalEmails
	summary.TotalStorageUsed = totalStorage
	summary.TotalBandwidth = totalBandwidth
	summary.PeakAPICalls = peakAPICalls
	summary.PeakAPICallsDate = peakDate

	// Calculate average
	if len(usages) > 0 {
		summary.AverageAPICalls = float64(totalAPICalls) / float64(len(usages))
	}

	return summary
}
