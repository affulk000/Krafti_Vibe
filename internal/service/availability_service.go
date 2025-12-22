package service

import (
	"context"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// AvailabilityService defines the interface for availability operations
type AvailabilityService interface {
	// CRUD operations
	CreateAvailability(ctx context.Context, tenantID uuid.UUID, req *dto.CreateAvailabilitySlotRequest) (*dto.AvailabilitySlotResponse, error)
	GetAvailability(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.AvailabilitySlotResponse, error)
	UpdateAvailability(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateAvailabilitySlotRequest) (*dto.AvailabilitySlotResponse, error)
	DeleteAvailability(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// Query operations
	ListAvailabilities(ctx context.Context, filter *dto.AvailabilitySlotFilter, tenantID uuid.UUID) (*dto.AvailabilitySlotListResponse, error)
	ListByType(ctx context.Context, artisanID uuid.UUID, availabilityType models.AvailabilityType, tenantID uuid.UUID, page, pageSize int) (*dto.AvailabilitySlotListResponse, error)
	GetByDayOfWeek(ctx context.Context, artisanID uuid.UUID, dayOfWeek int, tenantID uuid.UUID) ([]*dto.AvailabilitySlotResponse, error)
	GetWeeklySchedule(ctx context.Context, artisanID uuid.UUID, tenantID uuid.UUID, weekStart time.Time) (*dto.WeeklyScheduleResponse, error)

	// Availability checks
	CheckAvailability(ctx context.Context, req *dto.CheckAvailabilitySlotRequest, tenantID uuid.UUID) (*dto.AvailabilitySlotCheckResponse, error)

	// Bulk operations
	BulkCreateAvailability(ctx context.Context, tenantID uuid.UUID, req *dto.BulkCreateAvailabilitySlotRequest) ([]*dto.AvailabilitySlotResponse, error)
	DeleteByType(ctx context.Context, artisanID uuid.UUID, availabilityType models.AvailabilityType, tenantID uuid.UUID) error
}

type availabilityService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewAvailabilityService creates a new availability service
func NewAvailabilityService(repos *repository.Repositories, logger log.AllLogger) AvailabilityService {
	return &availabilityService{
		repos:  repos,
		logger: logger,
	}
}

func (s *availabilityService) CreateAvailability(ctx context.Context, tenantID uuid.UUID, req *dto.CreateAvailabilitySlotRequest) (*dto.AvailabilitySlotResponse, error) {
	// Verify artisan exists
	artisan, err := s.repos.Artisan.GetByID(ctx, req.ArtisanID)
	if err != nil {
		s.logger.Error("failed to get artisan", "artisan_id", req.ArtisanID, "error", err)
		return nil, errors.NewNotFoundError("artisan")
	}

	// Verify artisan belongs to same tenant
	if artisan.TenantID != tenantID {
		return nil, errors.NewForbiddenError("artisan does not belong to your tenant")
	}

	// Check for conflicts
	conflicts, err := s.repos.Availability.FindConflicts(ctx, req.ArtisanID, req.StartTime, req.EndTime, nil)
	if err != nil {
		s.logger.Error("failed to check conflicts", "error", err)
		return nil, errors.NewInternalError("failed to check availability conflicts", err)
	}

	if len(conflicts) > 0 {
		return nil, errors.NewConflictError("availability slot conflicts with existing schedule")
	}

	// Create availability
	availability := &models.Availability{
		TenantID:    tenantID,
		ArtisanID:   req.ArtisanID,
		Type:        req.Type,
		DayOfWeek:   req.DayOfWeek,
		Date:        req.Date,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsRecurring: req.IsRecurring,
		RecurUntil:  req.RecurUntil,
		Notes:       req.Notes,
	}

	if err := s.repos.Availability.Create(ctx, availability); err != nil {
		s.logger.Error("failed to create availability", "error", err)
		return nil, errors.NewInternalError("failed to create availability", err)
	}

	// Reload with relationships
	created, err := s.repos.Availability.GetByID(ctx, availability.ID)
	if err != nil {
		s.logger.Error("failed to reload availability", "error", err)
		return nil, errors.NewInternalError("failed to reload availability", err)
	}

	return dto.ToAvailabilitySlotResponse(created), nil
}

func (s *availabilityService) GetAvailability(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.AvailabilitySlotResponse, error) {
	availability, err := s.repos.Availability.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get availability", "id", id, "error", err)
		return nil, errors.NewNotFoundError("availability")
	}

	// Verify tenant access
	if availability.TenantID != tenantID {
		return nil, errors.NewForbiddenError("availability does not belong to your tenant")
	}

	return dto.ToAvailabilitySlotResponse(availability), nil
}

func (s *availabilityService) UpdateAvailability(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateAvailabilitySlotRequest) (*dto.AvailabilitySlotResponse, error) {
	// Get existing availability
	availability, err := s.repos.Availability.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get availability", "id", id, "error", err)
		return nil, errors.NewNotFoundError("availability")
	}

	// Verify tenant access
	if availability.TenantID != tenantID {
		return nil, errors.NewForbiddenError("availability does not belong to your tenant")
	}

	// Update fields
	if req.Type != nil {
		availability.Type = *req.Type
	}
	if req.DayOfWeek != nil {
		availability.DayOfWeek = req.DayOfWeek
	}
	if req.Date != nil {
		availability.Date = req.Date
	}
	if req.StartTime != nil {
		availability.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		availability.EndTime = *req.EndTime
	}
	if req.IsRecurring != nil {
		availability.IsRecurring = *req.IsRecurring
	}
	if req.RecurUntil != nil {
		availability.RecurUntil = req.RecurUntil
	}
	if req.Notes != nil {
		availability.Notes = *req.Notes
	}

	// Check for conflicts (excluding current availability)
	conflicts, err := s.repos.Availability.FindConflicts(ctx, availability.ArtisanID, availability.StartTime, availability.EndTime, &id)
	if err != nil {
		s.logger.Error("failed to check conflicts", "error", err)
		return nil, errors.NewInternalError("failed to check availability conflicts", err)
	}

	if len(conflicts) > 0 {
		return nil, errors.NewConflictError("availability slot conflicts with existing schedule")
	}

	// Update availability
	if err := s.repos.Availability.Update(ctx, availability); err != nil {
		s.logger.Error("failed to update availability", "error", err)
		return nil, errors.NewInternalError("failed to update availability", err)
	}

	// Reload with relationships
	updated, err := s.repos.Availability.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to reload availability", "error", err)
		return nil, errors.NewInternalError("failed to reload availability", err)
	}

	return dto.ToAvailabilitySlotResponse(updated), nil
}

func (s *availabilityService) DeleteAvailability(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// Get availability to verify tenant access
	availability, err := s.repos.Availability.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get availability", "id", id, "error", err)
		return errors.NewNotFoundError("availability")
	}

	// Verify tenant access
	if availability.TenantID != tenantID {
		return errors.NewForbiddenError("availability does not belong to your tenant")
	}

	if err := s.repos.Availability.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete availability", "error", err)
		return errors.NewInternalError("failed to delete availability", err)
	}

	return nil
}

func (s *availabilityService) ListAvailabilities(ctx context.Context, filter *dto.AvailabilitySlotFilter, tenantID uuid.UUID) (*dto.AvailabilitySlotListResponse, error) {
	var availabilities []*models.Availability
	var total int64
	var err error

	if filter.Type != nil {
		availabilities, total, err = s.repos.Availability.ListByArtisanAndType(ctx, filter.ArtisanID, *filter.Type, filter.Page, filter.PageSize)
	} else if filter.StartDate != nil && filter.EndDate != nil {
		availabilities, total, err = s.repos.Availability.ListByArtisanAndDateRange(ctx, filter.ArtisanID, *filter.StartDate, *filter.EndDate, filter.Page, filter.PageSize)
	} else {
		availabilities, total, err = s.repos.Availability.ListByArtisan(ctx, filter.ArtisanID, filter.Page, filter.PageSize)
	}

	if err != nil {
		s.logger.Error("failed to list availabilities", "error", err)
		return nil, errors.NewInternalError("failed to list availabilities", err)
	}

	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	return &dto.AvailabilitySlotListResponse{
		Availabilities: dto.ToAvailabilitySlotResponses(availabilities),
		Page:           filter.Page,
		PageSize:       filter.PageSize,
		TotalItems:     total,
		TotalPages:     totalPages,
		HasNext:        filter.Page < totalPages,
		HasPrevious:    filter.Page > 1,
	}, nil
}

func (s *availabilityService) ListByType(ctx context.Context, artisanID uuid.UUID, availabilityType models.AvailabilityType, tenantID uuid.UUID, page, pageSize int) (*dto.AvailabilitySlotListResponse, error) {
	availabilities, total, err := s.repos.Availability.ListByArtisanAndType(ctx, artisanID, availabilityType, page, pageSize)
	if err != nil {
		s.logger.Error("failed to list availabilities by type", "error", err)
		return nil, errors.NewInternalError("failed to list availabilities", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.AvailabilitySlotListResponse{
		Availabilities: dto.ToAvailabilitySlotResponses(availabilities),
		Page:           page,
		PageSize:       pageSize,
		TotalItems:     total,
		TotalPages:     totalPages,
		HasNext:        page < totalPages,
		HasPrevious:    page > 1,
	}, nil
}

func (s *availabilityService) GetByDayOfWeek(ctx context.Context, artisanID uuid.UUID, dayOfWeek int, tenantID uuid.UUID) ([]*dto.AvailabilitySlotResponse, error) {
	availabilities, err := s.repos.Availability.GetByArtisanAndDayOfWeek(ctx, artisanID, dayOfWeek)
	if err != nil {
		s.logger.Error("failed to get availabilities by day of week", "error", err)
		return nil, errors.NewInternalError("failed to get availabilities", err)
	}

	return dto.ToAvailabilitySlotResponses(availabilities), nil
}

func (s *availabilityService) GetWeeklySchedule(ctx context.Context, artisanID uuid.UUID, tenantID uuid.UUID, weekStart time.Time) (*dto.WeeklyScheduleResponse, error) {
	// Calculate week end (7 days from start)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Get all availabilities for the week
	availabilities, _, err := s.repos.Availability.ListByArtisanAndDateRange(ctx, artisanID, weekStart, weekEnd, 1, 1000)
	if err != nil {
		s.logger.Error("failed to get weekly schedule", "error", err)
		return nil, errors.NewInternalError("failed to get weekly schedule", err)
	}

	// Organize by day
	schedule := make(map[string][]*dto.AvailabilitySlotResponse)
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	for _, availability := range availabilities {
		var dayName string
		if availability.DayOfWeek != nil {
			dayName = dayNames[*availability.DayOfWeek]
		} else if availability.Date != nil {
			dayName = dayNames[availability.Date.Weekday()]
		}

		if dayName != "" {
			schedule[dayName] = append(schedule[dayName], dto.ToAvailabilitySlotResponse(availability))
		}
	}

	return &dto.WeeklyScheduleResponse{
		ArtisanID: artisanID,
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Schedule:  schedule,
	}, nil
}

func (s *availabilityService) CheckAvailability(ctx context.Context, req *dto.CheckAvailabilitySlotRequest, tenantID uuid.UUID) (*dto.AvailabilitySlotCheckResponse, error) {
	// Find conflicts
	conflicts, err := s.repos.Availability.FindConflicts(ctx, req.ArtisanID, req.StartTime, req.EndTime, nil)
	if err != nil {
		s.logger.Error("failed to check availability", "error", err)
		return nil, errors.NewInternalError("failed to check availability", err)
	}

	isAvailable := len(conflicts) == 0

	return &dto.AvailabilitySlotCheckResponse{
		IsAvailable:      isAvailable,
		ConflictingSlots: dto.ToAvailabilitySlotResponses(conflicts),
	}, nil
}

func (s *availabilityService) BulkCreateAvailability(ctx context.Context, tenantID uuid.UUID, req *dto.BulkCreateAvailabilitySlotRequest) ([]*dto.AvailabilitySlotResponse, error) {
	// Verify artisan exists
	artisan, err := s.repos.Artisan.GetByID(ctx, req.ArtisanID)
	if err != nil {
		s.logger.Error("failed to get artisan", "artisan_id", req.ArtisanID, "error", err)
		return nil, errors.NewNotFoundError("artisan")
	}

	// Verify artisan belongs to same tenant
	if artisan.TenantID != tenantID {
		return nil, errors.NewForbiddenError("artisan does not belong to your tenant")
	}

	// Create availability for each day
	var availabilities []*models.Availability
	for _, dayOfWeek := range req.DaysOfWeek {
		availability := &models.Availability{
			TenantID:    tenantID,
			ArtisanID:   req.ArtisanID,
			Type:        req.Type,
			DayOfWeek:   &dayOfWeek,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			IsRecurring: req.IsRecurring,
			RecurUntil:  req.RecurUntil,
			Notes:       req.Notes,
		}
		availabilities = append(availabilities, availability)
	}

	if err := s.repos.Availability.BulkCreate(ctx, availabilities); err != nil {
		s.logger.Error("failed to bulk create availabilities", "error", err)
		return nil, errors.NewInternalError("failed to create availabilities", err)
	}

	return dto.ToAvailabilitySlotResponses(availabilities), nil
}

func (s *availabilityService) DeleteByType(ctx context.Context, artisanID uuid.UUID, availabilityType models.AvailabilityType, tenantID uuid.UUID) error {
	if err := s.repos.Availability.DeleteByArtisanAndType(ctx, artisanID, availabilityType); err != nil {
		s.logger.Error("failed to delete availabilities by type", "error", err)
		return errors.NewInternalError("failed to delete availabilities", err)
	}

	return nil
}
