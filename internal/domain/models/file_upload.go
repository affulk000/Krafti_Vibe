package models

import "github.com/google/uuid"

type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeDocument FileType = "document"
	FileTypeVideo    FileType = "video"
	FileTypeOther    FileType = "other"
)

type FileUpload struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Uploaded By
	UploadedByID uuid.UUID `json:"uploaded_by_id" gorm:"type:uuid;not null;index" validate:"required"`

	// File Details
	FileName     string   `json:"file_name" gorm:"not null;size:255" validate:"required"`
	FileType     FileType `json:"file_type" gorm:"type:varchar(50);not null" validate:"required"`
	MimeType     string   `json:"mime_type" gorm:"size:100" validate:"required"`
	FileSize     int64    `json:"file_size" gorm:"not null" validate:"required,min=1"` // bytes
	FilePath     string   `json:"file_path" gorm:"not null;size:500" validate:"required"`
	FileURL      string   `json:"file_url" gorm:"not null;size:500" validate:"required,url"`
	ThumbnailURL string   `json:"thumbnail_url,omitempty" gorm:"size:500"`

	// Related Entity
	RelatedEntityType string     `json:"related_entity_type,omitempty" gorm:"size:50"` // booking, user, review
	RelatedEntityID   *uuid.UUID `json:"related_entity_id,omitempty" gorm:"type:uuid;index"`

	// Storage
	StorageProvider string `json:"storage_provider" gorm:"size:50;default:'s3'"` // s3, local, cloudinary
	StorageBucket   string `json:"storage_bucket,omitempty" gorm:"size:255"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	UploadedBy *User `json:"uploaded_by,omitempty" gorm:"foreignKey:UploadedByID"`
}

// Business Methods
func (f *FileUpload) IsImage() bool {
	return f.FileType == FileTypeImage
}

func (f *FileUpload) GetFileSizeKB() float64 {
	return float64(f.FileSize) / 1024
}

func (f *FileUpload) GetFileSizeMB() float64 {
	return float64(f.FileSize) / (1024 * 1024)
}
