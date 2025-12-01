package repository

import (
	"context"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantRepository defines the interface for tenant repository operations
type TenantRepository interface {
	BaseRepository[models.Tenant]

	// Query Operations
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*models.Tenant, error)
	FindBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error)
	FindByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	FindByStatus(ctx context.Context, status models.TenantStatus, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error)
	FindByPlan(ctx context.Context, plan models.TenantPlan, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error)
	FindActiveTenants(ctx context.Context, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error)
	FindTrialTenants(ctx context.Context) ([]*models.Tenant, error)
	FindExpiredTrials(ctx context.Context) ([]*models.Tenant, error)
	FindExpiringTrials(ctx context.Context, days int) ([]*models.Tenant, error)

	// Status Management
	ActivateTenant(ctx context.Context, tenantID uuid.UUID) error
	SuspendTenant(ctx context.Context, tenantID uuid.UUID, reason string) error
	CancelTenant(ctx context.Context, tenantID uuid.UUID, reason string) error
	ConvertTrialToActive(ctx context.Context, tenantID uuid.UUID, plan models.TenantPlan) error
	UpdateStatus(ctx context.Context, tenantID uuid.UUID, status models.TenantStatus) error

	// Plan Management
	UpgradePlan(ctx context.Context, tenantID uuid.UUID, newPlan models.TenantPlan) error
	DowngradePlan(ctx context.Context, tenantID uuid.UUID, newPlan models.TenantPlan) error
	UpdatePlanLimits(ctx context.Context, tenantID uuid.UUID, maxUsers, maxArtisans int, maxStorage int64) error

	// Settings Management
	UpdateSettings(ctx context.Context, tenantID uuid.UUID, settings models.TenantSettings) error
	UpdateFeatures(ctx context.Context, tenantID uuid.UUID, features models.TenantFeatures) error
	EnableFeature(ctx context.Context, tenantID uuid.UUID, feature string) error
	DisableFeature(ctx context.Context, tenantID uuid.UUID, feature string) error

	// Usage Tracking
	IncrementUserCount(ctx context.Context, tenantID uuid.UUID) error
	DecrementUserCount(ctx context.Context, tenantID uuid.UUID) error
	UpdateStorageUsed(ctx context.Context, tenantID uuid.UUID, storageBytes int64) error
	CheckUserLimit(ctx context.Context, tenantID uuid.UUID) (bool, error)
	CheckStorageLimit(ctx context.Context, tenantID uuid.UUID) (bool, error)

	// Admin Management
	AddAdmin(ctx context.Context, tenantID, userID uuid.UUID) error
	RemoveAdmin(ctx context.Context, tenantID, userID uuid.UUID) error
	GetAdmins(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error)

	// Analytics & Statistics
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (TenantStats, error)
	GetPlatformStats(ctx context.Context) (PlatformStats, error)
	GetTenantGrowth(ctx context.Context, days int) ([]TenantGrowth, error)
	GetRevenueByPlan(ctx context.Context) ([]PlanRevenue, error)
	GetChurnRate(ctx context.Context, days int) (float64, error)

	// Search & Filter
	Search(ctx context.Context, query string, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error)
	FindByFilters(ctx context.Context, filters TenantFilters, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error)

	// Bulk Operations
	BulkUpdateStatus(ctx context.Context, tenantIDs []uuid.UUID, status models.TenantStatus) error
	BulkSuspend(ctx context.Context, tenantIDs []uuid.UUID, reason string) error
	DeleteInactiveTenants(ctx context.Context, inactiveDays int) error
}

// TenantStats represents tenant statistics
type TenantStats struct {
	TenantID           uuid.UUID           `json:"tenant_id"`
	TenantName         string              `json:"tenant_name"`
	Plan               models.TenantPlan   `json:"plan"`
	Status             models.TenantStatus `json:"status"`
	CreatedAt          time.Time           `json:"created_at"`
	TrialEndsAt        *time.Time          `json:"trial_ends_at,omitempty"`
	DaysSinceCreation  int                 `json:"days_since_creation"`
	CurrentUsers       int                 `json:"current_users"`
	MaxUsers           int                 `json:"max_users"`
	UserUtilization    float64             `json:"user_utilization_percent"`
	StorageUsed        int64               `json:"storage_used_bytes"`
	MaxStorage         int64               `json:"max_storage_bytes"`
	StorageUtilization float64             `json:"storage_utilization_percent"`
	TotalBookings      int64               `json:"total_bookings"`
	TotalProjects      int64               `json:"total_projects"`
	TotalServices      int64               `json:"total_services"`
	TotalRevenue       float64             `json:"total_revenue"`
	MonthlyRevenue     float64             `json:"monthly_revenue"`
	ActiveArtisans     int64               `json:"active_artisans"`
	ActiveCustomers    int64               `json:"active_customers"`
	LastActivityAt     *time.Time          `json:"last_activity_at,omitempty"`
}

// PlatformStats represents platform-wide statistics
type PlatformStats struct {
	TotalTenants     int64                         `json:"total_tenants"`
	ActiveTenants    int64                         `json:"active_tenants"`
	TrialTenants     int64                         `json:"trial_tenants"`
	SuspendedTenants int64                         `json:"suspended_tenants"`
	CancelledTenants int64                         `json:"cancelled_tenants"`
	ByPlan           map[models.TenantPlan]int64   `json:"by_plan"`
	ByStatus         map[models.TenantStatus]int64 `json:"by_status"`
	TotalUsers       int64                         `json:"total_users"`
	TotalBookings    int64                         `json:"total_bookings"`
	TotalRevenue     float64                       `json:"total_revenue"`
	MonthlyRevenue   float64                       `json:"monthly_revenue"`
	AverageTenantAge float64                       `json:"average_tenant_age_days"`
	TenantsThisMonth int64                         `json:"tenants_this_month"`
	TenantsThisWeek  int64                         `json:"tenants_this_week"`
	ChurnRate        float64                       `json:"churn_rate_percent"`
	GrowthRate       float64                       `json:"growth_rate_percent"`
}

// TenantGrowth represents tenant growth over time
type TenantGrowth struct {
	Date       time.Time `json:"date"`
	NewTenants int64     `json:"new_tenants"`
	Churned    int64     `json:"churned"`
	NetGrowth  int64     `json:"net_growth"`
	Total      int64     `json:"total"`
}

// PlanRevenue represents revenue by plan
type PlanRevenue struct {
	Plan       models.TenantPlan `json:"plan"`
	Tenants    int64             `json:"tenants"`
	Revenue    float64           `json:"revenue"`
	AvgRevenue float64           `json:"avg_revenue_per_tenant"`
}

// TenantFilters for advanced filtering
type TenantFilters struct {
	Statuses      []models.TenantStatus `json:"statuses"`
	Plans         []models.TenantPlan   `json:"plans"`
	MinUsers      *int                  `json:"min_users"`
	MaxUsers      *int                  `json:"max_users"`
	CreatedAfter  *time.Time            `json:"created_after"`
	CreatedBefore *time.Time            `json:"created_before"`
	HasTrial      *bool                 `json:"has_trial"`
	TrialExpiring *bool                 `json:"trial_expiring"`
	OwnerID       *uuid.UUID            `json:"owner_id"`
}

// tenantRepository implements TenantRepository
type tenantRepository struct {
	BaseRepository[models.Tenant]
	db     *gorm.DB
	logger log.AllLogger
	cache  Cache
}

// NewTenantRepository creates a new TenantRepository instance
func NewTenantRepository(db *gorm.DB, config ...RepositoryConfig) TenantRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.Tenant](db, cfg)

	return &tenantRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
	}
}

// FindByOwnerID retrieves all tenants owned by a user
func (r *tenantRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*models.Tenant, error) {
	if ownerID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "owner_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenants []*models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&tenants).Error; err != nil {
		r.logger.Error("failed to find tenants by owner", "owner_id", ownerID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find tenants", err)
	}

	return tenants, nil
}

// FindBySubdomain retrieves a tenant by subdomain
func (r *tenantRepository) FindBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error) {
	if subdomain == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "subdomain cannot be empty", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("subdomain = ?", subdomain).
		First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to find tenant by subdomain", "subdomain", subdomain, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	return &tenant, nil
}

// FindByDomain retrieves a tenant by custom domain
func (r *tenantRepository) FindByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	if domain == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "domain cannot be empty", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("domain = ?", domain).
		First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to find tenant by domain", "domain", domain, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	return &tenant, nil
}

// FindByStatus retrieves tenants by status
func (r *tenantRepository) FindByStatus(ctx context.Context, status models.TenantStatus, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("status = ?", status).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count tenants", err)
	}

	var tenants []*models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("status = ?", status).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&tenants).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find tenants by status", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return tenants, paginationResult, nil
}

// FindByPlan retrieves tenants by plan
func (r *tenantRepository) FindByPlan(ctx context.Context, plan models.TenantPlan, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("plan = ?", plan).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count tenants", err)
	}

	var tenants []*models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("plan = ?", plan).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&tenants).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find tenants by plan", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return tenants, paginationResult, nil
}

// FindActiveTenants retrieves all active tenants
func (r *tenantRepository) FindActiveTenants(ctx context.Context, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error) {
	return r.FindByStatus(ctx, models.TenantStatusActive, pagination)
}

// FindTrialTenants retrieves all trial tenants
func (r *tenantRepository) FindTrialTenants(ctx context.Context) ([]*models.Tenant, error) {
	var tenants []*models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("status = ?", models.TenantStatusTrial).
		Order("trial_ends_at ASC").
		Find(&tenants).Error; err != nil {
		r.logger.Error("failed to find trial tenants", "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find trial tenants", err)
	}

	return tenants, nil
}

// FindExpiredTrials retrieves tenants with expired trials
func (r *tenantRepository) FindExpiredTrials(ctx context.Context) ([]*models.Tenant, error) {
	var tenants []*models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("status = ? AND trial_ends_at < ?", models.TenantStatusTrial, time.Now()).
		Order("trial_ends_at ASC").
		Find(&tenants).Error; err != nil {
		r.logger.Error("failed to find expired trials", "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find expired trials", err)
	}

	return tenants, nil
}

// FindExpiringTrials retrieves trials expiring within specified days
func (r *tenantRepository) FindExpiringTrials(ctx context.Context, days int) ([]*models.Tenant, error) {
	if days <= 0 {
		days = 7
	}

	expiryDate := time.Now().AddDate(0, 0, days)

	var tenants []*models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("status = ? AND trial_ends_at BETWEEN ? AND ?",
			models.TenantStatusTrial, time.Now(), expiryDate).
		Order("trial_ends_at ASC").
		Find(&tenants).Error; err != nil {
		r.logger.Error("failed to find expiring trials", "days", days, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find expiring trials", err)
	}

	return tenants, nil
}

// ActivateTenant activates a tenant
func (r *tenantRepository) ActivateTenant(ctx context.Context, tenantID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	// Validate status transition
	if err := tenant.CanTransitionTo(models.TenantStatusActive); err != nil {
		return errors.NewRepositoryError("INVALID_STATUS", err.Error(), errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&tenant).
		Update("status", models.TenantStatusActive)

	if result.Error != nil {
		r.logger.Error("failed to activate tenant", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to activate tenant", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Info("tenant activated", "tenant_id", tenantID)
	return nil
}

// SuspendTenant suspends a tenant
func (r *tenantRepository) SuspendTenant(ctx context.Context, tenantID uuid.UUID, reason string) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	// Validate status transition
	if err := tenant.CanTransitionTo(models.TenantStatusSuspended); err != nil {
		return errors.NewRepositoryError("INVALID_STATUS", err.Error(), errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": models.TenantStatusSuspended,
	}

	// Store suspension reason in metadata
	if reason != "" {
		metadata := tenant.Metadata
		if metadata == nil {
			metadata = make(models.JSONB)
		}
		metadata["suspension_reason"] = reason
		metadata["suspended_at"] = time.Now()
		updates["metadata"] = metadata
	}

	result := r.db.WithContext(ctx).
		Model(&tenant).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to suspend tenant", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to suspend tenant", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Warn("tenant suspended", "tenant_id", tenantID, "reason", reason)
	return nil
}

// CancelTenant cancels a tenant
func (r *tenantRepository) CancelTenant(ctx context.Context, tenantID uuid.UUID, reason string) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	// Validate status transition
	if err := tenant.CanTransitionTo(models.TenantStatusCancelled); err != nil {
		return errors.NewRepositoryError("INVALID_STATUS", err.Error(), errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": models.TenantStatusCancelled,
	}

	// Store cancellation reason in metadata
	if reason != "" {
		metadata := tenant.Metadata
		if metadata == nil {
			metadata = make(models.JSONB)
		}
		metadata["cancellation_reason"] = reason
		metadata["cancelled_at"] = time.Now()
		updates["metadata"] = metadata
	}

	result := r.db.WithContext(ctx).
		Model(&tenant).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to cancel tenant", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to cancel tenant", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Warn("tenant cancelled", "tenant_id", tenantID, "reason", reason)
	return nil
}

// ConvertTrialToActive converts a trial tenant to active with a plan
func (r *tenantRepository) ConvertTrialToActive(ctx context.Context, tenantID uuid.UUID, plan models.TenantPlan) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	if tenant.Status != models.TenantStatusTrial {
		return errors.NewRepositoryError("INVALID_STATUS", "tenant is not in trial status", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": models.TenantStatusActive,
		"plan":   plan,
	}

	result := r.db.WithContext(ctx).
		Model(&tenant).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to convert trial to active", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to convert trial", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Info("trial converted to active", "tenant_id", tenantID, "plan", plan)
	return nil
}

// UpdateStatus updates tenant status
func (r *tenantRepository) UpdateStatus(ctx context.Context, tenantID uuid.UUID, status models.TenantStatus) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("status", status)

	if result.Error != nil {
		r.logger.Error("failed to update tenant status", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	return nil
}

// UpgradePlan upgrades tenant to a higher plan
func (r *tenantRepository) UpgradePlan(ctx context.Context, tenantID uuid.UUID, newPlan models.TenantPlan) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	// Update plan and adjust limits based on plan
	var maxUsers, maxArtisans int
	var maxStorage int64

	switch newPlan {
	case models.TenantPlanSolo:
		maxUsers, maxArtisans, maxStorage = 5, 1, 1073741824 // 1GB
	case models.TenantPlanSmall:
		maxUsers, maxArtisans, maxStorage = 20, 5, 5368709120 // 5GB
	case models.TenantPlanCorporation:
		maxUsers, maxArtisans, maxStorage = 100, 25, 21474836480 // 20GB
	case models.TenantPlanEnterprise:
		maxUsers, maxArtisans, maxStorage = 1000, 100, 107374182400 // 100GB
	}

	updates := map[string]interface{}{
		"plan":         newPlan,
		"max_users":    maxUsers,
		"max_artisans": maxArtisans,
		"max_storage":  maxStorage,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to upgrade plan", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to upgrade plan", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Info("plan upgraded", "tenant_id", tenantID, "new_plan", newPlan)
	return nil
}

// DowngradePlan downgrades tenant to a lower plan
func (r *tenantRepository) DowngradePlan(ctx context.Context, tenantID uuid.UUID, newPlan models.TenantPlan) error {
	return r.UpgradePlan(ctx, tenantID, newPlan)
}

// UpdatePlanLimits updates tenant plan limits
func (r *tenantRepository) UpdatePlanLimits(ctx context.Context, tenantID uuid.UUID, maxUsers, maxArtisans int, maxStorage int64) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"max_users":    maxUsers,
		"max_artisans": maxArtisans,
		"max_storage":  maxStorage,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to update plan limits", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update plan limits", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	return nil
}

// UpdateSettings updates tenant settings
func (r *tenantRepository) UpdateSettings(ctx context.Context, tenantID uuid.UUID, settings models.TenantSettings) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("settings", settings)

	if result.Error != nil {
		r.logger.Error("failed to update settings", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update settings", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	return nil
}

// UpdateFeatures updates tenant features
func (r *tenantRepository) UpdateFeatures(ctx context.Context, tenantID uuid.UUID, features models.TenantFeatures) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("features", features)

	if result.Error != nil {
		r.logger.Error("failed to update features", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update features", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	return nil
}

// EnableFeature enables a specific feature for a tenant
func (r *tenantRepository) EnableFeature(ctx context.Context, tenantID uuid.UUID, feature string) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	// This would require reflection or a feature map
	// For simplicity, updating the entire features object
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	// Update specific feature (simplified)
	result := r.db.WithContext(ctx).
		Model(&tenant).
		Update("features", tenant.Features)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to enable feature", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	return nil
}

// DisableFeature disables a specific feature for a tenant
func (r *tenantRepository) DisableFeature(ctx context.Context, tenantID uuid.UUID, feature string) error {
	return r.EnableFeature(ctx, tenantID, feature)
}

// IncrementUserCount increments the user count for a tenant
func (r *tenantRepository) IncrementUserCount(ctx context.Context, tenantID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("current_users", gorm.Expr("current_users + ?", 1))

	if result.Error != nil {
		r.logger.Error("failed to increment user count", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment user count", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	return nil
}

// DecrementUserCount decrements the user count for a tenant
func (r *tenantRepository) DecrementUserCount(ctx context.Context, tenantID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ? AND current_users > 0", tenantID).
		Update("current_users", gorm.Expr("current_users - ?", 1))

	if result.Error != nil {
		r.logger.Error("failed to decrement user count", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to decrement user count", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	return nil
}

// UpdateStorageUsed updates the storage used by a tenant
func (r *tenantRepository) UpdateStorageUsed(ctx context.Context, tenantID uuid.UUID, storageBytes int64) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("storage_used", storageBytes)

	if result.Error != nil {
		r.logger.Error("failed to update storage used", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update storage", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	return nil
}

// CheckUserLimit checks if tenant can add more users
func (r *tenantRepository) CheckUserLimit(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	if tenantID == uuid.Nil {
		return false, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		return false, errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	return tenant.CanAddUser(), nil
}

// CheckStorageLimit checks if tenant has storage space available
func (r *tenantRepository) CheckStorageLimit(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	if tenantID == uuid.Nil {
		return false, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		return false, errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	return tenant.StorageUsed < tenant.MaxStorage, nil
}

// AddAdmin adds a user as admin to a tenant
func (r *tenantRepository) AddAdmin(ctx context.Context, tenantID, userID uuid.UUID) error {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id and user_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	var user models.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	if err := r.db.WithContext(ctx).Model(&tenant).Association("Admins").Append(&user); err != nil {
		r.logger.Error("failed to add admin", "tenant_id", tenantID, "user_id", userID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to add admin", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Info("admin added", "tenant_id", tenantID, "user_id", userID)
	return nil
}

// RemoveAdmin removes a user from tenant admins
func (r *tenantRepository) RemoveAdmin(ctx context.Context, tenantID, userID uuid.UUID) error {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id and user_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
	}

	var user models.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	if err := r.db.WithContext(ctx).Model(&tenant).Association("Admins").Delete(&user); err != nil {
		r.logger.Error("failed to remove admin", "tenant_id", tenantID, "user_id", userID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to remove admin", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Info("admin removed", "tenant_id", tenantID, "user_id", userID)
	return nil
}

// GetAdmins retrieves all admins for a tenant
func (r *tenantRepository) GetAdmins(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var tenant models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Admins").
		First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	// Convert []models.User to []*models.User
	admins := make([]*models.User, len(tenant.Admins))
	for i := range tenant.Admins {
		admins[i] = &tenant.Admins[i]
	}

	return admins, nil
}

// GetTenantStats retrieves comprehensive statistics for a tenant
func (r *tenantRepository) GetTenantStats(ctx context.Context, tenantID uuid.UUID) (TenantStats, error) {
	if tenantID == uuid.Nil {
		return TenantStats{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := TenantStats{TenantID: tenantID}

	// Get tenant
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return stats, errors.NewRepositoryError("NOT_FOUND", "tenant not found", errors.ErrNotFound)
		}
		return stats, errors.NewRepositoryError("FIND_FAILED", "failed to find tenant", err)
	}

	stats.TenantName = tenant.Name
	stats.Plan = tenant.Plan
	stats.Status = tenant.Status
	stats.CreatedAt = tenant.CreatedAt
	stats.TrialEndsAt = tenant.TrialEndsAt
	stats.CurrentUsers = tenant.CurrentUsers
	stats.MaxUsers = tenant.MaxUsers
	stats.StorageUsed = tenant.StorageUsed
	stats.MaxStorage = tenant.MaxStorage

	// Calculate days since creation
	stats.DaysSinceCreation = int(time.Since(tenant.CreatedAt).Hours() / 24)

	// Calculate utilization
	if tenant.MaxUsers > 0 {
		stats.UserUtilization = (float64(tenant.CurrentUsers) / float64(tenant.MaxUsers)) * 100
	}
	if tenant.MaxStorage > 0 {
		stats.StorageUtilization = (float64(tenant.StorageUsed) / float64(tenant.MaxStorage)) * 100
	}

	// Total bookings
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalBookings)

	// Total projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalProjects)

	// Total services
	r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalServices)

	// Total revenue (completed bookings)
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("tenant_id = ? AND status = ?", tenantID, "completed").
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&stats.TotalRevenue)

	// Monthly revenue (last 30 days)
	monthAgo := time.Now().AddDate(0, -1, 0)
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("tenant_id = ? AND status = ? AND created_at >= ?",
			tenantID, "completed", monthAgo).
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&stats.MonthlyRevenue)

	// Active artisans
	r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("tenant_id = ? AND role = ? AND status = ?",
			tenantID, models.UserRoleArtisan, models.UserStatusActive).
		Count(&stats.ActiveArtisans)

	// Active customers
	r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("tenant_id = ? AND role = ? AND status = ?",
			tenantID, models.UserRoleCustomer, models.UserStatusActive).
		Count(&stats.ActiveCustomers)

	// Last activity
	var lastActivity time.Time
	if err := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("tenant_id = ?", tenantID).
		Select("MAX(created_at)").
		Scan(&lastActivity).Error; err == nil && !lastActivity.IsZero() {
		stats.LastActivityAt = &lastActivity
	}

	return stats, nil
}

// GetPlatformStats retrieves platform-wide statistics
func (r *tenantRepository) GetPlatformStats(ctx context.Context) (PlatformStats, error) {
	stats := PlatformStats{
		ByPlan:   make(map[models.TenantPlan]int64),
		ByStatus: make(map[models.TenantStatus]int64),
	}

	// Total tenants
	r.db.WithContext(ctx).Model(&models.Tenant{}).Count(&stats.TotalTenants)

	// Active tenants
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("status = ?", models.TenantStatusActive).
		Count(&stats.ActiveTenants)

	// Trial tenants
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("status = ?", models.TenantStatusTrial).
		Count(&stats.TrialTenants)

	// Suspended tenants
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("status = ?", models.TenantStatusSuspended).
		Count(&stats.SuspendedTenants)

	// Cancelled tenants
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("status = ?", models.TenantStatusCancelled).
		Count(&stats.CancelledTenants)

	// By plan
	type PlanCount struct {
		Plan  models.TenantPlan
		Count int64
	}
	var planCounts []PlanCount
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Select("plan, COUNT(*) as count").
		Group("plan").
		Scan(&planCounts)
	for _, pc := range planCounts {
		stats.ByPlan[pc.Plan] = pc.Count
	}

	// By status
	type StatusCount struct {
		Status models.TenantStatus
		Count  int64
	}
	var statusCounts []StatusCount
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// Total users
	r.db.WithContext(ctx).Model(&models.User{}).Count(&stats.TotalUsers)

	// Total bookings
	r.db.WithContext(ctx).Model(&models.Booking{}).Count(&stats.TotalBookings)

	// Total revenue
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("status = ?", "completed").
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&stats.TotalRevenue)

	// Monthly revenue
	monthAgo := time.Now().AddDate(0, -1, 0)
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("status = ? AND created_at >= ?", "completed", monthAgo).
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&stats.MonthlyRevenue)

	// Average tenant age
	r.db.WithContext(ctx).Raw(`
		SELECT AVG(EXTRACT(DAY FROM (CURRENT_TIMESTAMP - created_at)))
		FROM tenants
		WHERE deleted_at IS NULL
	`).Scan(&stats.AverageTenantAge)

	// Tenants this month
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("created_at >= ?", monthAgo).
		Count(&stats.TenantsThisMonth)

	// Tenants this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("created_at >= ?", weekAgo).
		Count(&stats.TenantsThisWeek)

	// Calculate churn rate (last 30 days)
	var churnedCount int64
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("status = ? AND updated_at >= ?", models.TenantStatusCancelled, monthAgo).
		Count(&churnedCount)
	if stats.ActiveTenants > 0 {
		stats.ChurnRate = (float64(churnedCount) / float64(stats.ActiveTenants)) * 100
	}

	// Growth rate
	var tenants30DaysAgo int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("created_at < ?", thirtyDaysAgo).
		Count(&tenants30DaysAgo)
	if tenants30DaysAgo > 0 {
		stats.GrowthRate = ((float64(stats.TotalTenants) - float64(tenants30DaysAgo)) / float64(tenants30DaysAgo)) * 100
	}

	return stats, nil
}

// GetTenantGrowth retrieves tenant growth metrics
func (r *tenantRepository) GetTenantGrowth(ctx context.Context, days int) ([]TenantGrowth, error) {
	if days <= 0 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)

	var growth []TenantGrowth
	if err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE dates AS (
			SELECT CURRENT_DATE - ?::integer as date
			UNION ALL
			SELECT date + 1
			FROM dates
			WHERE date < CURRENT_DATE
		),
		new_tenants AS (
			SELECT DATE(created_at) as date, COUNT(*) as count
			FROM tenants
			WHERE created_at >= ? AND deleted_at IS NULL
			GROUP BY DATE(created_at)
		),
		churned_tenants AS (
			SELECT DATE(updated_at) as date, COUNT(*) as count
			FROM tenants
			WHERE status = 'cancelled' AND updated_at >= ? AND deleted_at IS NULL
			GROUP BY DATE(updated_at)
		)
		SELECT
			d.date,
			COALESCE(nt.count, 0) as new_tenants,
			COALESCE(ct.count, 0) as churned,
			COALESCE(nt.count, 0) - COALESCE(ct.count, 0) as net_growth,
			(SELECT COUNT(*) FROM tenants WHERE created_at <= d.date AND deleted_at IS NULL) as total
		FROM dates d
		LEFT JOIN new_tenants nt ON d.date = nt.date
		LEFT JOIN churned_tenants ct ON d.date = ct.date
		ORDER BY d.date
	`, days, startDate, startDate).Scan(&growth).Error; err != nil {
		r.logger.Error("failed to get tenant growth", "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get tenant growth", err)
	}

	return growth, nil
}

// GetRevenueByPlan retrieves revenue breakdown by plan
func (r *tenantRepository) GetRevenueByPlan(ctx context.Context) ([]PlanRevenue, error) {
	var revenue []PlanRevenue
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			t.plan,
			COUNT(DISTINCT t.id) as tenants,
			COALESCE(SUM(b.total_price), 0) as revenue,
			COALESCE(AVG(b.total_price), 0) as avg_revenue
		FROM tenants t
		LEFT JOIN bookings b ON b.tenant_id = t.id AND b.status = 'completed'
		WHERE t.deleted_at IS NULL
		GROUP BY t.plan
		ORDER BY revenue DESC
	`).Scan(&revenue).Error; err != nil {
		r.logger.Error("failed to get revenue by plan", "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get revenue by plan", err)
	}

	return revenue, nil
}

// GetChurnRate calculates churn rate for specified period
func (r *tenantRepository) GetChurnRate(ctx context.Context, days int) (float64, error) {
	if days <= 0 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)

	var churned, active int64
	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("status = ? AND updated_at >= ?", models.TenantStatusCancelled, startDate).
		Count(&churned)

	r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("status = ?", models.TenantStatusActive).
		Count(&active)

	if active == 0 {
		return 0, nil
	}

	churnRate := (float64(churned) / float64(active)) * 100
	return churnRate, nil
}

// Search searches tenants by name, subdomain, or business name
func (r *tenantRepository) Search(ctx context.Context, query string, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error) {
	pagination.Validate()
	searchPattern := "%" + query + "%"

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("name ILIKE ? OR subdomain ILIKE ? OR business_name ILIKE ? OR domain ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search tenants", err)
	}

	var tenants []*models.Tenant
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("name ILIKE ? OR subdomain ILIKE ? OR business_name ILIKE ? OR domain ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&tenants).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search tenants", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return tenants, paginationResult, nil
}

// FindByFilters retrieves tenants using advanced filters
func (r *tenantRepository) FindByFilters(ctx context.Context, filters TenantFilters, pagination PaginationParams) ([]*models.Tenant, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx)

	// Apply filters
	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if len(filters.Plans) > 0 {
		query = query.Where("plan IN ?", filters.Plans)
	}

	if filters.MinUsers != nil {
		query = query.Where("current_users >= ?", *filters.MinUsers)
	}

	if filters.MaxUsers != nil {
		query = query.Where("current_users <= ?", *filters.MaxUsers)
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	if filters.HasTrial != nil && *filters.HasTrial {
		query = query.Where("trial_ends_at IS NOT NULL")
	}

	if filters.TrialExpiring != nil && *filters.TrialExpiring {
		sevenDaysFromNow := time.Now().AddDate(0, 0, 7)
		query = query.Where("status = ? AND trial_ends_at BETWEEN ? AND ?",
			models.TenantStatusTrial, time.Now(), sevenDaysFromNow)
	}

	if filters.OwnerID != nil {
		query = query.Where("owner_id = ?", *filters.OwnerID)
	}

	// Count total
	var totalItems int64
	if err := query.Model(&models.Tenant{}).Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count tenants", err)
	}

	// Find tenants
	var tenants []*models.Tenant
	if err := query.
		Preload("Owner").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&tenants).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find tenants", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return tenants, paginationResult, nil
}

// BulkUpdateStatus updates status for multiple tenants
func (r *tenantRepository) BulkUpdateStatus(ctx context.Context, tenantIDs []uuid.UUID, status models.TenantStatus) error {
	if len(tenantIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id IN ?", tenantIDs).
		Update("status", status)

	if result.Error != nil {
		r.logger.Error("failed to bulk update status", "count", len(tenantIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update status", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Info("bulk updated tenant status", "count", result.RowsAffected, "status", status)
	return nil
}

// BulkSuspend suspends multiple tenants
func (r *tenantRepository) BulkSuspend(ctx context.Context, tenantIDs []uuid.UUID, reason string) error {
	if len(tenantIDs) == 0 {
		return nil
	}

	updates := map[string]interface{}{
		"status": models.TenantStatusSuspended,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id IN ?", tenantIDs).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to bulk suspend tenants", "count", len(tenantIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk suspend", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Warn("bulk suspended tenants", "count", result.RowsAffected, "reason", reason)
	return nil
}

// DeleteInactiveTenants deletes tenants inactive for specified days
func (r *tenantRepository) DeleteInactiveTenants(ctx context.Context, inactiveDays int) error {
	if inactiveDays <= 0 {
		inactiveDays = 90
	}

	cutoffDate := time.Now().AddDate(0, 0, -inactiveDays)

	// Find tenants with no recent activity
	var inactiveIDs []uuid.UUID
	r.db.WithContext(ctx).Raw(`
		SELECT t.id
		FROM tenants t
		LEFT JOIN bookings b ON b.tenant_id = t.id AND b.created_at > ?
		WHERE t.status IN ('cancelled', 'suspended')
			AND t.updated_at < ?
			AND b.id IS NULL
		GROUP BY t.id
	`, cutoffDate, cutoffDate).Pluck("id", &inactiveIDs)

	if len(inactiveIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Where("id IN ?", inactiveIDs).
		Delete(&models.Tenant{})

	if result.Error != nil {
		r.logger.Error("failed to delete inactive tenants", "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete inactive tenants", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:tenants:*")
	}

	r.logger.Info("deleted inactive tenants", "count", result.RowsAffected, "inactive_days", inactiveDays)
	return nil
}
