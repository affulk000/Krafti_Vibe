package handler

import (
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// DataExportHandler handles HTTP requests for data export operations
type DataExportHandler struct {
	exportService service.DataExportService
}

// NewDataExportHandler creates a new data export handler
func NewDataExportHandler(exportService service.DataExportService) *DataExportHandler {
	return &DataExportHandler{
		exportService: exportService,
	}
}

// ============================================================================
// Core Export Operations
// ============================================================================

// RequestDataExport godoc
// @Summary Request data export
// @Description Request a data export for GDPR compliance or backup purposes
// @Tags data-exports
// @Accept json
// @Produce json
// @Param export body dto.DataExportRequest true "Export request data"
// @Success 201 {object} dto.DataExportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /data-exports [post]
func (h *DataExportHandler) RequestDataExport(c *fiber.Ctx) error {
	var req dto.DataExportRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	// Get tenant ID from context
	tenantID, err := GetTenantID(c)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "Tenant ID is required", err)
	}

	// Get requesting user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized, "USER_REQUIRED", "User authentication required", err)
	}

	export, err := h.exportService.RequestDataExport(c.Context(), &req, tenantID, userID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, export, "Data export request created successfully")
}

// GetDataExportStatus godoc
// @Summary Get export status
// @Description Get the status and details of a data export request
// @Tags data-exports
// @Produce json
// @Param id path string true "Export ID"
// @Success 200 {object} dto.DataExportResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /data-exports/{id} [get]
func (h *DataExportHandler) GetDataExportStatus(c *fiber.Ctx) error {
	exportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid export ID", err)
	}

	tenantID, err := GetTenantID(c)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "Tenant ID is required", err)
	}

	export, err := h.exportService.GetDataExportStatus(c.Context(), exportID, tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, export)
}

// ListDataExports godoc
// @Summary List data exports
// @Description List all data export requests for the tenant
// @Tags data-exports
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Export status"
// @Param export_type query string false "Export type"
// @Success 200 {object} dto.DataExportListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /data-exports [get]
func (h *DataExportHandler) ListDataExports(c *fiber.Ctx) error {
	tenantID, err := GetTenantID(c)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "Tenant ID is required", err)
	}

	// Parse pagination and filters
	page, pageSize := ParsePagination(c)

	var statusPtr *string
	if status := c.Query("status"); status != "" {
		statusPtr = &status
	}

	filter := &dto.DataExportFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   statusPtr,
	}

	exports, err := h.exportService.ListDataExports(c.Context(), tenantID, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, exports)
}

// CancelDataExport godoc
// @Summary Cancel data export
// @Description Cancel a pending or processing data export request
// @Tags data-exports
// @Param id path string true "Export ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /data-exports/{id}/cancel [post]
func (h *DataExportHandler) CancelDataExport(c *fiber.Ctx) error {
	exportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid export ID", err)
	}

	tenantID, err := GetTenantID(c)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "Tenant ID is required", err)
	}

	if err := h.exportService.CancelDataExport(c.Context(), exportID, tenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message": "Export cancelled successfully",
	})
}

// ============================================================================
// Admin Operations
// ============================================================================

// GetPendingExports godoc
// @Summary Get pending exports
// @Description Get all pending export requests (admin only)
// @Tags data-exports
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /data-exports/pending [get]
func (h *DataExportHandler) GetPendingExports(c *fiber.Ctx) error {
	exports, err := h.exportService.GetPendingExports(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"exports": exports,
		"count":   len(exports),
	})
}

// GetExportsByStatus godoc
// @Summary Get exports by status
// @Description Get exports filtered by status (admin only)
// @Tags data-exports
// @Produce json
// @Param status query string true "Export status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.DataExportListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /data-exports/by-status [get]
func (h *DataExportHandler) GetExportsByStatus(c *fiber.Ctx) error {
	status := c.Query("status")
	if status == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "STATUS_REQUIRED", "Status parameter is required", nil)
	}

	page, pageSize := ParsePagination(c)
	filter := &dto.DataExportFilter{
		Page:     page,
		PageSize: pageSize,
	}

	exports, err := h.exportService.GetExportsByStatus(c.Context(), status, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, exports)
}

// DeleteExpiredExports godoc
// @Summary Delete expired exports
// @Description Delete all expired export files (admin only)
// @Tags data-exports
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /data-exports/cleanup [post]
func (h *DataExportHandler) DeleteExpiredExports(c *fiber.Ctx) error {
	count, err := h.exportService.DeleteExpiredExports(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message":        "Expired exports deleted successfully",
		"deleted_count":  count,
	})
}
