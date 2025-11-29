package models

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "active"
	APIKeyStatusInactive APIKeyStatus = "inactive"
	APIKeyStatusExpired  APIKeyStatus = "expired"
	APIKeyStatusRevoked  APIKeyStatus = "revoked"
)

type APIKey struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index" validate:"required"`

	// Created By
	CreatedByID uuid.UUID `json:"created_by_id" gorm:"type:uuid;not null" validate:"required"`

	// Key Details
	Name        string `json:"name" gorm:"not null;size:255" validate:"required,min=2,max=255"`
	Description string `json:"description,omitempty" gorm:"type:text"`
	KeyHash     string `json:"key_hash" gorm:"uniqueIndex;not null;size:255"` // Hashed API key
	KeyPrefix   string `json:"key_prefix" gorm:"size:10"`                     // First few chars for identification

	// Permissions
	Scopes []string `json:"scopes" gorm:"type:text[];not null" validate:"required,min=1"` // read:bookings, write:bookings

	// Status
	Status    APIKeyStatus `json:"status" gorm:"type:varchar(50);not null;default:'active'" validate:"required"`
	ExpiresAt *time.Time   `json:"expires_at,omitempty"`

	// Usage Tracking
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	UsageCount int64      `json:"usage_count" gorm:"default:0"`

	// Rate Limiting
	RateLimitPerHour int `json:"rate_limit_per_hour" gorm:"default:1000"`
	RateLimitPerDay  int `json:"rate_limit_per_day" gorm:"default:10000"`

	// Restrictions
	AllowedIPs []string `json:"allowed_ips,omitempty" gorm:"type:text[]"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant    *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	CreatedBy *User   `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID"`
}

// Business Methods
func (ak *APIKey) IsActive() bool {
	if ak.Status != APIKeyStatusActive {
		return false
	}

	if ak.ExpiresAt != nil && time.Now().After(*ak.ExpiresAt) {
		return false
	}

	return true
}

func (ak *APIKey) HasScope(scope string) bool {
	return slices.Contains(ak.Scopes, scope)
}

func (ak *APIKey) IsExpired() bool {
	return ak.ExpiresAt != nil && time.Now().After(*ak.ExpiresAt)
}
