package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan string

const (
	PlanFree       SubscriptionPlan = "free"       // Trial/Basic - Limited features
	PlanStarter    SubscriptionPlan = "starter"    // Small artisan business
	PlanPro        SubscriptionPlan = "pro"        // Growing business
	PlanBusiness   SubscriptionPlan = "business"   // Established business
	PlanEnterprise SubscriptionPlan = "enterprise" // Large operations
)

type SubscriptionStatus string

const (
	SubStatusActive    SubscriptionStatus = "active"
	SubStatusTrialing  SubscriptionStatus = "trialing"
	SubStatusPastDue   SubscriptionStatus = "past_due"
	SubStatusCanceled  SubscriptionStatus = "canceled"
	SubStatusSuspended SubscriptionStatus = "suspended"
)

type BillingInterval string

const (
	BillingMonthly  BillingInterval = "monthly"
	BillingYearly   BillingInterval = "yearly"
	BillingLifetime BillingInterval = "lifetime"
)

// Subscription represents tenant's subscription details
type Subscription struct {
	BaseModel
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;uniqueIndex;not null"`

	// Plan Details
	Plan            SubscriptionPlan   `json:"plan" gorm:"type:varchar(50);not null;default:'free'"`
	Status          SubscriptionStatus `json:"status" gorm:"type:varchar(50);not null;default:'trialing'"`
	BillingInterval BillingInterval    `json:"billing_interval" gorm:"type:varchar(20);not null;default:'monthly'"`

	// Pricing
	Amount          float64 `json:"amount" gorm:"type:decimal(10,2);not null;default:0"`
	Currency        string  `json:"currency" gorm:"size:3;default:'USD'"`
	DiscountPercent float64 `json:"discount_percent" gorm:"type:decimal(5,2);default:0"`

	// Subscription Lifecycle
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	CurrentPeriodStart time.Time  `json:"current_period_start" gorm:"not null"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end" gorm:"not null"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" gorm:"default:false"`

	// Payment Gateway References
	StripeSubscriptionID string `json:"stripe_subscription_id,omitempty" gorm:"uniqueIndex;size:255"`
	StripeCustomerID     string `json:"stripe_customer_id,omitempty" gorm:"size:255"`
	PaymentMethodID      string `json:"payment_method_id,omitempty" gorm:"size:255"`

	// Usage & Limits (enforced based on plan)
	MaxCustomers        int `json:"max_customers" gorm:"not null;default:10"`
	MaxProjects         int `json:"max_projects" gorm:"not null;default:5"`
	MaxStorageGB        int `json:"max_storage_gb" gorm:"not null;default:1"`
	MaxTeamMembers      int `json:"max_team_members" gorm:"not null;default:1"`
	MaxServicesListed   int `json:"max_services_listed" gorm:"not null;default:5"`
	MaxBookingsPerMonth int `json:"max_bookings_per_month" gorm:"not null;default:50"`
	CurrentStorageGB    int `json:"current_storage_gb" gorm:"default:0"`
	CurrentCustomers    int `json:"current_customers" gorm:"default:0"`
	CurrentProjects     int `json:"current_projects" gorm:"default:0"`

	// Feature Flags (enabled based on plan)
	Features SubscriptionFeatures `json:"features" gorm:"type:jsonb"`

	// Billing History
	NextBillingDate   *time.Time `json:"next_billing_date,omitempty"`
	LastPaymentDate   *time.Time `json:"last_payment_date,omitempty"`
	LastPaymentAmount float64    `json:"last_payment_amount" gorm:"type:decimal(10,2);default:0"`
	FailedPayments    int        `json:"failed_payments" gorm:"default:0"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// SubscriptionFeatures defines what features are enabled for the subscription
type SubscriptionFeatures struct {
	// Core Features
	BasicBooking       bool `json:"basic_booking"`       // All plans
	CustomerManagement bool `json:"customer_management"` // All plans
	ServiceCatalog     bool `json:"service_catalog"`     // All plans

	// Communication
	InAppMessaging      bool `json:"in_app_messaging"`     // Starter+
	EmailNotifications  bool `json:"email_notifications"`  // Starter+
	SMSNotifications    bool `json:"sms_notifications"`    // Pro+
	WhatsAppIntegration bool `json:"whatsapp_integration"` // Business+

	// Project Management
	BasicProjects       bool `json:"basic_projects"`        // Starter+
	AdvancedProjectMgmt bool `json:"advanced_project_mgmt"` // Pro+
	ProjectTemplates    bool `json:"project_templates"`     // Business+
	GanttCharts         bool `json:"gantt_charts"`          // Business+
	ResourceAllocation  bool `json:"resource_allocation"`   // Enterprise

	// Team & Collaboration
	TeamMembers     bool `json:"team_members"`      // Pro+
	RoleBasedAccess bool `json:"role_based_access"` // Pro+
	TeamChat        bool `json:"team_chat"`         // Business+
	TaskAssignment  bool `json:"task_assignment"`   // Pro+

	// Financial
	InvoiceGeneration bool `json:"invoice_generation"`  // All plans
	OnlinePayments    bool `json:"online_payments"`     // Starter+
	RecurringBilling  bool `json:"recurring_billing"`   // Pro+
	MultiCurrency     bool `json:"multi_currency"`      // Business+
	ExpenseTracking   bool `json:"expense_tracking"`    // Pro+
	ProfitLossReports bool `json:"profit_loss_reports"` // Business+

	// Analytics & Reporting
	BasicAnalytics    bool `json:"basic_analytics"`    // Starter+
	AdvancedAnalytics bool `json:"advanced_analytics"` // Pro+
	CustomReports     bool `json:"custom_reports"`     // Business+
	DataExport        bool `json:"data_export"`        // Pro+
	APIAccess         bool `json:"api_access"`         // Business+

	// Marketing & Client Acquisition
	PublicProfile       bool `json:"public_profile"`        // All plans
	OnlineBookingWidget bool `json:"online_booking_widget"` // Starter+
	SEOOptimization     bool `json:"seo_optimization"`      // Pro+
	MarketingAutomation bool `json:"marketing_automation"`  // Business+
	ReviewManagement    bool `json:"review_management"`     // Starter+
	LoyaltyPrograms     bool `json:"loyalty_programs"`      // Pro+

	// Branding & Customization
	CustomBranding bool `json:"custom_branding"` // Pro+
	CustomDomain   bool `json:"custom_domain"`   // Business+
	WhiteLabeling  bool `json:"white_labeling"`  // Enterprise
	MobileApp      bool `json:"mobile_app"`      // All plans

	// Operations
	InventoryManagement bool `json:"inventory_management"` // Pro+
	QuotationManagement bool `json:"quotation_management"` // Starter+
	ContractManagement  bool `json:"contract_management"`  // Business+
	DocumentStorage     bool `json:"document_storage"`     // Starter+
	DigitalSignatures   bool `json:"digital_signatures"`   // Pro+

	// Support & Training
	EmailSupport        bool `json:"email_support"`         // All plans
	PrioritySupport     bool `json:"priority_support"`      // Pro+
	DedicatedAccountMgr bool `json:"dedicated_account_mgr"` // Enterprise
	OnboardingTraining  bool `json:"onboarding_training"`   // Business+
}

// JSONB Scan and Value for SubscriptionFeatures
func (sf *SubscriptionFeatures) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, sf)
}

func (sf SubscriptionFeatures) Value() (driver.Value, error) {
	return json.Marshal(sf)
}

// Business Methods for Subscription
func (s *Subscription) IsActive() bool {
	return s.Status == SubStatusActive || s.Status == SubStatusTrialing
}

func (s *Subscription) IsTrialing() bool {
	return s.Status == SubStatusTrialing &&
		s.TrialEndsAt != nil &&
		time.Now().Before(*s.TrialEndsAt)
}

func (s *Subscription) DaysUntilRenewal() int {
	if s.NextBillingDate == nil {
		return 0
	}
	duration := time.Until(*s.NextBillingDate)
	return int(duration.Hours() / 24)
}

func (s *Subscription) CanAddCustomer() bool {
	return s.CurrentCustomers < s.MaxCustomers
}

func (s *Subscription) CanCreateProject() bool {
	return s.CurrentProjects < s.MaxProjects
}

func (s *Subscription) HasFeature(feature string) bool {
	// Use reflection or switch case to check feature
	// For now, returning true as example
	return true
}

func (s *Subscription) GetPlanLimits() map[string]int {
	return map[string]int{
		"customers":        s.MaxCustomers,
		"projects":         s.MaxProjects,
		"storage_gb":       s.MaxStorageGB,
		"team_members":     s.MaxTeamMembers,
		"services":         s.MaxServicesListed,
		"bookings_monthly": s.MaxBookingsPerMonth,
	}
}

// GetDefaultFeaturesForPlan returns default features for a subscription plan
func GetDefaultFeaturesForPlan(plan SubscriptionPlan) SubscriptionFeatures {
	switch plan {
	case PlanFree:
		return SubscriptionFeatures{
			BasicBooking:       true,
			CustomerManagement: true,
			ServiceCatalog:     true,
			InvoiceGeneration:  true,
			PublicProfile:      true,
			MobileApp:          true,
			EmailSupport:       true,
		}
	case PlanStarter:
		return SubscriptionFeatures{
			BasicBooking:        true,
			CustomerManagement:  true,
			ServiceCatalog:      true,
			InAppMessaging:      true,
			EmailNotifications:  true,
			BasicProjects:       true,
			InvoiceGeneration:   true,
			OnlinePayments:      true,
			OnlineBookingWidget: true,
			ReviewManagement:    true,
			QuotationManagement: true,
			DocumentStorage:     true,
			PublicProfile:       true,
			MobileApp:           true,
			EmailSupport:        true,
		}
	case PlanPro:
		starter := GetDefaultFeaturesForPlan(PlanStarter)
		starter.SMSNotifications = true
		starter.AdvancedProjectMgmt = true
		starter.TeamMembers = true
		starter.RoleBasedAccess = true
		starter.TaskAssignment = true
		starter.RecurringBilling = true
		starter.ExpenseTracking = true
		starter.BasicAnalytics = true
		starter.AdvancedAnalytics = true
		starter.DataExport = true
		starter.SEOOptimization = true
		starter.LoyaltyPrograms = true
		starter.CustomBranding = true
		starter.InventoryManagement = true
		starter.DigitalSignatures = true
		starter.PrioritySupport = true
		return starter
	case PlanBusiness:
		pro := GetDefaultFeaturesForPlan(PlanPro)
		pro.WhatsAppIntegration = true
		pro.ProjectTemplates = true
		pro.GanttCharts = true
		pro.TeamChat = true
		pro.MultiCurrency = true
		pro.ProfitLossReports = true
		pro.CustomReports = true
		pro.APIAccess = true
		pro.MarketingAutomation = true
		pro.CustomDomain = true
		pro.ContractManagement = true
		pro.OnboardingTraining = true
		return pro
	case PlanEnterprise:
		business := GetDefaultFeaturesForPlan(PlanBusiness)
		business.ResourceAllocation = true
		business.WhiteLabeling = true
		business.DedicatedAccountMgr = true
		return business
	default:
		return GetDefaultFeaturesForPlan(PlanFree)
	}
}

// GetDefaultLimitsForPlan returns default limits for a subscription plan
func GetDefaultLimitsForPlan(plan SubscriptionPlan) map[string]int {
	switch plan {
	case PlanFree:
		return map[string]int{
			"max_customers":          6,
			"max_projects":           2,
			"max_storage_gb":         1,
			"max_team_members":       1,
			"max_services_listed":    5,
			"max_bookings_per_month": 20,
		}
	case PlanStarter:
		return map[string]int{
			"max_customers":          25,
			"max_projects":           10,
			"max_storage_gb":         5,
			"max_team_members":       3,
			"max_services_listed":    20,
			"max_bookings_per_month": 100,
		}
	case PlanPro:
		return map[string]int{
			"max_customers":          150,
			"max_projects":           50,
			"max_storage_gb":         25,
			"max_team_members":       10,
			"max_services_listed":    100,
			"max_bookings_per_month": 500,
		}
	case PlanBusiness:
		return map[string]int{
			"max_customers":          1000,
			"max_projects":           200,
			"max_storage_gb":         100,
			"max_team_members":       50,
			"max_services_listed":    500,
			"max_bookings_per_month": 2000,
		}
	case PlanEnterprise:
		return map[string]int{
			"max_customers":          -1, // Unlimited
			"max_projects":           -1, // Unlimited
			"max_storage_gb":         500,
			"max_team_members":       -1, // Unlimited
			"max_services_listed":    -1, // Unlimited
			"max_bookings_per_month": -1, // Unlimited
		}
	default:
		return GetDefaultLimitsForPlan(PlanFree)
	}
}

type UsageRecord struct {
	ID          string `json:"id"`
	CustomerID  string `json:"customer_id"`
	Plan        string `json:"plan"`
	Date        string `json:"date"`
	Customers   int    `json:"customers"`
	Projects    int    `json:"projects"`
	StorageGB   int    `json:"storage_gb"`
	TeamMembers int    `json:"team_members"`
	Services    int    `json:"services"`
	Bookings    int    `json:"bookings"`
}
