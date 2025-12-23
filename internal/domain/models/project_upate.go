package models

import (
	"github.com/google/uuid"
)

type UpdateType string

const (
	UpdateTypeStatusChange UpdateType = "status_change"
	UpdateTypeMilestone    UpdateType = "milestone"
	UpdateTypeComment      UpdateType = "comment"
	UpdateTypeFile         UpdateType = "file"
	UpdateTypePayment      UpdateType = "payment"
	UpdateTypeSchedule     UpdateType = "schedule"
)

type ProjectUpdate struct {
	BaseModel

	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	ProjectID uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`

	Type        UpdateType `json:"type" gorm:"type:varchar(32);not null"`
	Title       string     `json:"title" gorm:"not null;size:255"`
	Description string     `json:"description,omitempty" gorm:"type:text"`

	// Visibility
	VisibleToCustomer bool `json:"visible_to_customer" gorm:"default:true"`

	// Attachments
	AttachmentURLs []string `json:"attachment_urls,omitempty" gorm:"type:text[]"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
