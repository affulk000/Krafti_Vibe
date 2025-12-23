package handler

import (
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantInvitationHandler handles tenant invitation HTTP requests
type TenantInvitationHandler struct {
	service service.TenantInvitationService
}

// NewTenantInvitationHandler creates a new tenant invitation handler
func NewTenantInvitationHandler(service service.TenantInvitationService) *TenantInvitationHandler {
	return &TenantInvitationHandler{
		service: service,
	}
}

// CreateInvitation creates a new tenant invitation
// @Summary Create tenant invitation
// @Tags Tenant Invitations
// @Accept json
// @Produce json
// @Param invitation body dto.CreateInvitationRequest true "Invitation details"
// @Success 201 {object} dto.InvitationResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 401 {object} handler.ErrorResponse
// @Router /api/v1/invitations [post]
func (h *TenantInvitationHandler) CreateInvitation(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateInvitationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	// Validate and sanitize
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
	}

	invitation, err := h.service.CreateInvitation(c.Context(), &req, authCtx.TenantID, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(invitation)
}

// GetInvitation retrieves an invitation by ID
// @Summary Get invitation
// @Tags Tenant Invitations
// @Produce json
// @Param id path string true "Invitation ID"
// @Success 200 {object} dto.InvitationResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/invitations/{id} [get]
func (h *TenantInvitationHandler) GetInvitation(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid invitation ID",
			"code":  "INVALID_INVITATION_ID",
		})
	}

	invitation, err := h.service.GetInvitation(c.Context(), id)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(invitation)
}

// GetInvitationByToken retrieves an invitation by token
// @Summary Get invitation by token
// @Tags Tenant Invitations
// @Produce json
// @Param token path string true "Invitation token"
// @Success 200 {object} dto.InvitationResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/invitations/token/{token} [get]
func (h *TenantInvitationHandler) GetInvitationByToken(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invitation token is required",
			"code":  "MISSING_TOKEN",
		})
	}

	invitation, err := h.service.GetInvitationByToken(c.Context(), token)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(invitation)
}

// ListInvitations lists all invitations for a tenant
// @Summary List invitations
// @Tags Tenant Invitations
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status (pending, accepted, expired)"
// @Success 200 {object} dto.InvitationListResponse
// @Router /api/v1/invitations [get]
func (h *TenantInvitationHandler) ListInvitations(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 20)

	filter := &dto.InvitationFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	invitations, err := h.service.ListInvitations(c.Context(), authCtx.TenantID, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(invitations)
}

// GetPendingInvitations retrieves all pending invitations
// @Summary Get pending invitations
// @Tags Tenant Invitations
// @Produce json
// @Success 200 {object} []dto.InvitationResponse
// @Router /api/v1/invitations/pending [get]
func (h *TenantInvitationHandler) GetPendingInvitations(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	invitations, err := h.service.GetPendingInvitations(c.Context(), authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(invitations)
}

// GetInvitationsByEmail retrieves invitations for a specific email
// @Summary Get invitations by email
// @Tags Tenant Invitations
// @Produce json
// @Param email query string true "Email address"
// @Success 200 {object} []dto.InvitationResponse
// @Router /api/v1/invitations/by-email [get]
func (h *TenantInvitationHandler) GetInvitationsByEmail(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email is required",
			"code":  "MISSING_EMAIL",
		})
	}

	invitations, err := h.service.GetInvitationsByEmail(c.Context(), email)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(invitations)
}

// AcceptInvitation accepts an invitation
// @Summary Accept invitation
// @Tags Tenant Invitations
// @Accept json
// @Produce json
// @Param request body dto.AcceptInvitationRequest true "Accept invitation request"
// @Success 200 {object} dto.AcceptInvitationResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /api/v1/invitations/accept [post]
func (h *TenantInvitationHandler) AcceptInvitation(c *fiber.Ctx) error {
	var req dto.AcceptInvitationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
	}

	response, err := h.service.AcceptInvitation(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(response)
}

// RevokeInvitation revokes an invitation
// @Summary Revoke invitation
// @Tags Tenant Invitations
// @Produce json
// @Param id path string true "Invitation ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/invitations/{id}/revoke [post]
func (h *TenantInvitationHandler) RevokeInvitation(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid invitation ID",
			"code":  "INVALID_INVITATION_ID",
		})
	}

	if err := h.service.RevokeInvitation(c.Context(), id, authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "invitation revoked successfully",
	})
}

// ResendInvitation resends an invitation
// @Summary Resend invitation
// @Tags Tenant Invitations
// @Produce json
// @Param id path string true "Invitation ID"
// @Success 200 {object} dto.InvitationResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/invitations/{id}/resend [post]
func (h *TenantInvitationHandler) ResendInvitation(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid invitation ID",
			"code":  "INVALID_INVITATION_ID",
		})
	}

	invitation, err := h.service.ResendInvitation(c.Context(), id, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(invitation)
}

// DeleteExpiredInvitations deletes all expired invitations
// @Summary Delete expired invitations
// @Tags Tenant Invitations
// @Produce json
// @Success 200 {object} handler.SuccessResponse
// @Router /api/v1/invitations/cleanup/expired [delete]
func (h *TenantInvitationHandler) DeleteExpiredInvitations(c *fiber.Ctx) error {
	count, err := h.service.DeleteExpiredInvitations(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "expired invitations deleted successfully",
		"count":   count,
	})
}
