package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines the interface for user service operations
type UserService interface {
	// CRUD Operations
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserDetailResponse, error)
	GetUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) (*dto.UserDetailResponse, error)
	GetUserByEmail(ctx context.Context, email string, tenantID *uuid.UUID) (*dto.UserResponse, error)
	GetUserByLogtoID(ctx context.Context, ZitadelUserID string) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserDetailResponse, error)
	DeleteUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error

	// User Queries
	ListUsers(ctx context.Context, filter *dto.UserFilter, requestingUserID uuid.UUID) (*dto.UserListResponse, error)
	SearchUsers(ctx context.Context, query string, tenantID *uuid.UUID, page, pageSize int) (*dto.UserListResponse, error)
	GetUsersByRole(ctx context.Context, tenantID *uuid.UUID, role models.UserRole, page, pageSize int) (*dto.UserListResponse, error)
	GetActiveUsers(ctx context.Context, tenantID *uuid.UUID, page, pageSize int) (*dto.UserListResponse, error)
	GetRecentlyActive(ctx context.Context, tenantID *uuid.UUID, hours int, page, pageSize int) (*dto.UserListResponse, error)

	// Authentication & Security
	UpdatePassword(ctx context.Context, userID uuid.UUID, req *dto.UpdatePasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
	ConfirmPasswordReset(ctx context.Context, req *dto.ConfirmPasswordResetRequest) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
	VerifyPhone(ctx context.Context, userID uuid.UUID) error
	RecordLogin(ctx context.Context, userID uuid.UUID) error
	RecordFailedLogin(ctx context.Context, userID uuid.UUID) error
	UnlockUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error
	GetLockedUsers(ctx context.Context, requestingUserID uuid.UUID) ([]*dto.UserResponse, error)

	// MFA Management
	SetupMFA(ctx context.Context, userID uuid.UUID) (*dto.MFASetupResponse, error)
	EnableMFA(ctx context.Context, userID uuid.UUID, req *dto.EnableMFARequest) error
	DisableMFA(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error
	VerifyMFA(ctx context.Context, userID uuid.UUID, req *dto.VerifyMFARequest) error

	// Role & Status Management
	UpdateRole(ctx context.Context, userID uuid.UUID, newRole models.UserRole, requestingUserID uuid.UUID) error
	UpdateStatus(ctx context.Context, userID uuid.UUID, newStatus models.UserStatus, requestingUserID uuid.UUID) error
	ActivateUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error
	DeactivateUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error
	SuspendUser(ctx context.Context, userID uuid.UUID, req *dto.SuspendUserRequest, requestingUserID uuid.UUID) error

	// Profile Management
	UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error
	UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
	UpdatePreferences(ctx context.Context, userID uuid.UUID, req *dto.UpdatePreferencesRequest) error

	// Compliance & GDPR
	AcceptTerms(ctx context.Context, userID uuid.UUID, version string) error
	AcceptPrivacyPolicy(ctx context.Context, userID uuid.UUID) error
	UpdateConsent(ctx context.Context, userID uuid.UUID, dataProcessing, marketing bool) error
	MarkForDeletion(ctx context.Context, userID uuid.UUID, scheduledDate time.Time) error
	GetUsersMarkedForDeletion(ctx context.Context, requestingUserID uuid.UUID) ([]*dto.UserResponse, error)
	PermanentlyDeleteUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error

	// Analytics
	GetUserStats(ctx context.Context, tenantID *uuid.UUID, requestingUserID uuid.UUID) (*dto.UserStatsResponse, error)
	GetRegistrationStats(ctx context.Context, startDate, endDate time.Time, requestingUserID uuid.UUID) (*dto.RegistrationStatsResponse, error)
	GetUserGrowth(ctx context.Context, tenantID *uuid.UUID, months int, requestingUserID uuid.UUID) ([]*dto.UserGrowthResponse, error)

	// Webhook Operations (Internal - bypass permission checks)
	CreateUserFromWebhook(ctx context.Context, user *models.User) error
	UpdateUserFromWebhook(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error
	DeleteUserFromWebhook(ctx context.Context, userID uuid.UUID) error
	UpdateUserStatusFromWebhook(ctx context.Context, userID uuid.UUID, status models.UserStatus) error
}

// userService implements UserService
type userService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewUserService creates a new user service
func NewUserService(repos *repository.Repositories, logger log.AllLogger) UserService {
	return &userService{
		repos:  repos,
		logger: logger,
	}
}

// ============================================================================
// CRUD Operations
// ============================================================================

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserDetailResponse, error) {
	s.logger.Info("creating user", "email", req.Email, "role", req.Role)

	// Validate tenant requirement for non-platform users
	if !req.IsPlatformUser && req.TenantID == nil {
		return nil, errors.NewValidationError("tenant_id is required for non-platform users")
	}

	// Check if email already exists
	existingUser, err := s.repos.User.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.NewValidationError("user with this email already exists")
	}

	// Hash password if provided
	var passwordHash string
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.NewServiceError("PASSWORD_HASH_FAILED", "Failed to hash password", err)
		}
		passwordHash = string(hash)
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = models.UserStatusPending
	}

	// Create user
	user := &models.User{
		TenantID:       req.TenantID,
		IsPlatformUser: req.IsPlatformUser,
		Email:          req.Email,
		PasswordHash:   passwordHash,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		PhoneNumber:    req.PhoneNumber,
		AvatarURL:      req.AvatarURL,
		Role:           req.Role,
		Status:         status,
		Timezone:       req.Timezone,
		Language:       req.Language,
		ZitadelUserID:  req.ZitadelUserID,
		Metadata:       req.Metadata,
	}

	// Set defaults
	if user.Timezone == "" {
		user.Timezone = "UTC"
	}
	if user.Language == "" {
		user.Language = "en"
	}

	if err := s.repos.User.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "Failed to create user", err)
	}

	s.logger.Info("user created successfully", "user_id", user.ID)
	return dto.ToUserDetailResponse(user), nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) (*dto.UserDetailResponse, error) {
	s.logger.Info("getting user", "user_id", userID, "requesting_user_id", requestingUserID)

	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	// M2M tokens (requestingUserID is empty) have already been validated by scope checks
	// They bypass user-level permission checks
	if requestingUserID == uuid.Nil {
		s.logger.Info("M2M token detected, bypassing user permission checks", "user_id", userID)
		return dto.ToUserDetailResponse(user), nil
	}

	// Check access permissions for user tokens
	requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
	if err != nil {
		return nil, errors.NewNotFoundError("requesting user")
	}

	// Users can access their own data
	if userID == requestingUserID {
		return dto.ToUserDetailResponse(user), nil
	}

	// Platform admins can access all users
	if requestingUser.IsPlatformAdmin() {
		return dto.ToUserDetailResponse(user), nil
	}

	// Tenant admins can access users in their tenant
	if requestingUser.IsTenantAdmin() || requestingUser.IsTenantOwner() {
		if user.TenantID != nil && requestingUser.TenantID != nil && *user.TenantID == *requestingUser.TenantID {
			return dto.ToUserDetailResponse(user), nil
		}
	}

	return nil, errors.NewValidationError("You don't have permission to view this user")
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(ctx context.Context, email string, tenantID *uuid.UUID) (*dto.UserResponse, error) {
	s.logger.Info("getting user by email", "email", email)

	var user *models.User
	var err error

	if tenantID != nil {
		user, err = s.repos.User.GetByEmailWithTenant(ctx, email, *tenantID)
	} else {
		user, err = s.repos.User.GetByEmail(ctx, email)
	}

	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	return dto.ToUserResponse(user), nil
}

// GetUserByLogtoID retrieves a user by Logto user ID
func (s *userService) GetUserByLogtoID(ctx context.Context, ZitadelUserID string) (*dto.UserResponse, error) {
	s.logger.Info("getting user by logto id", "logto_user_id", ZitadelUserID)

	user, err := s.repos.User.GetByZitadelID(ctx, ZitadelUserID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	return dto.ToUserResponse(user), nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserDetailResponse, error) {
	s.logger.Info("updating user", "user_id", userID, "requesting_user_id", requestingUserID)

	// Get existing user
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Check permissions
	if err := s.checkUserUpdatePermission(ctx, user, requestingUserID); err != nil {
		return nil, err
	}

	// Update fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}
	if req.Language != nil {
		user.Language = *req.Language
	}
	if req.EmailVerified != nil {
		user.EmailVerified = *req.EmailVerified
	}
	if req.PhoneVerified != nil {
		user.PhoneVerified = *req.PhoneVerified
	}
	if req.MFAEnabled != nil {
		user.MFAEnabled = *req.MFAEnabled
	}

	// Role and status updates require admin permissions
	// M2M tokens can update roles and status (already validated by scopes)
	if requestingUserID == uuid.Nil {
		// M2M token - allow role and status updates
		if req.Role != nil {
			user.Role = *req.Role
		}
		if req.Status != nil {
			user.Status = *req.Status
		}
	} else {
		// User token - check permissions
		requestingUser, _ := s.repos.User.GetByID(ctx, requestingUserID)
		if requestingUser != nil && (requestingUser.IsPlatformAdmin() || requestingUser.IsTenantAdmin() || requestingUser.IsTenantOwner()) {
			if req.Role != nil {
				user.Role = *req.Role
			}
			if req.Status != nil {
				user.Status = *req.Status
			}
		}
	}

	// Update metadata
	if req.Metadata != nil {
		if user.Metadata == nil {
			user.Metadata = make(models.JSONB)
		}
		maps.Copy(user.Metadata, req.Metadata)
	}

	if err := s.repos.User.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user", "user_id", userID, "error", err)
		return nil, errors.NewServiceError("UPDATE_FAILED", "Failed to update user", err)
	}

	s.logger.Info("user updated successfully", "user_id", userID)
	return dto.ToUserDetailResponse(user), nil
}

// DeleteUser soft deletes a user
func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error {
	s.logger.Info("deleting user", "user_id", userID, "requesting_user_id", requestingUserID)

	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return errors.NewNotFoundError("requesting user")
		}

		// Platform admins can delete any user
		if !requestingUser.IsPlatformAdmin() {
			// Tenant admins can only delete users in their tenant
			if !(requestingUser.IsTenantAdmin() || requestingUser.IsTenantOwner()) {
				return errors.NewValidationError("You don't have permission to delete users")
			}

			if user.TenantID == nil || requestingUser.TenantID == nil || *user.TenantID != *requestingUser.TenantID {
				return errors.NewValidationError("You can only delete users in your tenant")
			}

			// Can't delete tenant owners
			if user.IsTenantOwner() {
				return errors.NewValidationError("Cannot delete tenant owner")
			}
		}
	}

	if err := s.repos.User.Delete(ctx, userID); err != nil {
		s.logger.Error("failed to delete user", "user_id", userID, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete user", err)
	}

	s.logger.Info("user deleted successfully", "user_id", userID)
	return nil
}

// ============================================================================
// User Queries
// ============================================================================

// ListUsers retrieves a paginated list of users with filters
func (s *userService) ListUsers(ctx context.Context, filter *dto.UserFilter, requestingUserID uuid.UUID) (*dto.UserListResponse, error) {
	s.logger.Info("listing users", "requesting_user_id", requestingUserID)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, reqErr := s.repos.User.GetByID(ctx, requestingUserID)
		if reqErr != nil {
			return nil, errors.NewNotFoundError("requesting user")
		}

		// Platform admins can see all users
		// Tenant admins can only see users in their tenant
		if !requestingUser.IsPlatformAdmin() {
			if requestingUser.TenantID == nil {
				return nil, errors.NewValidationError("Invalid user tenant")
			}
			filter.TenantID = requestingUser.TenantID
		}
	}

	// Set defaults
	page := max(1, filter.Page)
	pageSize := min(100, max(1, filter.PageSize))

	// Build repository filters
	repoFilters := repository.UserFilters{
		Roles:             filter.Roles,
		Statuses:          filter.Statuses,
		IsPlatformUser:    filter.IsPlatformUser,
		EmailVerified:     filter.EmailVerified,
		PhoneVerified:     filter.PhoneVerified,
		MFAEnabled:        filter.MFAEnabled,
		IsLocked:          filter.IsLocked,
		MarkedForDeletion: filter.MarkedForDeletion,
		CreatedAfter:      filter.CreatedAfter,
		CreatedBefore:     filter.CreatedBefore,
		LastLoginAfter:    filter.LastLoginAfter,
		LastLoginBefore:   filter.LastLoginBefore,
	}

	if filter.TenantID != nil {
		repoFilters.TenantIDs = []uuid.UUID{*filter.TenantID}
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	// Use search if query provided
	var users []*models.User
	var paginationResult repository.PaginationResult
	var err error

	if filter.SearchQuery != "" {
		users, paginationResult, err = s.repos.User.Search(ctx, filter.SearchQuery, filter.TenantID, pagination)
	} else {
		users, paginationResult, err = s.repos.User.FindByFilters(ctx, repoFilters, pagination)
	}

	if err != nil {
		s.logger.Error("failed to list users", "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list users", err)
	}

	return &dto.UserListResponse{
		Users:       dto.ToUserResponses(users),
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// SearchUsers searches for users
func (s *userService) SearchUsers(ctx context.Context, query string, tenantID *uuid.UUID, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Info("searching users", "query", query)

	page = max(1, page)
	pageSize = min(100, max(1, pageSize))

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	users, paginationResult, err := s.repos.User.Search(ctx, query, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to search users", "error", err)
		return nil, errors.NewServiceError("SEARCH_FAILED", "Failed to search users", err)
	}

	return &dto.UserListResponse{
		Users:       dto.ToUserResponses(users),
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetUsersByRole retrieves users by role
func (s *userService) GetUsersByRole(ctx context.Context, tenantID *uuid.UUID, role models.UserRole, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Info("getting users by role", "role", role)

	page = max(1, page)
	pageSize = min(100, max(1, pageSize))

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	var users []*models.User
	var paginationResult repository.PaginationResult
	var err error

	if tenantID != nil {
		users, paginationResult, err = s.repos.User.GetTenantUsersByRole(ctx, *tenantID, role, pagination)
	} else {
		users, paginationResult, err = s.repos.User.GetByRole(ctx, role, pagination)
	}

	if err != nil {
		s.logger.Error("failed to get users by role", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get users by role", err)
	}

	return &dto.UserListResponse{
		Users:       dto.ToUserResponses(users),
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetActiveUsers retrieves active users
func (s *userService) GetActiveUsers(ctx context.Context, tenantID *uuid.UUID, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Info("getting active users")

	page = max(1, page)
	pageSize = min(100, max(1, pageSize))

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	users, paginationResult, err := s.repos.User.GetActiveUsers(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to get active users", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get active users", err)
	}

	return &dto.UserListResponse{
		Users:       dto.ToUserResponses(users),
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetRecentlyActive retrieves recently active users
func (s *userService) GetRecentlyActive(ctx context.Context, tenantID *uuid.UUID, hours int, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Info("getting recently active users", "hours", hours)

	if hours <= 0 {
		hours = 24
	}

	page = max(1, page)
	pageSize = min(100, max(1, pageSize))

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	users, paginationResult, err := s.repos.User.GetRecentlyActive(ctx, tenantID, hours, pagination)
	if err != nil {
		s.logger.Error("failed to get recently active users", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get recently active users", err)
	}

	return &dto.UserListResponse{
		Users:       dto.ToUserResponses(users),
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ============================================================================
// Authentication & Security
// ============================================================================

// UpdatePassword updates a user's password
func (s *userService) UpdatePassword(ctx context.Context, userID uuid.UUID, req *dto.UpdatePasswordRequest) error {
	s.logger.Info("updating password", "user_id", userID)

	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.NewValidationError("Current password is incorrect")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewServiceError("PASSWORD_HASH_FAILED", "Failed to hash password", err)
	}

	if err := s.repos.User.UpdatePassword(ctx, userID, string(hash)); err != nil {
		s.logger.Error("failed to update password", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update password", err)
	}

	s.logger.Info("password updated successfully", "user_id", userID)
	return nil
}

// ResetPassword initiates password reset
func (s *userService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	s.logger.Info("initiating password reset", "email", req.Email)

	// Check if user exists
	user, err := s.repos.User.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if user exists or not
		s.logger.Info("password reset requested for non-existent email", "email", req.Email)
		return nil
	}

	// Generate reset token (in production, store this in a secure way with expiration)
	// For now, we'll just log that a reset was requested
	s.logger.Info("password reset requested", "user_id", user.ID, "email", req.Email)

	// TODO: Send password reset email with token

	return nil
}

// ConfirmPasswordReset confirms password reset with token
func (s *userService) ConfirmPasswordReset(ctx context.Context, req *dto.ConfirmPasswordResetRequest) error {
	s.logger.Info("confirming password reset")

	// TODO: Validate token and get user ID
	// For now, this is a placeholder

	return errors.NewServiceError("NOT_IMPLEMENTED", "Password reset confirmation not yet implemented", nil)
}

// VerifyEmail marks a user's email as verified
func (s *userService) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("verifying email", "user_id", userID)

	if err := s.repos.User.VerifyEmail(ctx, userID); err != nil {
		s.logger.Error("failed to verify email", "user_id", userID, "error", err)
		return errors.NewServiceError("VERIFY_FAILED", "Failed to verify email", err)
	}

	s.logger.Info("email verified successfully", "user_id", userID)
	return nil
}

// VerifyPhone marks a user's phone as verified
func (s *userService) VerifyPhone(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("verifying phone", "user_id", userID)

	if err := s.repos.User.VerifyPhone(ctx, userID); err != nil {
		s.logger.Error("failed to verify phone", "user_id", userID, "error", err)
		return errors.NewServiceError("VERIFY_FAILED", "Failed to verify phone", err)
	}

	s.logger.Info("phone verified successfully", "user_id", userID)
	return nil
}

// RecordLogin records a successful login
func (s *userService) RecordLogin(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("recording login", "user_id", userID)

	if err := s.repos.User.RecordLogin(ctx, userID); err != nil {
		s.logger.Error("failed to record login", "user_id", userID, "error", err)
		return errors.NewServiceError("RECORD_FAILED", "Failed to record login", err)
	}

	return nil
}

// RecordFailedLogin records a failed login attempt
func (s *userService) RecordFailedLogin(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("recording failed login", "user_id", userID)

	if err := s.repos.User.RecordFailedLogin(ctx, userID); err != nil {
		s.logger.Error("failed to record failed login", "user_id", userID, "error", err)
		return errors.NewServiceError("RECORD_FAILED", "Failed to record failed login", err)
	}

	return nil
}

// UnlockUser unlocks a locked user account
func (s *userService) UnlockUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error {
	s.logger.Info("unlocking user", "user_id", userID, "requesting_user_id", requestingUserID)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() && !requestingUser.IsTenantAdmin() && !requestingUser.IsTenantOwner() {
			return errors.NewValidationError("You don't have permission to unlock users")
		}
	}

	if err := s.repos.User.UnlockUser(ctx, userID); err != nil {
		s.logger.Error("failed to unlock user", "user_id", userID, "error", err)
		return errors.NewServiceError("UNLOCK_FAILED", "Failed to unlock user", err)
	}

	s.logger.Info("user unlocked successfully", "user_id", userID)
	return nil
}

// GetLockedUsers retrieves all locked users
func (s *userService) GetLockedUsers(ctx context.Context, requestingUserID uuid.UUID) ([]*dto.UserResponse, error) {
	s.logger.Info("getting locked users")

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return nil, errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() && !requestingUser.IsTenantAdmin() && !requestingUser.IsTenantOwner() {
			return nil, errors.NewValidationError("You don't have permission to view locked users")
		}
	}

	users, err := s.repos.User.GetLockedUsers(ctx)
	if err != nil {
		s.logger.Error("failed to get locked users", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get locked users", err)
	}

	return dto.ToUserResponses(users), nil
}

// ============================================================================
// MFA Management
// ============================================================================

// SetupMFA generates MFA secret and QR code
func (s *userService) SetupMFA(ctx context.Context, userID uuid.UUID) (*dto.MFASetupResponse, error) {
	s.logger.Info("setting up MFA", "user_id", userID)

	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	if user.MFAEnabled {
		return nil, errors.NewValidationError("MFA is already enabled for this user")
	}

	// Generate secret
	secret, err := generateMFASecret()
	if err != nil {
		return nil, errors.NewServiceError("MFA_SETUP_FAILED", "Failed to generate MFA secret", err)
	}

	// Generate backup codes
	backupCodes := generateBackupCodes(10)

	// Generate QR code URL (otpauth:// URL)
	qrCodeURL := fmt.Sprintf("otpauth://totp/KraftiVibe:%s?secret=%s&issuer=KraftiVibe", user.Email, secret)

	return &dto.MFASetupResponse{
		Secret:      secret,
		QRCodeURL:   qrCodeURL,
		BackupCodes: backupCodes,
	}, nil
}

// EnableMFA enables MFA for a user
func (s *userService) EnableMFA(ctx context.Context, userID uuid.UUID, req *dto.EnableMFARequest) error {
	s.logger.Info("enabling MFA", "user_id", userID)

	// TODO: Verify the MFA code before enabling
	// For now, we'll just enable it

	if err := s.repos.User.EnableMFA(ctx, userID, req.Secret); err != nil {
		s.logger.Error("failed to enable MFA", "user_id", userID, "error", err)
		return errors.NewServiceError("MFA_ENABLE_FAILED", "Failed to enable MFA", err)
	}

	s.logger.Info("MFA enabled successfully", "user_id", userID)
	return nil
}

// DisableMFA disables MFA for a user
func (s *userService) DisableMFA(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error {
	s.logger.Info("disabling MFA", "user_id", userID)

	// M2M tokens bypass permission checks (already validated by scopes)
	// Users can disable their own MFA, or admins can disable it for others
	if requestingUserID != uuid.Nil && userID != requestingUserID {
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() && !requestingUser.IsTenantAdmin() && !requestingUser.IsTenantOwner() {
			return errors.NewValidationError("You don't have permission to disable MFA for other users")
		}
	}

	if err := s.repos.User.DisableMFA(ctx, userID); err != nil {
		s.logger.Error("failed to disable MFA", "user_id", userID, "error", err)
		return errors.NewServiceError("MFA_DISABLE_FAILED", "Failed to disable MFA", err)
	}

	s.logger.Info("MFA disabled successfully", "user_id", userID)
	return nil
}

// VerifyMFA verifies an MFA code
func (s *userService) VerifyMFA(ctx context.Context, userID uuid.UUID, req *dto.VerifyMFARequest) error {
	s.logger.Info("verifying MFA", "user_id", userID)

	// TODO: Implement actual MFA code verification
	// This would use a TOTP library to verify the code

	return nil
}

// ============================================================================
// Role & Status Management
// ============================================================================

// UpdateRole updates a user's role
func (s *userService) UpdateRole(ctx context.Context, userID uuid.UUID, newRole models.UserRole, requestingUserID uuid.UUID) error {
	s.logger.Info("updating user role", "user_id", userID, "new_role", newRole)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() && !requestingUser.IsTenantAdmin() && !requestingUser.IsTenantOwner() {
			return errors.NewValidationError("You don't have permission to update user roles")
		}
	}

	if err := s.repos.User.UpdateRole(ctx, userID, newRole); err != nil {
		s.logger.Error("failed to update user role", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update user role", err)
	}

	s.logger.Info("user role updated successfully", "user_id", userID, "new_role", newRole)
	return nil
}

// UpdateStatus updates a user's status
func (s *userService) UpdateStatus(ctx context.Context, userID uuid.UUID, newStatus models.UserStatus, requestingUserID uuid.UUID) error {
	s.logger.Info("updating user status", "user_id", userID, "new_status", newStatus)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() && !requestingUser.IsTenantAdmin() && !requestingUser.IsTenantOwner() {
			return errors.NewValidationError("You don't have permission to update user status")
		}
	}

	if err := s.repos.User.UpdateStatus(ctx, userID, newStatus); err != nil {
		s.logger.Error("failed to update user status", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update user status", err)
	}

	s.logger.Info("user status updated successfully", "user_id", userID, "new_status", newStatus)
	return nil
}

// ActivateUser activates a user
func (s *userService) ActivateUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error {
	return s.UpdateStatus(ctx, userID, models.UserStatusActive, requestingUserID)
}

// DeactivateUser deactivates a user
func (s *userService) DeactivateUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error {
	return s.UpdateStatus(ctx, userID, models.UserStatusInactive, requestingUserID)
}

// SuspendUser suspends a user
func (s *userService) SuspendUser(ctx context.Context, userID uuid.UUID, req *dto.SuspendUserRequest, requestingUserID uuid.UUID) error {
	s.logger.Info("suspending user", "user_id", userID, "reason", req.Reason)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() && !requestingUser.IsTenantAdmin() && !requestingUser.IsTenantOwner() {
			return errors.NewValidationError("You don't have permission to suspend users")
		}
	}

	if err := s.repos.User.SuspendUser(ctx, userID, req.Reason); err != nil {
		s.logger.Error("failed to suspend user", "user_id", userID, "error", err)
		return errors.NewServiceError("SUSPEND_FAILED", "Failed to suspend user", err)
	}

	s.logger.Info("user suspended successfully", "user_id", userID)
	return nil
}

// ============================================================================
// Profile Management
// ============================================================================

// UpdateProfile updates a user's profile
func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	s.logger.Info("updating user profile", "user_id", userID)

	if err := s.repos.User.UpdateProfile(ctx, userID, updates); err != nil {
		s.logger.Error("failed to update user profile", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update user profile", err)
	}

	s.logger.Info("user profile updated successfully", "user_id", userID)
	return nil
}

// UpdateAvatar updates a user's avatar
func (s *userService) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	s.logger.Info("updating user avatar", "user_id", userID)

	if err := s.repos.User.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		s.logger.Error("failed to update user avatar", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update user avatar", err)
	}

	s.logger.Info("user avatar updated successfully", "user_id", userID)
	return nil
}

// UpdatePreferences updates a user's preferences
func (s *userService) UpdatePreferences(ctx context.Context, userID uuid.UUID, req *dto.UpdatePreferencesRequest) error {
	s.logger.Info("updating user preferences", "user_id", userID)

	if err := s.repos.User.UpdatePreferences(ctx, userID, req.Timezone, req.Language); err != nil {
		s.logger.Error("failed to update user preferences", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update user preferences", err)
	}

	// Update consent if provided
	if req.DataProcessingConsent != nil || req.MarketingConsent != nil {
		dataProcessing := false
		marketing := false

		if req.DataProcessingConsent != nil {
			dataProcessing = *req.DataProcessingConsent
		}
		if req.MarketingConsent != nil {
			marketing = *req.MarketingConsent
		}

		if err := s.repos.User.UpdateConsent(ctx, userID, dataProcessing, marketing); err != nil {
			s.logger.Error("failed to update user consent", "user_id", userID, "error", err)
			return errors.NewServiceError("UPDATE_FAILED", "Failed to update user consent", err)
		}
	}

	s.logger.Info("user preferences updated successfully", "user_id", userID)
	return nil
}

// ============================================================================
// Compliance & GDPR
// ============================================================================

// AcceptTerms records terms acceptance
func (s *userService) AcceptTerms(ctx context.Context, userID uuid.UUID, version string) error {
	s.logger.Info("accepting terms", "user_id", userID, "version", version)

	if err := s.repos.User.AcceptTerms(ctx, userID, version); err != nil {
		s.logger.Error("failed to accept terms", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to accept terms", err)
	}

	s.logger.Info("terms accepted successfully", "user_id", userID)
	return nil
}

// AcceptPrivacyPolicy records privacy policy acceptance
func (s *userService) AcceptPrivacyPolicy(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("accepting privacy policy", "user_id", userID)

	if err := s.repos.User.AcceptPrivacyPolicy(ctx, userID); err != nil {
		s.logger.Error("failed to accept privacy policy", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to accept privacy policy", err)
	}

	s.logger.Info("privacy policy accepted successfully", "user_id", userID)
	return nil
}

// UpdateConsent updates user consent settings
func (s *userService) UpdateConsent(ctx context.Context, userID uuid.UUID, dataProcessing, marketing bool) error {
	s.logger.Info("updating consent", "user_id", userID)

	if err := s.repos.User.UpdateConsent(ctx, userID, dataProcessing, marketing); err != nil {
		s.logger.Error("failed to update consent", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update consent", err)
	}

	s.logger.Info("consent updated successfully", "user_id", userID)
	return nil
}

// MarkForDeletion marks a user for deletion
func (s *userService) MarkForDeletion(ctx context.Context, userID uuid.UUID, scheduledDate time.Time) error {
	s.logger.Info("marking user for deletion", "user_id", userID, "scheduled_date", scheduledDate)

	if err := s.repos.User.MarkForDeletion(ctx, userID, scheduledDate); err != nil {
		s.logger.Error("failed to mark user for deletion", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to mark user for deletion", err)
	}

	s.logger.Info("user marked for deletion successfully", "user_id", userID)
	return nil
}

// GetUsersMarkedForDeletion retrieves users marked for deletion
func (s *userService) GetUsersMarkedForDeletion(ctx context.Context, requestingUserID uuid.UUID) ([]*dto.UserResponse, error) {
	s.logger.Info("getting users marked for deletion")

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return nil, errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() {
			return nil, errors.NewValidationError("Only platform admins can view users marked for deletion")
		}
	}

	users, err := s.repos.User.GetUsersMarkedForDeletion(ctx)
	if err != nil {
		s.logger.Error("failed to get users marked for deletion", "error", err)
		return nil, errors.NewServiceError("QUERY_FAILED", "Failed to get users marked for deletion", err)
	}

	return dto.ToUserResponses(users), nil
}

// PermanentlyDeleteUser permanently deletes a user
func (s *userService) PermanentlyDeleteUser(ctx context.Context, userID uuid.UUID, requestingUserID uuid.UUID) error {
	s.logger.Info("permanently deleting user", "user_id", userID)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() {
			return errors.NewValidationError("Only platform admins can permanently delete users")
		}
	}

	if err := s.repos.User.PermanentlyDeleteUser(ctx, userID); err != nil {
		s.logger.Error("failed to permanently delete user", "user_id", userID, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to permanently delete user", err)
	}

	s.logger.Info("user permanently deleted successfully", "user_id", userID)
	return nil
}

// ============================================================================
// Analytics
// ============================================================================

// GetUserStats retrieves user statistics
func (s *userService) GetUserStats(ctx context.Context, tenantID *uuid.UUID, requestingUserID uuid.UUID) (*dto.UserStatsResponse, error) {
	s.logger.Info("getting user stats", "tenant_id", tenantID)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return nil, errors.NewNotFoundError("requesting user")
		}

		// Platform admins can see all stats
		// Tenant admins can only see their tenant stats
		if !requestingUser.IsPlatformAdmin() {
			if requestingUser.TenantID == nil {
				return nil, errors.NewValidationError("Invalid user tenant")
			}
			tenantID = requestingUser.TenantID
		}
	}

	stats, err := s.repos.User.GetUserStats(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get user stats", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get user stats", err)
	}

	return dto.ToUserStatsResponse(stats), nil
}

// GetRegistrationStats retrieves registration statistics
func (s *userService) GetRegistrationStats(ctx context.Context, startDate, endDate time.Time, requestingUserID uuid.UUID) (*dto.RegistrationStatsResponse, error) {
	s.logger.Info("getting registration stats", "start_date", startDate, "end_date", endDate)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return nil, errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() && !requestingUser.IsTenantAdmin() && !requestingUser.IsTenantOwner() {
			return nil, errors.NewValidationError("You don't have permission to view registration stats")
		}
	}

	stats, err := s.repos.User.GetRegistrationStats(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get registration stats", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get registration stats", err)
	}

	return dto.ToRegistrationStatsResponse(stats), nil
}

// GetUserGrowth retrieves user growth data
func (s *userService) GetUserGrowth(ctx context.Context, tenantID *uuid.UUID, months int, requestingUserID uuid.UUID) ([]*dto.UserGrowthResponse, error) {
	s.logger.Info("getting user growth", "tenant_id", tenantID, "months", months)

	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID != uuid.Nil {
		// Check permissions for user tokens
		requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
		if err != nil {
			return nil, errors.NewNotFoundError("requesting user")
		}

		if !requestingUser.IsPlatformAdmin() && !requestingUser.IsTenantAdmin() && !requestingUser.IsTenantOwner() {
			return nil, errors.NewValidationError("You don't have permission to view user growth")
		}

		// Platform admins can see all growth
		// Tenant admins can only see their tenant growth
		if !requestingUser.IsPlatformAdmin() {
			if requestingUser.TenantID == nil {
				return nil, errors.NewValidationError("Invalid user tenant")
			}
			tenantID = requestingUser.TenantID
		}
	}

	if months <= 0 {
		months = 12
	}

	data, err := s.repos.User.GetUserGrowth(ctx, tenantID, months)
	if err != nil {
		s.logger.Error("failed to get user growth", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get user growth", err)
	}

	return dto.ToUserGrowthResponses(data), nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// checkUserUpdatePermission checks if requesting user can update target user
func (s *userService) checkUserUpdatePermission(ctx context.Context, targetUser *models.User, requestingUserID uuid.UUID) error {
	// M2M tokens bypass user-level permission checks (already validated by scopes)
	if requestingUserID == uuid.Nil {
		return nil
	}

	// Users can update their own profile
	if targetUser.ID == requestingUserID {
		return nil
	}

	requestingUser, err := s.repos.User.GetByID(ctx, requestingUserID)
	if err != nil {
		return errors.NewNotFoundError("requesting user")
	}

	// Platform admins can update any user
	if requestingUser.IsPlatformAdmin() {
		return nil
	}

	// Tenant admins/owners can update users in their tenant
	if requestingUser.IsTenantAdmin() || requestingUser.IsTenantOwner() {
		if targetUser.TenantID != nil && requestingUser.TenantID != nil && *targetUser.TenantID == *requestingUser.TenantID {
			// Can't modify tenant owner unless you're the owner
			if targetUser.IsTenantOwner() && !requestingUser.IsTenantOwner() {
				return errors.NewValidationError("Only tenant owner can modify the owner account")
			}
			return nil
		}
	}

	return errors.NewValidationError("You don't have permission to update this user")
}

// generateMFASecret generates a random MFA secret
func generateMFASecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// generateBackupCodes generates backup codes for MFA
func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		bytes := make([]byte, 8)
		rand.Read(bytes)
		codes[i] = fmt.Sprintf("%X-%X", bytes[0:4], bytes[4:8])
	}
	return codes
}

// ============================================================================
// Webhook Operations (Internal - bypass permission checks)
// ============================================================================

// CreateUserFromWebhook creates a user from a webhook event (no permission checks)
func (s *userService) CreateUserFromWebhook(ctx context.Context, user *models.User) error {
	s.logger.Info("creating user from webhook", "email", user.Email, "logto_user_id", user.ZitadelUserID)

	// Check if user already exists by Logto ID
	if user.ZitadelUserID != "" {
		existingUser, err := s.repos.User.GetByZitadelID(ctx, user.ZitadelUserID)
		if err == nil && existingUser != nil {
			s.logger.Warn("user already exists with this Logto ID", "logto_user_id", user.ZitadelUserID)
			return errors.NewValidationError("user with this Logto ID already exists")
		}
	}

	// Check if user already exists by email
	existingUser, err := s.repos.User.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		s.logger.Warn("user already exists with this email", "email", user.Email)
		// Update the existing user's Logto ID if it's not set
		if existingUser.ZitadelUserID == "" && user.ZitadelUserID != "" {
			s.logger.Info("updating existing user with Logto ID", "user_id", existingUser.ID, "logto_user_id", user.ZitadelUserID)
			existingUser.ZitadelUserID = user.ZitadelUserID
			if err := s.repos.User.Update(ctx, existingUser); err != nil {
				return errors.NewServiceError("UPDATE_FAILED", "Failed to update user with Logto ID", err)
			}
			return nil
		}
		return errors.NewValidationError("user with this email already exists")
	}

	// Create user (no permission checks for webhook)
	if err := s.repos.User.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user from webhook", "error", err)
		return errors.NewServiceError("CREATE_FAILED", "Failed to create user from webhook", err)
	}

	s.logger.Info("user created from webhook successfully", "user_id", user.ID, "email", user.Email)
	return nil
}

// UpdateUserFromWebhook updates a user from a webhook event (no permission checks)
func (s *userService) UpdateUserFromWebhook(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	s.logger.Info("updating user from webhook", "user_id", userID)

	// Get existing user
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Apply updates directly to the user model
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if firstName, ok := updates["first_name"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := updates["last_name"].(string); ok {
		user.LastName = lastName
	}
	if avatarURL, ok := updates["avatar_url"].(string); ok {
		user.AvatarURL = avatarURL
	}
	if phoneNumber, ok := updates["phone_number"].(string); ok {
		user.PhoneNumber = phoneNumber
	}
	if emailVerified, ok := updates["email_verified"].(bool); ok {
		user.EmailVerified = emailVerified
	}
	if phoneVerified, ok := updates["phone_verified"].(bool); ok {
		user.PhoneVerified = phoneVerified
	}
	if status, ok := updates["status"].(models.UserStatus); ok {
		user.Status = status
	}
	if metadata, ok := updates["metadata"].([]byte); ok {
		// Unmarshal the JSON bytes to JSONB (map[string]any)
		var jsonbData models.JSONB
		if err := json.Unmarshal(metadata, &jsonbData); err == nil {
			user.Metadata = jsonbData
		}
	}
	if updatedAt, ok := updates["updated_at"].(time.Time); ok {
		user.UpdatedAt = updatedAt
	}

	// Update user in database
	if err := s.repos.User.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user from webhook", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update user from webhook", err)
	}

	s.logger.Info("user updated from webhook successfully", "user_id", userID)
	return nil
}

// DeleteUserFromWebhook soft deletes a user from a webhook event (no permission checks)
func (s *userService) DeleteUserFromWebhook(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("deleting user from webhook", "user_id", userID)

	// Get user first
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Soft delete (mark as inactive and marked for deletion)
	now := time.Now()
	deletionDate := now.Add(30 * 24 * time.Hour) // Schedule deletion in 30 days

	if err := s.repos.User.MarkForDeletion(ctx, userID, deletionDate); err != nil {
		s.logger.Error("failed to mark user for deletion from webhook", "user_id", userID, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete user from webhook", err)
	}

	s.logger.Info("user deleted from webhook successfully", "user_id", userID, "user_email", user.Email)
	return nil
}

// UpdateUserStatusFromWebhook updates a user's status from a webhook event (no permission checks)
func (s *userService) UpdateUserStatusFromWebhook(ctx context.Context, userID uuid.UUID, status models.UserStatus) error {
	s.logger.Info("updating user status from webhook", "user_id", userID, "status", status)

	if err := s.repos.User.UpdateStatus(ctx, userID, status); err != nil {
		s.logger.Error("failed to update user status from webhook", "user_id", userID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to update user status from webhook", err)
	}

	s.logger.Info("user status updated from webhook successfully", "user_id", userID, "status", status)
	return nil
}
