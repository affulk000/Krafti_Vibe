package repository

import (
	"context"
	"fmt"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArtisanRepository defines the interface for artisan repository operations
type ArtisanRepository interface {
	BaseRepository[models.Artisan]

	// FindByUserID retrieves an artisan by user ID
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Artisan, error)

	// FindByTenant retrieves all artisans for a tenant
	FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error)

	// FindAvailable retrieves available artisans
	FindAvailable(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error)

	// FindBySpecialization retrieves artisans by specialization
	FindBySpecialization(ctx context.Context, tenantID uuid.UUID, specialization string, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error)

	// FindTopRated retrieves top-rated artisans
	FindTopRated(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Artisan, error)

	// UpdateRating updates an artisan's rating
	UpdateRating(ctx context.Context, artisanID uuid.UUID, newRating float64, reviewCount int) error

	// UpdateAvailability updates availability status
	UpdateAvailability(ctx context.Context, artisanID uuid.UUID, isAvailable bool, note string) error

	// Search searches artisans by name, bio, or specialization
	Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error)

	// FindNearby finds artisans near a location (requires PostGIS)
	FindNearby(ctx context.Context, tenantID uuid.UUID, latitude, longitude float64, radiusKm int) ([]*models.Artisan, error)

	// IncrementBookingCount increments total bookings counter
	IncrementBookingCount(ctx context.Context, artisanID uuid.UUID) error

	// GetStatistics retrieves statistics for an artisan
	GetStatistics(ctx context.Context, artisanID uuid.UUID) (map[string]any, error)

	// BatchUpdateAvailability updates availability for multiple artisans
	BatchUpdateAvailability(ctx context.Context, artisanIDs []uuid.UUID, isAvailable bool) error

	// CountByTenant counts artisans for a tenant
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// artisanRepository implements ArtisanRepository
type artisanRepository struct {
	BaseRepository[models.Artisan]
	db     *gorm.DB
	logger log.AllLogger
}

// NewArtisanRepository creates a new artisan repository
func NewArtisanRepository(db *gorm.DB, config ...RepositoryConfig) ArtisanRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.Artisan](db, cfg)

	return &artisanRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
	}
}

// FindByUserID retrieves an artisan by user ID
func (r *artisanRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Artisan, error) {
	if userID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	var artisan models.Artisan
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Services").
		Where("user_id = ?", userID).
		First(&artisan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "artisan not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to find artisan by user_id", "user_id", userID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find artisan", err)
	}

	return &artisan, nil
}

// FindByTenant retrieves all artisans for a tenant
func (r *artisanRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count artisans", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count artisans", err)
	}

	// Find paginated results
	var artisans []*models.Artisan
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("rating DESC, review_count DESC").
		Find(&artisans).Error; err != nil {
		r.logger.Error("failed to find artisans", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find artisans", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return artisans, paginationResult, nil
}

// FindAvailable retrieves available artisans
func (r *artisanRepository) FindAvailable(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("tenant_id = ? AND is_available = ?", tenantID, true).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count available artisans", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count artisans", err)
	}

	// Find paginated results
	var artisans []*models.Artisan
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Services").
		Where("tenant_id = ? AND is_available = ?", tenantID, true).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("rating DESC, review_count DESC").
		Find(&artisans).Error; err != nil {
		r.logger.Error("failed to find available artisans", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find artisans", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return artisans, paginationResult, nil
}

// FindBySpecialization retrieves artisans by specialization
func (r *artisanRepository) FindBySpecialization(ctx context.Context, tenantID uuid.UUID, specialization string, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if specialization == "" {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "specialization cannot be empty", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("tenant_id = ? AND ? = ANY(specialization)", tenantID, specialization).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count artisans by specialization", "tenant_id", tenantID, "specialization", specialization, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count artisans", err)
	}

	// Find paginated results
	var artisans []*models.Artisan
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ? AND ? = ANY(specialization)", tenantID, specialization).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("rating DESC, review_count DESC").
		Find(&artisans).Error; err != nil {
		r.logger.Error("failed to find artisans by specialization", "tenant_id", tenantID, "specialization", specialization, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find artisans", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return artisans, paginationResult, nil
}

// FindTopRated retrieves top-rated artisans
func (r *artisanRepository) FindTopRated(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Artisan, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if limit <= 0 {
		limit = 10
	}

	var artisans []*models.Artisan
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ? AND is_available = ? AND rating >= ?", tenantID, true, 4.0).
		Order("rating DESC, review_count DESC").
		Limit(limit).
		Find(&artisans).Error; err != nil {
		r.logger.Error("failed to find top-rated artisans", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find top-rated artisans", err)
	}

	return artisans, nil
}

// UpdateRating updates an artisan's rating
func (r *artisanRepository) UpdateRating(ctx context.Context, artisanID uuid.UUID, newRating float64, reviewCount int) error {
	if artisanID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("id = ?", artisanID).
		Updates(map[string]any{
			"rating":       newRating,
			"review_count": reviewCount,
		})

	if result.Error != nil {
		r.logger.Error("failed to update artisan rating", "artisan_id", artisanID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update rating", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "artisan not found", errors.ErrNotFound)
	}

	return nil
}

// UpdateAvailability updates availability status
func (r *artisanRepository) UpdateAvailability(ctx context.Context, artisanID uuid.UUID, isAvailable bool, note string) error {
	if artisanID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("id = ?", artisanID).
		Updates(map[string]any{
			"is_available":      isAvailable,
			"availability_note": note,
		})

	if result.Error != nil {
		r.logger.Error("failed to update availability", "artisan_id", artisanID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update availability", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "artisan not found", errors.ErrNotFound)
	}

	return nil
}

// Search searches artisans by name, bio, or specialization
func (r *artisanRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	if query == "" {
		return r.FindAvailable(ctx, tenantID, pagination)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Joins("JOIN users ON users.id = artisans.user_id").
		Where("artisans.tenant_id = ?", tenantID).
		Where("(artisans.bio ILIKE ? OR CAST(artisans.specialization AS TEXT) ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?)",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count search results", "tenant_id", tenantID, "query", query, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count artisans", err)
	}

	// Find paginated results
	var artisans []*models.Artisan
	if err := r.db.WithContext(ctx).
		Preload("User").
		Joins("JOIN users ON users.id = artisans.user_id").
		Where("artisans.tenant_id = ?", tenantID).
		Where("(artisans.bio ILIKE ? OR CAST(artisans.specialization AS TEXT) ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?)",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("artisans.rating DESC, artisans.review_count DESC").
		Find(&artisans).Error; err != nil {
		r.logger.Error("failed to search artisans", "tenant_id", tenantID, "query", query, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search artisans", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return artisans, paginationResult, nil
}

// FindNearby finds artisans near a location (simplified without PostGIS)
func (r *artisanRepository) FindNearby(ctx context.Context, tenantID uuid.UUID, latitude, longitude float64, radiusKm int) ([]*models.Artisan, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	// Simplified implementation without PostGIS
	// In production, use ST_DWithin for proper geospatial queries
	var artisans []*models.Artisan
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ? AND is_available = ? AND service_radius >= ?", tenantID, true, 0).
		Order("rating DESC").
		Limit(50).
		Find(&artisans).Error; err != nil {
		r.logger.Error("failed to find nearby artisans", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find nearby artisans", err)
	}

	// TODO: Implement proper distance calculation with PostGIS:
	// SELECT *, ST_Distance(
	//   ST_MakePoint(location->>'longitude'::float, location->>'latitude'::float)::geography,
	//   ST_MakePoint(?, ?)::geography
	// ) AS distance
	// FROM artisans
	// WHERE ST_DWithin(
	//   ST_MakePoint(location->>'longitude'::float, location->>'latitude'::float)::geography,
	//   ST_MakePoint(?, ?)::geography,
	//   ? * 1000
	// )
	// ORDER BY distance ASC

	return artisans, nil
}

// IncrementBookingCount increments total bookings counter
func (r *artisanRepository) IncrementBookingCount(ctx context.Context, artisanID uuid.UUID) error {
	if artisanID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("id = ?", artisanID).
		UpdateColumn("total_bookings", gorm.Expr("total_bookings + ?", 1))

	if result.Error != nil {
		r.logger.Error("failed to increment booking count", "artisan_id", artisanID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment booking count", result.Error)
	}

	return nil
}

// GetStatistics retrieves statistics for an artisan
func (r *artisanRepository) GetStatistics(ctx context.Context, artisanID uuid.UUID) (map[string]any, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	var artisan models.Artisan
	if err := r.db.WithContext(ctx).
		Preload("Bookings").
		Preload("Projects").
		Preload("Reviews").
		Preload("Services").
		First(&artisan, artisanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "artisan not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get artisan statistics", "artisan_id", artisanID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to get statistics", err)
	}

	// Calculate statistics
	var completedBookings, cancelledBookings int
	var totalEarnings float64

	for _, booking := range artisan.Bookings {
		switch booking.Status {
		case models.BookingStatusCompleted:
			completedBookings++
			totalEarnings += booking.TotalPrice
		case models.BookingStatusCancelled:
			cancelledBookings++
		}
	}

	var activeProjects int
	for _, project := range artisan.Projects {
		if project.Status == models.ProjectStatusInProgress {
			activeProjects++
		}
	}

	var completionRate float64
	if artisan.TotalBookings > 0 {
		completionRate = (float64(completedBookings) / float64(artisan.TotalBookings)) * 100
	}

	stats := map[string]any{
		"total_bookings":     artisan.TotalBookings,
		"completed_bookings": completedBookings,
		"cancelled_bookings": cancelledBookings,
		"active_projects":    activeProjects,
		"total_projects":     len(artisan.Projects),
		"average_rating":     artisan.Rating,
		"review_count":       artisan.ReviewCount,
		"total_earnings":     totalEarnings,
		"completion_rate":    completionRate,
		"total_services":     len(artisan.Services),
		"years_experience":   artisan.YearsExperience,
		"is_available":       artisan.IsAvailable,
	}

	return stats, nil
}

// FindWithFilters finds artisans with multiple filters
func (r *artisanRepository) FindWithFilters(ctx context.Context, filters map[string]any, pagination PaginationParams) ([]*models.Artisan, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.Artisan{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "tenant_id":
			query = query.Where("tenant_id = ?", value)
		case "is_available":
			query = query.Where("is_available = ?", value)
		case "min_rating":
			query = query.Where("rating >= ?", value)
		case "min_experience":
			query = query.Where("years_experience >= ?", value)
		case "specialization":
			query = query.Where("? = ANY(specialization)", value)
		}
	}

	// Count total
	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count artisans", err)
	}

	// Find paginated results
	var artisans []*models.Artisan
	if err := query.
		Preload("User").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("rating DESC, review_count DESC").
		Find(&artisans).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find artisans", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return artisans, paginationResult, nil
}

// CountByTenant counts artisans for a tenant
func (r *artisanRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error; err != nil {
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count artisans", err)
	}
	return count, nil
}

// CountAvailableByTenant counts available artisans for a tenant
func (r *artisanRepository) CountAvailableByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("tenant_id = ? AND is_available = ?", tenantID, true).
		Count(&count).Error; err != nil {
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count available artisans", err)
	}
	return count, nil
}

// UpdateDashboardStats updates dashboard statistics for an artisan
func (r *artisanRepository) UpdateDashboardStats(ctx context.Context, artisanID uuid.UUID, stats map[string]any) error {
	if artisanID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := make(map[string]any)

	if projectsActive, ok := stats["projects_active"]; ok {
		updates["dashboard_projects_active"] = projectsActive
	}
	if tasksOpen, ok := stats["tasks_open"]; ok {
		updates["dashboard_tasks_open"] = tasksOpen
	}
	if tasksOverdue, ok := stats["tasks_overdue"]; ok {
		updates["dashboard_tasks_overdue"] = tasksOverdue
	}
	if nextDueAt, ok := stats["next_due_at"]; ok {
		updates["dashboard_next_due_at"] = nextDueAt
	}

	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("id = ?", artisanID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update dashboard stats", result.Error)
	}

	return nil
}

// BatchUpdateAvailability updates availability for multiple artisans
func (r *artisanRepository) BatchUpdateAvailability(ctx context.Context, artisanIDs []uuid.UUID, isAvailable bool) error {
	if len(artisanIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Artisan{}).
		Where("id IN ?", artisanIDs).
		Update("is_available", isAvailable)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED",
			fmt.Sprintf("failed to batch update availability for %d artisans", len(artisanIDs)),
			result.Error)
	}

	r.logger.Info("batch updated artisan availability",
		"count", result.RowsAffected,
		"is_available", isAvailable)

	return nil
}
