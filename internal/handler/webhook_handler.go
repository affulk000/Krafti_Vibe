package handler

import (
	"strconv"
	"time"

	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// WebhookHandler handles HTTP requests for webhook operations
type WebhookHandler struct {
	webhookService service.WebhookRepository
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService service.WebhookRepository) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// ============================================================================
// Webhook Event Management
// ============================================================================

// CreateWebhookEvent godoc
// @Summary Create a webhook event
// @Description Create a new webhook event for delivery
// @Tags webhooks
// @Accept json
// @Produce json
// @Param webhook body dto.CreateWebhookEventRequest true "Webhook event data"
// @Success 201 {object} dto.WebhookEventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks [post]
func (h *WebhookHandler) CreateWebhookEvent(c *fiber.Ctx) error {
	var req dto.CreateWebhookEventRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	event, err := h.webhookService.CreateWebhookEvent(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, event, "Webhook event created successfully")
}

// GetWebhookEvent godoc
// @Summary Get webhook event by ID
// @Description Get detailed webhook event information by ID
// @Tags webhooks
// @Produce json
// @Param id path string true "Webhook Event ID"
// @Success 200 {object} dto.WebhookEventResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/{id} [get]
func (h *WebhookHandler) GetWebhookEvent(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid webhook event ID", err)
	}

	event, err := h.webhookService.GetWebhookEvent(c.Context(), eventID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, event)
}

// ListWebhookEvents godoc
// @Summary List webhook events
// @Description List webhook events with filtering and pagination
// @Tags webhooks
// @Accept json
// @Produce json
// @Param filter body dto.WebhookEventFilter true "Filter parameters"
// @Success 200 {object} dto.WebhookEventListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/list [post]
func (h *WebhookHandler) ListWebhookEvents(c *fiber.Ctx) error {
	var filter dto.WebhookEventFilter
	if err := c.BodyParser(&filter); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid filter parameters", err)
	}

	events, err := h.webhookService.ListWebhookEvents(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, events)
}

// ============================================================================
// Delivery Operations
// ============================================================================

// DeliverWebhook godoc
// @Summary Deliver a webhook
// @Description Attempt to deliver a webhook event
// @Tags webhooks
// @Produce json
// @Param id path string true "Webhook Event ID"
// @Success 200 {object} dto.WebhookDeliveryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/{id}/deliver [post]
func (h *WebhookHandler) DeliverWebhook(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid webhook event ID", err)
	}

	delivery, err := h.webhookService.DeliverWebhook(c.Context(), eventID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, delivery)
}

// RetryWebhook godoc
// @Summary Retry a failed webhook
// @Description Retry delivery of a failed webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook Event ID"
// @Param request body dto.RetryWebhookRequest true "Retry options"
// @Success 200 {object} dto.WebhookDeliveryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/{id}/retry [post]
func (h *WebhookHandler) RetryWebhook(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid webhook event ID", err)
	}

	var req dto.RetryWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}
	req.WebhookEventID = eventID

	delivery, err := h.webhookService.RetryWebhook(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, delivery)
}

// RetryFailedWebhooks godoc
// @Summary Retry failed webhooks for a tenant
// @Description Retry all failed webhooks for a specific tenant
// @Tags webhooks
// @Produce json
// @Param tenantId path string true "Tenant ID"
// @Param limit query int false "Max webhooks to retry" default(10)
// @Success 200 {object} dto.WebhookRetryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/tenant/{tenantId}/retry-failed [post]
func (h *WebhookHandler) RetryFailedWebhooks(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	result, err := h.webhookService.RetryFailedWebhooks(c.Context(), tenantID, limit)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, result)
}

// BulkRetryWebhooks godoc
// @Summary Bulk retry webhooks
// @Description Retry multiple webhooks based on criteria
// @Tags webhooks
// @Accept json
// @Produce json
// @Param request body dto.BulkRetryRequest true "Bulk retry criteria"
// @Success 200 {object} dto.WebhookRetryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/bulk-retry [post]
func (h *WebhookHandler) BulkRetryWebhooks(c *fiber.Ctx) error {
	var req dto.BulkRetryRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	result, err := h.webhookService.BulkRetryWebhooks(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, result)
}

// ============================================================================
// Query Operations
// ============================================================================

// GetPendingWebhooks godoc
// @Summary Get pending webhooks
// @Description Get webhooks pending retry
// @Tags webhooks
// @Produce json
// @Param tenantId query string false "Tenant ID filter"
// @Param limit query int false "Max results" default(50)
// @Success 200 {array} dto.WebhookEventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/pending [get]
func (h *WebhookHandler) GetPendingWebhooks(c *fiber.Ctx) error {
	var tenantID uuid.UUID
	if tenantIDStr := c.Query("tenantId"); tenantIDStr != "" {
		var err error
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	webhooks, err := h.webhookService.GetPendingWebhooks(c.Context(), tenantID, limit)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, webhooks)
}

// GetFailedWebhooks godoc
// @Summary Get failed webhooks
// @Description Get webhooks that have failed delivery
// @Tags webhooks
// @Produce json
// @Param tenantId query string true "Tenant ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} dto.WebhookEventListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/failed [get]
func (h *WebhookHandler) GetFailedWebhooks(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenantId"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	webhooks, err := h.webhookService.GetFailedWebhooks(c.Context(), tenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, webhooks)
}

// GetDeliveredWebhooks godoc
// @Summary Get delivered webhooks
// @Description Get webhooks that have been successfully delivered
// @Tags webhooks
// @Produce json
// @Param tenantId query string true "Tenant ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} dto.WebhookEventListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/delivered [get]
func (h *WebhookHandler) GetDeliveredWebhooks(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenantId"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	webhooks, err := h.webhookService.GetDeliveredWebhooks(c.Context(), tenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, webhooks)
}

// GetRecentWebhooks godoc
// @Summary Get recent webhooks
// @Description Get recent webhooks within specified hours
// @Tags webhooks
// @Produce json
// @Param tenantId query string true "Tenant ID"
// @Param hours query int false "Hours to look back" default(24)
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} dto.WebhookEventListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/recent [get]
func (h *WebhookHandler) GetRecentWebhooks(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenantId"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	hours := 24
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	webhooks, err := h.webhookService.GetRecentWebhooks(c.Context(), tenantID, hours, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, webhooks)
}

// ============================================================================
// Analytics
// ============================================================================

// GetWebhookStats godoc
// @Summary Get webhook statistics
// @Description Get webhook delivery statistics for a tenant
// @Tags webhooks
// @Produce json
// @Param tenantId query string true "Tenant ID"
// @Param startDate query string false "Start date (RFC3339)"
// @Param endDate query string false "End date (RFC3339)"
// @Success 200 {object} dto.WebhookStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/stats [get]
func (h *WebhookHandler) GetWebhookStats(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenantId"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var startDate, endDate time.Time
	if startStr := c.Query("startDate"); startStr != "" {
		startDate, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid start date format", err)
		}
	}

	if endStr := c.Query("endDate"); endStr != "" {
		endDate, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid end date format", err)
		}
	}

	stats, err := h.webhookService.GetWebhookStats(c.Context(), tenantID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetWebhookAnalytics godoc
// @Summary Get webhook analytics
// @Description Get comprehensive webhook analytics for a tenant
// @Tags webhooks
// @Produce json
// @Param tenantId query string true "Tenant ID"
// @Param startDate query string false "Start date (RFC3339)"
// @Param endDate query string false "End date (RFC3339)"
// @Success 200 {object} dto.WebhookAnalyticsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/analytics [get]
func (h *WebhookHandler) GetWebhookAnalytics(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenantId"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	var startDate, endDate time.Time
	if startStr := c.Query("startDate"); startStr != "" {
		startDate, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid start date format", err)
		}
	}

	if endStr := c.Query("endDate"); endStr != "" {
		endDate, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid end date format", err)
		}
	}

	analytics, err := h.webhookService.GetWebhookAnalytics(c.Context(), tenantID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, analytics)
}

// GetFailureReasons godoc
// @Summary Get webhook failure reasons
// @Description Get top failure reasons for webhook delivery
// @Tags webhooks
// @Produce json
// @Param tenantId query string true "Tenant ID"
// @Param limit query int false "Max results" default(10)
// @Success 200 {array} dto.WebhookFailureReasonResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/failure-reasons [get]
func (h *WebhookHandler) GetFailureReasons(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenantId"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid tenant ID", err)
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	reasons, err := h.webhookService.GetFailureReasons(c.Context(), tenantID, limit)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, reasons)
}

// ============================================================================
// Cleanup Operations
// ============================================================================

// CleanupOldWebhooks godoc
// @Summary Cleanup old webhooks
// @Description Remove old webhook events from the database
// @Tags webhooks
// @Produce json
// @Param olderThanDays query int false "Delete webhooks older than this many days" default(90)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/cleanup/old [post]
func (h *WebhookHandler) CleanupOldWebhooks(c *fiber.Ctx) error {
	olderThanDays := 90
	if daysStr := c.Query("olderThanDays"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			olderThanDays = d
		}
	}

	count, err := h.webhookService.CleanupOldWebhooks(c.Context(), olderThanDays)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"deleted_count":   count,
		"older_than_days": olderThanDays,
		"message":         "Old webhooks cleaned up successfully",
	})
}

// CleanupDeliveredWebhooks godoc
// @Summary Cleanup delivered webhooks
// @Description Remove old delivered webhook events
// @Tags webhooks
// @Produce json
// @Param olderThanDays query int false "Delete webhooks older than this many days" default(30)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/cleanup/delivered [post]
func (h *WebhookHandler) CleanupDeliveredWebhooks(c *fiber.Ctx) error {
	olderThanDays := 30
	if daysStr := c.Query("olderThanDays"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			olderThanDays = d
		}
	}

	count, err := h.webhookService.CleanupDeliveredWebhooks(c.Context(), olderThanDays)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"deleted_count":   count,
		"older_than_days": olderThanDays,
		"message":         "Delivered webhooks cleaned up successfully",
	})
}

// PurgeFailedWebhooks godoc
// @Summary Purge failed webhooks
// @Description Remove permanently failed webhooks
// @Tags webhooks
// @Produce json
// @Param maxAttempts query int false "Max attempts threshold" default(3)
// @Param olderThanDays query int false "Delete webhooks older than this many days" default(7)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/cleanup/purge [post]
func (h *WebhookHandler) PurgeFailedWebhooks(c *fiber.Ctx) error {
	maxAttempts := 3
	if attemptsStr := c.Query("maxAttempts"); attemptsStr != "" {
		if a, err := strconv.Atoi(attemptsStr); err == nil && a > 0 {
			maxAttempts = a
		}
	}

	olderThanDays := 7
	if daysStr := c.Query("olderThanDays"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			olderThanDays = d
		}
	}

	count, err := h.webhookService.PurgeFailedWebhooks(c.Context(), maxAttempts, olderThanDays)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"deleted_count":   count,
		"max_attempts":    maxAttempts,
		"older_than_days": olderThanDays,
		"message":         "Failed webhooks purged successfully",
	})
}

// ============================================================================
// Background Processing
// ============================================================================

// ProcessPendingWebhooks godoc
// @Summary Process pending webhooks
// @Description Process pending webhooks in batch
// @Tags webhooks
// @Produce json
// @Param batchSize query int false "Batch size" default(100)
// @Success 200 {object} dto.WebhookRetryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/process-pending [post]
func (h *WebhookHandler) ProcessPendingWebhooks(c *fiber.Ctx) error {
	batchSize := 100
	if batchStr := c.Query("batchSize"); batchStr != "" {
		if b, err := strconv.Atoi(batchStr); err == nil && b > 0 {
			batchSize = b
		}
	}

	result, err := h.webhookService.ProcessPendingWebhooks(c.Context(), batchSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, result)
}

// ============================================================================
// Health & Monitoring
// ============================================================================

// HealthCheck godoc
// @Summary Webhook service health check
// @Description Check if webhook service is healthy
// @Tags webhooks
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/health [get]
func (h *WebhookHandler) HealthCheck(c *fiber.Ctx) error {
	if err := h.webhookService.HealthCheck(c.Context()); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"status":  "healthy",
		"service": "webhook",
	})
}

// GetServiceMetrics godoc
// @Summary Get webhook service metrics
// @Description Get comprehensive metrics for webhook service
// @Tags webhooks
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/metrics [get]
func (h *WebhookHandler) GetServiceMetrics(c *fiber.Ctx) error {
	metrics := h.webhookService.GetServiceMetrics(c.Context())

	return NewSuccessResponse(c, metrics)
}
