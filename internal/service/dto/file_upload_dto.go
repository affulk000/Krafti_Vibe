package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// File Upload Request DTOs
// ============================================================================

// UploadFileRequest represents a request to upload a file
type UploadFileRequest struct {
	FileName          string         `json:"file_name" validate:"required"`
	FileType          models.FileType `json:"file_type" validate:"required,oneof=image document video other"`
	MimeType          string         `json:"mime_type" validate:"required"`
	FileSize          int64          `json:"file_size" validate:"required,min=1"`
	FilePath          string         `json:"file_path" validate:"required"`
	FileURL           string         `json:"file_url" validate:"required,url"`
	ThumbnailURL      string         `json:"thumbnail_url,omitempty"`
	RelatedEntityType string         `json:"related_entity_type,omitempty" validate:"omitempty,oneof=booking user review project artisan"`
	RelatedEntityID   *uuid.UUID     `json:"related_entity_id,omitempty"`
	StorageProvider   string         `json:"storage_provider,omitempty" validate:"omitempty,oneof=s3 local cloudinary"`
	StorageBucket     string         `json:"storage_bucket,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

// UpdateFileRequest represents a request to update file metadata
type UpdateFileRequest struct {
	FileName          *string        `json:"file_name,omitempty"`
	RelatedEntityType *string        `json:"related_entity_type,omitempty" validate:"omitempty,oneof=booking user review project artisan"`
	RelatedEntityID   *uuid.UUID     `json:"related_entity_id,omitempty"`
	ThumbnailURL      *string        `json:"thumbnail_url,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

// FileUploadFilter represents filters for file queries
type FileUploadFilter struct {
	TenantID          uuid.UUID       `json:"tenant_id" validate:"required"`
	UploadedByID      *uuid.UUID      `json:"uploaded_by_id,omitempty"`
	FileType          *models.FileType `json:"file_type,omitempty"`
	RelatedEntityType *string         `json:"related_entity_type,omitempty"`
	RelatedEntityID   *uuid.UUID      `json:"related_entity_id,omitempty"`
	MinSize           *int64          `json:"min_size,omitempty"`
	MaxSize           *int64          `json:"max_size,omitempty"`
	SearchQuery       string          `json:"search_query,omitempty"`
	Page              int             `json:"page"`
	PageSize          int             `json:"page_size"`
}

// ============================================================================
// File Upload Response DTOs
// ============================================================================

// FileUploadResponse represents a file upload
type FileUploadResponse struct {
	ID                uuid.UUID       `json:"id"`
	TenantID          uuid.UUID       `json:"tenant_id"`
	UploadedByID      uuid.UUID       `json:"uploaded_by_id"`
	FileName          string          `json:"file_name"`
	FileType          models.FileType `json:"file_type"`
	MimeType          string          `json:"mime_type"`
	FileSize          int64           `json:"file_size"`
	FileSizeKB        float64         `json:"file_size_kb"`
	FileSizeMB        float64         `json:"file_size_mb"`
	FilePath          string          `json:"file_path"`
	FileURL           string          `json:"file_url"`
	ThumbnailURL      string          `json:"thumbnail_url,omitempty"`
	RelatedEntityType string          `json:"related_entity_type,omitempty"`
	RelatedEntityID   *uuid.UUID      `json:"related_entity_id,omitempty"`
	StorageProvider   string          `json:"storage_provider"`
	StorageBucket     string          `json:"storage_bucket,omitempty"`
	Metadata          models.JSONB    `json:"metadata,omitempty"`
	UploadedBy        *UserSummary    `json:"uploaded_by,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// FileUploadListResponse represents a paginated list of file uploads
type FileUploadListResponse struct {
	Files       []*FileUploadResponse `json:"files"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"page_size"`
	TotalItems  int64                 `json:"total_items"`
	TotalPages  int                   `json:"total_pages"`
	HasNext     bool                  `json:"has_next"`
	HasPrevious bool                  `json:"has_previous"`
}

// FileStatsResponse represents file statistics
type FileStatsResponse struct {
	TenantID           uuid.UUID              `json:"tenant_id"`
	TotalFiles         int64                  `json:"total_files"`
	StorageUsedBytes   int64                  `json:"storage_used_bytes"`
	StorageUsedMB      float64                `json:"storage_used_mb"`
	StorageUsedGB      float64                `json:"storage_used_gb"`
	FilesByType        map[string]int64       `json:"files_by_type"`
	RecentUploads30d   int64                  `json:"recent_uploads_30d"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToFileUploadResponse converts a FileUpload model to FileUploadResponse DTO
func ToFileUploadResponse(file *models.FileUpload) *FileUploadResponse {
	if file == nil {
		return nil
	}

	resp := &FileUploadResponse{
		ID:                file.ID,
		TenantID:          file.TenantID,
		UploadedByID:      file.UploadedByID,
		FileName:          file.FileName,
		FileType:          file.FileType,
		MimeType:          file.MimeType,
		FileSize:          file.FileSize,
		FileSizeKB:        file.GetFileSizeKB(),
		FileSizeMB:        file.GetFileSizeMB(),
		FilePath:          file.FilePath,
		FileURL:           file.FileURL,
		ThumbnailURL:      file.ThumbnailURL,
		RelatedEntityType: file.RelatedEntityType,
		RelatedEntityID:   file.RelatedEntityID,
		StorageProvider:   file.StorageProvider,
		StorageBucket:     file.StorageBucket,
		Metadata:          file.Metadata,
		CreatedAt:         file.CreatedAt,
		UpdatedAt:         file.UpdatedAt,
	}

	// Add uploader if available
	if file.UploadedBy != nil {
		resp.UploadedBy = &UserSummary{
			ID:        file.UploadedBy.ID,
			FirstName: file.UploadedBy.FirstName,
			LastName:  file.UploadedBy.LastName,
			Email:     file.UploadedBy.Email,
			AvatarURL: file.UploadedBy.AvatarURL,
		}
	}

	return resp
}

// ToFileUploadResponses converts multiple FileUpload models to DTOs
func ToFileUploadResponses(files []*models.FileUpload) []*FileUploadResponse {
	if files == nil {
		return nil
	}

	responses := make([]*FileUploadResponse, len(files))
	for i, file := range files {
		responses[i] = ToFileUploadResponse(file)
	}
	return responses
}
