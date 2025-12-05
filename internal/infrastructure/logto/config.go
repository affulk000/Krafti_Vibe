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
func DefaultRoles() []RoleDefinition {
	scopes := DefaultScopes()

	return []RoleDefinition{
		{
			Name:        "platform_admin",
			Description: "Platform-wide administrator with full access",
			Scopes: []string{
				scopes.AdminFull,
				scopes.TenantAdmin,
				scopes.UserManage,
			},
		},
		{
			Name:        "tenant_admin",
			Description: "Tenant administrator with full tenant access",
			Scopes: []string{
				scopes.TenantManage,
				scopes.UserManage,
				scopes.ArtisanManage,
				scopes.CustomerManage,
				scopes.ServiceWrite,
				scopes.BookingManage,
				scopes.ProjectManage,
				scopes.PaymentRead,
				scopes.ReviewModerate,
				scopes.ReportGenerate,
			},
		},
		{
			Name:        "artisan",
			Description: "Artisan with service and project management",
			Scopes: []string{
				scopes.ServiceRead,
				scopes.ServiceWrite,
				scopes.BookingRead,
				scopes.BookingWrite,
				scopes.ProjectRead,
				scopes.ProjectWrite,
				scopes.CustomerRead,
				scopes.PaymentRead,
			},
		},
		{
			Name:        "customer",
			Description: "Customer with booking and review capabilities",
			Scopes: []string{
				scopes.ServiceRead,
				scopes.BookingRead,
				scopes.BookingWrite,
				scopes.ProjectRead,
				scopes.PaymentRead,
				scopes.PaymentWrite,
				scopes.ReviewRead,
				scopes.ReviewWrite,
			},
		},
		{
			Name:        "project_manager",
			Description: "Project manager with project oversight",
			Scopes: []string{
				scopes.ProjectManage,
				scopes.BookingRead,
				scopes.ArtisanRead,
				scopes.CustomerRead,
				scopes.ReportRead,
			},
		},
		{
			Name:        "accountant",
			Description: "Financial oversight and payment management",
			Scopes: []string{
				scopes.PaymentRead,
				scopes.PaymentProcess,
				scopes.BookingRead,
				scopes.ProjectRead,
				scopes.ReportGenerate,
				scopes.ReportExport,
			},
		},
		{
			Name:        "support_staff",
			Description: "Customer support with read access",
			Scopes: []string{
				scopes.UserRead,
				scopes.CustomerRead,
				scopes.BookingRead,
				scopes.ProjectRead,
				scopes.ServiceRead,
				scopes.ReviewRead,
			},
		},
		{
			Name:        "m2m_service",
			Description: "Machine-to-machine service integration",
			Scopes: []string{
				scopes.ServiceRead,
				scopes.BookingRead,
				scopes.BookingWrite,
				scopes.CustomerRead,
				scopes.PaymentRead,
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
				{Name: "service:read", Description: "Read services"},
				{Name: "service:write", Description: "Create and update services"},
				{Name: "service:delete", Description: "Delete services"},
				{Name: "booking:read", Description: "Read bookings"},
				{Name: "booking:write", Description: "Create and update bookings"},
				{Name: "booking:manage", Description: "Full booking management"},
				{Name: "project:read", Description: "Read projects"},
				{Name: "project:write", Description: "Create and update projects"},
				{Name: "project:manage", Description: "Full project management"},
				{Name: "user:manage", Description: "Manage users"},
				{Name: "payment:process", Description: "Process payments"},
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
