package repository

import (
	"context"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	BaseRepository[models.User]

	// Authentication & Lookup
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByLogtoID(ctx context.Context, logtoUserID string) (*models.User, error)
	GetByEmailWithTenant(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error)
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
	VerifyPhone(ctx context.Context, userID uuid.UUID) error

	// Tenant Management
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	GetTenantUsersByRole(ctx context.Context, tenantID uuid.UUID, role models.UserRole, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	GetPlatformUsers(ctx context.Context, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)

	// Role Management
	UpdateRole(ctx context.Context, userID uuid.UUID, newRole models.UserRole) error
	GetByRole(ctx context.Context, role models.UserRole, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	GetArtisans(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	GetCustomers(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	GetTenantAdmins(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error)

	// Status Management
	UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error
	ActivateUser(ctx context.Context, userID uuid.UUID) error
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	SuspendUser(ctx context.Context, userID uuid.UUID, reason string) error
	GetActiveUsers(ctx context.Context, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	GetInactiveUsers(ctx context.Context, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error)

	// Security & Authentication
	RecordLogin(ctx context.Context, userID uuid.UUID) error
	RecordFailedLogin(ctx context.Context, userID uuid.UUID) error
	UnlockUser(ctx context.Context, userID uuid.UUID) error
	GetLockedUsers(ctx context.Context) ([]*models.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
	EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error
	DisableMFA(ctx context.Context, userID uuid.UUID) error
	UpdateSessionToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error

	// Compliance & GDPR
	AcceptTerms(ctx context.Context, userID uuid.UUID, version string) error
	AcceptPrivacyPolicy(ctx context.Context, userID uuid.UUID) error
	UpdateConsent(ctx context.Context, userID uuid.UUID, dataProcessing, marketing bool) error
	MarkForDeletion(ctx context.Context, userID uuid.UUID, scheduledDate time.Time) error
	GetUsersMarkedForDeletion(ctx context.Context) ([]*models.User, error)
	PermanentlyDeleteUser(ctx context.Context, userID uuid.UUID) error

	// Profile Management
	UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error
	UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
	UpdatePreferences(ctx context.Context, userID uuid.UUID, timezone, language string) error

	// Search & Filter
	Search(ctx context.Context, query string, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	FindByFilters(ctx context.Context, filters UserFilters, pagination PaginationParams) ([]*models.User, PaginationResult, error)
	GetRecentlyActive(ctx context.Context, tenantID *uuid.UUID, hours int, pagination PaginationParams) ([]*models.User, PaginationResult, error)

	// Analytics
	GetUserStats(ctx context.Context, tenantID *uuid.UUID) (UserStats, error)
	GetRegistrationStats(ctx context.Context, startDate, endDate time.Time) (RegistrationStats, error)
	GetUserGrowth(ctx context.Context, tenantID *uuid.UUID, months int) ([]UserGrowthData, error)
}

type UserFilters struct {
	Roles             []models.UserRole   `json:"roles"`
	Statuses          []models.UserStatus `json:"statuses"`
	TenantIDs         []uuid.UUID         `json:"tenant_ids"`
	IsPlatformUser    *bool               `json:"is_platform_user"`
	EmailVerified     *bool               `json:"email_verified"`
	PhoneVerified     *bool               `json:"phone_verified"`
	MFAEnabled        *bool               `json:"mfa_enabled"`
	IsLocked          *bool               `json:"is_locked"`
	MarkedForDeletion *bool               `json:"marked_for_deletion"`
	CreatedAfter      *time.Time          `json:"created_after"`
	CreatedBefore     *time.Time          `json:"created_before"`
	LastLoginAfter    *time.Time          `json:"last_login_after"`
	LastLoginBefore   *time.Time          `json:"last_login_before"`
}

type UserStats struct {
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

type RegistrationStats struct {
	StartDate          time.Time                 `json:"start_date"`
	EndDate            time.Time                 `json:"end_date"`
	TotalRegistrations int64                     `json:"total_registrations"`
	ByRole             map[models.UserRole]int64 `json:"by_role"`
	ByDay              []DailyRegistration       `json:"by_day"`
	AveragePerDay      float64                   `json:"average_per_day"`
}

type DailyRegistration struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

type UserGrowthData struct {
	Month      time.Time `json:"month"`
	NewUsers   int64     `json:"new_users"`
	TotalUsers int64     `json:"total_users"`
	Growth     float64   `json:"growth_percentage"`
}

// userRepository implements UserRepository
type userRepository struct {
	BaseRepository[models.User]
	db      *gorm.DB
	logger  log.AllLogger
	cache   Cache
	metrics MetricsCollector
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, config ...RepositoryConfig) UserRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 5 * time.Minute
	}

	baseRepo := NewBaseRepository[models.User](db, cfg)

	return &userRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
		metrics:        cfg.Metrics,
	}
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_email", "users", time.Since(start), nil)
		}
	}()

	if email == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "email cannot be empty", errors.ErrInvalidInput)
	}

	email = strings.ToLower(strings.TrimSpace(email))

	// Try cache first
	cacheKey := r.getCacheKey("email", email)
	if r.cache != nil {
		var user models.User
		if err := r.cache.GetJSON(ctx, cacheKey, &user); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit("users")
			}
			return &user, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss("users")
		}
	}

	var user models.User
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Where("LOWER(email) = ?", email).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get user by email", "email", email, "error", err)
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get user", err)
	}

	// Cache the result
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, user, 5*time.Minute); err != nil {
			r.logger.Warn("failed to cache user", "email", email, "error", err)
		}
	}

	return &user, nil
}

// GetByLogtoID retrieve a user by Logto user ID
func (r *userRepository) GetByLogtoID(ctx context.Context, logtoUserID string) (*models.User, error) {
	if logtoUserID == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "logto user ID cannot be empty", errors.ErrInvalidInput)
	}

	// try cache first
	cacheKey := r.getCacheKey("logto", logtoUserID)
	if r.cache != nil {
		var user models.User
		if err := r.cache.GetJSON(ctx, cacheKey, user); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit("users")
			}
			return &user, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheHit("users")
		}
	}

	var user models.User
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Where("logto_user_id = ?", logtoUserID).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get user by logto user ID", "logtoUserID", logtoUserID, "error", err)
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get user", err)
	}

	// Cache the result
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, user, 5*time.Minute); err != nil {
			r.logger.Warn("failed to cache user", "logtoUserID", logtoUserID, "error", err)
		}
	}

	return &user, nil
}

// GetByEmailWithTenant retrieves a user by email within a specific tenant
func (r *userRepository) GetByEmailWithTenant(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	if email == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "email cannot be empty", errors.ErrInvalidInput)
	}

	email = strings.ToLower(strings.TrimSpace(email))

	var user models.User
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Where("LOWER(email) = ? AND tenant_id = ?", email, tenantID).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get user", err)
	}

	return &user, nil
}

// VerifyEmail marks a user's email as verified
func (r *userRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("email_verified", true)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to verify email", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("email verified", "user_id", userID)
	return nil
}

// VerifyPhone marks a user's phone as verified
func (r *userRepository) VerifyPhone(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("phone_verified", true)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to verify phone", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("phone verified", "user_id", userID)
	return nil
}

// GetByTenantID retrieves all users for a tenant
func (r *userRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	var users []*models.User
	if err := query.
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// GetTenantUsersByRole retrieves users by role for a tenant
func (r *userRepository) GetTenantUsersByRole(ctx context.Context, tenantID uuid.UUID, role models.UserRole, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.User{}).
		Where("tenant_id = ? AND role = ?", tenantID, role)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	var users []*models.User
	if err := query.
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// GetPlatformUsers retrieves all users from the platform.
func (r *userRepository) GetPlatformUsers(ctx context.Context, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("is_platform_user = ?", true)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	var users []*models.User
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// CountByTenant counts the number of users in a tenant.
func (r *userRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error; err != nil {
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}
	return count, nil
}

// UpdateRole updates the role of a user.
func (r *userRepository) UpdateRole(ctx context.Context, userID uuid.UUID, newRole models.UserRole) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("role", newRole)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update role", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("user role updated", "user_id", userID, "new_role", newRole)
	return nil
}

// GetByRole retrieves users by role.
func (r *userRepository) GetByRole(ctx context.Context, role models.UserRole, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("role = ?", role)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	var users []*models.User
	if err := query.
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// GetArtisans retrieves all artisans from the database.
func (r *userRepository) GetArtisans(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	return r.GetTenantUsersByRole(ctx, tenantID, models.UserRoleArtisan, pagination)
}

// GetCustomers retrieves all customers from the database.
func (r *userRepository) GetCustomers(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	return r.GetTenantUsersByRole(ctx, tenantID, models.UserRoleCustomer, pagination)
}

// GetTenantAdmins retrieves all tenant admins from the database.
func (r *userRepository) GetTenantAdmins(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND role IN ?", tenantID, []models.UserRole{
			models.UserRoleTenantOwner,
			models.UserRoleTenantAdmin,
		}).Find(&users).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}
	return users, nil
}

// UpdateStatus updates the status of a user.
func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("status", status)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("user status updated", "user_id", userID, "status", status)
	return nil
}

// ActivateUser activates a user.
func (r *userRepository) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	return r.UpdateStatus(ctx, userID, models.UserStatusActive)
}

// DeactivateUser deactivates a user
func (r *userRepository) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	return r.UpdateStatus(ctx, userID, models.UserStatusInactive)
}

// SuspendUser suspends a user with a reason
func (r *userRepository) SuspendUser(ctx context.Context, userID uuid.UUID, reason string) error {
	updates := map[string]interface{}{
		"status":                models.UserStatusSuspended,
		"account_locked_reason": reason,
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to suspend user", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("user suspended", "user_id", userID, "reason", reason)
	return nil
}

// GetActiveUsers retrieves active users
func (r *userRepository) GetActiveUsers(ctx context.Context, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.User{}).Where("status = ?", models.UserStatusActive)

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	var users []*models.User
	if err := query.
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("last_login_at DESC NULLS LAST").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// GetInactiveUsers retrieves inactive users
func (r *userRepository) GetInactiveUsers(ctx context.Context, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.User{}).Where("status = ?", models.UserStatusInactive)

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	var users []*models.User
	if err := query.
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("updated_at DESC").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// RecordLogin records a successful login
func (r *userRepository) RecordLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_login_at":         now,
		"failed_login_attempts": 0,
		"last_failed_login_at":  nil,
		"locked_until":          nil,
		"account_locked_reason": "",
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to record login", result.Error)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	return nil
}

// RecordFailedLogin records a failed login attempt
func (r *userRepository) RecordFailedLogin(ctx context.Context, userID uuid.UUID) error {
	var user *models.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return errors.NewRepositoryError("GET_FAILED", "failed to get user", err)
	}

	user.RecordFailedLogin()

	if err := r.Update(ctx, user); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	return nil
}

// UnlockUser unlocks a user account
func (r *userRepository) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	updates := map[string]interface{}{
		"failed_login_attempts": 0,
		"locked_until":          nil,
		"account_locked_reason": "",
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to unlock user", result.Error)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("user unlocked", "user_id", userID)
	return nil
}

// GetLockedUsers returns a list of locked users
func (r *userRepository) GetLockedUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("locked_until IS NOT NULL AND locked_until > ?", time.Now()).
		Order("locked_until DESC").
		Find(&users).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find locked users", err)
	}
	return users, nil
}

// UpdatePassword updates a user's password
func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	if newPasswordHash == "" {
		return errors.NewRepositoryError("INVALID_INPUT", "password hash cannot be empty", errors.ErrInvalidInput)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"password_hash":        newPasswordHash,
		"password_changed_at":  now,
		"must_change_password": false,
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update password", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("password updated", "user_id", userID)
	return nil
}

// EnableMFA enables multi-factor authentication for a user
func (r *userRepository) EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error {
	if secret == "" {
		return errors.NewRepositoryError("INVALID_INPUT", "MFA secret cannot be empty", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"mfa_enabled": true,
		"mfa_secret":  secret,
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to enable MFA", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("MFA enabled", "user_id", userID)
	return nil
}

// DisableMFA disables MFA for a user
func (r *userRepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	updates := map[string]interface{}{
		"mfa_enabled": false,
		"mfa_secret":  "",
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to disable MFA", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("MFA disabled", "user_id", userID)
	return nil
}

// UpdateSessionToken updates a user's session token
func (r *userRepository) UpdateSessionToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	updates := map[string]interface{}{
		"session_token":      token,
		"session_expires_at": expiresAt,
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update session token", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	return nil
}

// AcceptTerms records that a user has accepted terms
func (r *userRepository) AcceptTerms(ctx context.Context, userID uuid.UUID, version string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"terms_accepted_at": now,
		"terms_version":     version,
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to accept terms", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("terms accepted", "user_id", userID, "version", version)
	return nil
}

// AcceptPrivacyPolicy records that a user has accepted the privacy policy
func (r *userRepository) AcceptPrivacyPolicy(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("privacy_policy_accepted_at", now)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to accept privacy policy", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("privacy policy accepted", "user_id", userID)
	return nil
}

// UpdateConsent updates user consent preferences
func (r *userRepository) UpdateConsent(ctx context.Context, userID uuid.UUID, dataProcessing, marketing bool) error {
	updates := map[string]interface{}{
		"data_processing_consent": dataProcessing,
		"marketing_consent":       marketing,
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update consent", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("consent updated", "user_id", userID, "data_processing", dataProcessing, "marketing", marketing)
	return nil
}

// MarkForDeletion marks a user for deletion (GDPR compliance)
func (r *userRepository) MarkForDeletion(ctx context.Context, userID uuid.UUID, scheduledDate time.Time) error {
	updates := map[string]interface{}{
		"marked_for_deletion":   true,
		"deletion_scheduled_at": scheduledDate,
		"status":                models.UserStatusInactive,
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark for deletion", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("user marked for deletion", "user_id", userID, "scheduled_date", scheduledDate)
	return nil
}

// GetUsersMarkedForDeletion retrieves all users marked for deletion
func (r *userRepository) GetUsersMarkedForDeletion(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("marked_for_deletion = ? AND deletion_scheduled_at <= ?", true, time.Now()).
		Order("deletion_scheduled_at ASC").
		Find(&users).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find users marked for deletion", err)
	}
	return users, nil
}

// PermanentlyDeleteUser permanently deletes a user and all related data
func (r *userRepository) PermanentlyDeleteUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete related records first

		// Delete artisan profile if exists
		if err := tx.Where("user_id = ?", userID).Delete(&models.Artisan{}).Error; err != nil {
			r.logger.Error("failed to delete artisan profile", "user_id", userID, "error", err)
		}

		// Delete customer profile if exists
		if err := tx.Where("user_id = ?", userID).Delete(&models.Customer{}).Error; err != nil {
			r.logger.Error("failed to delete customer profile", "user_id", userID, "error", err)
		}

		// Permanently delete the user (using Unscoped to bypass soft delete)
		if err := tx.Unscoped().Delete(&models.User{}, "id = ?", userID).Error; err != nil {
			return errors.NewRepositoryError("DELETE_FAILED", "failed to permanently delete user", err)
		}

		// Invalidate cache
		r.invalidateUserCache(ctx, userID)

		r.logger.Info("user permanently deleted", "user_id", userID)
		return nil
	})
}

// UpdateProfile updates user profile fields
func (r *userRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "no updates provided", errors.ErrInvalidInput)
	}

	// Whitelist allowed profile fields
	allowedFields := map[string]bool{
		"first_name":   true,
		"last_name":    true,
		"phone_number": true,
		"avatar_url":   true,
		"bio":          true,
	}

	// Filter updates to only allowed fields
	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "no valid profile fields to update", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(filteredUpdates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update profile", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("profile updated", "user_id", userID)
	return nil
}

// UpdateAvatar updates a user's avatar URL
func (r *userRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("avatar_url", avatarURL)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update avatar", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("avatar updated", "user_id", userID)
	return nil
}

// UpdatePreferences updates user preferences (timezone, language)
func (r *userRepository) UpdatePreferences(ctx context.Context, userID uuid.UUID, timezone, language string) error {
	updates := map[string]interface{}{}

	if timezone != "" {
		updates["timezone"] = timezone
	}

	if language != "" {
		updates["language"] = language
	}

	if len(updates) == 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "no preferences to update", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update preferences", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, userID)

	r.logger.Info("preferences updated", "user_id", userID)
	return nil
}

// Search searches for users by query string
func (r *userRepository) Search(ctx context.Context, query string, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()
	searchPattern := "%" + strings.ToLower(query) + "%"

	dbQuery := r.db.WithContext(ctx).Model(&models.User{}).
		Where("LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ?",
			searchPattern, searchPattern, searchPattern)

	if tenantID != nil {
		dbQuery = dbQuery.Where("tenant_id = ?", *tenantID)
	}

	var totalItems int64
	if err := dbQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	var users []*models.User
	if err := dbQuery.
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// FindByFilters retrieves users using advanced filters
func (r *userRepository) FindByFilters(ctx context.Context, filters UserFilters, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.User{})

	// Apply filters
	if len(filters.Roles) > 0 {
		query = query.Where("role IN ?", filters.Roles)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if len(filters.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filters.TenantIDs)
	}

	if filters.IsPlatformUser != nil {
		query = query.Where("is_platform_user = ?", *filters.IsPlatformUser)
	}

	if filters.EmailVerified != nil {
		query = query.Where("email_verified = ?", *filters.EmailVerified)
	}

	if filters.PhoneVerified != nil {
		query = query.Where("phone_verified = ?", *filters.PhoneVerified)
	}

	if filters.MFAEnabled != nil {
		query = query.Where("mfa_enabled = ?", *filters.MFAEnabled)
	}

	if filters.IsLocked != nil {
		if *filters.IsLocked {
			query = query.Where("locked_until IS NOT NULL AND locked_until > ?", time.Now())
		} else {
			query = query.Where("locked_until IS NULL OR locked_until <= ?", time.Now())
		}
	}

	if filters.MarkedForDeletion != nil {
		query = query.Where("marked_for_deletion = ?", *filters.MarkedForDeletion)
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	if filters.LastLoginAfter != nil {
		query = query.Where("last_login_at >= ?", *filters.LastLoginAfter)
	}

	if filters.LastLoginBefore != nil {
		query = query.Where("last_login_at <= ?", *filters.LastLoginBefore)
	}

	// Count total
	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	// Find users
	var users []*models.User
	if err := query.
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// GetRecentlyActive retrieves users who were active within the specified hours
func (r *userRepository) GetRecentlyActive(ctx context.Context, tenantID *uuid.UUID, hours int, pagination PaginationParams) ([]*models.User, PaginationResult, error) {
	pagination.Validate()

	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	query := r.db.WithContext(ctx).Model(&models.User{}).
		Where("last_login_at >= ?", cutoffTime)

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count users", err)
	}

	var users []*models.User
	if err := query.
		Preload("Tenant").
		Preload("ArtisanProfile").
		Preload("CustomerProfile").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("last_login_at DESC").
		Find(&users).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find users", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return users, paginationResult, nil
}

// GetUserStats retrieves user statistics
func (r *userRepository) GetUserStats(ctx context.Context, tenantID *uuid.UUID) (UserStats, error) {
	var stats UserStats

	query := r.db.WithContext(ctx).Model(&models.User{})
	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	// Total users
	if err := query.Count(&stats.TotalUsers).Error; err != nil {
		return stats, errors.NewRepositoryError("COUNT_FAILED", "failed to count total users", err)
	}

	// By status
	statusQuery := query
	if tenantID != nil {
		statusQuery = r.db.WithContext(ctx).Model(&models.User{}).Where("tenant_id = ?", *tenantID)
	}

	statusQuery.Where("status = ?", models.UserStatusActive).Count(&stats.ActiveUsers)
	statusQuery.Where("status = ?", models.UserStatusInactive).Count(&stats.InactiveUsers)
	statusQuery.Where("status = ?", models.UserStatusSuspended).Count(&stats.SuspendedUsers)
	statusQuery.Where("status = ?", models.UserStatusPending).Count(&stats.PendingUsers)

	// By role
	var roleResults []struct {
		Role  models.UserRole
		Count int64
	}
	roleQuery := r.db.WithContext(ctx).Model(&models.User{}).
		Select("role, COUNT(*) as count").
		Group("role")

	if tenantID != nil {
		roleQuery = roleQuery.Where("tenant_id = ?", *tenantID)
	}

	roleQuery.Scan(&roleResults)

	stats.ByRole = make(map[models.UserRole]int64)
	for _, result := range roleResults {
		stats.ByRole[result.Role] = result.Count
	}

	// By status detailed
	var statusResults []struct {
		Status models.UserStatus
		Count  int64
	}
	statusDetailQuery := r.db.WithContext(ctx).Model(&models.User{}).
		Select("status, COUNT(*) as count").
		Group("status")

	if tenantID != nil {
		statusDetailQuery = statusDetailQuery.Where("tenant_id = ?", *tenantID)
	}

	statusDetailQuery.Scan(&statusResults)

	stats.ByStatus = make(map[models.UserStatus]int64)
	for _, result := range statusResults {
		stats.ByStatus[result.Status] = result.Count
	}

	// Additional counts
	verifyQuery := query
	if tenantID != nil {
		verifyQuery = r.db.WithContext(ctx).Model(&models.User{}).Where("tenant_id = ?", *tenantID)
	}

	verifyQuery.Where("email_verified = ?", true).Count(&stats.EmailVerifiedCount)
	verifyQuery.Where("phone_verified = ?", true).Count(&stats.PhoneVerifiedCount)
	verifyQuery.Where("mfa_enabled = ?", true).Count(&stats.MFAEnabledCount)
	verifyQuery.Where("locked_until IS NOT NULL AND locked_until > ?", time.Now()).Count(&stats.LockedUsersCount)
	verifyQuery.Where("marked_for_deletion = ?", true).Count(&stats.MarkedForDeletion)

	return stats, nil
}

// GetRegistrationStats retrieves registration statistics for a date range
func (r *userRepository) GetRegistrationStats(ctx context.Context, startDate, endDate time.Time) (RegistrationStats, error) {
	stats := RegistrationStats{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Total registrations in date range
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&stats.TotalRegistrations).Error; err != nil {
		return stats, errors.NewRepositoryError("COUNT_FAILED", "failed to count registrations", err)
	}

	// By role
	var roleResults []struct {
		Role  models.UserRole
		Count int64
	}
	r.db.WithContext(ctx).Model(&models.User{}).
		Select("role, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("role").
		Scan(&roleResults)

	stats.ByRole = make(map[models.UserRole]int64)
	for _, result := range roleResults {
		stats.ByRole[result.Role] = result.Count
	}

	// By day
	var dailyResults []struct {
		Date  time.Time
		Count int64
	}
	r.db.WithContext(ctx).Model(&models.User{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyResults)

	stats.ByDay = make([]DailyRegistration, len(dailyResults))
	for i, result := range dailyResults {
		stats.ByDay[i] = DailyRegistration{
			Date:  result.Date,
			Count: result.Count,
		}
	}

	// Calculate average per day
	days := endDate.Sub(startDate).Hours() / 24
	if days > 0 {
		stats.AveragePerDay = float64(stats.TotalRegistrations) / days
	}

	return stats, nil
}

// GetUserGrowth retrieves user growth data for the specified number of months
func (r *userRepository) GetUserGrowth(ctx context.Context, tenantID *uuid.UUID, months int) ([]UserGrowthData, error) {
	growth := make([]UserGrowthData, 0, months)

	now := time.Now()

	for i := months - 1; i >= 0; i-- {
		monthStart := now.AddDate(0, -i, 0)
		monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		var monthData UserGrowthData
		monthData.Month = monthStart

		query := r.db.WithContext(ctx).Model(&models.User{})
		if tenantID != nil {
			query = query.Where("tenant_id = ?", *tenantID)
		}

		// New users this month
		query.Where("created_at BETWEEN ? AND ?", monthStart, monthEnd).
			Count(&monthData.NewUsers)

		// Total users up to end of month
		totalQuery := r.db.WithContext(ctx).Model(&models.User{})
		if tenantID != nil {
			totalQuery = totalQuery.Where("tenant_id = ?", *tenantID)
		}
		totalQuery.Where("created_at <= ?", monthEnd).Count(&monthData.TotalUsers)

		// Calculate growth percentage
		if i > 0 && len(growth) > 0 {
			previousTotal := growth[len(growth)-1].TotalUsers
			if previousTotal > 0 {
				monthData.Growth = ((float64(monthData.TotalUsers) - float64(previousTotal)) / float64(previousTotal)) * 100
			}
		}

		growth = append(growth, monthData)
	}

	return growth, nil
}

// Helper methods

func (r *userRepository) getCacheKey(prefix string, parts ...string) string {
	allParts := append([]string{"repo", "users", prefix}, parts...)
	return strings.Join(allParts, ":")
}

func (r *userRepository) invalidateUserCache(ctx context.Context, userID uuid.UUID) {
	if r.cache == nil {
		return
	}

	patterns := []string{
		r.getCacheKey("id", userID.String()),
		r.getCacheKey("email", "*"),
		r.getCacheKey("logto", "*"),
	}

	for _, pattern := range patterns {
		r.cache.DeletePattern(ctx, pattern)
	}
}
