package logto

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

// Config holds Logto configuration
type Config struct {
	// Logto tenant configuration
	Endpoint      string
	AppID         string
	AppSecret     string
	JWKSEndpoint  string
	TokenEndpoint string
	Issuer        string

	// API Resource configuration
	APIResourceIndicator string

	// Organization/Multi-tenancy settings
	EnableOrganizations bool

	// JWKS cache settings
	JWKSCacheTTL      time.Duration
	JWKSRefreshWindow time.Duration

	// Token validation settings
	ValidateAudience   bool
	ValidateIssuer     bool
	ClockSkewTolerance time.Duration

	// Feature flags
	EnableM2M     bool
	EnableRBAC    bool
	EnableLogging bool
}

// DefaultConfig returns default Logto configuration
func DefaultConfig() *Config {
	return &Config{
		JWKSCacheTTL:        15 * time.Minute,
		JWKSRefreshWindow:   1 * time.Minute,
		ValidateAudience:    true,
		ValidateIssuer:      true,
		ClockSkewTolerance:  5 * time.Minute,
		EnableOrganizations: true,
		EnableM2M:           true,
		EnableRBAC:          true,
		EnableLogging:       true,
	}
}

// LoadConfig loads configuration from environment or config file
func LoadConfig(endpoint, appID, appSecret string) (*Config, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("logto endpoint is required")
	}

	config := DefaultConfig()
	config.Endpoint = endpoint
	config.AppID = appID
	config.AppSecret = appSecret

	// Construct standard endpoints
	config.JWKSEndpoint = fmt.Sprintf("%s/oidc/jwks", endpoint)
	config.TokenEndpoint = fmt.Sprintf("%s/oidc/token", endpoint)
	config.Issuer = fmt.Sprintf("%s/oidc", endpoint)

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if c.JWKSEndpoint == "" {
		return fmt.Errorf("JWKS endpoint is required")
	}
	if c.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}
	if c.JWKSCacheTTL < time.Minute {
		return fmt.Errorf("JWKS cache TTL must be at least 1 minute")
	}
	return nil
}

// LogConfig logs the configuration (excluding sensitive data)
func (c *Config) LogConfig() {
	if !c.EnableLogging {
		return
	}

	log.Info("Logto Configuration:")
	log.Infof("  Endpoint: %s", c.Endpoint)
	log.Infof("  JWKS Endpoint: %s", c.JWKSEndpoint)
	log.Infof("  Issuer: %s", c.Issuer)
	log.Infof("  API Resource: %s", c.APIResourceIndicator)
	log.Infof("  Organizations Enabled: %v", c.EnableOrganizations)
	log.Infof("  M2M Enabled: %v", c.EnableM2M)
	log.Infof("  RBAC Enabled: %v", c.EnableRBAC)
}

// Scopes defines the application's permission scopes
type Scopes struct {
	// Service Management
	ServiceRead   string
	ServiceWrite  string
	ServiceDelete string

	// Booking Management
	BookingRead   string
	BookingWrite  string
	BookingDelete string
	BookingManage string

	// Project Management
	ProjectRead   string
	ProjectWrite  string
	ProjectDelete string
	ProjectManage string

	// User Management
	UserRead   string
	UserWrite  string
	UserDelete string
	UserManage string

	// Artisan Management
	ArtisanRead   string
	ArtisanWrite  string
	ArtisanManage string

	// Customer Management
	CustomerRead   string
	CustomerWrite  string
	CustomerManage string

	// Payment Management
	PaymentRead    string
	PaymentWrite   string
	PaymentProcess string

	// Invoice Management
	InvoiceRead  string
	InvoiceWrite string

	// Subscription Management
	SubscriptionRead  string
	SubscriptionWrite string

	// Message Management
	MessageRead  string
	MessageWrite string

	// Notification Management
	NotificationRead  string
	NotificationWrite string

	// Milestone Management
	MilestoneRead  string
	MilestoneWrite string

	// Task Management
	TaskRead  string
	TaskWrite string

	// Tenant/Organization Management
	TenantRead   string
	TenantWrite  string
	TenantManage string
	TenantAdmin  string

	// Review Management
	ReviewRead     string
	ReviewWrite    string
	ReviewModerate string

	// Report Management
	ReportRead     string
	ReportGenerate string
	ReportExport   string

	// Webhook Management
	WebhookRead   string
	WebhookWrite  string
	WebhookManage string

	// File Upload Management
	FileRead   string
	FileWrite  string
	FileDelete string
	FileManage string

	// Promo Code Management
	PromoRead   string
	PromoWrite  string
	PromoDelete string
	PromoApply  string

	// System Settings Management
	SettingsRead   string
	SettingsWrite  string
	SettingsDelete string
	SettingsManage string

	// Admin Scopes
	AdminRead  string
	AdminWrite string
	AdminFull  string
}

// DefaultScopes returns the default scope definitions
func DefaultScopes() *Scopes {
	return &Scopes{
		// Service Management
		ServiceRead:   "service:read",
		ServiceWrite:  "service:write",
		ServiceDelete: "service:delete",

		// Booking Management
		BookingRead:   "booking:read",
		BookingWrite:  "booking:write",
		BookingDelete: "booking:delete",
		BookingManage: "booking:manage",

		// Project Management
		ProjectRead:   "project:read",
		ProjectWrite:  "project:write",
		ProjectDelete: "project:delete",
		ProjectManage: "project:manage",

		// User Management
		UserRead:   "user:read",
		UserWrite:  "user:write",
		UserDelete: "user:delete",
		UserManage: "user:manage",

		// Artisan Management
		ArtisanRead:   "artisan:read",
		ArtisanWrite:  "artisan:write",
		ArtisanManage: "artisan:manage",

		// Customer Management
		CustomerRead:   "customer:read",
		CustomerWrite:  "customer:write",
		CustomerManage: "customer:manage",

		// Payment Management
		PaymentRead:    "payment:read",
		PaymentWrite:   "payment:write",
		PaymentProcess: "payment:process",

		// Invoice Management
		InvoiceRead:  "invoice:read",
		InvoiceWrite: "invoice:write",

		// Subscription Management
		SubscriptionRead:  "subscription:read",
		SubscriptionWrite: "subscription:write",

		// Message Management
		MessageRead:  "message:read",
		MessageWrite: "message:write",

		// Notification Management
		NotificationRead:  "notification:read",
		NotificationWrite: "notification:write",

		// Milestone Management
		MilestoneRead:  "milestone:read",
		MilestoneWrite: "milestone:write",

		// Task Management
		TaskRead:  "task:read",
		TaskWrite: "task:write",

		// Tenant/Organization Management
		TenantRead:   "tenant:read",
		TenantWrite:  "tenant:write",
		TenantManage: "tenant:manage",
		TenantAdmin:  "tenant:admin",

		// Review Management
		ReviewRead:     "review:read",
		ReviewWrite:    "review:write",
		ReviewModerate: "review:moderate",

		// Report Management
		ReportRead:     "report:read",
		ReportGenerate: "report:generate",
		ReportExport:   "report:export",

		// Webhook Management
		WebhookRead:   "webhook:read",
		WebhookWrite:  "webhook:write",
		WebhookManage: "webhook:manage",

		// File Upload Management
		FileRead:   "file:read",
		FileWrite:  "file:write",
		FileDelete: "file:delete",
		FileManage: "file:manage",

		// Promo Code Management
		PromoRead:   "promo:read",
		PromoWrite:  "promo:write",
		PromoDelete: "promo:delete",
		PromoApply:  "promo:apply",

		// System Settings Management
		SettingsRead:   "settings:read",
		SettingsWrite:  "settings:write",
		SettingsDelete: "settings:delete",
		SettingsManage: "settings:manage",

		// Admin Scopes
		AdminRead:  "admin:read",
		AdminWrite: "admin:write",
		AdminFull:  "admin:full",
	}
}

// RoleDefinition defines a role with its scopes
type RoleDefinition struct {
	Name        string
	Description string
	Scopes      []string
}

// DefaultRoles returns predefined role definitions
// These roles align with UserRole constants in internal/domain/models/user.go
func DefaultRoles() []RoleDefinition {
	scopes := DefaultScopes()

	return []RoleDefinition{
		// ============================================================================
		// Platform Roles (IsPlatformUser = true, TenantID = NULL)
		// ============================================================================
		{
			Name:        "platform_super_admin",
			Description: "Platform super administrator - App owner with unrestricted access",
			Scopes: []string{
				scopes.AdminFull,
				scopes.TenantAdmin,
				scopes.UserManage,
				scopes.SettingsManage,
				scopes.WebhookManage,
				scopes.FileManage,
				scopes.ReportExport,
			},
		},
		{
			Name:        "platform_admin",
			Description: "Platform administrator - Platform staff with full management access",
			Scopes: []string{
				scopes.AdminWrite,
				scopes.TenantManage,
				scopes.UserManage,
				scopes.SettingsWrite,
				scopes.ReportGenerate,
			},
		},
		{
			Name:        "platform_support",
			Description: "Platform support staff - Read access for customer support",
			Scopes: []string{
				scopes.AdminRead,
				scopes.TenantRead,
				scopes.UserRead,
				scopes.CustomerRead,
				scopes.BookingRead,
				scopes.ProjectRead,
				scopes.ServiceRead,
				scopes.ReviewRead,
				scopes.PaymentRead,
			},
		},

		// ============================================================================
		// Tenant Roles (IsPlatformUser = false, TenantID = NOT NULL)
		// ============================================================================
		{
			Name:        "tenant_owner",
			Description: "Tenant owner - Owns the organization with full control",
			Scopes: []string{
				scopes.TenantAdmin,
				scopes.TenantManage,
				scopes.UserManage,
				scopes.ArtisanManage,
				scopes.CustomerManage,
				scopes.ServiceWrite,
				scopes.BookingManage,
				scopes.ProjectManage,
				scopes.PaymentProcess,
				scopes.InvoiceWrite,
				scopes.SubscriptionWrite,
				scopes.ReviewModerate,
				scopes.ReportGenerate,
				scopes.WebhookManage,
				scopes.FileManage,
				scopes.SettingsManage,
				scopes.PromoWrite,
			},
		},
		{
			Name:        "tenant_admin",
			Description: "Tenant administrator - Admin of organization with management access",
			Scopes: []string{
				scopes.TenantManage,
				scopes.UserWrite,
				scopes.ArtisanManage,
				scopes.CustomerManage,
				scopes.ServiceWrite,
				scopes.BookingManage,
				scopes.ProjectManage,
				scopes.PaymentRead,
				scopes.InvoiceWrite,
				scopes.ReviewModerate,
				scopes.ReportGenerate,
				scopes.FileWrite,
				scopes.SettingsWrite,
			},
		},
		{
			Name:        "artisan",
			Description: "Artisan - Service provider with service and project management",
			Scopes: []string{
				scopes.ServiceRead,
				scopes.ServiceWrite,
				scopes.BookingRead,
				scopes.BookingWrite,
				scopes.ProjectRead,
				scopes.ProjectWrite,
				scopes.CustomerRead,
				scopes.PaymentRead,
				scopes.InvoiceRead,
				scopes.MilestoneWrite,
				scopes.TaskWrite,
				scopes.ReviewRead,
				scopes.MessageWrite,
				scopes.NotificationRead,
				scopes.FileWrite,
			},
		},
		{
			Name:        "team_member",
			Description: "Team member - Employee/staff with operational access",
			Scopes: []string{
				scopes.ServiceRead,
				scopes.BookingRead,
				scopes.BookingWrite,
				scopes.ProjectRead,
				scopes.ProjectWrite,
				scopes.CustomerRead,
				scopes.TaskWrite,
				scopes.MilestoneRead,
				scopes.MessageWrite,
				scopes.NotificationRead,
				scopes.FileRead,
			},
		},
		{
			Name:        "customer",
			Description: "Customer - Service consumer with booking and review capabilities",
			Scopes: []string{
				scopes.ServiceRead,
				scopes.ArtisanRead,
				scopes.BookingRead,
				scopes.BookingWrite,
				scopes.ProjectRead,
				scopes.PaymentRead,
				scopes.PaymentWrite,
				scopes.InvoiceRead,
				scopes.ReviewRead,
				scopes.ReviewWrite,
				scopes.MessageWrite,
				scopes.NotificationRead,
				scopes.FileRead,
				scopes.PromoApply,
			},
		},

		// ============================================================================
		// Additional Functional Roles (Can be assigned to team_member users)
		// ============================================================================
		{
			Name:        "project_manager",
			Description: "Project manager - Team member with enhanced project oversight",
			Scopes: []string{
				scopes.ProjectManage,
				scopes.BookingRead,
				scopes.BookingWrite,
				scopes.ArtisanRead,
				scopes.CustomerRead,
				scopes.MilestoneWrite,
				scopes.TaskWrite,
				scopes.ReportRead,
				scopes.FileWrite,
			},
		},
		{
			Name:        "accountant",
			Description: "Accountant - Team member with financial oversight",
			Scopes: []string{
				scopes.PaymentRead,
				scopes.PaymentProcess,
				scopes.InvoiceRead,
				scopes.InvoiceWrite,
				scopes.BookingRead,
				scopes.ProjectRead,
				scopes.SubscriptionRead,
				scopes.ReportGenerate,
				scopes.ReportExport,
			},
		},

		// ============================================================================
		// Service Roles
		// ============================================================================
		{
			Name:        "m2m_service",
			Description: "Machine-to-machine service integration",
			Scopes: []string{
				scopes.ServiceRead,
				scopes.BookingRead,
				scopes.BookingWrite,
				scopes.CustomerRead,
				scopes.PaymentRead,
				scopes.WebhookRead,
				scopes.NotificationWrite,
			},
		},
	}
}

// APIResource defines an API resource with its permissions
type APIResource struct {
	Indicator   string
	Name        string
	Description string
	Permissions []Permission
}

// Permission defines a specific permission
type Permission struct {
	Name        string
	Description string
}

// DefaultAPIResources returns the default API resource definitions
func DefaultAPIResources() []APIResource {
	return []APIResource{
		{
			Indicator:   "https://api.kraftivibe.com",
			Name:        "Krafti Vibe API",
			Description: "Main API for Krafti Vibe platform",
			Permissions: []Permission{
				// Service Management
				{Name: "service:read", Description: "Read services"},
				{Name: "service:write", Description: "Create and update services"},
				{Name: "service:delete", Description: "Delete services"},

				// Booking Management
				{Name: "booking:read", Description: "Read bookings"},
				{Name: "booking:write", Description: "Create and update bookings"},
				{Name: "booking:delete", Description: "Delete bookings"},
				{Name: "booking:manage", Description: "Full booking management"},

				// Project Management
				{Name: "project:read", Description: "Read projects"},
				{Name: "project:write", Description: "Create and update projects"},
				{Name: "project:delete", Description: "Delete projects"},
				{Name: "project:manage", Description: "Full project management"},

				// User Management
				{Name: "user:read", Description: "Read users"},
				{Name: "user:write", Description: "Create and update users"},
				{Name: "user:delete", Description: "Delete users"},
				{Name: "user:manage", Description: "Full user management"},

				// Artisan Management
				{Name: "artisan:read", Description: "Read artisan profiles"},
				{Name: "artisan:write", Description: "Create and update artisan profiles"},
				{Name: "artisan:manage", Description: "Full artisan management"},

				// Customer Management
				{Name: "customer:read", Description: "Read customer profiles"},
				{Name: "customer:write", Description: "Create and update customer profiles"},
				{Name: "customer:manage", Description: "Full customer management"},

				// Payment Management
				{Name: "payment:read", Description: "Read payment information"},
				{Name: "payment:write", Description: "Create and update payments"},
				{Name: "payment:process", Description: "Process payments"},

				// Invoice Management
				{Name: "invoice:read", Description: "Read invoices"},
				{Name: "invoice:write", Description: "Create and update invoices"},

				// Subscription Management
				{Name: "subscription:read", Description: "Read subscriptions"},
				{Name: "subscription:write", Description: "Create and update subscriptions"},

				// Message Management
				{Name: "message:read", Description: "Read messages"},
				{Name: "message:write", Description: "Send messages"},

				// Notification Management
				{Name: "notification:read", Description: "Read notifications"},
				{Name: "notification:write", Description: "Create and update notifications"},

				// Milestone Management
				{Name: "milestone:read", Description: "Read milestones"},
				{Name: "milestone:write", Description: "Create and update milestones"},

				// Task Management
				{Name: "task:read", Description: "Read tasks"},
				{Name: "task:write", Description: "Create and update tasks"},

				// Tenant/Organization Management
				{Name: "tenant:read", Description: "Read tenant information"},
				{Name: "tenant:write", Description: "Create and update tenant"},
				{Name: "tenant:manage", Description: "Manage tenant settings and members"},
				{Name: "tenant:admin", Description: "Full tenant administration"},

				// Review Management
				{Name: "review:read", Description: "Read reviews"},
				{Name: "review:write", Description: "Create and update reviews"},
				{Name: "review:moderate", Description: "Moderate reviews"},

				// Report Management
				{Name: "report:read", Description: "Read reports"},
				{Name: "report:generate", Description: "Generate reports"},
				{Name: "report:export", Description: "Export reports"},

				// Webhook Management
				{Name: "webhook:read", Description: "Read webhooks"},
				{Name: "webhook:write", Description: "Create and update webhooks"},
				{Name: "webhook:manage", Description: "Full webhook management"},

				// File Upload Management
				{Name: "file:read", Description: "Read files"},
				{Name: "file:write", Description: "Upload and update files"},
				{Name: "file:delete", Description: "Delete files"},
				{Name: "file:manage", Description: "Full file management"},

				// Promo Code Management
				{Name: "promo:read", Description: "Read promotional codes"},
				{Name: "promo:write", Description: "Create and update promotional codes"},
				{Name: "promo:delete", Description: "Delete promotional codes"},
				{Name: "promo:apply", Description: "Apply promotional codes"},

				// System Settings Management
				{Name: "settings:read", Description: "Read system settings"},
				{Name: "settings:write", Description: "Create and update system settings"},
				{Name: "settings:delete", Description: "Delete system settings"},
				{Name: "settings:manage", Description: "Full settings management"},

				// Admin Scopes
				{Name: "admin:read", Description: "Administrative read access"},
				{Name: "admin:write", Description: "Administrative write access"},
				{Name: "admin:full", Description: "Full administrative access"},
			},
		},
		{
			Indicator:   "https://api.kraftivibe.com/org",
			Name:        "Krafti Vibe Organization API",
			Description: "Organization-scoped API resources",
			Permissions: []Permission{
				{Name: "org:read", Description: "Read organization data"},
				{Name: "org:write", Description: "Modify organization data"},
				{Name: "org:member:invite", Description: "Invite organization members"},
				{Name: "org:member:manage", Description: "Manage organization members"},
				{Name: "org:billing:read", Description: "Read billing information"},
				{Name: "org:billing:manage", Description: "Manage billing"},
				{Name: "org:settings:read", Description: "Read organization settings"},
				{Name: "org:settings:write", Description: "Modify organization settings"},
			},
		},
	}
}
