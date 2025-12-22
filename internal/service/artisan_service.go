package service

import (
	"context"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// artisanService defines the interface for artisan operations
type ArtisanService interface {
	// Core Operations
	CreateArtisan(ctx context.Context, req *dto.CreateArtisanRequest) (*dto.ArtisanResponse, error)
	GetArtisan(ctx context.Context, id uuid.UUID) (*dto.ArtisanResponse, error)
	GetArtisanByUserID(ctx context.Context, userID uuid.UUID) (*dto.ArtisanResponse, error)
	UpdateArtisan(ctx context.Context, id uuid.UUID, req *dto.UpdateArtisanRequest) (*dto.ArtisanResponse, error)
	DeleteArtisan(ctx context.Context, id uuid.UUID) error

	// Query Operations
	ListArtisans(ctx context.Context, filter dto.ArtisanFilter) (*dto.ArtisanListResponse, error)
	SearchArtisans(ctx context.Context, query string, filter dto.ArtisanFilter) (*dto.ArtisanListResponse, error)
	GetAvailableArtisans(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.ArtisanListResponse, error)
	GetArtisansBySpecialization(ctx context.Context, tenantID uuid.UUID, specialization string, page, pageSize int) (*dto.ArtisanListResponse, error)
	GetTopRatedArtisans(ctx context.Context, tenantID uuid.UUID, limit int) ([]*dto.ArtisanResponse, error)
	FindNearbyArtisans(ctx context.Context, tenantID uuid.UUID, latitude, longitude float64, radiusKm int) ([]*dto.ArtisanResponse, error)

	// Availability Management
	UpdateAvailability(ctx context.Context, artisanID uuid.UUID, available bool, note string) error

	// Rating Management
	UpdateRating(ctx context.Context, artisanID uuid.UUID, newRating float64, reviewCount int) error

	// Statistics
	GetArtisanStats(ctx context.Context, artisanID uuid.UUID) (*dto.ArtisanStatsResponse, error)
	GetDashboardStats(ctx context.Context, artisanID uuid.UUID) (*dto.ArtisanServiceDashboardResponse, error)

	// Batch Operations
	BatchUpdateAvailability(ctx context.Context, artisanIDs []uuid.UUID, available bool) error

	// Health
	HealthCheck(ctx context.Context) error
}

// artisanService implements artisanService
type artisanService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewartisanService creates a new artisan service
func NewArtisanService(repos *repository.Repositories, logger log.AllLogger) ArtisanService {
	return &artisanService{
		repos:  repos,
		logger: logger,
	}
}

// CreateArtisan creates a new artisan profile
func (s *artisanService) CreateArtisan(ctx context.Context, req *dto.CreateArtisanRequest) (*dto.ArtisanResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	// Check if artisan already exists for this user
	existing, _ := s.repos.Artisan.FindByUserID(ctx, req.UserID)
	if existing != nil {
		return nil, errors.NewConflictError("artisan profile already exists for this user")
	}

	artisan := &models.Artisan{
		UserID:             req.UserID,
		TenantID:           req.TenantID,
		Bio:                req.Bio,
		Specialization:     req.Specialization,
		YearsExperience:    req.YearsExperience,
		Certifications:     req.Certifications,
		CommissionRate:     req.CommissionRate,
		AutoAcceptBookings: req.AutoAcceptBookings,
		BookingLeadTime:    req.BookingLeadTime,
		MaxAdvanceBooking:  req.MaxAdvanceBooking,
		ServiceRadius:      req.ServiceRadius,
		IsAvailable:        true,
		Rating:             0,
		ReviewCount:        0,
	}

	// Set location if provided
	if req.Location != nil {
		artisan.Location = *req.Location
	}

	if err := s.repos.Artisan.Create(ctx, artisan); err != nil {
		return nil, errors.NewServiceError("ARTISAN_CREATE_FAILED", "failed to create artisan", err)
	}

	s.logger.Info("artisan created", "artisan_id", artisan.ID, "user_id", req.UserID)
	return dto.ToArtisanResponse(artisan), nil
}

// GetArtisan retrieves an artisan by ID
func (s *artisanService) GetArtisan(ctx context.Context, id uuid.UUID) (*dto.ArtisanResponse, error) {
	artisan, err := s.repos.Artisan.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("artisan not found")
		}
		return nil, errors.NewServiceError("ARTISAN_GET_FAILED", "failed to get artisan", err)
	}

	return dto.ToArtisanResponse(artisan), nil
}

// GetArtisanByUserID retrieves an artisan by user ID
func (s *artisanService) GetArtisanByUserID(ctx context.Context, userID uuid.UUID) (*dto.ArtisanResponse, error) {
	artisan, err := s.repos.Artisan.FindByUserID(ctx, userID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("artisan not found")
		}
		return nil, errors.NewServiceError("ARTISAN_GET_FAILED", "failed to get artisan", err)
	}

	return dto.ToArtisanResponse(artisan), nil
}

// UpdateArtisan updates an artisan profile
func (s *artisanService) UpdateArtisan(ctx context.Context, id uuid.UUID, req *dto.UpdateArtisanRequest) (*dto.ArtisanResponse, error) {
	artisan, err := s.repos.Artisan.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("artisan not found")
		}
		return nil, errors.NewServiceError("ARTISAN_GET_FAILED", "failed to get artisan", err)
	}

	// Apply updates
	if req.Bio != nil {
		artisan.Bio = *req.Bio
	}
	if req.Specialization != nil {
		artisan.Specialization = req.Specialization
	}
	if req.YearsExperience != nil {
		artisan.YearsExperience = *req.YearsExperience
	}
	if req.CommissionRate != nil {
		artisan.CommissionRate = *req.CommissionRate
	}
	if req.AutoAcceptBookings != nil {
		artisan.AutoAcceptBookings = *req.AutoAcceptBookings
	}
	if req.BookingLeadTime != nil {
		artisan.BookingLeadTime = *req.BookingLeadTime
	}
	if req.MaxAdvanceBooking != nil {
		artisan.MaxAdvanceBooking = *req.MaxAdvanceBooking
	}
	if req.Location != nil {
		artisan.Location = *req.Location
	}
	if req.ServiceRadius != nil {
		artisan.ServiceRadius = *req.ServiceRadius
	}

	if err := s.repos.Artisan.Update(ctx, artisan); err != nil {
		return nil, errors.NewServiceError("ARTISAN_UPDATE_FAILED", "failed to update artisan", err)
	}

	s.logger.Info("artisan updated", "artisan_id", id)
	return dto.ToArtisanResponse(artisan), nil
}

// DeleteArtisan deletes an artisan
func (s *artisanService) DeleteArtisan(ctx context.Context, id uuid.UUID) error {
	if err := s.repos.Artisan.Delete(ctx, id); err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("artisan not found")
		}
		return errors.NewServiceError("ARTISAN_DELETE_FAILED", "failed to delete artisan", err)
	}

	s.logger.Info("artisan deleted", "artisan_id", id)
	return nil
}

// ListArtisans lists artisans with filtering
func (s *artisanService) ListArtisans(ctx context.Context, filter dto.ArtisanFilter) (*dto.ArtisanListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize:     filter.PageSize,
	}
	pagination.Validate()

	var artisans []*models.Artisan
	var paginationResult repository.PaginationResult
	var err error

	// Handle nil tenant_id for platform admins
	if filter.TenantID == nil || *filter.TenantID == uuid.Nil {
		// Platform admin querying all artisans across all tenants
		// For now, return empty list as this feature needs to be implemented at repo level
		s.logger.Warn("platform admin query for all artisans not yet implemented at repository level")
		artisans = []*models.Artisan{}
		paginationResult = repository.PaginationResult{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: 0,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}
	} else {
		// Tenant-specific query
		artisans, paginationResult, err = s.repos.Artisan.FindByTenant(ctx, *filter.TenantID, pagination)
		if err != nil {
			return nil, errors.NewServiceError("ARTISAN_LIST_FAILED", "failed to list artisans", err)
		}
	}

	return &dto.ArtisanListResponse{
		Artisans:    dto.ToArtisanResponses(artisans),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// SearchArtisans searches artisans
func (s *artisanService) SearchArtisans(ctx context.Context, query string, filter dto.ArtisanFilter) (*dto.ArtisanListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}
	pagination.Validate()

	// Extract tenant ID from pointer, default to uuid.Nil for platform admins
	var tenantID uuid.UUID
	if filter.TenantID != nil {
		tenantID = *filter.TenantID
	}

	artisans, paginationResult, err := s.repos.Artisan.Search(ctx, tenantID, query, pagination)
	if err != nil {
		return nil, errors.NewServiceError("ARTISAN_SEARCH_FAILED", "failed to search artisans", err)
	}

	return &dto.ArtisanListResponse{
		Artisans:    dto.ToArtisanResponses(artisans),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetAvailableArtisans gets available artisans
func (s *artisanService) GetAvailableArtisans(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.ArtisanListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	artisans, paginationResult, err := s.repos.Artisan.FindAvailable(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("ARTISAN_AVAILABLE_FAILED", "failed to get available artisans", err)
	}

	return &dto.ArtisanListResponse{
		Artisans:    dto.ToArtisanResponses(artisans),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetArtisansBySpecialization gets artisans by specialization
func (s *artisanService) GetArtisansBySpecialization(ctx context.Context, tenantID uuid.UUID, specialization string, page, pageSize int) (*dto.ArtisanListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	artisans, paginationResult, err := s.repos.Artisan.FindBySpecialization(ctx, tenantID, specialization, pagination)
	if err != nil {
		return nil, errors.NewServiceError("ARTISAN_SPECIALIZATION_FAILED", "failed to get artisans by specialization", err)
	}

	return &dto.ArtisanListResponse{
		Artisans:    dto.ToArtisanResponses(artisans),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetTopRatedArtisans gets top rated artisans
func (s *artisanService) GetTopRatedArtisans(ctx context.Context, tenantID uuid.UUID, limit int) ([]*dto.ArtisanResponse, error) {
	artisans, err := s.repos.Artisan.FindTopRated(ctx, tenantID, limit)
	if err != nil {
		return nil, errors.NewServiceError("ARTISAN_TOP_RATED_FAILED", "failed to get top rated artisans", err)
	}

	return dto.ToArtisanResponses(artisans), nil
}

// FindNearbyArtisans finds nearby artisans
func (s *artisanService) FindNearbyArtisans(ctx context.Context, tenantID uuid.UUID, latitude, longitude float64, radiusKm int) ([]*dto.ArtisanResponse, error) {
	artisans, err := s.repos.Artisan.FindNearby(ctx, tenantID, latitude, longitude, radiusKm)
	if err != nil {
		return nil, errors.NewServiceError("ARTISAN_NEARBY_FAILED", "failed to find nearby artisans", err)
	}

	return dto.ToArtisanResponses(artisans), nil
}

// UpdateAvailability updates artisan availability
func (s *artisanService) UpdateAvailability(ctx context.Context, artisanID uuid.UUID, available bool, note string) error {
	if err := s.repos.Artisan.UpdateAvailability(ctx, artisanID, available, note); err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("artisan not found")
		}
		return errors.NewServiceError("ARTISAN_AVAILABILITY_UPDATE_FAILED", "failed to update availability", err)
	}

	s.logger.Info("artisan availability updated", "artisan_id", artisanID, "available", available)
	return nil
}

// UpdateRating updates artisan rating
func (s *artisanService) UpdateRating(ctx context.Context, artisanID uuid.UUID, newRating float64, reviewCount int) error {
	if err := s.repos.Artisan.UpdateRating(ctx, artisanID, newRating, reviewCount); err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("artisan not found")
		}
		return errors.NewServiceError("ARTISAN_RATING_UPDATE_FAILED", "failed to update rating", err)
	}

	s.logger.Info("artisan rating updated", "artisan_id", artisanID, "rating", newRating)
	return nil
}

// GetArtisanStats gets artisan statistics
func (s *artisanService) GetArtisanStats(ctx context.Context, artisanID uuid.UUID) (*dto.ArtisanStatsResponse, error) {
	stats, err := s.repos.Artisan.GetStatistics(ctx, artisanID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("artisan not found")
		}
		return nil, errors.NewServiceError("ARTISAN_STATS_FAILED", "failed to get artisan stats", err)
	}

	return &dto.ArtisanStatsResponse{
		ArtisanID:         artisanID,
		TotalBookings:     getInt64(stats, "total_bookings"),
		CompletedBookings: getInt64(stats, "completed_bookings"),
		CancelledBookings: getInt64(stats, "cancelled_bookings"),
		ActiveProjects:    getInt64(stats, "active_projects"),
		TotalProjects:     getInt64(stats, "total_projects"),
		AverageRating:     getFloat64(stats, "average_rating"),
		ReviewCount:       getInt64(stats, "review_count"),
		TotalEarnings:     getFloat64(stats, "total_earnings"),
		CompletionRate:    getFloat64(stats, "completion_rate"),
		TotalServices:     getInt64(stats, "total_services"),
		YearsExperience:   getInt(stats, "years_experience"),
		IsAvailable:       getBool(stats, "is_available"),
	}, nil
}

// GetDashboardStats gets dashboard statistics
func (s *artisanService) GetDashboardStats(ctx context.Context, artisanID uuid.UUID) (*dto.ArtisanServiceDashboardResponse, error) {
	stats, err := s.GetArtisanStats(ctx, artisanID)
	if err != nil {
		return nil, err
	}

	return &dto.ArtisanServiceDashboardResponse{
		Stats:            stats,
		RecentBookings:   []*dto.BookingResponse{}, // Would fetch from booking service
		UpcomingBookings: []*dto.BookingResponse{}, // Would fetch from booking service
		RecentReviews:    []*dto.ReviewResponse{},  // Would fetch from review service
	}, nil
}

// BatchUpdateAvailability updates availability for multiple artisans
func (s *artisanService) BatchUpdateAvailability(ctx context.Context, artisanIDs []uuid.UUID, available bool) error {
	if err := s.repos.Artisan.BatchUpdateAvailability(ctx, artisanIDs, available); err != nil {
		return errors.NewServiceError("ARTISAN_BATCH_UPDATE_FAILED", "failed to batch update availability", err)
	}

	s.logger.Info("batch availability updated", "count", len(artisanIDs), "available", available)
	return nil
}

// HealthCheck performs a health check
func (s *artisanService) HealthCheck(ctx context.Context) error {
	_, err := s.repos.Artisan.CountByTenant(ctx, uuid.New())
	if err != nil && !errors.IsNotFoundError(err) {
		return errors.NewServiceError("HEALTH_CHECK_FAILED", "artisan service health check failed", err)
	}
	return nil
}

// Helper functions
func getInt64(m map[string]any, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		}
	}
	return 0
}

func getFloat64(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int64:
			return float64(val)
		case int:
			return float64(val)
		}
	}
	return 0
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if val, ok := v.(bool); ok {
			return val
		}
	}
	return false
}
