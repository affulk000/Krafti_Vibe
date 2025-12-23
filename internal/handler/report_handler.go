package handler

import (
	"strconv"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ReportHandler handles HTTP requests for report operations
type ReportHandler struct {
	reportService service.ReportService
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// CreateReport creates a new report
// @Summary Create a new report
// @Description Create a new report request for generation
// @Tags Reports
// @Accept json
// @Produce json
// @Param request body dto.CreateReportRequest true "Report creation request"
// @Success 201 {object} dto.ReportResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 401 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /reports [post]
func (h *ReportHandler) CreateReport(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	report, err := h.reportService.CreateReport(c.Context(), authCtx.TenantID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// GetReport retrieves a report by ID
// @Summary Get report by ID
// @Description Retrieve report details by ID
// @Tags Reports
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} dto.ReportResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /reports/{id} [get]
func (h *ReportHandler) GetReport(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	report, err := h.reportService.GetReport(c.Context(), reportID, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(report)
}

// UpdateReport updates a report
// @Summary Update report
// @Description Update report metadata
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param request body dto.UpdateReportRequest true "Update request"
// @Success 200 {object} dto.ReportResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /reports/{id} [put]
func (h *ReportHandler) UpdateReport(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	var req dto.UpdateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	report, err := h.reportService.UpdateReport(c.Context(), reportID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(report)
}

// DeleteReport deletes a report
// @Summary Delete report
// @Description Delete a report by ID
// @Tags Reports
// @Param id path string true "Report ID"
// @Success 204
// @Failure 400 {object} handler.ErrorResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /reports/{id} [delete]
func (h *ReportHandler) DeleteReport(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	if err := h.reportService.DeleteReport(c.Context(), reportID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListReports lists reports with filters
// @Summary List reports
// @Description List reports with optional filters
// @Tags Reports
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param type query string false "Report type"
// @Param status query string false "Report status"
// @Param is_scheduled query boolean false "Filter by scheduled"
// @Success 200 {object} dto.ReportListResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports [get]
func (h *ReportHandler) ListReports(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	filter := &dto.ReportFilter{
		TenantID: authCtx.TenantID,
		Page:     page,
		PageSize: pageSize,
	}

	// Optional filters
	if reportType := c.Query("type"); reportType != "" {
		rt := models.ReportType(reportType)
		filter.Type = &rt
	}

	if status := c.Query("status"); status != "" {
		st := models.ReportStatus(status)
		filter.Status = &st
	}

	if isScheduled := c.Query("is_scheduled"); isScheduled != "" {
		if scheduled, err := strconv.ParseBool(isScheduled); err == nil {
			filter.IsScheduled = &scheduled
		}
	}

	reports, err := h.reportService.ListReports(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(reports)
}

// GetPendingReports retrieves pending reports
// @Summary Get pending reports
// @Description Get all pending reports for the tenant
// @Tags Reports
// @Produce json
// @Success 200 {array} dto.ReportResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/pending [get]
func (h *ReportHandler) GetPendingReports(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reports, err := h.reportService.GetPendingReports(c.Context(), authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(reports)
}

// GetScheduledReports retrieves scheduled reports
// @Summary Get scheduled reports
// @Description Get all scheduled reports for the tenant
// @Tags Reports
// @Produce json
// @Success 200 {array} dto.ReportResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/scheduled [get]
func (h *ReportHandler) GetScheduledReports(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reports, err := h.reportService.GetScheduledReports(c.Context(), authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(reports)
}

// GetFailedReports retrieves failed reports
// @Summary Get failed reports
// @Description Get all failed reports for the tenant
// @Tags Reports
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ReportListResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/failed [get]
func (h *ReportHandler) GetFailedReports(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	reports, err := h.reportService.GetFailedReports(c.Context(), authCtx.TenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(reports)
}

// MarkAsGenerating marks a report as generating
// @Summary Mark report as generating
// @Description Mark a report as currently being generated
// @Tags Reports
// @Param id path string true "Report ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/{id}/generating [post]
func (h *ReportHandler) MarkAsGenerating(c *fiber.Ctx) error {
	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	if err := h.reportService.MarkAsGenerating(c.Context(), reportID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "report marked as generating",
	})
}

// MarkAsCompleted marks a report as completed
// @Summary Mark report as completed
// @Description Mark a report as completed with file URL
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param body body object true "Completion data"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/{id}/completed [post]
func (h *ReportHandler) MarkAsCompleted(c *fiber.Ctx) error {
	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	var body struct {
		FileURL string `json:"file_url" validate:"required,url"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if err := h.reportService.MarkAsCompleted(c.Context(), reportID, body.FileURL); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "report marked as completed",
	})
}

// MarkAsFailed marks a report as failed
// @Summary Mark report as failed
// @Description Mark a report as failed with error message
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param body body object true "Failure data"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/{id}/failed [post]
func (h *ReportHandler) MarkAsFailed(c *fiber.Ctx) error {
	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	var body struct {
		ErrorMessage string `json:"error_message" validate:"required"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if err := h.reportService.MarkAsFailed(c.Context(), reportID, body.ErrorMessage); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "report marked as failed",
	})
}

// RetryFailedReport retries a failed report
// @Summary Retry failed report
// @Description Retry generation of a failed report
// @Tags Reports
// @Param id path string true "Report ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/{id}/retry [post]
func (h *ReportHandler) RetryFailedReport(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	if err := h.reportService.RetryFailedReport(c.Context(), reportID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "report queued for retry",
	})
}

// EnableSchedule enables scheduling for a report
// @Summary Enable report schedule
// @Description Enable automatic scheduling for a report
// @Tags Reports
// @Param id path string true "Report ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/{id}/schedule/enable [post]
func (h *ReportHandler) EnableSchedule(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	if err := h.reportService.EnableSchedule(c.Context(), reportID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "report schedule enabled",
	})
}

// DisableSchedule disables scheduling for a report
// @Summary Disable report schedule
// @Description Disable automatic scheduling for a report
// @Tags Reports
// @Param id path string true "Report ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/{id}/schedule/disable [post]
func (h *ReportHandler) DisableSchedule(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	if err := h.reportService.DisableSchedule(c.Context(), reportID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "report schedule disabled",
	})
}

// UpdateScheduleCron updates schedule cron expression
// @Summary Update schedule cron
// @Description Update the cron expression for a scheduled report
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param body body object true "Cron data"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/{id}/schedule/cron [put]
func (h *ReportHandler) UpdateScheduleCron(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid report ID",
			"code":  "INVALID_REPORT_ID",
		})
	}

	var body struct {
		CronExpression string `json:"cron_expression" validate:"required"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if err := h.reportService.UpdateScheduleCron(c.Context(), reportID, authCtx.UserID, body.CronExpression); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "schedule cron updated",
	})
}

// GetReportStats retrieves report statistics
// @Summary Get report statistics
// @Description Get comprehensive report statistics for the tenant
// @Tags Reports
// @Produce json
// @Success 200 {object} dto.ReportStatsResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/stats [get]
func (h *ReportHandler) GetReportStats(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	stats, err := h.reportService.GetReportStats(c.Context(), authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(stats)
}

// GetReportTypeUsage retrieves report type usage
// @Summary Get report type usage
// @Description Get report type usage statistics over time
// @Tags Reports
// @Produce json
// @Param days query int false "Number of days" default(30)
// @Success 200 {array} repository.ReportTypeUsage
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/analytics/type-usage [get]
func (h *ReportHandler) GetReportTypeUsage(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))

	usage, err := h.reportService.GetReportTypeUsage(c.Context(), authCtx.TenantID, days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(usage)
}

// GetUserReportActivity retrieves user report activity
// @Summary Get user report activity
// @Description Get report activity for the current user
// @Tags Reports
// @Produce json
// @Success 200 {object} repository.UserReportActivity
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/analytics/user-activity [get]
func (h *ReportHandler) GetUserReportActivity(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	activity, err := h.reportService.GetUserReportActivity(c.Context(), authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(activity)
}

// GetReportGenerationMetrics retrieves generation metrics
// @Summary Get generation metrics
// @Description Get report generation performance metrics
// @Tags Reports
// @Produce json
// @Param days query int false "Number of days" default(30)
// @Success 200 {object} repository.ReportGenerationMetrics
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/analytics/generation-metrics [get]
func (h *ReportHandler) GetReportGenerationMetrics(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))

	metrics, err := h.reportService.GetReportGenerationMetrics(c.Context(), authCtx.TenantID, days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(metrics)
}

// DeleteOldReports deletes old reports
// @Summary Delete old reports
// @Description Delete reports older than specified days
// @Tags Reports
// @Param days query int false "Days old" default(90)
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/cleanup/old [delete]
func (h *ReportHandler) DeleteOldReports(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	days, _ := strconv.Atoi(c.Query("days", "90"))
	duration := time.Duration(days) * 24 * time.Hour

	if err := h.reportService.DeleteOldReports(c.Context(), authCtx.TenantID, duration); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "old reports deleted",
		"days":    days,
	})
}

// DeleteFailedReports deletes failed reports
// @Summary Delete failed reports
// @Description Delete failed reports older than specified days
// @Tags Reports
// @Param days query int false "Days old" default(30)
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/cleanup/failed [delete]
func (h *ReportHandler) DeleteFailedReports(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))
	duration := time.Duration(days) * 24 * time.Hour

	if err := h.reportService.DeleteFailedReports(c.Context(), authCtx.TenantID, duration); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "failed reports deleted",
		"days":    days,
	})
}

// SearchReports searches reports
// @Summary Search reports
// @Description Search reports by name or description
// @Tags Reports
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ReportListResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /reports/search [get]
func (h *ReportHandler) SearchReports(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "search query is required",
			"code":  "MISSING_QUERY",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	reports, err := h.reportService.SearchReports(c.Context(), authCtx.TenantID, query, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(reports)
}
