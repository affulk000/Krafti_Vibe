package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Task Request DTOs
// ============================================================================

// CreateTaskRequest represents a request to create a task
type CreateTaskRequest struct {
	ProjectID      uuid.UUID              `json:"project_id" validate:"required"`
	MilestoneID    *uuid.UUID             `json:"milestone_id,omitempty"`
	AssignedToID   *uuid.UUID             `json:"assigned_to_id,omitempty"`
	Title          string                 `json:"title" validate:"required,max=255"`
	Description    string                 `json:"description,omitempty"`
	Priority       models.TaskPriority    `json:"priority" validate:"required"`
	StartDate      *time.Time             `json:"start_date,omitempty"`
	DueDate        *time.Time             `json:"due_date,omitempty"`
	EstimatedHours float64                `json:"estimated_hours,omitempty" validate:"min=0"`
	EstimatedCost  float64                `json:"estimated_cost,omitempty" validate:"min=0"`
	DependsOn      []uuid.UUID            `json:"depends_on,omitempty"`
	Checklist      []models.ChecklistItem `json:"checklist,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	Title          *string                 `json:"title,omitempty" validate:"omitempty,max=255"`
	Description    *string                 `json:"description,omitempty"`
	Status         *models.TaskStatus      `json:"status,omitempty"`
	Priority       *models.TaskPriority    `json:"priority,omitempty"`
	AssignedToID   *uuid.UUID              `json:"assigned_to_id,omitempty"`
	MilestoneID    *uuid.UUID              `json:"milestone_id,omitempty"`
	StartDate      *time.Time              `json:"start_date,omitempty"`
	DueDate        *time.Time              `json:"due_date,omitempty"`
	EstimatedHours *float64                `json:"estimated_hours,omitempty" validate:"omitempty,min=0"`
	EstimatedCost  *float64                `json:"estimated_cost,omitempty" validate:"omitempty,min=0"`
	TrackedHours   *float64                `json:"tracked_hours,omitempty" validate:"omitempty,min=0"`
	ActualCost     *float64                `json:"actual_cost,omitempty" validate:"omitempty,min=0"`
	DependsOn      []uuid.UUID             `json:"depends_on,omitempty"`
	BlockedBy      []uuid.UUID             `json:"blocked_by,omitempty"`
	BlockReason    *string                 `json:"block_reason,omitempty"`
	Checklist      []models.ChecklistItem  `json:"checklist,omitempty"`
	AttachmentURLs []string                `json:"attachment_urls,omitempty"`
	Metadata       map[string]interface{}  `json:"metadata,omitempty"`
}

// ReorderTasksRequest represents a request to reorder tasks
type ReorderTasksRequest struct {
	ProjectID  uuid.UUID        `json:"project_id" validate:"required"`
	TaskOrders map[string]int   `json:"task_orders" validate:"required"` // task_id -> order_index
}

// TaskFilter represents filters for task queries
type TaskFilter struct {
	ProjectID    uuid.UUID            `json:"project_id"`
	MilestoneID  *uuid.UUID           `json:"milestone_id,omitempty"`
	AssignedToID *uuid.UUID           `json:"assigned_to_id,omitempty"`
	Status       *models.TaskStatus   `json:"status,omitempty"`
	Priority     *models.TaskPriority `json:"priority,omitempty"`
	IsOverdue    *bool                `json:"is_overdue,omitempty"`
	Page         int                  `json:"page"`
	PageSize     int                  `json:"page_size"`
}

// ============================================================================
// Task Response DTOs
// ============================================================================

// TaskResponse represents a task
type TaskResponse struct {
	ID             uuid.UUID              `json:"id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	ProjectID      uuid.UUID              `json:"project_id"`
	MilestoneID    *uuid.UUID             `json:"milestone_id,omitempty"`
	AssignedToID   *uuid.UUID             `json:"assigned_to_id,omitempty"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description,omitempty"`
	Status         models.TaskStatus      `json:"status"`
	Priority       models.TaskPriority    `json:"priority"`
	OrderIndex     int                    `json:"order_index"`
	StartDate      *time.Time             `json:"start_date,omitempty"`
	DueDate        *time.Time             `json:"due_date,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	EstimatedHours float64                `json:"estimated_hours"`
	TrackedHours   float64                `json:"tracked_hours"`
	EstimatedCost  float64                `json:"estimated_cost"`
	ActualCost     float64                `json:"actual_cost"`
	DependsOn      []uuid.UUID            `json:"depends_on,omitempty"`
	BlockedBy      []uuid.UUID            `json:"blocked_by,omitempty"`
	BlockReason    string                 `json:"block_reason,omitempty"`
	AttachmentURLs []string               `json:"attachment_urls,omitempty"`
	Checklist      []models.ChecklistItem `json:"checklist,omitempty"`
	Metadata       models.JSONB           `json:"metadata,omitempty"`
	IsCompleted    bool                   `json:"is_completed"`
	IsOverdue      bool                   `json:"is_overdue"`
	Progress       int                    `json:"progress"`
	AssignedTo     *UserSummary           `json:"assigned_to,omitempty"`
	Milestone      *MilestoneSummary      `json:"milestone,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// TaskListResponse represents a paginated list of tasks
type TaskListResponse struct {
	Tasks       []*TaskResponse `json:"tasks"`
	Page        int             `json:"page"`
	PageSize    int             `json:"page_size"`
	TotalItems  int64           `json:"total_items"`
	TotalPages  int             `json:"total_pages"`
	HasNext     bool            `json:"has_next"`
	HasPrevious bool            `json:"has_previous"`
}

// TaskStatsResponse represents task statistics
type TaskStatsResponse struct {
	ProjectID           uuid.UUID `json:"project_id,omitempty"`
	UserID              uuid.UUID `json:"user_id,omitempty"`
	TotalTasks          int64     `json:"total_tasks"`
	TodoTasks           int64     `json:"todo_tasks"`
	InProgressTasks     int64     `json:"in_progress_tasks"`
	BlockedTasks        int64     `json:"blocked_tasks"`
	ReviewTasks         int64     `json:"review_tasks"`
	CompletedTasks      int64     `json:"completed_tasks"`
	OverdueTasks        int64     `json:"overdue_tasks"`
	TotalEstimatedHours float64   `json:"total_estimated_hours"`
	TotalTrackedHours   float64   `json:"total_tracked_hours"`
	CompletionRate      float64   `json:"completion_rate"`
}

// UserSummary represents a summary of a user
type UserSummary struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}

// MilestoneSummary represents a summary of a milestone
type MilestoneSummary struct {
	ID      uuid.UUID               `json:"id"`
	Title   string                  `json:"title"`
	Status  models.MilestoneStatus  `json:"status"`
	DueDate *time.Time              `json:"due_date,omitempty"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToTaskResponse converts a Task model to TaskResponse DTO
func ToTaskResponse(task *models.ProjectTask) *TaskResponse {
	if task == nil {
		return nil
	}

	resp := &TaskResponse{
		ID:             task.ID,
		TenantID:       task.TenantID,
		ProjectID:      task.ProjectID,
		MilestoneID:    task.MilestoneID,
		AssignedToID:   task.AssignedToID,
		Title:          task.Title,
		Description:    task.Description,
		Status:         task.Status,
		Priority:       task.Priority,
		OrderIndex:     task.OrderIndex,
		StartDate:      task.StartDate,
		DueDate:        task.DueDate,
		CompletedAt:    task.CompletedAt,
		EstimatedHours: task.EstimatedHours,
		TrackedHours:   task.TrackedHours,
		EstimatedCost:  task.EstimatedCost,
		ActualCost:     task.ActualCost,
		DependsOn:      task.DependsOn,
		BlockedBy:      task.BlockedBy,
		BlockReason:    task.BlockReason,
		AttachmentURLs: task.AttachmentURLs,
		Checklist:      task.Checklist,
		Metadata:       task.Metadata,
		IsCompleted:    task.IsCompleted(),
		IsOverdue:      task.IsOverdue(),
		Progress:       task.ProgressPercentage(),
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}

	// Add assigned user if available
	if task.AssignedTo != nil {
		resp.AssignedTo = &UserSummary{
			ID:        task.AssignedTo.ID,
			FirstName: task.AssignedTo.FirstName,
			LastName:  task.AssignedTo.LastName,
			Email:     task.AssignedTo.Email,
			AvatarURL: task.AssignedTo.AvatarURL,
		}
	}

	// Add milestone if available
	if task.Milestone != nil {
		resp.Milestone = &MilestoneSummary{
			ID:      task.Milestone.ID,
			Title:   task.Milestone.Title,
			Status:  task.Milestone.Status,
			DueDate: task.Milestone.DueDate,
		}
	}

	return resp
}

// ToTaskResponses converts multiple Task models to DTOs
func ToTaskResponses(tasks []*models.ProjectTask) []*TaskResponse {
	if tasks == nil {
		return nil
	}

	responses := make([]*TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = ToTaskResponse(task)
	}
	return responses
}
