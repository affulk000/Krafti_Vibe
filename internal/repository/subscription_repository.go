package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionRepository defines the interface for subscription repository operations
type SubscriptionRepository interface {
	BaseRepository[models.Subscription]

	// Core Operations
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.Subscription, error)
	GetByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*models.Subscription, error)
	GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.Subscription, error)
	GetActiveSubscriptions(ctx context.Context, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error)
	GetTrialingSubscriptions(ctx context.Context, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error)
	GetExpiringTrials(ctx context.Context, daysUntilExpiry int, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error)

	// Plan Management
	GetByPlan(ctx context.Context, plan models.SubscriptionPlan, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error)
	UpgradePlan(ctx context.Context, tenantID uuid.UUID, newPlan models.SubscriptionPlan) error
	DowngradePlan(ctx context.Context, tenantID uuid.UUID, newPlan models.SubscriptionPlan, immediate bool) error
	CancelSubscription(ctx context.Context, tenantID uuid.UUID, cancelAtPeriodEnd bool) error
	ReactivateSubscription(ctx context.Context, tenantID uuid.UUID) error

	// Usage Tracking
	IncrementUsage(ctx context.Context, tenantID uuid.UUID, usageType string, amount int) error
	DecrementUsage(ctx context.Context, tenantID uuid.UUID, usageType string, amount int) error
	CheckLimit(ctx context.Context, tenantID uuid.UUID, limitType string) (bool, error)
	GetUsageStats(ctx context.Context, tenantID uuid.UUID) (UsageStats, error)
	ResetMonthlyUsage(ctx context.Context, tenantID uuid.UUID) error

	// Billing Operations
	RecordPayment(ctx context.Context, tenantID uuid.UUID, amount float64, paymentDate time.Time) error
	RecordFailedPayment(ctx context.Context, tenantID uuid.UUID) error
	ResetFailedPayments(ctx context.Context, tenantID uuid.UUID) error
	IncrementFailedPayments(ctx context.Context, tenantID uuid.UUID) error
	UpdateBillingCycle(ctx context.Context, tenantID uuid.UUID, nextBillingDate time.Time) error
	GetSubscriptionsDueForBilling(ctx context.Context, beforeDate time.Time) ([]*models.Subscription, error)
	GetSubscriptionsWithFailedPayments(ctx context.Context, minFailures int) ([]*models.Subscription, error)

	// Feature Management
	UpdateFeatures(ctx context.Context, tenantID uuid.UUID, features models.SubscriptionFeatures) error
	HasFeature(ctx context.Context, tenantID uuid.UUID, feature string) (bool, error)
	GetEnabledFeatures(ctx context.Context, tenantID uuid.UUID) ([]string, error)

	// Analytics & Reporting
	GetSubscriptionStats(ctx context.Context) (SubscriptionStats, error)
	GetRevenueByPlan(ctx context.Context, startDate, endDate time.Time) (map[models.SubscriptionPlan]float64, error)
	GetChurnRate(ctx context.Context, startDate, endDate time.Time) (float64, error)
	GetMRR(ctx context.Context) (float64, error)
	GetARR(ctx context.Context) (float64, error)

	// Lifecycle Management
	StartTrial(ctx context.Context, tenantID uuid.UUID, trialDays int) error
	EndTrial(ctx context.Context, tenantID uuid.UUID) error
	SuspendSubscription(ctx context.Context, tenantID uuid.UUID, reason string) error
	GetExpiredSubscriptions(ctx context.Context) ([]*models.Subscription, error)
	AutoRenewSubscriptions(ctx context.Context) error

	// Search & Filter
	FindByFilters(ctx context.Context, filters SubscriptionFilters, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error)
	SearchSubscriptions(ctx context.Context, query string, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error)
}

type SubscriptionFilters struct {
	Plans         []models.SubscriptionPlan   `json:"plans"`
	Statuses      []models.SubscriptionStatus `json:"statuses"`
	BillingCycles []models.BillingInterval    `json:"billing_cycles"`
	MinAmount     *float64                    `json:"min_amount"`
	MaxAmount     *float64                    `json:"max_amount"`
	IsTrialing    *bool                       `json:"is_trialing"`
	HasFailures   *bool                       `json:"has_failures"`
	CreatedAfter  *time.Time                  `json:"created_after"`
	CreatedBefore *time.Time                  `json:"created_before"`
}

type SubscriptionStats struct {
	TotalSubscriptions    int64 `json:"total_subscriptions"`
	ActiveSubscriptions   int64 `json:"active_subscriptions"`
	TrialSubscriptions    int64 `json:"trial_subscriptions"`
	CanceledSubscriptions int64 `json:"canceled_subscriptions"`

	ByPlan   map[models.SubscriptionPlan]int64   `json:"by_plan"`
	ByStatus map[models.SubscriptionStatus]int64 `json:"by_status"`

	TotalMRR              float64 `json:"total_mrr"`
	TotalARR              float64 `json:"total_arr"`
	AverageRevenuePerUser float64 `json:"average_revenue_per_user"`
	ChurnRate             float64 `json:"churn_rate"`
}

type UsageStats struct {
	TenantID uuid.UUID               `json:"tenant_id"`
	Plan     models.SubscriptionPlan `json:"plan"`

	CustomersUsed  int `json:"customers_used"`
	CustomersLimit int `json:"customers_limit"`

	ProjectsUsed  int `json:"projects_used"`
	ProjectsLimit int `json:"projects_limit"`

	StorageUsedGB  int `json:"storage_used_gb"`
	StorageLimitGB int `json:"storage_limit_gb"`

	CustomersPercent float64 `json:"customers_percent"`
	ProjectsPercent  float64 `json:"projects_percent"`
	StoragePercent   float64 `json:"storage_percent"`
}

// subscriptionRepository implements SubscriptionRepository
type subscriptionRepository struct {
	BaseRepository[models.Subscription]
	db      *gorm.DB
	logger  log.AllLogger
	cache   Cache
	metrics MetricsCollector
}

// NewSubscriptionRepository creates a new SubscriptionRepository instance
func NewSubscriptionRepository(db *gorm.DB, config ...RepositoryConfig) SubscriptionRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 10 * time.Minute
	}

	baseRepo := NewBaseRepository[models.Subscription](db, cfg)

	return &subscriptionRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
		metrics:        cfg.Metrics,
	}
}

// GetByTenantID retrieves a subscription by tenant ID
func (r *subscriptionRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.Subscription, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_tenant_id", "subscriptions", time.Since(start), nil)
		}
	}()

	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant ID cannot be nil", errors.ErrInvalidInput)
	}

	// Try cache first
	cacheKey := r.getCacheKey("tenant", tenantID.String())
	if r.cache != nil {
		var subscription models.Subscription
		if err := r.cache.GetJSON(ctx, cacheKey, &subscription); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit("subscriptions")
			}
			return &subscription, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss("subscriptions")
		}
	}

	var subscription models.Subscription
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("tenant_id = ?", tenantID).
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "subscription not found", errors.ErrNotFound)
		}
		if r.logger != nil {
			r.logger.Error("failed to get subscription by tenant ID", "tenant_id", tenantID, "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get subscription", err)
	}

	// Cache the result
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, subscription, 10*time.Minute); err != nil && r.logger != nil {
			r.logger.Warn("failed to cache subscription", "tenant_id", tenantID, "error", err)
		}
	}

	return &subscription, nil
}

// GetByStripeSubscriptionID retrieves a subscription by Stripe subscription ID
func (r *subscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*models.Subscription, error) {
	if stripeSubID == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "stripe subscription ID cannot be empty", errors.ErrInvalidInput)
	}

	var subscription models.Subscription
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("stripe_subscription_id = ?", stripeSubID).
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "subscription not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get subscription", err)
	}

	return &subscription, nil
}

// GetByStripeCustomerID retrieves a subscription by Stripe customer ID
func (r *subscriptionRepository) GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.Subscription, error) {
	if stripeCustomerID == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "stripe customer ID cannot be empty", errors.ErrInvalidInput)
	}

	var subscription models.Subscription
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("stripe_customer_id = ?", stripeCustomerID).
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "subscription not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get subscription", err)
	}

	return &subscription, nil
}

// GetActiveSubscriptions retrieves all active subscriptions
func (r *subscriptionRepository) GetActiveSubscriptions(ctx context.Context, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("status IN ?", []models.SubscriptionStatus{models.SubStatusActive, models.SubStatusTrialing})

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count subscriptions", err)
	}

	var subscriptions []*models.Subscription
	if err := query.
		Preload("Tenant").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&subscriptions).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find subscriptions", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return subscriptions, paginationResult, nil
}

// GetTrialingSubscriptions retrieves all trialing subscriptions
func (r *subscriptionRepository) GetTrialingSubscriptions(ctx context.Context, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("status = ?", models.SubStatusTrialing).
		Where("trial_ends_at IS NOT NULL").
		Where("trial_ends_at > ?", time.Now())

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count subscriptions", err)
	}

	var subscriptions []*models.Subscription
	if err := query.
		Preload("Tenant").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("trial_ends_at ASC").
		Find(&subscriptions).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find subscriptions", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return subscriptions, paginationResult, nil
}

// GetExpiringTrials retrieves trials expiring within specified days
func (r *subscriptionRepository) GetExpiringTrials(ctx context.Context, daysUntilExpiry int, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error) {
	pagination.Validate()

	expiryDate := time.Now().AddDate(0, 0, daysUntilExpiry)

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("status = ?", models.SubStatusTrialing).
		Where("trial_ends_at IS NOT NULL").
		Where("trial_ends_at BETWEEN ? AND ?", time.Now(), expiryDate)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count subscriptions", err)
	}

	var subscriptions []*models.Subscription
	if err := query.
		Preload("Tenant").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("trial_ends_at ASC").
		Find(&subscriptions).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find subscriptions", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return subscriptions, paginationResult, nil
}

// GetByPlan retrieves subscriptions by plan
func (r *subscriptionRepository) GetByPlan(ctx context.Context, plan models.SubscriptionPlan, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("plan = ?", plan)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count subscriptions", err)
	}

	var subscriptions []*models.Subscription
	if err := query.
		Preload("Tenant").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&subscriptions).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find subscriptions", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return subscriptions, paginationResult, nil
}

// UpgradePlan upgrades a subscription to a new plan
func (r *subscriptionRepository) UpgradePlan(ctx context.Context, tenantID uuid.UUID, newPlan models.SubscriptionPlan) error {
	subscription, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	subscription.Plan = newPlan

	// Update limits and features
	limits := models.GetDefaultLimitsForPlan(newPlan)
	subscription.MaxCustomers = limits["max_customers"]
	subscription.MaxProjects = limits["max_projects"]
	subscription.MaxStorageGB = limits["max_storage_gb"]
	subscription.MaxTeamMembers = limits["max_team_members"]
	subscription.MaxServicesListed = limits["max_services_listed"]
	subscription.MaxBookingsPerMonth = limits["max_bookings_per_month"]

	subscription.Features = models.GetDefaultFeaturesForPlan(newPlan)

	if err := r.Update(ctx, subscription); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("subscription upgraded", "tenant_id", tenantID, "new_plan", newPlan)
	}
	return nil
}

// DowngradePlan downgrades a subscription to a new plan
func (r *subscriptionRepository) DowngradePlan(ctx context.Context, tenantID uuid.UUID, newPlan models.SubscriptionPlan, immediate bool) error {
	subscription, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	if immediate {
		subscription.Plan = newPlan
		limits := models.GetDefaultLimitsForPlan(newPlan)
		subscription.MaxCustomers = limits["max_customers"]
		subscription.MaxProjects = limits["max_projects"]
		subscription.MaxStorageGB = limits["max_storage_gb"]
		subscription.MaxTeamMembers = limits["max_team_members"]
		subscription.MaxServicesListed = limits["max_services_listed"]
		subscription.MaxBookingsPerMonth = limits["max_bookings_per_month"]
		subscription.Features = models.GetDefaultFeaturesForPlan(newPlan)
	} else {
		subscription.CancelAtPeriodEnd = true
		// Store new plan in metadata for application at period end
		if subscription.Metadata == nil {
			subscription.Metadata = models.JSONB{}
		}
		subscription.Metadata["pending_downgrade_plan"] = string(newPlan)
	}

	if err := r.Update(ctx, subscription); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("subscription downgraded", "tenant_id", tenantID, "new_plan", newPlan, "immediate", immediate)
	}
	return nil
}

// CancelSubscription cancels a subscription
func (r *subscriptionRepository) CancelSubscription(ctx context.Context, tenantID uuid.UUID, cancelAtPeriodEnd bool) error {
	subscription, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	now := time.Now()
	subscription.CancelAtPeriodEnd = cancelAtPeriodEnd

	if !cancelAtPeriodEnd {
		subscription.Status = models.SubStatusCanceled
		subscription.CanceledAt = &now
	}

	if err := r.Update(ctx, subscription); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("subscription canceled", "tenant_id", tenantID, "at_period_end", cancelAtPeriodEnd)
	}
	return nil
}

// ReactivateSubscription reactivates a canceled subscription
func (r *subscriptionRepository) ReactivateSubscription(ctx context.Context, tenantID uuid.UUID) error {
	subscription, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	subscription.Status = models.SubStatusActive
	subscription.CanceledAt = nil
	subscription.CancelAtPeriodEnd = false

	if err := r.Update(ctx, subscription); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("subscription reactivated", "tenant_id", tenantID)
	}
	return nil
}

// IncrementUsage increments usage counter
func (r *subscriptionRepository) IncrementUsage(ctx context.Context, tenantID uuid.UUID, usageType string, amount int) error {
	subscription, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	switch usageType {
	case "customers":
		subscription.CurrentCustomers += amount
	case "projects":
		subscription.CurrentProjects += amount
	case "storage_gb":
		subscription.CurrentStorageGB += amount
	default:
		return errors.NewRepositoryError("INVALID_INPUT", fmt.Sprintf("unknown usage type: %s", usageType), errors.ErrInvalidInput)
	}

	if err := r.Update(ctx, subscription); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTenantCache(ctx, tenantID)

	return nil
}

// DecrementUsage decrements usage counter
func (r *subscriptionRepository) DecrementUsage(ctx context.Context, tenantID uuid.UUID, usageType string, amount int) error {
	subscription, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	switch usageType {
	case "customers":
		subscription.CurrentCustomers -= amount
		if subscription.CurrentCustomers < 0 {
			subscription.CurrentCustomers = 0
		}
	case "projects":
		subscription.CurrentProjects -= amount
		if subscription.CurrentProjects < 0 {
			subscription.CurrentProjects = 0
		}
	case "storage_gb":
		subscription.CurrentStorageGB -= amount
		if subscription.CurrentStorageGB < 0 {
			subscription.CurrentStorageGB = 0
		}
	default:
		return errors.NewRepositoryError("INVALID_INPUT", fmt.Sprintf("unknown usage type: %s", usageType), errors.ErrInvalidInput)
	}

	if err := r.Update(ctx, subscription); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTenantCache(ctx, tenantID)

	return nil
}

// CheckLimit checks if a usage limit has been reached
func (r *subscriptionRepository) CheckLimit(ctx context.Context, tenantID uuid.UUID, limitType string) (bool, error) {
	subscription, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return false, err
	}

	switch limitType {
	case "customers":
		if subscription.MaxCustomers == -1 {
			return true, nil // Unlimited
		}
		return subscription.CurrentCustomers < subscription.MaxCustomers, nil
	case "projects":
		if subscription.MaxProjects == -1 {
			return true, nil
		}
		return subscription.CurrentProjects < subscription.MaxProjects, nil
	case "storage_gb":
		if subscription.MaxStorageGB == -1 {
			return true, nil
		}
		return subscription.CurrentStorageGB < subscription.MaxStorageGB, nil
	default:
		return false, errors.NewRepositoryError("INVALID_INPUT", fmt.Sprintf("unknown limit type: %s", limitType), errors.ErrInvalidInput)
	}
}

// GetUsageStats retrieves usage statistics for a subscription
func (r *subscriptionRepository) GetUsageStats(ctx context.Context, tenantID uuid.UUID) (UsageStats, error) {
	subscription, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return UsageStats{}, err
	}

	return UsageStats{
		TenantID:         tenantID,
		Plan:             subscription.Plan,
		CustomersUsed:    subscription.CurrentCustomers,
		CustomersLimit:   subscription.MaxCustomers,
		ProjectsUsed:     subscription.CurrentProjects,
		ProjectsLimit:    subscription.MaxProjects,
		StorageUsedGB:    subscription.CurrentStorageGB,
		StorageLimitGB:   subscription.MaxStorageGB,
		CustomersPercent: calculatePercentage(subscription.CurrentCustomers, subscription.MaxCustomers),
		ProjectsPercent:  calculatePercentage(subscription.CurrentProjects, subscription.MaxProjects),
		StoragePercent:   calculatePercentage(subscription.CurrentStorageGB, subscription.MaxStorageGB),
	}, nil
}

// ResetMonthlyUsage resets all monthly usage counters for a tenant
func (r *subscriptionRepository) ResetMonthlyUsage(ctx context.Context, tenantID uuid.UUID) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("reset_monthly_usage", "subscriptions", time.Since(start), nil)
		}
	}()

	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant ID cannot be nil", errors.ErrInvalidInput)
	}

	// Fetch subscription
	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	// Reset only tracked usage fields
	sub.CurrentCustomers = 0
	sub.CurrentProjects = 0
	sub.CurrentStorageGB = 0

	// Update in DB
	if err := r.Update(ctx, sub); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to reset monthly usage", "tenant_id", tenantID, "error", err)
		}
		return err
	}

	// Invalidate cache
	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("monthly usage reset", "tenant_id", tenantID, "plan", sub.Plan)
	}
	return nil
}

// RecordPayment records a successful payment
func (r *subscriptionRepository) RecordPayment(ctx context.Context, tenantID uuid.UUID, amount float64, paymentDate time.Time) error {
	start := time.Now()
	defer func() {
		r.recordMetric("record_payment", time.Since(start), nil)
	}()

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	sub.LastPaymentDate = &paymentDate
	sub.LastPaymentAmount = amount
	sub.FailedPayments = 0

	if sub.NextBillingDate != nil && paymentDate.After(*sub.NextBillingDate) {
		// Advance billing cycle
		var nextDate time.Time
		switch sub.BillingInterval {
		case models.BillingMonthly:
			nextDate = paymentDate.AddDate(0, 1, 0)
		case models.BillingYearly:
			nextDate = paymentDate.AddDate(1, 0, 0)
		case models.BillingLifetime:
			nextDate = time.Time{}
		default:
			nextDate = paymentDate.AddDate(0, 1, 0)
		}
		if !nextDate.IsZero() {
			sub.NextBillingDate = &nextDate
			sub.CurrentPeriodStart = *sub.NextBillingDate
			sub.CurrentPeriodEnd = nextDate
		}
	}

	if err := r.Update(ctx, sub); err != nil {
		return err
	}

	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("payment recorded", "tenant_id", tenantID, "amount", amount)
	}
	return nil
}

// RecordFailedPayment records a failed payment attempt
func (r *subscriptionRepository) RecordFailedPayment(ctx context.Context, tenantID uuid.UUID) error {
	start := time.Now()
	defer func() {
		r.recordMetric("record_failed_payment", time.Since(start), nil)
	}()

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	sub.FailedPayments++
	if sub.FailedPayments >= 3 {
		sub.Status = models.SubStatusPastDue
	}

	if err := r.Update(ctx, sub); err != nil {
		return err
	}

	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Warn("payment failed", "tenant_id", tenantID, "failures", sub.FailedPayments)
	}
	return nil
}

// UpdateBillingCycle updates the next billing date
func (r *subscriptionRepository) UpdateBillingCycle(ctx context.Context, tenantID uuid.UUID, nextBillingDate time.Time) error {
	start := time.Now()
	defer func() {
		r.recordMetric("update_billing_cycle", time.Since(start), nil)
	}()

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	sub.NextBillingDate = &nextBillingDate
	if err := r.Update(ctx, sub); err != nil {
		return err
	}

	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("billing cycle updated", "tenant_id", tenantID, "next_date", nextBillingDate)
	}
	return nil
}

// GetSubscriptionsDueForBilling returns subscriptions due before a date
func (r *subscriptionRepository) GetSubscriptionsDueForBilling(ctx context.Context, beforeDate time.Time) ([]*models.Subscription, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("get_due_for_billing", time.Since(start), nil)
	}()

	var subs []*models.Subscription
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("status IN ? AND next_billing_date IS NOT NULL AND next_billing_date < ?",
			[]models.SubscriptionStatus{models.SubStatusActive, models.SubStatusTrialing}, beforeDate).
		Find(&subs).Error

	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get due subscriptions", "before", beforeDate, "error", err)
		}
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to query due subscriptions", err)
	}

	return subs, nil
}

// GetSubscriptionsWithFailedPayments returns subscriptions with failed payments
func (r *subscriptionRepository) GetSubscriptionsWithFailedPayments(ctx context.Context, minFailures int) ([]*models.Subscription, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("get_failed_payments", time.Since(start), nil)
	}()

	var subs []*models.Subscription
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("failed_payments >= ?", minFailures).
		Find(&subs).Error

	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get failed payment subs", "min_failures", minFailures, "error", err)
		}
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to query failed payments", err)
	}

	return subs, nil
}

// UpdateFeatures updates feature flags for a subscription
func (r *subscriptionRepository) UpdateFeatures(ctx context.Context, tenantID uuid.UUID, features models.SubscriptionFeatures) error {
	start := time.Now()
	defer func() {
		r.recordMetric("update_features", time.Since(start), nil)
	}()

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	sub.Features = features
	if err := r.Update(ctx, sub); err != nil {
		return err
	}

	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("features updated", "tenant_id", tenantID)
	}
	return nil
}

// HasFeature checks if a specific feature is enabled
func (r *subscriptionRepository) HasFeature(ctx context.Context, tenantID uuid.UUID, feature string) (bool, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("has_feature", time.Since(start), nil)
	}()

	cacheKey := r.getCacheKey("feature", tenantID.String(), feature)
	if r.cache != nil {
		var has bool
		if err := r.cache.GetJSON(ctx, cacheKey, &has); err == nil {
			r.recordCacheHit()
			return has, nil
		}
		r.recordCacheMiss()
	}

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return false, err
	}

	has := r.checkFeature(sub.Features, feature)
	if r.cache != nil {
		_ = r.cache.SetJSON(ctx, cacheKey, has, 5*time.Minute)
	}

	return has, nil
}

// GetEnabledFeatures returns all enabled feature names
func (r *subscriptionRepository) GetEnabledFeatures(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("get_enabled_features", time.Since(start), nil)
	}()

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	features := r.getEnabledFeatureList(sub.Features)
	return features, nil
}

// GetSubscriptionStats returns overall subscription statistics
func (r *subscriptionRepository) GetSubscriptionStats(ctx context.Context) (SubscriptionStats, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("get_stats", time.Since(start), nil)
	}()

	var stats SubscriptionStats

	// Total counts
	r.db.WithContext(ctx).Model(&models.Subscription{}).Count(&stats.TotalSubscriptions)
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("status = ?", models.SubStatusActive).Count(&stats.ActiveSubscriptions)
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("status = ?", models.SubStatusTrialing).Count(&stats.TrialSubscriptions)
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("status = ?", models.SubStatusCanceled).Count(&stats.CanceledSubscriptions)

	// By Plan
	var planCounts []struct {
		Plan  models.SubscriptionPlan
		Count int64
	}
	r.db.WithContext(ctx).Model(&models.Subscription{}).Select("plan, count(*) as count").Group("plan").Scan(&planCounts)
	stats.ByPlan = make(map[models.SubscriptionPlan]int64)
	for _, pc := range planCounts {
		stats.ByPlan[pc.Plan] = pc.Count
	}

	// By Status
	var statusCounts []struct {
		Status models.SubscriptionStatus
		Count  int64
	}
	r.db.WithContext(ctx).Model(&models.Subscription{}).Select("status, count(*) as count").Group("status").Scan(&statusCounts)
	stats.ByStatus = make(map[models.SubscriptionStatus]int64)
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// MRR & ARR
	stats.TotalMRR, _ = r.GetMRR(ctx)
	stats.TotalARR = stats.TotalMRR * 12
	stats.AverageRevenuePerUser = 0
	if stats.ActiveSubscriptions > 0 {
		stats.AverageRevenuePerUser = stats.TotalMRR / float64(stats.ActiveSubscriptions)
	}

	// Churn (simplified: canceled last 30 days / active 30 days ago)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var canceledLast30, active30Ago int64
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("canceled_at >= ?", thirtyDaysAgo).Count(&canceledLast30)
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("created_at <= ? AND status = ?", thirtyDaysAgo, models.SubStatusActive).Count(&active30Ago)
	if active30Ago > 0 {
		stats.ChurnRate = float64(canceledLast30) / float64(active30Ago) * 100
	}

	return stats, nil
}

// GetRevenueByPlan returns revenue per plan for a date range
func (r *subscriptionRepository) GetRevenueByPlan(ctx context.Context, startDate, endDate time.Time) (map[models.SubscriptionPlan]float64, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("revenue_by_plan", time.Since(start), nil)
	}()

	revenue := make(map[models.SubscriptionPlan]float64)

	var results []struct {
		Plan   models.SubscriptionPlan
		Amount float64
	}
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT s.plan, COALESCE(SUM(i.amount), 0) as amount
			FROM subscriptions s
			LEFT JOIN invoices i ON i.subscription_id = s.id
			WHERE i.paid_at BETWEEN ? AND ?
			GROUP BY s.plan
		`, startDate, endDate).
		Scan(&results).Error

	if err != nil {
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to calculate revenue", err)
	}

	for _, res := range results {
		revenue[res.Plan] = res.Amount
	}

	return revenue, nil
}

// GetChurnRate calculates churn rate between two dates
func (r *subscriptionRepository) GetChurnRate(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("get_churn_rate", time.Since(start), nil)
	}()

	var canceled, activeStart int64
	r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("canceled_at BETWEEN ? AND ?", startDate, endDate).
		Count(&canceled)

	r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("created_at <= ? AND (canceled_at IS NULL OR canceled_at > ?)", startDate, startDate).
		Count(&activeStart)

	if activeStart == 0 {
		return 0, nil
	}
	return (float64(canceled) / float64(activeStart)) * 100, nil
}

// GetMRR returns current Monthly Recurring Revenue
func (r *subscriptionRepository) GetMRR(ctx context.Context) (float64, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("get_mrr", time.Since(start), nil)
	}()

	var mrr float64
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM(amount), 0)
			FROM subscriptions
			WHERE status IN (?, ?)
			  AND billing_interval = ?
		`, models.SubStatusActive, models.SubStatusTrialing, models.BillingMonthly).
		Scan(&mrr).Error

	if err != nil {
		return 0, errors.NewRepositoryError("QUERY_FAILED", "failed to calculate MRR", err)
	}

	// Add yearly plans divided by 12
	var yearly float64
	r.db.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM(amount), 0)
			FROM subscriptions
			WHERE status IN (?, ?) AND billing_interval = ?
		`, models.SubStatusActive, models.SubStatusTrialing, models.BillingYearly).
		Scan(&yearly)

	mrr += yearly / 12
	return mrr, nil
}

// GetARR returns Annual Recurring Revenue
func (r *subscriptionRepository) GetARR(ctx context.Context) (float64, error) {
	mrr, err := r.GetMRR(ctx)
	if err != nil {
		return 0, err
	}
	return mrr * 12, nil
}

// StartTrial starts a trial period
func (r *subscriptionRepository) StartTrial(ctx context.Context, tenantID uuid.UUID, trialDays int) error {
	start := time.Now()
	defer func() {
		r.recordMetric("start_trial", time.Since(start), nil)
	}()

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	if sub.IsTrialing() {
		return errors.NewRepositoryError("ALREADY_TRIALING", "subscription already in trial", errors.ErrInvalidInput)
	}

	now := time.Now()
	trialEnd := now.AddDate(0, 0, trialDays)

	sub.Status = models.SubStatusTrialing
	sub.TrialEndsAt = &trialEnd
	sub.CurrentPeriodStart = now
	sub.CurrentPeriodEnd = trialEnd

	if err := r.Update(ctx, sub); err != nil {
		return err
	}

	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("trial started", "tenant_id", tenantID, "days", trialDays, "ends", trialEnd)
	}
	return nil
}

// EndTrial ends the trial and converts to paid
func (r *subscriptionRepository) EndTrial(ctx context.Context, tenantID uuid.UUID) error {
	start := time.Now()
	defer func() {
		r.recordMetric("end_trial", time.Since(start), nil)
	}()

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	if !sub.IsTrialing() {
		return errors.NewRepositoryError("NOT_TRIALING", "subscription not in trial", errors.ErrInvalidInput)
	}

	sub.Status = models.SubStatusActive
	sub.TrialEndsAt = nil

	// Set next billing
	nextBilling := sub.CurrentPeriodEnd
	sub.NextBillingDate = &nextBilling

	if err := r.Update(ctx, sub); err != nil {
		return err
	}

	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Info("trial ended", "tenant_id", tenantID)
	}
	return nil
}

// SuspendSubscription suspends a subscription
func (r *subscriptionRepository) SuspendSubscription(ctx context.Context, tenantID uuid.UUID, reason string) error {
	start := time.Now()
	defer func() {
		r.recordMetric("suspend", time.Since(start), nil)
	}()

	sub, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	sub.Status = models.SubStatusSuspended
	if sub.Metadata == nil {
		sub.Metadata = models.JSONB{}
	}
	sub.Metadata["suspend_reason"] = reason
	sub.Metadata["suspended_at"] = time.Now()

	if err := r.Update(ctx, sub); err != nil {
		return err
	}

	r.invalidateTenantCache(ctx, tenantID)

	if r.logger != nil {
		r.logger.Warn("subscription suspended", "tenant_id", tenantID, "reason", reason)
	}
	return nil
}

// GetExpiredSubscriptions returns expired trials/past-due
func (r *subscriptionRepository) GetExpiredSubscriptions(ctx context.Context) ([]*models.Subscription, error) {
	start := time.Now()
	defer func() {
		r.recordMetric("get_expired", time.Since(start), nil)
	}()

	now := time.Now()
	var subs []*models.Subscription
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("(status = ? AND trial_ends_at < ?) OR (status = ? AND failed_payments >= 3)",
			models.SubStatusTrialing, now, models.SubStatusPastDue).
		Find(&subs).Error

	if err != nil {
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get expired", err)
	}

	return subs, nil
}

// AutoRenewSubscriptions renews due subscriptions (run via cron)
func (r *subscriptionRepository) AutoRenewSubscriptions(ctx context.Context) error {
	start := time.Now()
	defer func() {
		r.recordMetric("auto_renew", time.Since(start), nil)
	}()

	dueSubs, err := r.GetSubscriptionsDueForBilling(ctx, time.Now())
	if err != nil {
		return err
	}

	for _, sub := range dueSubs {
		// In real app: trigger Stripe charge
		// For now: just advance cycle
		now := time.Now()
		var nextDate time.Time
		switch sub.BillingInterval {
		case models.BillingMonthly:
			nextDate = now.AddDate(0, 1, 0)
		case models.BillingYearly:
			nextDate = now.AddDate(1, 0, 0)
		default:
			continue
		}

		sub.CurrentPeriodStart = *sub.NextBillingDate
		sub.CurrentPeriodEnd = nextDate
		sub.NextBillingDate = &nextDate

		if err := r.Update(ctx, sub); err != nil {
			if r.logger != nil {
				r.logger.Error("failed to renew", "tenant_id", sub.TenantID, "error", err)
			}
			continue
		}
		r.invalidateTenantCache(ctx, sub.TenantID)
	}

	if r.logger != nil {
		r.logger.Info("auto-renew completed", "processed", len(dueSubs))
	}
	return nil
}

// FindByFilters finds subscriptions with complex filters
func (r *subscriptionRepository) FindByFilters(ctx context.Context, filters SubscriptionFilters, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.Subscription{})
	countQuery := r.db.WithContext(ctx).Model(&models.Subscription{})

	// Apply filters
	if len(filters.Plans) > 0 {
		query = query.Where("plan IN ?", filters.Plans)
		countQuery = countQuery.Where("plan IN ?", filters.Plans)
	}
	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
		countQuery = countQuery.Where("status IN ?", filters.Statuses)
	}
	if len(filters.BillingCycles) > 0 {
		query = query.Where("billing_interval IN ?", filters.BillingCycles)
		countQuery = countQuery.Where("billing_interval IN ?", filters.BillingCycles)
	}
	if filters.MinAmount != nil {
		query = query.Where("amount >= ?", *filters.MinAmount)
		countQuery = countQuery.Where("amount >= ?", *filters.MinAmount)
	}
	if filters.MaxAmount != nil {
		query = query.Where("amount <= ?", *filters.MaxAmount)
		countQuery = countQuery.Where("amount <= ?", *filters.MaxAmount)
	}
	if filters.IsTrialing != nil {
		if *filters.IsTrialing {
			query = query.Where("status = ? AND trial_ends_at > ?", models.SubStatusTrialing, time.Now())
			countQuery = countQuery.Where("status = ? AND trial_ends_at > ?", models.SubStatusTrialing, time.Now())
		} else {
			query = query.Where("status != ? OR trial_ends_at <= ?", models.SubStatusTrialing, time.Now())
			countQuery = countQuery.Where("status != ? OR trial_ends_at <= ?", models.SubStatusTrialing, time.Now())
		}
	}
	if filters.HasFailures != nil && *filters.HasFailures {
		query = query.Where("failed_payments > 0")
		countQuery = countQuery.Where("failed_payments > 0")
	}
	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
		countQuery = countQuery.Where("created_at >= ?", *filters.CreatedAfter)
	}
	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
		countQuery = countQuery.Where("created_at <= ?", *filters.CreatedBefore)
	}

	// Count
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count", err)
	}

	// Query
	var subs []*models.Subscription
	if err := query.
		Preload("Tenant").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&subs).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find", err)
	}

	result := CalculatePagination(pagination, total)
	return subs, result, nil
}

// SearchSubscriptions performs full-text search
func (r *subscriptionRepository) SearchSubscriptions(ctx context.Context, query string, pagination PaginationParams) ([]*models.Subscription, PaginationResult, error) {
	pagination.Validate()

	if strings.TrimSpace(query) == "" {
		return []*models.Subscription{}, PaginationResult{}, nil
	}

	search := "%" + query + "%"
	var total int64
	countQuery := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Joins("LEFT JOIN tenants ON tenants.id = subscriptions.tenant_id").
		Where("subscriptions.stripe_customer_id ILIKE ? OR tenants.email ILIKE ? OR tenants.name ILIKE ?", search, search, search)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, PaginationResult{}, err
	}

	var subs []*models.Subscription
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Joins("LEFT JOIN tenants ON tenants.id = subscriptions.tenant_id").
		Where("subscriptions.stripe_customer_id ILIKE ? OR tenants.email ILIKE ? OR tenants.name ILIKE ?", search, search, search).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&subs).Error; err != nil {
		return nil, PaginationResult{}, err
	}

	result := CalculatePagination(pagination, total)
	return subs, result, nil
}

// Helper: invalidate tenant cache
func (r *subscriptionRepository) invalidateTenantCache(ctx context.Context, tenantID uuid.UUID) {
	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.getCacheKey("tenant", tenantID.String()))
	}
}

// Helper: record metric
func (r *subscriptionRepository) recordMetric(op string, duration time.Duration, err error) {
	if r.metrics != nil {
		r.metrics.RecordOperation(op, "subscriptions", duration, err)
	}
}

// Helper: record cache hit/miss
func (r *subscriptionRepository) recordCacheHit() {
	if r.metrics != nil {
		r.metrics.RecordCacheHit("subscriptions")
	}
}

func (r *subscriptionRepository) recordCacheMiss() {
	if r.metrics != nil {
		r.metrics.RecordCacheMiss("subscriptions")
	}
}

// Helper: check feature by name (case-insensitive)
func (r *subscriptionRepository) checkFeature(features models.SubscriptionFeatures, feature string) bool {
	v := reflect.ValueOf(features)
	t := reflect.TypeOf(features)
	feature = strings.ReplaceAll(strings.ToLower(feature), " ", "_")

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if strings.ToLower(jsonTag) == feature {
			return v.Field(i).Bool()
		}
	}
	return false
}

// Helper: get enabled feature names
func (r *subscriptionRepository) getEnabledFeatureList(features models.SubscriptionFeatures) []string {
	var list []string
	v := reflect.ValueOf(features)
	t := reflect.TypeOf(features)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if v.Field(i).Bool() {
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			list = append(list, jsonTag)
		}
	}
	return list
}

// Helper methods
func (r *subscriptionRepository) getCacheKey(prefix string, parts ...string) string {
	allParts := append([]string{"repo", "subscriptions", prefix}, parts...)
	return strings.Join(allParts, ":")
}

func calculatePercentage(current, max int) float64 {
	if max == -1 || max == 0 {
		return 0.0
	}
	return (float64(current) / float64(max)) * 100.0
}

// ============================================================================
// Missing Method Implementations
// ============================================================================

// ResetFailedPayments resets the failed payments counter
func (r *subscriptionRepository) ResetFailedPayments(ctx context.Context, tenantID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("tenant_id = ?", tenantID).
		UpdateColumn("failed_payments", 0).Error; err != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to reset failed payments", err)
	}

	// Invalidate cache
	if r.cache != nil {
		cacheKey := r.getCacheKey("tenant", tenantID.String())
		r.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// IncrementFailedPayments increments the failed payments counter
func (r *subscriptionRepository) IncrementFailedPayments(ctx context.Context, tenantID uuid.UUID) error {
	return r.RecordFailedPayment(ctx, tenantID) // Same implementation
}
