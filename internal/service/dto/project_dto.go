package dto

import (
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"

	"github.com/google/uuid"
)

// ============================================================================
// Project Request DTOs
// ============================================================================

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	TenantID     uuid.UUID              `json:"tenant_id" validate:"required"`
	ArtisanID    uuid.UUID              `json:"artisan_id" validate:"required"`
	CustomerID   *uuid.UUID             `json:"customer_id,omitempty"`
	Title        string                 `json:"title" validate:"required,max=255"`
	Description  string                 `json:"description,omitempty"`
	Priority     models.ProjectPriority `json:"priority" validate:"required"`
	StartDate    *time.Time             `json:"start_date,omitempty"`
	DueDate      *time.Time             `json:"due_date,omitempty"`
	BudgetAmount float64                `json:"budget_amount,omitempty" validate:"min=0"`
	Currency     string                 `json:"currency,omitempty" validate:"len=3"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the create project request
func (r *CreateProjectRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if r.ArtisanID == uuid.Nil {
		return fmt.Errorf("artisan_id is required")
	}
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if r.StartDate != nil && r.DueDate != nil && r.DueDate.Before(*r.StartDate) {
		return fmt.Errorf("due_date must be after start_date")
	}
	if r.BudgetAmount < 0 {
		return fmt.Errorf("budget_amount cannot be negative")
	}
	return nil
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Title        *string                 `json:"title,omitempty" validate:"omitempty,max=255"`
	Description  *string                 `json:"description,omitempty"`
	Priority     *models.ProjectPriority `json:"priority,omitempty"`
	StartDate    *time.Time              `json:"start_date,omitempty"`
	DueDate      *time.Time              `json:"due_date,omitempty"`
	BudgetAmount *float64                `json:"budget_amount,omitempty" validate:"omitempty,min=0"`
	Currency     *string                 `json:"currency,omitempty" validate:"omitempty,len=3"`
	Tags         []string                `json:"tags,omitempty"`
	Metadata     map[string]interface{}  `json:"metadata,omitempty"`
}

// ProjectFilter represents filters for project queries
type ProjectFilter struct {
	TenantID   uuid.UUID               `json:"tenant_id"`
	ArtisanID  *uuid.UUID              `json:"artisan_id,omitempty"`
	CustomerID *uuid.UUID              `json:"customer_id,omitempty"`
	Status     *models.ProjectStatus   `json:"status,omitempty"`
	Priority   *models.ProjectPriority `json:"priority,omitempty"`
	Tags       []string                `json:"tags,omitempty"`
	FromDate   *time.Time              `json:"from_date,omitempty"`
	ToDate     *time.Time              `json:"to_date,omitempty"`
	IsOverdue  *bool                   `json:"is_overdue,omitempty"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
}

// ProjectStatusUpdateRequest represents a request to update project status
type ProjectStatusUpdateRequest struct {
	Status models.ProjectStatus `json:"status" validate:"required"`
	Reason string               `json:"reason,omitempty"`
}

// BulkProjectUpdateRequest represents bulk project update
type BulkProjectUpdateRequest struct {
	ProjectIDs []uuid.UUID           `json:"project_ids" validate:"required,min=1"`
	Status     *models.ProjectStatus `json:"status,omitempty"`
	ArtisanID  *uuid.UUID            `json:"artisan_id,omitempty"`
}

// ============================================================================
// Project Response DTOs
// ============================================================================

// ProjectResponse represents a project
type ProjectResponse struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	ArtisanID          uuid.UUID              `json:"artisan_id"`
	CustomerID         *uuid.UUID             `json:"customer_id,omitempty"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description,omitempty"`
	Status             models.ProjectStatus   `json:"status"`
	Priority           models.ProjectPriority `json:"priority"`
	StartDate          *time.Time             `json:"start_date,omitempty"`
	DueDate            *time.Time             `json:"due_date,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	BudgetAmount       float64                `json:"budget_amount"`
	Currency           string                 `json:"currency"`
	ProgressPercent    int                    `json:"progress_percent"`
	TasksTotal         int                    `json:"tasks_total"`
	TasksCompleted     int                    `json:"tasks_completed"`
	TasksOverdue       int                    `json:"tasks_overdue"`
	ActiveBlockedTasks int                    `json:"active_blocked_tasks"`
	Tags               []string               `json:"tags,omitempty"`
	Metadata           models.JSONB           `json:"metadata,omitempty"`
	ArtisanName        string                 `json:"artisan_name,omitempty"`
	CustomerName       string                 `json:"customer_name,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// ProjectListResponse represents a paginated list of projects
type ProjectListResponse struct {
	Projects    []*ProjectResponse `json:"projects"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
	TotalItems  int64              `json:"total_items"`
	TotalPages  int                `json:"total_pages"`
	HasNext     bool               `json:"has_next"`
	HasPrevious bool               `json:"has_previous"`
}

// ProjectStatsResponse represents project statistics
type ProjectStatsResponse struct {
	TotalProjects          int64                            `json:"total_projects"`
	ActiveProjects         int64                            `json:"active_projects"`
	CompletedProjects      int64                            `json:"completed_projects"`
	OnHoldProjects         int64                            `json:"on_hold_projects"`
	CancelledProjects      int64                            `json:"cancelled_projects"`
	OverdueProjects        int64                            `json:"overdue_projects"`
	ByStatus               map[models.ProjectStatus]int64   `json:"by_status"`
	ByPriority             map[models.ProjectPriority]int64 `json:"by_priority"`
	AverageProgress        float64                          `json:"average_progress"`
	TotalBudget            float64                          `json:"total_budget"`
	OnTimeProjects         int64                            `json:"on_time_projects"`
	CompletionRate         float64                          `json:"completion_rate"`
	AverageTasksPerProject float64                          `json:"average_tasks_per_project"`
}

// ArtisanProjectStatsResponse represents artisan project stats
type ArtisanProjectStatsResponse struct {
	ArtisanID              uuid.UUID `json:"artisan_id"`
	TotalProjects          int64     `json:"total_projects"`
	ActiveProjects         int64     `json:"active_projects"`
	CompletedProjects      int64     `json:"completed_projects"`
	AverageProgress        float64   `json:"average_progress"`
	TotalRevenue           float64   `json:"total_revenue"`
	OnTimeDeliveryRate     float64   `json:"on_time_delivery_rate"`
	CustomerSatisfaction   float64   `json:"customer_satisfaction"`
	AverageProjectDuration float64   `json:"average_project_duration_days"`
}

// CustomerProjectStatsResponse represents customer project stats
type CustomerProjectStatsResponse struct {
	CustomerID         uuid.UUID `json:"customer_id"`
	TotalProjects      int64     `json:"total_projects"`
	ActiveProjects     int64     `json:"active_projects"`
	CompletedProjects  int64     `json:"completed_projects"`
	TotalSpent         float64   `json:"total_spent"`
	AverageProjectCost float64   `json:"average_project_cost"`
}

// ProjectHealthResponse represents project health metrics
type ProjectHealthResponse struct {
	ProjectID          uuid.UUID `json:"project_id"`
	HealthScore        int       `json:"health_score"`
	IsOnTrack          bool      `json:"is_on_track"`
	IsOverBudget       bool      `json:"is_over_budget"`
	IsOverdue          bool      `json:"is_overdue"`
	BlockedTasksCount  int       `json:"blocked_tasks_count"`
	OverdueTasksCount  int       `json:"overdue_tasks_count"`
	CompletionVelocity float64   `json:"completion_velocity"`
	RiskLevel          string    `json:"risk_level"`
	Recommendations    []string  `json:"recommendations"`
}

// ProjectTimelineResponse represents timeline events
type ProjectTimelineResponse struct {
	Events []*TimelineEventResponse `json:"events"`
}

// TimelineEventResponse represents a timeline event
type TimelineEventResponse struct {
	Date        time.Time  `json:"date"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	UserName    string     `json:"user_name,omitempty"`
}

// ArtisanDashboardResponse represents artisan dashboard
type ArtisanDashboardResponse struct {
	ActiveProjects    []*ProjectResponse          `json:"active_projects"`
	UpcomingDeadlines []*ProjectResponse          `json:"upcoming_deadlines"`
	OverdueProjects   []*ProjectResponse          `json:"overdue_projects"`
	RecentlyCompleted []*ProjectResponse          `json:"recently_completed"`
	Statistics        ArtisanProjectStatsResponse `json:"statistics"`
	TasksSummary      TasksSummaryResponse        `json:"tasks_summary"`
}

// TenantProjectDashboardResponse represents tenant dashboard
type TenantProjectDashboardResponse struct {
	Statistics           ProjectStatsResponse          `json:"statistics"`
	ActiveProjects       []*ProjectResponse            `json:"active_projects"`
	HighPriorityProjects []*ProjectResponse            `json:"high_priority_projects"`
	OverdueProjects      []*ProjectResponse            `json:"overdue_projects"`
	RecentActivity       []*TimelineEventResponse      `json:"recent_activity"`
	TopArtisans          []*ArtisanPerformanceResponse `json:"top_artisans"`
}

// TasksSummaryResponse represents tasks summary
type TasksSummaryResponse struct {
	TotalTasks      int64 `json:"total_tasks"`
	CompletedTasks  int64 `json:"completed_tasks"`
	OverdueTasks    int64 `json:"overdue_tasks"`
	BlockedTasks    int64 `json:"blocked_tasks"`
	InProgressTasks int64 `json:"in_progress_tasks"`
}

// ArtisanPerformanceResponse represents artisan performance
type ArtisanPerformanceResponse struct {
	ArtisanID         uuid.UUID `json:"artisan_id"`
	ArtisanName       string    `json:"artisan_name"`
	ActiveProjects    int       `json:"active_projects"`
	CompletedProjects int       `json:"completed_projects"`
	AverageProgress   float64   `json:"average_progress"`
	OnTimeRate        float64   `json:"on_time_rate"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToProjectResponse converts a Project model to ProjectResponse DTO
func ToProjectResponse(project *models.Project) *ProjectResponse {
	if project == nil {
		return nil
	}

	resp := &ProjectResponse{
		ID:                 project.ID,
		TenantID:           project.TenantID,
		ArtisanID:          project.ArtisanID,
		CustomerID:         project.CustomerID,
		Title:              project.Title,
		Description:        project.Description,
		Status:             project.Status,
		Priority:           project.Priority,
		StartDate:          project.StartDate,
		DueDate:            project.DueDate,
		CompletedAt:        project.CompletedAt,
		BudgetAmount:       project.BudgetAmount,
		Currency:           project.Currency,
		ProgressPercent:    project.ProgressPercent,
		TasksTotal:         project.TasksTotal,
		TasksCompleted:     project.TasksCompleted,
		TasksOverdue:       project.TasksOverdue,
		ActiveBlockedTasks: project.ActiveBlockedTasks,
		Tags:               project.Tags,
		Metadata:           project.Metadata,
		CreatedAt:          project.CreatedAt,
		UpdatedAt:          project.UpdatedAt,
	}

	// Add artisan name if available
	if project.Artisan != nil && project.Artisan.User != nil {
		resp.ArtisanName = project.Artisan.User.FirstName + " " + project.Artisan.User.LastName
	}

	// Add customer name if available
	if project.Customer != nil && project.Customer.User != nil {
		resp.CustomerName = project.Customer.User.FirstName + " " + project.Customer.User.LastName
	}

	return resp
}

// ToProjectResponses converts multiple Project models to DTOs
func ToProjectResponses(projects []*models.Project) []*ProjectResponse {
	if projects == nil {
		return nil
	}

	responses := make([]*ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = ToProjectResponse(project)
	}
	return responses
}

// ToProjectStatsResponse converts ProjectStats to response DTO
func ToProjectStatsResponse(stats repository.ProjectStats) *ProjectStatsResponse {
	return &ProjectStatsResponse{
		TotalProjects:          stats.TotalProjects,
		ActiveProjects:         stats.ActiveProjects,
		CompletedProjects:      stats.CompletedProjects,
		OnHoldProjects:         stats.OnHoldProjects,
		CancelledProjects:      stats.CancelledProjects,
		OverdueProjects:        stats.OverdueProjects,
		ByStatus:               stats.ByStatus,
		ByPriority:             stats.ByPriority,
		AverageProgress:        stats.AverageProgress,
		TotalBudget:            stats.TotalBudget,
		OnTimeProjects:         stats.OnTimeProjects,
		CompletionRate:         stats.CompletionRate,
		AverageTasksPerProject: stats.AverageTasksPerProject,
	}
}

// ToArtisanProjectStatsResponse converts stats to response DTO
func ToArtisanProjectStatsResponse(artisanID uuid.UUID, stats repository.ArtisanProjectStats) *ArtisanProjectStatsResponse {
	return &ArtisanProjectStatsResponse{
		ArtisanID:              artisanID,
		TotalProjects:          stats.TotalProjects,
		ActiveProjects:         stats.ActiveProjects,
		CompletedProjects:      stats.CompletedProjects,
		AverageProgress:        stats.AverageProgress,
		TotalRevenue:           stats.TotalRevenue,
		OnTimeDeliveryRate:     stats.OnTimeDeliveryRate,
		CustomerSatisfaction:   stats.CustomerSatisfaction,
		AverageProjectDuration: stats.AverageProjectDuration,
	}
}

// ToCustomerProjectStatsResponse converts stats to response DTO
func ToCustomerProjectStatsResponse(customerID uuid.UUID, stats repository.CustomerProjectStats) *CustomerProjectStatsResponse {
	return &CustomerProjectStatsResponse{
		CustomerID:         customerID,
		TotalProjects:      stats.TotalProjects,
		ActiveProjects:     stats.ActiveProjects,
		CompletedProjects:  stats.CompletedProjects,
		TotalSpent:         stats.TotalSpent,
		AverageProjectCost: stats.AverageProjectCost,
	}
}

// ToProjectHealthResponse converts health to response DTO
func ToProjectHealthResponse(health repository.ProjectHealth) *ProjectHealthResponse {
	return &ProjectHealthResponse{
		ProjectID:          health.ProjectID,
		HealthScore:        health.HealthScore,
		IsOnTrack:          health.IsOnTrack,
		IsOverBudget:       health.IsOverBudget,
		IsOverdue:          health.IsOverdue,
		BlockedTasksCount:  health.BlockedTasksCount,
		OverdueTasksCount:  health.OverdueTasksCount,
		CompletionVelocity: health.CompletionVelocity,
		RiskLevel:          health.RiskLevel,
		Recommendations:    health.Recommendations,
	}
}

// ToTimelineEventResponses converts timeline events to response DTOs
func ToTimelineEventResponses(events []repository.TimelineEvent) []*TimelineEventResponse {
	responses := make([]*TimelineEventResponse, len(events))
	for i, event := range events {
		responses[i] = &TimelineEventResponse{
			Date:        event.Date,
			Type:        event.Type,
			Title:       event.Title,
			Description: event.Description,
			UserID:      event.UserID,
			UserName:    event.UserName,
		}
	}
	return responses
}

// ToArtisanDashboardResponse converts dashboard to response DTO
func ToArtisanDashboardResponse(dashboard repository.ArtisanDashboard, artisanID uuid.UUID) *ArtisanDashboardResponse {
	return &ArtisanDashboardResponse{
		ActiveProjects:    ToProjectResponses(dashboard.ActiveProjects),
		UpcomingDeadlines: ToProjectResponses(dashboard.UpcomingDeadlines),
		OverdueProjects:   ToProjectResponses(dashboard.OverdueProjects),
		RecentlyCompleted: ToProjectResponses(dashboard.RecentlyCompleted),
		Statistics:        *ToArtisanProjectStatsResponse(artisanID, dashboard.Statistics),
		TasksSummary: TasksSummaryResponse{
			TotalTasks:      dashboard.TasksSummary.TotalTasks,
			CompletedTasks:  dashboard.TasksSummary.CompletedTasks,
			OverdueTasks:    dashboard.TasksSummary.OverdueTasks,
			BlockedTasks:    dashboard.TasksSummary.BlockedTasks,
			InProgressTasks: dashboard.TasksSummary.InProgressTasks,
		},
	}
}

// ToTenantProjectDashboardResponse converts dashboard to response DTO
func ToTenantProjectDashboardResponse(dashboard repository.TenantProjectDashboard) *TenantProjectDashboardResponse {
	topArtisans := make([]*ArtisanPerformanceResponse, len(dashboard.TopArtisans))
	for i, perf := range dashboard.TopArtisans {
		topArtisans[i] = &ArtisanPerformanceResponse{
			ArtisanID:         perf.ArtisanID,
			ArtisanName:       perf.ArtisanName,
			ActiveProjects:    perf.ActiveProjects,
			CompletedProjects: perf.CompletedProjects,
			AverageProgress:   perf.AverageProgress,
			OnTimeRate:        perf.OnTimeRate,
		}
	}

	return &TenantProjectDashboardResponse{
		Statistics:           *ToProjectStatsResponse(dashboard.Statistics),
		ActiveProjects:       ToProjectResponses(dashboard.ActiveProjects),
		HighPriorityProjects: ToProjectResponses(dashboard.HighPriorityProjects),
		OverdueProjects:      ToProjectResponses(dashboard.OverdueProjects),
		RecentActivity:       ToTimelineEventResponses(dashboard.RecentActivity),
		TopArtisans:          topArtisans,
	}
}
