package models

import (
	"time"

	"github.com/google/uuid"
)

type ProjectStatus string

const (
	ProjectStatusPlanned    ProjectStatus = "planned"
	ProjectStatusInProgress ProjectStatus = "in_progress"
	ProjectStatusOnHold     ProjectStatus = "on_hold"
	ProjectStatusCompleted  ProjectStatus = "completed"
	ProjectStatusCancelled  ProjectStatus = "cancelled"
)

type ProjectPriority string

const (
	ProjectPriorityLow    ProjectPriority = "low"
	ProjectPriorityMedium ProjectPriority = "medium"
	ProjectPriorityHigh   ProjectPriority = "high"
)

// Project represents an artisan-led project for a customer.
// Designed for dashboard embedding and progress tracking.
type Project struct {
	BaseModel

	// Multi-tenancy
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_project_tenant_artisan"`
	ArtisanID  uuid.UUID  `json:"artisan_id" gorm:"type:uuid;not null;index:idx_project_tenant_artisan"`
	CustomerID *uuid.UUID `json:"customer_id,omitempty" gorm:"type:uuid;index"`

	// Core details
	Title       string          `json:"title" gorm:"not null;size:255"`
	Description string          `json:"description,omitempty" gorm:"type:text"`
	Status      ProjectStatus   `json:"status" gorm:"type:varchar(32);not null;default:'planned';index"`
	Priority    ProjectPriority `json:"priority" gorm:"type:varchar(16);not null;default:'medium';index"`

	// Timing
	StartDate   *time.Time `json:"start_date,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Budget
	BudgetAmount float64 `json:"budget_amount,omitempty" gorm:"type:decimal(12,2);default:0"`
	Currency     string  `json:"currency,omitempty" gorm:"size:3;default:'USD'"`

	// Progress snapshot (denormalized for quick dashboard reads)
	ProgressPercent    int `json:"progress_percent" gorm:"default:0"` // 0-100
	TasksTotal         int `json:"tasks_total" gorm:"default:0"`
	TasksCompleted     int `json:"tasks_completed" gorm:"default:0"`
	TasksOverdue       int `json:"tasks_overdue" gorm:"default:0"`
	ActiveBlockedTasks int `json:"active_blocked_tasks" gorm:"default:0"`

	// Tags and metadata
	Tags     []string `json:"tags,omitempty" gorm:"type:text[]"`
	Metadata JSONB    `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant     *Tenant            `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Artisan    *Artisan           `json:"artisan,omitempty" gorm:"foreignKey:ArtisanID"`
	Customer   *Customer          `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Tasks      []ProjectTask      `json:"tasks,omitempty" gorm:"foreignKey:ProjectID"`
	Milestones []ProjectMilestone `json:"milestones,omitempty" gorm:"foreignKey:ProjectID"`
	Updates    []ProjectUpdate    `json:"updates,omitempty" gorm:"foreignKey:ProjectID"`
}
