package repository

import (
	"context"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AvailabilityRepository defines the interface for availability data operations
type AvailabilityRepository interface {
	// CRUD operations
	Create(ctx context.Context, availability *models.Availability) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Availability, error)
	Update(ctx context.Context, availability *models.Availability) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	ListByArtisan(ctx context.Context, artisanID uuid.UUID, page, pageSize int) ([]*models.Availability, int64, error)
	ListByArtisanAndType(ctx context.Context, artisanID uuid.UUID, availabilityType models.AvailabilityType, page, pageSize int) ([]*models.Availability, int64, error)
	ListByArtisanAndDateRange(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*models.Availability, int64, error)
	GetByArtisanAndDayOfWeek(ctx context.Context, artisanID uuid.UUID, dayOfWeek int) ([]*models.Availability, error)

	// Conflict detection
	FindConflicts(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) ([]*models.Availability, error)

	// Availability checks
	CheckAvailability(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time) (bool, error)

	// Bulk operations
	BulkCreate(ctx context.Context, availabilities []*models.Availability) error
	DeleteByArtisanAndType(ctx context.Context, artisanID uuid.UUID, availabilityType models.AvailabilityType) error
}

type availabilityRepository struct {
	db *gorm.DB
}

// NewAvailabilityRepository creates a new availability repository
func NewAvailabilityRepository(db *gorm.DB) AvailabilityRepository {
	return &availabilityRepository{db: db}
}

func (r *availabilityRepository) Create(ctx context.Context, availability *models.Availability) error {
	return r.db.WithContext(ctx).Create(availability).Error
}

func (r *availabilityRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Availability, error) {
	var availability models.Availability
	err := r.db.WithContext(ctx).
		Preload("Artisan").
		First(&availability, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &availability, nil
}

func (r *availabilityRepository) Update(ctx context.Context, availability *models.Availability) error {
	return r.db.WithContext(ctx).Save(availability).Error
}

func (r *availabilityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Availability{}, "id = ?", id).Error
}

func (r *availabilityRepository) ListByArtisan(ctx context.Context, artisanID uuid.UUID, page, pageSize int) ([]*models.Availability, int64, error) {
	var availabilities []*models.Availability
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Availability{}).
		Where("artisan_id = ?", artisanID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.
		Preload("Artisan").
		Order("day_of_week ASC, start_time ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&availabilities).Error

	return availabilities, total, err
}

func (r *availabilityRepository) ListByArtisanAndType(ctx context.Context, artisanID uuid.UUID, availabilityType models.AvailabilityType, page, pageSize int) ([]*models.Availability, int64, error) {
	var availabilities []*models.Availability
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Availability{}).
		Where("artisan_id = ? AND type = ?", artisanID, availabilityType)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.
		Preload("Artisan").
		Order("day_of_week ASC, start_time ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&availabilities).Error

	return availabilities, total, err
}

func (r *availabilityRepository) ListByArtisanAndDateRange(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*models.Availability, int64, error) {
	var availabilities []*models.Availability
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Availability{}).
		Where("artisan_id = ?", artisanID).
		Where("(date IS NULL OR (date >= ? AND date <= ?))", startDate, endDate)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.
		Preload("Artisan").
		Order("date ASC, start_time ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&availabilities).Error

	return availabilities, total, err
}

func (r *availabilityRepository) GetByArtisanAndDayOfWeek(ctx context.Context, artisanID uuid.UUID, dayOfWeek int) ([]*models.Availability, error) {
	var availabilities []*models.Availability
	err := r.db.WithContext(ctx).
		Where("artisan_id = ? AND day_of_week = ?", artisanID, dayOfWeek).
		Preload("Artisan").
		Order("start_time ASC").
		Find(&availabilities).Error

	return availabilities, err
}

func (r *availabilityRepository) FindConflicts(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) ([]*models.Availability, error) {
	var availabilities []*models.Availability

	query := r.db.WithContext(ctx).
		Where("artisan_id = ?", artisanID).
		Where("start_time < ? AND end_time > ?", endTime, startTime)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.
		Preload("Artisan").
		Find(&availabilities).Error

	return availabilities, err
}

func (r *availabilityRepository) CheckAvailability(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Availability{}).
		Where("artisan_id = ?", artisanID).
		Where("type = ?", models.AvailabilityTypeRegular).
		Where("start_time <= ? AND end_time >= ?", startTime, endTime).
		Count(&count).Error

	return count > 0, err
}

func (r *availabilityRepository) BulkCreate(ctx context.Context, availabilities []*models.Availability) error {
	return r.db.WithContext(ctx).Create(availabilities).Error
}

func (r *availabilityRepository) DeleteByArtisanAndType(ctx context.Context, artisanID uuid.UUID, availabilityType models.AvailabilityType) error {
	return r.db.WithContext(ctx).
		Where("artisan_id = ? AND type = ?", artisanID, availabilityType).
		Delete(&models.Availability{}).Error
}
