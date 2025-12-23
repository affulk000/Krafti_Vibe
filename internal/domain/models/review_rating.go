package models

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index:idx_review_tenant_artisan"`

	// References
	BookingID  uuid.UUID `json:"booking_id" gorm:"type:uuid;not null;uniqueIndex" validate:"required"`
	ArtisanID  uuid.UUID `json:"artisan_id" gorm:"type:uuid;not null;index:idx_review_tenant_artisan;index:idx_review_artisan_rating" validate:"required"`
	CustomerID uuid.UUID `json:"customer_id" gorm:"type:uuid;not null;index" validate:"required"`
	ServiceID  uuid.UUID `json:"service_id" gorm:"type:uuid;not null;index" validate:"required"`

	// Rating
	Rating int `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5;index:idx_review_artisan_rating" validate:"required,min=1,max=5"`

	// Review Content
	Title   string `json:"title,omitempty" gorm:"size:255"`
	Comment string `json:"comment,omitempty" gorm:"type:text"`

	// Additional Ratings (optional breakdowns)
	QualityRating         *int `json:"quality_rating,omitempty" gorm:"check:quality_rating >= 1 AND quality_rating <= 5"`
	ProfessionalismRating *int `json:"professionalism_rating,omitempty" gorm:"check:professionalism_rating >= 1 AND professionalism_rating <= 5"`
	ValueRating           *int `json:"value_rating,omitempty" gorm:"check:value_rating >= 1 AND value_rating <= 5"`
	TimelinessRating      *int `json:"timeliness_rating,omitempty" gorm:"check:timeliness_rating >= 1 AND timeliness_rating <= 5"`

	// Media
	PhotoURLs []string `json:"photo_urls,omitempty" gorm:"type:text[]"`

	// Response
	ResponseText string     `json:"response_text,omitempty" gorm:"type:text"`
	ResponsedAt  *time.Time `json:"responsed_at,omitempty"`
	ResponsedBy  *uuid.UUID `json:"responsed_by,omitempty"`

	// Moderation
	IsPublished   bool       `json:"is_published" gorm:"default:true"`
	IsFlagged     bool       `json:"is_flagged" gorm:"default:false"`
	FlaggedReason string     `json:"flagged_reason,omitempty" gorm:"type:text"`
	ModeratedAt   *time.Time `json:"moderated_at,omitempty"`
	ModeratedBy   *uuid.UUID `json:"moderated_by,omitempty"`

	// Helpful Votes
	HelpfulCount    int `json:"helpful_count" gorm:"default:0"`
	NotHelpfulCount int `json:"not_helpful_count" gorm:"default:0"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Booking  *Booking `json:"booking,omitempty" gorm:"foreignKey:BookingID"`
	Artisan  *User    `json:"artisan,omitempty" gorm:"foreignKey:ArtisanID"`
	Customer *User    `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Service  *Service `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
}

// Business Methods
func (r *Review) IsPositive() bool {
	return r.Rating >= 4
}

func (r *Review) HasResponse() bool {
	return r.ResponseText != ""
}
