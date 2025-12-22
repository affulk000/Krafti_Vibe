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

type CustomerRepository interface {
	BaseRepository[models.Customer]

	// Core Operations
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Customer, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Customer, PaginationResult, error)
	GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*models.Customer, error)

	// Loyalty & Rewards
	AddLoyaltyPoints(ctx context.Context, customerID uuid.UUID, points int) error
	DeductLoyaltyPoints(ctx context.Context, customerID uuid.UUID, points int) error
	GetByLoyaltyTier(ctx context.Context, tenantID uuid.UUID, tier string, pagination PaginationParams) ([]*models.Customer, PaginationResult, error)
	GetTopLoyaltyCustomers(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Customer, error)
	UpdateTotalSpent(ctx context.Context, customerID uuid.UUID, amount float64) error

	// Statistics
	IncrementBookingCount(ctx context.Context, customerID uuid.UUID) error
	IncrementCompletedBookings(ctx context.Context, customerID uuid.UUID) error
	IncrementCancelledBookings(ctx context.Context, customerID uuid.UUID) error
	GetCustomerStats(ctx context.Context, customerID uuid.UUID) (CustomerStats, error)
	UpdateBookingStatistics(ctx context.Context, customerID uuid.UUID, completed, cancelled int) error

	// Preferences
	UpdatePreferredArtisans(ctx context.Context, customerID uuid.UUID, artisanIDs []uuid.UUID) error
	AddPreferredArtisan(ctx context.Context, customerID uuid.UUID, artisanID uuid.UUID) error
	RemovePreferredArtisan(ctx context.Context, customerID uuid.UUID, artisanID uuid.UUID) error
	GetCustomersWithPreferredArtisan(ctx context.Context, artisanID uuid.UUID) ([]*models.Customer, error)

	// Payment Methods
	UpdateDefaultPaymentMethod(ctx context.Context, customerID uuid.UUID, paymentMethodID string) error
	GetCustomersWithPaymentMethod(ctx context.Context, tenantID uuid.UUID) ([]*models.Customer, error)

	// Location
	UpdatePrimaryLocation(ctx context.Context, customerID uuid.UUID, location models.Location) error
	GetCustomersByLocation(ctx context.Context, tenantID uuid.UUID, latitude, longitude, radiusKM float64) ([]*models.Customer, error)

	// Communication Preferences
	UpdateNotificationPreferences(ctx context.Context, customerID uuid.UUID, email, sms, push bool) error
	GetCustomersOptedInForEmail(ctx context.Context, tenantID uuid.UUID) ([]*models.Customer, error)
	GetCustomersOptedInForSMS(ctx context.Context, tenantID uuid.UUID) ([]*models.Customer, error)

	// Analytics & Reporting
	GetCustomerAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (CustomerAnalytics, error)
	GetTopSpendingCustomers(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Customer, error)
	GetMostActiveCustomers(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Customer, error)
	GetCustomerRetentionRate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error)
	GetAverageCustomerValue(ctx context.Context, tenantID uuid.UUID) (float64, error)
	GetCustomerLifetimeValue(ctx context.Context, customerID uuid.UUID) (float64, error)
	GetChurnedCustomers(ctx context.Context, tenantID uuid.UUID, inactiveDays int) ([]*models.Customer, error)

	// Segmentation
	GetCustomersBySpendingRange(ctx context.Context, tenantID uuid.UUID, minSpent, maxSpent float64, pagination PaginationParams) ([]*models.Customer, PaginationResult, error)
	GetCustomersByBookingCount(ctx context.Context, tenantID uuid.UUID, minBookings, maxBookings int, pagination PaginationParams) ([]*models.Customer, PaginationResult, error)
	GetHighValueCustomers(ctx context.Context, tenantID uuid.UUID, minSpent float64) ([]*models.Customer, error)
	GetAtRiskCustomers(ctx context.Context, tenantID uuid.UUID, inactiveDays int, minPreviousBookings int) ([]*models.Customer, error)

	// Search & Filter
	Search(ctx context.Context, query string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Customer, PaginationResult, error)
	FindByFilters(ctx context.Context, filters CustomerFilters, pagination PaginationParams) ([]*models.Customer, PaginationResult, error)

	// Bulk Operations
	BulkUpdateLoyaltyPoints(ctx context.Context, customerIDs []uuid.UUID, points int) error
	BulkUpdateNotificationPreferences(ctx context.Context, customerIDs []uuid.UUID, email, sms, push bool) error
}

// CustomerStats represents customer statistics
type CustomerStats struct {
	CustomerID          uuid.UUID `json:"customer_id"`
	TotalBookings       int       `json:"total_bookings"`
	CompletedBookings   int       `json:"completed_bookings"`
	CancelledBookings   int       `json:"cancelled_bookings"`
	TotalSpent          float64   `json:"total_spent"`
	AverageBookingValue float64   `json:"average_booking_value"`
	LoyaltyPoints       int       `json:"loyalty_points"`
	LoyaltyTier         string    `json:"loyalty_tier"`
	LastBookingDate     time.Time `json:"last_booking_date"`
	DaysSinceLastVisit  int       `json:"days_since_last_visit"`
	CompletionRate      float64   `json:"completion_rate"`
	CancellationRate    float64   `json:"cancellation_rate"`
}

// CustomerAnalytics represents analytics for customers
type CustomerAnalytics struct {
	TotalCustomers       int64            `json:"total_customers"`
	NewCustomers         int64            `json:"new_customers"`
	ActiveCustomers      int64            `json:"active_customers"`
	ChurnedCustomers     int64            `json:"churned_customers"`
	AverageLifetimeValue float64          `json:"average_lifetime_value"`
	AverageBookings      float64          `json:"average_bookings"`
	RetentionRate        float64          `json:"retention_rate"`
	ByLoyaltyTier        map[string]int64 `json:"by_loyalty_tier"`
}

// CustomerFilters for advanced filtering
type CustomerFilters struct {
	TenantID            uuid.UUID   `json:"tenant_id"`
	UserIDs             []uuid.UUID `json:"user_ids"`
	LoyaltyTiers        []string    `json:"loyalty_tiers"`
	MinLoyaltyPoints    *int        `json:"min_loyalty_points"`
	MaxLoyaltyPoints    *int        `json:"max_loyalty_points"`
	MinTotalSpent       *float64    `json:"min_total_spent"`
	MaxTotalSpent       *float64    `json:"max_total_spent"`
	MinBookings         *int        `json:"min_bookings"`
	MaxBookings         *int        `json:"max_bookings"`
	PreferredArtisanIDs []uuid.UUID `json:"preferred_artisan_ids"`
	HasPaymentMethod    *bool       `json:"has_payment_method"`
	EmailNotifications  *bool       `json:"email_notifications"`
	SMSNotifications    *bool       `json:"sms_notifications"`
	CreatedAfter        *time.Time  `json:"created_after"`
	CreatedBefore       *time.Time  `json:"created_before"`
}

type customerRepository struct {
	BaseRepository[models.Customer]
	db      *gorm.DB
	logger  log.AllLogger
	cache   Cache
	metrics MetricsCollector
}

func NewCustomerRepository(db *gorm.DB, config ...RepositoryConfig) CustomerRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 5 * time.Minute
	}

	baseRepo := NewBaseRepository[models.Customer](db, cfg)

	return &customerRepository{
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

func (r *customerRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find customer", err)
	}

	return &customer, nil
}

func (r *customerRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Customer, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Customer{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count customers", err)
	}

	var customers []*models.Customer
	if err := query.
		Preload("User").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&customers).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find customers", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return customers, paginationResult, nil
}

func (r *customerRepository) GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find customer", err)
	}

	return &customer, nil
}

//------------------------------------------------------------
// Loyalty & Rewards
//------------------------------------------------------------

func (r *customerRepository) AddLoyaltyPoints(ctx context.Context, customerID uuid.UUID, points int) error {
	if points <= 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "points must be positive", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		UpdateColumn("loyalty_points", gorm.Expr("loyalty_points + ?", points))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to add loyalty points", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) DeductLoyaltyPoints(ctx context.Context, customerID uuid.UUID, points int) error {
	if points <= 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "points must be positive", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ? AND loyalty_points >= ?", customerID, points).
		UpdateColumn("loyalty_points", gorm.Expr("loyalty_points - ?", points))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to deduct loyalty points", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("INSUFFICIENT_POINTS", "insufficient loyalty points", errors.ErrInvalidInput)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) GetByLoyaltyTier(ctx context.Context, tenantID uuid.UUID, tier string, pagination PaginationParams) ([]*models.Customer, PaginationResult, error) {
	pagination.Validate()

	var minPoints, maxPoints int
	switch strings.ToLower(tier) {
	case "platinum":
		minPoints = 1000
		maxPoints = 999999
	case "gold":
		minPoints = 500
		maxPoints = 999
	case "silver":
		minPoints = 100
		maxPoints = 499
	case "bronze":
		minPoints = 0
		maxPoints = 99
	default:
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "invalid loyalty tier", errors.ErrInvalidInput)
	}

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Customer{}).
		Where("tenant_id = ? AND loyalty_points BETWEEN ? AND ?", tenantID, minPoints, maxPoints)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count customers", err)
	}

	var customers []*models.Customer
	if err := query.
		Preload("User").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("loyalty_points DESC").
		Find(&customers).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find customers", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return customers, paginationResult, nil
}

func (r *customerRepository) GetTopLoyaltyCustomers(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ?", tenantID).
		Order("loyalty_points DESC").
		Limit(limit).
		Find(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find top customers", err)
	}

	return customers, nil
}

func (r *customerRepository) UpdateTotalSpent(ctx context.Context, customerID uuid.UUID, amount float64) error {
	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		UpdateColumn("total_spent", gorm.Expr("total_spent + ?", amount))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update total spent", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

//------------------------------------------------------------
// Statistics
//------------------------------------------------------------

func (r *customerRepository) IncrementBookingCount(ctx context.Context, customerID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		UpdateColumn("total_bookings", gorm.Expr("total_bookings + ?", 1))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment booking count", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) IncrementCompletedBookings(ctx context.Context, customerID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		UpdateColumn("completed_bookings", gorm.Expr("completed_bookings + ?", 1))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment completed bookings", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) IncrementCancelledBookings(ctx context.Context, customerID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		UpdateColumn("cancelled_bookings", gorm.Expr("cancelled_bookings + ?", 1))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment cancelled bookings", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) GetCustomerStats(ctx context.Context, customerID uuid.UUID) (CustomerStats, error) {
	customer, err := r.GetByID(ctx, customerID)
	if err != nil {
		return CustomerStats{}, err
	}

	stats := CustomerStats{
		CustomerID:        customerID,
		TotalBookings:     customer.TotalBookings,
		CompletedBookings: customer.CompletedBookings,
		CancelledBookings: customer.CancelledBookings,
		TotalSpent:        customer.TotalSpent,
		LoyaltyPoints:     customer.LoyaltyPoints,
		LoyaltyTier:       customer.GetLoyaltyTier(),
	}

	if customer.TotalBookings > 0 {
		stats.AverageBookingValue = customer.TotalSpent / float64(customer.TotalBookings)
		stats.CompletionRate = float64(customer.CompletedBookings) / float64(customer.TotalBookings) * 100
		stats.CancellationRate = float64(customer.CancelledBookings) / float64(customer.TotalBookings) * 100
	}

	// Get last booking date
	var lastBooking models.Booking
	if err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("start_time DESC").
		First(&lastBooking).Error; err == nil {
		stats.LastBookingDate = lastBooking.StartTime
		stats.DaysSinceLastVisit = int(time.Since(lastBooking.StartTime).Hours() / 24)
	}

	return stats, nil
}

func (r *customerRepository) UpdateBookingStatistics(ctx context.Context, customerID uuid.UUID, completed, cancelled int) error {
	updates := map[string]any{}

	if completed != 0 {
		updates["completed_bookings"] = gorm.Expr("completed_bookings + ?", completed)
	}

	if cancelled != 0 {
		updates["cancelled_bookings"] = gorm.Expr("cancelled_bookings + ?", cancelled)
	}

	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update statistics", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

//------------------------------------------------------------
// Preferences
//------------------------------------------------------------

func (r *customerRepository) UpdatePreferredArtisans(ctx context.Context, customerID uuid.UUID, artisanIDs []uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		Update("preferred_artisans", artisanIDs)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update preferred artisans", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) AddPreferredArtisan(ctx context.Context, customerID uuid.UUID, artisanID uuid.UUID) error {
	customer, err := r.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	// Check if already exists
	if slices.Contains(customer.PreferredArtisans, artisanID) {
		return nil // Already preferred
	}

	customer.PreferredArtisans = append(customer.PreferredArtisans, artisanID)

	if err := r.Update(ctx, customer); err != nil {
		return err
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) RemovePreferredArtisan(ctx context.Context, customerID uuid.UUID, artisanID uuid.UUID) error {
	customer, err := r.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	filtered := make([]uuid.UUID, 0)
	for _, id := range customer.PreferredArtisans {
		if id != artisanID {
			filtered = append(filtered, id)
		}
	}

	customer.PreferredArtisans = filtered

	if err := r.Update(ctx, customer); err != nil {
		return err
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) GetCustomersWithPreferredArtisan(ctx context.Context, artisanID uuid.UUID) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("? = ANY(preferred_artisans)", artisanID).
		Find(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find customers", err)
	}

	return customers, nil
}

//------------------------------------------------------------
// Payment Methods
//------------------------------------------------------------

func (r *customerRepository) UpdateDefaultPaymentMethod(ctx context.Context, customerID uuid.UUID, paymentMethodID string) error {
	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		Update("default_payment_method_id", paymentMethodID)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update payment method", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) GetCustomersWithPaymentMethod(ctx context.Context, tenantID uuid.UUID) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ? AND default_payment_method_id IS NOT NULL AND default_payment_method_id != ''", tenantID).
		Find(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find customers", err)
	}

	return customers, nil
}

//------------------------------------------------------------
// Location
//------------------------------------------------------------

func (r *customerRepository) UpdatePrimaryLocation(ctx context.Context, customerID uuid.UUID, location models.Location) error {
	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		Update("primary_location", location)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update location", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) GetCustomersByLocation(ctx context.Context, tenantID uuid.UUID, latitude, longitude, radiusKM float64) ([]*models.Customer, error) {
	// Using PostGIS-style distance calculation
	query := `
		SELECT * FROM customers
		WHERE tenant_id = ?
		AND ST_DWithin(
			ST_SetSRID(ST_MakePoint((primary_location->>'longitude')::float, (primary_location->>'latitude')::float), 4326)::geography,
			ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
			?
		)
	`

	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Raw(query, tenantID, longitude, latitude, radiusKM*1000).
		Scan(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find customers by location", err)
	}

	return customers, nil
}

//------------------------------------------------------------
// Communication Preferences
//------------------------------------------------------------

func (r *customerRepository) UpdateNotificationPreferences(ctx context.Context, customerID uuid.UUID, email, sms, push bool) error {
	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		Updates(map[string]any{
			"email_notifications": email,
			"sms_notifications":   sms,
			"push_notifications":  push,
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update preferences", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "customer not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, customerID)
	return nil
}

func (r *customerRepository) GetCustomersOptedInForEmail(ctx context.Context, tenantID uuid.UUID) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ? AND email_notifications = ?", tenantID, true).
		Find(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find customers", err)
	}

	return customers, nil
}

func (r *customerRepository) GetCustomersOptedInForSMS(ctx context.Context, tenantID uuid.UUID) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ? AND sms_notifications = ?", tenantID, true).
		Find(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find customers", err)
	}

	return customers, nil
}

//------------------------------------------------------------
// Analytics & Reporting
//------------------------------------------------------------

func (r *customerRepository) GetCustomerAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (CustomerAnalytics, error) {
	analytics := CustomerAnalytics{
		ByLoyaltyTier: make(map[string]int64),
	}

	// Total customers
	r.db.WithContext(ctx).Model(&models.Customer{}).
		Where("tenant_id = ?", tenantID).
		Count(&analytics.TotalCustomers)

	// New customers in period
	query := r.db.WithContext(ctx).Model(&models.Customer{}).Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	query.Count(&analytics.NewCustomers)

	// Active customers (with bookings in period)
	r.db.WithContext(ctx).
		Raw(`
			SELECT COUNT(DISTINCT c.id)
			FROM customers c
			INNER JOIN bookings b ON b.customer_id = c.id
			WHERE c.tenant_id = ? AND b.start_time BETWEEN ? AND ?
		`, tenantID, startDate, endDate).
		Scan(&analytics.ActiveCustomers)

	// Average lifetime value
	var avgLTV float64
	r.db.WithContext(ctx).Model(&models.Customer{}).
		Select("COALESCE(AVG(total_spent), 0)").
		Where("tenant_id = ?", tenantID).
		Scan(&avgLTV)
	analytics.AverageLifetimeValue = avgLTV

	// Average bookings
	var avgBookings float64
	r.db.WithContext(ctx).Model(&models.Customer{}).
		Select("COALESCE(AVG(total_bookings), 0)").
		Where("tenant_id = ?", tenantID).
		Scan(&avgBookings)
	analytics.AverageBookings = avgBookings

	// By loyalty tier
	var customers []*models.Customer
	r.db.WithContext(ctx).Select("loyalty_points").Where("tenant_id = ?", tenantID).Find(&customers)

	tierCounts := map[string]int64{
		"Bronze":   0,
		"Silver":   0,
		"Gold":     0,
		"Platinum": 0,
	}

	for _, c := range customers {
		tier := c.GetLoyaltyTier()
		tierCounts[tier]++
	}
	analytics.ByLoyaltyTier = tierCounts

	// Retention rate
	retentionRate, _ := r.GetCustomerRetentionRate(ctx, tenantID, startDate, endDate)
	analytics.RetentionRate = retentionRate

	return analytics, nil
}

func (r *customerRepository) GetTopSpendingCustomers(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ?", tenantID).
		Order("total_spent DESC").
		Limit(limit).
		Find(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find top spending customers", err)
	}

	return customers, nil
}

func (r *customerRepository) GetMostActiveCustomers(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ?", tenantID).
		Order("total_bookings DESC").
		Limit(limit).
		Find(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find most active customers", err)
	}

	return customers, nil
}

func (r *customerRepository) GetCustomerRetentionRate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var customersAtStart, customersRetained int64

	// Customers who existed at start of period
	r.db.WithContext(ctx).Model(&models.Customer{}).
		Where("tenant_id = ? AND created_at < ?", tenantID, startDate).
		Count(&customersAtStart)

	if customersAtStart == 0 {
		return 0, nil
	}

	// Customers who made bookings in the period
	r.db.WithContext(ctx).
		Raw(`
			SELECT COUNT(DISTINCT c.id)
			FROM customers c
			INNER JOIN bookings b ON b.customer_id = c.id
			WHERE c.tenant_id = ? AND c.created_at < ? AND b.start_time BETWEEN ? AND ?
		`, tenantID, startDate, startDate, endDate).
		Scan(&customersRetained)

	return float64(customersRetained) / float64(customersAtStart) * 100, nil
}

func (r *customerRepository) GetAverageCustomerValue(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	var avgValue float64
	if err := r.db.WithContext(ctx).Model(&models.Customer{}).
		Select("COALESCE(AVG(total_spent), 0)").
		Where("tenant_id = ?", tenantID).
		Scan(&avgValue).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to calculate average value", err)
	}

	return avgValue, nil
}

func (r *customerRepository) GetCustomerLifetimeValue(ctx context.Context, customerID uuid.UUID) (float64, error) {
	customer, err := r.GetByID(ctx, customerID)
	if err != nil {
		return 0, err
	}

	return customer.TotalSpent, nil
}

func (r *customerRepository) GetChurnedCustomers(ctx context.Context, tenantID uuid.UUID, inactiveDays int) ([]*models.Customer, error) {
	cutoffDate := time.Now().AddDate(0, 0, -inactiveDays)

	var customers []*models.Customer
	err := r.db.WithContext(ctx).
		Preload("User").
		Joins("LEFT JOIN bookings ON bookings.customer_id = customers.id").
		Where("customers.tenant_id = ?", tenantID).
		Group("customers.id").
		Having("MAX(bookings.start_time) < ? OR MAX(bookings.start_time) IS NULL", cutoffDate).
		Find(&customers).Error

	if err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find churned customers", err)
	}

	return customers, nil
}

//------------------------------------------------------------
// Segmentation
//------------------------------------------------------------

func (r *customerRepository) GetCustomersBySpendingRange(ctx context.Context, tenantID uuid.UUID, minSpent, maxSpent float64, pagination PaginationParams) ([]*models.Customer, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Customer{}).
		Where("tenant_id = ? AND total_spent BETWEEN ? AND ?", tenantID, minSpent, maxSpent)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count customers", err)
	}

	var customers []*models.Customer
	if err := query.
		Preload("User").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("total_spent DESC").
		Find(&customers).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find customers", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return customers, paginationResult, nil
}

func (r *customerRepository) GetCustomersByBookingCount(ctx context.Context, tenantID uuid.UUID, minBookings, maxBookings int, pagination PaginationParams) ([]*models.Customer, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Customer{}).
		Where("tenant_id = ? AND total_bookings BETWEEN ? AND ?", tenantID, minBookings, maxBookings)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count customers", err)
	}

	var customers []*models.Customer
	if err := query.
		Preload("User").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("total_bookings DESC").
		Find(&customers).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find customers", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return customers, paginationResult, nil
}

func (r *customerRepository) GetHighValueCustomers(ctx context.Context, tenantID uuid.UUID, minSpent float64) ([]*models.Customer, error) {
	var customers []*models.Customer
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ? AND total_spent >= ?", tenantID, minSpent).
		Order("total_spent DESC").
		Find(&customers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find high value customers", err)
	}

	return customers, nil
}

func (r *customerRepository) GetAtRiskCustomers(ctx context.Context, tenantID uuid.UUID, inactiveDays int, minPreviousBookings int) ([]*models.Customer, error) {
	cutoffDate := time.Now().AddDate(0, 0, -inactiveDays)

	var customers []*models.Customer
	err := r.db.WithContext(ctx).
		Preload("User").
		Joins("LEFT JOIN bookings ON bookings.customer_id = customers.id").
		Where("customers.tenant_id = ? AND customers.total_bookings >= ?", tenantID, minPreviousBookings).
		Group("customers.id").
		Having("MAX(bookings.start_time) < ?", cutoffDate).
		Find(&customers).Error

	if err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find at-risk customers", err)
	}

	return customers, nil
}

//------------------------------------------------------------
// Search & Filter
//------------------------------------------------------------

func (r *customerRepository) Search(ctx context.Context, query string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Customer, PaginationResult, error) {
	pagination.Validate()

	like := fmt.Sprintf("%%%s%%", strings.TrimSpace(query))

	countQuery := r.db.WithContext(ctx).Model(&models.Customer{}).
		Joins("LEFT JOIN users ON users.id = customers.user_id").
		Where("customers.tenant_id = ?", tenantID)

	if query != "" {
		countQuery = countQuery.Where(
			"users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ? OR customers.notes ILIKE ?",
			like, like, like, like,
		)
	}

	var totalItems int64
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count customers", err)
	}

	dataQuery := r.db.WithContext(ctx).
		Preload("User").
		Joins("LEFT JOIN users ON users.id = customers.user_id").
		Where("customers.tenant_id = ?", tenantID)

	if query != "" {
		dataQuery = dataQuery.Where(
			"users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ? OR customers.notes ILIKE ?",
			like, like, like, like,
		)
	}

	var customers []*models.Customer
	if err := dataQuery.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("customers.created_at DESC").
		Find(&customers).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search customers", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return customers, paginationResult, nil
}

func (r *customerRepository) FindByFilters(ctx context.Context, filters CustomerFilters, pagination PaginationParams) ([]*models.Customer, PaginationResult, error) {
	// Allow platform admins to query customers across all tenants (tenant_id can be nil)
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.Customer{})
	query = r.applyCustomerFilters(query, filters)

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count customers", err)
	}

	var customers []*models.Customer
	if err := r.applyCustomerFilters(r.db.WithContext(ctx).Model(&models.Customer{}), filters).
		Preload("User").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&customers).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to apply filters", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return customers, paginationResult, nil
}

//------------------------------------------------------------
// Bulk Operations
//------------------------------------------------------------

func (r *customerRepository) BulkUpdateLoyaltyPoints(ctx context.Context, customerIDs []uuid.UUID, points int) error {
	if len(customerIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id IN ?", customerIDs).
		UpdateColumn("loyalty_points", gorm.Expr("loyalty_points + ?", points))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update loyalty points", result.Error)
	}

	for _, id := range customerIDs {
		r.InvalidateCache(ctx, id)
	}

	return nil
}

func (r *customerRepository) BulkUpdateNotificationPreferences(ctx context.Context, customerIDs []uuid.UUID, email, sms, push bool) error {
	if len(customerIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id IN ?", customerIDs).
		Updates(map[string]any{
			"email_notifications": email,
			"sms_notifications":   sms,
			"push_notifications":  push,
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update preferences", result.Error)
	}

	for _, id := range customerIDs {
		r.InvalidateCache(ctx, id)
	}

	return nil
}

//------------------------------------------------------------
// Helper Methods
//------------------------------------------------------------

func (r *customerRepository) applyCustomerFilters(query *gorm.DB, filters CustomerFilters) *gorm.DB {
	// Only filter by tenant_id if it's not nil (platform admins can query all tenants)
	if filters.TenantID != uuid.Nil {
		query = query.Where("tenant_id = ?", filters.TenantID)
	}

	if len(filters.UserIDs) > 0 {
		query = query.Where("user_id IN ?", filters.UserIDs)
	}

	if len(filters.LoyaltyTiers) > 0 {
		var conditions []string
		var args []any

		for _, tier := range filters.LoyaltyTiers {
			switch strings.ToLower(tier) {
			case "platinum":
				conditions = append(conditions, "loyalty_points >= ?")
				args = append(args, 1000)
			case "gold":
				conditions = append(conditions, "(loyalty_points >= ? AND loyalty_points < ?)")
				args = append(args, 500, 1000)
			case "silver":
				conditions = append(conditions, "(loyalty_points >= ? AND loyalty_points < ?)")
				args = append(args, 100, 500)
			case "bronze":
				conditions = append(conditions, "loyalty_points < ?")
				args = append(args, 100)
			}
		}

		if len(conditions) > 0 {
			query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
		}
	}

	if filters.MinLoyaltyPoints != nil {
		query = query.Where("loyalty_points >= ?", *filters.MinLoyaltyPoints)
	}

	if filters.MaxLoyaltyPoints != nil {
		query = query.Where("loyalty_points <= ?", *filters.MaxLoyaltyPoints)
	}

	if filters.MinTotalSpent != nil {
		query = query.Where("total_spent >= ?", *filters.MinTotalSpent)
	}

	if filters.MaxTotalSpent != nil {
		query = query.Where("total_spent <= ?", *filters.MaxTotalSpent)
	}

	if filters.MinBookings != nil {
		query = query.Where("total_bookings >= ?", *filters.MinBookings)
	}

	if filters.MaxBookings != nil {
		query = query.Where("total_bookings <= ?", *filters.MaxBookings)
	}

	if len(filters.PreferredArtisanIDs) > 0 {
		query = query.Where("preferred_artisans && ?", filters.PreferredArtisanIDs)
	}

	if filters.HasPaymentMethod != nil {
		if *filters.HasPaymentMethod {
			query = query.Where("default_payment_method_id IS NOT NULL AND default_payment_method_id != ''")
		} else {
			query = query.Where("default_payment_method_id IS NULL OR default_payment_method_id = ''")
		}
	}

	if filters.EmailNotifications != nil {
		query = query.Where("email_notifications = ?", *filters.EmailNotifications)
	}

	if filters.SMSNotifications != nil {
		query = query.Where("sms_notifications = ?", *filters.SMSNotifications)
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	return query
}

func (r *customerRepository) InvalidateCache(ctx context.Context, customerID uuid.UUID) error {
	if r.cache != nil {
		cacheKey := fmt.Sprintf("customer:%s", customerID.String())
		if err := r.cache.Delete(ctx, cacheKey); err != nil {
			if r.logger != nil {
				r.logger.Warnf("Failed to invalidate cache for customer %s: %v", customerID, err)
			}
			return err
		}
	}
	return nil
}
