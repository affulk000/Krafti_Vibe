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

// AuditLogRepository defines the interface for audit log repository operations
type AuditLogRepository interface {
	BaseRepository[models.AuditLog]

	// FindByTenant retrieves audit logs for a tenant
	FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error)

	// FindByUser retrieves audit logs for a user
	FindByUser(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error)

	// FindByEntity retrieves audit logs for an entity
	FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error)

	// FindByAction retrieves audit logs by action type
	FindByAction(ctx context.Context, tenantID uuid.UUID, action models.AuditAction, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error)

	// FindByDateRange retrieves audit logs within a date range
	FindByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error)

	// FindRecent retrieves recent audit logs
	FindRecent(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.AuditLog, error)

	// Search searches audit logs
	Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error)

	// GetEntityHistory retrieves complete history for an entity
	GetEntityHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.AuditLog, error)

	// GetUserActivity retrieves user activity summary
	GetUserActivity(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (map[string]any, error)

	// GetSystemActivity retrieves system-wide activity summary
	GetSystemActivity(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]any, error)

	// CleanupOldLogs removes audit logs older than specified duration
	CleanupOldLogs(ctx context.Context, retentionDays int) error

	// CountByAction counts logs by action type
	CountByAction(ctx context.Context, tenantID uuid.UUID, action models.AuditAction) (int64, error)

	// FindSuspiciousActivity finds potentially suspicious activities
	FindSuspiciousActivity(ctx context.Context, tenantID uuid.UUID, hours int) ([]*models.AuditLog, error)

	// FindByIPAddress retrieves audit logs by IP address
	FindByIPAddress(ctx context.Context, tenantID uuid.UUID, ipAddress string, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error)
}

// auditLogRepository implements AuditLogRepository
type auditLogRepository struct {
	BaseRepository[models.AuditLog]
	db     *gorm.DB
	logger log.AllLogger
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB, config ...RepositoryConfig) AuditLogRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.AuditLog](db, cfg)

	return &auditLogRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
	}
}

// FindByTenant retrieves audit logs for a tenant
func (r *auditLogRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count audit logs", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count audit logs", err)
	}

	// Find paginated results
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to find audit logs", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find audit logs", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return logs, paginationResult, nil
}

// FindByUser retrieves audit logs for a user
func (r *auditLogRepository) FindByUser(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("user_id = ?", userID).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count user audit logs", "user_id", userID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count audit logs", err)
	}

	// Find paginated results
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to find user audit logs", "user_id", userID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find audit logs", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return logs, paginationResult, nil
}

// FindByEntity retrieves audit logs for an entity
func (r *auditLogRepository) FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error) {
	if entityType == "" {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "entity_type cannot be empty", errors.ErrInvalidInput)
	}
	if entityID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "entity_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count entity audit logs", "entity_type", entityType, "entity_id", entityID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count audit logs", err)
	}

	// Find paginated results
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to find entity audit logs", "entity_type", entityType, "entity_id", entityID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find audit logs", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return logs, paginationResult, nil
}

// FindByAction retrieves audit logs by action type
func (r *auditLogRepository) FindByAction(ctx context.Context, tenantID uuid.UUID, action models.AuditAction, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("tenant_id = ? AND action = ?", tenantID, action).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count action audit logs", "tenant_id", tenantID, "action", action, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count audit logs", err)
	}

	// Find paginated results
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND action = ?", tenantID, action).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to find action audit logs", "tenant_id", tenantID, "action", action, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find audit logs", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return logs, paginationResult, nil
}

// FindByDateRange retrieves audit logs within a date range
func (r *auditLogRepository) FindByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, startDate, endDate).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count audit logs by date range", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count audit logs", err)
	}

	// Find paginated results
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, startDate, endDate).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to find audit logs by date range", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find audit logs", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return logs, paginationResult, nil
}

// FindRecent retrieves recent audit logs
func (r *auditLogRepository) FindRecent(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.AuditLog, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if limit <= 0 {
		limit = 50
	}

	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to find recent audit logs", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find recent audit logs", err)
	}

	return logs, nil
}

// Search searches audit logs
func (r *auditLogRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("tenant_id = ? AND (description ILIKE ? OR entity_type ILIKE ? OR user_email ILIKE ?)", tenantID, searchPattern, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count search results", "tenant_id", tenantID, "query", query, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count audit logs", err)
	}

	// Find paginated results
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (description ILIKE ? OR entity_type ILIKE ? OR user_email ILIKE ?)", tenantID, searchPattern, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to search audit logs", "tenant_id", tenantID, "query", query, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search audit logs", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return logs, paginationResult, nil
}

// GetEntityHistory retrieves complete history for an entity
func (r *auditLogRepository) GetEntityHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.AuditLog, error) {
	if entityType == "" || entityID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "entity_type and entity_id are required", errors.ErrInvalidInput)
	}

	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at ASC").
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to get entity history", "entity_type", entityType, "entity_id", entityID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to get entity history", err)
	}

	return logs, nil
}

// GetUserActivity retrieves user activity summary
func (r *auditLogRepository) GetUserActivity(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (map[string]any, error) {
	if userID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	type ActionCount struct {
		Action models.AuditAction
		Count  int64
	}

	var actionCounts []ActionCount
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Group("action").
		Scan(&actionCounts).Error; err != nil {
		r.logger.Error("failed to get user activity", "user_id", userID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get user activity", err)
	}

	activity := make(map[string]any)
	activity["user_id"] = userID
	activity["start_date"] = startDate
	activity["end_date"] = endDate

	actionMap := make(map[string]int64)
	var total int64
	for _, ac := range actionCounts {
		actionMap[string(ac.Action)] = ac.Count
		total += ac.Count
	}
	activity["actions"] = actionMap
	activity["total_actions"] = total

	return activity, nil
}

// GetSystemActivity retrieves system-wide activity summary
func (r *auditLogRepository) GetSystemActivity(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]any, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	type ActionCount struct {
		Action models.AuditAction
		Count  int64
	}

	var actionCounts []ActionCount
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, startDate, endDate).
		Group("action").
		Scan(&actionCounts).Error; err != nil {
		r.logger.Error("failed to get system activity", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get system activity", err)
	}

	activity := make(map[string]any)
	activity["tenant_id"] = tenantID
	activity["start_date"] = startDate
	activity["end_date"] = endDate

	actionMap := make(map[string]int64)
	var total int64
	for _, ac := range actionCounts {
		actionMap[string(ac.Action)] = ac.Count
		total += ac.Count
	}
	activity["actions"] = actionMap
	activity["total_actions"] = total

	// Get unique users
	var uniqueUsers int64
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at >= ? AND created_at <= ? AND user_id IS NOT NULL", tenantID, startDate, endDate).
		Distinct("user_id").
		Count(&uniqueUsers)
	activity["unique_users"] = uniqueUsers

	return activity, nil
}

// CleanupOldLogs removes audit logs older than specified duration
func (r *auditLogRepository) CleanupOldLogs(ctx context.Context, retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 90 // Default to 90 days
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&models.AuditLog{})

	if result.Error != nil {
		r.logger.Error("failed to cleanup old audit logs", "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to cleanup old audit logs", result.Error)
	}

	r.logger.Info("cleaned up old audit logs", "count", result.RowsAffected, "retention_days", retentionDays)
	return nil
}

// CountByAction counts logs by action type
func (r *auditLogRepository) CountByAction(ctx context.Context, tenantID uuid.UUID, action models.AuditAction) (int64, error) {
	if tenantID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("tenant_id = ? AND action = ?", tenantID, action).
		Count(&count).Error; err != nil {
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count logs by action", err)
	}

	return count, nil
}

// FindSuspiciousActivity finds potentially suspicious activities
func (r *auditLogRepository) FindSuspiciousActivity(ctx context.Context, tenantID uuid.UUID, hours int) ([]*models.AuditLog, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if hours <= 0 {
		hours = 24
	}

	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	// Find failed login attempts, deletes, and other suspicious actions
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at >= ? AND action IN ?", tenantID, cutoffTime,
			[]models.AuditAction{models.AuditActionDelete, models.AuditActionUpdate}).
		Order("created_at DESC").
		Limit(100).
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to find suspicious activity", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find suspicious activity", err)
	}

	return logs, nil
}

// FindByIPAddress retrieves audit logs by IP address
func (r *auditLogRepository) FindByIPAddress(ctx context.Context, tenantID uuid.UUID, ipAddress string, pagination PaginationParams) ([]*models.AuditLog, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}
	if ipAddress == "" {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "ip_address cannot be empty", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("tenant_id = ? AND ip_address = ?", tenantID, ipAddress).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count logs by IP", "tenant_id", tenantID, "ip", ipAddress, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count audit logs", err)
	}

	// Find paginated results
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND ip_address = ?", tenantID, ipAddress).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		r.logger.Error("failed to find logs by IP", "tenant_id", tenantID, "ip", ipAddress, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find audit logs", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return logs, paginationResult, nil
}
