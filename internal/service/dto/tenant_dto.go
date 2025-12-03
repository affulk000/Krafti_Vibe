package dto

import (
	"fmt"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Request DTOs
// ============================================================================

// CreateTenantRequest represents the request to create a tenant
type CreateTenantRequest struct {
	OwnerID       uuid.UUID              `json:"owner_id" validate:"required"`
	Name          string                 `json:"name" validate:"required,min=2,max=100"`
	Subdomain     string                 `json:"subdomain" validate:"required,min=3,max=63,alphanum"`
	BusinessName  string                 `json:"business_name" validate:"required,min=2,max=200"`
	BusinessEmail string                 `json:"business_email" validate:"required,email"`
	BusinessPhone string                 `json:"business_phone,omitempty"`
	TaxID         string                 `json:"tax_id,omitempty"`
	Plan          models.TenantPlan      `json:"plan" validate:"required"`
	TrialDays     int                    `json:"trial_days" validate:"min=0,max=90"`
	Settings      *models.TenantSettings `json:"settings,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the create tenant request
func (r *CreateTenantRequest) Validate() error {
	if r.OwnerID == uuid.Nil {
		return fmt.Errorf("owner ID is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("tenant name is required")
	}
	if len(r.Name) < 2 || len(r.Name) > 100 {
		return fmt.Errorf("tenant name must be between 2 and 100 characters")
	}
	if strings.TrimSpace(r.Subdomain) == "" {
		return fmt.Errorf("subdomain is required")
	}
	if len(r.Subdomain) < 3 || len(r.Subdomain) > 63 {
		return fmt.Errorf("subdomain must be between 3 and 63 characters")
	}
	if strings.TrimSpace(r.BusinessName) == "" {
		return fmt.Errorf("business name is required")
	}
	if strings.TrimSpace(r.BusinessEmail) == "" {
		return fmt.Errorf("business email is required")
	}
	if r.TrialDays < 0 || r.TrialDays > 90 {
		return fmt.Errorf("trial days must be between 0 and 90")
	}
	return nil
}

// Sanitize sanitizes the create tenant request
func (r *CreateTenantRequest) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Subdomain = strings.ToLower(strings.TrimSpace(r.Subdomain))
	r.BusinessName = strings.TrimSpace(r.BusinessName)
	r.BusinessEmail = strings.ToLower(strings.TrimSpace(r.BusinessEmail))
	r.BusinessPhone = strings.TrimSpace(r.BusinessPhone)
	r.TaxID = strings.TrimSpace(r.TaxID)
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	Name          *string                `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Domain        *string                `json:"domain,omitempty" validate:"omitempty,min=3,max=255"`
	BusinessName  *string                `json:"business_name,omitempty" validate:"omitempty,min=2,max=200"`
	BusinessEmail *string                `json:"business_email,omitempty" validate:"omitempty,email"`
	BusinessPhone *string                `json:"business_phone,omitempty"`
	TaxID         *string                `json:"tax_id,omitempty"`
	LogoURL       *string                `json:"logo_url,omitempty" validate:"omitempty,url"`
	PrimaryColor  *string                `json:"primary_color,omitempty" validate:"omitempty,len=7"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the update tenant request
func (r *UpdateTenantRequest) Validate() error {
	if r.Name != nil && (len(*r.Name) < 2 || len(*r.Name) > 100) {
		return fmt.Errorf("tenant name must be between 2 and 100 characters")
	}
	if r.Domain != nil && (len(*r.Domain) < 3 || len(*r.Domain) > 255) {
		return fmt.Errorf("domain must be between 3 and 255 characters")
	}
	if r.BusinessName != nil && (len(*r.BusinessName) < 2 || len(*r.BusinessName) > 200) {
		return fmt.Errorf("business name must be between 2 and 200 characters")
	}
	if r.PrimaryColor != nil && len(*r.PrimaryColor) != 7 {
		return fmt.Errorf("primary color must be a valid hex color (e.g., #FF5733)")
	}
	return nil
}

// Sanitize sanitizes the update tenant request
func (r *UpdateTenantRequest) Sanitize() {
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		r.Name = &trimmed
	}
	if r.Domain != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*r.Domain))
		r.Domain = &trimmed
	}
	if r.BusinessName != nil {
		trimmed := strings.TrimSpace(*r.BusinessName)
		r.BusinessName = &trimmed
	}
	if r.BusinessEmail != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*r.BusinessEmail))
		r.BusinessEmail = &trimmed
	}
	if r.BusinessPhone != nil {
		trimmed := strings.TrimSpace(*r.BusinessPhone)
		r.BusinessPhone = &trimmed
	}
	if r.TaxID != nil {
		trimmed := strings.TrimSpace(*r.TaxID)
		r.TaxID = &trimmed
	}
	if r.PrimaryColor != nil {
		trimmed := strings.ToUpper(strings.TrimSpace(*r.PrimaryColor))
		r.PrimaryColor = &trimmed
	}
}

// UpdateTenantPlanRequest represents the request to update tenant plan
type UpdateTenantPlanRequest struct {
	Plan models.TenantPlan `json:"plan" validate:"required"`
}

// Validate validates the update tenant plan request
func (r *UpdateTenantPlanRequest) Validate() error {
	if r.Plan == "" {
		return fmt.Errorf("plan is required")
	}
	return nil
}

// SuspendTenantRequest represents the request to suspend a tenant
type SuspendTenantRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500"`
}

// Validate validates the suspend tenant request
func (r *SuspendTenantRequest) Validate() error {
	if strings.TrimSpace(r.Reason) == "" {
		return fmt.Errorf("suspension reason is required")
	}
	if len(r.Reason) < 5 || len(r.Reason) > 500 {
		return fmt.Errorf("suspension reason must be between 5 and 500 characters")
	}
	return nil
}

// CancelTenantRequest represents the request to cancel a tenant
type CancelTenantRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500"`
}

// Validate validates the cancel tenant request
func (r *CancelTenantRequest) Validate() error {
	if strings.TrimSpace(r.Reason) == "" {
		return fmt.Errorf("cancellation reason is required")
	}
	if len(r.Reason) < 5 || len(r.Reason) > 500 {
		return fmt.Errorf("cancellation reason must be between 5 and 500 characters")
	}
	return nil
}

// UpdateStorageUsageRequest represents the request to update storage usage
type UpdateStorageUsageRequest struct {
	BytesUsed int64 `json:"bytes_used" validate:"required,min=0"`
}

// Validate validates the update storage usage request
func (r *UpdateStorageUsageRequest) Validate() error {
	if r.BytesUsed < 0 {
		return fmt.Errorf("bytes used cannot be negative")
	}
	return nil
}

// ValidateLimitsRequest represents the request to validate tenant limits
type ValidateLimitsRequest struct {
	Action string `json:"action" validate:"required"`
}

// Validate validates the validate limits request
func (r *ValidateLimitsRequest) Validate() error {
	if strings.TrimSpace(r.Action) == "" {
		return fmt.Errorf("action is required")
	}
	return nil
}

// ============================================================================
// Filter DTOs
// ============================================================================

// TenantFilter represents filters for listing tenants
type TenantFilter struct {
	Status         *models.TenantStatus `json:"status,omitempty"`
	Plan           *models.TenantPlan   `json:"plan,omitempty"`
	OwnerID        *uuid.UUID           `json:"owner_id,omitempty"`
	IsTrialExpired *bool                `json:"is_trial_expired,omitempty"`
	SearchQuery    string               `json:"search_query,omitempty"`
	CreatedAfter   *time.Time           `json:"created_after,omitempty"`
	CreatedBefore  *time.Time           `json:"created_before,omitempty"`
	Page           int                  `json:"page" validate:"min=1"`
	PageSize       int                  `json:"page_size" validate:"min=1,max=100"`
	SortBy         string               `json:"sort_by,omitempty"`
	SortOrder      string               `json:"sort_order,omitempty"` // asc or desc
}

// Validate validates the tenant filter
func (f *TenantFilter) Validate() error {
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
	return nil
}

// SearchTenantsRequest represents the request to search tenants
type SearchTenantsRequest struct {
	Query    string `json:"query" validate:"required,min=2"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
}

// Validate validates the search tenants request
func (r *SearchTenantsRequest) Validate() error {
	if strings.TrimSpace(r.Query) == "" {
		return fmt.Errorf("search query is required")
	}
	if len(r.Query) < 2 {
		return fmt.Errorf("search query must be at least 2 characters")
	}
	if r.Page < 1 {
		r.Page = 1
	}
	r.PageSize = max(1, min(r.PageSize, 100))
	if r.PageSize == 0 {
		r.PageSize = 20
	}
	return nil
}

// CheckSubdomainRequest represents the request to check subdomain availability
type CheckSubdomainRequest struct {
	Subdomain string `json:"subdomain" validate:"required,min=3,max=63"`
}

// Validate validates the check subdomain request
func (r *CheckSubdomainRequest) Validate() error {
	if strings.TrimSpace(r.Subdomain) == "" {
		return fmt.Errorf("subdomain is required")
	}
	if len(r.Subdomain) < 3 || len(r.Subdomain) > 63 {
		return fmt.Errorf("subdomain must be between 3 and 63 characters")
	}
	return nil
}

// Sanitize sanitizes the check subdomain request
func (r *CheckSubdomainRequest) Sanitize() {
	r.Subdomain = strings.ToLower(strings.TrimSpace(r.Subdomain))
}

// CheckDomainRequest represents the request to check domain availability
type CheckDomainRequest struct {
	Domain string `json:"domain" validate:"required,fqdn"`
}

// Validate validates the check domain request
func (r *CheckDomainRequest) Validate() error {
	if strings.TrimSpace(r.Domain) == "" {
		return fmt.Errorf("domain is required")
	}
	if len(r.Domain) < 3 || len(r.Domain) > 255 {
		return fmt.Errorf("domain must be between 3 and 255 characters")
	}
	return nil
}

// Sanitize sanitizes the check domain request
func (r *CheckDomainRequest) Sanitize() {
	r.Domain = strings.ToLower(strings.TrimSpace(r.Domain))
}

// UpdateTenantSettingsRequest represents the request to update tenant settings
type UpdateTenantSettingsRequest struct {
	Timezone                *string `json:"timezone,omitempty"`
	DateFormat              *string `json:"date_format,omitempty"`
	TimeFormat              *string `json:"time_format,omitempty"`
	Currency                *string `json:"currency,omitempty"`
	Language                *string `json:"language,omitempty"`
	AllowPublicBooking      *bool   `json:"allow_public_booking,omitempty"`
	RequireEmailVerify      *bool   `json:"require_email_verification,omitempty"`
	BookingApprovalRequired *bool   `json:"booking_approval_required,omitempty"`
	AutoAcceptBookings      *bool   `json:"auto_accept_bookings,omitempty"`
	EnableNotifications     *bool   `json:"enable_notifications,omitempty"`
	EnableSMS               *bool   `json:"enable_sms_notifications,omitempty"`
	EnableEmailReminders    *bool   `json:"enable_email_reminders,omitempty"`
	EnableWaitlist          *bool   `json:"enable_waitlist,omitempty"`
	EnableReviews           *bool   `json:"enable_reviews,omitempty"`
}

// UpdateTenantFeaturesRequest represents the request to update tenant features
type UpdateTenantFeaturesRequest struct {
	CanBookServices       *bool `json:"can_book_services,omitempty"`
	CanManageProjects     *bool `json:"can_manage_projects,omitempty"`
	CanAccessReports      *bool `json:"can_access_reports,omitempty"`
	CanUseAPI             *bool `json:"can_use_api,omitempty"`
	CanUseWebhooks        *bool `json:"can_use_webhooks,omitempty"`
	CanExportData         *bool `json:"can_export_data,omitempty"`
	CanCustomizeBranding  *bool `json:"can_customize_branding,omitempty"`
	CanWhiteLabel         *bool `json:"can_white_label,omitempty"`
	CanUseAdvancedReports *bool `json:"can_use_advanced_reports,omitempty"`
	CanPrioritizeSupport  *bool `json:"can_prioritize_support,omitempty"`
}

// ============================================================================
// Response DTOs
// ============================================================================

// TenantResponse represents a tenant response
// Note: SubscriptionID is string to match Tenant model (not *uuid.UUID)
// Note: CurrentArtisans is not in Tenant model, we calculate from MaxArtisans
type TenantResponse struct {
	ID                uuid.UUID              `json:"id"`
	OwnerID           uuid.UUID              `json:"owner_id"`
	Name              string                 `json:"name"`
	Subdomain         string                 `json:"subdomain"`
	Domain            string                 `json:"domain,omitempty"`
	BusinessName      string                 `json:"business_name"`
	BusinessEmail     string                 `json:"business_email"`
	BusinessPhone     string                 `json:"business_phone,omitempty"`
	TaxID             string                 `json:"tax_id,omitempty"`
	LogoURL           string                 `json:"logo_url,omitempty"`
	PrimaryColor      string                 `json:"primary_color,omitempty"`
	Plan              models.TenantPlan      `json:"plan"`
	Status            models.TenantStatus    `json:"status"`
	TrialEndsAt       *time.Time             `json:"trial_ends_at,omitempty"`
	SubscriptionID    string                 `json:"subscription_id,omitempty"`
	Settings          models.TenantSettings  `json:"settings"`
	Features          models.TenantFeatures  `json:"features"`
	MaxUsers          int                    `json:"max_users"`
	MaxArtisans       int                    `json:"max_artisans"`
	MaxStorage        int64                  `json:"max_storage"`
	CurrentUsers      int                    `json:"current_users"`
	StorageUsed       int64                  `json:"storage_used"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	IsTrialExpired    bool                   `json:"is_trial_expired"`
	DaysUntilExpiry   *int                   `json:"days_until_expiry,omitempty"`
	StoragePercentage float64                `json:"storage_percentage"`
	UsersPercentage   float64                `json:"users_percentage"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// TenantListResponse represents a paginated list of tenants
type TenantListResponse struct {
	Tenants    []*TenantResponse `json:"tenants"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalItems int64             `json:"total_items"`
	TotalPages int               `json:"total_pages"`
}

// TenantStats represents statistics for a tenant
type TenantStats struct {
	TenantID          uuid.UUID `json:"tenant_id"`
	TotalUsers        int       `json:"total_users"`
	TotalArtisans     int       `json:"total_artisans"`
	TotalCustomers    int       `json:"total_customers"`
	TotalBookings     int       `json:"total_bookings"`
	TotalProjects     int       `json:"total_projects"`
	TotalServices     int       `json:"total_services"`
	TotalRevenue      float64   `json:"total_revenue"`
	StorageUsedMB     int64     `json:"storage_used_mb"`
	StorageLimitMB    int64     `json:"storage_limit_mb"`
	StoragePercentage float64   `json:"storage_percentage"`
	ActiveBookings    int       `json:"active_bookings"`
	CompletedBookings int       `json:"completed_bookings"`
	CancelledBookings int       `json:"cancelled_bookings"`
	PendingBookings   int       `json:"pending_bookings"`
	RevenueThisMonth  float64   `json:"revenue_this_month"`
	RevenueLastMonth  float64   `json:"revenue_last_month"`
	GrowthRate        float64   `json:"growth_rate"`
	AverageRating     float64   `json:"average_rating"`
	TotalReviews      int       `json:"total_reviews"`
}

// TenantDetailsResponse represents detailed tenant information
type TenantDetailsResponse struct {
	Tenant *TenantResponse `json:"tenant"`
	Stats  *TenantStats    `json:"stats"`
}

// SubdomainAvailabilityResponse represents subdomain availability check response
type SubdomainAvailabilityResponse struct {
	Subdomain string `json:"subdomain"`
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

// DomainAvailabilityResponse represents domain availability check response
type DomainAvailabilityResponse struct {
	Domain    string `json:"domain"`
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

// TenantLimitsResponse represents tenant limits information
type TenantLimitsResponse struct {
	TenantID          uuid.UUID         `json:"tenant_id"`
	Plan              models.TenantPlan `json:"plan"`
	MaxUsers          int               `json:"max_users"`
	CurrentUsers      int               `json:"current_users"`
	UsersRemaining    int               `json:"users_remaining"`
	MaxArtisans       int               `json:"max_artisans"`
	CurrentArtisans   int               `json:"current_artisans"`
	ArtisansRemaining int               `json:"artisans_remaining"`
	MaxStorage        int64             `json:"max_storage"`
	StorageUsed       int64             `json:"storage_used"`
	StorageRemaining  int64             `json:"storage_remaining"`
	CanAddUser        bool              `json:"can_add_user"`
	CanAddArtisan     bool              `json:"can_add_artisan"`
	CanUploadFile     bool              `json:"can_upload_file"`
}

// TenantActivityResponse represents tenant activity
type TenantActivityResponse struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	TenantName    string    `json:"tenant_name"`
	LastActivity  time.Time `json:"last_activity"`
	ActivityCount int       `json:"activity_count"`
	IsActive      bool      `json:"is_active"`
}

// TenantHealthResponse represents tenant health check
type TenantHealthResponse struct {
	TenantID           uuid.UUID           `json:"tenant_id"`
	Status             models.TenantStatus `json:"status"`
	IsHealthy          bool                `json:"is_healthy"`
	HealthScore        int                 `json:"health_score"` // 0-100
	Issues             []string            `json:"issues,omitempty"`
	Warnings           []string            `json:"warnings,omitempty"`
	StorageHealth      string              `json:"storage_health"`
	UserLimitHealth    string              `json:"user_limit_health"`
	SubscriptionHealth string              `json:"subscription_health"`
	LastChecked        time.Time           `json:"last_checked"`
}

// ============================================================================
// Utility Functions
// ============================================================================

// ToTenantResponse converts a models.Tenant to TenantResponse
func ToTenantResponse(tenant *models.Tenant) *TenantResponse {
	if tenant == nil {
		return nil
	}

	response := &TenantResponse{
		ID:             tenant.ID,
		OwnerID:        tenant.OwnerID,
		Name:           tenant.Name,
		Subdomain:      tenant.Subdomain,
		Domain:         tenant.Domain,
		BusinessName:   tenant.BusinessName,
		BusinessEmail:  tenant.BusinessEmail,
		BusinessPhone:  tenant.BusinessPhone,
		TaxID:          tenant.TaxID,
		LogoURL:        tenant.LogoURL,
		PrimaryColor:   tenant.PrimaryColor,
		Plan:           tenant.Plan,
		Status:         tenant.Status,
		TrialEndsAt:    tenant.TrialEndsAt,
		SubscriptionID: tenant.SubscriptionID, // string type matches
		Settings:       tenant.Settings,
		Features:       tenant.Features,
		MaxUsers:       tenant.MaxUsers,
		MaxArtisans:    tenant.MaxArtisans,
		MaxStorage:     tenant.MaxStorage,
		CurrentUsers:   tenant.CurrentUsers,
		StorageUsed:    tenant.StorageUsed,
		Metadata:       tenant.Metadata,
		CreatedAt:      tenant.CreatedAt,
		UpdatedAt:      tenant.UpdatedAt,
	}

	// Calculate trial expiry
	if tenant.TrialEndsAt != nil {
		response.IsTrialExpired = time.Now().After(*tenant.TrialEndsAt)
		if !response.IsTrialExpired {
			days := int(time.Until(*tenant.TrialEndsAt).Hours() / 24)
			response.DaysUntilExpiry = &days
		}
	}

	// Calculate percentages
	if tenant.MaxStorage > 0 {
		response.StoragePercentage = (float64(tenant.StorageUsed) / float64(tenant.MaxStorage)) * 100
	}
	if tenant.MaxUsers > 0 {
		response.UsersPercentage = (float64(tenant.CurrentUsers) / float64(tenant.MaxUsers)) * 100
	}

	return response
}

// ToTenantResponses converts multiple models.Tenant to TenantResponse slice
func ToTenantResponses(tenants []*models.Tenant) []*TenantResponse {
	responses := make([]*TenantResponse, 0, len(tenants))
	for _, tenant := range tenants {
		responses = append(responses, ToTenantResponse(tenant))
	}
	return responses
}
