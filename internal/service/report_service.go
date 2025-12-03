package service

import (
	"context"
	"fmt"
	"maps"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// ReportService defines the interface for report service operations
type ReportService interface {
	// CRUD Operations
	CreateReport(ctx context.Context, tenantID, requestedByID uuid.UUID, req *dto.CreateReportRequest) (*dto.ReportResponse, error)
	GetReport(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*dto.ReportResponse, error)
	UpdateReport(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *dto.UpdateReportRequest) (*dto.ReportResponse, error)
	DeleteReport(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Report Management
	ListReports(ctx context.Context, filter *dto.ReportFilter) (*dto.ReportListResponse, error)
	GetPendingReports(ctx context.Context, tenantID uuid.UUID) ([]*dto.ReportResponse, error)
	GetScheduledReports(ctx context.Context, tenantID uuid.UUID) ([]*dto.ReportResponse, error)
	GetFailedReports(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.ReportListResponse, error)

	// Status Management
	MarkAsGenerating(ctx context.Context, reportID uuid.UUID) error
	MarkAsCompleted(ctx context.Context, reportID uuid.UUID, fileURL string) error
	MarkAsFailed(ctx context.Context, reportID uuid.UUID, errorMessage string) error
	RetryFailedReport(ctx context.Context, reportID uuid.UUID, userID uuid.UUID) error

	// Scheduled Reports
	EnableSchedule(ctx context.Context, reportID uuid.UUID, userID uuid.UUID) error
	DisableSchedule(ctx context.Context, reportID uuid.UUID, userID uuid.UUID) error
	UpdateScheduleCron(ctx context.Context, reportID uuid.UUID, userID uuid.UUID, cronExpr string) error

	// Queue Management
	GetNextPendingReport(ctx context.Context) (*dto.ReportResponse, error)

	// Statistics & Analytics
	GetReportStats(ctx context.Context, tenantID uuid.UUID) (*dto.ReportStatsResponse, error)
	GetReportTypeUsage(ctx context.Context, tenantID uuid.UUID, days int) ([]repository.ReportTypeUsage, error)
	GetUserReportActivity(ctx context.Context, userID uuid.UUID) (*repository.UserReportActivity, error)
	GetReportGenerationMetrics(ctx context.Context, tenantID uuid.UUID, days int) (*repository.ReportGenerationMetrics, error)

	// Cleanup Operations
	DeleteOldReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error
	DeleteFailedReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error
	ArchiveCompletedReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error

	// Search & Filter
	SearchReports(ctx context.Context, tenantID uuid.UUID, query string, page, pageSize int) (*dto.ReportListResponse, error)
}

// reportService implements ReportService
type reportService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewReportService creates a new ReportService instance
func NewReportService(repos *repository.Repositories, logger log.AllLogger) ReportService {
	return &reportService{
		repos:  repos,
		logger: logger,
	}
}

// CreateReport creates a new report
func (s *reportService) CreateReport(ctx context.Context, tenantID, requestedByID uuid.UUID, req *dto.CreateReportRequest) (*dto.ReportResponse, error) {
	s.logger.Info("creating report", "tenant_id", tenantID, "type", req.Type)

	// Validate date range
	if req.EndDate.Before(req.StartDate) || req.EndDate.Equal(req.StartDate) {
		return nil, errors.NewValidationError("End date must be after start date")
	}

	// Verify user exists
	user, err := s.repos.User.GetByID(ctx, requestedByID)
	if err != nil {
		s.logger.Error("user not found", "user_id", requestedByID, "error", err)
		return nil, errors.NewNotFoundError("user")
	}

	// Verify user belongs to tenant
	if user.TenantID == nil || *user.TenantID != tenantID {
		s.logger.Warn("user does not belong to tenant", "user_id", requestedByID, "tenant_id", tenantID)
		return nil, errors.NewValidationError("User does not belong to this tenant")
	}

	// Default file format if not provided
	fileFormat := req.FileFormat
	if fileFormat == "" {
		fileFormat = "pdf"
	}

	// Validate file format
	validFormats := map[string]bool{"pdf": true, "csv": true, "xlsx": true}
	if !validFormats[fileFormat] {
		return nil, errors.NewValidationError("File format must be pdf, csv, or xlsx")
	}

	// Validate scheduled report
	if req.IsScheduled && req.ScheduleCron == "" {
		return nil, errors.NewValidationError("Schedule cron expression is required for scheduled reports")
	}

	// Create report model
	report := &models.Report{
		TenantID:      tenantID,
		Type:          req.Type,
		Name:          req.Name,
		Description:   req.Description,
		Status:        models.ReportStatusPending,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Filters:       req.Filters,
		FileFormat:    fileFormat,
		RequestedByID: requestedByID,
		IsScheduled:   req.IsScheduled,
		ScheduleCron:  req.ScheduleCron,
		Metadata:      req.Metadata,
	}

	// Save to database
	if err := s.repos.Report.Create(ctx, report); err != nil {
		s.logger.Error("failed to create report", "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "Failed to create report", err)
	}

	s.logger.Info("report created", "report_id", report.ID)

	// Load relationships
	created, err := s.repos.Report.GetByID(ctx, report.ID)
	if err != nil {
		s.logger.Error("failed to load report with relationships", "report_id", report.ID, "error", err)
		// Return basic response without relationships
		return dto.ToReportResponse(report), nil
	}

	return dto.ToReportResponse(created), nil
}

// GetReport retrieves a report by ID
func (s *reportService) GetReport(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*dto.ReportResponse, error) {
	s.logger.Info("retrieving report", "report_id", id, "user_id", userID)

	report, err := s.repos.Report.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("report not found", "report_id", id, "error", err)
		return nil, errors.NewNotFoundError("report")
	}

	// Verify user has access (must be requester or admin in same tenant)
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Check access: user must be requester or belong to same tenant
	if report.RequestedByID != userID && (user.TenantID == nil || report.TenantID != *user.TenantID) {
		s.logger.Warn("unauthorized access attempt", "report_id", id, "user_id", userID)
		return nil, errors.NewValidationError("You do not have access to this report")
	}

	return dto.ToReportResponse(report), nil
}

// UpdateReport updates a report
func (s *reportService) UpdateReport(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *dto.UpdateReportRequest) (*dto.ReportResponse, error) {
	s.logger.Info("updating report", "report_id", id, "user_id", userID)

	// Get existing report
	report, err := s.repos.Report.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("report not found", "report_id", id, "error", err)
		return nil, errors.NewNotFoundError("report")
	}

	// Verify user is the requester
	if report.RequestedByID != userID {
		s.logger.Warn("unauthorized update attempt", "report_id", id, "user_id", userID)
		return nil, errors.NewValidationError("Only the requester can update this report")
	}

	// Cannot update completed or generating reports
	if report.Status == models.ReportStatusCompleted || report.Status == models.ReportStatusGenerating {
		return nil, errors.NewValidationError(fmt.Sprintf("Cannot update report in %s status", report.Status))
	}

	// Update fields
	if req.Name != nil {
		report.Name = *req.Name
	}

	if req.Description != nil {
		report.Description = *req.Description
	}

	if req.ScheduleCron != nil {
		report.ScheduleCron = *req.ScheduleCron
	}

	if req.Metadata != nil {
		if report.Metadata == nil {
			report.Metadata = make(models.JSONB)
		}
		maps.Copy(report.Metadata, req.Metadata)
	}

	// Save changes
	if err := s.repos.Report.Update(ctx, report); err != nil {
		s.logger.Error("failed to update report", "report_id", id, "error", err)
		return nil, errors.NewServiceError("UPDATE_FAILED", "Failed to update report", err)
	}

	s.logger.Info("report updated", "report_id", id)
	return dto.ToReportResponse(report), nil
}

// DeleteReport deletes a report
func (s *reportService) DeleteReport(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("deleting report", "report_id", id, "user_id", userID)

	// Get existing report
	report, err := s.repos.Report.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("report not found", "report_id", id, "error", err)
		return errors.NewNotFoundError("report")
	}

	// Verify user is the requester
	if report.RequestedByID != userID {
		s.logger.Warn("unauthorized delete attempt", "report_id", id, "user_id", userID)
		return errors.NewValidationError("Only the requester can delete this report")
	}

	// Cannot delete generating reports
	if report.Status == models.ReportStatusGenerating {
		return errors.NewValidationError("Cannot delete report while it is being generated")
	}

	if err := s.repos.Report.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete report", "report_id", id, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete report", err)
	}

	s.logger.Info("report deleted", "report_id", id)
	return nil
}

// ListReports retrieves reports with filters
func (s *reportService) ListReports(ctx context.Context, filter *dto.ReportFilter) (*dto.ReportListResponse, error) {
	s.logger.Info("listing reports", "tenant_id", filter.TenantID)

	// Build repository filters
	repoFilters := repository.ReportFilters{}

	if filter.Type != nil {
		repoFilters.Types = []models.ReportType{*filter.Type}
	}

	if filter.Status != nil {
		repoFilters.Statuses = []models.ReportStatus{*filter.Status}
	}

	if filter.IsScheduled != nil {
		repoFilters.IsScheduled = filter.IsScheduled
	}

	if filter.RequestedByID != nil {
		repoFilters.RequestedBy = filter.RequestedByID
	}

	if filter.StartDate != nil {
		repoFilters.StartDate = filter.StartDate
	}

	if filter.EndDate != nil {
		repoFilters.EndDate = filter.EndDate
	}

	// Set defaults
	page := max(1, filter.Page)
	pageSize := min(100, max(1, filter.PageSize))

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	reports, paginationResult, err := s.repos.Report.FindByFilters(ctx, filter.TenantID, repoFilters, pagination)
	if err != nil {
		s.logger.Error("failed to list reports", "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list reports", err)
	}

	return &dto.ReportListResponse{
		Reports:     dto.ToReportResponses(reports),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetPendingReports retrieves all pending reports for a tenant
func (s *reportService) GetPendingReports(ctx context.Context, tenantID uuid.UUID) ([]*dto.ReportResponse, error) {
	s.logger.Info("getting pending reports", "tenant_id", tenantID)

	reports, err := s.repos.Report.FindPendingReports(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get pending reports", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get pending reports", err)
	}

	return dto.ToReportResponses(reports), nil
}

// GetScheduledReports retrieves all scheduled reports for a tenant
func (s *reportService) GetScheduledReports(ctx context.Context, tenantID uuid.UUID) ([]*dto.ReportResponse, error) {
	s.logger.Info("getting scheduled reports", "tenant_id", tenantID)

	reports, err := s.repos.Report.FindScheduledReports(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get scheduled reports", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get scheduled reports", err)
	}

	return dto.ToReportResponses(reports), nil
}

// GetFailedReports retrieves all failed reports for a tenant
func (s *reportService) GetFailedReports(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.ReportListResponse, error) {
	s.logger.Info("getting failed reports", "tenant_id", tenantID)

	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	reports, paginationResult, err := s.repos.Report.FindFailedReports(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to get failed reports", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get failed reports", err)
	}

	return &dto.ReportListResponse{
		Reports:     dto.ToReportResponses(reports),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// MarkAsGenerating marks a report as being generated
func (s *reportService) MarkAsGenerating(ctx context.Context, reportID uuid.UUID) error {
	s.logger.Info("marking report as generating", "report_id", reportID)

	if err := s.repos.Report.MarkAsGenerating(ctx, reportID); err != nil {
		s.logger.Error("failed to mark report as generating", "report_id", reportID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to mark report as generating", err)
	}

	return nil
}

// MarkAsCompleted marks a report as completed with file URL
func (s *reportService) MarkAsCompleted(ctx context.Context, reportID uuid.UUID, fileURL string) error {
	s.logger.Info("marking report as completed", "report_id", reportID)

	if fileURL == "" {
		return errors.NewValidationError("File URL is required")
	}

	if err := s.repos.Report.MarkAsCompleted(ctx, reportID, fileURL); err != nil {
		s.logger.Error("failed to mark report as completed", "report_id", reportID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to mark report as completed", err)
	}

	return nil
}

// MarkAsFailed marks a report as failed with error message
func (s *reportService) MarkAsFailed(ctx context.Context, reportID uuid.UUID, errorMessage string) error {
	s.logger.Info("marking report as failed", "report_id", reportID)

	if err := s.repos.Report.MarkAsFailed(ctx, reportID, errorMessage); err != nil {
		s.logger.Error("failed to mark report as failed", "report_id", reportID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to mark report as failed", err)
	}

	return nil
}

// RetryFailedReport retries a failed report
func (s *reportService) RetryFailedReport(ctx context.Context, reportID uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("retrying failed report", "report_id", reportID, "user_id", userID)

	// Get existing report
	report, err := s.repos.Report.GetByID(ctx, reportID)
	if err != nil {
		s.logger.Error("report not found", "report_id", reportID, "error", err)
		return errors.NewNotFoundError("report")
	}

	// Verify user is the requester
	if report.RequestedByID != userID {
		s.logger.Warn("unauthorized retry attempt", "report_id", reportID, "user_id", userID)
		return errors.NewValidationError("Only the requester can retry this report")
	}

	// Retry
	if err := s.repos.Report.RetryFailedReport(ctx, reportID); err != nil {
		s.logger.Error("failed to retry report", "report_id", reportID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to retry report", err)
	}

	s.logger.Info("report queued for retry", "report_id", reportID)
	return nil
}

// EnableSchedule enables scheduling for a report
func (s *reportService) EnableSchedule(ctx context.Context, reportID uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("enabling schedule", "report_id", reportID, "user_id", userID)

	// Get existing report
	report, err := s.repos.Report.GetByID(ctx, reportID)
	if err != nil {
		s.logger.Error("report not found", "report_id", reportID, "error", err)
		return errors.NewNotFoundError("report")
	}

	// Verify user is the requester
	if report.RequestedByID != userID {
		s.logger.Warn("unauthorized schedule update", "report_id", reportID, "user_id", userID)
		return errors.NewValidationError("Only the requester can update schedule")
	}

	// Verify cron expression exists
	if report.ScheduleCron == "" {
		return errors.NewValidationError("Schedule cron expression must be set first")
	}

	if err := s.repos.Report.EnableSchedule(ctx, reportID); err != nil {
		s.logger.Error("failed to enable schedule", "report_id", reportID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to enable schedule", err)
	}

	return nil
}

// DisableSchedule disables scheduling for a report
func (s *reportService) DisableSchedule(ctx context.Context, reportID uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("disabling schedule", "report_id", reportID, "user_id", userID)

	// Get existing report
	report, err := s.repos.Report.GetByID(ctx, reportID)
	if err != nil {
		s.logger.Error("report not found", "report_id", reportID, "error", err)
		return errors.NewNotFoundError("report")
	}

	// Verify user is the requester
	if report.RequestedByID != userID {
		s.logger.Warn("unauthorized schedule update", "report_id", reportID, "user_id", userID)
		return errors.NewValidationError("Only the requester can update schedule")
	}

	if err := s.repos.Report.DisableSchedule(ctx, reportID); err != nil {
		s.logger.Error("failed to disable schedule", "report_id", reportID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to disable schedule", err)
	}

	return nil
}

// UpdateScheduleCron updates the cron expression for a scheduled report
func (s *reportService) UpdateScheduleCron(ctx context.Context, reportID uuid.UUID, userID uuid.UUID, cronExpr string) error {
	s.logger.Info("updating schedule cron", "report_id", reportID, "user_id", userID)

	if cronExpr == "" {
		return errors.NewValidationError("Cron expression is required")
	}

	// Get existing report
	report, err := s.repos.Report.GetByID(ctx, reportID)
	if err != nil {
		s.logger.Error("report not found", "report_id", reportID, "error", err)
		return errors.NewNotFoundError("report")
	}

	// Verify user is the requester
	if report.RequestedByID != userID {
		s.logger.Warn("unauthorized schedule update", "report_id", reportID, "user_id", userID)
		return errors.NewValidationError("Only the requester can update schedule")
	}

	// Update cron expression
	report.ScheduleCron = cronExpr

	if err := s.repos.Report.Update(ctx, report); err != nil {
		s.logger.Error("failed to update schedule cron", "report_id", reportID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update schedule", err)
	}

	return nil
}

// GetNextPendingReport retrieves the next pending report to be processed
func (s *reportService) GetNextPendingReport(ctx context.Context) (*dto.ReportResponse, error) {
	s.logger.Info("getting next pending report")

	report, err := s.repos.Report.GetNextPendingReport(ctx)
	if err != nil {
		s.logger.Error("failed to get next pending report", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get next pending report", err)
	}

	if report == nil {
		return nil, nil // No pending reports
	}

	return dto.ToReportResponse(report), nil
}

// GetReportStats retrieves report statistics for a tenant
func (s *reportService) GetReportStats(ctx context.Context, tenantID uuid.UUID) (*dto.ReportStatsResponse, error) {
	s.logger.Info("getting report stats", "tenant_id", tenantID)

	stats, err := s.repos.Report.GetReportStats(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get report stats", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get report stats", err)
	}

	return &dto.ReportStatsResponse{
		TenantID:          tenantID,
		TotalReports:      stats.TotalReports,
		CompletedReports:  stats.CompletedReports,
		PendingReports:    stats.PendingReports,
		FailedReports:     stats.FailedReports,
		ScheduledReports:  stats.ScheduledReports,
		ByType:            stats.ByType,
		ByStatus:          stats.ByStatus,
		ByFormat:          stats.ByFormat,
		ReportsThisWeek:   stats.ReportsThisWeek,
		ReportsThisMonth:  stats.ReportsThisMonth,
		AvgGenerationTime: stats.AvgGenerationTime,
		SuccessRate:       stats.SuccessRate,
		MostRequestedType: stats.MostRequestedType,
	}, nil
}

// GetReportTypeUsage retrieves report type usage over time
func (s *reportService) GetReportTypeUsage(ctx context.Context, tenantID uuid.UUID, days int) ([]repository.ReportTypeUsage, error) {
	s.logger.Info("getting report type usage", "tenant_id", tenantID, "days", days)

	if days <= 0 {
		days = 30
	}

	usage, err := s.repos.Report.GetReportTypeUsage(ctx, tenantID, days)
	if err != nil {
		s.logger.Error("failed to get report type usage", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get report type usage", err)
	}

	return usage, nil
}

// GetUserReportActivity retrieves report activity for a user
func (s *reportService) GetUserReportActivity(ctx context.Context, userID uuid.UUID) (*repository.UserReportActivity, error) {
	s.logger.Info("getting user report activity", "user_id", userID)

	activity, err := s.repos.Report.GetUserReportActivity(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user report activity", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get user report activity", err)
	}

	return &activity, nil
}

// GetReportGenerationMetrics retrieves report generation metrics
func (s *reportService) GetReportGenerationMetrics(ctx context.Context, tenantID uuid.UUID, days int) (*repository.ReportGenerationMetrics, error) {
	s.logger.Info("getting report generation metrics", "tenant_id", tenantID, "days", days)

	if days <= 0 {
		days = 30
	}

	metrics, err := s.repos.Report.GetReportGenerationMetrics(ctx, tenantID, days)
	if err != nil {
		s.logger.Error("failed to get report generation metrics", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get report generation metrics", err)
	}

	return &metrics, nil
}

// DeleteOldReports deletes reports older than specified duration
func (s *reportService) DeleteOldReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error {
	s.logger.Info("deleting old reports", "tenant_id", tenantID, "older_than", olderThan)

	if err := s.repos.Report.DeleteOldReports(ctx, tenantID, olderThan); err != nil {
		s.logger.Error("failed to delete old reports", "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete old reports", err)
	}

	return nil
}

// DeleteFailedReports deletes failed reports older than specified duration
func (s *reportService) DeleteFailedReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error {
	s.logger.Info("deleting failed reports", "tenant_id", tenantID, "older_than", olderThan)

	if err := s.repos.Report.DeleteFailedReports(ctx, tenantID, olderThan); err != nil {
		s.logger.Error("failed to delete failed reports", "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete failed reports", err)
	}

	return nil
}

// ArchiveCompletedReports archives completed reports older than specified duration
func (s *reportService) ArchiveCompletedReports(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error {
	s.logger.Info("archiving completed reports", "tenant_id", tenantID, "older_than", olderThan)

	if err := s.repos.Report.ArchiveCompletedReports(ctx, tenantID, olderThan); err != nil {
		s.logger.Error("failed to archive completed reports", "error", err)
		return errors.NewServiceError("ARCHIVE_FAILED", "Failed to archive completed reports", err)
	}

	return nil
}

// SearchReports searches reports by name or description
func (s *reportService) SearchReports(ctx context.Context, tenantID uuid.UUID, query string, page, pageSize int) (*dto.ReportListResponse, error) {
	s.logger.Info("searching reports", "tenant_id", tenantID, "query", query)

	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	reports, paginationResult, err := s.repos.Report.Search(ctx, tenantID, query, pagination)
	if err != nil {
		s.logger.Error("failed to search reports", "error", err)
		return nil, errors.NewServiceError("SEARCH_FAILED", "Failed to search reports", err)
	}

	return &dto.ReportListResponse{
		Reports:     dto.ToReportResponses(reports),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}
