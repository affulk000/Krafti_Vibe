package models

import (
	"time"

	"github.com/google/uuid"
)

type MilestoneStatus string

const (
	MilestoneStatusPending    MilestoneStatus = "pending"
	MilestoneStatusInProgress MilestoneStatus = "in_progress"
	MilestoneStatusCompleted  MilestoneStatus = "completed"
	MilestoneStatusCancelled  MilestoneStatus = "cancelled"
)

type ProjectMilestone struct {
	BaseModel

	// Multi-tenancy
	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	ProjectID uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index"`

	// Milestone Details
	Title       string          `json:"title" gorm:"not null;size:255"`
	Description string          `json:"description,omitempty" gorm:"type:text"`
	Status      MilestoneStatus `json:"status" gorm:"type:varchar(32);not null;default:'pending'"`
	OrderIndex  int             `json:"order_index" gorm:"default:0"`

	// Timing
	StartDate   *time.Time `json:"start_date,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Payment tied to milestone
	PaymentAmount      float64    `json:"payment_amount" gorm:"type:decimal(12,2);default:0"`
	PaymentPercentage  float64    `json:"payment_percentage" gorm:"type:decimal(5,2);default:0"`
	IsPaymentMilestone bool       `json:"is_payment_milestone" gorm:"default:false"`
	PaymentReceived    bool       `json:"payment_received" gorm:"default:false"`
	PaymentReceivedAt  *time.Time `json:"payment_received_at,omitempty"`

	// Deliverables
	Deliverables    []string `json:"deliverables,omitempty" gorm:"type:text[]"`
	AttachmentURLs  []string `json:"attachment_urls,omitempty" gorm:"type:text[]"`
	CompletionProof []string `json:"completion_proof,omitempty" gorm:"type:text[]"`

	// Customer approval
	RequiresApproval   bool       `json:"requires_approval" gorm:"default:false"`
	ApprovedByCustomer bool       `json:"approved_by_customer" gorm:"default:false"`
	ApprovedAt         *time.Time `json:"approved_at,omitempty"`
	RejectionReason    string     `json:"rejection_reason,omitempty" gorm:"type:text"`

	// Notes
	ArtisanNotes  string `json:"artisan_notes,omitempty" gorm:"type:text"`
	CustomerNotes string `json:"customer_notes,omitempty" gorm:"type:text"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Project *Project      `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Tasks   []ProjectTask `json:"tasks,omitempty" gorm:"foreignKey:MilestoneID"`
}

// Business Methods
func (m *ProjectMilestone) IsCompleted() bool {
	return m.Status == MilestoneStatusCompleted
}

func (m *ProjectMilestone) IsOverdue() bool {
	if m.DueDate == nil || m.IsCompleted() {
		return false
	}
	return time.Now().After(*m.DueDate)
}

func (m *ProjectMilestone) CanComplete() bool {
	if m.RequiresApproval {
		return m.ApprovedByCustomer
	}
	return true
}
