package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// SDK Client Request DTOs
// ============================================================================

// CreateSDKClientRequest represents a request to create an SDK client
type CreateSDKClientRequest struct {
	Name           string                  `json:"name" validate:"required,min=2,max=255"`
	Description    string                  `json:"description,omitempty"`
	Platform       models.SDKPlatform      `json:"platform" validate:"required"`
	Environment    models.SDKEnvironment   `json:"environment" validate:"required"`
	AppIdentifier  string                  `json:"app_identifier,omitempty"`
	AppVersion     string                  `json:"app_version,omitempty"`
	AllowedOrigins []string                `json:"allowed_origins,omitempty"`
	AllowedIPs     []string                `json:"allowed_ips,omitempty"`
	WebhookURL     string                  `json:"webhook_url,omitempty" validate:"omitempty,url"`
	Configuration  models.SDKConfig        `json:"configuration"`
	Scopes         []string                `json:"scopes,omitempty"`
	Permissions    models.SDKPermissions   `json:"permissions"`
	RateLimitConfig models.RateLimitConfig `json:"rate_limit_config"`
	Tags           []string                `json:"tags,omitempty"`
	IsActive       bool                    `json:"is_active"`
}

// UpdateSDKClientRequest represents a request to update an SDK client
type UpdateSDKClientRequest struct {
	Name            *string                  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description     *string                  `json:"description,omitempty"`
	AppIdentifier   *string                  `json:"app_identifier,omitempty"`
	AppVersion      *string                  `json:"app_version,omitempty"`
	AllowedOrigins  *[]string                `json:"allowed_origins,omitempty"`
	AllowedIPs      *[]string                `json:"allowed_ips,omitempty"`
	WebhookURL      *string                  `json:"webhook_url,omitempty" validate:"omitempty,url"`
	Configuration   *models.SDKConfig        `json:"configuration,omitempty"`
	Scopes          *[]string                `json:"scopes,omitempty"`
	Permissions     *models.SDKPermissions   `json:"permissions,omitempty"`
	RateLimitConfig *models.RateLimitConfig  `json:"rate_limit_config,omitempty"`
	Tags            *[]string                `json:"tags,omitempty"`
	IsActive        *bool                    `json:"is_active,omitempty"`
}

// CreateSDKKeyRequest represents a request to create an SDK key
type CreateSDKKeyRequest struct {
	ClientID        uuid.UUID               `json:"client_id" validate:"required"`
	Name            string                  `json:"name" validate:"required,min=2,max=255"`
	Description     string                  `json:"description,omitempty"`
	Environment     models.SDKEnvironment   `json:"environment" validate:"required"`
	Scopes          []string                `json:"scopes,omitempty"`
	RateLimitConfig *models.RateLimitConfig `json:"rate_limit_config,omitempty"`
	ExpiresInDays   *int                    `json:"expires_in_days,omitempty" validate:"omitempty,min=1,max=3650"`
}

// UpdateSDKKeyRequest represents a request to update an SDK key
type UpdateSDKKeyRequest struct {
	Name            *string                  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description     *string                  `json:"description,omitempty"`
	Scopes          *[]string                `json:"scopes,omitempty"`
	RateLimitConfig *models.RateLimitConfig  `json:"rate_limit_config,omitempty"`
	Status          *models.SDKKeyStatus     `json:"status,omitempty"`
}

// RevokeSDKKeyRequest represents a request to revoke an SDK key
type RevokeSDKKeyRequest struct {
	Reason string `json:"reason,omitempty" validate:"max=500"`
}

// RotateSDKKeyRequest represents a request to rotate an SDK key
type RotateSDKKeyRequest struct {
	ExpiresInDays *int `json:"expires_in_days,omitempty" validate:"omitempty,min=1,max=3650"`
}

// SDKClientFilter represents filters for SDK client queries
type SDKClientFilter struct {
	Platform    *models.SDKPlatform    `json:"platform,omitempty"`
	Environment *models.SDKEnvironment `json:"environment,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Page        int                    `json:"page"`
	PageSize    int                    `json:"page_size"`
}

// SDKKeyFilter represents filters for SDK key queries
type SDKKeyFilter struct {
	ClientID    *uuid.UUID             `json:"client_id,omitempty"`
	Environment *models.SDKEnvironment `json:"environment,omitempty"`
	Status      *models.SDKKeyStatus   `json:"status,omitempty"`
	Page        int                    `json:"page"`
	PageSize    int                    `json:"page_size"`
}

// SDKUsageFilter represents filters for SDK usage queries
type SDKUsageFilter struct {
	ClientID   *uuid.UUID `json:"client_id,omitempty"`
	KeyID      *uuid.UUID `json:"key_id,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	StatusCode *int       `json:"status_code,omitempty"`
	IsError    *bool      `json:"is_error,omitempty"`
	Endpoint   *string    `json:"endpoint,omitempty"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}

// ============================================================================
// SDK Client Response DTOs
// ============================================================================

// SDKClientResponse represents an SDK client
type SDKClientResponse struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Platform        models.SDKPlatform     `json:"platform"`
	Environment     models.SDKEnvironment  `json:"environment"`
	AppIdentifier   string                 `json:"app_identifier,omitempty"`
	AppVersion      string                 `json:"app_version,omitempty"`
	SDKVersion      string                 `json:"sdk_version,omitempty"`
	AllowedOrigins  []string               `json:"allowed_origins,omitempty"`
	AllowedIPs      []string               `json:"allowed_ips,omitempty"`
	WebhookURL      string                 `json:"webhook_url,omitempty"`
	Configuration   models.SDKConfig       `json:"configuration"`
	Scopes          []string               `json:"scopes"`
	Permissions     models.SDKPermissions  `json:"permissions"`
	RateLimitConfig models.RateLimitConfig `json:"rate_limit_config"`
	TotalRequests   int64                  `json:"total_requests"`
	TotalErrors     int64                  `json:"total_errors"`
	LastUsedAt      *time.Time             `json:"last_used_at,omitempty"`
	LastUsedIP      string                 `json:"last_used_ip,omitempty"`
	LastUsedVersion string                 `json:"last_used_version,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	IsActive        bool                   `json:"is_active"`
	KeyCount        int                    `json:"key_count"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// SDKClientListResponse represents a paginated list of SDK clients
type SDKClientListResponse struct {
	Clients     []*SDKClientResponse `json:"clients"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"page_size"`
	TotalItems  int64                `json:"total_items"`
	TotalPages  int                  `json:"total_pages"`
	HasNext     bool                 `json:"has_next"`
	HasPrevious bool                 `json:"has_previous"`
}

// SDKKeyResponse represents an SDK key
type SDKKeyResponse struct {
	ID              uuid.UUID               `json:"id"`
	ClientID        uuid.UUID               `json:"client_id"`
	TenantID        uuid.UUID               `json:"tenant_id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description,omitempty"`
	KeyPrefix       string                  `json:"key_prefix"`
	Environment     models.SDKEnvironment   `json:"environment"`
	Scopes          []string                `json:"scopes"`
	RateLimitConfig *models.RateLimitConfig `json:"rate_limit_config,omitempty"`
	ExpiresAt       *time.Time              `json:"expires_at,omitempty"`
	TotalRequests   int64                   `json:"total_requests"`
	TotalErrors     int64                   `json:"total_errors"`
	LastUsedAt      *time.Time              `json:"last_used_at,omitempty"`
	LastUsedIP      string                  `json:"last_used_ip,omitempty"`
	Status          models.SDKKeyStatus     `json:"status"`
	CreatedBy       uuid.UUID               `json:"created_by,omitempty"`
	RevokedAt       *time.Time              `json:"revoked_at,omitempty"`
	RevokedBy       *uuid.UUID              `json:"revoked_by,omitempty"`
	RevokeReason    string                  `json:"revoke_reason,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// SDKKeyWithSecretResponse includes the actual API key (only returned once at creation)
type SDKKeyWithSecretResponse struct {
	*SDKKeyResponse
	APIKey string `json:"api_key"` // Only returned at creation
}

// SDKKeyListResponse represents a paginated list of SDK keys
type SDKKeyListResponse struct {
	Keys        []*SDKKeyResponse `json:"keys"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
	TotalItems  int64             `json:"total_items"`
	TotalPages  int               `json:"total_pages"`
	HasNext     bool              `json:"has_next"`
	HasPrevious bool              `json:"has_previous"`
}

// SDKUsageResponse represents SDK usage data
type SDKUsageResponse struct {
	ID           uuid.UUID  `json:"id"`
	ClientID     uuid.UUID  `json:"client_id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	KeyID        *uuid.UUID `json:"key_id,omitempty"`
	Endpoint     string     `json:"endpoint"`
	Method       string     `json:"method"`
	StatusCode   int        `json:"status_code"`
	ResponseTime int64      `json:"response_time"`
	RequestSize  int64      `json:"request_size"`
	ResponseSize int64      `json:"response_size"`
	IPAddress    string     `json:"ip_address,omitempty"`
	UserAgent    string     `json:"user_agent,omitempty"`
	SDKVersion   string     `json:"sdk_version,omitempty"`
	AppVersion   string     `json:"app_version,omitempty"`
	Country      string     `json:"country,omitempty"`
	Region       string     `json:"region,omitempty"`
	City         string     `json:"city,omitempty"`
	IsError      bool       `json:"is_error"`
	ErrorCode    string     `json:"error_code,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	Timestamp    time.Time  `json:"timestamp"`
}

// SDKUsageListResponse represents a paginated list of SDK usage
type SDKUsageListResponse struct {
	Usage       []*SDKUsageResponse `json:"usage"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
	TotalItems  int64               `json:"total_items"`
	TotalPages  int                 `json:"total_pages"`
	HasNext     bool                `json:"has_next"`
	HasPrevious bool                `json:"has_previous"`
}

// SDKUsageStatsResponse represents aggregated usage statistics
type SDKUsageStatsResponse struct {
	ClientID         uuid.UUID         `json:"client_id"`
	TotalRequests    int64             `json:"total_requests"`
	TotalErrors      int64             `json:"total_errors"`
	ErrorRate        float64           `json:"error_rate"`
	AvgResponseTime  float64           `json:"avg_response_time"`
	TotalDataSent    int64             `json:"total_data_sent"`
	TotalDataReceived int64            `json:"total_data_received"`
	TopEndpoints     []EndpointStat    `json:"top_endpoints"`
	StatusCodeDist   map[string]int64  `json:"status_code_distribution"`
	RequestsByDay    []DailyStat       `json:"requests_by_day"`
	ErrorsByType     map[string]int64  `json:"errors_by_type"`
	GeographicDist   []GeographicStat  `json:"geographic_distribution"`
}

// EndpointStat represents statistics for an endpoint
type EndpointStat struct {
	Endpoint      string  `json:"endpoint"`
	RequestCount  int64   `json:"request_count"`
	ErrorCount    int64   `json:"error_count"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

// DailyStat represents daily statistics
type DailyStat struct {
	Date     time.Time `json:"date"`
	Requests int64     `json:"requests"`
	Errors   int64     `json:"errors"`
}

// GeographicStat represents geographic distribution
type GeographicStat struct {
	Country      string `json:"country"`
	RequestCount int64  `json:"request_count"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToSDKClientResponse converts an SDKClient model to SDKClientResponse DTO
func ToSDKClientResponse(client *models.SDKClient) *SDKClientResponse {
	if client == nil {
		return nil
	}

	return &SDKClientResponse{
		ID:              client.ID,
		TenantID:        client.TenantID,
		Name:            client.Name,
		Description:     client.Description,
		Platform:        client.Platform,
		Environment:     client.Environment,
		AppIdentifier:   client.AppIdentifier,
		AppVersion:      client.AppVersion,
		SDKVersion:      client.SDKVersion,
		AllowedOrigins:  client.AllowedOrigins,
		AllowedIPs:      client.AllowedIPs,
		WebhookURL:      client.WebhookURL,
		Configuration:   client.Configuration,
		Scopes:          client.Scopes,
		Permissions:     client.Permissions,
		RateLimitConfig: client.RateLimitConfig,
		TotalRequests:   client.TotalRequests,
		TotalErrors:     client.TotalErrors,
		LastUsedAt:      client.LastUsedAt,
		LastUsedIP:      client.LastUsedIP,
		LastUsedVersion: client.LastUsedVersion,
		Tags:            client.Tags,
		IsActive:        client.IsActive,
		KeyCount:        len(client.APIKeys),
		CreatedAt:       client.CreatedAt,
		UpdatedAt:       client.UpdatedAt,
	}
}

// ToSDKClientResponses converts multiple SDKClient models to DTOs
func ToSDKClientResponses(clients []*models.SDKClient) []*SDKClientResponse {
	if clients == nil {
		return nil
	}

	responses := make([]*SDKClientResponse, len(clients))
	for i, client := range clients {
		responses[i] = ToSDKClientResponse(client)
	}
	return responses
}

// ToSDKKeyResponse converts an SDKKey model to SDKKeyResponse DTO
func ToSDKKeyResponse(key *models.SDKKey) *SDKKeyResponse {
	if key == nil {
		return nil
	}

	return &SDKKeyResponse{
		ID:              key.ID,
		ClientID:        key.ClientID,
		TenantID:        key.TenantID,
		Name:            key.Name,
		Description:     key.Description,
		KeyPrefix:       key.KeyPrefix,
		Environment:     key.Environment,
		Scopes:          key.Scopes,
		RateLimitConfig: key.RateLimitConfig,
		ExpiresAt:       key.ExpiresAt,
		TotalRequests:   key.TotalRequests,
		TotalErrors:     key.TotalErrors,
		LastUsedAt:      key.LastUsedAt,
		LastUsedIP:      key.LastUsedIP,
		Status:          key.Status,
		CreatedBy:       key.CreatedBy,
		RevokedAt:       key.RevokedAt,
		RevokedBy:       key.RevokedBy,
		RevokeReason:    key.RevokeReason,
		CreatedAt:       key.CreatedAt,
		UpdatedAt:       key.UpdatedAt,
	}
}

// ToSDKKeyResponses converts multiple SDKKey models to DTOs
func ToSDKKeyResponses(keys []*models.SDKKey) []*SDKKeyResponse {
	if keys == nil {
		return nil
	}

	responses := make([]*SDKKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = ToSDKKeyResponse(key)
	}
	return responses
}

// ToSDKUsageResponse converts an SDKUsage model to SDKUsageResponse DTO
func ToSDKUsageResponse(usage *models.SDKUsage) *SDKUsageResponse {
	if usage == nil {
		return nil
	}

	return &SDKUsageResponse{
		ID:           usage.ID,
		ClientID:     usage.ClientID,
		TenantID:     usage.TenantID,
		KeyID:        usage.KeyID,
		Endpoint:     usage.Endpoint,
		Method:       usage.Method,
		StatusCode:   usage.StatusCode,
		ResponseTime: usage.ResponseTime,
		RequestSize:  usage.RequestSize,
		ResponseSize: usage.ResponseSize,
		IPAddress:    usage.IPAddress,
		UserAgent:    usage.UserAgent,
		SDKVersion:   usage.SDKVersion,
		AppVersion:   usage.AppVersion,
		Country:      usage.Country,
		Region:       usage.Region,
		City:         usage.City,
		IsError:      usage.IsError,
		ErrorCode:    usage.ErrorCode,
		ErrorMessage: usage.ErrorMessage,
		Timestamp:    usage.Timestamp,
	}
}

// ToSDKUsageResponses converts multiple SDKUsage models to DTOs
func ToSDKUsageResponses(usages []*models.SDKUsage) []*SDKUsageResponse {
	if usages == nil {
		return nil
	}

	responses := make([]*SDKUsageResponse, len(usages))
	for i, usage := range usages {
		responses[i] = ToSDKUsageResponse(usage)
	}
	return responses
}
