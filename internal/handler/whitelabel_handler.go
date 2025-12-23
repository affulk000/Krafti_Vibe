package handler

import (
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// WhiteLabelHandler handles whitelabel HTTP requests
type WhiteLabelHandler struct {
	service service.WhiteLabelService
}

// NewWhiteLabelHandler creates a new whitelabel handler
func NewWhiteLabelHandler(service service.WhiteLabelService) *WhiteLabelHandler {
	return &WhiteLabelHandler{
		service: service,
	}
}

// CreateWhiteLabel creates a new whitelabel configuration
// @Summary Create whitelabel configuration
// @Tags WhiteLabel
// @Accept json
// @Produce json
// @Param whitelabel body dto.CreateWhiteLabelRequest true "WhiteLabel details"
// @Success 201 {object} dto.WhiteLabelResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 401 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel [post]
func (h *WhiteLabelHandler) CreateWhiteLabel(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateWhiteLabelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	whitelabel, err := h.service.CreateWhiteLabel(c.Context(), authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(whitelabel)
}

// GetWhiteLabel retrieves a whitelabel configuration by ID
// @Summary Get whitelabel configuration
// @Tags WhiteLabel
// @Produce json
// @Param id path string true "WhiteLabel ID"
// @Success 200 {object} dto.WhiteLabelResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/{id} [get]
func (h *WhiteLabelHandler) GetWhiteLabel(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid whitelabel ID",
			"code":  "INVALID_WHITELABEL_ID",
		})
	}

	whitelabel, err := h.service.GetWhiteLabel(c.Context(), id, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(whitelabel)
}

// GetMyWhiteLabel retrieves the current tenant's whitelabel configuration
// @Summary Get my whitelabel configuration
// @Tags WhiteLabel
// @Produce json
// @Success 200 {object} dto.WhiteLabelResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/me [get]
func (h *WhiteLabelHandler) GetMyWhiteLabel(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	whitelabel, err := h.service.GetWhiteLabelByTenant(c.Context(), authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(whitelabel)
}

// GetPublicWhiteLabel retrieves public whitelabel configuration
// @Summary Get public whitelabel configuration
// @Tags WhiteLabel
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} dto.PublicWhiteLabelResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/public [get]
func (h *WhiteLabelHandler) GetPublicWhiteLabel(c *fiber.Ctx) error {
	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tenant_id is required",
			"code":  "MISSING_TENANT_ID",
		})
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid tenant_id",
			"code":  "INVALID_TENANT_ID",
		})
	}

	whitelabel, err := h.service.GetPublicWhiteLabel(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(whitelabel)
}

// GetPublicWhiteLabelByDomain retrieves public whitelabel by custom domain
// @Summary Get public whitelabel by domain
// @Tags WhiteLabel
// @Produce json
// @Param domain query string true "Custom domain"
// @Success 200 {object} dto.PublicWhiteLabelResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/domain [get]
func (h *WhiteLabelHandler) GetPublicWhiteLabelByDomain(c *fiber.Ctx) error {
	domain := c.Query("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "domain is required",
			"code":  "MISSING_DOMAIN",
		})
	}

	whitelabel, err := h.service.GetPublicWhiteLabelByDomain(c.Context(), domain)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(whitelabel)
}

// UpdateWhiteLabel updates a whitelabel configuration
// @Summary Update whitelabel configuration
// @Tags WhiteLabel
// @Accept json
// @Produce json
// @Param id path string true "WhiteLabel ID"
// @Param whitelabel body dto.UpdateWhiteLabelRequest true "Updated whitelabel details"
// @Success 200 {object} dto.WhiteLabelResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/{id} [put]
func (h *WhiteLabelHandler) UpdateWhiteLabel(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid whitelabel ID",
			"code":  "INVALID_WHITELABEL_ID",
		})
	}

	var req dto.UpdateWhiteLabelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	whitelabel, err := h.service.UpdateWhiteLabel(c.Context(), id, authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(whitelabel)
}

// DeleteWhiteLabel deletes a whitelabel configuration
// @Summary Delete whitelabel configuration
// @Tags WhiteLabel
// @Produce json
// @Param id path string true "WhiteLabel ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/{id} [delete]
func (h *WhiteLabelHandler) DeleteWhiteLabel(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid whitelabel ID",
			"code":  "INVALID_WHITELABEL_ID",
		})
	}

	if err := h.service.DeleteWhiteLabel(c.Context(), id, authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "whitelabel deleted successfully",
	})
}

// UpdateColorScheme updates only the color scheme
// @Summary Update color scheme
// @Tags WhiteLabel
// @Accept json
// @Produce json
// @Param colors body dto.UpdateColorSchemeRequest true "Color scheme"
// @Success 200 {object} dto.WhiteLabelResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/colors [put]
func (h *WhiteLabelHandler) UpdateColorScheme(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.UpdateColorSchemeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	whitelabel, err := h.service.UpdateColorScheme(c.Context(), authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(whitelabel)
}

// UpdateBranding updates only branding assets
// @Summary Update branding assets
// @Tags WhiteLabel
// @Accept json
// @Produce json
// @Param branding body dto.UpdateBrandingRequest true "Branding assets"
// @Success 200 {object} dto.WhiteLabelResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/branding [put]
func (h *WhiteLabelHandler) UpdateBranding(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.UpdateBrandingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	whitelabel, err := h.service.UpdateBranding(c.Context(), authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(whitelabel)
}

// UpdateDomain updates custom domain configuration
// @Summary Update custom domain
// @Tags WhiteLabel
// @Accept json
// @Produce json
// @Param domain body dto.UpdateDomainRequest true "Domain configuration"
// @Success 200 {object} dto.WhiteLabelResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/domain [put]
func (h *WhiteLabelHandler) UpdateDomain(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.UpdateDomainRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	whitelabel, err := h.service.UpdateDomain(c.Context(), authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(whitelabel)
}

// ActivateWhiteLabel activates whitelabel configuration
// @Summary Activate whitelabel
// @Tags WhiteLabel
// @Produce json
// @Success 200 {object} handler.SuccessResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/activate [post]
func (h *WhiteLabelHandler) ActivateWhiteLabel(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	if err := h.service.ActivateWhiteLabel(c.Context(), authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "whitelabel activated successfully",
	})
}

// DeactivateWhiteLabel deactivates whitelabel configuration
// @Summary Deactivate whitelabel
// @Tags WhiteLabel
// @Produce json
// @Success 200 {object} handler.SuccessResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/deactivate [post]
func (h *WhiteLabelHandler) DeactivateWhiteLabel(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	if err := h.service.DeactivateWhiteLabel(c.Context(), authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "whitelabel deactivated successfully",
	})
}

// CheckDomainAvailability checks if a domain is available
// @Summary Check domain availability
// @Tags WhiteLabel
// @Produce json
// @Param domain query string true "Domain to check"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /api/v1/whitelabel/check-domain [get]
func (h *WhiteLabelHandler) CheckDomainAvailability(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	domain := c.Query("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "domain is required",
			"code":  "MISSING_DOMAIN",
		})
	}

	available, err := h.service.CheckDomainAvailability(c.Context(), domain, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"available": available,
		"domain":    domain,
	})
}
