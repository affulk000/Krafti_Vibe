package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	// Platform Roles (TenantID = NULL, IsPlatformUser = true)
	UserRolePlatformSuperAdmin UserRole = "platform_super_admin" // App owner
	UserRolePlatformAdmin      UserRole = "platform_admin"       // Platform staff
	UserRolePlatformSupport    UserRole = "platform_support"     // Support staff

	// Tenant Roles (TenantID = NOT NULL, IsPlatformUser = false)
	UserRoleTenantOwner UserRole = "tenant_owner" // Owns the organization
	UserRoleTenantAdmin UserRole = "tenant_admin" // Admin of organization
	UserRoleArtisan     UserRole = "artisan"      // Service provider
	UserRoleTeamMember  UserRole = "team_member"  // Employee/staff
	UserRoleCustomer    UserRole = "customer"     // Service consumer
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

type User struct {
	BaseModel

	// Multi-tenancy - NULLABLE for platform admins
	TenantID       *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index:idx_user_tenant_email"`
	IsPlatformUser bool       `json:"is_platform_user" gorm:"default:false;index"`

	// Authentication
	Email        string `json:"email" gorm:"not null;size:255;uniqueIndex:idx_user_email"`
	PasswordHash string `json:"-" gorm:"size:255"`
	LogtoUserID  string `json:"logto_user_id,omitempty" gorm:"uniqueIndex;size:255"`

	// Basic Info
	FirstName   string `json:"first_name" gorm:"not null;size:100"`
	LastName    string `json:"last_name" gorm:"not null;size:100"`
	PhoneNumber string `json:"phone_number,omitempty" gorm:"size:20"`
	AvatarURL   string `json:"avatar_url,omitempty" gorm:"size:500"`

	// Role & Status
	Role   UserRole   `json:"role" gorm:"type:varchar(50);not null;index"`
	Status UserStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending'"`

	// Preferences
	Timezone      string `json:"timezone" gorm:"size:50;default:'UTC'"`
	Language      string `json:"language" gorm:"size:10;default:'en'"`
	EmailVerified bool   `json:"email_verified" gorm:"default:false"`
	PhoneVerified bool   `json:"phone_verified" gorm:"default:false"`

	// Enhanced Security
	MFAEnabled          bool       `json:"mfa_enabled" gorm:"default:false"`
	MFASecret           string     `json:"-" gorm:"size:255"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled" gorm:"default:false"`
	TwoFactorSecret     string     `json:"-" gorm:"size:255"`
	BackupCodes         []string   `json:"-" gorm:"type:text[]"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastPasswordResetAt *time.Time `json:"last_password_reset_at,omitempty"`
	PasswordChangedAt   *time.Time `json:"password_changed_at,omitempty"`
	MustChangePassword  bool       `json:"must_change_password" gorm:"default:false"`
	FailedLoginAttempts int        `json:"failed_login_attempts" gorm:"default:0"`
	LastFailedLoginAt   *time.Time `json:"last_failed_login_at,omitempty"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	AccountLockedReason string     `json:"account_locked_reason,omitempty" gorm:"size:255"`

	// Session Management
	SessionToken     string     `json:"-" gorm:"size:255;index"`
	SessionExpiresAt *time.Time `json:"session_expires_at,omitempty"`
	RefreshTokens    []string   `json:"-" gorm:"type:text[]"`

	// Compliance (GDPR/CCPA)
	TermsAcceptedAt         *time.Time `json:"terms_accepted_at,omitempty"`
	TermsVersion            string     `json:"terms_version,omitempty" gorm:"size:50"`
	PrivacyPolicyAcceptedAt *time.Time `json:"privacy_policy_accepted_at,omitempty"`
	DataProcessingConsent   bool       `json:"data_processing_consent" gorm:"default:false"`
	MarketingConsent        bool       `json:"marketing_consent" gorm:"default:false"`
	DataRetentionDays       int        `json:"data_retention_days" gorm:"default:730"`
	MarkedForDeletion       bool       `json:"marked_for_deletion" gorm:"default:false;index"`
	DeletionScheduledAt     *time.Time `json:"deletion_scheduled_at,omitempty"`

	// Profiles
	ArtisanProfile  *Artisan  `json:"artisan_profile,omitempty" gorm:"foreignKey:UserID"`
	CustomerProfile *Customer `json:"customer_profile,omitempty" gorm:"foreignKey:UserID"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant           *Tenant         `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	OwnedTenants     []Tenant        `json:"owned_tenants,omitempty" gorm:"foreignKey:OwnerID"`
	AssignedTasks    []ProjectTask   `json:"assigned_tasks,omitempty" gorm:"foreignKey:AssignedToID"`
	ProjectUpdates   []ProjectUpdate `json:"project_updates,omitempty" gorm:"foreignKey:UserID"`
	Projects         []Project       `json:"projects,omitempty" gorm:"foreignKey:ArtisanID"`
	Bookings         []Booking       `json:"bookings,omitempty" gorm:"foreignKey:ArtisanID"`
	CustomerBookings []Booking       `json:"customer_bookings,omitempty" gorm:"foreignKey:CustomerID"`
	Reviews          []Review        `json:"reviews,omitempty" gorm:"foreignKey:ArtisanID"`
	CustomerReviews  []Review        `json:"customer_reviews,omitempty" gorm:"foreignKey:CustomerID"`
}

// Business Methods
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u *User) IsPlatformAdmin() bool {
	return u.IsPlatformUser && (u.Role == UserRolePlatformSuperAdmin || u.Role == UserRolePlatformAdmin)
}

func (u *User) IsPlatformSuperAdmin() bool {
	return u.IsPlatformUser && u.Role == UserRolePlatformSuperAdmin
}

func (u *User) IsTenantOwner() bool {
	return !u.IsPlatformUser && u.Role == UserRoleTenantOwner
}

func (u *User) IsTenantAdmin() bool {
	return !u.IsPlatformUser && u.Role == UserRoleTenantAdmin
}

func (u *User) IsArtisan() bool {
	return u.Role == UserRoleArtisan
}

func (u *User) IsCustomer() bool {
	return u.Role == UserRoleCustomer
}

func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && time.Now().Before(*u.LockedUntil)
}

func (u *User) RequiresMFA() bool {
	return u.MFAEnabled || u.TwoFactorEnabled
}

func (u *User) CanAccessTenant(tenantID uuid.UUID) bool {
	if u.IsPlatformAdmin() {
		return true // Platform admins can access all tenants
	}
	return u.TenantID != nil && *u.TenantID == tenantID
}

func (u *User) CanManageTenant(tenantID uuid.UUID) bool {
	if u.IsPlatformAdmin() {
		return true
	}
	if u.TenantID != nil && *u.TenantID == tenantID {
		return u.Role == UserRoleTenantOwner || u.Role == UserRoleTenantAdmin
	}
	return false
}

// Security Methods
func (u *User) RecordFailedLogin() {
	u.FailedLoginAttempts++
	now := time.Now()
	u.LastFailedLoginAt = &now

	if u.FailedLoginAttempts >= 5 {
		lockUntil := now.Add(15 * time.Minute)
		u.LockedUntil = &lockUntil
		u.AccountLockedReason = "Too many failed login attempts"
	}
}

func (u *User) RecordSuccessfulLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.FailedLoginAttempts = 0
	u.LastFailedLoginAt = nil
}

func (u *User) UnlockAccount() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.AccountLockedReason = ""
}

// Validate hooks for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Validate role and tenant relationship
	if u.IsPlatformUser {
		// Platform users cannot have a tenant
		if u.TenantID != nil {
			return errors.New("platform users cannot belong to a tenant")
		}
		// Validate platform role
		if u.Role != UserRolePlatformSuperAdmin &&
			u.Role != UserRolePlatformAdmin &&
			u.Role != UserRolePlatformSupport {
			return errors.New("invalid platform user role")
		}
	} else {
		// Tenant users must have a tenant (except during onboarding)
		if u.Status != UserStatusPending && u.TenantID == nil {
			return errors.New("tenant users must belong to a tenant")
		}
		// Validate tenant role
		if u.Role == UserRolePlatformSuperAdmin ||
			u.Role == UserRolePlatformAdmin ||
			u.Role == UserRolePlatformSupport {
			return errors.New("cannot assign platform role to tenant user")
		}
	}

	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Prevent changing user from platform to tenant or vice versa
	var original User
	if err := tx.First(&original, u.ID).Error; err == nil {
		if original.IsPlatformUser != u.IsPlatformUser {
			return errors.New("cannot change platform user status")
		}
	}
	return nil
}
