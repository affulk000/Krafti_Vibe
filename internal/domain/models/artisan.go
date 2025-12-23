package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// StringArray is a custom type for handling []string in JSONB
type StringArray []string

// CertificationArray is a custom type for handling []Certification in JSONB
type CertificationArray []Certification

// PortfolioArray is a custom type for handling []PortfolioItem in JSONB
type PortfolioArray []PortfolioItem

type Artisan struct {
	BaseModel

	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;index:idx_artisan_tenant_status;not null"`

	// Professional Info
	Bio             string      `json:"bio,omitempty" gorm:"type:text"`
	Specialization  StringArray `json:"specialization" gorm:"type:jsonb"`
	YearsExperience int         `json:"years_experience" gorm:"default:0"`

	// Certifications & Portfolio
	Certifications CertificationArray `json:"certifications,omitempty" gorm:"type:jsonb"`
	Portfolio      PortfolioArray     `json:"portfolio,omitempty" gorm:"type:jsonb"`

	// Ratings & Reviews
	Rating        float64 `json:"rating" gorm:"type:decimal(3,2);default:0"`
	ReviewCount   int     `json:"review_count" gorm:"default:0"`
	TotalBookings int     `json:"total_bookings" gorm:"default:0"`

	// Availability
	IsAvailable      bool   `json:"is_available" gorm:"default:true;index:idx_artisan_tenant_status"`
	AvailabilityNote string `json:"availability_note,omitempty" gorm:"size:500"`

	// Commission & Payment
	CommissionRate   float64 `json:"commission_rate" gorm:"type:decimal(5,2);default:0"` // Percentage
	PaymentAccountID string  `json:"payment_account_id,omitempty" gorm:"size:255"`

	// Settings
	AutoAcceptBookings   bool `json:"auto_accept_bookings" gorm:"default:false"`
	BookingLeadTime      int  `json:"booking_lead_time" gorm:"default:60"`   // minutes
	MaxAdvanceBooking    int  `json:"max_advance_booking" gorm:"default:90"` // days
	SimultaneousBookings int  `json:"simultaneous_bookings" gorm:"default:1"`

	// Location
	Location      Location `json:"location,omitempty" gorm:"type:jsonb"`
	ServiceRadius int      `json:"service_radius" gorm:"default:0"` // km, 0 = no travel

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	User         *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Availability []Availability `json:"availability,omitempty" gorm:"foreignKey:ArtisanID"`
	Services     []Service      `json:"services,omitempty" gorm:"foreignKey:ArtisanID"`
	Projects     []Project      `json:"projects,omitempty" gorm:"foreignKey:ArtisanID"`
	Bookings     []Booking      `json:"bookings,omitempty" gorm:"foreignKey:ArtisanID"`
	Reviews      []Review       `json:"reviews,omitempty" gorm:"foreignKey:ArtisanID"`

	// Dashboard (computed/embedded, not persisted)
	DashboardProjectsActive int        `json:"dashboard_projects_active,omitempty" gorm:"-"`
	DashboardTasksOpen      int        `json:"dashboard_tasks_open,omitempty" gorm:"-"`
	DashboardTasksOverdue   int        `json:"dashboard_tasks_overdue,omitempty" gorm:"-"`
	DashboardNextDueAt      *time.Time `json:"dashboard_next_due_at,omitempty" gorm:"-"`
}

type Certification struct {
	Name       string     `json:"name" validate:"required"`
	IssuedBy   string     `json:"issued_by" validate:"required"`
	IssuedDate time.Time  `json:"issued_date"`
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	FileURL    string     `json:"file_url,omitempty"`
}

type PortfolioItem struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	ImageURLs   []string  `json:"image_urls"`
	Date        time.Time `json:"date"`
	Tags        []string  `json:"tags,omitempty"`
}

type Location struct {
	Address    string  `json:"address"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude" validate:"latitude"`
	Longitude  float64 `json:"longitude" validate:"longitude"`
}

// Scan and Value methods for JSONB types
func (c *Certification) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &c)
}

func (c Certification) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (l *Location) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &l)
}

func (l Location) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, s)
}

func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

func (c *CertificationArray) Scan(value interface{}) error {
	if value == nil {
		*c = []Certification{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, c)
}

func (c CertificationArray) Value() (driver.Value, error) {
	if len(c) == 0 {
		return json.Marshal([]Certification{})
	}
	return json.Marshal(c)
}

func (p *PortfolioArray) Scan(value interface{}) error {
	if value == nil {
		*p = []PortfolioItem{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, p)
}

func (p PortfolioArray) Value() (driver.Value, error) {
	if len(p) == 0 {
		return json.Marshal([]PortfolioItem{})
	}
	return json.Marshal(p)
}
