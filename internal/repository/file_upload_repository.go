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

// FileUploadRepository defines the interface for file upload repository operations
type FileUploadRepository interface {
	BaseRepository[models.FileUpload]

	// FindByTenantID retrieves files for a tenant with pagination
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.FileUpload, PaginationResult, error)

	// FindByUploadedBy retrieves files uploaded by a user
	FindByUploadedBy(ctx context.Context, userID uuid.UUID) ([]*models.FileUpload, error)

	// FindByEntity retrieves files related to an entity
	FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.FileUpload, error)

	// FindByFileType retrieves files by type
	FindByFileType(ctx context.Context, tenantID uuid.UUID, fileType models.FileType) ([]*models.FileUpload, error)

	// GetStorageUsage calculates total storage used by a tenant
	GetStorageUsage(ctx context.Context, tenantID uuid.UUID) (int64, error)

	// GetFileStats retrieves file statistics
	GetFileStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error)

	// CleanupOrphanedFiles removes files not associated with any entity
	CleanupOrphanedFiles(ctx context.Context, olderThan time.Duration) error

	// FindExpiredFiles retrieves files that have expired
	FindExpiredFiles(ctx context.Context) ([]*models.FileUpload, error)

	// UpdateFileAccess updates last accessed timestamp
	UpdateFileAccess(ctx context.Context, fileID uuid.UUID) error

	// CountByTenant counts files for a tenant
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)

	// Search searches files by filename
	Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.FileUpload, PaginationResult, error)
}

// fileUploadRepository implements FileUploadRepository
type fileUploadRepository struct {
	BaseRepository[models.FileUpload]
	db     *gorm.DB
	logger log.AllLogger
}

// NewFileUploadRepository creates a new FileUploadRepository instance
func NewFileUploadRepository(db *gorm.DB, config ...RepositoryConfig) FileUploadRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.FileUpload](db, cfg)

	return &fileUploadRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
	}
}

// FindByTenantID retrieves files for a tenant with pagination
func (r *fileUploadRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.FileUpload, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.FileUpload{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count files", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count files", err)
	}

	// Find paginated results
	var files []*models.FileUpload
	if err := r.db.WithContext(ctx).
		Preload("UploadedBy").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&files).Error; err != nil {
		r.logger.Error("failed to find files", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find files", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return files, paginationResult, nil
}

// FindByUploadedBy retrieves files uploaded by a user
func (r *fileUploadRepository) FindByUploadedBy(ctx context.Context, userID uuid.UUID) ([]*models.FileUpload, error) {
	if userID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	var files []*models.FileUpload
	if err := r.db.WithContext(ctx).
		Where("uploaded_by_id = ?", userID).
		Order("created_at DESC").
		Find(&files).Error; err != nil {
		r.logger.Error("failed to find files by uploader", "user_id", userID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find files", err)
	}

	return files, nil
}

// FindByEntity retrieves files related to an entity
func (r *fileUploadRepository) FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.FileUpload, error) {
	if entityType == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "entity_type cannot be empty", errors.ErrInvalidInput)
	}
	if entityID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "entity_id cannot be nil", errors.ErrInvalidInput)
	}

	var files []*models.FileUpload
	if err := r.db.WithContext(ctx).
		Where("related_entity_type = ? AND related_entity_id = ?", entityType, entityID).
		Preload("UploadedBy").
		Order("created_at DESC").
		Find(&files).Error; err != nil {
		r.logger.Error("failed to find files by entity", "entity_type", entityType, "entity_id", entityID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find files", err)
	}

	return files, nil
}

// FindByFileType retrieves files by type
func (r *fileUploadRepository) FindByFileType(ctx context.Context, tenantID uuid.UUID, fileType models.FileType) ([]*models.FileUpload, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var files []*models.FileUpload
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND file_type = ?", tenantID, fileType).
		Order("created_at DESC").
		Find(&files).Error; err != nil {
		r.logger.Error("failed to find files by type", "tenant_id", tenantID, "file_type", fileType, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find files", err)
	}

	return files, nil
}

// GetStorageUsage calculates total storage used by a tenant
func (r *fileUploadRepository) GetStorageUsage(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	if tenantID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var totalSize int64
	if err := r.db.WithContext(ctx).
		Model(&models.FileUpload{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize).Error; err != nil {
		r.logger.Error("failed to calculate storage usage", "tenant_id", tenantID, "error", err)
		return 0, errors.NewRepositoryError("QUERY_FAILED", "failed to calculate storage usage", err)
	}

	return totalSize, nil
}

// GetFileStats retrieves file statistics
func (r *fileUploadRepository) GetFileStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := make(map[string]interface{})

	// Total files
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.FileUpload{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		r.logger.Error("failed to count total files", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get file stats", err)
	}
	stats["total_files"] = total

	// Storage usage
	storageUsed, err := r.GetStorageUsage(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	stats["storage_used_bytes"] = storageUsed
	stats["storage_used_mb"] = float64(storageUsed) / (1024 * 1024)
	stats["storage_used_gb"] = float64(storageUsed) / (1024 * 1024 * 1024)

	// Files by type
	type TypeCount struct {
		FileType models.FileType
		Count    int64
	}
	var typeCounts []TypeCount
	if err := r.db.WithContext(ctx).
		Model(&models.FileUpload{}).
		Select("file_type, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("file_type").
		Scan(&typeCounts).Error; err != nil {
		r.logger.Error("failed to count files by type", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get file stats", err)
	}

	typeMap := make(map[string]int64)
	for _, tc := range typeCounts {
		typeMap[string(tc.FileType)] = tc.Count
	}
	stats["by_type"] = typeMap

	// Recent uploads (last 30 days)
	var recentCount int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := r.db.WithContext(ctx).
		Model(&models.FileUpload{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, thirtyDaysAgo).
		Count(&recentCount).Error; err != nil {
		r.logger.Error("failed to count recent files", "tenant_id", tenantID, "error", err)
	} else {
		stats["recent_uploads_30d"] = recentCount
	}

	return stats, nil
}

// CleanupOrphanedFiles removes files not associated with any entity
func (r *fileUploadRepository) CleanupOrphanedFiles(ctx context.Context, olderThan time.Duration) error {
	cutoffDate := time.Now().Add(-olderThan)

	result := r.db.WithContext(ctx).
		Where("related_entity_id IS NULL AND created_at < ?", cutoffDate).
		Delete(&models.FileUpload{})

	if result.Error != nil {
		r.logger.Error("failed to cleanup orphaned files", "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to cleanup orphaned files", result.Error)
	}

	r.logger.Info("cleaned up orphaned files", "count", result.RowsAffected, "older_than", olderThan)
	return nil
}

// FindExpiredFiles retrieves files that have expired
func (r *fileUploadRepository) FindExpiredFiles(ctx context.Context) ([]*models.FileUpload, error) {
	var files []*models.FileUpload
	if err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Find(&files).Error; err != nil {
		r.logger.Error("failed to find expired files", "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find expired files", err)
	}

	return files, nil
}

// UpdateFileAccess updates last accessed timestamp
func (r *fileUploadRepository) UpdateFileAccess(ctx context.Context, fileID uuid.UUID) error {
	if fileID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "file_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.FileUpload{}).
		Where("id = ?", fileID).
		Update("last_accessed_at", time.Now())

	if result.Error != nil {
		r.logger.Error("failed to update file access", "file_id", fileID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update file access", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "file not found", errors.ErrNotFound)
	}

	return nil
}

// CountByTenant counts files for a tenant
func (r *fileUploadRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	if tenantID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.FileUpload{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error; err != nil {
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count files", err)
	}

	return count, nil
}

// FindByMimeType retrieves files by MIME type
func (r *fileUploadRepository) FindByMimeType(ctx context.Context, tenantID uuid.UUID, mimeType string) ([]*models.FileUpload, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var files []*models.FileUpload
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND mime_type = ?", tenantID, mimeType).
		Order("created_at DESC").
		Find(&files).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find files", err)
	}

	return files, nil
}

// FindLargeFiles retrieves files larger than specified size
func (r *fileUploadRepository) FindLargeFiles(ctx context.Context, tenantID uuid.UUID, minSizeBytes int64) ([]*models.FileUpload, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var files []*models.FileUpload
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND file_size >= ?", tenantID, minSizeBytes).
		Order("file_size DESC").
		Find(&files).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find large files", err)
	}

	return files, nil
}

// Search searches files by filename
func (r *fileUploadRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.FileUpload, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.FileUpload{}).
		Where("tenant_id = ? AND (file_name ILIKE ? OR original_name ILIKE ?)", tenantID, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count files", err)
	}

	// Find paginated results
	var files []*models.FileUpload
	if err := r.db.WithContext(ctx).
		Preload("UploadedBy").
		Where("tenant_id = ? AND (file_name ILIKE ? OR original_name ILIKE ?)", tenantID, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&files).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search files", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return files, paginationResult, nil
}
