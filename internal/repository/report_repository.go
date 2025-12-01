package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportRepository defines the interface for report repository operations
type ReportRepository interface {
	BaseRepository[models.Report]

	// Query Operations
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Report, PaginationResult, error)
	FindByType(ctx context.Context, tenantID uuid.UUID, reportType models.ReportType, pagination PaginationParams) ([]*models.Report, PaginationResult, error)
	FindByStatus(ctx context.Context, tenantID uuid.UUID, status models.ReportStatus, pagination PaginationParams) ([]*models.Report, PaginationResult, error)
	FindByRequestedBy(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.Report, PaginationResult, error)
	FindByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Report, PaginationResult, error)
	FindScheduledReports(ctx context.Context, tenantID uuid.UUID) ([]*models.Report, error)
	FindPendingReports(ctx context.Context, tenantID uuid.UUID) ([]*models.Report, error)
	FindFailedReports(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Report, PaginationResult, error)

	// Status Management
	UpdateStatus(ctx context.Context, reportID uuid.UUID, status models.ReportStatus) error
	MarkAsGenerating(ctx context.Context, reportID uuid.UUID) error
	MarkAsCompleted(ctx context.Context, reportID uuid.UUID, fileURL string) error
	MarkAsFailed(ctx context.Context, reportID uuid.UUID, errorMessage string) error

	// Scheduled Reports
	FindScheduledDue(ctx context.Context) ([]*models.Report, error)
	UpdateNextScheduledRun(ctx context.Context, reportID uuid.UUID, nextRun time.Time) error
	EnableSchedule(ctx context.Context, reportID uuid.UUID) error
	DisableSchedule(ctx context.Context, reportID uuid.UUID) error

	// Report Generation Queue
	GetNextPendingReport(ctx context.Context) (*models.Report, error)
	RetryFailedReport(ctx context.Context, reportID uuid.UUID) error

	// Analytics & Statistics
	GetReportStats(ctx context.Context, tenantID uuid.UUID) (ReportStats, error)
	GetReportTypeUsage(ctx context.Context, tenantID uuid.UUID, days int) ([]ReportTypeUsage, error)
	GetUserReportActivity(ctx context.Context, userID uuid.UUID) (UserReportActivity, error)
	GetReportGenerationMetrics(ctx context.Context, tenantID uuid.UUID, days int) (ReportGenerationMetrics, error)

	// Cleanup Operations
	DeleteOldReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error
	DeleteFailedReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error
	ArchiveCompletedReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error

	// Bulk Operations
	BulkUpdateStatus(ctx context.Context, reportIDs []uuid.UUID, status models.ReportStatus) error
	BulkDelete(ctx context.Context, reportIDs []uuid.UUID) error

	// Search & Filter
	Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.Report, PaginationResult, error)
	FindByFilters(ctx context.Context, tenantID uuid.UUID, filters ReportFilters, pagination PaginationParams) ([]*models.Report, PaginationResult, error)
}

// ReportStats represents report statistics
type ReportStats struct {
	TotalReports      int64                         `json:"total_reports"`
	CompletedReports  int64                         `json:"completed_reports"`
	PendingReports    int64                         `json:"pending_reports"`
	FailedReports     int64                         `json:"failed_reports"`
	ScheduledReports  int64                         `json:"scheduled_reports"`
	ByType            map[models.ReportType]int64   `json:"by_type"`
	ByStatus          map[models.ReportStatus]int64 `json:"by_status"`
	ByFormat          map[string]int64              `json:"by_format"`
	ReportsThisWeek   int64                         `json:"reports_this_week"`
	ReportsThisMonth  int64                         `json:"reports_this_month"`
	AvgGenerationTime float64                       `json:"avg_generation_time_seconds"`
	SuccessRate       float64                       `json:"success_rate"`
	MostRequestedType models.ReportType             `json:"most_requested_type"`
}

// ReportTypeUsage represents report type usage over time
type ReportTypeUsage struct {
	ReportType models.ReportType `json:"report_type"`
	Count      int64             `json:"count"`
	Date       time.Time         `json:"date"`
}

// UserReportActivity represents user report activity
type UserReportActivity struct {
	UserID           uuid.UUID         `json:"user_id"`
	TotalReports     int64             `json:"total_reports"`
	CompletedReports int64             `json:"completed_reports"`
	FailedReports    int64             `json:"failed_reports"`
	ScheduledReports int64             `json:"scheduled_reports"`
	LastReportAt     *time.Time        `json:"last_report_at,omitempty"`
	MostUsedType     models.ReportType `json:"most_used_type"`
	PreferredFormat  string            `json:"preferred_format"`
}

// ReportGenerationMetrics represents report generation metrics
type ReportGenerationMetrics struct {
	TotalGenerated     int64   `json:"total_generated"`
	SuccessfulReports  int64   `json:"successful_reports"`
	FailedReports      int64   `json:"failed_reports"`
	AvgGenerationTime  float64 `json:"avg_generation_time_seconds"`
	MinGenerationTime  float64 `json:"min_generation_time_seconds"`
	MaxGenerationTime  float64 `json:"max_generation_time_seconds"`
	SuccessRate        float64 `json:"success_rate"`
	FailureRate        float64 `json:"failure_rate"`
	ReportsPerDay      float64 `json:"reports_per_day"`
	PeakGenerationHour int     `json:"peak_generation_hour"`
}

// ReportFilters for advanced filtering
type ReportFilters struct {
	Types       []models.ReportType   `json:"types"`
	Statuses    []models.ReportStatus `json:"statuses"`
	Formats     []string              `json:"formats"`
	StartDate   *time.Time            `json:"start_date"`
	EndDate     *time.Time            `json:"end_date"`
	IsScheduled *bool                 `json:"is_scheduled"`
	RequestedBy *uuid.UUID            `json:"requested_by"`
}

// reportRepository implements ReportRepository
type reportRepository struct {
	BaseRepository[models.Report]
	db     *gorm.DB
	logger log.AllLogger
	cache  Cache
}

// NewReportRepository creates a new ReportRepository instance
func NewReportRepository(db *gorm.DB, config ...RepositoryConfig) ReportRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.Report](db, cfg)

	return &reportRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
	}
}

// FindByTenantID retrieves all reports for a tenant
func (r *reportRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Report, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count reports", err)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		r.logger.Error("failed to find reports", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find reports", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return reports, paginationResult, nil
}

// FindByType retrieves reports by type
func (r *reportRepository) FindByType(ctx context.Context, tenantID uuid.UUID, reportType models.ReportType, pagination PaginationParams) ([]*models.Report, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND type = ?", tenantID, reportType).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count reports", err)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Where("tenant_id = ? AND type = ?", tenantID, reportType).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find reports by type", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return reports, paginationResult, nil
}

// FindByStatus retrieves reports by status
func (r *reportRepository) FindByStatus(ctx context.Context, tenantID uuid.UUID, status models.ReportStatus, pagination PaginationParams) ([]*models.Report, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count reports", err)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find reports by status", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return reports, paginationResult, nil
}

// FindByRequestedBy retrieves reports requested by a user
func (r *reportRepository) FindByRequestedBy(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.Report, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("requested_by_id = ?", userID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count reports", err)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Where("requested_by_id = ?", userID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		r.logger.Error("failed to find reports by user", "user_id", userID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find reports", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return reports, paginationResult, nil
}

// FindByDateRange retrieves reports within a date range
func (r *reportRepository) FindByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Report, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND start_date >= ? AND end_date <= ?", tenantID, startDate, endDate).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count reports", err)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Where("tenant_id = ? AND start_date >= ? AND end_date <= ?", tenantID, startDate, endDate).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find reports by date range", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return reports, paginationResult, nil
}

// FindScheduledReports retrieves all scheduled reports for a tenant
func (r *reportRepository) FindScheduledReports(ctx context.Context, tenantID uuid.UUID) ([]*models.Report, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Where("tenant_id = ? AND is_scheduled = ?", tenantID, true).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		r.logger.Error("failed to find scheduled reports", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find scheduled reports", err)
	}

	return reports, nil
}

// FindPendingReports retrieves all pending reports for a tenant
func (r *reportRepository) FindPendingReports(ctx context.Context, tenantID uuid.UUID) ([]*models.Report, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Where("tenant_id = ? AND status = ?", tenantID, models.ReportStatusPending).
		Order("created_at ASC").
		Find(&reports).Error; err != nil {
		r.logger.Error("failed to find pending reports", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find pending reports", err)
	}

	return reports, nil
}

// FindFailedReports retrieves all failed reports for a tenant
func (r *reportRepository) FindFailedReports(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Report, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.ReportStatusFailed).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count failed reports", err)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Where("tenant_id = ? AND status = ?", tenantID, models.ReportStatusFailed).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find failed reports", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return reports, paginationResult, nil
}

// UpdateStatus updates the status of a report
func (r *reportRepository) UpdateStatus(ctx context.Context, reportID uuid.UUID, status models.ReportStatus) error {
	if reportID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "report_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id = ?", reportID).
		Update("status", status)

	if result.Error != nil {
		r.logger.Error("failed to update report status", "report_id", reportID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update report status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "report not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	return nil
}

// MarkAsGenerating marks a report as being generated
func (r *reportRepository) MarkAsGenerating(ctx context.Context, reportID uuid.UUID) error {
	if reportID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "report_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id = ? AND status = ?", reportID, models.ReportStatusPending).
		Update("status", models.ReportStatusGenerating)

	if result.Error != nil {
		r.logger.Error("failed to mark report as generating", "report_id", reportID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark report as generating", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "report not found or not in pending status", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	return nil
}

// MarkAsCompleted marks a report as completed with file URL
func (r *reportRepository) MarkAsCompleted(ctx context.Context, reportID uuid.UUID, fileURL string) error {
	if reportID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "report_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.ReportStatusCompleted,
		"file_url":     fileURL,
		"generated_at": now,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id = ?", reportID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to mark report as completed", "report_id", reportID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark report as completed", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "report not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	r.logger.Info("report completed", "report_id", reportID, "file_url", fileURL)
	return nil
}

// MarkAsFailed marks a report as failed with error message
func (r *reportRepository) MarkAsFailed(ctx context.Context, reportID uuid.UUID, errorMessage string) error {
	if reportID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "report_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status":        models.ReportStatusFailed,
		"error_message": errorMessage,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id = ?", reportID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to mark report as failed", "report_id", reportID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark report as failed", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "report not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	r.logger.Warn("report failed", "report_id", reportID, "error", errorMessage)
	return nil
}

// FindScheduledDue retrieves scheduled reports that are due for generation
func (r *reportRepository) FindScheduledDue(ctx context.Context) ([]*models.Report, error) {
	// This would typically check against a next_run_at field
	// For simplicity, we'll return all scheduled reports
	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Preload("Tenant").
		Where("is_scheduled = ?", true).
		Find(&reports).Error; err != nil {
		r.logger.Error("failed to find scheduled reports due", "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find scheduled reports", err)
	}

	return reports, nil
}

// UpdateNextScheduledRun updates the next scheduled run time
func (r *reportRepository) UpdateNextScheduledRun(ctx context.Context, reportID uuid.UUID, nextRun time.Time) error {
	if reportID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "report_id cannot be nil", errors.ErrInvalidInput)
	}

	// Store next run in metadata
	var report models.Report
	if err := r.db.WithContext(ctx).First(&report, reportID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "report not found", errors.ErrNotFound)
	}

	if report.Metadata == nil {
		report.Metadata = make(models.JSONB)
	}
	report.Metadata["next_run_at"] = nextRun

	if err := r.db.WithContext(ctx).
		Model(&report).
		Update("metadata", report.Metadata).Error; err != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update next scheduled run", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	return nil
}

// EnableSchedule enables scheduling for a report
func (r *reportRepository) EnableSchedule(ctx context.Context, reportID uuid.UUID) error {
	if reportID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "report_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id = ?", reportID).
		Update("is_scheduled", true)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to enable schedule", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "report not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	return nil
}

// DisableSchedule disables scheduling for a report
func (r *reportRepository) DisableSchedule(ctx context.Context, reportID uuid.UUID) error {
	if reportID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "report_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id = ?", reportID).
		Update("is_scheduled", false)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to disable schedule", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "report not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	return nil
}

// GetNextPendingReport retrieves the next pending report to be processed
func (r *reportRepository) GetNextPendingReport(ctx context.Context) (*models.Report, error) {
	var report models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Preload("Tenant").
		Where("status = ?", models.ReportStatusPending).
		Order("created_at ASC").
		First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No pending reports
		}
		r.logger.Error("failed to get next pending report", "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to get next pending report", err)
	}

	return &report, nil
}

// RetryFailedReport resets a failed report to pending status
func (r *reportRepository) RetryFailedReport(ctx context.Context, reportID uuid.UUID) error {
	if reportID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "report_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status":        models.ReportStatusPending,
		"error_message": "",
		"file_url":      "",
	}

	result := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id = ? AND status = ?", reportID, models.ReportStatusFailed).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to retry report", "report_id", reportID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to retry report", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "report not found or not in failed status", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	r.logger.Info("report queued for retry", "report_id", reportID)
	return nil
}

// GetReportStats retrieves comprehensive report statistics for a tenant
func (r *reportRepository) GetReportStats(ctx context.Context, tenantID uuid.UUID) (ReportStats, error) {
	if tenantID == uuid.Nil {
		return ReportStats{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := ReportStats{
		ByType:   make(map[models.ReportType]int64),
		ByStatus: make(map[models.ReportStatus]int64),
		ByFormat: make(map[string]int64),
	}

	// Total reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalReports)

	// Completed reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.ReportStatusCompleted).
		Count(&stats.CompletedReports)

	// Pending reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.ReportStatusPending).
		Count(&stats.PendingReports)

	// Failed reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.ReportStatusFailed).
		Count(&stats.FailedReports)

	// Scheduled reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND is_scheduled = ?", tenantID, true).
		Count(&stats.ScheduledReports)

	// Reports by type
	type TypeCount struct {
		Type  models.ReportType
		Count int64
	}
	var typeCounts []TypeCount
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Select("type, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("type").
		Scan(&typeCounts)

	for _, tc := range typeCounts {
		stats.ByType[tc.Type] = tc.Count
	}

	// Most requested type
	if len(typeCounts) > 0 {
		maxCount := int64(0)
		for _, tc := range typeCounts {
			if tc.Count > maxCount {
				maxCount = tc.Count
				stats.MostRequestedType = tc.Type
			}
		}
	}

	// Reports by status
	type StatusCount struct {
		Status models.ReportStatus
		Count  int64
	}
	var statusCounts []StatusCount
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// Reports by format
	type FormatCount struct {
		Format string
		Count  int64
	}
	var formatCounts []FormatCount
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Select("file_format, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("file_format").
		Scan(&formatCounts)

	for _, fc := range formatCounts {
		stats.ByFormat[fc.Format] = fc.Count
	}

	// Reports this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, weekAgo).
		Count(&stats.ReportsThisWeek)

	// Reports this month
	monthAgo := time.Now().AddDate(0, -1, 0)
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, monthAgo).
		Count(&stats.ReportsThisMonth)

	// Average generation time (in seconds)
	r.db.WithContext(ctx).Raw(`
		SELECT AVG(EXTRACT(EPOCH FROM (generated_at - created_at)))
		FROM reports
		WHERE tenant_id = ?
			AND status = ?
			AND generated_at IS NOT NULL
	`, tenantID, models.ReportStatusCompleted).Scan(&stats.AvgGenerationTime)

	// Success rate
	if stats.TotalReports > 0 {
		stats.SuccessRate = (float64(stats.CompletedReports) / float64(stats.TotalReports)) * 100
	}

	return stats, nil
}

// GetReportTypeUsage retrieves report type usage over time
func (r *reportRepository) GetReportTypeUsage(ctx context.Context, tenantID uuid.UUID, days int) ([]ReportTypeUsage, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if days <= 0 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)

	var usage []ReportTypeUsage
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			type as report_type,
			DATE(created_at) as date,
			COUNT(*) as count
		FROM reports
		WHERE tenant_id = ?
			AND created_at >= ?
			AND deleted_at IS NULL
		GROUP BY type, DATE(created_at)
		ORDER BY date DESC, count DESC
	`, tenantID, startDate).Scan(&usage).Error; err != nil {
		r.logger.Error("failed to get report type usage", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get report type usage", err)
	}

	return usage, nil
}

// GetUserReportActivity retrieves report activity for a user
func (r *reportRepository) GetUserReportActivity(ctx context.Context, userID uuid.UUID) (UserReportActivity, error) {
	if userID == uuid.Nil {
		return UserReportActivity{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	activity := UserReportActivity{
		UserID: userID,
	}

	// Total reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("requested_by_id = ?", userID).
		Count(&activity.TotalReports)

	// Completed reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("requested_by_id = ? AND status = ?", userID, models.ReportStatusCompleted).
		Count(&activity.CompletedReports)

	// Failed reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("requested_by_id = ? AND status = ?", userID, models.ReportStatusFailed).
		Count(&activity.FailedReports)

	// Scheduled reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("requested_by_id = ? AND is_scheduled = ?", userID, true).
		Count(&activity.ScheduledReports)

	// Last report
	var lastReport models.Report
	if err := r.db.WithContext(ctx).
		Where("requested_by_id = ?", userID).
		Order("created_at DESC").
		First(&lastReport).Error; err == nil {
		activity.LastReportAt = &lastReport.CreatedAt
	}

	// Most used type
	type TypeCount struct {
		Type  models.ReportType
		Count int64
	}
	var typeCount TypeCount
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Select("type, COUNT(*) as count").
		Where("requested_by_id = ?", userID).
		Group("type").
		Order("count DESC").
		First(&typeCount).Error; err == nil {
		activity.MostUsedType = typeCount.Type
	}

	// Preferred format
	type FormatCount struct {
		Format string
		Count  int64
	}
	var formatCount FormatCount
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Select("file_format, COUNT(*) as count").
		Where("requested_by_id = ?", userID).
		Group("file_format").
		Order("count DESC").
		First(&formatCount).Error; err == nil {
		activity.PreferredFormat = formatCount.Format
	}

	return activity, nil
}

// GetReportGenerationMetrics retrieves report generation metrics
func (r *reportRepository) GetReportGenerationMetrics(ctx context.Context, tenantID uuid.UUID, days int) (ReportGenerationMetrics, error) {
	if tenantID == uuid.Nil {
		return ReportGenerationMetrics{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if days <= 0 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)
	metrics := ReportGenerationMetrics{}

	// Total generated (completed or failed)
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND created_at >= ? AND status IN (?)",
			tenantID, startDate, []models.ReportStatus{models.ReportStatusCompleted, models.ReportStatusFailed}).
		Count(&metrics.TotalGenerated)

	// Successful reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND created_at >= ? AND status = ?",
			tenantID, startDate, models.ReportStatusCompleted).
		Count(&metrics.SuccessfulReports)

	// Failed reports
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND created_at >= ? AND status = ?",
			tenantID, startDate, models.ReportStatusFailed).
		Count(&metrics.FailedReports)

	// Average generation time
	r.db.WithContext(ctx).Raw(`
		SELECT AVG(EXTRACT(EPOCH FROM (generated_at - created_at)))
		FROM reports
		WHERE tenant_id = ?
			AND created_at >= ?
			AND status = ?
			AND generated_at IS NOT NULL
	`, tenantID, startDate, models.ReportStatusCompleted).Scan(&metrics.AvgGenerationTime)

	// Min generation time
	r.db.WithContext(ctx).Raw(`
		SELECT MIN(EXTRACT(EPOCH FROM (generated_at - created_at)))
		FROM reports
		WHERE tenant_id = ?
			AND created_at >= ?
			AND status = ?
			AND generated_at IS NOT NULL
	`, tenantID, startDate, models.ReportStatusCompleted).Scan(&metrics.MinGenerationTime)

	// Max generation time
	r.db.WithContext(ctx).Raw(`
		SELECT MAX(EXTRACT(EPOCH FROM (generated_at - created_at)))
		FROM reports
		WHERE tenant_id = ?
			AND created_at >= ?
			AND status = ?
			AND generated_at IS NOT NULL
	`, tenantID, startDate, models.ReportStatusCompleted).Scan(&metrics.MaxGenerationTime)

	// Success rate
	if metrics.TotalGenerated > 0 {
		metrics.SuccessRate = (float64(metrics.SuccessfulReports) / float64(metrics.TotalGenerated)) * 100
		metrics.FailureRate = (float64(metrics.FailedReports) / float64(metrics.TotalGenerated)) * 100
	}

	// Reports per day
	if days > 0 {
		metrics.ReportsPerDay = float64(metrics.TotalGenerated) / float64(days)
	}

	// Peak generation hour
	type HourCount struct {
		Hour  int
		Count int64
	}
	var hourCount HourCount
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			EXTRACT(HOUR FROM created_at)::int as hour,
			COUNT(*) as count
		FROM reports
		WHERE tenant_id = ?
			AND created_at >= ?
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY count DESC
		LIMIT 1
	`, tenantID, startDate).Scan(&hourCount).Error; err == nil {
		metrics.PeakGenerationHour = hourCount.Hour
	}

	return metrics, nil
}

// DeleteOldReports deletes reports older than specified duration
func (r *reportRepository) DeleteOldReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	cutoffDate := time.Now().Add(-olderThan)

	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at < ?", tenantID, cutoffDate).
		Delete(&models.Report{})

	if result.Error != nil {
		r.logger.Error("failed to delete old reports", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete old reports", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	r.logger.Info("deleted old reports", "tenant_id", tenantID, "count", result.RowsAffected, "older_than", olderThan)
	return nil
}

// DeleteFailedReports deletes failed reports older than specified duration
func (r *reportRepository) DeleteFailedReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	cutoffDate := time.Now().Add(-olderThan)

	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ? AND created_at < ?",
			tenantID, models.ReportStatusFailed, cutoffDate).
		Delete(&models.Report{})

	if result.Error != nil {
		r.logger.Error("failed to delete failed reports", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete failed reports", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	r.logger.Info("deleted failed reports", "tenant_id", tenantID, "count", result.RowsAffected)
	return nil
}

// ArchiveCompletedReports archives (soft deletes) completed reports older than specified duration
func (r *reportRepository) ArchiveCompletedReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	cutoffDate := time.Now().Add(-olderThan)

	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ? AND created_at < ?",
			tenantID, models.ReportStatusCompleted, cutoffDate).
		Delete(&models.Report{})

	if result.Error != nil {
		r.logger.Error("failed to archive completed reports", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to archive reports", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	r.logger.Info("archived completed reports", "tenant_id", tenantID, "count", result.RowsAffected, "older_than", olderThan)
	return nil
}

// BulkUpdateStatus updates status for multiple reports
func (r *reportRepository) BulkUpdateStatus(ctx context.Context, reportIDs []uuid.UUID, status models.ReportStatus) error {
	if len(reportIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id IN ?", reportIDs).
		Update("status", status)

	if result.Error != nil {
		r.logger.Error("failed to bulk update report status", "count", len(reportIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update status", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	r.logger.Info("bulk updated report status", "count", result.RowsAffected, "status", status)
	return nil
}

// BulkDelete deletes multiple reports
func (r *reportRepository) BulkDelete(ctx context.Context, reportIDs []uuid.UUID) error {
	if len(reportIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Where("id IN ?", reportIDs).
		Delete(&models.Report{})

	if result.Error != nil {
		r.logger.Error("failed to bulk delete reports", "count", len(reportIDs), "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to bulk delete reports", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:reports:*")
	}

	r.logger.Info("bulk deleted reports", "count", result.RowsAffected)
	return nil
}

// Search searches reports by name or description
func (r *reportRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.Report, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("tenant_id = ? AND (name ILIKE ? OR description ILIKE ?)",
			tenantID, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count reports", err)
	}

	var reports []*models.Report
	if err := r.db.WithContext(ctx).
		Preload("RequestedBy").
		Where("tenant_id = ? AND (name ILIKE ? OR description ILIKE ?)",
			tenantID, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search reports", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return reports, paginationResult, nil
}

// FindByFilters retrieves reports using advanced filters
func (r *reportRepository) FindByFilters(ctx context.Context, tenantID uuid.UUID, filters ReportFilters, pagination PaginationParams) ([]*models.Report, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	// Apply filters
	if len(filters.Types) > 0 {
		query = query.Where("type IN ?", filters.Types)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if len(filters.Formats) > 0 {
		query = query.Where("file_format IN ?", filters.Formats)
	}

	if filters.StartDate != nil {
		query = query.Where("start_date >= ?", *filters.StartDate)
	}

	if filters.EndDate != nil {
		query = query.Where("end_date <= ?", *filters.EndDate)
	}

	if filters.IsScheduled != nil {
		query = query.Where("is_scheduled = ?", *filters.IsScheduled)
	}

	if filters.RequestedBy != nil {
		query = query.Where("requested_by_id = ?", *filters.RequestedBy)
	}

	// Count total
	var totalItems int64
	if err := query.Model(&models.Report{}).Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count reports", err)
	}

	// Find reports
	var reports []*models.Report
	if err := query.
		Preload("RequestedBy").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find reports", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return reports, paginationResult, nil
}
