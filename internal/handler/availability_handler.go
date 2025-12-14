package handler

import (
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AvailabilityHandler handles availability HTTP requests
type AvailabilityHandler struct {
	service service.AvailabilityService
}

// NewAvailabilityHandler creates a new availability handler
func NewAvailabilityHandler(service service.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{
		service: service,
	}
}

// CreateAvailability creates a new availability slot
// @Summary Create availability slot
// @Tags Availability
// @Accept json
// @Produce json
// @Param availability body dto.CreateAvailabilitySlotRequest true "Availability details"
// @Success 201 {object} dto.AvailabilitySlotResponse
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Router /api/v1/availability [post]
func (h *AvailabilityHandler) CreateAvailability(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateAvailabilitySlotRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	availability, err := h.service.CreateAvailability(c.Context(), authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(availability)
}

// GetAvailability retrieves an availability slot by ID
// @Summary Get availability slot
// @Tags Availability
// @Produce json
// @Param id path string true "Availability ID"
// @Success 200 {object} dto.AvailabilitySlotResponse
// @Failure 404 {object} fiber.Map
// @Router /api/v1/availability/{id} [get]
func (h *AvailabilityHandler) GetAvailability(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid availability ID",
			"code":  "INVALID_AVAILABILITY_ID",
		})
	}

	availability, err := h.service.GetAvailability(c.Context(), id, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(availability)
}

// UpdateAvailability updates an availability slot
// @Summary Update availability slot
// @Tags Availability
// @Accept json
// @Produce json
// @Param id path string true "Availability ID"
// @Param availability body dto.UpdateAvailabilitySlotRequest true "Updated availability details"
// @Success 200 {object} dto.AvailabilitySlotResponse
// @Failure 400 {object} fiber.Map
// @Router /api/v1/availability/{id} [put]
func (h *AvailabilityHandler) UpdateAvailability(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid availability ID",
			"code":  "INVALID_AVAILABILITY_ID",
		})
	}

	var req dto.UpdateAvailabilitySlotRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	availability, err := h.service.UpdateAvailability(c.Context(), id, authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(availability)
}

// DeleteAvailability deletes an availability slot
// @Summary Delete availability slot
// @Tags Availability
// @Produce json
// @Param id path string true "Availability ID"
// @Success 200 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /api/v1/availability/{id} [delete]
func (h *AvailabilityHandler) DeleteAvailability(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid availability ID",
			"code":  "INVALID_AVAILABILITY_ID",
		})
	}

	if err := h.service.DeleteAvailability(c.Context(), id, authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "availability deleted successfully",
	})
}

// ListAvailabilities lists availability slots with filtering
// @Summary List availability slots
// @Tags Availability
// @Produce json
// @Param artisan_id query string true "Artisan ID"
// @Param type query string false "Filter by type"
// @Param day_of_week query int false "Filter by day of week (0-6)"
// @Param start_date query string false "Filter by start date"
// @Param end_date query string false "Filter by end date"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.AvailabilitySlotListResponse
// @Router /api/v1/availability [get]
func (h *AvailabilityHandler) ListAvailabilities(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	artisanIDStr := c.Query("artisan_id")
	if artisanIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "artisan_id is required",
			"code":  "MISSING_ARTISAN_ID",
		})
	}

	artisanID, err := uuid.Parse(artisanIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid artisan_id",
			"code":  "INVALID_ARTISAN_ID",
		})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 20)

	filter := &dto.AvailabilitySlotFilter{
		ArtisanID: artisanID,
		Page:      page,
		PageSize:  pageSize,
	}

	// Apply type filter if provided
	if availType := c.Query("type"); availType != "" {
		t := models.AvailabilityType(availType)
		filter.Type = &t
	}

	// Apply day of week filter if provided
	if dayOfWeek := c.QueryInt("day_of_week", -1); dayOfWeek >= 0 && dayOfWeek <= 6 {
		filter.DayOfWeek = &dayOfWeek
	}

	// Apply date range filter if provided
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	availabilities, err := h.service.ListAvailabilities(c.Context(), filter, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(availabilities)
}

// ListByType lists availability slots by type
// @Summary List availability slots by type
// @Tags Availability
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param type path string true "Availability type"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.AvailabilitySlotListResponse
// @Router /api/v1/availability/artisan/{artisan_id}/type/{type} [get]
func (h *AvailabilityHandler) ListByType(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid artisan_id",
			"code":  "INVALID_ARTISAN_ID",
		})
	}

	availType := models.AvailabilityType(c.Params("type"))
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 20)

	availabilities, err := h.service.ListByType(c.Context(), artisanID, availType, authCtx.TenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(availabilities)
}

// GetByDayOfWeek retrieves availability slots for a specific day of week
// @Summary Get availability by day of week
// @Tags Availability
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param day path int true "Day of week (0-6, 0=Sunday)"
// @Success 200 {object} []dto.AvailabilitySlotResponse
// @Router /api/v1/availability/artisan/{artisan_id}/day/{day} [get]
func (h *AvailabilityHandler) GetByDayOfWeek(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid artisan_id",
			"code":  "INVALID_ARTISAN_ID",
		})
	}

	dayOfWeek, err := c.ParamsInt("day")
	if err != nil || dayOfWeek < 0 || dayOfWeek > 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid day of week (must be 0-6)",
			"code":  "INVALID_DAY_OF_WEEK",
		})
	}

	availabilities, err := h.service.GetByDayOfWeek(c.Context(), artisanID, dayOfWeek, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(availabilities)
}

// GetWeeklySchedule retrieves weekly schedule for an artisan
// @Summary Get weekly schedule
// @Tags Availability
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param week_start query string false "Week start date (YYYY-MM-DD)" default(current week)
// @Success 200 {object} dto.WeeklyScheduleResponse
// @Router /api/v1/availability/artisan/{artisan_id}/weekly [get]
func (h *AvailabilityHandler) GetWeeklySchedule(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid artisan_id",
			"code":  "INVALID_ARTISAN_ID",
		})
	}

	// Parse week start date or use current week
	var weekStart time.Time
	if weekStartStr := c.Query("week_start"); weekStartStr != "" {
		weekStart, err = time.Parse("2006-01-02", weekStartStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid week_start date format (use YYYY-MM-DD)",
				"code":  "INVALID_DATE_FORMAT",
			})
		}
	} else {
		// Default to start of current week (Sunday)
		now := time.Now()
		weekStart = now.AddDate(0, 0, -int(now.Weekday()))
	}

	schedule, err := h.service.GetWeeklySchedule(c.Context(), artisanID, authCtx.TenantID, weekStart)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(schedule)
}

// CheckAvailability checks if an artisan is available for a given time slot
// @Summary Check availability
// @Tags Availability
// @Accept json
// @Produce json
// @Param request body dto.CheckAvailabilitySlotRequest true "Availability check request"
// @Success 200 {object} dto.AvailabilitySlotCheckResponse
// @Failure 400 {object} fiber.Map
// @Router /api/v1/availability/check [post]
func (h *AvailabilityHandler) CheckAvailability(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CheckAvailabilitySlotRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	result, err := h.service.CheckAvailability(c.Context(), &req, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(result)
}

// BulkCreateAvailability creates multiple availability slots
// @Summary Bulk create availability
// @Tags Availability
// @Accept json
// @Produce json
// @Param request body dto.BulkCreateAvailabilitySlotRequest true "Bulk availability request"
// @Success 201 {object} []dto.AvailabilitySlotResponse
// @Failure 400 {object} fiber.Map
// @Router /api/v1/availability/bulk [post]
func (h *AvailabilityHandler) BulkCreateAvailability(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.BulkCreateAvailabilitySlotRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	availabilities, err := h.service.BulkCreateAvailability(c.Context(), authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(availabilities)
}

// DeleteByType deletes all availability slots of a specific type
// @Summary Delete availability by type
// @Tags Availability
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param type path string true "Availability type"
// @Success 200 {object} fiber.Map
// @Router /api/v1/availability/artisan/{artisan_id}/type/{type} [delete]
func (h *AvailabilityHandler) DeleteByType(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid artisan_id",
			"code":  "INVALID_ARTISAN_ID",
		})
	}

	availType := models.AvailabilityType(c.Params("type"))

	if err := h.service.DeleteByType(c.Context(), artisanID, availType, authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "availabilities deleted successfully",
	})
}
