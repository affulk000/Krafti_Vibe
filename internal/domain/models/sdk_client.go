package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// SDKPlatform represents the platform type for SDK clients
type SDKPlatform string

const (
	SDKPlatformIOS         SDKPlatform = "ios"
	SDKPlatformAndroid     SDKPlatform = "android"
	SDKPlatformWeb         SDKPlatform = "web"
	SDKPlatformReactNative SDKPlatform = "react_native"
	SDKPlatformFlutter     SDKPlatform = "flutter"
	SDKPlatformNode        SDKPlatform = "node"
	SDKPlatformPython      SDKPlatform = "python"
	SDKPlatformGo          SDKPlatform = "go"
	SDKPlatformJava        SDKPlatform = "java"
	SDKPlatformPHP         SDKPlatform = "php"
	SDKPlatformRuby        SDKPlatform = "ruby"
	SDKPlatformDotNet      SDKPlatform = "dotnet"
)

// SDKEnvironment represents the environment type
type SDKEnvironment string

const (
	SDKEnvironmentProduction  SDKEnvironment = "production"
	SDKEnvironmentStaging     SDKEnvironment = "staging"
	SDKEnvironmentDevelopment SDKEnvironment = "development"
	SDKEnvironmentTesting     SDKEnvironment = "testing"
)

// SDKKeyStatus represents the status of an SDK key
type SDKKeyStatus string

const (
	SDKKeyStatusActive    SDKKeyStatus = "active"
	SDKKeyStatusInactive  SDKKeyStatus = "inactive"
	SDKKeyStatusRevoked   SDKKeyStatus = "revoked"
	SDKKeyStatusExpired   SDKKeyStatus = "expired"
	SDKKeyStatusSuspended SDKKeyStatus = "suspended"
)

// SDKClient represents an SDK client application
type SDKClient struct {
	BaseModel

	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index:idx_sdk_clients_tenant"`
	Tenant   *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`

	// Basic Information
	Name        string         `json:"name" gorm:"not null;size:255"`
	Description string         `json:"description,omitempty" gorm:"type:text"`
	Platform    SDKPlatform    `json:"platform" gorm:"type:varchar(50);not null;index:idx_sdk_clients_platform"`
	Environment SDKEnvironment `json:"environment" gorm:"type:varchar(50);not null;default:'production'"`

	// Application Details
	AppIdentifier string `json:"app_identifier,omitempty" gorm:"size:255"` // Bundle ID, Package Name, etc.
	AppVersion    string `json:"app_version,omitempty" gorm:"size:50"`
	SDKVersion    string `json:"sdk_version,omitempty" gorm:"size:50"`

	// Configuration
	AllowedOrigins []string  `json:"allowed_origins,omitempty" gorm:"type:text[]"`
	AllowedIPs     []string  `json:"allowed_ips,omitempty" gorm:"type:text[]"`
	WebhookURL     string    `json:"webhook_url,omitempty" gorm:"size:500"`
	WebhookSecret  string    `json:"webhook_secret,omitempty" gorm:"size:255"`
	Configuration  SDKConfig `json:"configuration" gorm:"type:jsonb"`

	// Permissions & Scopes
	Scopes      []string       `json:"scopes" gorm:"type:text[]"`
	Permissions SDKPermissions `json:"permissions" gorm:"type:jsonb"`

	// Rate Limiting
	RateLimitConfig RateLimitConfig `json:"rate_limit_config" gorm:"type:jsonb"`

	// Usage Tracking
	TotalRequests   int64      `json:"total_requests" gorm:"default:0"`
	TotalErrors     int64      `json:"total_errors" gorm:"default:0"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP      string     `json:"last_used_ip,omitempty" gorm:"size:45"`
	LastUsedVersion string     `json:"last_used_version,omitempty" gorm:"size:50"`

	// Metadata
	Metadata JSONB    `json:"metadata,omitempty" gorm:"type:jsonb"`
	Tags     []string `json:"tags,omitempty" gorm:"type:text[]"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true;index:idx_sdk_clients_active"`

	// Relationships
	APIKeys []SDKKey `json:"api_keys,omitempty" gorm:"foreignKey:ClientID"`
}

// SDKKey represents an API key for SDK authentication
type SDKKey struct {
	BaseModel

	ClientID uuid.UUID  `json:"client_id" gorm:"type:uuid;not null;index:idx_sdk_keys_client"`
	Client   *SDKClient `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	TenantID uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_sdk_keys_tenant"`

	// Key Information
	Name        string `json:"name" gorm:"not null;size:255"`
	Description string `json:"description,omitempty" gorm:"type:text"`
	KeyHash     string `json:"-" gorm:"not null;uniqueIndex;size:255"` // Hashed API key (never return raw key after creation)
	KeyPrefix   string `json:"key_prefix" gorm:"not null;size:20"`     // First few chars for identification

	// Environment & Scope
	Environment SDKEnvironment `json:"environment" gorm:"type:varchar(50);not null;default:'production'"`
	Scopes      []string       `json:"scopes" gorm:"type:text[]"`

	// Rate Limiting (overrides client settings if set)
	RateLimitConfig *RateLimitConfig `json:"rate_limit_config,omitempty" gorm:"type:jsonb"`

	// Expiration
	ExpiresAt *time.Time `json:"expires_at,omitempty" gorm:"index:idx_sdk_keys_expiration"`

	// Usage Tracking
	TotalRequests int64      `json:"total_requests" gorm:"default:0"`
	TotalErrors   int64      `json:"total_errors" gorm:"default:0"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP    string     `json:"last_used_ip,omitempty" gorm:"size:45"`

	// Status
	Status SDKKeyStatus `json:"status" gorm:"type:varchar(50);not null;default:'active';index:idx_sdk_keys_status"`

	// Metadata
	CreatedBy    uuid.UUID  `json:"created_by,omitempty" gorm:"type:uuid"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	RevokedBy    *uuid.UUID `json:"revoked_by,omitempty" gorm:"type:uuid"`
	RevokeReason string     `json:"revoke_reason,omitempty" gorm:"type:text"`
}

// SDKUsage represents SDK usage analytics
type SDKUsage struct {
	BaseModel

	ClientID uuid.UUID  `json:"client_id" gorm:"type:uuid;not null;index:idx_sdk_usage_client"`
	Client   *SDKClient `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	TenantID uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_sdk_usage_tenant"`
	KeyID    *uuid.UUID `json:"key_id,omitempty" gorm:"type:uuid;index:idx_sdk_usage_key"`

	// Request Information
	Endpoint     string `json:"endpoint" gorm:"size:500;index:idx_sdk_usage_endpoint"`
	Method       string `json:"method" gorm:"size:10"`
	StatusCode   int    `json:"status_code" gorm:"index:idx_sdk_usage_status"`
	ResponseTime int64  `json:"response_time"` // milliseconds
	RequestSize  int64  `json:"request_size"`  // bytes
	ResponseSize int64  `json:"response_size"` // bytes

	// Client Information
	IPAddress  string `json:"ip_address,omitempty" gorm:"size:45;index:idx_sdk_usage_ip"`
	UserAgent  string `json:"user_agent,omitempty" gorm:"type:text"`
	SDKVersion string `json:"sdk_version,omitempty" gorm:"size:50;index:idx_sdk_usage_version"`
	AppVersion string `json:"app_version,omitempty" gorm:"size:50"`

	// Geographic Information
	Country string `json:"country,omitempty" gorm:"size:2"`
	Region  string `json:"region,omitempty" gorm:"size:100"`
	City    string `json:"city,omitempty" gorm:"size:100"`

	// Error Information
	IsError      bool   `json:"is_error" gorm:"default:false;index:idx_sdk_usage_error"`
	ErrorCode    string `json:"error_code,omitempty" gorm:"size:50"`
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`

	// Timestamp (indexed for time-series queries)
	Timestamp time.Time `json:"timestamp" gorm:"not null;index:idx_sdk_usage_timestamp"`

	// Additional Data
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`
}

// SDKConfig defines SDK-specific configuration
type SDKConfig struct {
	// API Configuration
	BaseURL       string `json:"base_url,omitempty"`
	Timeout       int    `json:"timeout,omitempty"` // seconds
	RetryAttempts int    `json:"retry_attempts,omitempty"`
	RetryDelay    int    `json:"retry_delay,omitempty"` // milliseconds

	// Features
	EnableLogging     bool `json:"enable_logging"`
	EnableAnalytics   bool `json:"enable_analytics"`
	EnableCaching     bool `json:"enable_caching"`
	EnableOfflineMode bool `json:"enable_offline_mode"`

	// Security
	RequireHTTPS       bool     `json:"require_https"`
	AllowInsecure      bool     `json:"allow_insecure"`
	CertificatePinning bool     `json:"certificate_pinning"`
	PublicKeyPins      []string `json:"public_key_pins,omitempty"`

	// Custom Settings
	CustomSettings map[string]any `json:"custom_settings,omitempty"`
}

// SDKPermissions defines granular permissions for SDK clients
type SDKPermissions struct {
	// Read Permissions
	CanReadBookings      bool `json:"can_read_bookings"`
	CanReadServices      bool `json:"can_read_services"`
	CanReadCustomers     bool `json:"can_read_customers"`
	CanReadArtisans      bool `json:"can_read_artisans"`
	CanReadProjects      bool `json:"can_read_projects"`
	CanReadInvoices      bool `json:"can_read_invoices"`
	CanReadPayments      bool `json:"can_read_payments"`
	CanReadReviews       bool `json:"can_read_reviews"`
	CanReadMessages      bool `json:"can_read_messages"`
	CanReadNotifications bool `json:"can_read_notifications"`

	// Write Permissions
	CanCreateBookings bool `json:"can_create_bookings"`
	CanUpdateBookings bool `json:"can_update_bookings"`
	CanCancelBookings bool `json:"can_cancel_bookings"`
	CanCreateReviews  bool `json:"can_create_reviews"`
	CanSendMessages   bool `json:"can_send_messages"`
	CanUploadFiles    bool `json:"can_upload_files"`
	CanMakePayments   bool `json:"can_make_payments"`

	// Special Permissions
	CanAccessAnalytics bool `json:"can_access_analytics"`
	CanManageProfile   bool `json:"can_manage_profile"`
	CanAccessWebhooks  bool `json:"can_access_webhooks"`
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool           `json:"enabled"`
	RequestsPerMinute int            `json:"requests_per_minute"`
	RequestsPerHour   int            `json:"requests_per_hour"`
	RequestsPerDay    int            `json:"requests_per_day"`
	BurstSize         int            `json:"burst_size"`
	CustomLimits      map[string]int `json:"custom_limits,omitempty"` // endpoint -> limit
}

// Scan/Value implementations for JSONB fields
func (c *SDKConfig) Scan(value any) error {
	if value == nil {
		*c = SDKConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan SDKConfig")
	}
	return json.Unmarshal(bytes, c)
}

func (c SDKConfig) Value() (driver.Value, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	if string(data) == "{}" || string(data) == "null" {
		return nil, nil
	}
	return data, nil
}

func (p *SDKPermissions) Scan(value any) error {
	if value == nil {
		*p = SDKPermissions{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan SDKPermissions")
	}
	return json.Unmarshal(bytes, p)
}

func (p SDKPermissions) Value() (driver.Value, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	if string(data) == "{}" || string(data) == "null" {
		return nil, nil
	}
	return data, nil
}

func (r *RateLimitConfig) Scan(value any) error {
	if value == nil {
		*r = RateLimitConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan RateLimitConfig")
	}
	return json.Unmarshal(bytes, r)
}

func (r RateLimitConfig) Value() (driver.Value, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	if string(data) == "{}" || string(data) == "null" {
		return nil, nil
	}
	return data, nil
}

// TableName specifies the table name for the SDKClient model
func (SDKClient) TableName() string {
	return "sdk_clients"
}

// TableName specifies the table name for the SDKKey model
func (SDKKey) TableName() string {
	return "sdk_keys"
}

// TableName specifies the table name for the SDKUsage model
func (SDKUsage) TableName() string {
	return "sdk_usage"
}

// IsExpired checks if the SDK key has expired
func (k *SDKKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsValid checks if the SDK key is valid for use
func (k *SDKKey) IsValid() bool {
	return k.Status == SDKKeyStatusActive && !k.IsExpired()
}

// CanAccess checks if the SDK key has access to a specific scope
func (k *SDKKey) CanAccess(scope string) bool {
	if !k.IsValid() {
		return false
	}
	for _, s := range k.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}
