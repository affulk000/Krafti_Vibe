package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantPlan string

const (
	TenantPlanSolo        TenantPlan = "solo"
	TenantPlanSmall       TenantPlan = "small"
	TenantPlanCorporation TenantPlan = "corporation"
	TenantPlanEnterprise  TenantPlan = "enterprise"
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusCancelled TenantStatus = "cancelled"
	TenantStatusTrial     TenantStatus = "trial"
)

type Tenant struct {
	BaseModel

	// Owner is REQUIRED
	OwnerID uuid.UUID `json:"owner_id" gorm:"type:uuid;not null;index" validate:"required"`
	Owner   *User     `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`

	// Secondary admins
	Admins []User `json:"admins,omitempty" gorm:"many2many:tenant_admins"`

	// Basic Info
	Name      string `json:"name" gorm:"not null;size:255" validate:"required,min=2,max=255"`
	Subdomain string `json:"subdomain" gorm:"uniqueIndex;not null;size:63" validate:"required,alphanum,min=3,max=63"`
	Domain    string `json:"domain,omitempty" gorm:"size:255" validate:"omitempty,fqdn"`

	// Business Details
	BusinessName  string `json:"business_name" gorm:"size:255"`
	BusinessEmail string `json:"business_email" gorm:"size:255" validate:"omitempty,email"`
	BusinessPhone string `json:"business_phone" gorm:"size:20"`
	TaxID         string `json:"tax_id,omitempty" gorm:"size:50"`

	// Plan & Billing
	Plan              TenantPlan   `json:"plan" gorm:"type:varchar(50);not null;default:'solo'" validate:"required"`
	Status            TenantStatus `json:"status" gorm:"type:varchar(50);not null;default:'trial'" validate:"required"`
	TrialEndsAt       *time.Time   `json:"trial_ends_at,omitempty"`
	SubscriptionID    string       `json:"subscription_id,omitempty" gorm:"size:255"`
	BillingCustomerID string       `json:"billing_customer_id,omitempty" gorm:"size:255"`

	// Settings & Configuration
	Settings     TenantSettings `json:"settings" gorm:"type:jsonb"`
	Features     TenantFeatures `json:"features" gorm:"type:jsonb"`
	Integrations JSONB          `json:"integrations,omitempty" gorm:"type:jsonb"`

	// Branding
	LogoURL      string `json:"logo_url,omitempty" gorm:"size:500"`
	PrimaryColor string `json:"primary_color,omitempty" gorm:"size:7" validate:"omitempty,hexcolor"`

	// Limits & Usage
	MaxUsers     int   `json:"max_users" gorm:"default:10"`
	MaxArtisans  int   `json:"max_artisans" gorm:"default:5"`
	MaxStorage   int64 `json:"max_storage" gorm:"default:1073741824"` // 1GB
	CurrentUsers int   `json:"current_users" gorm:"default:0"`
	StorageUsed  int64 `json:"storage_used" gorm:"default:0"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Users    []User    `json:"users,omitempty" gorm:"foreignKey:TenantID"`
	Bookings []Booking `json:"bookings,omitempty" gorm:"foreignKey:TenantID"`
	Services []Service `json:"services,omitempty" gorm:"foreignKey:TenantID"`
}

// TenantSettings defines operational settings for a tenant
type TenantSettings struct {
	// Booking Settings
	BookingApprovalRequired  bool    `json:"booking_approval_required"`
	AutoAcceptBookings       bool    `json:"auto_accept_bookings"`
	MinBookingLeadTimeHours  int     `json:"min_booking_lead_time_hours" validate:"min=0"` // Default: 24
	MaxAdvanceBookingDays    int     `json:"max_advance_booking_days" validate:"min=1"`    // Default: 90
	AllowRecurringBookings   bool    `json:"allow_recurring_bookings"`
	RequireDepositBooking    bool    `json:"require_deposit_booking"`
	DefaultDepositPercentage float64 `json:"default_deposit_percentage" validate:"min=0,max=100"`

	// Cancellation Policy
	CancellationPolicy      string  `json:"cancellation_policy"` // flexible, moderate, strict
	FullRefundHours         int     `json:"full_refund_hours" validate:"min=0"`
	PartialRefundHours      int     `json:"partial_refund_hours" validate:"min=0"`
	PartialRefundPercentage float64 `json:"partial_refund_percentage" validate:"min=0,max=100"`
	NoShowFeePercentage     float64 `json:"no_show_fee_percentage" validate:"min=0,max=100"`

	// Payment Settings
	DefaultCurrency        string   `json:"default_currency" validate:"len=3"` // USD, EUR, GBP, GHS
	AcceptedPaymentMethods []string `json:"accepted_payment_methods"`          // card, cash, bank_transfer
	AutoChargeOnCompletion bool     `json:"auto_charge_on_completion"`
	EnableTipping          bool     `json:"enable_tipping"`
	DefaultTipPercentages  []int    `json:"default_tip_percentages"` // [10, 15, 20]

	// Commission & Pricing
	PlatformCommissionRate float64 `json:"platform_commission_rate" validate:"min=0,max=100"`
	TaxRate                float64 `json:"tax_rate" validate:"min=0,max=100"`
	IncludeTaxInPrice      bool    `json:"include_tax_in_price"`

	// Notification Settings
	EmailNotificationsEnabled bool  `json:"email_notifications_enabled"`
	SMSNotificationsEnabled   bool  `json:"sms_notifications_enabled"`
	PushNotificationsEnabled  bool  `json:"push_notifications_enabled"`
	NotifyOnNewBooking        bool  `json:"notify_on_new_booking"`
	NotifyOnCancellation      bool  `json:"notify_on_cancellation"`
	NotifyOnPayment           bool  `json:"notify_on_payment"`
	NotifyOnReview            bool  `json:"notify_on_review"`
	ReminderBeforeHours       []int `json:"reminder_before_hours"` // [24, 1]

	// Business Hours
	DefaultTimezone string               `json:"default_timezone"`
	BusinessHours   map[string]TimeRange `json:"business_hours"` // {"monday": {"start": "09:00", "end": "17:00"}}
	Holidays        []Holiday            `json:"holidays,omitempty"`

	// Customer Settings
	AllowCustomerSelfBooking bool `json:"allow_customer_self_booking"`
	RequireCustomerAccount   bool `json:"require_customer_account"`
	CollectCustomerNotes     bool `json:"collect_customer_notes"`
	EnableCustomerReviews    bool `json:"enable_customer_reviews"`
	ReviewsRequireApproval   bool `json:"reviews_require_approval"`

	// Team & Staff
	AllowTeamMemberBooking bool `json:"allow_team_member_booking"`
	RequireTaskAssignment  bool `json:"require_task_assignment"`
	EnableTimeTracking     bool `json:"enable_time_tracking"`

	// Branding & Customization
	ShowPoweredBy          bool   `json:"show_powered_by"`
	CustomEmailFooter      string `json:"custom_email_footer,omitempty"`
	BookingConfirmationMsg string `json:"booking_confirmation_msg,omitempty"`
	CancellationPolicyText string `json:"cancellation_policy_text,omitempty"`

	// Advanced
	EnableWaitlist          bool `json:"enable_waitlist"`
	MaxSimultaneousBookings int  `json:"max_simultaneous_bookings" validate:"min=1"`
	BufferBetweenBookings   int  `json:"buffer_between_bookings" validate:"min=0"` // minutes
	EnableOverbooking       bool `json:"enable_overbooking"`
	OverbookingPercentage   int  `json:"overbooking_percentage" validate:"min=0,max=100"`

	// Privacy & Compliance
	RequireTermsAcceptance bool `json:"require_terms_acceptance"`
	RequirePrivacyConsent  bool `json:"require_privacy_consent"`
	DataRetentionDays      int  `json:"data_retention_days" validate:"min=1"` // Default: 730 (2 years)
	AnonymizeDataAfterDays int  `json:"anonymize_data_after_days" validate:"min=1"`
	GDPRCompliant          bool `json:"gdpr_compliant"`
	AllowDataExport        bool `json:"allow_data_export"`

	// API & Integration
	WebhookURL          string   `json:"webhook_url,omitempty" validate:"omitempty,url"`
	WebhookEvents       []string `json:"webhook_events,omitempty"`
	APIRateLimitPerHour int      `json:"api_rate_limit_per_hour" validate:"min=0"`
	APIRateLimitPerDay  int      `json:"api_rate_limit_per_day" validate:"min=0"`

	// Localization
	DefaultLanguage    string   `json:"default_language"` // en, es, fr
	SupportedLanguages []string `json:"supported_languages"`
	DateFormat         string   `json:"date_format"`                           // MM/DD/YYYY, DD/MM/YYYY
	TimeFormat         string   `json:"time_format"`                           // 12h, 24h
	WeekStartsOn       int      `json:"week_starts_on" validate:"min=0,max=6"` // 0=Sunday, 1=Monday
}

// TenantFeatures defines feature flags for a tenant
type TenantFeatures struct {
	// Core Features (always enabled for all plans)
	BasicBooking       bool `json:"basic_booking"`
	CustomerManagement bool `json:"customer_management"`
	ServiceCatalog     bool `json:"service_catalog"`

	// Booking & Scheduling
	AdvancedBooking      bool `json:"advanced_booking"`
	RecurringBookings    bool `json:"recurring_bookings"`
	GroupBookings        bool `json:"group_bookings"`
	WaitlistManagement   bool `json:"waitlist_management"`
	BookingReminders     bool `json:"booking_reminders"`
	AvailabilityCalendar bool `json:"availability_calendar"`
	ResourceManagement   bool `json:"resource_management"`
	MultiLocationBooking bool `json:"multi_location_booking"`

	// Communication
	InAppMessaging      bool `json:"in_app_messaging"`
	EmailNotifications  bool `json:"email_notifications"`
	SMSNotifications    bool `json:"sms_notifications"`
	WhatsAppIntegration bool `json:"whatsapp_integration"`
	PushNotifications   bool `json:"push_notifications"`
	VideoConferencing   bool `json:"video_conferencing"`
	ChatSupport         bool `json:"chat_support"`

	// Project Management
	BasicProjects       bool `json:"basic_projects"`
	AdvancedProjectMgmt bool `json:"advanced_project_mgmt"`
	ProjectTemplates    bool `json:"project_templates"`
	GanttCharts         bool `json:"gantt_charts"`
	TaskManagement      bool `json:"task_management"`
	TimeTracking        bool `json:"time_tracking"`
	ResourceAllocation  bool `json:"resource_allocation"`
	ProjectMilestones   bool `json:"project_milestones"`
	DependencyTracking  bool `json:"dependency_tracking"`
	ProjectDashboard    bool `json:"project_dashboard"`

	// Team & Collaboration
	TeamMembers     bool `json:"team_members"`
	RoleBasedAccess bool `json:"role_based_access"`
	TeamChat        bool `json:"team_chat"`
	TaskAssignment  bool `json:"task_assignment"`
	TeamCalendar    bool `json:"team_calendar"`
	TeamPerformance bool `json:"team_performance"`
	ShiftManagement bool `json:"shift_management"`

	// Financial & Payments
	OnlinePayments    bool `json:"online_payments"`
	InvoiceGeneration bool `json:"invoice_generation"`
	RecurringBilling  bool `json:"recurring_billing"`
	MultiCurrency     bool `json:"multi_currency"`
	ExpenseTracking   bool `json:"expense_tracking"`
	ProfitLossReports bool `json:"profit_loss_reports"`
	PaymentGateway    bool `json:"payment_gateway"`
	QuoteManagement   bool `json:"quote_management"`
	DepositManagement bool `json:"deposit_management"`
	TipCollection     bool `json:"tip_collection"`
	RefundProcessing  bool `json:"refund_processing"`
	TaxCalculation    bool `json:"tax_calculation"`
	PaymentSchedules  bool `json:"payment_schedules"`

	// Analytics & Reporting
	BasicAnalytics      bool `json:"basic_analytics"`
	AdvancedAnalytics   bool `json:"advanced_analytics"`
	CustomReports       bool `json:"custom_reports"`
	DataExport          bool `json:"data_export"`
	DataVisualization   bool `json:"data_visualization"`
	RevenueReports      bool `json:"revenue_reports"`
	BookingAnalytics    bool `json:"booking_analytics"`
	CustomerInsights    bool `json:"customer_insights"`
	PredictiveAnalytics bool `json:"predictive_analytics"`
	RealTimeDashboard   bool `json:"real_time_dashboard"`

	// Marketing & Client Acquisition
	PublicProfile          bool `json:"public_profile"`
	OnlineBookingWidget    bool `json:"online_booking_widget"`
	SEOOptimization        bool `json:"seo_optimization"`
	MarketingAutomation    bool `json:"marketing_automation"`
	ReviewManagement       bool `json:"review_management"`
	LoyaltyPrograms        bool `json:"loyalty_programs"`
	ReferralProgram        bool `json:"referral_program"`
	EmailCampaigns         bool `json:"email_campaigns"`
	SocialMediaIntegration bool `json:"social_media_integration"`
	PromoCodeManagement    bool `json:"promo_code_management"`
	CustomerSegmentation   bool `json:"customer_segmentation"`

	// Branding & Customization
	CustomBranding       bool `json:"custom_branding"`
	CustomDomain         bool `json:"custom_domain"`
	WhiteLabeling        bool `json:"white_labeling"`
	CustomEmailTemplates bool `json:"custom_email_templates"`
	CustomFields         bool `json:"custom_fields"`
	CustomWorkflows      bool `json:"custom_workflows"`
	ThemeCustomization   bool `json:"theme_customization"`

	// Mobile & Apps
	MobileApp      bool `json:"mobile_app"`
	MobileBooking  bool `json:"mobile_booking"`
	OfflineMode    bool `json:"offline_mode"`
	GPSTracking    bool `json:"gps_tracking"`
	MobilePayments bool `json:"mobile_payments"`

	// Operations
	InventoryManagement bool `json:"inventory_management"`
	QuotationManagement bool `json:"quotation_management"`
	ContractManagement  bool `json:"contract_management"`
	DocumentStorage     bool `json:"document_storage"`
	DigitalSignatures   bool `json:"digital_signatures"`
	SupplierManagement  bool `json:"supplier_management"`
	PurchaseOrders      bool `json:"purchase_orders"`
	QualityControl      bool `json:"quality_control"`
	SafetyCompliance    bool `json:"safety_compliance"`

	// Customer Experience
	CustomerPortal      bool `json:"customer_portal"`
	SelfServiceBooking  bool `json:"self_service_booking"`
	BookingHistory      bool `json:"booking_history"`
	SavedPaymentMethods bool `json:"saved_payment_methods"`
	FavoriteArtisans    bool `json:"favorite_artisans"`
	RatingAndReviews    bool `json:"rating_and_reviews"`
	ServiceReminders    bool `json:"service_reminders"`
	CustomerFeedback    bool `json:"customer_feedback"`

	// Support & Training
	EmailSupport        bool `json:"email_support"`
	PhoneSupport        bool `json:"phone_support"`
	PrioritySupport     bool `json:"priority_support"`
	DedicatedAccountMgr bool `json:"dedicated_account_mgr"`
	OnboardingTraining  bool `json:"onboarding_training"`
	VideoTutorials      bool `json:"video_tutorials"`
	KnowledgeBase       bool `json:"knowledge_base"`
	LiveChat            bool `json:"live_chat"`

	// Integration & API
	APIAccess              bool `json:"api_access"`
	WebhookSupport         bool `json:"webhook_support"`
	ThirdPartyIntegrations bool `json:"third_party_integrations"`
	ZapierIntegration      bool `json:"zapier_integration"`
	CalendarSync           bool `json:"calendar_sync"`          // Google, Outlook
	AccountingIntegration  bool `json:"accounting_integration"` // QuickBooks, Xero
	CRMIntegration         bool `json:"crm_integration"`
	PaymentGatewayOptions  bool `json:"payment_gateway_options"`

	// Security & Compliance
	TwoFactorAuth            bool `json:"two_factor_auth"`
	SSO                      bool `json:"sso"` // Single Sign-On
	DataEncryption           bool `json:"data_encryption"`
	AuditLogs                bool `json:"audit_logs"`
	IPWhitelisting           bool `json:"ip_whitelisting"`
	BackupAndRecovery        bool `json:"backup_and_recovery"`
	ComplianceCertifications bool `json:"compliance_certifications"`
	DataResidency            bool `json:"data_residency"`

	// Advanced Features
	AIRecommendations       bool `json:"ai_recommendations"`
	AutomatedWorkflows      bool `json:"automated_workflows"`
	MultiTenantSupport      bool `json:"multi_tenant_support"`
	CustomIntegrations      bool `json:"custom_integrations"`
	AdvancedSecurity        bool `json:"advanced_security"`
	DisasterRecovery        bool `json:"disaster_recovery"`
	SLAGuarantees           bool `json:"sla_guarantees"`
	DedicatedInfrastructure bool `json:"dedicated_infrastructure"`
}

// TimeRange represents a time range in HH:MM format
type TimeRange struct {
	Start string `json:"start" validate:"required"` // "09:00"
	End   string `json:"end" validate:"required"`   // "17:00"
}

// Holiday represents a business holiday
type Holiday struct {
	Name string `json:"name" validate:"required"`
	Date string `json:"date" validate:"required"` // YYYY-MM-DD
	Type string `json:"type"`                     // public, company, religious
}

// Scan implements the sql.Scanner interface for TenantSettings
func (ts *TenantSettings) Scan(value any) error {
	if value == nil {
		*ts = GetDefaultTenantSettings()
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, ts)
}

// Value implements the driver.Valuer interface for TenantSettings
func (ts TenantSettings) Value() (driver.Value, error) {
	return json.Marshal(ts)
}

// Scan implements the sql.Scanner interface for TenantFeatures
func (tf *TenantFeatures) Scan(value any) error {
	if value == nil {
		*tf = TenantFeatures{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, tf)
}

// Value implements the driver.Valuer interface for TenantFeatures
func (tf TenantFeatures) Value() (driver.Value, error) {
	return json.Marshal(tf)
}

// GetDefaultTenantSettings returns sensible defaults for a new tenant
func GetDefaultTenantSettings() TenantSettings {
	return TenantSettings{
		// Booking defaults
		BookingApprovalRequired:  false,
		AutoAcceptBookings:       true,
		MinBookingLeadTimeHours:  24,
		MaxAdvanceBookingDays:    90,
		AllowRecurringBookings:   true,
		RequireDepositBooking:    false,
		DefaultDepositPercentage: 20.0,

		// Cancellation policy
		CancellationPolicy:      "moderate",
		FullRefundHours:         24,
		PartialRefundHours:      12,
		PartialRefundPercentage: 50.0,
		NoShowFeePercentage:     100.0,

		// Payment defaults
		DefaultCurrency:        "USD",
		AcceptedPaymentMethods: []string{"card", "cash"},
		AutoChargeOnCompletion: false,
		EnableTipping:          true,
		DefaultTipPercentages:  []int{10, 15, 20},

		// Commission & pricing
		PlatformCommissionRate: 10.0,
		TaxRate:                0.0,
		IncludeTaxInPrice:      false,

		// Notifications
		EmailNotificationsEnabled: true,
		SMSNotificationsEnabled:   false,
		PushNotificationsEnabled:  true,
		NotifyOnNewBooking:        true,
		NotifyOnCancellation:      true,
		NotifyOnPayment:           true,
		NotifyOnReview:            true,
		ReminderBeforeHours:       []int{24, 1},

		// Business hours
		DefaultTimezone: "UTC",
		BusinessHours: map[string]TimeRange{
			"monday":    {Start: "09:00", End: "17:00"},
			"tuesday":   {Start: "09:00", End: "17:00"},
			"wednesday": {Start: "09:00", End: "17:00"},
			"thursday":  {Start: "09:00", End: "17:00"},
			"friday":    {Start: "09:00", End: "17:00"},
			"saturday":  {Start: "10:00", End: "14:00"},
			"sunday":    {Start: "closed", End: "closed"},
		},

		// Customer settings
		AllowCustomerSelfBooking: true,
		RequireCustomerAccount:   false,
		CollectCustomerNotes:     true,
		EnableCustomerReviews:    true,
		ReviewsRequireApproval:   false,

		// Team settings
		AllowTeamMemberBooking: true,
		RequireTaskAssignment:  false,
		EnableTimeTracking:     false,

		// Advanced
		EnableWaitlist:          false,
		MaxSimultaneousBookings: 1,
		BufferBetweenBookings:   0,
		EnableOverbooking:       false,
		OverbookingPercentage:   0,

		// Privacy
		RequireTermsAcceptance: true,
		RequirePrivacyConsent:  true,
		DataRetentionDays:      730,
		AnonymizeDataAfterDays: 1095,
		GDPRCompliant:          true,
		AllowDataExport:        true,

		// API
		APIRateLimitPerHour: 1000,
		APIRateLimitPerDay:  10000,

		// Localization
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en"},
		DateFormat:         "MM/DD/YYYY",
		TimeFormat:         "12h",
		WeekStartsOn:       0, // Sunday
	}
}

// Helper methods for TenantSettings
func (ts *TenantSettings) IsBusinessHoursOn(dayOfWeek string) bool {
	if hours, ok := ts.BusinessHours[dayOfWeek]; ok {
		return hours.Start != "closed"
	}
	return false
}

func (ts *TenantSettings) GetCancellationRefundPercentage(hoursUntilBooking float64) float64 {
	if hoursUntilBooking >= float64(ts.FullRefundHours) {
		return 100.0
	}
	if hoursUntilBooking >= float64(ts.PartialRefundHours) {
		return ts.PartialRefundPercentage
	}
	return 0.0
}

func (ts *TenantSettings) SupportsPaymentMethod(method string) bool {
	return slices.Contains(ts.AcceptedPaymentMethods, method)
}

// Helper methods for TenantFeatures
func (tf *TenantFeatures) HasFeature(feature string) bool {
	// Use reflection or switch case to check individual features
	// This is a simplified version
	switch feature {
	case "advanced_project_mgmt":
		return tf.AdvancedProjectMgmt
	case "team_members":
		return tf.TeamMembers
	case "api_access":
		return tf.APIAccess
	case "custom_branding":
		return tf.CustomBranding
	case "white_labeling":
		return tf.WhiteLabeling
	default:
		return false
	}
}

func (tf *TenantFeatures) GetEnabledFeaturesCount() int {
	count := 0
	// Count all true boolean fields
	// This would be better with reflection in production
	if tf.BasicBooking {
		count++
	}
	if tf.AdvancedProjectMgmt {
		count++
	}
	if tf.TeamMembers {
		count++
	}
	// ... add all other features
	return count
}

// Business Methods
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

func (t *Tenant) IsTrial() bool {
	return t.Status == TenantStatusTrial
}

func (t *Tenant) TrialExpired() bool {
	return t.TrialEndsAt != nil && time.Now().After(*t.TrialEndsAt)
}

func (t *Tenant) CanAddUser() bool {
	return t.CurrentUsers < t.MaxUsers
}

// State Machine
func (t *Tenant) CanTransitionTo(newStatus TenantStatus) error {
	validTransitions := map[TenantStatus][]TenantStatus{
		TenantStatusTrial: {
			TenantStatusActive,
			TenantStatusCancelled,
		},
		TenantStatusActive: {
			TenantStatusSuspended,
			TenantStatusCancelled,
		},
		TenantStatusSuspended: {
			TenantStatusActive,
			TenantStatusCancelled,
		},
		TenantStatusCancelled: {}, // Terminal state
	}

	allowed := validTransitions[t.Status]
	if valid := slices.Contains(allowed, newStatus); !valid {
		return fmt.Errorf("cannot transition tenant from %s to %s", t.Status, newStatus)
	}
	return nil
}

func (t *Tenant) Suspend(reason string) error {
	if err := t.CanTransitionTo(TenantStatusSuspended); err != nil {
		return err
	}
	t.Status = TenantStatusSuspended
	// TODO: Log to audit trail
	return nil
}

func (t *Tenant) Activate() error {
	if err := t.CanTransitionTo(TenantStatusActive); err != nil {
		return err
	}
	t.Status = TenantStatusActive
	return nil
}

// Validation Hooks
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	if t.OwnerID == uuid.Nil {
		return errors.New("tenant must have an owner")
	}

	// Verify owner exists and has correct role
	var owner User
	if err := tx.First(&owner, t.OwnerID).Error; err != nil {
		return errors.New("tenant owner not found")
	}

	if !owner.IsPlatformAdmin() && owner.Role != UserRoleTenantOwner {
		return errors.New("user must have tenant_owner role or be platform admin")
	}

	return nil
}
