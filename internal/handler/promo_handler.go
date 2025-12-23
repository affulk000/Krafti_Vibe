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

// PromoCodeHandler handles HTTP requests for promo code operations
type PromoCodeHandler struct {
	promoService service.PromoCodeService
}

// NewPromoCodeHandler creates a new PromoCodeHandler
func NewPromoCodeHandler(promoService service.PromoCodeService) *PromoCodeHandler {
	return &PromoCodeHandler{
		promoService: promoService,
	}
}

// CreatePromoCode creates a new promo code
// @Summary Create a new promo code
// @Description Create a new promotional discount code
// @Tags PromoCodes
// @Accept json
// @Produce json
// @Param request body dto.CreatePromoCodeRequest true "Promo code creation request"
// @Success 201 {object} dto.PromoCodeResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 401 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /promo-codes [post]
func (h *PromoCodeHandler) CreatePromoCode(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreatePromoCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	tenantID := &authCtx.TenantID
	promo, err := h.promoService.CreatePromoCode(c.Context(), tenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(promo)
}

// GetPromoCode retrieves a promo code by ID
// @Summary Get promo code by ID
// @Description Retrieve promo code details by ID
// @Tags PromoCodes
// @Produce json
// @Param id path string true "Promo Code ID"
// @Success 200 {object} dto.PromoCodeResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /promo-codes/{id} [get]
func (h *PromoCodeHandler) GetPromoCode(c *fiber.Ctx) error {
	promoID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid promo code ID",
			"code":  "INVALID_PROMO_ID",
		})
	}

	promo, err := h.promoService.GetPromoCode(c.Context(), promoID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promo)
}

// GetPromoCodeByCode retrieves a promo code by code string
// @Summary Get promo code by code
// @Description Retrieve promo code details by code string
// @Tags PromoCodes
// @Produce json
// @Param code path string true "Promo Code"
// @Success 200 {object} dto.PromoCodeResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /promo-codes/code/{code} [get]
func (h *PromoCodeHandler) GetPromoCodeByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "promo code is required",
			"code":  "MISSING_CODE",
		})
	}

	promo, err := h.promoService.GetPromoCodeByCode(c.Context(), code)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promo)
}

// UpdatePromoCode updates a promo code
// @Summary Update promo code
// @Description Update promo code details
// @Tags PromoCodes
// @Accept json
// @Produce json
// @Param id path string true "Promo Code ID"
// @Param request body dto.UpdatePromoCodeRequest true "Update request"
// @Success 200 {object} dto.PromoCodeResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /promo-codes/{id} [put]
func (h *PromoCodeHandler) UpdatePromoCode(c *fiber.Ctx) error {
	promoID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid promo code ID",
			"code":  "INVALID_PROMO_ID",
		})
	}

	var req dto.UpdatePromoCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	promo, err := h.promoService.UpdatePromoCode(c.Context(), promoID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promo)
}

// DeletePromoCode deletes a promo code
// @Summary Delete promo code
// @Description Delete a promo code by ID
// @Tags PromoCodes
// @Param id path string true "Promo Code ID"
// @Success 204
// @Failure 400 {object} handler.ErrorResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /promo-codes/{id} [delete]
func (h *PromoCodeHandler) DeletePromoCode(c *fiber.Ctx) error {
	promoID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid promo code ID",
			"code":  "INVALID_PROMO_ID",
		})
	}

	if err := h.promoService.DeletePromoCode(c.Context(), promoID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListPromoCodes lists promo codes with filters
// @Summary List promo codes
// @Description List promo codes with optional filters
// @Tags PromoCodes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param type query string false "Discount type (percentage, fixed)"
// @Param is_active query boolean false "Filter by active status"
// @Param is_expired query boolean false "Filter by expired status"
// @Success 200 {object} dto.PromoCodeListResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes [get]
func (h *PromoCodeHandler) ListPromoCodes(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	filter := &dto.PromoCodeFilter{
		TenantID: &authCtx.TenantID,
		Page:     page,
		PageSize: pageSize,
	}

	// Optional filters
	if discountType := c.Query("type"); discountType != "" {
		dt := models.DiscountType(discountType)
		filter.Type = &dt
	}

	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filter.IsActive = &active
		}
	}

	if isExpired := c.Query("is_expired"); isExpired != "" {
		if expired, err := strconv.ParseBool(isExpired); err == nil {
			filter.IsExpired = &expired
		}
	}

	promos, err := h.promoService.ListPromoCodes(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promos)
}

// GetActivePromoCodes retrieves active promo codes
// @Summary Get active promo codes
// @Description Get all active promo codes for the tenant
// @Tags PromoCodes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PromoCodeListResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/active [get]
func (h *PromoCodeHandler) GetActivePromoCodes(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	tenantID := &authCtx.TenantID
	promos, err := h.promoService.GetActivePromoCodes(c.Context(), tenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promos)
}

// GetExpiredPromoCodes retrieves expired promo codes
// @Summary Get expired promo codes
// @Description Get all expired promo codes for the tenant
// @Tags PromoCodes
// @Produce json
// @Success 200 {array} dto.PromoCodeResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/expired [get]
func (h *PromoCodeHandler) GetExpiredPromoCodes(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	promos, err := h.promoService.GetExpiredPromoCodes(c.Context(), authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promos)
}

// GetExpiringPromoCodes retrieves promo codes expiring soon
// @Summary Get expiring promo codes
// @Description Get promo codes expiring within specified days
// @Tags PromoCodes
// @Produce json
// @Param days query int false "Days until expiry" default(7)
// @Success 200 {array} dto.PromoCodeResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/expiring [get]
func (h *PromoCodeHandler) GetExpiringPromoCodes(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	days, _ := strconv.Atoi(c.Query("days", "7"))

	promos, err := h.promoService.GetExpiringPromoCodes(c.Context(), authCtx.TenantID, days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promos)
}

// SearchPromoCodes searches promo codes
// @Summary Search promo codes
// @Description Search promo codes by code or description
// @Tags PromoCodes
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PromoCodeListResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/search [get]
func (h *PromoCodeHandler) SearchPromoCodes(c *fiber.Ctx) error {
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

	tenantID := &authCtx.TenantID
	promos, err := h.promoService.SearchPromoCodes(c.Context(), query, tenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promos)
}

// ValidatePromoCode validates a promo code
// @Summary Validate promo code
// @Description Validate a promo code and calculate discount
// @Tags PromoCodes
// @Accept json
// @Produce json
// @Param request body dto.ValidatePromoCodeRequest true "Validation request"
// @Success 200 {object} dto.PromoCodeValidationResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/validate [post]
func (h *PromoCodeHandler) ValidatePromoCode(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.ValidatePromoCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	tenantID := &authCtx.TenantID
	validation, err := h.promoService.ValidatePromoCode(c.Context(), &req, tenantID, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(validation)
}

// ApplyPromoCode applies a promo code
// @Summary Apply promo code
// @Description Apply a promo code to an order
// @Tags PromoCodes
// @Accept json
// @Produce json
// @Param body body object true "Apply request"
// @Success 200 {object} dto.PromoCodeValidationResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/apply [post]
func (h *PromoCodeHandler) ApplyPromoCode(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var body struct {
		Code      string     `json:"code" validate:"required"`
		Amount    float64    `json:"amount" validate:"required,min=0"`
		ServiceID *uuid.UUID `json:"service_id,omitempty"`
		ArtisanID *uuid.UUID `json:"artisan_id,omitempty"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	tenantID := &authCtx.TenantID
	result, err := h.promoService.ApplyPromoCode(
		c.Context(),
		body.Code,
		tenantID,
		authCtx.UserID,
		body.Amount,
		body.ServiceID,
		body.ArtisanID,
	)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(result)
}

// ActivatePromoCode activates a promo code
// @Summary Activate promo code
// @Description Activate a promo code
// @Tags PromoCodes
// @Param id path string true "Promo Code ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/{id}/activate [post]
func (h *PromoCodeHandler) ActivatePromoCode(c *fiber.Ctx) error {
	promoID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid promo code ID",
			"code":  "INVALID_PROMO_ID",
		})
	}

	if err := h.promoService.ActivatePromoCode(c.Context(), promoID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "promo code activated",
	})
}

// DeactivatePromoCode deactivates a promo code
// @Summary Deactivate promo code
// @Description Deactivate a promo code
// @Tags PromoCodes
// @Param id path string true "Promo Code ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/{id}/deactivate [post]
func (h *PromoCodeHandler) DeactivatePromoCode(c *fiber.Ctx) error {
	promoID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid promo code ID",
			"code":  "INVALID_PROMO_ID",
		})
	}

	if err := h.promoService.DeactivatePromoCode(c.Context(), promoID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "promo code deactivated",
	})
}

// BulkActivate bulk activates promo codes
// @Summary Bulk activate promo codes
// @Description Activate multiple promo codes at once
// @Tags PromoCodes
// @Accept json
// @Produce json
// @Param body body object true "Bulk activate request"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/bulk/activate [post]
func (h *PromoCodeHandler) BulkActivate(c *fiber.Ctx) error {
	var body struct {
		IDs []uuid.UUID `json:"ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if err := h.promoService.BulkActivate(c.Context(), body.IDs); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "promo codes activated",
		"count":   len(body.IDs),
	})
}

// BulkDeactivate bulk deactivates promo codes
// @Summary Bulk deactivate promo codes
// @Description Deactivate multiple promo codes at once
// @Tags PromoCodes
// @Accept json
// @Produce json
// @Param body body object true "Bulk deactivate request"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/bulk/deactivate [post]
func (h *PromoCodeHandler) BulkDeactivate(c *fiber.Ctx) error {
	var body struct {
		IDs []uuid.UUID `json:"ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if err := h.promoService.BulkDeactivate(c.Context(), body.IDs); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "promo codes deactivated",
		"count":   len(body.IDs),
	})
}

// BulkDelete bulk deletes promo codes
// @Summary Bulk delete promo codes
// @Description Delete multiple promo codes at once
// @Tags PromoCodes
// @Accept json
// @Produce json
// @Param body body object true "Bulk delete request"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/bulk/delete [delete]
func (h *PromoCodeHandler) BulkDelete(c *fiber.Ctx) error {
	var body struct {
		IDs []uuid.UUID `json:"ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if err := h.promoService.BulkDelete(c.Context(), body.IDs); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "promo codes deleted",
		"count":   len(body.IDs),
	})
}

// GetValidPromoCodesForService retrieves valid promo codes for a service
// @Summary Get valid promo codes for service
// @Description Get all valid promo codes applicable to a service
// @Tags PromoCodes
// @Produce json
// @Param service_id path string true "Service ID"
// @Success 200 {array} dto.PromoCodeResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/service/{service_id} [get]
func (h *PromoCodeHandler) GetValidPromoCodesForService(c *fiber.Ctx) error {
	serviceID, err := uuid.Parse(c.Params("service_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid service ID",
			"code":  "INVALID_SERVICE_ID",
		})
	}

	promos, err := h.promoService.GetValidPromoCodesForService(c.Context(), serviceID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promos)
}

// GetValidPromoCodesForArtisan retrieves valid promo codes for an artisan
// @Summary Get valid promo codes for artisan
// @Description Get all valid promo codes applicable to an artisan
// @Tags PromoCodes
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Success 200 {array} dto.PromoCodeResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/artisan/{artisan_id} [get]
func (h *PromoCodeHandler) GetValidPromoCodesForArtisan(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid artisan ID",
			"code":  "INVALID_ARTISAN_ID",
		})
	}

	promos, err := h.promoService.GetValidPromoCodesForArtisan(c.Context(), artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(promos)
}

// GetPromoCodeStats retrieves promo code statistics
// @Summary Get promo code statistics
// @Description Get detailed statistics for a promo code
// @Tags PromoCodes
// @Produce json
// @Param id path string true "Promo Code ID"
// @Success 200 {object} dto.PromoCodeStatsResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/{id}/stats [get]
func (h *PromoCodeHandler) GetPromoCodeStats(c *fiber.Ctx) error {
	promoID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid promo code ID",
			"code":  "INVALID_PROMO_ID",
		})
	}

	stats, err := h.promoService.GetPromoCodeStats(c.Context(), promoID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(stats)
}

// GetTopPerformingPromoCodes retrieves top performing promo codes
// @Summary Get top performing promo codes
// @Description Get top performing promo codes by usage
// @Tags PromoCodes
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} repository.PromoCodePerformance
// @Failure 400 {object} handler.ErrorResponse
// @Router /promo-codes/analytics/top-performing [get]
func (h *PromoCodeHandler) GetTopPerformingPromoCodes(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid start date format, use YYYY-MM-DD",
				"code":  "INVALID_DATE_FORMAT",
			})
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default: 1 month ago
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid end date format, use YYYY-MM-DD",
				"code":  "INVALID_DATE_FORMAT",
			})
		}
	} else {
		endDate = time.Now()
	}

	performance, err := h.promoService.GetTopPerformingPromoCodes(
		c.Context(),
		authCtx.TenantID,
		limit,
		startDate,
		endDate,
	)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(performance)
}
