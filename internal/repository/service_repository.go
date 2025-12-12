package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository/types"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceRepository defines the interface for service repository operations
type ServiceRepository interface {
	BaseRepository[models.Service]

	// Query Operations
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Service, PaginationResult, error)
	FindByCategory(ctx context.Context, tenantID uuid.UUID, category models.ServiceCategory, pagination PaginationParams) ([]*models.Service, PaginationResult, error)
	FindByArtisanID(ctx context.Context, artisanID uuid.UUID) ([]*models.Service, error)
	FindActiveServices(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Service, PaginationResult, error)
	FindByTags(ctx context.Context, tenantID uuid.UUID, tags []string, pagination PaginationParams) ([]*models.Service, PaginationResult, error)
	FindByPriceRange(ctx context.Context, tenantID uuid.UUID, minPrice, maxPrice float64, pagination PaginationParams) ([]*models.Service, PaginationResult, error)
	FindOrganizationWideServices(ctx context.Context, tenantID uuid.UUID) ([]*models.Service, error)

	// Search & Discovery
	Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.Service, PaginationResult, error)
	FindPopularServices(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Service, error)
	FindRecommendedServices(ctx context.Context, tenantID uuid.UUID, customerID uuid.UUID, limit int) ([]*models.Service, error)

	// Availability Management
	ActivateService(ctx context.Context, serviceID uuid.UUID) error
	DeactivateService(ctx context.Context, serviceID uuid.UUID) error
	UpdateAvailability(ctx context.Context, serviceID uuid.UUID, isActive bool) error

	// Pricing Management
	UpdatePrice(ctx context.Context, serviceID uuid.UUID, newPrice float64) error
	UpdateDeposit(ctx context.Context, serviceID uuid.UUID, depositAmount float64) error
	BulkUpdatePrices(ctx context.Context, serviceIDs []uuid.UUID, priceAdjustment float64, isPercentage bool) error

	// Addon Management
	AddServiceAddon(ctx context.Context, serviceID, addonID uuid.UUID) error
	RemoveServiceAddon(ctx context.Context, serviceID, addonID uuid.UUID) error
	GetServiceAddons(ctx context.Context, serviceID uuid.UUID) ([]*models.ServiceAddon, error)

	// Analytics & Statistics
	GetServiceStats(ctx context.Context, tenantID uuid.UUID) (types.ServiceStats, error)
	GetCategoryStats(ctx context.Context, tenantID uuid.UUID) ([]types.CategoryStats, error)
	GetServicePerformance(ctx context.Context, serviceID uuid.UUID) (types.ServicePerformance, error)
	GetRevenueByService(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]types.ServiceRevenue, error)
	GetServiceBookingTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]types.ServiceTrend, error)

	// Bulk Operations
	BulkActivate(ctx context.Context, serviceIDs []uuid.UUID) error
	BulkDeactivate(ctx context.Context, serviceIDs []uuid.UUID) error
	BulkUpdateCategory(ctx context.Context, serviceIDs []uuid.UUID, category models.ServiceCategory) error
	BulkDelete(ctx context.Context, serviceIDs []uuid.UUID) error

	// Advanced Filtering
	FindByFilters(ctx context.Context, tenantID uuid.UUID, filters types.ServiceFilters, pagination PaginationParams) ([]*models.Service, PaginationResult, error)
	GetCategoriesWithCount(ctx context.Context, tenantID uuid.UUID) ([]types.CategoryCount, error)
}

// serviceRepository implements ServiceRepository
type serviceRepository struct {
	BaseRepository[models.Service]
	db     *gorm.DB
	logger log.AllLogger
	cache  Cache
}

// NewServiceRepository creates a new ServiceRepository instance
func NewServiceRepository(db *gorm.DB, config ...RepositoryConfig) ServiceRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.Service](db, cfg)

	return &serviceRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
	}
}

// FindByTenantID retrieves all services for a tenant
func (r *serviceRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Service, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count services", err)
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Addons").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&services).Error; err != nil {
		r.logger.Error("failed to find services", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find services", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return services, paginationResult, nil
}

// FindByCategory retrieves services by category
func (r *serviceRepository) FindByCategory(ctx context.Context, tenantID uuid.UUID, category models.ServiceCategory, pagination PaginationParams) ([]*models.Service, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ? AND category = ?", tenantID, category).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count services", err)
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Addons").
		Where("tenant_id = ? AND category = ?", tenantID, category).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("name ASC").
		Find(&services).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find services by category", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return services, paginationResult, nil
}

// FindByArtisanID retrieves all services offered by an artisan
func (r *serviceRepository) FindByArtisanID(ctx context.Context, artisanID uuid.UUID) ([]*models.Service, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Addons").
		Where("artisan_id = ?", artisanID).
		Order("is_active DESC, name ASC").
		Find(&services).Error; err != nil {
		r.logger.Error("failed to find services by artisan", "artisan_id", artisanID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find services", err)
	}

	return services, nil
}

// FindActiveServices retrieves all active services for a tenant
func (r *serviceRepository) FindActiveServices(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Service, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count active services", err)
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Addons").
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("category ASC, name ASC").
		Find(&services).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find active services", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return services, paginationResult, nil
}

// FindByTags retrieves services by tags
func (r *serviceRepository) FindByTags(ctx context.Context, tenantID uuid.UUID, tags []string, pagination PaginationParams) ([]*models.Service, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}
	if len(tags) == 0 {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tags cannot be empty", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ? AND tags && ?", tenantID, tags).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count services", err)
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Addons").
		Where("tenant_id = ? AND tags && ?", tenantID, tags).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("name ASC").
		Find(&services).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find services by tags", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return services, paginationResult, nil
}

// FindByPriceRange retrieves services within a price range
func (r *serviceRepository) FindByPriceRange(ctx context.Context, tenantID uuid.UUID, minPrice, maxPrice float64, pagination PaginationParams) ([]*models.Service, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ? AND price BETWEEN ? AND ?", tenantID, minPrice, maxPrice).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count services", err)
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Addons").
		Where("tenant_id = ? AND price BETWEEN ? AND ?", tenantID, minPrice, maxPrice).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("price ASC").
		Find(&services).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find services by price range", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return services, paginationResult, nil
}

// FindOrganizationWideServices retrieves services not tied to a specific artisan
func (r *serviceRepository) FindOrganizationWideServices(ctx context.Context, tenantID uuid.UUID) ([]*models.Service, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Addons").
		Where("tenant_id = ? AND artisan_id IS NULL", tenantID).
		Order("category ASC, name ASC").
		Find(&services).Error; err != nil {
		r.logger.Error("failed to find org-wide services", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find organization-wide services", err)
	}

	return services, nil
}

// Search searches services by name or description
func (r *serviceRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.Service, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ? AND (name ILIKE ? OR description ILIKE ?)",
			tenantID, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count services", err)
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Addons").
		Where("tenant_id = ? AND (name ILIKE ? OR description ILIKE ?)",
			tenantID, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("name ASC").
		Find(&services).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search services", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return services, paginationResult, nil
}

// FindPopularServices retrieves most booked services
func (r *serviceRepository) FindPopularServices(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Service, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if limit <= 0 {
		limit = 10
	}

	var services []*models.Service
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Addons").
		Joins("LEFT JOIN bookings ON bookings.service_id = services.id").
		Where("services.tenant_id = ? AND services.is_active = ?", tenantID, true).
		Group("services.id").
		Order("COUNT(bookings.id) DESC").
		Limit(limit).
		Find(&services).Error; err != nil {
		r.logger.Error("failed to find popular services", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find popular services", err)
	}

	return services, nil
}

// FindRecommendedServices retrieves recommended services for a customer
func (r *serviceRepository) FindRecommendedServices(ctx context.Context, tenantID uuid.UUID, customerID uuid.UUID, limit int) ([]*models.Service, error) {
	if tenantID == uuid.Nil || customerID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id and customer_id cannot be nil", errors.ErrInvalidInput)
	}

	if limit <= 0 {
		limit = 5
	}

	// Get categories the customer has booked before
	var bookedCategories []models.ServiceCategory
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Select("DISTINCT services.category").
		Joins("JOIN services ON services.id = bookings.service_id").
		Where("bookings.customer_id = ? AND bookings.tenant_id = ?", customerID, tenantID).
		Pluck("category", &bookedCategories)

	// Recommend services from those categories
	var services []*models.Service
	query := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Addons").
		Where("tenant_id = ? AND is_active = ?", tenantID, true)

	if len(bookedCategories) > 0 {
		query = query.Where("category IN ?", bookedCategories)
	}

	if err := query.
		Order("RANDOM()").
		Limit(limit).
		Find(&services).Error; err != nil {
		r.logger.Error("failed to find recommended services", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find recommended services", err)
	}

	return services, nil
}

// ActivateService activates a service
func (r *serviceRepository) ActivateService(ctx context.Context, serviceID uuid.UUID) error {
	return r.UpdateAvailability(ctx, serviceID, true)
}

// DeactivateService deactivates a service
func (r *serviceRepository) DeactivateService(ctx context.Context, serviceID uuid.UUID) error {
	return r.UpdateAvailability(ctx, serviceID, false)
}

// UpdateAvailability updates service availability
func (r *serviceRepository) UpdateAvailability(ctx context.Context, serviceID uuid.UUID, isActive bool) error {
	if serviceID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "service_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("id = ?", serviceID).
		Update("is_active", isActive)

	if result.Error != nil {
		r.logger.Error("failed to update service availability", "service_id", serviceID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update service availability", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "service not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	return nil
}

// UpdatePrice updates service price
func (r *serviceRepository) UpdatePrice(ctx context.Context, serviceID uuid.UUID, newPrice float64) error {
	if serviceID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "service_id cannot be nil", errors.ErrInvalidInput)
	}
	if newPrice < 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "price cannot be negative", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("id = ?", serviceID).
		Update("price", newPrice)

	if result.Error != nil {
		r.logger.Error("failed to update service price", "service_id", serviceID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update service price", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "service not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	r.logger.Info("service price updated", "service_id", serviceID, "new_price", newPrice)
	return nil
}

// UpdateDeposit updates service deposit amount
func (r *serviceRepository) UpdateDeposit(ctx context.Context, serviceID uuid.UUID, depositAmount float64) error {
	if serviceID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "service_id cannot be nil", errors.ErrInvalidInput)
	}
	if depositAmount < 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "deposit amount cannot be negative", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"deposit_amount":   depositAmount,
		"requires_deposit": depositAmount > 0,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("id = ?", serviceID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to update service deposit", "service_id", serviceID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update service deposit", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "service not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	return nil
}

// BulkUpdatePrices updates prices for multiple services
func (r *serviceRepository) BulkUpdatePrices(ctx context.Context, serviceIDs []uuid.UUID, priceAdjustment float64, isPercentage bool) error {
	if len(serviceIDs) == 0 {
		return nil
	}

	var result *gorm.DB
	if isPercentage {
		// Update by percentage - cast to numeric to ensure decimal division
		result = r.db.WithContext(ctx).
			Model(&models.Service{}).
			Where("id IN ?", serviceIDs).
			Update("price", gorm.Expr("price * (1 + ?::numeric / 100)", priceAdjustment))
	} else {
		// Update by fixed amount
		result = r.db.WithContext(ctx).
			Model(&models.Service{}).
			Where("id IN ?", serviceIDs).
			Update("price", gorm.Expr("price + ?", priceAdjustment))
	}

	if result.Error != nil {
		r.logger.Error("failed to bulk update prices", "count", len(serviceIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update prices", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	r.logger.Info("bulk updated service prices", "count", result.RowsAffected, "adjustment", priceAdjustment, "is_percentage", isPercentage)
	return nil
}

// AddServiceAddon adds an addon to a service
func (r *serviceRepository) AddServiceAddon(ctx context.Context, serviceID, addonID uuid.UUID) error {
	if serviceID == uuid.Nil || addonID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "service_id and addon_id cannot be nil", errors.ErrInvalidInput)
	}

	// Use association to add the addon
	var service models.Service
	if err := r.db.WithContext(ctx).First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "service not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find service", err)
	}

	var addon models.ServiceAddon
	if err := r.db.WithContext(ctx).First(&addon, addonID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "addon not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find addon", err)
	}

	if err := r.db.WithContext(ctx).Model(&service).Association("Addons").Append(&addon); err != nil {
		r.logger.Error("failed to add service addon", "service_id", serviceID, "addon_id", addonID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to add service addon", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	r.logger.Info("added service addon", "service_id", serviceID, "addon_id", addonID)
	return nil
}

// RemoveServiceAddon removes an addon from a service
func (r *serviceRepository) RemoveServiceAddon(ctx context.Context, serviceID, addonID uuid.UUID) error {
	if serviceID == uuid.Nil || addonID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "service_id and addon_id cannot be nil", errors.ErrInvalidInput)
	}

	var service models.Service
	if err := r.db.WithContext(ctx).First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "service not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find service", err)
	}

	var addon models.ServiceAddon
	if err := r.db.WithContext(ctx).First(&addon, addonID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "addon not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find addon", err)
	}

	if err := r.db.WithContext(ctx).Model(&service).Association("Addons").Delete(&addon); err != nil {
		r.logger.Error("failed to remove service addon", "service_id", serviceID, "addon_id", addonID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to remove service addon", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	r.logger.Info("removed service addon", "service_id", serviceID, "addon_id", addonID)
	return nil
}

// GetServiceAddons retrieves all addons for a service
func (r *serviceRepository) GetServiceAddons(ctx context.Context, serviceID uuid.UUID) ([]*models.ServiceAddon, error) {
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

// GetServiceStats retrieves comprehensive service statistics
func (r *serviceRepository) GetServiceStats(ctx context.Context, tenantID uuid.UUID) (types.ServiceStats, error) {
	if tenantID == uuid.Nil {
		return types.ServiceStats{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := types.ServiceStats{
		ByCategory: make(map[models.ServiceCategory]int64),
	}

	// Total services
	r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalServices)

	// Active services
	r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Count(&stats.ActiveServices)

	// Inactive services
	r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ? AND is_active = ?", tenantID, false).
		Count(&stats.InactiveServices)

	// Services by category
	type CategoryCount struct {
		Category models.ServiceCategory
		Count    int64
	}
	var categoryCounts []CategoryCount
	r.db.WithContext(ctx).
		Model(&models.Service{}).
		Select("category, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("category").
		Scan(&categoryCounts)

	for _, cc := range categoryCounts {
		stats.ByCategory[cc.Category] = cc.Count
	}

	// Most popular category
	if len(categoryCounts) > 0 {
		maxCount := int64(0)
		for _, cc := range categoryCounts {
			if cc.Count > maxCount {
				maxCount = cc.Count
				stats.MostPopularCategory = cc.Category
			}
		}
	}

	// Average duration
	r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ?", tenantID).
		Select("AVG(duration_minutes)").
		Scan(&stats.AverageDuration)

	// Average price
	r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ?", tenantID).
		Select("AVG(price)").
		Scan(&stats.AveragePrice)

	// Total bookings
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Joins("JOIN services ON services.id = bookings.service_id").
		Where("services.tenant_id = ?", tenantID).
		Count(&stats.TotalBookings)

	// Total revenue (from completed bookings)
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Joins("JOIN services ON services.id = bookings.service_id").
		Where("services.tenant_id = ? AND bookings.status = ?", tenantID, "completed").
		Select("COALESCE(SUM(bookings.total_price), 0)").
		Scan(&stats.TotalRevenue)

	// Services with deposit
	r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("tenant_id = ? AND requires_deposit = ?", tenantID, true).
		Count(&stats.ServicesWithDeposit)

	// Highest priced service
	var highestPriced models.Service
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("price DESC").
		First(&highestPriced).Error; err == nil {

		var bookingCount int64
		r.db.WithContext(ctx).
			Model(&models.Booking{}).
			Where("service_id = ?", highestPriced.ID).
			Count(&bookingCount)

		stats.HighestPricedService = &types.ServiceSummary{
			ID:       highestPriced.ID,
			Name:     highestPriced.Name,
			Category: highestPriced.Category,
			Price:    highestPriced.Price,
			Bookings: bookingCount,
		}
	}

	// Most booked service
	type ServiceBookingCount struct {
		ServiceID uuid.UUID
		Count     int64
	}
	var mostBooked ServiceBookingCount
	if err := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Select("service_id, COUNT(*) as count").
		Joins("JOIN services ON services.id = bookings.service_id").
		Where("services.tenant_id = ?", tenantID).
		Group("service_id").
		Order("count DESC").
		First(&mostBooked).Error; err == nil {

		var service models.Service
		if err := r.db.WithContext(ctx).First(&service, mostBooked.ServiceID).Error; err == nil {
			stats.MostBookedService = &types.ServiceSummary{
				ID:       service.ID,
				Name:     service.Name,
				Category: service.Category,
				Price:    service.Price,
				Bookings: mostBooked.Count,
			}
		}
	}

	return stats, nil
}

// GetCategoryStats retrieves statistics per category
func (r *serviceRepository) GetCategoryStats(ctx context.Context, tenantID uuid.UUID) ([]types.CategoryStats, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var stats []types.CategoryStats
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			s.category,
			COUNT(DISTINCT s.id) as service_count,
			COUNT(DISTINCT b.id) as total_bookings,
			COALESCE(SUM(CASE WHEN b.status = 'completed' THEN b.total_price ELSE 0 END), 0) as total_revenue,
			AVG(s.price) as average_price,
			COUNT(DISTINCT CASE WHEN s.is_active THEN s.id END) as active_count
		FROM services s
		LEFT JOIN bookings b ON b.service_id = s.id
		WHERE s.tenant_id = ?
		GROUP BY s.category
		ORDER BY total_bookings DESC
	`, tenantID).Scan(&stats).Error; err != nil {
		r.logger.Error("failed to get category stats", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get category stats", err)
	}

	return stats, nil
}

// GetServicePerformance retrieves performance metrics for a service
func (r *serviceRepository) GetServicePerformance(ctx context.Context, serviceID uuid.UUID) (types.ServicePerformance, error) {
	if serviceID == uuid.Nil {
		return types.ServicePerformance{}, errors.NewRepositoryError("INVALID_INPUT", "service_id cannot be nil", errors.ErrInvalidInput)
	}

	perf := types.ServicePerformance{
		ServiceID: serviceID,
	}

	// Get service details
	var service models.Service
	if err := r.db.WithContext(ctx).First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return perf, errors.NewRepositoryError("NOT_FOUND", "service not found", errors.ErrNotFound)
		}
		return perf, errors.NewRepositoryError("FIND_FAILED", "failed to find service", err)
	}
	perf.ServiceName = service.Name

	// Total bookings
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("service_id = ?", serviceID).
		Count(&perf.TotalBookings)

	// Completed bookings
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("service_id = ? AND status = ?", serviceID, "completed").
		Count(&perf.CompletedBookings)

	// Cancelled bookings
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("service_id = ? AND status = ?", serviceID, "cancelled").
		Count(&perf.CancelledBookings)

	// Total revenue
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("service_id = ? AND status = ?", serviceID, "completed").
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&perf.TotalRevenue)

	// Average rating
	r.db.WithContext(ctx).
		Model(&models.Review{}).
		Joins("JOIN bookings ON bookings.id = reviews.booking_id").
		Where("bookings.service_id = ?", serviceID).
		Select("AVG(rating)").
		Scan(&perf.AverageRating)

	// Review count
	r.db.WithContext(ctx).
		Model(&models.Review{}).
		Joins("JOIN bookings ON bookings.id = reviews.booking_id").
		Where("bookings.service_id = ?", serviceID).
		Count(&perf.ReviewCount)

	// Completion rate
	if perf.TotalBookings > 0 {
		perf.CompletionRate = (float64(perf.CompletedBookings) / float64(perf.TotalBookings)) * 100
		perf.CancellationRate = (float64(perf.CancelledBookings) / float64(perf.TotalBookings)) * 100
	}

	// Bookings this month
	monthAgo := time.Now().AddDate(0, -1, 0)
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("service_id = ? AND created_at >= ?", serviceID, monthAgo).
		Count(&perf.BookingsThisMonth)

	// Revenue this month
	r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("service_id = ? AND status = ? AND created_at >= ?", serviceID, "completed", monthAgo).
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&perf.RevenueThisMonth)

	// Popularity score (simple calculation: bookings * rating)
	perf.PopularityScore = float64(perf.TotalBookings) * perf.AverageRating

	return perf, nil
}

// GetRevenueByService retrieves revenue breakdown by service
func (r *serviceRepository) GetRevenueByService(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]types.ServiceRevenue, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var revenue []types.ServiceRevenue
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			s.id as service_id,
			s.name as service_name,
			s.category,
			COUNT(b.id) as bookings,
			COALESCE(SUM(b.total_price), 0) as revenue,
			COALESCE(AVG(b.total_price), 0) as average_price
		FROM services s
		LEFT JOIN bookings b ON b.service_id = s.id
			AND b.status = 'completed'
			AND b.created_at BETWEEN ? AND ?
		WHERE s.tenant_id = ?
		GROUP BY s.id, s.name, s.category
		ORDER BY revenue DESC
	`, startDate, endDate, tenantID).Scan(&revenue).Error; err != nil {
		r.logger.Error("failed to get revenue by service", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get revenue by service", err)
	}

	return revenue, nil
}

// GetServiceBookingTrends retrieves service booking trends
func (r *serviceRepository) GetServiceBookingTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]types.ServiceTrend, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if days <= 0 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)

	var trends []types.ServiceTrend
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			DATE(b.created_at) as date,
			s.id as service_id,
			s.name as service_name,
			COUNT(b.id) as bookings,
			COALESCE(SUM(CASE WHEN b.status = 'completed' THEN b.total_price ELSE 0 END), 0) as revenue
		FROM bookings b
		JOIN services s ON s.id = b.service_id
		WHERE s.tenant_id = ?
			AND b.created_at >= ?
		GROUP BY DATE(b.created_at), s.id, s.name
		ORDER BY date DESC, bookings DESC
	`, tenantID, startDate).Scan(&trends).Error; err != nil {
		r.logger.Error("failed to get service booking trends", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get service booking trends", err)
	}

	return trends, nil
}

// BulkActivate activates multiple services
func (r *serviceRepository) BulkActivate(ctx context.Context, serviceIDs []uuid.UUID) error {
	if len(serviceIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("id IN ?", serviceIDs).
		Update("is_active", true)

	if result.Error != nil {
		r.logger.Error("failed to bulk activate services", "count", len(serviceIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk activate services", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	r.logger.Info("bulk activated services", "count", result.RowsAffected)
	return nil
}

// BulkDeactivate deactivates multiple services
func (r *serviceRepository) BulkDeactivate(ctx context.Context, serviceIDs []uuid.UUID) error {
	if len(serviceIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("id IN ?", serviceIDs).
		Update("is_active", false)

	if result.Error != nil {
		r.logger.Error("failed to bulk deactivate services", "count", len(serviceIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk deactivate services", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	r.logger.Info("bulk deactivated services", "count", result.RowsAffected)
	return nil
}

// BulkUpdateCategory updates category for multiple services
func (r *serviceRepository) BulkUpdateCategory(ctx context.Context, serviceIDs []uuid.UUID, category models.ServiceCategory) error {
	if len(serviceIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Where("id IN ?", serviceIDs).
		Update("category", category)

	if result.Error != nil {
		r.logger.Error("failed to bulk update category", "count", len(serviceIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update category", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	r.logger.Info("bulk updated service category", "count", result.RowsAffected, "category", category)
	return nil
}

// BulkDelete deletes multiple services
func (r *serviceRepository) BulkDelete(ctx context.Context, serviceIDs []uuid.UUID) error {
	if len(serviceIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Where("id IN ?", serviceIDs).
		Delete(&models.Service{})

	if result.Error != nil {
		r.logger.Error("failed to bulk delete services", "count", len(serviceIDs), "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to bulk delete services", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:services:*")
	}

	r.logger.Info("bulk deleted services", "count", result.RowsAffected)
	return nil
}

// FindByFilters retrieves services using advanced filters
func (r *serviceRepository) FindByFilters(ctx context.Context, tenantID uuid.UUID, filters types.ServiceFilters, pagination PaginationParams) ([]*models.Service, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	// Apply filters
	if len(filters.Categories) > 0 {
		query = query.Where("category IN ?", filters.Categories)
	}

	if filters.MinPrice != nil {
		query = query.Where("price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("price <= ?", *filters.MaxPrice)
	}

	if filters.MinDuration != nil {
		query = query.Where("duration_minutes >= ?", *filters.MinDuration)
	}

	if filters.MaxDuration != nil {
		query = query.Where("duration_minutes <= ?", *filters.MaxDuration)
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	if filters.ArtisanID != nil {
		query = query.Where("artisan_id = ?", *filters.ArtisanID)
	}

	if len(filters.Tags) > 0 {
		query = query.Where("tags && ?", filters.Tags)
	}

	if filters.RequiresDeposit != nil {
		query = query.Where("requires_deposit = ?", *filters.RequiresDeposit)
	}

	// Count total
	var totalItems int64
	if err := query.Model(&models.Service{}).Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count services", err)
	}

	// Find services
	var services []*models.Service
	if err := query.
		Preload("Artisan").
		Preload("Addons").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("name ASC").
		Find(&services).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find services", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return services, paginationResult, nil
}

// GetCategoriesWithCount retrieves all categories with service counts
func (r *serviceRepository) GetCategoriesWithCount(ctx context.Context, tenantID uuid.UUID) ([]types.CategoryCount, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var counts []types.CategoryCount
	if err := r.db.WithContext(ctx).
		Model(&models.Service{}).
		Select("category, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("category").
		Order("count DESC").
		Scan(&counts).Error; err != nil {
		r.logger.Error("failed to get categories with count", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get categories with count", err)
	}

	return counts, nil
}
