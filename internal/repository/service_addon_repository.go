package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"context"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceAddonRepository defines the interface for service addon repository operations
type ServiceAddonRepository interface {
	BaseRepository[models.ServiceAddon]

	// Query Operations
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.ServiceAddon, PaginationResult, error)
	FindActiveAddons(ctx context.Context, tenantID uuid.UUID) ([]*models.ServiceAddon, error)
	FindByPriceRange(ctx context.Context, tenantID uuid.UUID, minPrice, maxPrice float64) ([]*models.ServiceAddon, error)
	FindByServiceID(ctx context.Context, serviceID uuid.UUID) ([]*models.ServiceAddon, error)

	// Search
	Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.ServiceAddon, PaginationResult, error)

	// Availability Management
	ActivateAddon(ctx context.Context, addonID uuid.UUID) error
	DeactivateAddon(ctx context.Context, addonID uuid.UUID) error

	// Pricing Management
	UpdatePrice(ctx context.Context, addonID uuid.UUID, newPrice float64) error
	BulkUpdatePrices(ctx context.Context, addonIDs []uuid.UUID, priceAdjustment float64, isPercentage bool) error

	// Analytics
	GetAddonStats(ctx context.Context, tenantID uuid.UUID) (AddonStats, error)
	GetAddonUsage(ctx context.Context, addonID uuid.UUID) (AddonUsage, error)
	GetPopularAddons(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.ServiceAddon, error)

	// Bulk Operations
	BulkActivate(ctx context.Context, addonIDs []uuid.UUID) error
	BulkDeactivate(ctx context.Context, addonIDs []uuid.UUID) error
	BulkDelete(ctx context.Context, addonIDs []uuid.UUID) error
}

// AddonStats represents addon statistics
type AddonStats struct {
	TotalAddons    int64   `json:"total_addons"`
	ActiveAddons   int64   `json:"active_addons"`
	InactiveAddons int64   `json:"inactive_addons"`
	AveragePrice   float64 `json:"average_price"`
	TotalRevenue   float64 `json:"total_revenue"`
	TotalUsage     int64   `json:"total_usage"`
	AvgDuration    float64 `json:"avg_duration_minutes"`
}

// AddonUsage represents usage metrics for an addon
type AddonUsage struct {
	AddonID          uuid.UUID `json:"addon_id"`
	AddonName        string    `json:"addon_name"`
	TimesUsed        int64     `json:"times_used"`
	TotalRevenue     float64   `json:"total_revenue"`
	ServicesCount    int64     `json:"services_count"`
	UsageThisMonth   int64     `json:"usage_this_month"`
	RevenueThisMonth float64   `json:"revenue_this_month"`
}

// serviceAddonRepository implements ServiceAddonRepository
type serviceAddonRepository struct {
	BaseRepository[models.ServiceAddon]
	db     *gorm.DB
	logger log.AllLogger
	cache  Cache
}

// NewServiceAddonRepository creates a new ServiceAddonRepository instance
func NewServiceAddonRepository(db *gorm.DB, config ...RepositoryConfig) ServiceAddonRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.ServiceAddon](db, cfg)

	return &serviceAddonRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
	}
}

// FindByTenantID retrieves all addons for a tenant
func (r *serviceAddonRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.ServiceAddon, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count addons", err)
	}

	var addons []*models.ServiceAddon
	if err := r.db.WithContext(ctx).
		Preload("Services").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("name ASC").
		Find(&addons).Error; err != nil {
		r.logger.Error("failed to find addons", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find addons", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return addons, paginationResult, nil
}

// FindActiveAddons retrieves all active addons for a tenant
func (r *serviceAddonRepository) FindActiveAddons(ctx context.Context, tenantID uuid.UUID) ([]*models.ServiceAddon, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var addons []*models.ServiceAddon
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("name ASC").
		Find(&addons).Error; err != nil {
		r.logger.Error("failed to find active addons", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find active addons", err)
	}

	return addons, nil
}

// FindByPriceRange retrieves addons within a price range
func (r *serviceAddonRepository) FindByPriceRange(ctx context.Context, tenantID uuid.UUID, minPrice, maxPrice float64) ([]*models.ServiceAddon, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var addons []*models.ServiceAddon
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND price BETWEEN ? AND ?", tenantID, minPrice, maxPrice).
		Order("price ASC").
		Find(&addons).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find addons by price range", err)
	}

	return addons, nil
}

// FindByServiceID retrieves all addons for a service
func (r *serviceAddonRepository) FindByServiceID(ctx context.Context, serviceID uuid.UUID) ([]*models.ServiceAddon, error) {
	if serviceID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "service_id cannot be nil", errors.ErrInvalidInput)
	}

	var service models.Service
	if err := r.db.WithContext(ctx).
		Preload("Addons").
		First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "service not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find service", err)
	}

	// Convert []ServiceAddon to []*ServiceAddon
	addons := make([]*models.ServiceAddon, len(service.Addons))
	for i := range service.Addons {
		addons[i] = &service.Addons[i]
	}

	return addons, nil
}

// Search searches addons by name or description
func (r *serviceAddonRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.ServiceAddon, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("tenant_id = ? AND (name ILIKE ? OR description ILIKE ?)",
			tenantID, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count addons", err)
	}

	var addons []*models.ServiceAddon
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (name ILIKE ? OR description ILIKE ?)",
			tenantID, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("name ASC").
		Find(&addons).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search addons", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return addons, paginationResult, nil
}

// ActivateAddon activates an addon
func (r *serviceAddonRepository) ActivateAddon(ctx context.Context, addonID uuid.UUID) error {
	if addonID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "addon_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("id = ?", addonID).
		Update("is_active", true)

	if result.Error != nil {
		r.logger.Error("failed to activate addon", "addon_id", addonID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to activate addon", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "addon not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:service_addons:*")
	}

	return nil
}

// DeactivateAddon deactivates an addon
func (r *serviceAddonRepository) DeactivateAddon(ctx context.Context, addonID uuid.UUID) error {
	if addonID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "addon_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("id = ?", addonID).
		Update("is_active", false)

	if result.Error != nil {
		r.logger.Error("failed to deactivate addon", "addon_id", addonID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to deactivate addon", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "addon not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:service_addons:*")
	}

	return nil
}

// UpdatePrice updates addon price
func (r *serviceAddonRepository) UpdatePrice(ctx context.Context, addonID uuid.UUID, newPrice float64) error {
	if addonID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "addon_id cannot be nil", errors.ErrInvalidInput)
	}
	if newPrice < 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "price cannot be negative", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("id = ?", addonID).
		Update("price", newPrice)

	if result.Error != nil {
		r.logger.Error("failed to update addon price", "addon_id", addonID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update addon price", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "addon not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:service_addons:*")
	}

	r.logger.Info("addon price updated", "addon_id", addonID, "new_price", newPrice)
	return nil
}

// BulkUpdatePrices updates prices for multiple addons
func (r *serviceAddonRepository) BulkUpdatePrices(ctx context.Context, addonIDs []uuid.UUID, priceAdjustment float64, isPercentage bool) error {
	if len(addonIDs) == 0 {
		return nil
	}

	var result *gorm.DB
	if isPercentage {
		// Update by percentage
		result = r.db.WithContext(ctx).
			Model(&models.ServiceAddon{}).
			Where("id IN ?", addonIDs).
			Update("price", gorm.Expr("price * (1 + ? / 100)", priceAdjustment))
	} else {
		// Update by fixed amount
		result = r.db.WithContext(ctx).
			Model(&models.ServiceAddon{}).
			Where("id IN ?", addonIDs).
			Update("price", gorm.Expr("price + ?", priceAdjustment))
	}

	if result.Error != nil {
		r.logger.Error("failed to bulk update addon prices", "count", len(addonIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update addon prices", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:service_addons:*")
	}

	r.logger.Info("bulk updated addon prices", "count", result.RowsAffected, "adjustment", priceAdjustment, "is_percentage", isPercentage)
	return nil
}

// GetAddonStats retrieves addon statistics
func (r *serviceAddonRepository) GetAddonStats(ctx context.Context, tenantID uuid.UUID) (AddonStats, error) {
	if tenantID == uuid.Nil {
		return AddonStats{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := AddonStats{}

	// Total addons
	r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalAddons)

	// Active addons
	r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Count(&stats.ActiveAddons)

	// Inactive addons
	stats.InactiveAddons = stats.TotalAddons - stats.ActiveAddons

	// Average price
	r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("tenant_id = ?", tenantID).
		Select("AVG(price)").
		Scan(&stats.AveragePrice)

	// Average duration
	r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("tenant_id = ? AND duration_minutes > 0", tenantID).
		Select("AVG(duration_minutes)").
		Scan(&stats.AvgDuration)

	// Total usage (count from booking_addons table would be needed)
	// For now, count from many-to-many relationship
	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM service_addon_relations sar
		JOIN service_addons sa ON sa.id = sar.service_addon_id
		WHERE sa.tenant_id = ?
	`, tenantID).Scan(&stats.TotalUsage)

	// Total revenue from bookings with addons (would need booking_addons table)
	// Placeholder for now
	stats.TotalRevenue = 0

	return stats, nil
}

// GetAddonUsage retrieves usage metrics for an addon
func (r *serviceAddonRepository) GetAddonUsage(ctx context.Context, addonID uuid.UUID) (AddonUsage, error) {
	if addonID == uuid.Nil {
		return AddonUsage{}, errors.NewRepositoryError("INVALID_INPUT", "addon_id cannot be nil", errors.ErrInvalidInput)
	}

	usage := AddonUsage{
		AddonID: addonID,
	}

	// Get addon details
	var addon models.ServiceAddon
	if err := r.db.WithContext(ctx).First(&addon, addonID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return usage, errors.NewRepositoryError("NOT_FOUND", "addon not found", errors.ErrNotFound)
		}
		return usage, errors.NewRepositoryError("FIND_FAILED", "failed to find addon", err)
	}
	usage.AddonName = addon.Name

	// Count services using this addon
	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT service_id)
		FROM service_addon_relations
		WHERE service_addon_id = ?
	`, addonID).Scan(&usage.ServicesCount)

	// Times used would require booking_addons tracking
	usage.TimesUsed = 0
	usage.TotalRevenue = 0
	usage.UsageThisMonth = 0
	usage.RevenueThisMonth = 0

	return usage, nil
}

// GetPopularAddons retrieves most used addons
func (r *serviceAddonRepository) GetPopularAddons(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.ServiceAddon, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if limit <= 0 {
		limit = 10
	}

	var addons []*models.ServiceAddon
	if err := r.db.WithContext(ctx).Raw(`
		SELECT sa.*
		FROM service_addons sa
		LEFT JOIN service_addon_relations sar ON sar.service_addon_id = sa.id
		WHERE sa.tenant_id = ? AND sa.is_active = ?
		GROUP BY sa.id
		ORDER BY COUNT(sar.service_id) DESC
		LIMIT ?
	`, tenantID, true, limit).Scan(&addons).Error; err != nil {
		r.logger.Error("failed to find popular addons", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find popular addons", err)
	}

	return addons, nil
}

// BulkActivate activates multiple addons
func (r *serviceAddonRepository) BulkActivate(ctx context.Context, addonIDs []uuid.UUID) error {
	if len(addonIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("id IN ?", addonIDs).
		Update("is_active", true)

	if result.Error != nil {
		r.logger.Error("failed to bulk activate addons", "count", len(addonIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk activate addons", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:service_addons:*")
	}

	r.logger.Info("bulk activated addons", "count", result.RowsAffected)
	return nil
}

// BulkDeactivate deactivates multiple addons
func (r *serviceAddonRepository) BulkDeactivate(ctx context.Context, addonIDs []uuid.UUID) error {
	if len(addonIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.ServiceAddon{}).
		Where("id IN ?", addonIDs).
		Update("is_active", false)

	if result.Error != nil {
		r.logger.Error("failed to bulk deactivate addons", "count", len(addonIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk deactivate addons", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:service_addons:*")
	}

	r.logger.Info("bulk deactivated addons", "count", result.RowsAffected)
	return nil
}

// BulkDelete deletes multiple addons
func (r *serviceAddonRepository) BulkDelete(ctx context.Context, addonIDs []uuid.UUID) error {
	if len(addonIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Where("id IN ?", addonIDs).
		Delete(&models.ServiceAddon{})

	if result.Error != nil {
		r.logger.Error("failed to bulk delete addons", "count", len(addonIDs), "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to bulk delete addons", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:service_addons:*")
	}

	r.logger.Info("bulk deleted addons", "count", result.RowsAffected)
	return nil
}
