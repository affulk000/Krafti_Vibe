package handler

import (
	"time"

	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantUsageHandler handles HTTP requests for tenant usage tracking
type TenantUsageHandler struct {
	usageService service.TenantUsageService
}

// NewTenantUsageHandler creates a new tenant usage handler
func NewTenantUsageHandler(usageService service.TenantUsageService) *TenantUsageHandler {
	return &TenantUsageHandler{
		usageService: usageService,
	}
}

// ============================================================================
// Usage Query Operations
// ============================================================================

// GetDailyUsage godoc
// @Summary Get daily usage
// @Description Get usage statistics for a specific day
// @Tags tenant-usage
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} dto.UsageTrackingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{tenant_id}/usage/daily [get]
func (h *TenantUsageHandler) GetDailyUsage(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	// Parse date parameter
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid date format. Use YYYY-MM-DD", err)
	}

	usage, err := h.usageService.GetDailyUsage(c.Context(), tenantID, date)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, usage)
}

// GetUsageHistory godoc
// @Summary Get usage history
// @Description Get usage history for a date range
// @Tags tenant-usage
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.UsageHistoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{tenant_id}/usage/history [get]
func (h *TenantUsageHandler) GetUsageHistory(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	// Parse date parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "DATES_REQUIRED", "Both start_date and end_date are required", nil)
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_START_DATE", "Invalid start_date format. Use YYYY-MM-DD", err)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_END_DATE", "Invalid end_date format. Use YYYY-MM-DD", err)
	}

	history, err := h.usageService.GetUsageHistory(c.Context(), tenantID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, history)
}

// GetAPIUsageStats godoc
// @Summary Get API usage stats
// @Description Get API usage statistics for a specified number of days
// @Tags tenant-usage
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param days query int false "Number of days (default: 30)"
// @Success 200 {object} dto.UsageSummary
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{tenant_id}/usage/api [get]
func (h *TenantUsageHandler) GetAPIUsageStats(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	days := c.QueryInt("days", 30)
	if days < 1 || days > 365 {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DAYS", "Days must be between 1 and 365", nil)
	}

	stats, err := h.usageService.GetAPIUsageStats(c.Context(), tenantID, days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetPeakUsage godoc
// @Summary Get peak usage
// @Description Get peak usage statistics for a specified number of days
// @Tags tenant-usage
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param days query int false "Number of days (default: 30)"
// @Success 200 {object} dto.UsageTrackingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{tenant_id}/usage/peak [get]
func (h *TenantUsageHandler) GetPeakUsage(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	days := c.QueryInt("days", 30)
	if days < 1 || days > 365 {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DAYS", "Days must be between 1 and 365", nil)
	}

	peak, err := h.usageService.GetPeakUsage(c.Context(), tenantID, days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, peak)
}

// GetAverageUsage godoc
// @Summary Get average usage
// @Description Get average usage statistics for a specified number of days
// @Tags tenant-usage
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param days query int false "Number of days (default: 30)"
// @Success 200 {object} dto.UsageSummary
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{tenant_id}/usage/average [get]
func (h *TenantUsageHandler) GetAverageUsage(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	days := c.QueryInt("days", 30)
	if days < 1 || days > 365 {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DAYS", "Days must be between 1 and 365", nil)
	}

	avg, err := h.usageService.GetAverageUsage(c.Context(), tenantID, days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, avg)
}

// ============================================================================
// Usage Tracking Operations
// ============================================================================

// TrackFeatureUsage godoc
// @Summary Track feature usage
// @Description Track usage of a specific feature
// @Tags tenant-usage
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param usage body dto.TrackFeatureUsageRequest true "Feature usage data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{tenant_id}/usage/track [post]
func (h *TenantUsageHandler) TrackFeatureUsage(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	var req dto.TrackFeatureUsageRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.usageService.TrackFeatureUsage(c.Context(), &req, tenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message": "Feature usage tracked successfully",
	})
}

// IncrementAPIUsage godoc
// @Summary Increment API usage
// @Description Increment API call counter for tenant
// @Tags tenant-usage
// @Param tenant_id path string true "Tenant ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tenants/{tenant_id}/usage/api/increment [post]
func (h *TenantUsageHandler) IncrementAPIUsage(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	if err := h.usageService.IncrementAPIUsage(c.Context(), tenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message": "API usage incremented",
	})
}

// CheckAPIRateLimit godoc
// @Summary Check API rate limit
// @Description Check if tenant has exceeded API rate limit
// @Tags tenant-usage
// @Param tenant_id path string true "Tenant ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /tenants/{tenant_id}/usage/api/check [get]
func (h *TenantUsageHandler) CheckAPIRateLimit(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	allowed, err := h.usageService.CheckAPIRateLimit(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	if !allowed {
		return NewErrorResponse(c, fiber.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "API rate limit exceeded", nil)
	}

	return NewSuccessResponse(c, fiber.Map{
		"allowed": true,
		"message": "Within rate limit",
	})
}

// ============================================================================
// Admin Operations
// ============================================================================

// DeleteOldUsageRecords godoc
// @Summary Delete old usage records
// @Description Delete usage records older than specified days (admin only)
// @Tags tenant-usage
// @Param older_than_days query int true "Delete records older than days"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /usage/cleanup [delete]
func (h *TenantUsageHandler) DeleteOldUsageRecords(c *fiber.Ctx) error {
	days := c.QueryInt("older_than_days", 0)
	if days < 90 {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DAYS", "Can only delete records older than 90 days", nil)
	}

	count, err := h.usageService.DeleteOldUsageRecords(c.Context(), days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message":       "Old usage records deleted successfully",
		"deleted_count": count,
	})
}
