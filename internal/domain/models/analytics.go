package models

import (
	"time"

	"github.com/google/uuid"
)

type AnalyticsEventType string

const (
	AnalyticsEventPageView         AnalyticsEventType = "page_view"
	AnalyticsEventClick            AnalyticsEventType = "click"
	AnalyticsEventBookingStarted   AnalyticsEventType = "booking_started"
	AnalyticsEventBookingAbandoned AnalyticsEventType = "booking_abandoned"
	AnalyticsEventSearchPerformed  AnalyticsEventType = "search_performed"
	AnalyticsEventProfileViewed    AnalyticsEventType = "profile_viewed"
)

type AnalyticsEvent struct {
	BaseModel

	// Multi-tenancy
	TenantID *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"`

	// User
	UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	SessionID   string     `json:"session_id" gorm:"size:255;index" validate:"required"`
	AnonymousID string     `json:"anonymous_id,omitempty" gorm:"size:255;index"`

	// Event Details
	EventType  AnalyticsEventType `json:"event_type" gorm:"type:varchar(100);not null;index" validate:"required"`
	EventName  string             `json:"event_name" gorm:"size:255;not null" validate:"required"`
	Properties JSONB              `json:"properties,omitempty" gorm:"type:jsonb"`

	// Context
	PageURL   string `json:"page_url,omitempty" gorm:"size:500"`
	PageTitle string `json:"page_title,omitempty" gorm:"size:255"`
	Referrer  string `json:"referrer,omitempty" gorm:"size:500"`

	// Device & Location
	UserAgent  string `json:"user_agent,omitempty" gorm:"size:500"`
	IPAddress  string `json:"ip_address,omitempty" gorm:"size:45"`
	Country    string `json:"country,omitempty" gorm:"size:2"`
	City       string `json:"city,omitempty" gorm:"size:100"`
	DeviceType string `json:"device_type,omitempty" gorm:"size:50"` // mobile, tablet, desktop
	Browser    string `json:"browser,omitempty" gorm:"size:50"`
	OS         string `json:"os,omitempty" gorm:"size:50"`

	// Timestamp
	EventTimestamp time.Time `json:"event_timestamp" gorm:"not null;index" validate:"required"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
