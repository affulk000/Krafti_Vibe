package models

import (
	"time"

	"github.com/google/uuid"
)

type ReportType string

const (
	ReportTypeBookingSummary     ReportType = "booking_summary"
	ReportTypeRevenue            ReportType = "revenue"
	ReportTypeArtisanPerformance ReportType = "artisan_performance"
	ReportTypeCustomerAnalytics  ReportType = "customer_analytics"
	ReportTypePayments           ReportType = "payments"
)

type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
)

type Report struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index" validate:"required"`

	// Report Details
	Type        ReportType `json:"type" gorm:"type:varchar(50);not null;index" validate:"required"`
	Name        string     `json:"name" gorm:"not null;size:255" validate:"required"`
	Description string     `json:"description,omitempty" gorm:"type:text"`

	// Status
	Status ReportStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending'" validate:"required"`

	// Date Range
	StartDate time.Time `json:"start_date" gorm:"not null" validate:"required"`
	EndDate   time.Time `json:"end_date" gorm:"not null" validate:"required,gtfield=StartDate"`

	// Filters
	Filters JSONB `json:"filters,omitempty" gorm:"type:jsonb"`

	// Output
	FileURL     string     `json:"file_url,omitempty" gorm:"size:500"`
	FileFormat  string     `json:"file_format" gorm:"size:10;default:'pdf'"` // pdf, csv, xlsx
	GeneratedAt *time.Time `json:"generated_at,omitempty"`

	// Error Handling
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`

	// Requested By
	RequestedByID uuid.UUID `json:"requested_by_id" gorm:"type:uuid;not null" validate:"required"`

	// Scheduled
	IsScheduled  bool   `json:"is_scheduled" gorm:"default:false"`
	ScheduleCron string `json:"schedule_cron,omitempty" gorm:"size:100"` // cron expression

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant      *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	RequestedBy *User   `json:"requested_by,omitempty" gorm:"foreignKey:RequestedByID"`
}

// Business Methods
func (r *Report) IsCompleted() bool {
	return r.Status == ReportStatusCompleted
}

func (r *Report) IsFailed() bool {
	return r.Status == ReportStatusFailed
}
