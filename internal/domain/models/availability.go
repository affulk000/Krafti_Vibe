package models

import (
	"time"

	"github.com/google/uuid"
)

type AvailabilityType string

const (
	AvailabilityTypeRegular   AvailabilityType = "regular"
	AvailabilityTypeException AvailabilityType = "exception"
	AvailabilityTypeBreak     AvailabilityType = "break"
	AvailabilityTypeTimeOff   AvailabilityType = "time_off"
)

type Availability struct {
	BaseModel

	// Multi-tenancy
	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	ArtisanID uuid.UUID `json:"artisan_id" gorm:"type:uuid;not null;index"`

	// Type
	Type AvailabilityType `json:"type" gorm:"type:varchar(50);not null" validate:"required"`

	// Timing
	DayOfWeek *int       `json:"day_of_week,omitempty" gorm:"check:day_of_week >= 0 AND day_of_week <= 6"` // 0=Sunday
	Date      *time.Time `json:"date,omitempty"`                                                           // For specific date overrides
	StartTime time.Time  `json:"start_time" gorm:"not null" validate:"required"`
	EndTime   time.Time  `json:"end_time" gorm:"not null" validate:"required,gtfield=StartTime"`

	// Recurrence
	IsRecurring bool       `json:"is_recurring" gorm:"default:false"`
	RecurUntil  *time.Time `json:"recur_until,omitempty"`

	// Notes
	Notes string `json:"notes,omitempty" gorm:"type:text"`

	// Relationships
	Artisan *User `json:"artisan,omitempty" gorm:"foreignKey:ArtisanID"`
}

// Business Methods
func (a *Availability) IsAvailableAt(t time.Time) bool {
	if a.Type != AvailabilityTypeRegular {
		return false
	}

	if a.Date != nil && !a.Date.Equal(t) {
		return false
	}

	if a.DayOfWeek != nil && *a.DayOfWeek != int(t.Weekday()) {
		return false
	}

	return t.After(a.StartTime) && t.Before(a.EndTime)
}
