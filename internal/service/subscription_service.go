package service

import (
	"context"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// SubscriptionService defines the interface for subscription service operations
type SubscriptionService interface {
	// Core Operations
	CreateSubscription(ctx context.Context, req *dto.CreateSubscriptionRequest) (*dto.SubscriptionResponse, error)
	GetSubscription(ctx context.Context, tenantID uuid.UUID) (*dto.SubscriptionResponse, error)
	UpdateSubscription(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateSubscriptionRequest) (*dto.SubscriptionResponse, error)
	DeleteSubscription(ctx context.Context, tenantID uuid.UUID) error

	// Plan Management
	ChangePlan(ctx context.Context, tenantID uuid.UUID, req *dto.ChangePlanRequest) (*dto.BillingPreviewResponse, error)
	PreviewPlanChange(ctx context.Context, tenantID uuid.UUID, req *dto.ChangePlanRequest) (*dto.BillingPreviewResponse, error)
	GetPlanComparison(ctx context.Context) (*dto.PlanComparisonResponse, error)

	// Subscription Lifecycle
	StartTrial(ctx context.Context, tenantID uuid.UUID, trialDays int) (*dto.SubscriptionResponse, error)
	EndTrial(ctx context.Context, tenantID uuid.UUID) (*dto.SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, tenantID uuid.UUID, req *dto.CancelSubscriptionRequest) (*dto.SubscriptionResponse, error)
	ReactivateSubscription(ctx context.Context, tenantID uuid.UUID, req *dto.ReactivateSubscriptionRequest) (*dto.SubscriptionResponse, error)
	SuspendSubscription(ctx context.Context, tenantID uuid.UUID, reason string) (*dto.SubscriptionResponse, error)

	// Billing Operations
	UpdateBillingInterval(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateBillingIntervalRequest) (*dto.SubscriptionResponse, error)
	ProcessPayment(ctx context.Context, tenantID uuid.UUID, req *dto.ProcessPaymentRequest) (*dto.PaymentResponse, error)
	ApplyPromoCode(ctx context.Context, tenantID uuid.UUID, req *dto.ApplyPromoCodeRequest) (*dto.SubscriptionResponse, error)
	GetInvoices(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.PaymentHistoryResponse, error)
	GetPaymentHistory(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.PaymentHistoryResponse, error)

	// Usage & Limits
	GetUsage(ctx context.Context, tenantID uuid.UUID) (*dto.UsageResponse, error)
	UpdateUsage(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateUsageRequest) error
	CheckLimits(ctx context.Context, tenantID uuid.UUID, limitType string) (bool, error)
	EnforceUsageLimits(ctx context.Context, tenantID uuid.UUID) error
	ResetMonthlyUsage(ctx context.Context, tenantID uuid.UUID) error

	// Feature Management
	GetFeatureAccess(ctx context.Context, tenantID uuid.UUID) (*dto.FeatureAccessResponse, error)
	HasFeature(ctx context.Context, tenantID uuid.UUID, feature string) (bool, error)
	GetEnabledFeatures(ctx context.Context, tenantID uuid.UUID) ([]string, error)

	// Analytics & Reporting
	GetSubscriptionStats(ctx context.Context) (*dto.SubscriptionStatsResponse, error)
	GetSubscriptionList(ctx context.Context, filter dto.SubscriptionFilter) (*dto.SubscriptionListResponse, error)
	GetAnalytics(ctx context.Context, filter dto.AnalyticsFilter) (map[string]interface{}, error)
	GetChurnAnalysis(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error)
	GetRevenueAnalysis(ctx context.Context, filter dto.AnalyticsFilter) (map[string]interface{}, error)

	// Background Operations
	ProcessExpiringTrials(ctx context.Context) error
	ProcessFailedPayments(ctx context.Context) error
	ProcessSubscriptionRenewals(ctx context.Context) error
	CleanupExpiredSubscriptions(ctx context.Context) error

	// Health & Monitoring
	HealthCheck(ctx context.Context) error
	GetServiceMetrics(ctx context.Context) map[string]interface{}
}

// subscriptionService implements SubscriptionService
type subscriptionService struct {
	repos          *repository.Repositories
	paymentService PaymentService
	logger         log.AllLogger
}

// NewSubscriptionService creates a new SubscriptionService instance
func NewSubscriptionService(repos *repository.Repositories, paymentService PaymentService, logger log.AllLogger) SubscriptionService {
	return &subscriptionService{
		repos:          repos,
		paymentService: paymentService,
		logger:         logger,
	}
}

// ============================================================================
// Core Operations
// ============================================================================

// CreateSubscription creates a new subscription for a tenant
func (s *subscriptionService) CreateSubscription(ctx context.Context, req *dto.CreateSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	// Check if subscription already exists
	existingSub, err := s.repos.Subscription.GetByTenantID(ctx, req.TenantID)
	if err != nil && !errors.IsNotFoundError(err) {
		return nil, errors.NewServiceError("SUBSCRIPTION_CHECK_FAILED", "failed to check existing subscription", err)
	}
	if existingSub != nil {
		return nil, errors.NewConflictError("subscription already exists for tenant")
	}

	// Get plan pricing and limits
	pricing := s.getPlanPricing(req.Plan, req.BillingInterval)
	limits := models.GetDefaultLimitsForPlan(req.Plan)
	features := models.GetDefaultFeaturesForPlan(req.Plan)

	// Create subscription model
	now := time.Now()
	subscription := &models.Subscription{
		TenantID:        req.TenantID,
		Plan:            req.Plan,
		Status:          models.SubStatusTrialing,
		BillingInterval: req.BillingInterval,
		Amount:          pricing,
		Currency:        "USD", // Default currency
		DiscountPercent: 0,
		PaymentMethodID: req.PaymentMethodID,
		Features:        features,
		Metadata:        req.Metadata,
	}

	// Set limits
	subscription.MaxCustomers = limits["max_customers"]
	subscription.MaxProjects = limits["max_projects"]
	subscription.MaxStorageGB = limits["max_storage_gb"]
	subscription.MaxTeamMembers = limits["max_team_members"]
	subscription.MaxServicesListed = limits["max_services_listed"]
	subscription.MaxBookingsPerMonth = limits["max_bookings_per_month"]

	// Set trial or active period
	if req.TrialDays > 0 && req.Plan != models.PlanFree {
		trialEnd := now.AddDate(0, 0, req.TrialDays)
		subscription.TrialEndsAt = &trialEnd
		subscription.CurrentPeriodStart = now
		subscription.CurrentPeriodEnd = trialEnd
	} else if req.Plan == models.PlanFree {
		// Free plan is always active, no trial
		subscription.Status = models.SubStatusActive
		subscription.CurrentPeriodStart = now
		subscription.CurrentPeriodEnd = now.AddDate(1, 0, 0) // 1 year for free
	} else {
		// No trial, start as active
		subscription.Status = models.SubStatusActive
		subscription.CurrentPeriodStart = now
		switch req.BillingInterval {
		case models.BillingMonthly:
			subscription.CurrentPeriodEnd = now.AddDate(0, 1, 0)
			subscription.NextBillingDate = &subscription.CurrentPeriodEnd
		case models.BillingYearly:
			subscription.CurrentPeriodEnd = now.AddDate(1, 0, 0)
			subscription.NextBillingDate = &subscription.CurrentPeriodEnd
		case models.BillingLifetime:
			subscription.CurrentPeriodEnd = now.AddDate(10, 0, 0) // 10 years for lifetime
		}
	}

	// Create in repository
	if err := s.repos.Subscription.Create(ctx, subscription); err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_CREATE_FAILED", "failed to create subscription", err)
	}

	// Apply promo code if provided
	if req.PromoCode != "" {
		if err := s.applyPromoCodeInternal(ctx, subscription.TenantID, req.PromoCode); err != nil {
			s.logger.Warn("failed to apply promo code during subscription creation", "tenant_id", req.TenantID, "promo_code", req.PromoCode, "error", err)
		}
	}

	// Log subscription creation
	s.logger.Info("subscription created", "tenant_id", req.TenantID, "plan", req.Plan, "billing_interval", req.BillingInterval)

	return dto.ToSubscriptionResponse(subscription), nil
}

// GetSubscription retrieves a subscription by tenant ID
func (s *subscriptionService) GetSubscription(ctx context.Context, tenantID uuid.UUID) (*dto.SubscriptionResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("subscription not found")
		}
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	return dto.ToSubscriptionResponse(subscription), nil
}

// UpdateSubscription updates a subscription
func (s *subscriptionService) UpdateSubscription(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("subscription not found")
		}
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	// Update fields
	if req.PaymentMethodID != nil {
		subscription.PaymentMethodID = *req.PaymentMethodID
	}
	if req.Metadata != nil {
		subscription.Metadata = req.Metadata
	}

	if err := s.repos.Subscription.Update(ctx, subscription); err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_UPDATE_FAILED", "failed to update subscription", err)
	}

	s.logger.Info("subscription updated", "tenant_id", tenantID)
	return dto.ToSubscriptionResponse(subscription), nil
}

// DeleteSubscription deletes a subscription
func (s *subscriptionService) DeleteSubscription(ctx context.Context, tenantID uuid.UUID) error {
	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("subscription not found")
		}
		return errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	// Cancel with Stripe if needed
	if subscription.StripeSubscriptionID != "" {
		// In real implementation, cancel with Stripe
		s.logger.Info("would cancel Stripe subscription", "stripe_id", subscription.StripeSubscriptionID)
	}

	if err := s.repos.Subscription.Delete(ctx, subscription.ID); err != nil {
		return errors.NewServiceError("SUBSCRIPTION_DELETE_FAILED", "failed to delete subscription", err)
	}

	s.logger.Info("subscription deleted", "tenant_id", tenantID)
	return nil
}

// ============================================================================
// Plan Management
// ============================================================================

// ChangePlan changes the subscription plan
func (s *subscriptionService) ChangePlan(ctx context.Context, tenantID uuid.UUID, req *dto.ChangePlanRequest) (*dto.BillingPreviewResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("subscription not found")
		}
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	// Check if already on the requested plan
	if subscription.Plan == req.NewPlan {
		return nil, errors.NewValidationError("already on the requested plan")
	}

	// Validate plan change direction
	isUpgrade := s.isPlanUpgrade(subscription.Plan, req.NewPlan)
	if !isUpgrade && !req.ChangeImmediate {
		// Downgrades at period end only
		if err := s.repos.Subscription.DowngradePlan(ctx, tenantID, req.NewPlan, false); err != nil {
			return nil, errors.NewServiceError("PLAN_CHANGE_FAILED", "failed to schedule plan change", err)
		}
	} else {
		// Upgrades or immediate changes
		if err := s.repos.Subscription.UpgradePlan(ctx, tenantID, req.NewPlan); err != nil {
			return nil, errors.NewServiceError("PLAN_CHANGE_FAILED", "failed to change plan", err)
		}
	}

	// Calculate billing preview
	preview := s.calculateBillingPreview(subscription, req.NewPlan, req.ChangeImmediate)

	s.logger.Info("plan changed", "tenant_id", tenantID, "old_plan", subscription.Plan, "new_plan", req.NewPlan, "immediate", req.ChangeImmediate)

	return preview, nil
}

// PreviewPlanChange previews the billing impact of a plan change
func (s *subscriptionService) PreviewPlanChange(ctx context.Context, tenantID uuid.UUID, req *dto.ChangePlanRequest) (*dto.BillingPreviewResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("subscription not found")
		}
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	preview := s.calculateBillingPreview(subscription, req.NewPlan, req.ChangeImmediate)
	return preview, nil
}

// GetPlanComparison returns comparison of all available plans
func (s *subscriptionService) GetPlanComparison(ctx context.Context) (*dto.PlanComparisonResponse, error) {
	plans := []models.SubscriptionPlan{
		models.PlanFree, models.PlanStarter, models.PlanPro,
		models.PlanBusiness, models.PlanEnterprise,
	}

	var planDetails []dto.PlanDetails
	for _, plan := range plans {
		details := dto.PlanDetails{
			Plan:         plan,
			Name:         s.getPlanName(plan),
			Description:  s.getPlanDescription(plan),
			MonthlyPrice: s.getPlanPricing(plan, models.BillingMonthly),
			YearlyPrice:  s.getPlanPricing(plan, models.BillingYearly),
			Features:     models.GetDefaultFeaturesForPlan(plan),
			Limits:       models.GetDefaultLimitsForPlan(plan),
			Popular:      plan == models.PlanPro,
			Recommended:  plan == models.PlanBusiness,
		}

		// Calculate yearly discount
		if details.MonthlyPrice > 0 && details.YearlyPrice > 0 {
			monthlyTotal := details.MonthlyPrice * 12
			if monthlyTotal > details.YearlyPrice {
				details.YearlyDiscount = ((monthlyTotal - details.YearlyPrice) / monthlyTotal) * 100
			}
		}

		planDetails = append(planDetails, details)
	}

	return &dto.PlanComparisonResponse{Plans: planDetails}, nil
}

// ============================================================================
// Subscription Lifecycle
// ============================================================================

// StartTrial starts a trial period
func (s *subscriptionService) StartTrial(ctx context.Context, tenantID uuid.UUID, trialDays int) (*dto.SubscriptionResponse, error) {
	if trialDays < 0 || trialDays > 90 {
		return nil, errors.NewValidationError("trial days must be between 0 and 90")
	}

	if err := s.repos.Subscription.StartTrial(ctx, tenantID, trialDays); err != nil {
		return nil, errors.NewServiceError("TRIAL_START_FAILED", "failed to start trial", err)
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	s.logger.Info("trial started", "tenant_id", tenantID, "days", trialDays)
	return dto.ToSubscriptionResponse(subscription), nil
}

// EndTrial ends the trial period and converts to paid
func (s *subscriptionService) EndTrial(ctx context.Context, tenantID uuid.UUID) (*dto.SubscriptionResponse, error) {
	if err := s.repos.Subscription.EndTrial(ctx, tenantID); err != nil {
		return nil, errors.NewServiceError("TRIAL_END_FAILED", "failed to end trial", err)
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	s.logger.Info("trial ended", "tenant_id", tenantID)
	return dto.ToSubscriptionResponse(subscription), nil
}

// CancelSubscription cancels a subscription
func (s *subscriptionService) CancelSubscription(ctx context.Context, tenantID uuid.UUID, req *dto.CancelSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	if err := s.repos.Subscription.CancelSubscription(ctx, tenantID, req.CancelAtPeriodEnd); err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_CANCEL_FAILED", "failed to cancel subscription", err)
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	// Log cancellation reason
	if req.Reason != "" {
		s.logger.Info("subscription canceled", "tenant_id", tenantID, "reason", req.Reason, "at_period_end", req.CancelAtPeriodEnd)
	}

	return dto.ToSubscriptionResponse(subscription), nil
}

// ReactivateSubscription reactivates a canceled subscription
func (s *subscriptionService) ReactivateSubscription(ctx context.Context, tenantID uuid.UUID, req *dto.ReactivateSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	if err := s.repos.Subscription.ReactivateSubscription(ctx, tenantID); err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_REACTIVATE_FAILED", "failed to reactivate subscription", err)
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	s.logger.Info("subscription reactivated", "tenant_id", tenantID)
	return dto.ToSubscriptionResponse(subscription), nil
}

// SuspendSubscription suspends a subscription
func (s *subscriptionService) SuspendSubscription(ctx context.Context, tenantID uuid.UUID, reason string) (*dto.SubscriptionResponse, error) {
	if err := s.repos.Subscription.SuspendSubscription(ctx, tenantID, reason); err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_SUSPEND_FAILED", "failed to suspend subscription", err)
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	s.logger.Info("subscription suspended", "tenant_id", tenantID, "reason", reason)
	return dto.ToSubscriptionResponse(subscription), nil
}

// ============================================================================
// Usage & Limits
// ============================================================================

// GetUsage retrieves current usage statistics
func (s *subscriptionService) GetUsage(ctx context.Context, tenantID uuid.UUID) (*dto.UsageResponse, error) {
	usageStats, err := s.repos.Subscription.GetUsageStats(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("USAGE_GET_FAILED", "failed to get usage statistics", err)
	}

	// Get additional usage data (team members, services, bookings)
	teamMembersUsed := 0   // TODO: Get from user repository
	servicesUsed := 0      // TODO: Get from service repository
	bookingsThisMonth := 0 // TODO: Get from booking repository

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	response := &dto.UsageResponse{
		TenantID:            tenantID,
		Plan:                usageStats.Plan,
		CustomersUsed:       usageStats.CustomersUsed,
		CustomersLimit:      usageStats.CustomersLimit,
		CustomersPercentage: usageStats.CustomersPercent,
		ProjectsUsed:        usageStats.ProjectsUsed,
		ProjectsLimit:       usageStats.ProjectsLimit,
		ProjectsPercentage:  usageStats.ProjectsPercent,
		StorageUsedGB:       usageStats.StorageUsedGB,
		StorageLimitGB:      usageStats.StorageLimitGB,
		StoragePercentage:   usageStats.StoragePercent,
		TeamMembersUsed:     teamMembersUsed,
		TeamMembersLimit:    subscription.MaxTeamMembers,
		ServicesUsed:        servicesUsed,
		ServicesLimit:       subscription.MaxServicesListed,
		BookingsThisMonth:   bookingsThisMonth,
		BookingsLimit:       subscription.MaxBookingsPerMonth,
	}

	// Calculate percentages
	response.TeamMembersPercentage = dto.CalculatePercentage(teamMembersUsed, subscription.MaxTeamMembers)
	response.ServicesPercentage = dto.CalculatePercentage(servicesUsed, subscription.MaxServicesListed)
	response.BookingsPercentage = dto.CalculatePercentage(bookingsThisMonth, subscription.MaxBookingsPerMonth)

	// Check for over-limit conditions
	var overLimitReasons []string
	if response.CustomersPercentage > 100 {
		overLimitReasons = append(overLimitReasons, "customers")
	}
	if response.ProjectsPercentage > 100 {
		overLimitReasons = append(overLimitReasons, "projects")
	}
	if response.StoragePercentage > 100 {
		overLimitReasons = append(overLimitReasons, "storage")
	}
	if response.TeamMembersPercentage > 100 {
		overLimitReasons = append(overLimitReasons, "team_members")
	}
	if response.ServicesPercentage > 100 {
		overLimitReasons = append(overLimitReasons, "services")
	}
	if response.BookingsPercentage > 100 {
		overLimitReasons = append(overLimitReasons, "bookings")
	}

	response.IsOverLimit = len(overLimitReasons) > 0
	response.OverLimitReasons = overLimitReasons

	return response, nil
}

// UpdateUsage updates usage statistics
func (s *subscriptionService) UpdateUsage(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateUsageRequest) error {
	if err := req.Validate(); err != nil {
		return errors.NewValidationError("invalid request: " + err.Error())
	}

	switch req.Operation {
	case "increment":
		return s.repos.Subscription.IncrementUsage(ctx, tenantID, req.UsageType, req.Amount)
	case "decrement":
		return s.repos.Subscription.DecrementUsage(ctx, tenantID, req.UsageType, req.Amount)
	case "set":
		// For set operation, we need to get current usage and calculate the difference
		usageStats, err := s.repos.Subscription.GetUsageStats(ctx, tenantID)
		if err != nil {
			return err
		}

		var currentValue int
		switch req.UsageType {
		case "customers":
			currentValue = usageStats.CustomersUsed
		case "projects":
			currentValue = usageStats.ProjectsUsed
		case "storage_gb":
			currentValue = usageStats.StorageUsedGB
		default:
			return fmt.Errorf("unsupported usage type for set operation: %s", req.UsageType)
		}

		diff := req.Amount - currentValue
		if diff > 0 {
			return s.repos.Subscription.IncrementUsage(ctx, tenantID, req.UsageType, diff)
		} else if diff < 0 {
			return s.repos.Subscription.DecrementUsage(ctx, tenantID, req.UsageType, -diff)
		}
		return nil
	default:
		return fmt.Errorf("unsupported operation: %s", req.Operation)
	}
}

// CheckLimits checks if a specific limit has been reached
func (s *subscriptionService) CheckLimits(ctx context.Context, tenantID uuid.UUID, limitType string) (bool, error) {
	return s.repos.Subscription.CheckLimit(ctx, tenantID, limitType)
}

// EnforceUsageLimits enforces usage limits for a tenant
func (s *subscriptionService) EnforceUsageLimits(ctx context.Context, tenantID uuid.UUID) error {
	usage, err := s.GetUsage(ctx, tenantID)
	if err != nil {
		return err
	}

	if usage.IsOverLimit {
		s.logger.Warn("tenant is over usage limits", "tenant_id", tenantID, "over_limit_reasons", usage.OverLimitReasons)

		// In a real implementation, you might:
		// 1. Send notifications to the tenant
		// 2. Temporarily suspend certain features
		// 3. Force an upgrade
		// 4. Block new resource creation

		return fmt.Errorf("tenant is over usage limits: %v", usage.OverLimitReasons)
	}

	return nil
}

// ResetMonthlyUsage resets monthly usage counters
func (s *subscriptionService) ResetMonthlyUsage(ctx context.Context, tenantID uuid.UUID) error {
	return s.repos.Subscription.ResetMonthlyUsage(ctx, tenantID)
}

// ============================================================================
// Helper Methods
// ============================================================================

// getPlanPricing returns the pricing for a plan and billing interval
func (s *subscriptionService) getPlanPricing(plan models.SubscriptionPlan, interval models.BillingInterval) float64 {
	pricing := map[models.SubscriptionPlan]map[models.BillingInterval]float64{
		models.PlanFree: {
			models.BillingMonthly:  0,
			models.BillingYearly:   0,
			models.BillingLifetime: 0,
		},
		models.PlanStarter: {
			models.BillingMonthly:  29.99,
			models.BillingYearly:   299.99, // ~17% discount
			models.BillingLifetime: 999.99,
		},
		models.PlanPro: {
			models.BillingMonthly:  79.99,
			models.BillingYearly:   799.99, // ~17% discount
			models.BillingLifetime: 2499.99,
		},
		models.PlanBusiness: {
			models.BillingMonthly:  199.99,
			models.BillingYearly:   1999.99, // ~17% discount
			models.BillingLifetime: 5999.99,
		},
		models.PlanEnterprise: {
			models.BillingMonthly:  499.99,
			models.BillingYearly:   4999.99, // ~17% discount
			models.BillingLifetime: 14999.99,
		},
	}

	if planPricing, exists := pricing[plan]; exists {
		if price, exists := planPricing[interval]; exists {
			return price
		}
	}

	return 0
}

// getPlanName returns the display name for a plan
func (s *subscriptionService) getPlanName(plan models.SubscriptionPlan) string {
	names := map[models.SubscriptionPlan]string{
		models.PlanFree:       "Free",
		models.PlanStarter:    "Starter",
		models.PlanPro:        "Professional",
		models.PlanBusiness:   "Business",
		models.PlanEnterprise: "Enterprise",
	}
	if name, exists := names[plan]; exists {
		return name
	}
	return string(plan)
}

// getPlanDescription returns the description for a plan
func (s *subscriptionService) getPlanDescription(plan models.SubscriptionPlan) string {
	descriptions := map[models.SubscriptionPlan]string{
		models.PlanFree:       "Perfect for getting started with basic features",
		models.PlanStarter:    "Ideal for small artisan businesses",
		models.PlanPro:        "Advanced features for growing businesses",
		models.PlanBusiness:   "Comprehensive solution for established businesses",
		models.PlanEnterprise: "Full-featured solution for large operations",
	}
	if desc, exists := descriptions[plan]; exists {
		return desc
	}
	return ""
}

// isPlanUpgrade determines if a plan change is an upgrade
func (s *subscriptionService) isPlanUpgrade(currentPlan, newPlan models.SubscriptionPlan) bool {
	planLevels := map[models.SubscriptionPlan]int{
		models.PlanFree:       0,
		models.PlanStarter:    1,
		models.PlanPro:        2,
		models.PlanBusiness:   3,
		models.PlanEnterprise: 4,
	}

	currentLevel, currentExists := planLevels[currentPlan]
	newLevel, newExists := planLevels[newPlan]

	if !currentExists || !newExists {
		return false
	}

	return newLevel > currentLevel
}

// calculateBillingPreview calculates billing preview for plan changes
func (s *subscriptionService) calculateBillingPreview(subscription *models.Subscription, newPlan models.SubscriptionPlan, immediate bool) *dto.BillingPreviewResponse {
	newAmount := s.getPlanPricing(newPlan, subscription.BillingInterval)

	var prorationAmount float64
	var effectiveDate time.Time
	var nextBillingDate time.Time

	if immediate {
		// Calculate proration
		now := time.Now()
		totalPeriod := subscription.CurrentPeriodEnd.Sub(subscription.CurrentPeriodStart)
		remainingPeriod := subscription.CurrentPeriodEnd.Sub(now)

		if totalPeriod > 0 && remainingPeriod > 0 {
			remainingRatio := float64(remainingPeriod) / float64(totalPeriod)

			// Credit remaining time on old plan
			oldCredit := subscription.Amount * remainingRatio

			// Charge remaining time on new plan
			newCharge := newAmount * remainingRatio

			prorationAmount = newCharge - oldCredit
		}

		effectiveDate = now
		nextBillingDate = subscription.CurrentPeriodEnd
	} else {
		prorationAmount = 0
		effectiveDate = subscription.CurrentPeriodEnd
		nextBillingDate = subscription.CurrentPeriodEnd
	}

	return &dto.BillingPreviewResponse{
		CurrentPlan:     subscription.Plan,
		NewPlan:         newPlan,
		ProrationAmount: prorationAmount,
		NewAmount:       newAmount,
		NextBillingDate: nextBillingDate,
		EffectiveDate:   effectiveDate,
		Currency:        subscription.Currency,
		ChangeImmediate: immediate,
	}
}

// applyPromoCodeInternal applies a promo code (internal helper)
func (s *subscriptionService) applyPromoCodeInternal(ctx context.Context, tenantID uuid.UUID, promoCode string) error {
	// In a real implementation, this would:
	// 1. Validate the promo code
	// 2. Check if it's still valid (not expired, usage limits)
	// 3. Calculate the discount
	// 4. Apply it to the subscription
	// 5. Record the usage

	s.logger.Info("promo code applied", "tenant_id", tenantID, "promo_code", promoCode)
	return nil
}

// ============================================================================
// Feature Management
// ============================================================================

// HasFeature checks if a tenant has access to a specific feature
func (s *subscriptionService) HasFeature(ctx context.Context, tenantID uuid.UUID, feature string) (bool, error) {
	return s.repos.Subscription.HasFeature(ctx, tenantID, feature)
}

// GetEnabledFeatures returns all enabled features for a tenant
func (s *subscriptionService) GetEnabledFeatures(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	return s.repos.Subscription.GetEnabledFeatures(ctx, tenantID)
}

// GetFeatureAccess returns detailed feature access information
func (s *subscriptionService) GetFeatureAccess(ctx context.Context, tenantID uuid.UUID) (*dto.FeatureAccessResponse, error) {
	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("subscription not found")
		}
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	enabledFeatures, err := s.repos.Subscription.GetEnabledFeatures(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("FEATURES_GET_FAILED", "failed to get enabled features", err)
	}

	// Create feature details map
	featureDetails := s.buildFeatureDetailsMap(subscription.Plan, subscription.Features)

	// Separate enabled and disabled features
	var disabledFeatures []string
	for featureName := range featureDetails {
		if !featureDetails[featureName].Enabled {
			disabledFeatures = append(disabledFeatures, featureName)
		}
	}

	return &dto.FeatureAccessResponse{
		TenantID:         tenantID,
		Plan:             subscription.Plan,
		EnabledFeatures:  enabledFeatures,
		DisabledFeatures: disabledFeatures,
		FeatureDetails:   featureDetails,
	}, nil
}

// buildFeatureDetailsMap creates a detailed feature map
func (s *subscriptionService) buildFeatureDetailsMap(plan models.SubscriptionPlan, features models.SubscriptionFeatures) map[string]dto.FeatureDetail {
	// This would be a comprehensive mapping of all features
	// For brevity, showing a few examples
	return map[string]dto.FeatureDetail{
		"basic_booking": {
			Name:        "Basic Booking",
			Description: "Create and manage basic bookings",
			Enabled:     features.BasicBooking,
		},
		"advanced_project_mgmt": {
			Name:            "Advanced Project Management",
			Description:     "Advanced project planning and tracking tools",
			Enabled:         features.AdvancedProjectMgmt,
			RequiredPlan:    models.PlanPro,
			UpgradeRequired: plan < models.PlanPro && !features.AdvancedProjectMgmt,
		},
		"white_labeling": {
			Name:            "White Labeling",
			Description:     "Remove branding and use custom branding",
			Enabled:         features.WhiteLabeling,
			RequiredPlan:    models.PlanEnterprise,
			UpgradeRequired: plan < models.PlanEnterprise && !features.WhiteLabeling,
		},
	}
}

// UpdateBillingInterval updates the billing interval
func (s *subscriptionService) UpdateBillingInterval(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateBillingIntervalRequest) (*dto.SubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("subscription not found")
		}
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	// Check if already on requested interval
	if subscription.BillingInterval == req.BillingInterval {
		return dto.ToSubscriptionResponse(subscription), nil
	}

	// Update pricing for new interval
	newAmount := s.getPlanPricing(subscription.Plan, req.BillingInterval)
	subscription.BillingInterval = req.BillingInterval
	subscription.Amount = newAmount

	// Update billing dates if immediate
	if req.ChangeImmediate {
		now := time.Now()
		subscription.CurrentPeriodStart = now

		switch req.BillingInterval {
		case models.BillingMonthly:
			subscription.CurrentPeriodEnd = now.AddDate(0, 1, 0)
		case models.BillingYearly:
			subscription.CurrentPeriodEnd = now.AddDate(1, 0, 0)
		case models.BillingLifetime:
			subscription.CurrentPeriodEnd = now.AddDate(10, 0, 0)
		}

		subscription.NextBillingDate = &subscription.CurrentPeriodEnd
	}

	if err := s.repos.Subscription.Update(ctx, subscription); err != nil {
		return nil, errors.NewServiceError("BILLING_INTERVAL_UPDATE_FAILED", "failed to update billing interval", err)
	}

	s.logger.Info("billing interval updated", "tenant_id", tenantID, "interval", req.BillingInterval, "immediate", req.ChangeImmediate)
	return dto.ToSubscriptionResponse(subscription), nil
}

// ProcessPayment processes a payment for a subscription
func (s *subscriptionService) ProcessPayment(ctx context.Context, tenantID uuid.UUID, req *dto.ProcessPaymentRequest) (*dto.PaymentResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("subscription not found")
		}
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	// Process payment through payment service
	// This is a placeholder - in real implementation, integrate with Stripe/PayPal/etc.
	paymentID := uuid.New()
	now := time.Now()

	// Record successful payment
	if err := s.repos.Subscription.RecordPayment(ctx, tenantID, req.Amount, now); err != nil {
		return nil, errors.NewServiceError("PAYMENT_RECORD_FAILED", "failed to record payment", err)
	}

	s.logger.Info("payment processed", "tenant_id", tenantID, "amount", req.Amount, "payment_id", paymentID)

	return &dto.PaymentResponse{
		ID:             paymentID,
		SubscriptionID: subscription.ID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         "paid",
		Method:         "card", // Default
		Description:    req.Description,
		ProcessedAt:    &now,
		CreatedAt:      now,
	}, nil
}

// ApplyPromoCode applies a promotional code to a subscription
func (s *subscriptionService) ApplyPromoCode(ctx context.Context, tenantID uuid.UUID, req *dto.ApplyPromoCodeRequest) (*dto.SubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	if err := s.applyPromoCodeInternal(ctx, tenantID, req.PromoCode); err != nil {
		return nil, errors.NewServiceError("PROMO_CODE_APPLY_FAILED", "failed to apply promo code", err)
	}

	subscription, err := s.repos.Subscription.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTION_GET_FAILED", "failed to get subscription", err)
	}

	return dto.ToSubscriptionResponse(subscription), nil
}

// GetInvoices retrieves invoices for a subscription
func (s *subscriptionService) GetInvoices(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.PaymentHistoryResponse, error) {
	// This would integrate with your invoice/billing system
	// Placeholder implementation
	return &dto.PaymentHistoryResponse{
		Payments:   []dto.PaymentResponse{},
		Page:       page,
		PageSize:   pageSize,
		TotalItems: 0,
		TotalPages: 0,
	}, nil
}

// GetPaymentHistory retrieves payment history for a subscription
func (s *subscriptionService) GetPaymentHistory(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.PaymentHistoryResponse, error) {
	// This would integrate with your payment system
	// Placeholder implementation
	return &dto.PaymentHistoryResponse{
		Payments:   []dto.PaymentResponse{},
		Page:       page,
		PageSize:   pageSize,
		TotalItems: 0,
		TotalPages: 0,
	}, nil
}

// ============================================================================
// Analytics & Reporting
// ============================================================================

// GetSubscriptionStats returns subscription statistics
func (s *subscriptionService) GetSubscriptionStats(ctx context.Context) (*dto.SubscriptionStatsResponse, error) {
	stats, err := s.repos.Subscription.GetSubscriptionStats(ctx)
	if err != nil {
		return nil, errors.NewServiceError("STATS_GET_FAILED", "failed to get subscription statistics", err)
	}

	// Convert repository stats to DTO
	response := &dto.SubscriptionStatsResponse{
		TotalSubscriptions:    stats.TotalSubscriptions,
		ActiveSubscriptions:   stats.ActiveSubscriptions,
		TrialSubscriptions:    stats.TrialSubscriptions,
		CanceledSubscriptions: stats.CanceledSubscriptions,
		ByPlan:                stats.ByPlan,
		ByStatus:              stats.ByStatus,
		TotalMRR:              stats.TotalMRR,
		TotalARR:              stats.TotalARR,
		AverageRevenuePerUser: stats.AverageRevenuePerUser,
		ChurnRate:             stats.ChurnRate,
	}

	// Add additional calculations
	if stats.TotalSubscriptions > 0 {
		response.GrowthRate = s.calculateGrowthRate(ctx)
		response.CustomerLifetimeValue = s.calculateCustomerLifetimeValue(ctx)
	}

	return response, nil
}

// GetSubscriptionList returns filtered list of subscriptions
func (s *subscriptionService) GetSubscriptionList(ctx context.Context, filter dto.SubscriptionFilter) (*dto.SubscriptionListResponse, error) {
	if err := filter.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid filter: " + err.Error())
	}

	// Convert DTO filter to repository filter
	repoFilter := repository.SubscriptionFilters{
		Plans:         filter.Plans,
		Statuses:      filter.Statuses,
		BillingCycles: filter.BillingIntervals,
		MinAmount:     filter.MinAmount,
		MaxAmount:     filter.MaxAmount,
		IsTrialing:    filter.IsTrialing,
		HasFailures:   filter.HasFailedPayments,
		CreatedAfter:  filter.CreatedAfter,
		CreatedBefore: filter.CreatedBefore,
	}

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	subscriptions, paginationResult, err := s.repos.Subscription.FindByFilters(ctx, repoFilter, pagination)
	if err != nil {
		return nil, errors.NewServiceError("SUBSCRIPTIONS_LIST_FAILED", "failed to list subscriptions", err)
	}

	// Calculate pagination manually if fields are missing
	totalPages := int((paginationResult.TotalItems + int64(paginationResult.PageSize) - 1) / int64(paginationResult.PageSize))
	hasNext := paginationResult.Page < totalPages
	hasPrevious := paginationResult.Page > 1

	return &dto.SubscriptionListResponse{
		Subscriptions: dto.ToSubscriptionResponses(subscriptions),
		Page:          paginationResult.Page,
		PageSize:      paginationResult.PageSize,
		TotalItems:    paginationResult.TotalItems,
		TotalPages:    totalPages,
		HasNext:       hasNext,
		HasPrevious:   hasPrevious,
	}, nil
}

// GetAnalytics returns general subscription analytics
func (s *subscriptionService) GetAnalytics(ctx context.Context, filter dto.AnalyticsFilter) (map[string]interface{}, error) {
	if err := filter.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid filter: " + err.Error())
	}

	// This would include complex analytics queries
	// Placeholder implementation
	analytics := map[string]interface{}{
		"period":        fmt.Sprintf("%s to %s", filter.StartDate.Format("2006-01-02"), filter.EndDate.Format("2006-01-02")),
		"group_by":      filter.GroupBy,
		"total_revenue": 0.0,
		"new_signups":   0,
		"cancellations": 0,
		"upgrades":      0,
		"downgrades":    0,
	}

	return analytics, nil
}

// GetChurnAnalysis returns churn analysis
func (s *subscriptionService) GetChurnAnalysis(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	churnRate, err := s.repos.Subscription.GetChurnRate(ctx, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("CHURN_ANALYSIS_FAILED", "failed to calculate churn analysis", err)
	}

	analysis := map[string]interface{}{
		"period":               fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"churn_rate":           churnRate,
		"churn_by_plan":        map[string]float64{},
		"cancellation_reasons": map[string]int{},
		"retention_rate":       100 - churnRate,
	}

	return analysis, nil
}

// GetRevenueAnalysis returns revenue analysis
func (s *subscriptionService) GetRevenueAnalysis(ctx context.Context, filter dto.AnalyticsFilter) (map[string]interface{}, error) {
	if err := filter.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid filter: " + err.Error())
	}

	revenueByPlan, err := s.repos.Subscription.GetRevenueByPlan(ctx, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, errors.NewServiceError("REVENUE_ANALYSIS_FAILED", "failed to calculate revenue analysis", err)
	}

	mrr, _ := s.repos.Subscription.GetMRR(ctx)
	arr, _ := s.repos.Subscription.GetARR(ctx)

	analysis := map[string]interface{}{
		"period":          fmt.Sprintf("%s to %s", filter.StartDate.Format("2006-01-02"), filter.EndDate.Format("2006-01-02")),
		"revenue_by_plan": revenueByPlan,
		"current_mrr":     mrr,
		"current_arr":     arr,
		"currency":        filter.Currency,
	}

	return analysis, nil
}

// ============================================================================
// Background Operations
// ============================================================================

// ProcessExpiringTrials processes trials that are expiring soon
func (s *subscriptionService) ProcessExpiringTrials(ctx context.Context) error {
	// Get trials expiring in 3 days
	expiringTrials, _, err := s.repos.Subscription.GetExpiringTrials(ctx, 3, repository.PaginationParams{Page: 1, PageSize: 100})
	if err != nil {
		return errors.NewServiceError("EXPIRING_TRIALS_GET_FAILED", "failed to get expiring trials", err)
	}

	processed := 0
	for _, trial := range expiringTrials {
		// Send notification (implement notification service)
		s.logger.Info("trial expiring soon", "tenant_id", trial.TenantID, "expires_at", trial.TrialEndsAt)
		processed++
	}

	s.logger.Info("processed expiring trials", "count", processed)
	return nil
}

// ProcessFailedPayments processes subscriptions with failed payments
func (s *subscriptionService) ProcessFailedPayments(ctx context.Context) error {
	failedSubs, err := s.repos.Subscription.GetSubscriptionsWithFailedPayments(ctx, 1)
	if err != nil {
		return errors.NewServiceError("FAILED_PAYMENTS_GET_FAILED", "failed to get subscriptions with failed payments", err)
	}

	processed := 0
	for _, sub := range failedSubs {
		// Retry payment or suspend subscription based on failure count
		if sub.FailedPayments >= 3 {
			err := s.repos.Subscription.SuspendSubscription(ctx, sub.TenantID, "Payment failures")
			if err != nil {
				s.logger.Error("failed to suspend subscription", "tenant_id", sub.TenantID, "error", err)
				continue
			}
			s.logger.Info("subscription suspended due to payment failures", "tenant_id", sub.TenantID)
		} else {
			// Send payment reminder
			s.logger.Info("payment reminder sent", "tenant_id", sub.TenantID, "failures", sub.FailedPayments)
		}
		processed++
	}

	s.logger.Info("processed failed payments", "count", processed)
	return nil
}

// ProcessSubscriptionRenewals processes subscription renewals
func (s *subscriptionService) ProcessSubscriptionRenewals(ctx context.Context) error {
	return s.repos.Subscription.AutoRenewSubscriptions(ctx)
}

// CleanupExpiredSubscriptions cleans up expired subscriptions
func (s *subscriptionService) CleanupExpiredSubscriptions(ctx context.Context) error {
	expiredSubs, err := s.repos.Subscription.GetExpiredSubscriptions(ctx)
	if err != nil {
		return errors.NewServiceError("EXPIRED_SUBSCRIPTIONS_GET_FAILED", "failed to get expired subscriptions", err)
	}

	processed := 0
	for _, sub := range expiredSubs {
		// Convert expired trials to free plan or suspend
		if sub.Status == models.SubStatusTrialing {
			err := s.repos.Subscription.UpgradePlan(ctx, sub.TenantID, models.PlanFree)
			if err != nil {
				s.logger.Error("failed to convert expired trial to free", "tenant_id", sub.TenantID, "error", err)
				continue
			}
			s.logger.Info("expired trial converted to free plan", "tenant_id", sub.TenantID)
		} else {
			err := s.repos.Subscription.SuspendSubscription(ctx, sub.TenantID, "Expired subscription")
			if err != nil {
				s.logger.Error("failed to suspend expired subscription", "tenant_id", sub.TenantID, "error", err)
				continue
			}
			s.logger.Info("expired subscription suspended", "tenant_id", sub.TenantID)
		}
		processed++
	}

	s.logger.Info("processed expired subscriptions", "count", processed)
	return nil
}

// ============================================================================
// Health & Monitoring
// ============================================================================

// HealthCheck performs a health check on the subscription service
func (s *subscriptionService) HealthCheck(ctx context.Context) error {
	// Test database connectivity
	_, err := s.repos.Subscription.GetSubscriptionStats(ctx)
	if err != nil {
		return fmt.Errorf("subscription repository health check failed: %w", err)
	}

	return nil
}

// GetServiceMetrics returns service metrics
func (s *subscriptionService) GetServiceMetrics(ctx context.Context) map[string]interface{} {
	stats, _ := s.repos.Subscription.GetSubscriptionStats(ctx)

	return map[string]interface{}{
		"total_subscriptions":  stats.TotalSubscriptions,
		"active_subscriptions": stats.ActiveSubscriptions,
		"total_mrr":            stats.TotalMRR,
		"total_arr":            stats.TotalARR,
		"service_status":       "healthy",
	}
}

// Helper methods for calculations
func (s *subscriptionService) calculateGrowthRate(ctx context.Context) float64 {
	// Calculate month-over-month growth rate
	// Placeholder implementation
	return 5.2 // 5.2% growth
}

func (s *subscriptionService) calculateCustomerLifetimeValue(ctx context.Context) float64 {
	// Calculate average customer lifetime value
	// Placeholder implementation
	return 1250.0 // $1,250 CLV
}
