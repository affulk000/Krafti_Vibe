package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityMedium   TaskPriority = "medium"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

type ProjectTask struct {
	BaseModel

	// Multi-tenancy
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	ProjectID   uuid.UUID  `json:"project_id" gorm:"type:uuid;not null;index"`
	MilestoneID *uuid.UUID `json:"milestone_id,omitempty" gorm:"type:uuid;index"`

	// Assignment
	AssignedToID *uuid.UUID `json:"assigned_to_id,omitempty" gorm:"type:uuid;index:idx_task_assigned_status"`

	// Core details
	Title       string       `json:"title" gorm:"not null;size:255"`
	Description string       `json:"description,omitempty" gorm:"type:text"`
	Status      TaskStatus   `json:"status" gorm:"type:varchar(32);not null;default:'todo';index:idx_task_assigned_status"`
	Priority    TaskPriority `json:"priority" gorm:"type:varchar(16);not null;default:'medium';index"`
	OrderIndex  int          `json:"order_index" gorm:"default:0"`

	// Timing
	StartDate   *time.Time `json:"start_date,omitempty" gorm:"type:datetime"`
	DueDate     *time.Time `json:"due_date,omitempty" gorm:"type:datetime"`
	CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"type:datetime"`

	// Estimates and tracking
	EstimatedHours float64 `json:"estimated_hours" gorm:"type:decimal(8,2);default:0"`
	TrackedHours   float64 `json:"tracked_hours" gorm:"type:decimal(8,2);default:0"`

	// Cost tracking (materials, labor)
	EstimatedCost float64 `json:"estimated_cost" gorm:"type:decimal(10,2);default:0"`
	ActualCost    float64 `json:"actual_cost" gorm:"type:decimal(10,2);default:0"`

	// Dependencies
	DependsOn   []uuid.UUID `json:"depends_on,omitempty" gorm:"type:uuid[]"`
	BlockedBy   []uuid.UUID `json:"blocked_by,omitempty" gorm:"type:uuid[]"`
	BlockReason string      `json:"block_reason,omitempty" gorm:"type:text"`

	// Attachments
	AttachmentURLs []string `json:"attachment_urls,omitempty" gorm:"type:text[]"`

	// Checklist items
	Checklist []ChecklistItem `json:"checklist,omitempty" gorm:"type:jsonb"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Project    *Project          `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Milestone  *ProjectMilestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	AssignedTo *User             `json:"assigned_to,omitempty" gorm:"foreignKey:AssignedToID"`
}

type ChecklistItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	IsCompleted bool   `json:"is_completed"`
	Order       int    `json:"order"`
}

// Business Methods
func (t *ProjectTask) IsCompleted() bool {
	return t.Status == TaskStatusDone
}

func (t *ProjectTask) IsOverdue() bool {
	if t.DueDate == nil || t.IsCompleted() {
		return false
	}
	return time.Now().After(*t.DueDate)
}

func (t *ProjectTask) ProgressPercentage() int {
	if len(t.Checklist) == 0 {
		return 0
	}
	completed := 0
	for _, item := range t.Checklist {
		if item.IsCompleted {
			completed++
		}
	}
	return (completed * 100) / len(t.Checklist)
}
