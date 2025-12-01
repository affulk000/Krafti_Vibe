package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PromoCodeRepository interface {
	BaseRepository[models.PromoCode]

	// Core Operations
	GetByCode(ctx context.Context, code string) (*models.PromoCode, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error)
	GetPlatformWideCodes(ctx context.Context, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error)

	// Validation & Usage
	ValidateCode(ctx context.Context, code string, tenantID *uuid.UUID, amount float64, serviceID *uuid.UUID, artisanID *uuid.UUID) (*models.PromoCode, float64, error)
	IncrementUsage(ctx context.Context, promoCodeID uuid.UUID) error
	IncrementUserUsage(ctx context.Context, promoCodeID, userID uuid.UUID) error
	GetUserUsageCount(ctx context.Context, promoCodeID, userID uuid.UUID) (int, error)
	CanUserUsePromoCode(ctx context.Context, promoCodeID, userID uuid.UUID) (bool, error)

	// Status Operations
	Activate(ctx context.Context, promoCodeID uuid.UUID) error
	Deactivate(ctx context.Context, promoCodeID uuid.UUID) error
	UpdateUsageCount(ctx context.Context, promoCodeID uuid.UUID, count int) error

	// Query Operations
	GetActivePromoCodes(ctx context.Context, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error)
	GetExpiredPromoCodes(ctx context.Context, tenantID uuid.UUID) ([]*models.PromoCode, error)
	GetExpiringPromoCodes(ctx context.Context, tenantID uuid.UUID, days int) ([]*models.PromoCode, error)
	GetValidPromoCodesForService(ctx context.Context, serviceID uuid.UUID) ([]*models.PromoCode, error)
	GetValidPromoCodesForArtisan(ctx context.Context, artisanID uuid.UUID) ([]*models.PromoCode, error)

	// Analytics
	GetPromoCodeStats(ctx context.Context, promoCodeID uuid.UUID) (PromoCodeStats, error)
	GetTopPerformingPromoCodes(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]PromoCodePerformance, error)
	GetPromoCodeUsageByPeriod(ctx context.Context, promoCodeID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]PromoCodeUsageData, error)

	// Search & Filter
	Search(ctx context.Context, query string, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error)
	FindByFilters(ctx context.Context, filters PromoCodeFilters, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error)

	// Bulk Operations
	BulkActivate(ctx context.Context, promoCodeIDs []uuid.UUID) error
	BulkDeactivate(ctx context.Context, promoCodeIDs []uuid.UUID) error
	BulkDelete(ctx context.Context, promoCodeIDs []uuid.UUID) error
}

// PromoCodeStats represents statistics for a promo code
type PromoCodeStats struct {
	PromoCodeID     uuid.UUID `json:"promo_code_id"`
	Code            string    `json:"code"`
	TotalUses       int       `json:"total_uses"`
	UniqueUsers     int       `json:"unique_users"`
	TotalDiscount   float64   `json:"total_discount"`
	AverageDiscount float64   `json:"average_discount"`
	TotalRevenue    float64   `json:"total_revenue"`
	ConversionRate  float64   `json:"conversion_rate"`
	RemainingUses   int       `json:"remaining_uses"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
}

// PromoCodePerformance represents performance metrics
type PromoCodePerformance struct {
	PromoCodeID    uuid.UUID `json:"promo_code_id"`
	Code           string    `json:"code"`
	UsageCount     int       `json:"usage_count"`
	TotalDiscount  float64   `json:"total_discount"`
	TotalRevenue   float64   `json:"total_revenue"`
	UniqueUsers    int       `json:"unique_users"`
	ConversionRate float64   `json:"conversion_rate"`
}

// PromoCodeUsageData represents usage data for a period
type PromoCodeUsageData struct {
	Period      time.Time `json:"period"`
	UsageCount  int       `json:"usage_count"`
	Discount    float64   `json:"discount"`
	Revenue     float64   `json:"revenue"`
	UniqueUsers int       `json:"unique_users"`
}

// PromoCodeFilters for advanced filtering
type PromoCodeFilters struct {
	TenantID       *uuid.UUID           `json:"tenant_id"`
	Type           *models.DiscountType `json:"type"`
	IsActive       *bool                `json:"is_active"`
	MinValue       *float64             `json:"min_value"`
	MaxValue       *float64             `json:"max_value"`
	ServiceIDs     []uuid.UUID          `json:"service_ids"`
	ArtisanIDs     []uuid.UUID          `json:"artisan_ids"`
	ValidNow       bool                 `json:"valid_now"`
	ExpiringWithin *int                 `json:"expiring_within"` // days
}

type promoCodeRepository struct {
	BaseRepository[models.PromoCode]
	db      *gorm.DB
	logger  log.AllLogger
	cache   Cache
	metrics MetricsCollector
}

func NewPromoCodeRepository(db *gorm.DB, config ...RepositoryConfig) PromoCodeRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 5 * time.Minute
	}

	baseRepo := NewBaseRepository[models.PromoCode](db, cfg)

	return &promoCodeRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
		metrics:        cfg.Metrics,
	}
}

//------------------------------------------------------------
// Core Operations
//------------------------------------------------------------

func (r *promoCodeRepository) GetByCode(ctx context.Context, code string) (*models.PromoCode, error) {
	var promoCode models.PromoCode
	if err := r.db.WithContext(ctx).
		Where("UPPER(code) = UPPER(?)", code).
		First(&promoCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "promo code not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find promo code", err)
	}

	return &promoCode, nil
}

func (r *promoCodeRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.PromoCode{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count promo codes", err)
	}

	var promoCodes []*models.PromoCode
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&promoCodes).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find promo codes", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return promoCodes, paginationResult, nil
}

func (r *promoCodeRepository) GetPlatformWideCodes(ctx context.Context, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.PromoCode{}).Where("tenant_id IS NULL")

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count promo codes", err)
	}

	var promoCodes []*models.PromoCode
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&promoCodes).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find promo codes", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return promoCodes, paginationResult, nil
}

//------------------------------------------------------------
// Validation & Usage
//------------------------------------------------------------

func (r *promoCodeRepository) ValidateCode(ctx context.Context, code string, tenantID *uuid.UUID, amount float64, serviceID *uuid.UUID, artisanID *uuid.UUID) (*models.PromoCode, float64, error) {
	promoCode, err := r.GetByCode(ctx, code)
	if err != nil {
		return nil, 0, err
	}

	// Check if promo code is valid
	if !promoCode.IsValid() {
		return nil, 0, errors.NewRepositoryError("INVALID_PROMO", "promo code is not valid", errors.ErrInvalidInput)
	}

	// Check tenant applicability
	if promoCode.TenantID != nil && tenantID != nil && *promoCode.TenantID != *tenantID {
		return nil, 0, errors.NewRepositoryError("INVALID_PROMO", "promo code not applicable for this tenant", errors.ErrInvalidInput)
	}

	// Check minimum order amount
	if amount < promoCode.MinOrderAmount {
		return nil, 0, errors.NewRepositoryError("INVALID_PROMO", fmt.Sprintf("minimum order amount is %.2f", promoCode.MinOrderAmount), errors.ErrInvalidInput)
	}

	// Check service restriction
	if len(promoCode.ApplicableServices) > 0 && serviceID != nil {

		if !slices.Contains(promoCode.ApplicableServices, *serviceID) {
			return nil, 0, errors.NewRepositoryError("INVALID_PROMO", "promo code not applicable for this service", errors.ErrInvalidInput)
		}
	}

	// Check artisan restriction
	if len(promoCode.ApplicableArtisans) > 0 && artisanID != nil {

		if !slices.Contains(promoCode.ApplicableArtisans, *artisanID) {
			return nil, 0, errors.NewRepositoryError("INVALID_PROMO", "promo code not applicable for this artisan", errors.ErrInvalidInput)
		}
	}

	discount := promoCode.CalculateDiscount(amount)
	return promoCode, discount, nil
}

func (r *promoCodeRepository) IncrementUsage(ctx context.Context, promoCodeID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.PromoCode{}).
		Where("id = ?", promoCodeID).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", 1))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment usage", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "promo code not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, promoCodeID)
	return nil
}

func (r *promoCodeRepository) IncrementUserUsage(ctx context.Context, promoCodeID, userID uuid.UUID) error {
	// This would typically be tracked in a separate promo_code_usage table
	// For now, we'll just increment the main usage count
	return r.IncrementUsage(ctx, promoCodeID)
}

func (r *promoCodeRepository) GetUserUsageCount(ctx context.Context, promoCodeID, userID uuid.UUID) (int, error) {
	// This would query a promo_code_usage table
	// Placeholder implementation
	return 0, nil
}

func (r *promoCodeRepository) CanUserUsePromoCode(ctx context.Context, promoCodeID, userID uuid.UUID) (bool, error) {
	promoCode, err := r.GetByID(ctx, promoCodeID)
	if err != nil {
		return false, err
	}

	if promoCode.MaxUsesPerUser <= 0 {
		return true, nil
	}

	usageCount, err := r.GetUserUsageCount(ctx, promoCodeID, userID)
	if err != nil {
		return false, err
	}

	return usageCount < promoCode.MaxUsesPerUser, nil
}

//------------------------------------------------------------
// Status Operations
//------------------------------------------------------------

func (r *promoCodeRepository) Activate(ctx context.Context, promoCodeID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.PromoCode{}).
		Where("id = ?", promoCodeID).
		Update("is_active", true)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to activate promo code", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "promo code not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, promoCodeID)
	return nil
}

func (r *promoCodeRepository) Deactivate(ctx context.Context, promoCodeID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.PromoCode{}).
		Where("id = ?", promoCodeID).
		Update("is_active", false)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to deactivate promo code", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "promo code not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, promoCodeID)
	return nil
}

func (r *promoCodeRepository) UpdateUsageCount(ctx context.Context, promoCodeID uuid.UUID, count int) error {
	result := r.db.WithContext(ctx).
		Model(&models.PromoCode{}).
		Where("id = ?", promoCodeID).
		Update("used_count", count)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update usage count", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "promo code not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, promoCodeID)
	return nil
}

//------------------------------------------------------------
// Query Operations
//------------------------------------------------------------

func (r *promoCodeRepository) GetActivePromoCodes(ctx context.Context, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.PromoCode{}).Where("is_active = ?", true)

	if tenantID != nil {
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count promo codes", err)
	}

	var promoCodes []*models.PromoCode
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&promoCodes).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find promo codes", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return promoCodes, paginationResult, nil
}

func (r *promoCodeRepository) GetExpiredPromoCodes(ctx context.Context, tenantID uuid.UUID) ([]*models.PromoCode, error) {
	now := time.Now()

	var promoCodes []*models.PromoCode
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND expires_at IS NOT NULL AND expires_at < ?", tenantID, now).
		Order("expires_at DESC").
		Find(&promoCodes).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find expired promo codes", err)
	}

	return promoCodes, nil
}

func (r *promoCodeRepository) GetExpiringPromoCodes(ctx context.Context, tenantID uuid.UUID, days int) ([]*models.PromoCode, error) {
	now := time.Now()
	deadline := now.AddDate(0, 0, days)

	var promoCodes []*models.PromoCode
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ? AND expires_at IS NOT NULL AND expires_at BETWEEN ? AND ?",
			tenantID, true, now, deadline).
		Order("expires_at ASC").
		Find(&promoCodes).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find expiring promo codes", err)
	}

	return promoCodes, nil
}

func (r *promoCodeRepository) GetValidPromoCodesForService(ctx context.Context, serviceID uuid.UUID) ([]*models.PromoCode, error) {
	now := time.Now()

	var promoCodes []*models.PromoCode
	if err := r.db.WithContext(ctx).
		Where("is_active = ? AND starts_at <= ? AND (expires_at IS NULL OR expires_at > ?) AND (? = ANY(applicable_services) OR CARDINALITY(applicable_services) = 0)",
			true, now, now, serviceID).
		Order("value DESC").
		Find(&promoCodes).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find valid promo codes", err)
	}

	return promoCodes, nil
}

func (r *promoCodeRepository) GetValidPromoCodesForArtisan(ctx context.Context, artisanID uuid.UUID) ([]*models.PromoCode, error) {
	now := time.Now()

	var promoCodes []*models.PromoCode
	if err := r.db.WithContext(ctx).
		Where("is_active = ? AND starts_at <= ? AND (expires_at IS NULL OR expires_at > ?) AND (? = ANY(applicable_artisans) OR CARDINALITY(applicable_artisans) = 0)",
			true, now, now, artisanID).
		Order("value DESC").
		Find(&promoCodes).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find valid promo codes", err)
	}

	return promoCodes, nil
}

//------------------------------------------------------------
// Analytics
//------------------------------------------------------------

func (r *promoCodeRepository) GetPromoCodeStats(ctx context.Context, promoCodeID uuid.UUID) (PromoCodeStats, error) {
	promoCode, err := r.GetByID(ctx, promoCodeID)
	if err != nil {
		return PromoCodeStats{}, err
	}

	stats := PromoCodeStats{
		PromoCodeID: promoCodeID,
		Code:        promoCode.Code,
		TotalUses:   promoCode.UsedCount,
	}

	if promoCode.MaxUses > 0 {
		stats.RemainingUses = max(0, promoCode.MaxUses-promoCode.UsedCount)

	} else {
		stats.RemainingUses = -1 // unlimited
	}

	if promoCode.ExpiresAt != nil {
		days := time.Until(*promoCode.ExpiresAt).Hours() / 24
		stats.DaysUntilExpiry = int(days)
	} else {
		stats.DaysUntilExpiry = -1 // no expiry
	}

	// Additional stats would come from bookings/payments tables
	// Placeholder values
	stats.UniqueUsers = promoCode.UsedCount
	stats.TotalDiscount = 0
	stats.AverageDiscount = 0
	stats.TotalRevenue = 0
	stats.ConversionRate = 0

	return stats, nil
}

func (r *promoCodeRepository) GetTopPerformingPromoCodes(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]PromoCodePerformance, error) {
	// This would join with bookings table
	// Placeholder implementation
	var results []PromoCodePerformance
	return results, nil
}

func (r *promoCodeRepository) GetPromoCodeUsageByPeriod(ctx context.Context, promoCodeID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]PromoCodeUsageData, error) {
	// This would aggregate from bookings/usage table
	// Placeholder implementation
	var results []PromoCodeUsageData
	return results, nil
}

//------------------------------------------------------------
// Search & Filter
//------------------------------------------------------------

func (r *promoCodeRepository) Search(ctx context.Context, query string, tenantID *uuid.UUID, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error) {
	pagination.Validate()

	like := fmt.Sprintf("%%%s%%", strings.TrimSpace(query))

	countQuery := r.db.WithContext(ctx).Model(&models.PromoCode{})
	if tenantID != nil {
		countQuery = countQuery.Where("tenant_id = ?", *tenantID)
	}
	if query != "" {
		countQuery = countQuery.Where("UPPER(code) ILIKE ? OR description ILIKE ?", strings.ToUpper(like), like)
	}

	var totalItems int64
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count promo codes", err)
	}

	dataQuery := r.db.WithContext(ctx).Model(&models.PromoCode{})
	if tenantID != nil {
		dataQuery = dataQuery.Where("tenant_id = ?", *tenantID)
	}
	if query != "" {
		dataQuery = dataQuery.Where("UPPER(code) ILIKE ? OR description ILIKE ?", strings.ToUpper(like), like)
	}

	var promoCodes []*models.PromoCode
	if err := dataQuery.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&promoCodes).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search promo codes", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return promoCodes, paginationResult, nil
}

func (r *promoCodeRepository) FindByFilters(ctx context.Context, filters PromoCodeFilters, pagination PaginationParams) ([]*models.PromoCode, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.PromoCode{})
	query = r.applyPromoCodeFilters(query, filters)

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count promo codes", err)
	}

	var promoCodes []*models.PromoCode
	if err := r.applyPromoCodeFilters(r.db.WithContext(ctx).Model(&models.PromoCode{}), filters).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&promoCodes).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to apply filters", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return promoCodes, paginationResult, nil
}

//------------------------------------------------------------
// Bulk Operations
//------------------------------------------------------------

func (r *promoCodeRepository) BulkActivate(ctx context.Context, promoCodeIDs []uuid.UUID) error {
	if len(promoCodeIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.PromoCode{}).
		Where("id IN ?", promoCodeIDs).
		Update("is_active", true)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk activate promo codes", result.Error)
	}

	for _, id := range promoCodeIDs {
		r.InvalidateCache(ctx, id)
	}

	return nil
}

func (r *promoCodeRepository) BulkDeactivate(ctx context.Context, promoCodeIDs []uuid.UUID) error {
	if len(promoCodeIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.PromoCode{}).
		Where("id IN ?", promoCodeIDs).
		Update("is_active", false)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk deactivate promo codes", result.Error)
	}

	for _, id := range promoCodeIDs {
		r.InvalidateCache(ctx, id)
	}

	return nil
}

func (r *promoCodeRepository) BulkDelete(ctx context.Context, promoCodeIDs []uuid.UUID) error {
	if len(promoCodeIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Where("id IN ?", promoCodeIDs).
		Delete(&models.PromoCode{})

	if result.Error != nil {
		return errors.NewRepositoryError("DELETE_FAILED", "failed to bulk delete promo codes", result.Error)
	}

	for _, id := range promoCodeIDs {
		r.InvalidateCache(ctx, id)
	}

	return nil
}

//------------------------------------------------------------
// Helper Methods
//------------------------------------------------------------

func (r *promoCodeRepository) applyPromoCodeFilters(query *gorm.DB, filters PromoCodeFilters) *gorm.DB {
	if filters.TenantID != nil {
		query = query.Where("tenant_id = ?", *filters.TenantID)
	}

	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	if filters.MinValue != nil {
		query = query.Where("value >= ?", *filters.MinValue)
	}

	if filters.MaxValue != nil {
		query = query.Where("value <= ?", *filters.MaxValue)
	}

	if len(filters.ServiceIDs) > 0 {
		query = query.Where("applicable_services && ?", filters.ServiceIDs)
	}

	if len(filters.ArtisanIDs) > 0 {
		query = query.Where("applicable_artisans && ?", filters.ArtisanIDs)
	}

	if filters.ValidNow {
		now := time.Now()
		query = query.Where("is_active = ? AND starts_at <= ? AND (expires_at IS NULL OR expires_at > ?)", true, now, now)
	}

	if filters.ExpiringWithin != nil {
		now := time.Now()
		deadline := now.AddDate(0, 0, *filters.ExpiringWithin)
		query = query.Where("expires_at IS NOT NULL AND expires_at BETWEEN ? AND ?", now, deadline)
	}

	return query
}

func (r *promoCodeRepository) InvalidateCache(ctx context.Context, promoCodeID uuid.UUID) error {
	if r.cache != nil {
		cacheKey := fmt.Sprintf("promo_code:%s", promoCodeID.String())
		if err := r.cache.Delete(ctx, cacheKey); err != nil {
			if r.logger != nil {
				r.logger.Warnf("Failed to invalidate cache for promo code %s: %v", promoCodeID, err)
			}
			return err
		}
	}
	return nil
}
