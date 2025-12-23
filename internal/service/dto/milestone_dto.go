package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Milestone Request DTOs
// ============================================================================

// CreateMilestoneRequest represents a request to create a milestone
type CreateMilestoneRequest struct {
	ProjectID          uuid.UUID              `json:"project_id" validate:"required"`
	Title              string                 `json:"title" validate:"required,max=255"`
	Description        string                 `json:"description,omitempty"`
	StartDate          *time.Time             `json:"start_date,omitempty"`
	DueDate            *time.Time             `json:"due_date,omitempty"`
	PaymentAmount      float64                `json:"payment_amount,omitempty" validate:"min=0"`
	PaymentPercentage  float64                `json:"payment_percentage,omitempty" validate:"min=0,max=100"`
	IsPaymentMilestone bool                   `json:"is_payment_milestone"`
	RequiresApproval   bool                   `json:"requires_approval"`
	Deliverables       []string               `json:"deliverables,omitempty"`
	ArtisanNotes       string                 `json:"artisan_notes,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateMilestoneRequest represents a request to update a milestone
type UpdateMilestoneRequest struct {
	Title              *string                 `json:"title,omitempty" validate:"omitempty,max=255"`
	Description        *string                 `json:"description,omitempty"`
	Status             *models.MilestoneStatus `json:"status,omitempty"`
	StartDate          *time.Time              `json:"start_date,omitempty"`
	DueDate            *time.Time              `json:"due_date,omitempty"`
	PaymentAmount      *float64                `json:"payment_amount,omitempty" validate:"omitempty,min=0"`
	PaymentPercentage  *float64                `json:"payment_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	IsPaymentMilestone *bool                   `json:"is_payment_milestone,omitempty"`
	RequiresApproval   *bool                   `json:"requires_approval,omitempty"`
	Deliverables       []string                `json:"deliverables,omitempty"`
	AttachmentURLs     []string                `json:"attachment_urls,omitempty"`
	CompletionProof    []string                `json:"completion_proof,omitempty"`
	ArtisanNotes       *string                 `json:"artisan_notes,omitempty"`
	CustomerNotes      *string                 `json:"customer_notes,omitempty"`
	Metadata           map[string]interface{}  `json:"metadata,omitempty"`
}

// ApproveMilestoneRequest represents a request to approve a milestone
type ApproveMilestoneRequest struct {
	CustomerNotes string `json:"customer_notes,omitempty"`
}

// RejectMilestoneRequest represents a request to reject a milestone
type RejectMilestoneRequest struct {
	RejectionReason string `json:"rejection_reason" validate:"required"`
	CustomerNotes   string `json:"customer_notes,omitempty"`
}

// ReorderMilestonesRequest represents a request to reorder milestones
type ReorderMilestonesRequest struct {
	ProjectID       uuid.UUID      `json:"project_id" validate:"required"`
	MilestoneOrders map[string]int `json:"milestone_orders" validate:"required"` // milestone_id -> order_index
}

// MilestoneFilter represents filters for milestone queries
type MilestoneFilter struct {
	ProjectID          uuid.UUID               `json:"project_id"`
	Status             *models.MilestoneStatus `json:"status,omitempty"`
	IsPaymentMilestone *bool                   `json:"is_payment_milestone,omitempty"`
	RequiresApproval   *bool                   `json:"requires_approval,omitempty"`
	IsOverdue          *bool                   `json:"is_overdue,omitempty"`
}

// ============================================================================
// Milestone Response DTOs
// ============================================================================

// MilestoneResponse represents a milestone
type MilestoneResponse struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	ProjectID          uuid.UUID              `json:"project_id"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description,omitempty"`
	Status             models.MilestoneStatus `json:"status"`
	OrderIndex         int                    `json:"order_index"`
	StartDate          *time.Time             `json:"start_date,omitempty"`
	DueDate            *time.Time             `json:"due_date,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	PaymentAmount      float64                `json:"payment_amount"`
	PaymentPercentage  float64                `json:"payment_percentage"`
	IsPaymentMilestone bool                   `json:"is_payment_milestone"`
	PaymentReceived    bool                   `json:"payment_received"`
	PaymentReceivedAt  *time.Time             `json:"payment_received_at,omitempty"`
	Deliverables       []string               `json:"deliverables,omitempty"`
	AttachmentURLs     []string               `json:"attachment_urls,omitempty"`
	CompletionProof    []string               `json:"completion_proof,omitempty"`
	RequiresApproval   bool                   `json:"requires_approval"`
	ApprovedByCustomer bool                   `json:"approved_by_customer"`
	ApprovedAt         *time.Time             `json:"approved_at,omitempty"`
	RejectionReason    string                 `json:"rejection_reason,omitempty"`
	ArtisanNotes       string                 `json:"artisan_notes,omitempty"`
	CustomerNotes      string                 `json:"customer_notes,omitempty"`
	Metadata           models.JSONB           `json:"metadata,omitempty"`
	IsCompleted        bool                   `json:"is_completed"`
	IsOverdue          bool                   `json:"is_overdue"`
	CanComplete        bool                   `json:"can_complete"`
	TaskCount          int                    `json:"task_count"`
	CompletedTaskCount int                    `json:"completed_task_count"`
	Tasks              []*TaskResponse        `json:"tasks,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// MilestoneListResponse represents a paginated list of milestones
type MilestoneListResponse struct {
	Milestones []*MilestoneResponse `json:"milestones"`
	TotalItems int                  `json:"total_items"`
}

// MilestoneStatsResponse represents milestone statistics
type MilestoneStatsResponse struct {
	ProjectID            uuid.UUID `json:"project_id"`
	TotalMilestones      int64     `json:"total_milestones"`
	PendingMilestones    int64     `json:"pending_milestones"`
	InProgressMilestones int64     `json:"in_progress_milestones"`
	CompletedMilestones  int64     `json:"completed_milestones"`
	OverdueMilestones    int64     `json:"overdue_milestones"`
	TotalPaymentAmount   float64   `json:"total_payment_amount"`
	ReceivedPayment      float64   `json:"received_payment"`
	CompletionRate       float64   `json:"completion_rate"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToMilestoneResponse converts a Milestone model to MilestoneResponse DTO
func ToMilestoneResponse(milestone *models.ProjectMilestone) *MilestoneResponse {
	if milestone == nil {
		return nil
	}

	resp := &MilestoneResponse{
		ID:                 milestone.ID,
		TenantID:           milestone.TenantID,
		ProjectID:          milestone.ProjectID,
		Title:              milestone.Title,
		Description:        milestone.Description,
		Status:             milestone.Status,
		OrderIndex:         milestone.OrderIndex,
		StartDate:          milestone.StartDate,
		DueDate:            milestone.DueDate,
		CompletedAt:        milestone.CompletedAt,
		PaymentAmount:      milestone.PaymentAmount,
		PaymentPercentage:  milestone.PaymentPercentage,
		IsPaymentMilestone: milestone.IsPaymentMilestone,
		PaymentReceived:    milestone.PaymentReceived,
		PaymentReceivedAt:  milestone.PaymentReceivedAt,
		Deliverables:       milestone.Deliverables,
		AttachmentURLs:     milestone.AttachmentURLs,
		CompletionProof:    milestone.CompletionProof,
		RequiresApproval:   milestone.RequiresApproval,
		ApprovedByCustomer: milestone.ApprovedByCustomer,
		ApprovedAt:         milestone.ApprovedAt,
		RejectionReason:    milestone.RejectionReason,
		ArtisanNotes:       milestone.ArtisanNotes,
		CustomerNotes:      milestone.CustomerNotes,
		Metadata:           milestone.Metadata,
		IsCompleted:        milestone.IsCompleted(),
		IsOverdue:          milestone.IsOverdue(),
		CanComplete:        milestone.CanComplete(),
		CreatedAt:          milestone.CreatedAt,
		UpdatedAt:          milestone.UpdatedAt,
	}

	// Add tasks if available
	if milestone.Tasks != nil {
		resp.Tasks = ToTaskResponses(convertTaskSlice(milestone.Tasks))
		resp.TaskCount = len(milestone.Tasks)

		// Count completed tasks
		completedCount := 0
		for _, task := range milestone.Tasks {
			if task.Status == models.TaskStatusDone {
				completedCount++
			}
		}
		resp.CompletedTaskCount = completedCount
	}

	return resp
}

// ToMilestoneResponses converts multiple Milestone models to DTOs
func ToMilestoneResponses(milestones []*models.ProjectMilestone) []*MilestoneResponse {
	if milestones == nil {
		return nil
	}

	responses := make([]*MilestoneResponse, len(milestones))
	for i, milestone := range milestones {
		responses[i] = ToMilestoneResponse(milestone)
	}
	return responses
}

// convertTaskSlice converts []ProjectTask to []*ProjectTask
func convertTaskSlice(tasks []models.ProjectTask) []*models.ProjectTask {
	result := make([]*models.ProjectTask, len(tasks))
	for i := range tasks {
		result[i] = &tasks[i]
	}
	return result
}
