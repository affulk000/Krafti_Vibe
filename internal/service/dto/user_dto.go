package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"

	"github.com/google/uuid"
)

// ============================================================================
// User Request DTOs
// ============================================================================

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	TenantID       *uuid.UUID        `json:"tenant_id,omitempty"`
	IsPlatformUser bool              `json:"is_platform_user"`
	Email          string            `json:"email" validate:"required,email,max=255"`
	Password       string            `json:"password,omitempty" validate:"omitempty,min=8"`
	FirstName      string            `json:"first_name" validate:"required,max=100"`
	LastName       string            `json:"last_name" validate:"required,max=100"`
	PhoneNumber    string            `json:"phone_number,omitempty" validate:"omitempty,max=20"`
	AvatarURL      string            `json:"avatar_url,omitempty" validate:"omitempty,max=500,url"`
	Role           models.UserRole   `json:"role" validate:"required"`
	Status         models.UserStatus `json:"status,omitempty"`
	Timezone       string            `json:"timezone,omitempty" validate:"omitempty,max=50"`
	Language       string            `json:"language,omitempty" validate:"omitempty,max=10"`
	ZitadelUserID  string            `json:"zitadel_user_id,omitempty" validate:"omitempty,max=255"`
	Metadata       map[string]any    `json:"metadata,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FirstName     *string            `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName      *string            `json:"last_name,omitempty" validate:"omitempty,max=100"`
	PhoneNumber   *string            `json:"phone_number,omitempty" validate:"omitempty,max=20"`
	AvatarURL     *string            `json:"avatar_url,omitempty" validate:"omitempty,max=500,url"`
	Role          *models.UserRole   `json:"role,omitempty"`
	Status        *models.UserStatus `json:"status,omitempty"`
	Timezone      *string            `json:"timezone,omitempty" validate:"omitempty,max=50"`
	Language      *string            `json:"language,omitempty" validate:"omitempty,max=10"`
	EmailVerified *bool              `json:"email_verified,omitempty"`
	PhoneVerified *bool              `json:"phone_verified,omitempty"`
	MFAEnabled    *bool              `json:"mfa_enabled,omitempty"`
	Metadata      map[string]any     `json:"metadata,omitempty"`
}

// UpdatePasswordRequest represents a request to update password
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordRequest represents a request to reset password
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmPasswordResetRequest represents a request to confirm password reset
type ConfirmPasswordResetRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UpdatePreferencesRequest represents a request to update user preferences
type UpdatePreferencesRequest struct {
	Timezone              string `json:"timezone" validate:"required,max=50"`
	Language              string `json:"language" validate:"required,max=10"`
	DataProcessingConsent *bool  `json:"data_processing_consent,omitempty"`
	MarketingConsent      *bool  `json:"marketing_consent,omitempty"`
}

// EnableMFARequest represents a request to enable MFA
type EnableMFARequest struct {
	Secret string `json:"secret" validate:"required"`
	Code   string `json:"code" validate:"required,len=6"`
}

// VerifyMFARequest represents a request to verify MFA code
type VerifyMFARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// SuspendUserRequest represents a request to suspend a user
type SuspendUserRequest struct {
	Reason string `json:"reason" validate:"required,max=255"`
}

// UserFilter represents filters for user queries
type UserFilter struct {
	TenantID          *uuid.UUID          `json:"tenant_id,omitempty"`
	Roles             []models.UserRole   `json:"roles,omitempty"`
	Statuses          []models.UserStatus `json:"statuses,omitempty"`
	IsPlatformUser    *bool               `json:"is_platform_user,omitempty"`
	EmailVerified     *bool               `json:"email_verified,omitempty"`
	PhoneVerified     *bool               `json:"phone_verified,omitempty"`
	MFAEnabled        *bool               `json:"mfa_enabled,omitempty"`
	IsLocked          *bool               `json:"is_locked,omitempty"`
	MarkedForDeletion *bool               `json:"marked_for_deletion,omitempty"`
	SearchQuery       string              `json:"search_query,omitempty"`
	CreatedAfter      *time.Time          `json:"created_after,omitempty"`
	CreatedBefore     *time.Time          `json:"created_before,omitempty"`
	LastLoginAfter    *time.Time          `json:"last_login_after,omitempty"`
	LastLoginBefore   *time.Time          `json:"last_login_before,omitempty"`
	Page              int                 `json:"page"`
	PageSize          int                 `json:"page_size"`
}

// ============================================================================
// User Response DTOs
// ============================================================================

// UserResponse represents a user with basic information
type UserResponse struct {
	ID                      uuid.UUID         `json:"id"`
	TenantID                *uuid.UUID        `json:"tenant_id,omitempty"`
	IsPlatformUser          bool              `json:"is_platform_user"`
	Email                   string            `json:"email"`
	FirstName               string            `json:"first_name"`
	LastName                string            `json:"last_name"`
	FullName                string            `json:"full_name"`
	PhoneNumber             string            `json:"phone_number,omitempty"`
	AvatarURL               string            `json:"avatar_url,omitempty"`
	Role                    models.UserRole   `json:"role"`
	Status                  models.UserStatus `json:"status"`
	Timezone                string            `json:"timezone"`
	Language                string            `json:"language"`
	EmailVerified           bool              `json:"email_verified"`
	PhoneVerified           bool              `json:"phone_verified"`
	MFAEnabled              bool              `json:"mfa_enabled"`
	TwoFactorEnabled        bool              `json:"two_factor_enabled"`
	LastLoginAt             *time.Time        `json:"last_login_at,omitempty"`
	TermsAcceptedAt         *time.Time        `json:"terms_accepted_at,omitempty"`
	PrivacyPolicyAcceptedAt *time.Time        `json:"privacy_policy_accepted_at,omitempty"`
	DataProcessingConsent   bool              `json:"data_processing_consent"`
	MarketingConsent        bool              `json:"marketing_consent"`
	IsActive                bool              `json:"is_active"`
	IsLocked                bool              `json:"is_locked"`
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
}

// UserDetailResponse represents detailed user information
type UserDetailResponse struct {
	UserResponse
	ZitadelUserID       string         `json:"zitadel_user_id,omitempty"`
	MustChangePassword  bool           `json:"must_change_password"`
	FailedLoginAttempts int            `json:"failed_login_attempts"`
	LastFailedLoginAt   *time.Time     `json:"last_failed_login_at,omitempty"`
	LastPasswordResetAt *time.Time     `json:"last_password_reset_at,omitempty"`
	PasswordChangedAt   *time.Time     `json:"password_changed_at,omitempty"`
	LockedUntil         *time.Time     `json:"locked_until,omitempty"`
	AccountLockedReason string         `json:"account_locked_reason,omitempty"`
	SessionExpiresAt    *time.Time     `json:"session_expires_at,omitempty"`
	TermsVersion        string         `json:"terms_version,omitempty"`
	DataRetentionDays   int            `json:"data_retention_days"`
	MarkedForDeletion   bool           `json:"marked_for_deletion"`
	DeletionScheduledAt *time.Time     `json:"deletion_scheduled_at,omitempty"`
	Metadata            map[string]any `json:"metadata,omitempty"`
}

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users       []*UserResponse `json:"users"`
	Page        int             `json:"page"`
	PageSize    int             `json:"page_size"`
	TotalItems  int64           `json:"total_items"`
	TotalPages  int             `json:"total_pages"`
	HasNext     bool            `json:"has_next"`
	HasPrevious bool            `json:"has_previous"`
}

// UserStatsResponse represents user statistics
type UserStatsResponse struct {
	TotalUsers         int64                       `json:"total_users"`
	ActiveUsers        int64                       `json:"active_users"`
	InactiveUsers      int64                       `json:"inactive_users"`
	SuspendedUsers     int64                       `json:"suspended_users"`
	PendingUsers       int64                       `json:"pending_users"`
	ByRole             map[models.UserRole]int64   `json:"by_role"`
	ByStatus           map[models.UserStatus]int64 `json:"by_status"`
	EmailVerifiedCount int64                       `json:"email_verified_count"`
	PhoneVerifiedCount int64                       `json:"phone_verified_count"`
	MFAEnabledCount    int64                       `json:"mfa_enabled_count"`
	LockedUsersCount   int64                       `json:"locked_users_count"`
	MarkedForDeletion  int64                       `json:"marked_for_deletion"`
}

// RegistrationStatsResponse represents registration statistics
type RegistrationStatsResponse struct {
	StartDate          time.Time                 `json:"start_date"`
	EndDate            time.Time                 `json:"end_date"`
	TotalRegistrations int64                     `json:"total_registrations"`
	ByRole             map[models.UserRole]int64 `json:"by_role"`
	ByDay              []DailyRegistrationData   `json:"by_day"`
	AveragePerDay      float64                   `json:"average_per_day"`
}

// DailyRegistrationData represents daily registration data
type DailyRegistrationData struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

// UserGrowthResponse represents user growth data
type UserGrowthResponse struct {
	Month      time.Time `json:"month"`
	NewUsers   int64     `json:"new_users"`
	TotalUsers int64     `json:"total_users"`
	Growth     float64   `json:"growth_percentage"`
}

// MFASetupResponse represents MFA setup information
type MFASetupResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToUserResponse converts a User model to UserResponse DTO
func ToUserResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}

	return &UserResponse{
		ID:                      user.ID,
		TenantID:                user.TenantID,
		IsPlatformUser:          user.IsPlatformUser,
		Email:                   user.Email,
		FirstName:               user.FirstName,
		LastName:                user.LastName,
		FullName:                user.FullName(),
		PhoneNumber:             user.PhoneNumber,
		AvatarURL:               user.AvatarURL,
		Role:                    user.Role,
		Status:                  user.Status,
		Timezone:                user.Timezone,
		Language:                user.Language,
		EmailVerified:           user.EmailVerified,
		PhoneVerified:           user.PhoneVerified,
		MFAEnabled:              user.MFAEnabled,
		TwoFactorEnabled:        user.TwoFactorEnabled,
		LastLoginAt:             user.LastLoginAt,
		TermsAcceptedAt:         user.TermsAcceptedAt,
		PrivacyPolicyAcceptedAt: user.PrivacyPolicyAcceptedAt,
		DataProcessingConsent:   user.DataProcessingConsent,
		MarketingConsent:        user.MarketingConsent,
		IsActive:                user.IsActive(),
		IsLocked:                user.IsLocked(),
		CreatedAt:               user.CreatedAt,
		UpdatedAt:               user.UpdatedAt,
	}
}

// ToUserDetailResponse converts a User model to UserDetailResponse DTO
func ToUserDetailResponse(user *models.User) *UserDetailResponse {
	if user == nil {
		return nil
	}

	return &UserDetailResponse{
		UserResponse:        *ToUserResponse(user),
		ZitadelUserID:       user.ZitadelUserID,
		MustChangePassword:  user.MustChangePassword,
		FailedLoginAttempts: user.FailedLoginAttempts,
		LastFailedLoginAt:   user.LastFailedLoginAt,
		LastPasswordResetAt: user.LastPasswordResetAt,
		PasswordChangedAt:   user.PasswordChangedAt,
		LockedUntil:         user.LockedUntil,
		AccountLockedReason: user.AccountLockedReason,
		SessionExpiresAt:    user.SessionExpiresAt,
		TermsVersion:        user.TermsVersion,
		DataRetentionDays:   user.DataRetentionDays,
		MarkedForDeletion:   user.MarkedForDeletion,
		DeletionScheduledAt: user.DeletionScheduledAt,
		Metadata:            user.Metadata,
	}
}

// ToUserResponses converts multiple User models to DTOs
func ToUserResponses(users []*models.User) []*UserResponse {
	if users == nil {
		return nil
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(user)
	}
	return responses
}

// ToUserStatsResponse converts repository UserStats to DTO
func ToUserStatsResponse(stats repository.UserStats) *UserStatsResponse {
	return &UserStatsResponse{
		TotalUsers:         stats.TotalUsers,
		ActiveUsers:        stats.ActiveUsers,
		InactiveUsers:      stats.InactiveUsers,
		SuspendedUsers:     stats.SuspendedUsers,
		PendingUsers:       stats.PendingUsers,
		ByRole:             stats.ByRole,
		ByStatus:           stats.ByStatus,
		EmailVerifiedCount: stats.EmailVerifiedCount,
		PhoneVerifiedCount: stats.PhoneVerifiedCount,
		MFAEnabledCount:    stats.MFAEnabledCount,
		LockedUsersCount:   stats.LockedUsersCount,
		MarkedForDeletion:  stats.MarkedForDeletion,
	}
}

// ToRegistrationStatsResponse converts repository RegistrationStats to DTO
func ToRegistrationStatsResponse(stats repository.RegistrationStats) *RegistrationStatsResponse {
	byDay := make([]DailyRegistrationData, len(stats.ByDay))
	for i, day := range stats.ByDay {
		byDay[i] = DailyRegistrationData{
			Date:  day.Date,
			Count: day.Count,
		}
	}

	return &RegistrationStatsResponse{
		StartDate:          stats.StartDate,
		EndDate:            stats.EndDate,
		TotalRegistrations: stats.TotalRegistrations,
		ByRole:             stats.ByRole,
		ByDay:              byDay,
		AveragePerDay:      stats.AveragePerDay,
	}
}

// ToUserGrowthResponses converts repository UserGrowthData to DTOs
func ToUserGrowthResponses(data []repository.UserGrowthData) []*UserGrowthResponse {
	if data == nil {
		return nil
	}

	responses := make([]*UserGrowthResponse, len(data))
	for i, d := range data {
		responses[i] = &UserGrowthResponse{
			Month:      d.Month,
			NewUsers:   d.NewUsers,
			TotalUsers: d.TotalUsers,
			Growth:     d.Growth,
		}
	}
	return responses
}
