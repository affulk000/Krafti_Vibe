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

// Data export related errors
var (
	// ErrExportRequestNotFound is returned when export request is not found
	ErrExportRequestNotFound = errors.New("export request not found")

	// ErrExportAlreadyInProgress is returned when an export is already in progress
	ErrExportAlreadyInProgress = errors.New("data export already in progress")

	// ErrExportCannotBeCancelled is returned when export cannot be cancelled
	ErrExportCannotBeCancelled = errors.New("export cannot be cancelled in current state")

	// ErrExportExpired is returned when export download link has expired
	ErrExportExpired = errors.New("export download link has expired")
)

// Export status constants
const (
	ExportStatusPending    = "pending"
	ExportStatusProcessing = "processing"
	ExportStatusCompleted  = "completed"
	ExportStatusFailed     = "failed"
	ExportStatusCancelled  = "cancelled"
)

// Export type constants
const (
	ExportTypeFull    = "full"
	ExportTypePartial = "partial"
	ExportTypeGDPR    = "gdpr"
)

// DataExportService defines the interface for data export operations
type DataExportService interface {
	// Core Operations
	RequestDataExport(ctx context.Context, req *dto.DataExportRequest, tenantID uuid.UUID, requestedBy uuid.UUID) (*dto.DataExportResponse, error)
	GetDataExportStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.DataExportResponse, error)
	ListDataExports(ctx context.Context, tenantID uuid.UUID, filter *dto.DataExportFilter) (*dto.DataExportListResponse, error)
	CancelDataExport(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// Processing Operations (for background workers)
	StartProcessing(ctx context.Context, id uuid.UUID) error
	MarkCompleted(ctx context.Context, id uuid.UUID, fileURL string, fileSize int64) error
	MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) error

	// Query Operations
	GetPendingExports(ctx context.Context) ([]*dto.DataExportResponse, error)
	GetExportsByStatus(ctx context.Context, status string, pagination *dto.DataExportFilter) (*dto.DataExportListResponse, error)

	// Cleanup Operations
	DeleteExpiredExports(ctx context.Context) (int64, error)
	DeleteExportsByTenant(ctx context.Context, tenantID uuid.UUID) error
}

// dataExportService implements DataExportService
type dataExportService struct {
	repos           *repository.Repositories
	logger          *zap.Logger
	downloadExpiry  time.Duration // How long download links are valid
	maxExportPerDay int           // Maximum exports per tenant per day
}

// DataExportServiceConfig holds configuration for the data export service
type DataExportServiceConfig struct {
	DownloadExpiry  time.Duration
	MaxExportPerDay int
}

// NewDataExportService creates a new data export service
func NewDataExportService(
	repos *repository.Repositories,
	logger *zap.Logger,
	config ...DataExportServiceConfig,
) DataExportService {
	// Default configuration
	downloadExpiry := 24 * time.Hour
	maxExportPerDay := 5

	if len(config) > 0 {
		if config[0].DownloadExpiry > 0 {
			downloadExpiry = config[0].DownloadExpiry
		}
		if config[0].MaxExportPerDay > 0 {
			maxExportPerDay = config[0].MaxExportPerDay
		}
	}

	return &dataExportService{
		repos:           repos,
		logger:          logger,
		downloadExpiry:  downloadExpiry,
		maxExportPerDay: maxExportPerDay,
	}
}

// ============================================================================
// Core Operations
// ============================================================================

// RequestDataExport creates a new data export request
func (s *dataExportService) RequestDataExport(ctx context.Context, req *dto.DataExportRequest, tenantID uuid.UUID, requestedBy uuid.UUID) (*dto.DataExportResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	s.logger.Info("requesting data export",
		zap.String("tenant_id", tenantID.String()),
		zap.String("export_type", req.ExportType),
		zap.String("format", req.Format),
	)

	// Verify tenant exists and is active
	tenant, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	if tenant.Status == models.TenantStatusCancelled {
		return nil, ErrTenantCancelled
	}

	// Check for existing in-progress export
	pendingExports, err := s.repos.DataExport.FindPendingByTenant(ctx, tenantID)
	if err == nil && len(pendingExports) > 0 {
		return nil, ErrExportAlreadyInProgress
	}

	// Create export request
	exportRequest := &models.DataExportRequest{
		TenantID:    tenantID,
		RequestedBy: requestedBy,
		ExportType:  req.ExportType,
		Status:      ExportStatusPending,
	}

	if err := s.repos.DataExport.Create(ctx, exportRequest); err != nil {
		s.logger.Error("failed to create export request", zap.Error(err))
		return nil, fmt.Errorf("failed to create export request: %w", err)
	}

	s.logger.Info("data export requested successfully",
		zap.String("export_id", exportRequest.ID.String()),
		zap.String("tenant_id", tenantID.String()),
	)

	// TODO: Queue export job asynchronously via job queue/message broker
	// Example: s.jobQueue.Enqueue("data_export", exportRequest.ID)

	return dto.ToDataExportResponse(exportRequest), nil
}

// GetDataExportStatus retrieves the status of a data export request
func (s *dataExportService) GetDataExportStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.DataExportResponse, error) {
	exportRequest, err := s.repos.DataExport.GetByID(ctx, id)
	if err != nil {
		return nil, ErrExportRequestNotFound
	}

	// Verify tenant ownership
	if exportRequest.TenantID != tenantID {
		return nil, ErrExportRequestNotFound // Don't expose existence to other tenants
	}

	response := dto.ToDataExportResponse(exportRequest)

	// Check if download link has expired
	if exportRequest.Status == ExportStatusCompleted && exportRequest.ExpiresAt != nil {
		if time.Now().After(*exportRequest.ExpiresAt) {
			response.Status = "expired"
			response.FileURL = "" // Don't return expired URL
		}
	}

	return response, nil
}

// ListDataExports lists data export requests for a tenant
func (s *dataExportService) ListDataExports(ctx context.Context, tenantID uuid.UUID, filter *dto.DataExportFilter) (*dto.DataExportListResponse, error) {
	if filter == nil {
		filter = &dto.DataExportFilter{
			Page:     1,
			PageSize: 20,
		}
	}

	// Validate pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	exports, result, err := s.repos.DataExport.FindByTenant(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to list data exports", zap.Error(err))
		return nil, fmt.Errorf("failed to list data exports: %w", err)
	}

	responses := make([]*dto.DataExportResponse, 0, len(exports))
	for _, export := range exports {
		response := dto.ToDataExportResponse(export)

		// Apply status filter if provided
		if filter.Status != nil && *filter.Status != "" {
			if response.Status != *filter.Status {
				continue
			}
		}

		// Mark expired downloads
		if export.Status == ExportStatusCompleted && export.ExpiresAt != nil {
			if time.Now().After(*export.ExpiresAt) {
				response.Status = "expired"
				response.FileURL = ""
			}
		}

		responses = append(responses, response)
	}

	return &dto.DataExportListResponse{
		Exports:    responses,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: result.TotalPages,
	}, nil
}

// CancelDataExport cancels a pending data export request
func (s *dataExportService) CancelDataExport(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	s.logger.Info("cancelling data export",
		zap.String("export_id", id.String()),
		zap.String("tenant_id", tenantID.String()),
	)

	exportRequest, err := s.repos.DataExport.GetByID(ctx, id)
	if err != nil {
		return ErrExportRequestNotFound
	}

	// Verify tenant ownership
	if exportRequest.TenantID != tenantID {
		return ErrExportRequestNotFound
	}

	// Can only cancel pending exports
	if exportRequest.Status != ExportStatusPending {
		return fmt.Errorf("%w: current status is %s", ErrExportCannotBeCancelled, exportRequest.Status)
	}

	if err := s.repos.DataExport.UpdateStatus(ctx, id, ExportStatusCancelled); err != nil {
		s.logger.Error("failed to cancel export", zap.Error(err))
		return fmt.Errorf("failed to cancel export: %w", err)
	}

	s.logger.Info("data export cancelled successfully", zap.String("export_id", id.String()))
	return nil
}

// ============================================================================
// Processing Operations (for background workers)
// ============================================================================

// StartProcessing marks an export as being processed
func (s *dataExportService) StartProcessing(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("starting export processing", zap.String("export_id", id.String()))

	exportRequest, err := s.repos.DataExport.GetByID(ctx, id)
	if err != nil {
		return ErrExportRequestNotFound
	}

	if exportRequest.Status != ExportStatusPending {
		return fmt.Errorf("export is not in pending state: %s", exportRequest.Status)
	}

	if err := s.repos.DataExport.UpdateStatus(ctx, id, ExportStatusProcessing); err != nil {
		s.logger.Error("failed to update export status", zap.Error(err))
		return fmt.Errorf("failed to start processing: %w", err)
	}

	return nil
}

// MarkCompleted marks an export as completed with file details
func (s *dataExportService) MarkCompleted(ctx context.Context, id uuid.UUID, fileURL string, fileSize int64) error {
	s.logger.Info("marking export as completed",
		zap.String("export_id", id.String()),
		zap.Int64("file_size", fileSize),
	)

	exportRequest, err := s.repos.DataExport.GetByID(ctx, id)
	if err != nil {
		return ErrExportRequestNotFound
	}

	if exportRequest.Status != ExportStatusProcessing {
		return fmt.Errorf("export is not in processing state: %s", exportRequest.Status)
	}

	expiresAt := time.Now().Add(s.downloadExpiry)

	if err := s.repos.DataExport.MarkCompleted(ctx, id, fileURL, expiresAt); err != nil {
		s.logger.Error("failed to mark export completed", zap.Error(err))
		return fmt.Errorf("failed to mark completed: %w", err)
	}

	s.logger.Info("export completed successfully",
		zap.String("export_id", id.String()),
		zap.Time("expires_at", expiresAt),
	)

	// TODO: Send notification to user that export is ready

	return nil
}

// MarkFailed marks an export as failed with error details
func (s *dataExportService) MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) error {
	s.logger.Error("marking export as failed",
		zap.String("export_id", id.String()),
		zap.String("error", errorMessage),
	)

	if err := s.repos.DataExport.SetError(ctx, id, errorMessage); err != nil {
		s.logger.Error("failed to mark export failed", zap.Error(err))
		return fmt.Errorf("failed to mark failed: %w", err)
	}

	// TODO: Send notification to user about failure

	return nil
}

// ============================================================================
// Query Operations
// ============================================================================

// GetPendingExports retrieves all pending exports for processing
func (s *dataExportService) GetPendingExports(ctx context.Context) ([]*dto.DataExportResponse, error) {
	filter := &dto.DataExportFilter{
		Page:     1,
		PageSize: 100,
	}
	status := ExportStatusPending
	filter.Status = &status

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	exports, _, err := s.repos.DataExport.FindByStatus(ctx, ExportStatusPending, pagination)
	if err != nil {
		s.logger.Error("failed to get pending exports", zap.Error(err))
		return nil, fmt.Errorf("failed to get pending exports: %w", err)
	}

	responses := make([]*dto.DataExportResponse, 0, len(exports))
	for _, export := range exports {
		responses = append(responses, dto.ToDataExportResponse(export))
	}

	return responses, nil
}

// GetExportsByStatus retrieves exports filtered by status (admin use)
func (s *dataExportService) GetExportsByStatus(ctx context.Context, status string, filter *dto.DataExportFilter) (*dto.DataExportListResponse, error) {
	if filter == nil {
		filter = &dto.DataExportFilter{
			Page:     1,
			PageSize: 20,
		}
	}

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	exports, result, err := s.repos.DataExport.FindByStatus(ctx, status, pagination)
	if err != nil {
		s.logger.Error("failed to get exports by status", zap.Error(err))
		return nil, fmt.Errorf("failed to get exports: %w", err)
	}

	responses := make([]*dto.DataExportResponse, 0, len(exports))
	for _, export := range exports {
		responses = append(responses, dto.ToDataExportResponse(export))
	}

	return &dto.DataExportListResponse{
		Exports:    responses,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: result.TotalPages,
	}, nil
}

// ============================================================================
// Cleanup Operations
// ============================================================================

// DeleteExpiredExports deletes expired export records and their associated files
func (s *dataExportService) DeleteExpiredExports(ctx context.Context) (int64, error) {
	s.logger.Info("cleaning up expired exports")

	// First, get expired exports to delete their files
	expiredExports, err := s.repos.DataExport.FindExpiredExports(ctx)
	if err != nil {
		s.logger.Error("failed to find expired exports", zap.Error(err))
		return 0, fmt.Errorf("failed to find expired exports: %w", err)
	}

	// Delete associated files from storage
	for _, export := range expiredExports {
		if export.FileURL != "" {
			// TODO: Delete file from storage (S3, GCS, etc.)
			// Example: s.storageService.DeleteFile(ctx, export.FileURL)
			s.logger.Info("deleting export file",
				zap.String("export_id", export.ID.String()),
				zap.String("file_url", export.FileURL),
			)
		}
	}

	// Delete the records
	count, err := s.repos.DataExport.DeleteExpiredExports(ctx)
	if err != nil {
		s.logger.Error("failed to delete expired exports", zap.Error(err))
		return 0, fmt.Errorf("failed to delete expired exports: %w", err)
	}

	s.logger.Info("expired exports deleted", zap.Int64("count", count))
	return count, nil
}

// DeleteExportsByTenant deletes all exports for a tenant (for tenant deletion)
func (s *dataExportService) DeleteExportsByTenant(ctx context.Context, tenantID uuid.UUID) error {
	s.logger.Info("deleting all exports for tenant", zap.String("tenant_id", tenantID.String()))

	// Get all exports for tenant to delete their files
	filter := &dto.DataExportFilter{
		Page:     1,
		PageSize: 1000,
	}

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	exports, _, err := s.repos.DataExport.FindByTenant(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to find tenant exports", zap.Error(err))
		return fmt.Errorf("failed to find tenant exports: %w", err)
	}

	// Delete associated files
	for _, export := range exports {
		if export.FileURL != "" {
			// TODO: Delete file from storage
			s.logger.Info("deleting export file",
				zap.String("export_id", export.ID.String()),
				zap.String("file_url", export.FileURL),
			)
		}
	}

	// Delete all export records
	if err := s.repos.DataExport.DeleteByTenant(ctx, tenantID); err != nil {
		s.logger.Error("failed to delete tenant exports", zap.Error(err))
		return fmt.Errorf("failed to delete tenant exports: %w", err)
	}

	s.logger.Info("tenant exports deleted", zap.String("tenant_id", tenantID.String()))
	return nil
}
