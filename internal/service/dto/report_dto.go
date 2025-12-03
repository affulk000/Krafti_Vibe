package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Report Request DTOs
// ============================================================================

// CreateReportRequest represents a request to create a report
type CreateReportRequest struct {
	Type        models.ReportType `json:"type" validate:"required"`
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description,omitempty"`
	StartDate   time.Time         `json:"start_date" validate:"required"`
	EndDate     time.Time         `json:"end_date" validate:"required,gtfield=StartDate"`
	FileFormat  string            `json:"file_format,omitempty" validate:"omitempty,oneof=pdf csv xlsx"`
	Filters     map[string]any    `json:"filters,omitempty"`
	IsScheduled bool              `json:"is_scheduled"`
	ScheduleCron string           `json:"schedule_cron,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// UpdateReportRequest represents a request to update a report
type UpdateReportRequest struct {
	Name         *string        `json:"name,omitempty"`
	Description  *string        `json:"description,omitempty"`
	ScheduleCron *string        `json:"schedule_cron,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// ReportFilter represents filters for report queries
type ReportFilter struct {
	TenantID      uuid.UUID             `json:"tenant_id" validate:"required"`
	Type          *models.ReportType    `json:"type,omitempty"`
	Status        *models.ReportStatus  `json:"status,omitempty"`
	RequestedByID *uuid.UUID            `json:"requested_by_id,omitempty"`
	IsScheduled   *bool                 `json:"is_scheduled,omitempty"`
	StartDate     *time.Time            `json:"start_date,omitempty"`
	EndDate       *time.Time            `json:"end_date,omitempty"`
	Page          int                   `json:"page"`
	PageSize      int                   `json:"page_size"`
}

// ============================================================================
// Report Response DTOs
// ============================================================================

// ReportResponse represents a report
type ReportResponse struct {
	ID            uuid.UUID            `json:"id"`
	TenantID      uuid.UUID            `json:"tenant_id"`
	Type          models.ReportType    `json:"type"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Status        models.ReportStatus  `json:"status"`
	StartDate     time.Time            `json:"start_date"`
	EndDate       time.Time            `json:"end_date"`
	Filters       models.JSONB         `json:"filters,omitempty"`
	FileURL       string               `json:"file_url,omitempty"`
	FileFormat    string               `json:"file_format"`
	GeneratedAt   *time.Time           `json:"generated_at,omitempty"`
	ErrorMessage  string               `json:"error_message,omitempty"`
	RequestedByID uuid.UUID            `json:"requested_by_id"`
	RequestedBy   *UserSummary         `json:"requested_by,omitempty"`
	IsScheduled   bool                 `json:"is_scheduled"`
	ScheduleCron  string               `json:"schedule_cron,omitempty"`
	IsCompleted   bool                 `json:"is_completed"`
	IsFailed      bool                 `json:"is_failed"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// ReportListResponse represents a paginated list of reports
type ReportListResponse struct {
	Reports     []*ReportResponse `json:"reports"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
	TotalItems  int64             `json:"total_items"`
	TotalPages  int               `json:"total_pages"`
	HasNext     bool              `json:"has_next"`
	HasPrevious bool              `json:"has_previous"`
}

// ReportStatsResponse represents report statistics
type ReportStatsResponse struct {
	TenantID              uuid.UUID                           `json:"tenant_id"`
	TotalReports          int64                               `json:"total_reports"`
	CompletedReports      int64                               `json:"completed_reports"`
	PendingReports        int64                               `json:"pending_reports"`
	FailedReports         int64                               `json:"failed_reports"`
	ScheduledReports      int64                               `json:"scheduled_reports"`
	ByType                map[models.ReportType]int64         `json:"by_type"`
	ByStatus              map[models.ReportStatus]int64       `json:"by_status"`
	ByFormat              map[string]int64                    `json:"by_format"`
	ReportsThisWeek       int64                               `json:"reports_this_week"`
	ReportsThisMonth      int64                               `json:"reports_this_month"`
	AvgGenerationTime     float64                             `json:"avg_generation_time_seconds"`
	SuccessRate           float64                             `json:"success_rate"`
	MostRequestedType     models.ReportType                   `json:"most_requested_type"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToReportResponse converts a Report model to ReportResponse DTO
func ToReportResponse(report *models.Report) *ReportResponse {
	if report == nil {
		return nil
	}

	resp := &ReportResponse{
		ID:            report.ID,
		TenantID:      report.TenantID,
		Type:          report.Type,
		Name:          report.Name,
		Description:   report.Description,
		Status:        report.Status,
		StartDate:     report.StartDate,
		EndDate:       report.EndDate,
		Filters:       report.Filters,
		FileURL:       report.FileURL,
		FileFormat:    report.FileFormat,
		GeneratedAt:   report.GeneratedAt,
		ErrorMessage:  report.ErrorMessage,
		RequestedByID: report.RequestedByID,
		IsScheduled:   report.IsScheduled,
		ScheduleCron:  report.ScheduleCron,
		IsCompleted:   report.IsCompleted(),
		IsFailed:      report.IsFailed(),
		CreatedAt:     report.CreatedAt,
		UpdatedAt:     report.UpdatedAt,
	}

	// Add requester if available
	if report.RequestedBy != nil {
		resp.RequestedBy = &UserSummary{
			ID:        report.RequestedBy.ID,
			FirstName: report.RequestedBy.FirstName,
			LastName:  report.RequestedBy.LastName,
			Email:     report.RequestedBy.Email,
			AvatarURL: report.RequestedBy.AvatarURL,
		}
	}

	return resp
}

// ToReportResponses converts multiple Report models to DTOs
func ToReportResponses(reports []*models.Report) []*ReportResponse {
	if reports == nil {
		return nil
	}

	responses := make([]*ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = ToReportResponse(report)
	}
	return responses
}
