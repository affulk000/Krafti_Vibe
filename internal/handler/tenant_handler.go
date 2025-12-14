package handler

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantHandler handles HTTP requests for tenant operations
type TenantHandler struct {
	tenantService service.TenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService service.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// ============================================================================
// Core CRUD Operations
// ============================================================================

// CreateTenant godoc
// @Summary Create a new tenant
// @Description Create a new tenant organization
// @Tags tenants
// @Accept json
// @Produce json
// @Param tenant body dto.CreateTenantRequest true "Tenant creation data"
// @Success 201 {object} dto.TenantResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants [post]
func (h *TenantHandler) CreateTenant(c *fiber.Ctx) error {
	var req dto.CreateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	tenant, err := h.tenantService.CreateTenant(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, tenant, "Tenant created successfully")
}

// GetTenant godoc
// @Summary Get tenant by ID
// @Description Get detailed tenant information by ID
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} dto.TenantResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id} [get]
func (h *TenantHandler) GetTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	tenant, err := h.tenantService.GetTenant(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, tenant)
}

// GetTenantBySubdomain godoc
// @Summary Get tenant by subdomain
// @Description Get tenant information by subdomain
// @Tags tenants
// @Produce json
// @Param subdomain path string true "Subdomain"
// @Success 200 {object} dto.TenantResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/subdomain/{subdomain} [get]
func (h *TenantHandler) GetTenantBySubdomain(c *fiber.Ctx) error {
	subdomain := c.Params("subdomain")
	if subdomain == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_SUBDOMAIN", "Subdomain is required", nil)
	}

	tenant, err := h.tenantService.GetTenantBySubdomain(c.Context(), subdomain)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, tenant)
}

// GetTenantByDomain godoc
// @Summary Get tenant by domain
// @Description Get tenant information by custom domain
// @Tags tenants
// @Produce json
// @Param domain path string true "Domain"
// @Success 200 {object} dto.TenantResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/domain/{domain} [get]
func (h *TenantHandler) GetTenantByDomain(c *fiber.Ctx) error {
	domain := c.Params("domain")
	if domain == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_DOMAIN", "Domain is required", nil)
	}

	tenant, err := h.tenantService.GetTenantByDomain(c.Context(), domain)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, tenant)
}

// UpdateTenant godoc
// @Summary Update tenant
// @Description Update tenant information
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param tenant body dto.UpdateTenantRequest true "Update data"
// @Success 200 {object} dto.TenantResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req dto.UpdateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	tenant, err := h.tenantService.UpdateTenant(c.Context(), tenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, tenant, "Tenant updated successfully")
}

// DeleteTenant godoc
// @Summary Delete tenant
// @Description Delete a tenant
// @Tags tenants
// @Param id path string true "Tenant ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id} [delete]
func (h *TenantHandler) DeleteTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	if err := h.tenantService.DeleteTenant(c.Context(), tenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// ============================================================================
// Listing and Search
// ============================================================================

// ListTenants godoc
// @Summary List tenants
// @Description Get a paginated list of tenants with filters
// @Tags tenants
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param plan query string false "Filter by plan"
// @Success 200 {object} dto.TenantListResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants [get]
func (h *TenantHandler) ListTenants(c *fiber.Ctx) error {
	filter := &dto.TenantFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := models.TenantStatus(statusStr)
		filter.Status = &status
	}

	if planStr := c.Query("plan"); planStr != "" {
		plan := models.TenantPlan(planStr)
		filter.Plan = &plan
	}

	tenants, err := h.tenantService.ListTenants(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, tenants)
}

// SearchTenants godoc
// @Summary Search tenants
// @Description Search for tenants
// @Tags tenants
// @Accept json
// @Produce json
// @Param search body dto.SearchTenantsRequest true "Search criteria"
// @Success 200 {object} dto.TenantListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/search [post]
func (h *TenantHandler) SearchTenants(c *fiber.Ctx) error {
	var req dto.SearchTenantsRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	tenants, err := h.tenantService.SearchTenants(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, tenants)
}

// ============================================================================
// Tenant Status Management
// ============================================================================

// ActivateTenant godoc
// @Summary Activate tenant
// @Description Activate a tenant account
// @Tags tenants
// @Param id path string true "Tenant ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/activate [post]
func (h *TenantHandler) ActivateTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	if err := h.tenantService.ActivateTenant(c.Context(), tenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Tenant activated successfully")
}

// SuspendTenant godoc
// @Summary Suspend tenant
// @Description Suspend a tenant account
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param suspend body dto.SuspendTenantRequest true "Suspension data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/suspend [post]
func (h *TenantHandler) SuspendTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req dto.SuspendTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.tenantService.SuspendTenant(c.Context(), &req, tenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Tenant suspended successfully")
}

// CancelTenant godoc
// @Summary Cancel tenant
// @Description Cancel a tenant subscription
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param cancel body dto.CancelTenantRequest true "Cancellation data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/cancel [post]
func (h *TenantHandler) CancelTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req dto.CancelTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.tenantService.CancelTenant(c.Context(), &req, tenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Tenant cancelled successfully")
}

// ============================================================================
// Plan and Settings Management
// ============================================================================

// UpdateTenantPlan godoc
// @Summary Update tenant plan
// @Description Update tenant subscription plan
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param plan body dto.UpdateTenantPlanRequest true "Plan data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/plan [put]
func (h *TenantHandler) UpdateTenantPlan(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req dto.UpdateTenantPlanRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.tenantService.UpdateTenantPlan(c.Context(), tenantID, &req); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Tenant plan updated successfully")
}

// UpdateTenantSettings godoc
// @Summary Update tenant settings
// @Description Update tenant settings
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param settings body dto.UpdateTenantSettingsRequest true "Settings data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/settings [put]
func (h *TenantHandler) UpdateTenantSettings(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req dto.UpdateTenantSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.tenantService.UpdateTenantSettings(c.Context(), tenantID, &req); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Tenant settings updated successfully")
}

// UpdateTenantFeatures godoc
// @Summary Update tenant features
// @Description Update tenant feature flags
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param features body dto.UpdateTenantFeaturesRequest true "Features data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/features [put]
func (h *TenantHandler) UpdateTenantFeatures(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req dto.UpdateTenantFeaturesRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.tenantService.UpdateTenantFeatures(c.Context(), tenantID, &req); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Tenant features updated successfully")
}

// ============================================================================
// Validation and Availability
// ============================================================================

// CheckSubdomainAvailability godoc
// @Summary Check subdomain availability
// @Description Check if a subdomain is available
// @Tags tenants
// @Accept json
// @Produce json
// @Param subdomain body dto.CheckSubdomainRequest true "Subdomain to check"
// @Success 200 {object} dto.SubdomainAvailabilityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/check-subdomain [post]
func (h *TenantHandler) CheckSubdomainAvailability(c *fiber.Ctx) error {
	var req dto.CheckSubdomainRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	availability, err := h.tenantService.CheckSubdomainAvailability(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, availability)
}

// CheckDomainAvailability godoc
// @Summary Check domain availability
// @Description Check if a domain is available
// @Tags tenants
// @Accept json
// @Produce json
// @Param domain body dto.CheckDomainRequest true "Domain to check"
// @Success 200 {object} dto.DomainAvailabilityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/check-domain [post]
func (h *TenantHandler) CheckDomainAvailability(c *fiber.Ctx) error {
	var req dto.CheckDomainRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	availability, err := h.tenantService.CheckDomainAvailability(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, availability)
}

// ============================================================================
// Statistics and Monitoring
// ============================================================================

// GetTenantStats godoc
// @Summary Get tenant statistics
// @Description Get comprehensive tenant statistics
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} dto.TenantStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/stats [get]
func (h *TenantHandler) GetTenantStats(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	stats, err := h.tenantService.GetTenantStats(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetTenantDetails godoc
// @Summary Get tenant details
// @Description Get detailed tenant information
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} dto.TenantDetailsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/details [get]
func (h *TenantHandler) GetTenantDetails(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	details, err := h.tenantService.GetTenantDetails(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, details)
}

// GetTenantHealth godoc
// @Summary Get tenant health
// @Description Get tenant health metrics
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} dto.TenantHealthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/health [get]
func (h *TenantHandler) GetTenantHealth(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	health, err := h.tenantService.GetTenantHealth(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, health)
}

// GetTenantLimits godoc
// @Summary Get tenant limits
// @Description Get tenant plan limits and usage
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} dto.TenantLimitsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/limits [get]
func (h *TenantHandler) GetTenantLimits(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	limits, err := h.tenantService.GetTenantLimits(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, limits)
}

// ============================================================================
// Trial Management
// ============================================================================

// ExtendTrial godoc
// @Summary Extend tenant trial
// @Description Extend tenant trial period
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param extend body ExtendTrialRequest true "Extension data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{id}/extend-trial [post]
func (h *TenantHandler) ExtendTrial(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req ExtendTrialRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.tenantService.ExtendTrial(c.Context(), tenantID, req.Days); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Trial extended successfully")
}

// GetExpiredTrials godoc
// @Summary Get expired trials
// @Description Get tenants with expired trials
// @Tags tenants
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/expired-trials [get]
func (h *TenantHandler) GetExpiredTrials(c *fiber.Ctx) error {
	tenants, err := h.tenantService.GetExpiredTrials(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, tenants)
}

// ============================================================================
// Request Types
// ============================================================================

type ExtendTrialRequest struct {
	Days int `json:"days"`
}
