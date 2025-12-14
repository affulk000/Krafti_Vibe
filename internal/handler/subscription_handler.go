package handler

import (
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SubscriptionHandler handles HTTP requests for subscription operations
type SubscriptionHandler struct {
	subscriptionService service.SubscriptionService
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(subscriptionService service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// CreateSubscription creates a new subscription
func (h *SubscriptionHandler) CreateSubscription(c *fiber.Ctx) error {
	var req dto.CreateSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	subscription, err := h.subscriptionService.CreateSubscription(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, subscription, "Subscription created successfully")
}

// GetSubscription retrieves a subscription by tenant ID
func (h *SubscriptionHandler) GetSubscription(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	subscription, err := h.subscriptionService.GetSubscription(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, subscription)
}

// UpdateSubscription updates a subscription
func (h *SubscriptionHandler) UpdateSubscription(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req dto.UpdateSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	subscription, err := h.subscriptionService.UpdateSubscription(c.Context(), tenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, subscription, "Subscription updated successfully")
}

// CancelSubscription cancels a subscription
func (h *SubscriptionHandler) CancelSubscription(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var req dto.CancelSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	subscription, err := h.subscriptionService.CancelSubscription(c.Context(), tenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, subscription, "Subscription cancelled successfully")
}

// ListSubscriptions retrieves all subscriptions with pagination
func (h *SubscriptionHandler) ListSubscriptions(c *fiber.Ctx) error {
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	filter := dto.SubscriptionFilter{
		Page:     page,
		PageSize: pageSize,
	}

	subscriptions, err := h.subscriptionService.GetSubscriptionList(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, subscriptions)
}
