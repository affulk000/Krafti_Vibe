package repository

import (
	"context"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// TenantInvitation Repository
// ============================================================================

// TenantInvitationRepository defines the interface for tenant invitation operations
type TenantInvitationRepository interface {
	BaseRepository[models.TenantInvitation]

	// Query Operations
	FindByToken(ctx context.Context, token string) (*models.TenantInvitation, error)
	FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.TenantInvitation, PaginationResult, error)
	FindByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]*models.TenantInvitation, error)
	FindByEmail(ctx context.Context, email string) ([]*models.TenantInvitation, error)
	FindPendingByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantInvitation, error)
	FindExpiredInvitations(ctx context.Context) ([]*models.TenantInvitation, error)

	// Bulk Operations
	DeleteExpiredInvitations(ctx context.Context) (int64, error)
	DeleteByTenant(ctx context.Context, tenantID uuid.UUID) error
}

// ============================================================================
// TenantUsageTracking Repository
// ============================================================================

// TenantUsageTrackingRepository defines the interface for usage tracking operations
type TenantUsageTrackingRepository interface {
	BaseRepository[models.TenantUsageTracking]

	// Query Operations
	FindByTenantAndDate(ctx context.Context, tenantID uuid.UUID, date time.Time) (*models.TenantUsageTracking, error)
	FindByTenantAndDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.TenantUsageTracking, error)
	FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.TenantUsageTracking, PaginationResult, error)

	// Aggregation Operations
	GetTotalAPICallsForPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (int64, error)
	GetAverageAPICallsPerDay(ctx context.Context, tenantID uuid.UUID, days int) (float64, error)
	GetPeakUsageDay(ctx context.Context, tenantID uuid.UUID, days int) (*models.TenantUsageTracking, error)

	// Update Operations
	IncrementAPICall(ctx context.Context, tenantID uuid.UUID, date time.Time) error
	IncrementFeatureUsage(ctx context.Context, tenantID uuid.UUID, date time.Time, feature string, count int) error
	UpdateStorageUsage(ctx context.Context, tenantID uuid.UUID, date time.Time, storageGB int64) error
	UpdateBandwidthUsage(ctx context.Context, tenantID uuid.UUID, date time.Time, bandwidthGB int64) error

	// Cleanup Operations
	DeleteOldRecords(ctx context.Context, olderThan time.Time) (int64, error)
	ArchiveOldRecords(ctx context.Context, olderThan time.Time) error
}

// ============================================================================
// DataExportRequest Repository
// ============================================================================

// DataExportRequestRepository defines the interface for data export request operations
type DataExportRequestRepository interface {
	BaseRepository[models.DataExportRequest]

	// Query Operations
	FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.DataExportRequest, PaginationResult, error)
	FindPendingByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.DataExportRequest, error)
	FindByStatus(ctx context.Context, status string, pagination PaginationParams) ([]*models.DataExportRequest, PaginationResult, error)
	FindExpiredExports(ctx context.Context) ([]*models.DataExportRequest, error)

	// Update Operations
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	SetFileURL(ctx context.Context, id uuid.UUID, fileURL string, expiresAt time.Time) error
	SetError(ctx context.Context, id uuid.UUID, errorMessage string) error
	MarkCompleted(ctx context.Context, id uuid.UUID, fileURL string, expiresAt time.Time) error

	// Cleanup Operations
	DeleteExpiredExports(ctx context.Context) (int64, error)
	DeleteByTenant(ctx context.Context, tenantID uuid.UUID) error
}

// ============================================================================
// Repository Implementations
// ============================================================================

// tenantInvitationRepository implements TenantInvitationRepository
type tenantInvitationRepository struct {
	BaseRepository[models.TenantInvitation]
	db     *gorm.DB
	logger log.AllLogger
}

// NewTenantInvitationRepository creates a new TenantInvitationRepository instance
func NewTenantInvitationRepository(db *gorm.DB, config ...RepositoryConfig) TenantInvitationRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.TenantInvitation](db, cfg)

	return &tenantInvitationRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
	}
}

// FindByToken retrieves an invitation by token
func (r *tenantInvitationRepository) FindByToken(ctx context.Context, token string) (*models.TenantInvitation, error) {
	if token == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "token cannot be empty", errors.ErrInvalidInput)
	}

	var invitation models.TenantInvitation
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("Inviter").
		Where("token = ?", token).
		First(&invitation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "invitation not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find invitation", err)
	}

	return &invitation, nil
}

// FindByTenant retrieves invitations for a tenant
func (r *tenantInvitationRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.TenantInvitation, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.TenantInvitation{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invitations", err)
	}

	var invitations []*models.TenantInvitation
	if err := r.db.WithContext(ctx).
		Preload("Inviter").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&invitations).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find invitations", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invitations, paginationResult, nil
}

// FindByTenantAndEmail retrieves invitations for a specific email in a tenant
func (r *tenantInvitationRepository) FindByTenantAndEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]*models.TenantInvitation, error) {
	var invitations []*models.TenantInvitation
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND email = ?", tenantID, email).
		Order("created_at DESC").
		Find(&invitations).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find invitations", err)
	}

	return invitations, nil
}

// FindByEmail retrieves all invitations for an email
func (r *tenantInvitationRepository) FindByEmail(ctx context.Context, email string) ([]*models.TenantInvitation, error) {
	var invitations []*models.TenantInvitation
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("email = ?", email).
		Order("created_at DESC").
		Find(&invitations).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find invitations", err)
	}

	return invitations, nil
}

// FindPendingByTenant retrieves pending invitations for a tenant
func (r *tenantInvitationRepository) FindPendingByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantInvitation, error) {
	var invitations []*models.TenantInvitation
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND accepted_at IS NULL AND expires_at > ?", tenantID, time.Now()).
		Order("created_at DESC").
		Find(&invitations).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find pending invitations", err)
	}

	return invitations, nil
}

// FindExpiredInvitations retrieves all expired invitations
func (r *tenantInvitationRepository) FindExpiredInvitations(ctx context.Context) ([]*models.TenantInvitation, error) {
	var invitations []*models.TenantInvitation
	if err := r.db.WithContext(ctx).
		Where("accepted_at IS NULL AND expires_at < ?", time.Now()).
		Find(&invitations).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find expired invitations", err)
	}

	return invitations, nil
}

// DeleteExpiredInvitations deletes all expired invitations
func (r *tenantInvitationRepository) DeleteExpiredInvitations(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("accepted_at IS NULL AND expires_at < ?", time.Now()).
		Delete(&models.TenantInvitation{})

	if result.Error != nil {
		return 0, errors.NewRepositoryError("DELETE_FAILED", "failed to delete expired invitations", result.Error)
	}

	return result.RowsAffected, nil
}

// DeleteByTenant deletes all invitations for a tenant
func (r *tenantInvitationRepository) DeleteByTenant(ctx context.Context, tenantID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Delete(&models.TenantInvitation{})

	if result.Error != nil {
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete tenant invitations", result.Error)
	}

	return nil
}

// ============================================================================
// TenantUsageTracking Repository Implementation
// ============================================================================

// tenantUsageTrackingRepository implements TenantUsageTrackingRepository
type tenantUsageTrackingRepository struct {
	BaseRepository[models.TenantUsageTracking]
	db     *gorm.DB
	logger log.AllLogger
}

// NewTenantUsageTrackingRepository creates a new TenantUsageTrackingRepository instance
func NewTenantUsageTrackingRepository(db *gorm.DB, config ...RepositoryConfig) TenantUsageTrackingRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.TenantUsageTracking](db, cfg)

	return &tenantUsageTrackingRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
	}
}

// FindByTenantAndDate retrieves usage for a specific date
func (r *tenantUsageTrackingRepository) FindByTenantAndDate(ctx context.Context, tenantID uuid.UUID, date time.Time) (*models.TenantUsageTracking, error) {
	normalizedDate := date.Truncate(24 * time.Hour)

	var usage models.TenantUsageTracking
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND date = ?", tenantID, normalizedDate).
		First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "usage record not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find usage record", err)
	}

	return &usage, nil
}

// FindByTenantAndDateRange retrieves usage for a date range
func (r *tenantUsageTrackingRepository) FindByTenantAndDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.TenantUsageTracking, error) {
	var usages []*models.TenantUsageTracking
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Order("date ASC").
		Find(&usages).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find usage records", err)
	}

	return usages, nil
}

// FindByTenant retrieves usage records for a tenant
func (r *tenantUsageTrackingRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.TenantUsageTracking, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.TenantUsageTracking{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count usage records", err)
	}

	var usages []*models.TenantUsageTracking
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("date DESC").
		Find(&usages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find usage records", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return usages, paginationResult, nil
}

// GetTotalAPICallsForPeriod returns total API calls for a period
func (r *tenantUsageTrackingRepository) GetTotalAPICallsForPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.TenantUsageTracking{}).
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Select("COALESCE(SUM(api_calls_count), 0)").
		Scan(&total).Error; err != nil {
		return 0, errors.NewRepositoryError("QUERY_FAILED", "failed to sum API calls", err)
	}

	return total, nil
}

// GetAverageAPICallsPerDay returns average API calls per day
func (r *tenantUsageTrackingRepository) GetAverageAPICallsPerDay(ctx context.Context, tenantID uuid.UUID, days int) (float64, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	var average float64
	if err := r.db.WithContext(ctx).
		Model(&models.TenantUsageTracking{}).
		Where("tenant_id = ? AND date >= ?", tenantID, startDate).
		Select("COALESCE(AVG(api_calls_count), 0)").
		Scan(&average).Error; err != nil {
		return 0, errors.NewRepositoryError("QUERY_FAILED", "failed to calculate average", err)
	}

	return average, nil
}

// GetPeakUsageDay returns the day with highest API calls
func (r *tenantUsageTrackingRepository) GetPeakUsageDay(ctx context.Context, tenantID uuid.UUID, days int) (*models.TenantUsageTracking, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	var usage models.TenantUsageTracking
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND date >= ?", tenantID, startDate).
		Order("api_calls_count DESC").
		First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find peak usage", err)
	}

	return &usage, nil
}

// IncrementAPICall increments the API call counter
func (r *tenantUsageTrackingRepository) IncrementAPICall(ctx context.Context, tenantID uuid.UUID, date time.Time) error {
	normalizedDate := date.Truncate(24 * time.Hour)

	result := r.db.WithContext(ctx).
		Model(&models.TenantUsageTracking{}).
		Where("tenant_id = ? AND date = ?", tenantID, normalizedDate).
		Update("api_calls_count", gorm.Expr("api_calls_count + 1"))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment API call", result.Error)
	}

	return nil
}

// IncrementFeatureUsage increments feature usage counter
func (r *tenantUsageTrackingRepository) IncrementFeatureUsage(ctx context.Context, tenantID uuid.UUID, date time.Time, feature string, count int) error {
	normalizedDate := date.Truncate(24 * time.Hour)

	var column string
	switch feature {
	case "booking":
		column = "bookings_created"
	case "project":
		column = "projects_created"
	case "sms":
		column = "sms_sent"
	case "email":
		column = "emails_sent"
	default:
		return errors.NewRepositoryError("INVALID_INPUT", "unknown feature", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.TenantUsageTracking{}).
		Where("tenant_id = ? AND date = ?", tenantID, normalizedDate).
		Update(column, gorm.Expr(column+" + ?", count))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment feature usage", result.Error)
	}

	return nil
}

// UpdateStorageUsage updates storage usage
func (r *tenantUsageTrackingRepository) UpdateStorageUsage(ctx context.Context, tenantID uuid.UUID, date time.Time, storageGB int64) error {
	normalizedDate := date.Truncate(24 * time.Hour)

	result := r.db.WithContext(ctx).
		Model(&models.TenantUsageTracking{}).
		Where("tenant_id = ? AND date = ?", tenantID, normalizedDate).
		Update("storage_used_gb", storageGB)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update storage usage", result.Error)
	}

	return nil
}

// UpdateBandwidthUsage updates bandwidth usage
func (r *tenantUsageTrackingRepository) UpdateBandwidthUsage(ctx context.Context, tenantID uuid.UUID, date time.Time, bandwidthGB int64) error {
	normalizedDate := date.Truncate(24 * time.Hour)

	result := r.db.WithContext(ctx).
		Model(&models.TenantUsageTracking{}).
		Where("tenant_id = ? AND date = ?", tenantID, normalizedDate).
		Update("bandwidth_used_gb", gorm.Expr("bandwidth_used_gb + ?", bandwidthGB))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update bandwidth usage", result.Error)
	}

	return nil
}

// DeleteOldRecords deletes records older than specified time
func (r *tenantUsageTrackingRepository) DeleteOldRecords(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("date < ?", olderThan).
		Delete(&models.TenantUsageTracking{})

	if result.Error != nil {
		return 0, errors.NewRepositoryError("DELETE_FAILED", "failed to delete old records", result.Error)
	}

	return result.RowsAffected, nil
}

// ArchiveOldRecords archives old records (placeholder - implement based on archival strategy)
func (r *tenantUsageTrackingRepository) ArchiveOldRecords(ctx context.Context, olderThan time.Time) error {
	// This would typically move records to an archive table or cold storage
	// Implementation depends on specific requirements
	return nil
}

// ============================================================================
// DataExportRequest Repository Implementation
// ============================================================================

// dataExportRequestRepository implements DataExportRequestRepository
type dataExportRequestRepository struct {
	BaseRepository[models.DataExportRequest]
	db     *gorm.DB
	logger log.AllLogger
}

// NewDataExportRequestRepository creates a new DataExportRequestRepository instance
func NewDataExportRequestRepository(db *gorm.DB, config ...RepositoryConfig) DataExportRequestRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.DataExportRequest](db, cfg)

	return &dataExportRequestRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
	}
}

// FindByTenant retrieves export requests for a tenant
func (r *dataExportRequestRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.DataExportRequest, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.DataExportRequest{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count export requests", err)
	}

	var exports []*models.DataExportRequest
	if err := r.db.WithContext(ctx).
		Preload("Requester").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&exports).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find export requests", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return exports, paginationResult, nil
}

// FindPendingByTenant retrieves pending exports for a tenant
func (r *dataExportRequestRepository) FindPendingByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.DataExportRequest, error) {
	var exports []*models.DataExportRequest
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status IN ?", tenantID, []string{"pending", "processing"}).
		Order("created_at DESC").
		Find(&exports).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find pending exports", err)
	}

	return exports, nil
}

// FindByStatus retrieves exports by status
func (r *dataExportRequestRepository) FindByStatus(ctx context.Context, status string, pagination PaginationParams) ([]*models.DataExportRequest, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.DataExportRequest{}).
		Where("status = ?", status).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count exports", err)
	}

	var exports []*models.DataExportRequest
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("Requester").
		Where("status = ?", status).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&exports).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find exports", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return exports, paginationResult, nil
}

// FindExpiredExports retrieves expired exports
func (r *dataExportRequestRepository) FindExpiredExports(ctx context.Context) ([]*models.DataExportRequest, error) {
	var exports []*models.DataExportRequest
	if err := r.db.WithContext(ctx).
		Where("status = 'completed' AND expires_at < ?", time.Now()).
		Find(&exports).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find expired exports", err)
	}

	return exports, nil
}

// UpdateStatus updates the status of an export request
func (r *dataExportRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).
		Model(&models.DataExportRequest{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "export request not found", errors.ErrNotFound)
	}

	return nil
}

// SetFileURL sets the file URL and expiry for a completed export
func (r *dataExportRequestRepository) SetFileURL(ctx context.Context, id uuid.UUID, fileURL string, expiresAt time.Time) error {
	updates := map[string]any{
		"file_url":   fileURL,
		"expires_at": expiresAt,
	}

	result := r.db.WithContext(ctx).
		Model(&models.DataExportRequest{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to set file URL", result.Error)
	}

	return nil
}

// SetError sets an error message on a failed export
func (r *dataExportRequestRepository) SetError(ctx context.Context, id uuid.UUID, errorMessage string) error {
	updates := map[string]any{
		"status": "failed",
	}

	result := r.db.WithContext(ctx).
		Model(&models.DataExportRequest{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to set error", result.Error)
	}

	return nil
}

// MarkCompleted marks an export as completed
func (r *dataExportRequestRepository) MarkCompleted(ctx context.Context, id uuid.UUID, fileURL string, expiresAt time.Time) error {
	now := time.Now()
	updates := map[string]any{
		"status":       "completed",
		"file_url":     fileURL,
		"expires_at":   expiresAt,
		"completed_at": now,
	}

	result := r.db.WithContext(ctx).
		Model(&models.DataExportRequest{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark completed", result.Error)
	}

	return nil
}

// DeleteExpiredExports deletes expired export records and their files
func (r *dataExportRequestRepository) DeleteExpiredExports(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("status = 'completed' AND expires_at < ?", time.Now()).
		Delete(&models.DataExportRequest{})

	if result.Error != nil {
		return 0, errors.NewRepositoryError("DELETE_FAILED", "failed to delete expired exports", result.Error)
	}

	return result.RowsAffected, nil
}

// DeleteByTenant deletes all export requests for a tenant
func (r *dataExportRequestRepository) DeleteByTenant(ctx context.Context, tenantID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Delete(&models.DataExportRequest{})

	if result.Error != nil {
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete tenant exports", result.Error)
	}

	return nil
}
