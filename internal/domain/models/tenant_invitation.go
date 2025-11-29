package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TenantInvitation struct {
	BaseModel
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Email      string     `json:"email" gorm:"not null;size:255;index"`
	Role       UserRole   `json:"role" gorm:"type:varchar(50);not null"`
	Token      string     `json:"token" gorm:"uniqueIndex;not null;size:255"`
	ExpiresAt  time.Time  `json:"expires_at" gorm:"not null"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	InvitedBy  uuid.UUID  `json:"invited_by" gorm:"type:uuid;not null"`

	Tenant  *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Inviter *User   `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

func (ti *TenantInvitation) IsExpired() bool {
	return time.Now().After(ti.ExpiresAt)
}

func (ti *TenantInvitation) IsAccepted() bool {
	return ti.AcceptedAt != nil
}

type TenantUsageTracking struct {
	BaseModel
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex:idx_tenant_usage_date"`
	Date     time.Time `json:"date" gorm:"not null;uniqueIndex:idx_tenant_usage_date;index"`

	// API Usage
	APICallsCount int64 `json:"api_calls_count" gorm:"default:0"`
	APICallsLimit int64 `json:"api_calls_limit" gorm:"default:10000"`

	// Storage
	StorageUsedGB   int64 `json:"storage_used_gb" gorm:"default:0"`
	BandwidthUsedGB int64 `json:"bandwidth_used_gb" gorm:"default:0"`

	// Features
	BookingsCreated int `json:"bookings_created" gorm:"default:0"`
	ProjectsCreated int `json:"projects_created" gorm:"default:0"`
	SMSSent         int `json:"sms_sent" gorm:"default:0"`
	EmailsSent      int `json:"emails_sent" gorm:"default:0"`

	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

func (tut *TenantUsageTracking) CanMakeAPICall() bool {
	return tut.APICallsCount < tut.APICallsLimit
}

func (tut *TenantUsageTracking) IncrementAPICall() error {
	if !tut.CanMakeAPICall() {
		return errors.New("API rate limit exceeded for today")
	}
	tut.APICallsCount++
	return nil
}

type DataExportRequest struct {
	BaseModel
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	RequestedBy uuid.UUID  `json:"requested_by" gorm:"type:uuid;not null"`
	ExportType  string     `json:"export_type" gorm:"size:50;not null"` // full, partial, gdpr
	Status      string     `json:"status" gorm:"size:50;not null;default:'pending'"`
	FileURL     string     `json:"file_url,omitempty" gorm:"size:500"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	Tenant    *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Requester *User   `json:"requester,omitempty" gorm:"foreignKey:RequestedBy"`
}
