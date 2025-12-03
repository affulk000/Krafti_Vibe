package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tenant-related errors
var (
	// ErrTenantNotFound is returned when a tenant is not found
	ErrTenantNotFound = errors.New("tenant not found")

	// ErrSubdomainTaken is returned when a subdomain is already taken
	ErrSubdomainTaken = errors.New("subdomain is already taken")

	// ErrDomainTaken is returned when a domain is already taken
	ErrDomainTaken = errors.New("domain is already taken")

	// ErrInvalidSubdomain is returned when subdomain format is invalid
	ErrInvalidSubdomain = errors.New("invalid subdomain format")

	// ErrTenantLimitReached is returned when tenant reaches a plan limit
	ErrTenantLimitReached = errors.New("tenant limit reached")

	// ErrTenantSuspended is returned when tenant is suspended
	ErrTenantSuspended = errors.New("tenant is suspended")

	// ErrTenantCancelled is returned when tenant is cancelled
	ErrTenantCancelled = errors.New("tenant is cancelled")

	// ErrOwnerAlreadyHasTenant is returned when user already owns a tenant
	ErrOwnerAlreadyHasTenant = errors.New("user already belongs to a tenant")

	// ErrInvalidPlan is returned when plan is invalid
	ErrInvalidPlan = errors.New("invalid tenant plan")

	// ErrInvalidStatusTransition is returned when status transition is not allowed
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// TenantService defines the interface for core tenant operations
type TenantService interface {
	// Core CRUD Operations
	CreateTenant(ctx context.Context, req *dto.CreateTenantRequest) (*dto.TenantResponse, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*dto.TenantResponse, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*dto.TenantResponse, error)
	GetTenantByDomain(ctx context.Context, domain string) (*dto.TenantResponse, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, req *dto.UpdateTenantRequest) (*dto.TenantResponse, error)
	DeleteTenant(ctx context.Context, id uuid.UUID) error

	// Listing and Search
	ListTenants(ctx context.Context, filter *dto.TenantFilter) (*dto.TenantListResponse, error)
	SearchTenants(ctx context.Context, req *dto.SearchTenantsRequest) (*dto.TenantListResponse, error)

	// Tenant Status Management
	ActivateTenant(ctx context.Context, id uuid.UUID) error
	SuspendTenant(ctx context.Context, req *dto.SuspendTenantRequest, tenantID uuid.UUID) error
	CancelTenant(ctx context.Context, req *dto.CancelTenantRequest, tenantID uuid.UUID) error

	// Plan and Settings Management
	UpdateTenantPlan(ctx context.Context, id uuid.UUID, req *dto.UpdateTenantPlanRequest) error
	UpdateTenantSettings(ctx context.Context, id uuid.UUID, req *dto.UpdateTenantSettingsRequest) error
	UpdateTenantFeatures(ctx context.Context, id uuid.UUID, req *dto.UpdateTenantFeaturesRequest) error

	// Validation and Availability
	CheckSubdomainAvailability(ctx context.Context, req *dto.CheckSubdomainRequest) (*dto.SubdomainAvailabilityResponse, error)
	CheckDomainAvailability(ctx context.Context, req *dto.CheckDomainRequest) (*dto.DomainAvailabilityResponse, error)
	ValidateTenantLimits(ctx context.Context, req *dto.ValidateLimitsRequest, tenantID uuid.UUID) error

	// Statistics and Monitoring
	GetTenantStats(ctx context.Context, id uuid.UUID) (*dto.TenantStats, error)
	GetTenantDetails(ctx context.Context, id uuid.UUID) (*dto.TenantDetailsResponse, error)
	GetTenantLimits(ctx context.Context, id uuid.UUID) (*dto.TenantLimitsResponse, error)
	GetTenantHealth(ctx context.Context, id uuid.UUID) (*dto.TenantHealthResponse, error)

	// Trial Management
	GetExpiredTrials(ctx context.Context) ([]*dto.TenantResponse, error)
	ExtendTrial(ctx context.Context, id uuid.UUID, days int) error

	// Usage Counter Management
	IncrementUserCount(ctx context.Context, id uuid.UUID) error
	DecrementUserCount(ctx context.Context, id uuid.UUID) error
	UpdateStorageUsage(ctx context.Context, req *dto.UpdateStorageUsageRequest, tenantID uuid.UUID) error
}

// tenantService implements TenantService
type tenantService struct {
	repos  *repository.Repositories
	logger *zap.Logger
}

// NewTenantService creates a new tenant service
func NewTenantService(
	repos *repository.Repositories,
	logger *zap.Logger,
) TenantService {
	return &tenantService{
		repos:  repos,
		logger: logger,
	}
}

// ============================================================================
// Core CRUD Operations
// ============================================================================

// CreateTenant creates a new tenant organization
func (s *tenantService) CreateTenant(ctx context.Context, req *dto.CreateTenantRequest) (*dto.TenantResponse, error) {
	req.Sanitize()
	if err := req.Validate(); err != nil {
		s.logger.Warn("invalid create tenant request", zap.Error(err))
		return nil, fmt.Errorf("validation error: %w", err)
	}

	s.logger.Info("creating tenant",
		zap.String("name", req.Name),
		zap.String("subdomain", req.Subdomain),
		zap.String("owner_id", req.OwnerID.String()),
		zap.String("plan", string(req.Plan)),
	)

	// Validate subdomain format
	if !isValidSubdomain(req.Subdomain) {
		return nil, ErrInvalidSubdomain
	}

	// Check if subdomain is available
	existing, err := s.repos.Tenant.FindBySubdomain(ctx, req.Subdomain)
	if err == nil && existing != nil {
		s.logger.Warn("subdomain already taken", zap.String("subdomain", req.Subdomain))
		return nil, ErrSubdomainTaken
	}

	// Verify owner exists
	owner, err := s.repos.User.GetByID(ctx, req.OwnerID)
	if err != nil {
		s.logger.Error("failed to find owner", zap.Error(err))
		return nil, fmt.Errorf("owner not found: %w", err)
	}

	// Check if owner already has a tenant
	if owner.TenantID != nil {
		return nil, ErrOwnerAlreadyHasTenant
	}

	// Get plan limits and features
	limits := getLimitsForPlan(req.Plan)
	features := getFeaturesForPlan(req.Plan)

	// Create tenant
	tenant := &models.Tenant{
		OwnerID:       req.OwnerID,
		Name:          req.Name,
		Subdomain:     req.Subdomain,
		BusinessName:  req.BusinessName,
		BusinessEmail: req.BusinessEmail,
		BusinessPhone: req.BusinessPhone,
		TaxID:         req.TaxID,
		Plan:          req.Plan,
		Status:        models.TenantStatusTrial,
		MaxUsers:      limits.MaxUsers,
		MaxArtisans:   limits.MaxArtisans,
		MaxStorage:    limits.MaxStorage,
		Features:      features,
		CurrentUsers:  1, // Owner counts as first user
	}

	// Set trial expiry
	if req.TrialDays > 0 {
		trialEnd := time.Now().AddDate(0, 0, req.TrialDays)
		tenant.TrialEndsAt = &trialEnd
	}

	// Apply custom settings if provided
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}

	// Apply metadata if provided
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}

	if err := s.repos.Tenant.Create(ctx, tenant); err != nil {
		s.logger.Error("failed to create tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Update owner with tenant reference
	owner.TenantID = &tenant.ID
	owner.Role = models.UserRoleTenantOwner
	if err := s.repos.User.Update(ctx, owner); err != nil {
		s.logger.Error("failed to update owner with tenant", zap.Error(err))
		// Don't fail - tenant is created
	}

	s.logger.Info("tenant created successfully",
		zap.String("tenant_id", tenant.ID.String()),
		zap.String("subdomain", tenant.Subdomain),
	)

	return dto.ToTenantResponse(tenant), nil
}

// GetTenant retrieves a tenant by ID
func (s *tenantService) GetTenant(ctx context.Context, id uuid.UUID) (*dto.TenantResponse, error) {
	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	return dto.ToTenantResponse(tenant), nil
}

// GetTenantBySubdomain retrieves a tenant by subdomain
func (s *tenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (*dto.TenantResponse, error) {
	subdomain = strings.ToLower(strings.TrimSpace(subdomain))
	tenant, err := s.repos.Tenant.FindBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	return dto.ToTenantResponse(tenant), nil
}

// GetTenantByDomain retrieves a tenant by custom domain
func (s *tenantService) GetTenantByDomain(ctx context.Context, domain string) (*dto.TenantResponse, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	tenant, err := s.repos.Tenant.FindByDomain(ctx, domain)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	return dto.ToTenantResponse(tenant), nil
}

// UpdateTenant updates tenant details
func (s *tenantService) UpdateTenant(ctx context.Context, id uuid.UUID, req *dto.UpdateTenantRequest) (*dto.TenantResponse, error) {
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	// Check domain availability if being updated
	if req.Domain != nil && *req.Domain != tenant.Domain {
		existing, err := s.repos.Tenant.FindByDomain(ctx, *req.Domain)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrDomainTaken
		}
		tenant.Domain = *req.Domain
	}

	// Update fields if provided
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.BusinessName != nil {
		tenant.BusinessName = *req.BusinessName
	}
	if req.BusinessEmail != nil {
		tenant.BusinessEmail = *req.BusinessEmail
	}
	if req.BusinessPhone != nil {
		tenant.BusinessPhone = *req.BusinessPhone
	}
	if req.TaxID != nil {
		tenant.TaxID = *req.TaxID
	}
	if req.LogoURL != nil {
		tenant.LogoURL = *req.LogoURL
	}
	if req.PrimaryColor != nil {
		tenant.PrimaryColor = *req.PrimaryColor
	}
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}

	if err := s.repos.Tenant.Update(ctx, tenant); err != nil {
		s.logger.Error("failed to update tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	s.logger.Info("tenant updated successfully", zap.String("tenant_id", id.String()))
	return dto.ToTenantResponse(tenant), nil
}

// DeleteTenant soft deletes a tenant
func (s *tenantService) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	s.logger.Warn("deleting tenant", zap.String("tenant_id", id.String()))

	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return ErrTenantNotFound
	}

	// Update tenant status to cancelled first
	tenant.Status = models.TenantStatusCancelled
	if err := s.repos.Tenant.Update(ctx, tenant); err != nil {
		s.logger.Error("failed to update tenant status", zap.Error(err))
	}

	// Soft delete the tenant
	if err := s.repos.Tenant.SoftDelete(ctx, id); err != nil {
		s.logger.Error("failed to delete tenant", zap.Error(err))
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	s.logger.Info("tenant deleted successfully", zap.String("tenant_id", id.String()))
	return nil
}

// ============================================================================
// Listing and Search
// ============================================================================

// ListTenants lists tenants with filtering and pagination
func (s *tenantService) ListTenants(ctx context.Context, filter *dto.TenantFilter) (*dto.TenantListResponse, error) {
	if filter == nil {
		filter = &dto.TenantFilter{Page: 1, PageSize: 20}
	}
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	// Build repository filters
	repoFilters := repository.TenantFilters{}
	if filter.Status != nil {
		repoFilters.Statuses = []models.TenantStatus{*filter.Status}
	}
	if filter.Plan != nil {
		repoFilters.Plans = []models.TenantPlan{*filter.Plan}
	}
	if filter.OwnerID != nil {
		repoFilters.OwnerID = filter.OwnerID
	}
	if filter.CreatedAfter != nil {
		repoFilters.CreatedAfter = filter.CreatedAfter
	}
	if filter.CreatedBefore != nil {
		repoFilters.CreatedBefore = filter.CreatedBefore
	}

	tenants, result, err := s.repos.Tenant.FindByFilters(ctx, repoFilters, pagination)
	if err != nil {
		s.logger.Error("failed to list tenants", zap.Error(err))
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return &dto.TenantListResponse{
		Tenants:    dto.ToTenantResponses(tenants),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: result.TotalPages,
	}, nil
}

// SearchTenants searches tenants by query
func (s *tenantService) SearchTenants(ctx context.Context, req *dto.SearchTenantsRequest) (*dto.TenantListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	pagination := repository.PaginationParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	tenants, result, err := s.repos.Tenant.Search(ctx, req.Query, pagination)
	if err != nil {
		s.logger.Error("failed to search tenants", zap.Error(err))
		return nil, fmt.Errorf("failed to search tenants: %w", err)
	}

	return &dto.TenantListResponse{
		Tenants:    dto.ToTenantResponses(tenants),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: result.TotalPages,
	}, nil
}

// ============================================================================
// Tenant Status Management
// ============================================================================

// ActivateTenant activates a tenant
func (s *tenantService) ActivateTenant(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("activating tenant", zap.String("tenant_id", id.String()))

	if err := s.repos.Tenant.ActivateTenant(ctx, id); err != nil {
		s.logger.Error("failed to activate tenant", zap.Error(err))
		return fmt.Errorf("failed to activate tenant: %w", err)
	}

	s.logger.Info("tenant activated successfully", zap.String("tenant_id", id.String()))
	return nil
}

// SuspendTenant suspends a tenant
func (s *tenantService) SuspendTenant(ctx context.Context, req *dto.SuspendTenantRequest, tenantID uuid.UUID) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	s.logger.Warn("suspending tenant",
		zap.String("tenant_id", tenantID.String()),
		zap.String("reason", req.Reason),
	)

	if err := s.repos.Tenant.SuspendTenant(ctx, tenantID, req.Reason); err != nil {
		s.logger.Error("failed to suspend tenant", zap.Error(err))
		return fmt.Errorf("failed to suspend tenant: %w", err)
	}

	s.logger.Info("tenant suspended successfully", zap.String("tenant_id", tenantID.String()))
	return nil
}

// CancelTenant cancels a tenant
func (s *tenantService) CancelTenant(ctx context.Context, req *dto.CancelTenantRequest, tenantID uuid.UUID) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	s.logger.Warn("cancelling tenant",
		zap.String("tenant_id", tenantID.String()),
		zap.String("reason", req.Reason),
	)

	if err := s.repos.Tenant.CancelTenant(ctx, tenantID, req.Reason); err != nil {
		s.logger.Error("failed to cancel tenant", zap.Error(err))
		return fmt.Errorf("failed to cancel tenant: %w", err)
	}

	s.logger.Info("tenant cancelled successfully", zap.String("tenant_id", tenantID.String()))
	return nil
}

// ============================================================================
// Plan and Settings Management
// ============================================================================

// UpdateTenantPlan updates the tenant's subscription plan
func (s *tenantService) UpdateTenantPlan(ctx context.Context, id uuid.UUID, req *dto.UpdateTenantPlanRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return ErrTenantNotFound
	}

	s.logger.Info("updating tenant plan",
		zap.String("tenant_id", id.String()),
		zap.String("old_plan", string(tenant.Plan)),
		zap.String("new_plan", string(req.Plan)),
	)

	// Get new plan limits and features
	limits := getLimitsForPlan(req.Plan)
	features := getFeaturesForPlan(req.Plan)

	// Update plan and limits
	tenant.Plan = req.Plan
	tenant.MaxUsers = limits.MaxUsers
	tenant.MaxArtisans = limits.MaxArtisans
	tenant.MaxStorage = limits.MaxStorage
	tenant.Features = features

	// If upgrading from trial, activate the tenant
	if tenant.Status == models.TenantStatusTrial {
		tenant.Status = models.TenantStatusActive
	}

	if err := s.repos.Tenant.Update(ctx, tenant); err != nil {
		s.logger.Error("failed to update tenant plan", zap.Error(err))
		return fmt.Errorf("failed to update plan: %w", err)
	}

	s.logger.Info("tenant plan updated successfully", zap.String("tenant_id", id.String()))
	return nil
}

// UpdateTenantSettings updates tenant settings
func (s *tenantService) UpdateTenantSettings(ctx context.Context, id uuid.UUID, req *dto.UpdateTenantSettingsRequest) error {
	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return ErrTenantNotFound
	}

	// Update settings fields if provided
	settings := tenant.Settings

	if req.Timezone != nil {
		settings.DefaultTimezone = *req.Timezone
	}
	if req.DateFormat != nil {
		settings.DateFormat = *req.DateFormat
	}
	if req.TimeFormat != nil {
		settings.TimeFormat = *req.TimeFormat
	}
	if req.Currency != nil {
		settings.DefaultCurrency = *req.Currency
	}
	if req.Language != nil {
		settings.DefaultLanguage = *req.Language
	}
	if req.AllowPublicBooking != nil {
		settings.AllowRecurringBookings = *req.AllowPublicBooking
	}
	if req.RequireEmailVerify != nil {
		settings.RequireCustomerAccount = *req.RequireEmailVerify
	}
	if req.BookingApprovalRequired != nil {
		settings.BookingApprovalRequired = *req.BookingApprovalRequired
	}
	if req.AutoAcceptBookings != nil {
		settings.AutoAcceptBookings = *req.AutoAcceptBookings
	}
	if req.EnableNotifications != nil {
		settings.EmailNotificationsEnabled = *req.EnableNotifications
	}
	if req.EnableSMS != nil {
		settings.SMSNotificationsEnabled = *req.EnableSMS
	}
	if req.EnableEmailReminders != nil {
		settings.EmailNotificationsEnabled = *req.EnableEmailReminders
	}
	if req.EnableWaitlist != nil {
		settings.EnableWaitlist = *req.EnableWaitlist
	}
	if req.EnableReviews != nil {
		settings.EnableCustomerReviews = *req.EnableReviews
	}

	if err := s.repos.Tenant.UpdateSettings(ctx, id, settings); err != nil {
		s.logger.Error("failed to update tenant settings", zap.Error(err))
		return fmt.Errorf("failed to update settings: %w", err)
	}

	s.logger.Info("tenant settings updated successfully", zap.String("tenant_id", id.String()))
	return nil
}

// UpdateTenantFeatures updates tenant feature flags
func (s *tenantService) UpdateTenantFeatures(ctx context.Context, id uuid.UUID, req *dto.UpdateTenantFeaturesRequest) error {
	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return ErrTenantNotFound
	}

	features := tenant.Features

	if req.CanBookServices != nil {
		features.BasicBooking = *req.CanBookServices
	}
	if req.CanManageProjects != nil {
		features.BasicProjects = *req.CanManageProjects
	}
	if req.CanAccessReports != nil {
		features.BasicAnalytics = *req.CanAccessReports
	}
	if req.CanUseAPI != nil {
		features.APIAccess = *req.CanUseAPI
	}
	if req.CanUseWebhooks != nil {
		features.WebhookSupport = *req.CanUseWebhooks
	}
	if req.CanExportData != nil {
		features.DataExport = *req.CanExportData
	}
	if req.CanCustomizeBranding != nil {
		features.CustomBranding = *req.CanCustomizeBranding
	}
	if req.CanWhiteLabel != nil {
		features.WhiteLabeling = *req.CanWhiteLabel
	}
	if req.CanUseAdvancedReports != nil {
		features.AdvancedAnalytics = *req.CanUseAdvancedReports
	}
	if req.CanPrioritizeSupport != nil {
		features.PrioritySupport = *req.CanPrioritizeSupport
	}

	if err := s.repos.Tenant.UpdateFeatures(ctx, id, features); err != nil {
		s.logger.Error("failed to update tenant features", zap.Error(err))
		return fmt.Errorf("failed to update features: %w", err)
	}

	s.logger.Info("tenant features updated successfully", zap.String("tenant_id", id.String()))
	return nil
}

// ============================================================================
// Validation and Availability
// ============================================================================

// CheckSubdomainAvailability checks if a subdomain is available
func (s *tenantService) CheckSubdomainAvailability(ctx context.Context, req *dto.CheckSubdomainRequest) (*dto.SubdomainAvailabilityResponse, error) {
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if !isValidSubdomain(req.Subdomain) {
		return &dto.SubdomainAvailabilityResponse{
			Subdomain: req.Subdomain,
			Available: false,
			Message:   "Invalid subdomain format",
		}, nil
	}

	existing, err := s.repos.Tenant.FindBySubdomain(ctx, req.Subdomain)
	if err == nil && existing != nil {
		return &dto.SubdomainAvailabilityResponse{
			Subdomain: req.Subdomain,
			Available: false,
			Message:   "Subdomain is already taken",
		}, nil
	}

	return &dto.SubdomainAvailabilityResponse{
		Subdomain: req.Subdomain,
		Available: true,
		Message:   "Subdomain is available",
	}, nil
}

// CheckDomainAvailability checks if a custom domain is available
func (s *tenantService) CheckDomainAvailability(ctx context.Context, req *dto.CheckDomainRequest) (*dto.DomainAvailabilityResponse, error) {
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	existing, err := s.repos.Tenant.FindByDomain(ctx, req.Domain)
	if err == nil && existing != nil {
		return &dto.DomainAvailabilityResponse{
			Domain:    req.Domain,
			Available: false,
			Message:   "Domain is already taken",
		}, nil
	}

	return &dto.DomainAvailabilityResponse{
		Domain:    req.Domain,
		Available: true,
		Message:   "Domain is available",
	}, nil
}

// ValidateTenantLimits validates if tenant can perform an action based on plan limits
func (s *tenantService) ValidateTenantLimits(ctx context.Context, req *dto.ValidateLimitsRequest, tenantID uuid.UUID) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	tenant, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		return ErrTenantNotFound
	}

	switch req.Action {
	case "add_user":
		if tenant.MaxUsers > 0 && tenant.CurrentUsers >= tenant.MaxUsers {
			return fmt.Errorf("%w: user limit reached (%d/%d)", ErrTenantLimitReached, tenant.CurrentUsers, tenant.MaxUsers)
		}
	case "upload_file":
		if tenant.MaxStorage > 0 && tenant.StorageUsed >= tenant.MaxStorage {
			return fmt.Errorf("%w: storage limit reached", ErrTenantLimitReached)
		}
	}

	return nil
}

// ============================================================================
// Statistics and Monitoring
// ============================================================================

// GetTenantStats retrieves statistics for a tenant
func (s *tenantService) GetTenantStats(ctx context.Context, id uuid.UUID) (*dto.TenantStats, error) {
	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	stats := &dto.TenantStats{
		TenantID:          tenant.ID,
		TotalUsers:        tenant.CurrentUsers,
		StorageUsedMB:     tenant.StorageUsed / (1024 * 1024),
		StorageLimitMB:    tenant.MaxStorage / (1024 * 1024),
		StoragePercentage: 0,
	}

	if tenant.MaxStorage > 0 {
		stats.StoragePercentage = (float64(tenant.StorageUsed) / float64(tenant.MaxStorage)) * 100
	}

	return stats, nil
}

// GetTenantDetails retrieves detailed tenant information with stats
func (s *tenantService) GetTenantDetails(ctx context.Context, id uuid.UUID) (*dto.TenantDetailsResponse, error) {
	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	stats, _ := s.GetTenantStats(ctx, id)

	return &dto.TenantDetailsResponse{
		Tenant: dto.ToTenantResponse(tenant),
		Stats:  stats,
	}, nil
}

// GetTenantLimits retrieves current limits and usage for a tenant
func (s *tenantService) GetTenantLimits(ctx context.Context, id uuid.UUID) (*dto.TenantLimitsResponse, error) {
	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	usersRemaining := 0
	if tenant.MaxUsers > 0 {
		usersRemaining = tenant.MaxUsers - tenant.CurrentUsers
		if usersRemaining < 0 {
			usersRemaining = 0
		}
	}

	storageRemaining := int64(0)
	if tenant.MaxStorage > 0 {
		storageRemaining = tenant.MaxStorage - tenant.StorageUsed
		if storageRemaining < 0 {
			storageRemaining = 0
		}
	}

	return &dto.TenantLimitsResponse{
		TenantID:          tenant.ID,
		Plan:              tenant.Plan,
		MaxUsers:          tenant.MaxUsers,
		CurrentUsers:      tenant.CurrentUsers,
		UsersRemaining:    usersRemaining,
		MaxArtisans:       tenant.MaxArtisans,
		CurrentArtisans:   0, // Would need user role count
		ArtisansRemaining: tenant.MaxArtisans,
		MaxStorage:        tenant.MaxStorage,
		StorageUsed:       tenant.StorageUsed,
		StorageRemaining:  storageRemaining,
		CanAddUser:        tenant.MaxUsers <= 0 || tenant.CurrentUsers < tenant.MaxUsers,
		CanAddArtisan:     true, // Simplified
		CanUploadFile:     tenant.MaxStorage <= 0 || tenant.StorageUsed < tenant.MaxStorage,
	}, nil
}

// GetTenantHealth checks the health status of a tenant
func (s *tenantService) GetTenantHealth(ctx context.Context, id uuid.UUID) (*dto.TenantHealthResponse, error) {
	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	health := &dto.TenantHealthResponse{
		TenantID:    tenant.ID,
		Status:      tenant.Status,
		IsHealthy:   true,
		HealthScore: 100,
		LastChecked: time.Now(),
	}

	var issues []string
	var warnings []string

	// Check storage
	if tenant.MaxStorage > 0 {
		usagePercent := float64(tenant.StorageUsed) / float64(tenant.MaxStorage) * 100
		if usagePercent >= 90 {
			issues = append(issues, "Storage usage is critical (>90%)")
			health.StorageHealth = "critical"
		} else if usagePercent >= 75 {
			warnings = append(warnings, "Storage usage is high (>75%)")
			health.StorageHealth = "warning"
		} else {
			health.StorageHealth = "healthy"
		}
	}

	// Check user limits
	if tenant.MaxUsers > 0 {
		usagePercent := float64(tenant.CurrentUsers) / float64(tenant.MaxUsers) * 100
		if usagePercent >= 100 {
			warnings = append(warnings, "User limit reached")
			health.UserLimitHealth = "limit_reached"
		} else if usagePercent >= 80 {
			warnings = append(warnings, "Approaching user limit (>80%)")
			health.UserLimitHealth = "warning"
		} else {
			health.UserLimitHealth = "healthy"
		}
	}

	// Check status
	if tenant.Status == models.TenantStatusSuspended {
		issues = append(issues, "Tenant is suspended")
		health.SubscriptionHealth = "suspended"
	} else if tenant.Status == models.TenantStatusCancelled {
		issues = append(issues, "Tenant is cancelled")
		health.SubscriptionHealth = "cancelled"
	} else {
		health.SubscriptionHealth = "active"
	}

	health.Issues = issues
	health.Warnings = warnings
	health.IsHealthy = len(issues) == 0
	health.HealthScore = 100 - (len(issues) * 25) - (len(warnings) * 10)
	if health.HealthScore < 0 {
		health.HealthScore = 0
	}

	return health, nil
}

// ============================================================================
// Trial Management
// ============================================================================

// GetExpiredTrials retrieves all tenants with expired trials
func (s *tenantService) GetExpiredTrials(ctx context.Context) ([]*dto.TenantResponse, error) {
	tenants, err := s.repos.Tenant.FindExpiredTrials(ctx)
	if err != nil {
		s.logger.Error("failed to find expired trials", zap.Error(err))
		return nil, fmt.Errorf("failed to find expired trials: %w", err)
	}
	return dto.ToTenantResponses(tenants), nil
}

// ExtendTrial extends a tenant's trial period
func (s *tenantService) ExtendTrial(ctx context.Context, id uuid.UUID, days int) error {
	if days <= 0 || days > 90 {
		return fmt.Errorf("trial extension must be between 1 and 90 days")
	}

	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return ErrTenantNotFound
	}

	if tenant.Status != models.TenantStatusTrial {
		return fmt.Errorf("tenant is not in trial status")
	}

	newExpiry := time.Now().AddDate(0, 0, days)
	if tenant.TrialEndsAt != nil && tenant.TrialEndsAt.After(time.Now()) {
		newExpiry = tenant.TrialEndsAt.AddDate(0, 0, days)
	}

	tenant.TrialEndsAt = &newExpiry
	if err := s.repos.Tenant.Update(ctx, tenant); err != nil {
		s.logger.Error("failed to extend trial", zap.Error(err))
		return fmt.Errorf("failed to extend trial: %w", err)
	}

	s.logger.Info("trial extended",
		zap.String("tenant_id", id.String()),
		zap.Int("days", days),
		zap.Time("new_expiry", newExpiry),
	)

	return nil
}

// ============================================================================
// Usage Counter Management
// ============================================================================

// IncrementUserCount increments the user count for a tenant
func (s *tenantService) IncrementUserCount(ctx context.Context, id uuid.UUID) error {
	tenant, err := s.repos.Tenant.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	if tenant.MaxUsers > 0 && tenant.CurrentUsers >= tenant.MaxUsers {
		return ErrTenantLimitReached
	}

	if err := s.repos.Tenant.IncrementUserCount(ctx, id); err != nil {
		s.logger.Error("failed to increment user count", zap.Error(err))
		return fmt.Errorf("failed to increment user count: %w", err)
	}

	return nil
}

// DecrementUserCount decrements the user count for a tenant
func (s *tenantService) DecrementUserCount(ctx context.Context, id uuid.UUID) error {
	if err := s.repos.Tenant.DecrementUserCount(ctx, id); err != nil {
		s.logger.Error("failed to decrement user count", zap.Error(err))
		return fmt.Errorf("failed to decrement user count: %w", err)
	}
	return nil
}

// UpdateStorageUsage updates the storage used by a tenant
func (s *tenantService) UpdateStorageUsage(ctx context.Context, req *dto.UpdateStorageUsageRequest, tenantID uuid.UUID) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	tenant, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	if tenant.MaxStorage > 0 && req.BytesUsed > tenant.MaxStorage {
		return fmt.Errorf("%w: storage limit exceeded", ErrTenantLimitReached)
	}

	if err := s.repos.Tenant.UpdateStorageUsed(ctx, tenantID, req.BytesUsed); err != nil {
		s.logger.Error("failed to update storage usage", zap.Error(err))
		return fmt.Errorf("failed to update storage usage: %w", err)
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// isValidSubdomain validates subdomain format
func isValidSubdomain(subdomain string) bool {
	if len(subdomain) < 3 || len(subdomain) > 63 {
		return false
	}

	// RFC 1123 subdomain validation
	pattern := `^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`
	matched, err := regexp.MatchString(pattern, subdomain)
	if err != nil || !matched {
		return false
	}

	// Reserved subdomains
	reserved := map[string]bool{
		"www": true, "api": true, "admin": true, "app": true,
		"mail": true, "ftp": true, "localhost": true, "dashboard": true,
		"portal": true, "status": true, "help": true, "support": true,
		"blog": true, "cdn": true, "static": true, "assets": true,
	}

	return !reserved[subdomain]
}

// PlanLimits defines limits for a tenant plan
type PlanLimits struct {
	MaxUsers    int
	MaxArtisans int
	MaxStorage  int64
}

// getLimitsForPlan returns the limits for a given plan
func getLimitsForPlan(plan models.TenantPlan) PlanLimits {
	limits := map[models.TenantPlan]PlanLimits{
		models.TenantPlanSolo: {
			MaxUsers:    5,
			MaxArtisans: 1,
			MaxStorage:  1073741824, // 1GB
		},
		models.TenantPlanSmall: {
			MaxUsers:    20,
			MaxArtisans: 5,
			MaxStorage:  5368709120, // 5GB
		},
		models.TenantPlanCorporation: {
			MaxUsers:    100,
			MaxArtisans: 25,
			MaxStorage:  21474836480, // 20GB
		},
		models.TenantPlanEnterprise: {
			MaxUsers:    1000,
			MaxArtisans: 100,
			MaxStorage:  107374182400, // 100GB
		},
	}

	if l, ok := limits[plan]; ok {
		return l
	}
	return limits[models.TenantPlanSolo]
}

// getFeaturesForPlan returns the feature set for a given plan
func getFeaturesForPlan(plan models.TenantPlan) models.TenantFeatures {
	features := models.TenantFeatures{}

	switch plan {
	case models.TenantPlanSolo:
		features.BasicBooking = true
		features.CustomerManagement = true
		features.ServiceCatalog = true
		features.EmailNotifications = true
		features.BasicAnalytics = true

	case models.TenantPlanSmall:
		features.BasicBooking = true
		features.AdvancedBooking = true
		features.CustomerManagement = true
		features.ServiceCatalog = true
		features.EmailNotifications = true
		features.SMSNotifications = true
		features.BasicProjects = true
		features.TeamMembers = true
		features.BasicAnalytics = true
		features.OnlinePayments = true
		features.InvoiceGeneration = true

	case models.TenantPlanCorporation:
		features.BasicBooking = true
		features.AdvancedBooking = true
		features.CustomerManagement = true
		features.ServiceCatalog = true
		features.EmailNotifications = true
		features.SMSNotifications = true
		features.BasicProjects = true
		features.TeamMembers = true
		features.BasicAnalytics = true
		features.AdvancedAnalytics = true
		features.OnlinePayments = true
		features.InvoiceGeneration = true
		features.APIAccess = true
		features.WebhookSupport = true
		features.CustomBranding = true
		features.DataExport = true

	case models.TenantPlanEnterprise:
		features.BasicBooking = true
		features.AdvancedBooking = true
		features.CustomerManagement = true
		features.ServiceCatalog = true
		features.EmailNotifications = true
		features.SMSNotifications = true
		features.BasicProjects = true
		features.TeamMembers = true
		features.BasicAnalytics = true
		features.AdvancedAnalytics = true
		features.OnlinePayments = true
		features.InvoiceGeneration = true
		features.APIAccess = true
		features.WebhookSupport = true
		features.CustomBranding = true
		features.WhiteLabeling = true
		features.DataExport = true
		features.SSO = true
		features.AIRecommendations = true
		features.PrioritySupport = true
		features.SLAGuarantees = true
	}

	return features
}
