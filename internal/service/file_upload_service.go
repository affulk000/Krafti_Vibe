package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// FileUploadService defines the interface for file upload operations
type FileUploadService interface {
	// CRUD Operations
	UploadFile(ctx context.Context, tenantID, uploadedByID uuid.UUID, req *dto.UploadFileRequest) (*dto.FileUploadResponse, error)
	GetFile(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.FileUploadResponse, error)
	UpdateFile(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateFileRequest) (*dto.FileUploadResponse, error)
	DeleteFile(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// List & Filter Operations
	ListFiles(ctx context.Context, filter *dto.FileUploadFilter) (*dto.FileUploadListResponse, error)
	ListFilesByUploader(ctx context.Context, uploaderID uuid.UUID, tenantID uuid.UUID) ([]*dto.FileUploadResponse, error)
	ListFilesByEntity(ctx context.Context, entityType string, entityID uuid.UUID, tenantID uuid.UUID) ([]*dto.FileUploadResponse, error)
	ListFilesByType(ctx context.Context, tenantID uuid.UUID, fileType models.FileType) ([]*dto.FileUploadResponse, error)
	SearchFiles(ctx context.Context, tenantID uuid.UUID, query string, page, pageSize int) (*dto.FileUploadListResponse, error)

	// Statistics & Analytics
	GetFileStats(ctx context.Context, tenantID uuid.UUID) (*dto.FileStatsResponse, error)
	GetStorageUsage(ctx context.Context, tenantID uuid.UUID) (int64, error)

	// Utilities
	ValidateFileUpload(req *dto.UploadFileRequest) error
	CleanupOrphanedFiles(ctx context.Context, olderThanDays int) error
	UpdateFileAccess(ctx context.Context, fileID uuid.UUID) error
}

// fileUploadService implements FileUploadService
type fileUploadService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewFileUploadService creates a new file upload service
func NewFileUploadService(repos *repository.Repositories, logger log.AllLogger) FileUploadService {
	return &fileUploadService{
		repos:  repos,
		logger: logger,
	}
}

// UploadFile uploads a new file
func (s *fileUploadService) UploadFile(ctx context.Context, tenantID, uploadedByID uuid.UUID, req *dto.UploadFileRequest) (*dto.FileUploadResponse, error) {
	// Validate request
	if err := s.ValidateFileUpload(req); err != nil {
		return nil, err
	}

	// Verify uploader exists
	uploader, err := s.repos.User.GetByID(ctx, uploadedByID)
	if err != nil {
		s.logger.Error("failed to get uploader", "uploader_id", uploadedByID, "error", err)
		return nil, errors.NewNotFoundError("uploader")
	}

	// Verify uploader belongs to tenant
	if uploader.TenantID != nil && *uploader.TenantID != tenantID {
		return nil, errors.NewValidationError("Uploader does not belong to tenant")
	}

	// Verify related entity if provided
	if req.RelatedEntityID != nil && req.RelatedEntityType != "" {
		if err := s.verifyRelatedEntity(ctx, req.RelatedEntityType, *req.RelatedEntityID, tenantID); err != nil {
			return nil, err
		}
	}

	// Set default storage provider
	storageProvider := req.StorageProvider
	if storageProvider == "" {
		storageProvider = "s3"
	}

	// Create file upload record
	file := &models.FileUpload{
		TenantID:          tenantID,
		UploadedByID:      uploadedByID,
		FileName:          req.FileName,
		FileType:          req.FileType,
		MimeType:          req.MimeType,
		FileSize:          req.FileSize,
		FilePath:          req.FilePath,
		FileURL:           req.FileURL,
		ThumbnailURL:      req.ThumbnailURL,
		RelatedEntityType: req.RelatedEntityType,
		RelatedEntityID:   req.RelatedEntityID,
		StorageProvider:   storageProvider,
		StorageBucket:     req.StorageBucket,
		Metadata:          req.Metadata,
	}

	if err := s.repos.FileUpload.Create(ctx, file); err != nil {
		s.logger.Error("failed to create file upload", "error", err)
		return nil, errors.NewRepositoryError("CREATE_FAILED", "Failed to upload file", err)
	}

	// Load uploader relationship
	file.UploadedBy = uploader

	s.logger.Info("file uploaded", "file_id", file.ID, "file_name", req.FileName, "size_mb", file.GetFileSizeMB())
	return dto.ToFileUploadResponse(file), nil
}

// GetFile retrieves a file by ID
func (s *fileUploadService) GetFile(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.FileUploadResponse, error) {
	file, err := s.repos.FileUpload.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get file", "id", id, "error", err)
		return nil, errors.NewNotFoundError("file")
	}

	// Verify tenant access
	if file.TenantID != tenantID {
		return nil, errors.NewNotFoundError("file")
	}

	// Load uploader relationship
	if file.UploadedByID != uuid.Nil {
		uploader, err := s.repos.User.GetByID(ctx, file.UploadedByID)
		if err == nil {
			file.UploadedBy = uploader
		}
	}

	return dto.ToFileUploadResponse(file), nil
}

// UpdateFile updates a file's metadata
func (s *fileUploadService) UpdateFile(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateFileRequest) (*dto.FileUploadResponse, error) {
	// Get existing file
	file, err := s.repos.FileUpload.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get file", "id", id, "error", err)
		return nil, errors.NewNotFoundError("file")
	}

	// Verify tenant access
	if file.TenantID != tenantID {
		return nil, errors.NewNotFoundError("file")
	}

	// Update fields
	if req.FileName != nil {
		file.FileName = *req.FileName
	}
	if req.ThumbnailURL != nil {
		file.ThumbnailURL = *req.ThumbnailURL
	}
	if req.RelatedEntityType != nil {
		file.RelatedEntityType = *req.RelatedEntityType
	}
	if req.RelatedEntityID != nil {
		// Verify related entity
		if req.RelatedEntityType != nil {
			if err := s.verifyRelatedEntity(ctx, *req.RelatedEntityType, *req.RelatedEntityID, tenantID); err != nil {
				return nil, err
			}
		}
		file.RelatedEntityID = req.RelatedEntityID
	}
	if req.Metadata != nil {
		file.Metadata = req.Metadata
	}

	// Save file
	if err := s.repos.FileUpload.Update(ctx, file); err != nil {
		s.logger.Error("failed to update file", "id", id, "error", err)
		return nil, errors.NewRepositoryError("UPDATE_FAILED", "Failed to update file", err)
	}

	s.logger.Info("file updated", "id", id)

	// Get updated file
	updated, err := s.repos.FileUpload.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to retrieve updated file", err)
	}

	return dto.ToFileUploadResponse(updated), nil
}

// DeleteFile deletes a file
func (s *fileUploadService) DeleteFile(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// Get file
	file, err := s.repos.FileUpload.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get file", "id", id, "error", err)
		return errors.NewNotFoundError("file")
	}

	// Verify tenant access
	if file.TenantID != tenantID {
		return errors.NewNotFoundError("file")
	}

	// Delete from database
	if err := s.repos.FileUpload.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete file", "id", id, "error", err)
		return errors.NewRepositoryError("DELETE_FAILED", "Failed to delete file", err)
	}

	s.logger.Info("file deleted", "id", id, "file_name", file.FileName)
	return nil
}

// ListFiles lists files with filtering
func (s *fileUploadService) ListFiles(ctx context.Context, filter *dto.FileUploadFilter) (*dto.FileUploadListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(filter.Page, 1),
		PageSize: max(filter.PageSize, 1),
	}
	pagination.PageSize = min(pagination.PageSize, 100)

	var files []*models.FileUpload
	var paginationResult repository.PaginationResult
	var err error

	// Use appropriate repository method based on filters
	if filter.SearchQuery != "" {
		files, paginationResult, err = s.repos.FileUpload.Search(ctx, filter.TenantID, filter.SearchQuery, pagination)
	} else if filter.RelatedEntityType != nil && filter.RelatedEntityID != nil {
		files, err = s.repos.FileUpload.FindByEntity(ctx, *filter.RelatedEntityType, *filter.RelatedEntityID)
		// Manual pagination for non-paginated results
		totalItems := int64(len(files))
		startIdx := pagination.Offset()
		endIdx := min(startIdx+pagination.Limit(), len(files))
		if startIdx < len(files) {
			files = files[startIdx:endIdx]
		} else {
			files = []*models.FileUpload{}
		}
		paginationResult = repository.CalculatePagination(pagination, totalItems)
	} else if filter.FileType != nil {
		files, err = s.repos.FileUpload.FindByFileType(ctx, filter.TenantID, *filter.FileType)
		// Manual pagination
		totalItems := int64(len(files))
		startIdx := pagination.Offset()
		endIdx := min(startIdx+pagination.Limit(), len(files))
		if startIdx < len(files) {
			files = files[startIdx:endIdx]
		} else {
			files = []*models.FileUpload{}
		}
		paginationResult = repository.CalculatePagination(pagination, totalItems)
	} else if filter.UploadedByID != nil {
		files, err = s.repos.FileUpload.FindByUploadedBy(ctx, *filter.UploadedByID)
		// Manual pagination
		totalItems := int64(len(files))
		startIdx := pagination.Offset()
		endIdx := min(startIdx+pagination.Limit(), len(files))
		if startIdx < len(files) {
			files = files[startIdx:endIdx]
		} else {
			files = []*models.FileUpload{}
		}
		paginationResult = repository.CalculatePagination(pagination, totalItems)
	} else {
		files, paginationResult, err = s.repos.FileUpload.FindByTenantID(ctx, filter.TenantID, pagination)
	}

	if err != nil {
		s.logger.Error("failed to list files", "error", err)
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to list files", err)
	}

	return &dto.FileUploadListResponse{
		Files:       dto.ToFileUploadResponses(files),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListFilesByUploader lists files uploaded by a user
func (s *fileUploadService) ListFilesByUploader(ctx context.Context, uploaderID uuid.UUID, tenantID uuid.UUID) ([]*dto.FileUploadResponse, error) {
	files, err := s.repos.FileUpload.FindByUploadedBy(ctx, uploaderID)
	if err != nil {
		s.logger.Error("failed to list files by uploader", "uploader_id", uploaderID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list files", err)
	}

	// Filter by tenant
	var tenantFiles []*models.FileUpload
	for _, file := range files {
		if file.TenantID == tenantID {
			tenantFiles = append(tenantFiles, file)
		}
	}

	return dto.ToFileUploadResponses(tenantFiles), nil
}

// ListFilesByEntity lists files related to an entity
func (s *fileUploadService) ListFilesByEntity(ctx context.Context, entityType string, entityID uuid.UUID, tenantID uuid.UUID) ([]*dto.FileUploadResponse, error) {
	files, err := s.repos.FileUpload.FindByEntity(ctx, entityType, entityID)
	if err != nil {
		s.logger.Error("failed to list files by entity", "entity_type", entityType, "entity_id", entityID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list files", err)
	}

	// Filter by tenant
	var tenantFiles []*models.FileUpload
	for _, file := range files {
		if file.TenantID == tenantID {
			tenantFiles = append(tenantFiles, file)
		}
	}

	return dto.ToFileUploadResponses(tenantFiles), nil
}

// ListFilesByType lists files by type
func (s *fileUploadService) ListFilesByType(ctx context.Context, tenantID uuid.UUID, fileType models.FileType) ([]*dto.FileUploadResponse, error) {
	files, err := s.repos.FileUpload.FindByFileType(ctx, tenantID, fileType)
	if err != nil {
		s.logger.Error("failed to list files by type", "file_type", fileType, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list files", err)
	}

	return dto.ToFileUploadResponses(files), nil
}

// SearchFiles searches files by query
func (s *fileUploadService) SearchFiles(ctx context.Context, tenantID uuid.UUID, query string, page, pageSize int) (*dto.FileUploadListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	files, paginationResult, err := s.repos.FileUpload.Search(ctx, tenantID, query, pagination)
	if err != nil {
		s.logger.Error("failed to search files", "query", query, "error", err)
		return nil, errors.NewServiceError("SEARCH_FAILED", "Failed to search files", err)
	}

	return &dto.FileUploadListResponse{
		Files:       dto.ToFileUploadResponses(files),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetFileStats retrieves file statistics
func (s *fileUploadService) GetFileStats(ctx context.Context, tenantID uuid.UUID) (*dto.FileStatsResponse, error) {
	stats, err := s.repos.FileUpload.GetFileStats(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get file stats", "tenant_id", tenantID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get file stats", err)
	}

	// Type assert the stats
	totalFiles, _ := stats["total_files"].(int64)
	storageUsedBytes, _ := stats["storage_used_bytes"].(int64)
	storageUsedMB, _ := stats["storage_used_mb"].(float64)
	storageUsedGB, _ := stats["storage_used_gb"].(float64)
	recentUploads30d, _ := stats["recent_uploads_30d"].(int64)
	byType, _ := stats["by_type"].(map[string]int64)

	return &dto.FileStatsResponse{
		TenantID:         tenantID,
		TotalFiles:       totalFiles,
		StorageUsedBytes: storageUsedBytes,
		StorageUsedMB:    storageUsedMB,
		StorageUsedGB:    storageUsedGB,
		FilesByType:      byType,
		RecentUploads30d: recentUploads30d,
	}, nil
}

// GetStorageUsage returns total storage used by a tenant
func (s *fileUploadService) GetStorageUsage(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	usage, err := s.repos.FileUpload.GetStorageUsage(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get storage usage", "tenant_id", tenantID, "error", err)
		return 0, errors.NewServiceError("QUERY_FAILED", "Failed to get storage usage", err)
	}

	return usage, nil
}

// ValidateFileUpload validates file upload request
func (s *fileUploadService) ValidateFileUpload(req *dto.UploadFileRequest) error {
	// Validate file size (max 100MB)
	const maxFileSize = 100 * 1024 * 1024 // 100MB
	if req.FileSize > maxFileSize {
		return errors.NewValidationError(fmt.Sprintf("File size exceeds maximum allowed size of %d MB", maxFileSize/(1024*1024)))
	}

	// Validate file type
	validFileTypes := map[models.FileType]bool{
		models.FileTypeImage:    true,
		models.FileTypeDocument: true,
		models.FileTypeVideo:    true,
		models.FileTypeOther:    true,
	}
	if !validFileTypes[req.FileType] {
		return errors.NewValidationError(fmt.Sprintf("Invalid file type: %s", req.FileType))
	}

	// Validate MIME type for images
	if req.FileType == models.FileTypeImage {
		allowedMimeTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml"}
		if !slices.Contains(allowedMimeTypes, req.MimeType) {
			return errors.NewValidationError(fmt.Sprintf("Invalid MIME type for image: %s", req.MimeType))
		}
	}

	// Validate related entity type if provided
	if req.RelatedEntityType != "" {
		validEntityTypes := []string{"booking", "user", "review", "project", "artisan"}
		if !slices.Contains(validEntityTypes, req.RelatedEntityType) {
			return errors.NewValidationError(fmt.Sprintf("Invalid related entity type: %s", req.RelatedEntityType))
		}

		if req.RelatedEntityID == nil {
			return errors.NewValidationError("Related entity ID is required when entity type is specified")
		}
	}

	return nil
}

// CleanupOrphanedFiles cleans up orphaned files
func (s *fileUploadService) CleanupOrphanedFiles(ctx context.Context, olderThanDays int) error {
	if olderThanDays <= 0 {
		return errors.NewValidationError("olderThanDays must be greater than zero")
	}

	duration := time.Duration(olderThanDays) * 24 * time.Hour
	if err := s.repos.FileUpload.CleanupOrphanedFiles(ctx, duration); err != nil {
		s.logger.Error("failed to cleanup orphaned files", "error", err)
		return errors.NewServiceError("CLEANUP_FAILED", "Failed to cleanup orphaned files", err)
	}

	s.logger.Info("orphaned files cleanup completed", "older_than_days", olderThanDays)
	return nil
}

// UpdateFileAccess updates file access timestamp
func (s *fileUploadService) UpdateFileAccess(ctx context.Context, fileID uuid.UUID) error {
	if err := s.repos.FileUpload.UpdateFileAccess(ctx, fileID); err != nil {
		s.logger.Error("failed to update file access", "file_id", fileID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update file access", err)
	}

	return nil
}

// verifyRelatedEntity verifies that a related entity exists and belongs to the tenant
func (s *fileUploadService) verifyRelatedEntity(ctx context.Context, entityType string, entityID uuid.UUID, tenantID uuid.UUID) error {
	switch strings.ToLower(entityType) {
	case "user":
		user, err := s.repos.User.GetByID(ctx, entityID)
		if err != nil {
			return errors.NewNotFoundError("related user")
		}
		if user.TenantID != nil && *user.TenantID != tenantID {
			return errors.NewValidationError("Related user does not belong to tenant")
		}

	case "booking":
		booking, err := s.repos.Booking.GetByID(ctx, entityID)
		if err != nil {
			return errors.NewNotFoundError("related booking")
		}
		if booking.TenantID != tenantID {
			return errors.NewValidationError("Related booking does not belong to tenant")
		}

	case "project":
		project, err := s.repos.Project.GetByID(ctx, entityID)
		if err != nil {
			return errors.NewNotFoundError("related project")
		}
		if project.TenantID != tenantID {
			return errors.NewValidationError("Related project does not belong to tenant")
		}

	default:
		// For other entity types, we'll skip verification for now
		s.logger.Warn("skipping verification for entity type", "entity_type", entityType)
	}

	return nil
}
