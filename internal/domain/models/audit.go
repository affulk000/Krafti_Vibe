package models

import (
	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionLogin  AuditAction = "login"
	AuditActionLogout AuditAction = "logout"
	AuditActionExport AuditAction = "export"
)

type AuditLog struct {
	BaseModel

	// Multi-tenancy
	TenantID *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"`

	// Actor
	UserID    *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	UserEmail string     `json:"user_email" gorm:"size:255;index"`
	UserRole  UserRole   `json:"user_role" gorm:"type:varchar(50);index"`

	// Action
	Action     AuditAction `json:"action" gorm:"type:varchar(50);not null;index" validate:"required"`
	EntityType string      `json:"entity_type" gorm:"size:50;not null;index" validate:"required"` // booking, user, payment
	EntityID   uuid.UUID   `json:"entity_id" gorm:"type:uuid;not null;index" validate:"required"`

	// Details
	Description string `json:"description" gorm:"type:text"`
	OldValues   JSONB  `json:"old_values,omitempty" gorm:"type:jsonb"`
	NewValues   JSONB  `json:"new_values,omitempty" gorm:"type:jsonb"`

	// Request Info
	IPAddress string `json:"ip_address" gorm:"size:45"`
	UserAgent string `json:"user_agent" gorm:"size:500"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
