package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Availability Request DTOs
// ============================================================================

// CreateAvailabilitySlotRequest represents a request to create an availability slot
type CreateAvailabilitySlotRequest struct {
	ArtisanID   uuid.UUID               `json:"artisan_id" validate:"required"`
	Type        models.AvailabilityType `json:"type" validate:"required"`
	DayOfWeek   *int                    `json:"day_of_week,omitempty"`
	Date        *time.Time              `json:"date,omitempty"`
	StartTime   time.Time               `json:"start_time" validate:"required"`
	EndTime     time.Time               `json:"end_time" validate:"required,gtfield=StartTime"`
	IsRecurring bool                    `json:"is_recurring"`
	RecurUntil  *time.Time              `json:"recur_until,omitempty"`
	Notes       string                  `json:"notes,omitempty"`
}

// UpdateAvailabilitySlotRequest represents a request to update an availability slot
type UpdateAvailabilitySlotRequest struct {
	Type        *models.AvailabilityType `json:"type,omitempty"`
	DayOfWeek   *int                     `json:"day_of_week,omitempty"`
	Date        *time.Time               `json:"date,omitempty"`
	StartTime   *time.Time               `json:"start_time,omitempty"`
	EndTime     *time.Time               `json:"end_time,omitempty"`
	IsRecurring *bool                    `json:"is_recurring,omitempty"`
	RecurUntil  *time.Time               `json:"recur_until,omitempty"`
	Notes       *string                  `json:"notes,omitempty"`
}

// AvailabilitySlotFilter represents filters for availability queries
type AvailabilitySlotFilter struct {
	ArtisanID   uuid.UUID                `json:"artisan_id"`
	Type        *models.AvailabilityType `json:"type,omitempty"`
	DayOfWeek   *int                     `json:"day_of_week,omitempty"`
	StartDate   *time.Time               `json:"start_date,omitempty"`
	EndDate     *time.Time               `json:"end_date,omitempty"`
	IsRecurring *bool                    `json:"is_recurring,omitempty"`
	Page        int                      `json:"page"`
	PageSize    int                      `json:"page_size"`
}

// CheckAvailabilitySlotRequest represents a request to check availability
type CheckAvailabilitySlotRequest struct {
	ArtisanID uuid.UUID `json:"artisan_id" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

// BulkCreateAvailabilitySlotRequest represents a bulk availability creation request
type BulkCreateAvailabilitySlotRequest struct {
	ArtisanID   uuid.UUID               `json:"artisan_id" validate:"required"`
	Type        models.AvailabilityType `json:"type" validate:"required"`
	DaysOfWeek  []int                   `json:"days_of_week" validate:"required,min=1"`
	StartTime   time.Time               `json:"start_time" validate:"required"`
	EndTime     time.Time               `json:"end_time" validate:"required"`
	IsRecurring bool                    `json:"is_recurring"`
	RecurUntil  *time.Time              `json:"recur_until,omitempty"`
	Notes       string                  `json:"notes,omitempty"`
}

// ============================================================================
// Availability Response DTOs
// ============================================================================

// AvailabilityResponse represents an availability slot
type AvailabilitySlotResponse struct {
	ID          uuid.UUID               `json:"id"`
	TenantID    uuid.UUID               `json:"tenant_id"`
	ArtisanID   uuid.UUID               `json:"artisan_id"`
	Type        models.AvailabilityType `json:"type"`
	DayOfWeek   *int                    `json:"day_of_week,omitempty"`
	DayName     string                  `json:"day_name,omitempty"`
	Date        *time.Time              `json:"date,omitempty"`
	StartTime   time.Time               `json:"start_time"`
	EndTime     time.Time               `json:"end_time"`
	IsRecurring bool                    `json:"is_recurring"`
	RecurUntil  *time.Time              `json:"recur_until,omitempty"`
	Notes       string                  `json:"notes,omitempty"`
	Artisan     *UserSummary            `json:"artisan,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// AvailabilitySlotListResponse represents a paginated list of availability slots
type AvailabilitySlotListResponse struct {
	Availabilities []*AvailabilitySlotResponse `json:"availabilities"`
	Page           int                         `json:"page"`
	PageSize       int                         `json:"page_size"`
	TotalItems     int64                       `json:"total_items"`
	TotalPages     int                         `json:"total_pages"`
	HasNext        bool                        `json:"has_next"`
	HasPrevious    bool                        `json:"has_previous"`
}

// AvailabilitySlotCheckResponse represents the result of an availability check
type AvailabilitySlotCheckResponse struct {
	IsAvailable      bool                        `json:"is_available"`
	ConflictingSlots []*AvailabilitySlotResponse `json:"conflicting_slots,omitempty"`
	AvailableSlots   []*AvailabilitySlotResponse `json:"available_slots,omitempty"`
	SuggestedTimes   []TimeSlot                  `json:"suggested_times,omitempty"`
}

// TimeSlot represents a suggested time slot
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int       `json:"duration_minutes"`
}

// WeeklyScheduleResponse represents a weekly schedule
type WeeklyScheduleResponse struct {
	ArtisanID uuid.UUID                              `json:"artisan_id"`
	WeekStart time.Time                              `json:"week_start"`
	WeekEnd   time.Time                              `json:"week_end"`
	Schedule  map[string][]*AvailabilitySlotResponse `json:"schedule"` // day -> slots
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToAvailabilitySlotResponse converts an Availability model to AvailabilitySlotResponse DTO
func ToAvailabilitySlotResponse(availability *models.Availability) *AvailabilitySlotResponse {
	if availability == nil {
		return nil
	}

	resp := &AvailabilitySlotResponse{
		ID:          availability.ID,
		TenantID:    availability.TenantID,
		ArtisanID:   availability.ArtisanID,
		Type:        availability.Type,
		DayOfWeek:   availability.DayOfWeek,
		Date:        availability.Date,
		StartTime:   availability.StartTime,
		EndTime:     availability.EndTime,
		IsRecurring: availability.IsRecurring,
		RecurUntil:  availability.RecurUntil,
		Notes:       availability.Notes,
		CreatedAt:   availability.CreatedAt,
		UpdatedAt:   availability.UpdatedAt,
	}

	// Add day name if day of week is set
	if availability.DayOfWeek != nil {
		dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		if *availability.DayOfWeek >= 0 && *availability.DayOfWeek < 7 {
			resp.DayName = dayNames[*availability.DayOfWeek]
		}
	}

	// Add artisan if available
	if availability.Artisan != nil {
		resp.Artisan = &UserSummary{
			ID:        availability.Artisan.ID,
			FirstName: availability.Artisan.FirstName,
			LastName:  availability.Artisan.LastName,
			Email:     availability.Artisan.Email,
			AvatarURL: availability.Artisan.AvatarURL,
		}
	}

	return resp
}

// ToAvailabilitySlotResponses converts multiple Availability models to DTOs
func ToAvailabilitySlotResponses(availabilities []*models.Availability) []*AvailabilitySlotResponse {
	if availabilities == nil {
		return nil
	}

	responses := make([]*AvailabilitySlotResponse, len(availabilities))
	for i, availability := range availabilities {
		responses[i] = ToAvailabilitySlotResponse(availability)
	}
	return responses
}
