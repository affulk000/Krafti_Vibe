package handler

import (
	"strconv"

	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ArtisanHandler handles HTTP requests for artisan operations
type ArtisanHandler struct {
	artisanService service.ArtisanService
}

// NewArtisanHandler creates a new artisan handler
func NewArtisanHandler(artisanService service.ArtisanService) *ArtisanHandler {
	return &ArtisanHandler{
		artisanService: artisanService,
	}
}

// ============================================================================
// CRUD Operations
// ============================================================================

// CreateArtisan godoc
// @Summary Create a new artisan
// @Description Create a new artisan profile
// @Tags artisans
// @Accept json
// @Produce json
// @Param artisan body dto.CreateArtisanRequest true "Artisan creation data"
// @Success 201 {object} dto.ArtisanResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans [post]
func (h *ArtisanHandler) CreateArtisan(c *fiber.Ctx) error {
	var req dto.CreateArtisanRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	artisan, err := h.artisanService.CreateArtisan(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, artisan, "Artisan created successfully")
}

// GetArtisan godoc
// @Summary Get artisan by ID
// @Description Get detailed artisan information by ID
// @Tags artisans
// @Produce json
// @Param id path string true "Artisan ID"
// @Success 200 {object} dto.ArtisanResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/{id} [get]
func (h *ArtisanHandler) GetArtisan(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	artisan, err := h.artisanService.GetArtisan(c.Context(), artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisan)
}

// GetArtisanByUserID godoc
// @Summary Get artisan by user ID
// @Description Get artisan profile by user ID
// @Tags artisans
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} dto.ArtisanResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/user/{user_id} [get]
func (h *ArtisanHandler) GetArtisanByUserID(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	artisan, err := h.artisanService.GetArtisanByUserID(c.Context(), userID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisan)
}

// UpdateArtisan godoc
// @Summary Update artisan
// @Description Update artisan information
// @Tags artisans
// @Accept json
// @Produce json
// @Param id path string true "Artisan ID"
// @Param artisan body dto.UpdateArtisanRequest true "Update data"
// @Success 200 {object} dto.ArtisanResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/{id} [put]
func (h *ArtisanHandler) UpdateArtisan(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	var req dto.UpdateArtisanRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	artisan, err := h.artisanService.UpdateArtisan(c.Context(), artisanID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisan, "Artisan updated successfully")
}

// DeleteArtisan godoc
// @Summary Delete artisan
// @Description Soft delete an artisan profile
// @Tags artisans
// @Param id path string true "Artisan ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/{id} [delete]
func (h *ArtisanHandler) DeleteArtisan(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	if err := h.artisanService.DeleteArtisan(c.Context(), artisanID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// ============================================================================
// Query Operations
// ============================================================================

// ListArtisans godoc
// @Summary List artisans
// @Description Get a paginated list of artisans with filters
// @Tags artisans
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param tenant_id query string false "Filter by tenant ID"
// @Param specialization query string false "Filter by specialization"
// @Param is_available query boolean false "Filter by availability"
// @Success 200 {object} dto.ArtisanListResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans [get]
func (h *ArtisanHandler) ListArtisans(c *fiber.Ctx) error {
	// Get auth context for tenant isolation
	authCtx := MustGetAuthContext(c)

	filter := dto.ArtisanFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	// Use tenant ID from auth context (for tenant isolation)
	// Platform users can optionally filter by tenant via query param
	if authCtx.TenantID != uuid.Nil {
		filter.TenantID = &authCtx.TenantID
	} else if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		// Platform users can filter by specific tenant
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = &tenantID
		}
	}

	// Parse specialization filter
	if spec := c.Query("specialization"); spec != "" {
		filter.Specialization = &spec
	}

	// Parse availability filter
	if availStr := c.Query("is_available"); availStr != "" {
		isAvail := availStr == "true"
		filter.IsAvailable = &isAvail
	}

	// Parse min rating filter
	if ratingStr := c.Query("min_rating"); ratingStr != "" {
		if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil {
			filter.MinRating = &rating
		}
	}

	artisans, err := h.artisanService.ListArtisans(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisans)
}

// SearchArtisans godoc
// @Summary Search artisans
// @Description Search for artisans by name or specialization
// @Tags artisans
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ArtisanListResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/search [get]
func (h *ArtisanHandler) SearchArtisans(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_QUERY", "Search query is required", nil)
	}

	filter := dto.ArtisanFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = &tid
		}
	}

	artisans, err := h.artisanService.SearchArtisans(c.Context(), query, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisans)
}

// GetAvailableArtisans godoc
// @Summary Get available artisans
// @Description Get all currently available artisans
// @Tags artisans
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ArtisanListResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/available [get]
func (h *ArtisanHandler) GetAvailableArtisans(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	artisans, err := h.artisanService.GetAvailableArtisans(c.Context(), tenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisans)
}

// GetArtisansBySpecialization godoc
// @Summary Get artisans by specialization
// @Description Get artisans filtered by specialization
// @Tags artisans
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param specialization query string true "Specialization"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ArtisanListResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/by-specialization [get]
func (h *ArtisanHandler) GetArtisansBySpecialization(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	specialization := c.Query("specialization")
	if specialization == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_SPECIALIZATION", "Specialization is required", nil)
	}

	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	artisans, err := h.artisanService.GetArtisansBySpecialization(c.Context(), tenantID, specialization, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisans)
}

// GetTopRatedArtisans godoc
// @Summary Get top rated artisans
// @Description Get artisans with highest ratings
// @Tags artisans
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/top-rated [get]
func (h *ArtisanHandler) GetTopRatedArtisans(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	limit := getIntQuery(c, "limit", 10)

	artisans, err := h.artisanService.GetTopRatedArtisans(c.Context(), tenantID, limit)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisans)
}

// FindNearbyArtisans godoc
// @Summary Find nearby artisans
// @Description Find artisans near a specific location
// @Tags artisans
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param latitude query number true "Latitude"
// @Param longitude query number true "Longitude"
// @Param radius query int false "Radius in km" default(10)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/nearby [get]
func (h *ArtisanHandler) FindNearbyArtisans(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")

	if latStr == "" || lonStr == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_LOCATION", "Latitude and longitude are required", nil)
	}

	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_LATITUDE", "Invalid latitude value", err)
	}

	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_LONGITUDE", "Invalid longitude value", err)
	}

	radiusKm := getIntQuery(c, "radius", 10)

	artisans, err := h.artisanService.FindNearbyArtisans(c.Context(), tenantID, latitude, longitude, radiusKm)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, artisans)
}

// ============================================================================
// Availability Management
// ============================================================================

// UpdateAvailability godoc
// @Summary Update artisan availability
// @Description Update artisan's availability status
// @Tags artisans
// @Accept json
// @Produce json
// @Param id path string true "Artisan ID"
// @Param availability body AvailabilityUpdateRequest true "Availability data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/{id}/availability [put]
func (h *ArtisanHandler) UpdateAvailability(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	var req AvailabilityUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.artisanService.UpdateAvailability(c.Context(), artisanID, req.Available, req.Note); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Availability updated successfully")
}

// BatchUpdateAvailability godoc
// @Summary Batch update availability
// @Description Update availability for multiple artisans
// @Tags artisans
// @Accept json
// @Produce json
// @Param batch body BatchAvailabilityUpdateRequest true "Batch update data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/batch/availability [put]
func (h *ArtisanHandler) BatchUpdateAvailability(c *fiber.Ctx) error {
	var req BatchAvailabilityUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if len(req.ArtisanIDs) == 0 {
		return NewErrorResponse(c, fiber.StatusBadRequest, "EMPTY_IDS", "Artisan IDs list cannot be empty", nil)
	}

	if err := h.artisanService.BatchUpdateAvailability(c.Context(), req.ArtisanIDs, req.Available); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Availability updated successfully")
}

// ============================================================================
// Statistics
// ============================================================================

// GetArtisanStats godoc
// @Summary Get artisan statistics
// @Description Get comprehensive statistics for an artisan
// @Tags artisans
// @Produce json
// @Param id path string true "Artisan ID"
// @Success 200 {object} dto.ArtisanStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/{id}/stats [get]
func (h *ArtisanHandler) GetArtisanStats(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	stats, err := h.artisanService.GetArtisanStats(c.Context(), artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetDashboardStats godoc
// @Summary Get artisan dashboard statistics
// @Description Get dashboard statistics for an artisan
// @Tags artisans
// @Produce json
// @Param id path string true "Artisan ID"
// @Success 200 {object} dto.ArtisanServiceDashboardResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /artisans/{id}/dashboard [get]
func (h *ArtisanHandler) GetDashboardStats(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	dashboard, err := h.artisanService.GetDashboardStats(c.Context(), artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, dashboard)
}

// ============================================================================
// Request Types
// ============================================================================

type AvailabilityUpdateRequest struct {
	Available bool   `json:"available"`
	Note      string `json:"note,omitempty"`
}

type BatchAvailabilityUpdateRequest struct {
	ArtisanIDs []uuid.UUID `json:"artisan_ids"`
	Available  bool        `json:"available"`
}
