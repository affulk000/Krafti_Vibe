package dto

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Request DTOs
// ============================================================================

// CreateSubscriptionRequest represents the request to create a subscription
type CreateSubscriptionRequest struct {
	TenantID        uuid.UUID               `json:"tenant_id" validate:"required"`
	Plan            models.SubscriptionPlan `json:"plan" validate:"required"`
	BillingInterval models.BillingInterval  `json:"billing_interval" validate:"required"`
	PaymentMethodID string                  `json:"payment_method_id,omitempty"`
	PromoCode       string                  `json:"promo_code,omitempty"`
	TrialDays       int                     `json:"trial_days" validate:"min=0,max=90"`
	Metadata        map[string]any          `json:"metadata,omitempty"`
}

// Validate validates the create subscription request
func (r *CreateSubscriptionRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if r.Plan == "" {
		return fmt.Errorf("subscription plan is required")
	}
	if r.BillingInterval == "" {
		return fmt.Errorf("billing interval is required")
	}
	if r.TrialDays < 0 || r.TrialDays > 90 {
		return fmt.Errorf("trial days must be between 0 and 90")
	}

	// Validate plan
	validPlans := []models.SubscriptionPlan{
		models.PlanFree, models.PlanStarter, models.PlanPro,
		models.PlanBusiness, models.PlanEnterprise,
	}
	valid := slices.Contains(validPlans, r.Plan)
	if !valid {
		return fmt.Errorf("invalid subscription plan: %s", r.Plan)
	}

	// Validate billing interval
	validIntervals := []models.BillingInterval{
		models.BillingMonthly, models.BillingYearly, models.BillingLifetime,
	}
	valid = slices.Contains(validIntervals, r.BillingInterval)
	if !valid {
		return fmt.Errorf("invalid billing interval: %s", r.BillingInterval)
	}

	return nil
}

// UpdateSubscriptionRequest represents the request to update a subscription
type UpdateSubscriptionRequest struct {
	PaymentMethodID *string        `json:"payment_method_id,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// ChangePlanRequest represents the request to change subscription plan
type ChangePlanRequest struct {
	NewPlan         models.SubscriptionPlan `json:"new_plan" validate:"required"`
	ChangeImmediate bool                    `json:"change_immediate" validate:"omitempty"`
	PromoCode       string                  `json:"promo_code,omitempty"`
}

// Validate validates the change plan request
func (r *ChangePlanRequest) Validate() error {
	if r.NewPlan == "" {
		return fmt.Errorf("new plan is required")
	}

	validPlans := []models.SubscriptionPlan{
		models.PlanFree, models.PlanStarter, models.PlanPro,
		models.PlanBusiness, models.PlanEnterprise,
	}
	valid := slices.Contains(validPlans, r.NewPlan)
	if !valid {
		return fmt.Errorf("invalid subscription plan: %s", r.NewPlan)
	}

	return nil
}

// CancelSubscriptionRequest represents the request to cancel a subscription
type CancelSubscriptionRequest struct {
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
	Reason            string `json:"reason,omitempty" validate:"max=500"`
	Feedback          string `json:"feedback,omitempty" validate:"max=1000"`
}

// Validate validates the cancel subscription request
func (r *CancelSubscriptionRequest) Validate() error {
	if len(r.Reason) > 500 {
		return fmt.Errorf("reason must be 500 characters or less")
	}
	if len(r.Feedback) > 1000 {
		return fmt.Errorf("feedback must be 1000 characters or less")
	}
	return nil
}

// ReactivateSubscriptionRequest represents the request to reactivate a subscription
type ReactivateSubscriptionRequest struct {
	PaymentMethodID string         `json:"payment_method_id,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// UpdateBillingIntervalRequest represents the request to update billing interval
type UpdateBillingIntervalRequest struct {
	BillingInterval models.BillingInterval `json:"billing_interval" validate:"required"`
	ChangeImmediate bool                   `json:"change_immediate"`
}

// Validate validates the update billing interval request
func (r *UpdateBillingIntervalRequest) Validate() error {
	if r.BillingInterval == "" {
		return fmt.Errorf("billing interval is required")
	}

	validIntervals := []models.BillingInterval{
		models.BillingMonthly, models.BillingYearly, models.BillingLifetime,
	}
	valid := slices.Contains(validIntervals, r.BillingInterval)
	if !valid {
		return fmt.Errorf("invalid billing interval: %s", r.BillingInterval)
	}

	return nil
}

// ProcessPaymentRequest represents the request to process a payment
type ProcessPaymentRequest struct {
	Amount          float64        `json:"amount" validate:"required,min=0"`
	Currency        string         `json:"currency" validate:"required,len=3"`
	PaymentMethodID string         `json:"payment_method_id" validate:"required"`
	Description     string         `json:"description,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// Validate validates the process payment request
func (r *ProcessPaymentRequest) Validate() error {
	if r.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if len(r.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-character code")
	}
	if strings.TrimSpace(r.PaymentMethodID) == "" {
		return fmt.Errorf("payment method ID is required")
	}
	return nil
}

// ApplyPromoCodeRequest represents the request to apply a promo code
type ApplyPromoCodeRequest struct {
	PromoCode string `json:"promo_code" validate:"required"`
}

// Validate validates the apply promo code request
func (r *ApplyPromoCodeRequest) Validate() error {
	if strings.TrimSpace(r.PromoCode) == "" {
		return fmt.Errorf("promo code is required")
	}
	return nil
}

// UpdateUsageRequest represents the request to update usage
type UpdateUsageRequest struct {
	UsageType string `json:"usage_type" validate:"required"`
	Amount    int    `json:"amount" validate:"required"`
	Operation string `json:"operation" validate:"required"` // increment, decrement, set
}

// Validate validates the update usage request
func (r *UpdateUsageRequest) Validate() error {
	if strings.TrimSpace(r.UsageType) == "" {
		return fmt.Errorf("usage type is required")
	}

	validTypes := []string{"customers", "projects", "storage_gb", "team_members", "services", "bookings"}
	valid := slices.Contains(validTypes, r.UsageType)
	if !valid {
		return fmt.Errorf("invalid usage type: %s", r.UsageType)
	}

	validOps := []string{"increment", "decrement", "set"}
	valid = slices.Contains(validOps, r.Operation)
	if !valid {
		return fmt.Errorf("invalid operation: %s", r.Operation)
	}

	if r.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	return nil
}

// ============================================================================
// Filter DTOs
// ============================================================================

// SubscriptionFilter represents filters for listing subscriptions
type SubscriptionFilter struct {
	Plans             []models.SubscriptionPlan   `json:"plans,omitempty"`
	Statuses          []models.SubscriptionStatus `json:"statuses,omitempty"`
	BillingIntervals  []models.BillingInterval    `json:"billing_intervals,omitempty"`
	MinAmount         *float64                    `json:"min_amount,omitempty"`
	MaxAmount         *float64                    `json:"max_amount,omitempty"`
	IsTrialing        *bool                       `json:"is_trialing,omitempty"`
	HasFailedPayments *bool                       `json:"has_failed_payments,omitempty"`
	DueForRenewal     *bool                       `json:"due_for_renewal,omitempty"`
	CreatedAfter      *time.Time                  `json:"created_after,omitempty"`
	CreatedBefore     *time.Time                  `json:"created_before,omitempty"`
	TrialEndingInDays *int                        `json:"trial_ending_in_days,omitempty"`
	Page              int                         `json:"page" validate:"min=1"`
	PageSize          int                         `json:"page_size" validate:"min=1,max=100"`
	SortBy            string                      `json:"sort_by,omitempty"`
	SortOrder         string                      `json:"sort_order,omitempty"` // asc or desc
}

// Validate validates the subscription filter
func (f *SubscriptionFilter) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}
	f.PageSize = max(1, min(f.PageSize, 100))
	if f.PageSize == 0 {
		f.PageSize = 20
	}
	if f.SortOrder != "" && f.SortOrder != "asc" && f.SortOrder != "desc" {
		return fmt.Errorf("sort order must be 'asc' or 'desc'")
	}
	if f.MinAmount != nil && *f.MinAmount < 0 {
		return fmt.Errorf("min amount cannot be negative")
	}
	if f.MaxAmount != nil && *f.MaxAmount < 0 {
		return fmt.Errorf("max amount cannot be negative")
	}
	if f.MinAmount != nil && f.MaxAmount != nil && *f.MinAmount > *f.MaxAmount {
		return fmt.Errorf("min amount cannot be greater than max amount")
	}
	if f.TrialEndingInDays != nil && *f.TrialEndingInDays < 0 {
		return fmt.Errorf("trial ending in days cannot be negative")
	}
	return nil
}

// AnalyticsFilter represents filters for subscription analytics
type AnalyticsFilter struct {
	StartDate     time.Time                   `json:"start_date" validate:"required"`
	EndDate       time.Time                   `json:"end_date" validate:"required"`
	GroupBy       string                      `json:"group_by,omitempty"` // day, week, month, year
	Plans         []models.SubscriptionPlan   `json:"plans,omitempty"`
	Statuses      []models.SubscriptionStatus `json:"statuses,omitempty"`
	IncludeTrials bool                        `json:"include_trials"`
	Currency      string                      `json:"currency,omitempty"`
}

// Validate validates the analytics filter
func (f *AnalyticsFilter) Validate() error {
	if f.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if f.EndDate.IsZero() {
		return fmt.Errorf("end date is required")
	}
	if f.StartDate.After(f.EndDate) {
		return fmt.Errorf("start date cannot be after end date")
	}

	if f.GroupBy != "" {
		validGroupBy := []string{"day", "week", "month", "year"}

		if !slices.Contains(validGroupBy, f.GroupBy) {
			return fmt.Errorf("invalid group by: %s", f.GroupBy)
		}
	}

	return nil
}

// ============================================================================
// Response DTOs
// ============================================================================

// SubscriptionResponse represents a subscription response
type SubscriptionResponse struct {
	ID              uuid.UUID                 `json:"id"`
	TenantID        uuid.UUID                 `json:"tenant_id"`
	Plan            models.SubscriptionPlan   `json:"plan"`
	Status          models.SubscriptionStatus `json:"status"`
	BillingInterval models.BillingInterval    `json:"billing_interval"`
	Amount          float64                   `json:"amount"`
	Currency        string                    `json:"currency"`
	DiscountPercent float64                   `json:"discount_percent"`

	// Lifecycle dates
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`

	// Payment info
	StripeSubscriptionID string `json:"stripe_subscription_id,omitempty"`
	StripeCustomerID     string `json:"stripe_customer_id,omitempty"`
	PaymentMethodID      string `json:"payment_method_id,omitempty"`

	// Usage & limits
	MaxCustomers        int `json:"max_customers"`
	MaxProjects         int `json:"max_projects"`
	MaxStorageGB        int `json:"max_storage_gb"`
	MaxTeamMembers      int `json:"max_team_members"`
	MaxServicesListed   int `json:"max_services_listed"`
	MaxBookingsPerMonth int `json:"max_bookings_per_month"`
	CurrentStorageGB    int `json:"current_storage_gb"`
	CurrentCustomers    int `json:"current_customers"`
	CurrentProjects     int `json:"current_projects"`

	// Features
	Features models.SubscriptionFeatures `json:"features"`

	// Billing
	NextBillingDate   *time.Time `json:"next_billing_date,omitempty"`
	LastPaymentDate   *time.Time `json:"last_payment_date,omitempty"`
	LastPaymentAmount float64    `json:"last_payment_amount"`
	FailedPayments    int        `json:"failed_payments"`

	// Calculated fields
	IsTrialing       bool               `json:"is_trialing"`
	DaysUntilExpiry  *int               `json:"days_until_expiry,omitempty"`
	DaysUntilRenewal *int               `json:"days_until_renewal,omitempty"`
	UsagePercentages map[string]float64 `json:"usage_percentages"`

	// Metadata
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// SubscriptionListResponse represents a paginated list of subscriptions
type SubscriptionListResponse struct {
	Subscriptions []*SubscriptionResponse `json:"subscriptions"`
	Page          int                     `json:"page"`
	PageSize      int                     `json:"page_size"`
	TotalItems    int64                   `json:"total_items"`
	TotalPages    int                     `json:"total_pages"`
	HasNext       bool                    `json:"has_next"`
	HasPrevious   bool                    `json:"has_previous"`
}

// SubscriptionStatsResponse represents subscription statistics
type SubscriptionStatsResponse struct {
	TotalSubscriptions    int64 `json:"total_subscriptions"`
	ActiveSubscriptions   int64 `json:"active_subscriptions"`
	TrialSubscriptions    int64 `json:"trial_subscriptions"`
	CanceledSubscriptions int64 `json:"canceled_subscriptions"`
	PastDueSubscriptions  int64 `json:"past_due_subscriptions"`

	ByPlan   map[models.SubscriptionPlan]int64   `json:"by_plan"`
	ByStatus map[models.SubscriptionStatus]int64 `json:"by_status"`

	TotalMRR              float64 `json:"total_mrr"`
	TotalARR              float64 `json:"total_arr"`
	AverageRevenuePerUser float64 `json:"average_revenue_per_user"`
	ChurnRate             float64 `json:"churn_rate"`
	GrowthRate            float64 `json:"growth_rate"`

	RevenueByPlan         map[models.SubscriptionPlan]float64 `json:"revenue_by_plan"`
	CustomerLifetimeValue float64                             `json:"customer_lifetime_value"`
}

// UsageResponse represents current usage statistics
type UsageResponse struct {
	TenantID uuid.UUID               `json:"tenant_id"`
	Plan     models.SubscriptionPlan `json:"plan"`

	CustomersUsed       int     `json:"customers_used"`
	CustomersLimit      int     `json:"customers_limit"`
	CustomersPercentage float64 `json:"customers_percentage"`

	ProjectsUsed       int     `json:"projects_used"`
	ProjectsLimit      int     `json:"projects_limit"`
	ProjectsPercentage float64 `json:"projects_percentage"`

	StorageUsedGB     int     `json:"storage_used_gb"`
	StorageLimitGB    int     `json:"storage_limit_gb"`
	StoragePercentage float64 `json:"storage_percentage"`

	TeamMembersUsed       int     `json:"team_members_used"`
	TeamMembersLimit      int     `json:"team_members_limit"`
	TeamMembersPercentage float64 `json:"team_members_percentage"`

	ServicesUsed       int     `json:"services_used"`
	ServicesLimit      int     `json:"services_limit"`
	ServicesPercentage float64 `json:"services_percentage"`

	BookingsThisMonth  int     `json:"bookings_this_month"`
	BookingsLimit      int     `json:"bookings_limit"`
	BookingsPercentage float64 `json:"bookings_percentage"`

	IsOverLimit      bool     `json:"is_over_limit"`
	OverLimitReasons []string `json:"over_limit_reasons,omitempty"`
}

// PlanComparisonResponse represents plan comparison data
type PlanComparisonResponse struct {
	Plans []PlanDetails `json:"plans"`
}

// PlanDetails represents detailed information about a plan
type PlanDetails struct {
	Plan           models.SubscriptionPlan     `json:"plan"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	MonthlyPrice   float64                     `json:"monthly_price"`
	YearlyPrice    float64                     `json:"yearly_price"`
	YearlyDiscount float64                     `json:"yearly_discount"`
	Features       models.SubscriptionFeatures `json:"features"`
	Limits         map[string]int              `json:"limits"`
	Popular        bool                        `json:"popular"`
	Recommended    bool                        `json:"recommended"`
}

// PaymentHistoryResponse represents payment history
type PaymentHistoryResponse struct {
	Payments   []PaymentResponse `json:"payments"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalItems int64             `json:"total_items"`
	TotalPages int               `json:"total_pages"`
}

// PaymentResponse represents a payment record
type PaymentResponse struct {
	ID             uuid.UUID  `json:"id"`
	SubscriptionID uuid.UUID  `json:"subscription_id"`
	Amount         float64    `json:"amount"`
	Currency       string     `json:"currency"`
	Status         string     `json:"status"`
	Method         string     `json:"method"`
	Description    string     `json:"description"`
	FailureReason  string     `json:"failure_reason,omitempty"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// FeatureAccessResponse represents feature access information
type FeatureAccessResponse struct {
	TenantID         uuid.UUID                `json:"tenant_id"`
	Plan             models.SubscriptionPlan  `json:"plan"`
	EnabledFeatures  []string                 `json:"enabled_features"`
	DisabledFeatures []string                 `json:"disabled_features"`
	FeatureDetails   map[string]FeatureDetail `json:"feature_details"`
}

// FeatureDetail represents detailed information about a feature
type FeatureDetail struct {
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Enabled         bool                    `json:"enabled"`
	RequiredPlan    models.SubscriptionPlan `json:"required_plan,omitempty"`
	UpgradeRequired bool                    `json:"upgrade_required"`
}

// BillingPreviewResponse represents billing preview for plan changes
type BillingPreviewResponse struct {
	CurrentPlan     models.SubscriptionPlan `json:"current_plan"`
	NewPlan         models.SubscriptionPlan `json:"new_plan"`
	ProrationAmount float64                 `json:"proration_amount"`
	NewAmount       float64                 `json:"new_amount"`
	NextBillingDate time.Time               `json:"next_billing_date"`
	EffectiveDate   time.Time               `json:"effective_date"`
	Currency        string                  `json:"currency"`
	ChangeImmediate bool                    `json:"change_immediate"`
}

// ============================================================================
// Utility Functions
// ============================================================================

// ToSubscriptionResponse converts a models.Subscription to SubscriptionResponse
func ToSubscriptionResponse(subscription *models.Subscription) *SubscriptionResponse {
	if subscription == nil {
		return nil
	}

	response := &SubscriptionResponse{
		ID:                   subscription.ID,
		TenantID:             subscription.TenantID,
		Plan:                 subscription.Plan,
		Status:               subscription.Status,
		BillingInterval:      subscription.BillingInterval,
		Amount:               subscription.Amount,
		Currency:             subscription.Currency,
		DiscountPercent:      subscription.DiscountPercent,
		TrialEndsAt:          subscription.TrialEndsAt,
		CurrentPeriodStart:   subscription.CurrentPeriodStart,
		CurrentPeriodEnd:     subscription.CurrentPeriodEnd,
		CanceledAt:           subscription.CanceledAt,
		CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd,
		StripeSubscriptionID: subscription.StripeSubscriptionID,
		StripeCustomerID:     subscription.StripeCustomerID,
		PaymentMethodID:      subscription.PaymentMethodID,
		MaxCustomers:         subscription.MaxCustomers,
		MaxProjects:          subscription.MaxProjects,
		MaxStorageGB:         subscription.MaxStorageGB,
		MaxTeamMembers:       subscription.MaxTeamMembers,
		MaxServicesListed:    subscription.MaxServicesListed,
		MaxBookingsPerMonth:  subscription.MaxBookingsPerMonth,
		CurrentStorageGB:     subscription.CurrentStorageGB,
		CurrentCustomers:     subscription.CurrentCustomers,
		CurrentProjects:      subscription.CurrentProjects,
		Features:             subscription.Features,
		NextBillingDate:      subscription.NextBillingDate,
		LastPaymentDate:      subscription.LastPaymentDate,
		LastPaymentAmount:    subscription.LastPaymentAmount,
		FailedPayments:       subscription.FailedPayments,
		Metadata:             subscription.Metadata,
		CreatedAt:            subscription.CreatedAt,
		UpdatedAt:            subscription.UpdatedAt,
	}

	// Calculate trial status
	response.IsTrialing = subscription.IsTrialing()

	// Calculate days until expiry/renewal
	if subscription.TrialEndsAt != nil && response.IsTrialing {
		days := int(time.Until(*subscription.TrialEndsAt).Hours() / 24)
		if days >= 0 {
			response.DaysUntilExpiry = &days
		}
	}

	if subscription.NextBillingDate != nil {
		days := int(time.Until(*subscription.NextBillingDate).Hours() / 24)
		if days >= 0 {
			response.DaysUntilRenewal = &days
		}
	}

	// Calculate usage percentages
	response.UsagePercentages = map[string]float64{
		"customers": calculatePercentage(subscription.CurrentCustomers, subscription.MaxCustomers),
		"projects":  calculatePercentage(subscription.CurrentProjects, subscription.MaxProjects),
		"storage":   calculatePercentage(subscription.CurrentStorageGB, subscription.MaxStorageGB),
	}

	return response
}

// ToSubscriptionResponses converts multiple models.Subscription to SubscriptionResponse slice
func ToSubscriptionResponses(subscriptions []*models.Subscription) []*SubscriptionResponse {
	responses := make([]*SubscriptionResponse, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		responses = append(responses, ToSubscriptionResponse(subscription))
	}
	return responses
}

// Helper function to calculate percentage
func calculatePercentage(current, max int) float64 {
	if max == -1 || max == 0 { // Unlimited or no limit
		return 0.0
	}
	if current > max {
		return 100.0
	}
	return (float64(current) / float64(max)) * 100.0
}

// CalculatePercentage is the exported version for use in other packages
func CalculatePercentage(current, max int) float64 {
	return calculatePercentage(current, max)
}
