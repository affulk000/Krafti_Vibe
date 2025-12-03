package service

import (
	"context"
	"encoding/json"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// MilestoneService defines the interface for milestone operations
type MilestoneService interface {
	// CRUD Operations
	CreateMilestone(ctx context.Context, req *dto.CreateMilestoneRequest) (*dto.MilestoneResponse, error)
	GetMilestone(ctx context.Context, id uuid.UUID) (*dto.MilestoneResponse, error)
	UpdateMilestone(ctx context.Context, id uuid.UUID, req *dto.UpdateMilestoneRequest) (*dto.MilestoneResponse, error)
	DeleteMilestone(ctx context.Context, id uuid.UUID) error

	// Query Operations
	ListMilestonesByProject(ctx context.Context, projectID uuid.UUID) ([]*dto.MilestoneResponse, error)
	ListMilestonesByStatus(ctx context.Context, projectID uuid.UUID, status models.MilestoneStatus) ([]*dto.MilestoneResponse, error)
	ListOverdueMilestones(ctx context.Context, projectID uuid.UUID) ([]*dto.MilestoneResponse, error)
	ListUpcomingMilestones(ctx context.Context, projectID uuid.UUID, days int) ([]*dto.MilestoneResponse, error)
	ListPaymentMilestones(ctx context.Context, projectID uuid.UUID) ([]*dto.MilestoneResponse, error)

	// Status Operations
	UpdateMilestoneStatus(ctx context.Context, id uuid.UUID, status models.MilestoneStatus) error
	CompleteMilestone(ctx context.Context, id uuid.UUID) error

	// Payment Operations
	MarkPaymentReceived(ctx context.Context, id uuid.UUID) error

	// Approval Operations
	ApproveMilestone(ctx context.Context, id uuid.UUID, req *dto.ApproveMilestoneRequest) error
	RejectMilestone(ctx context.Context, id uuid.UUID, req *dto.RejectMilestoneRequest) error

	// Order Operations
	ReorderMilestones(ctx context.Context, req *dto.ReorderMilestonesRequest) error

	// Statistics
	GetMilestoneStats(ctx context.Context, projectID uuid.UUID) (*dto.MilestoneStatsResponse, error)
}

// milestoneService implements MilestoneService
type milestoneService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewMilestoneService creates a new milestone service
func NewMilestoneService(repos *repository.Repositories, logger log.AllLogger) MilestoneService {
	return &milestoneService{
		repos:  repos,
		logger: logger,
	}
}

// CreateMilestone creates a new milestone
func (s *milestoneService) CreateMilestone(ctx context.Context, req *dto.CreateMilestoneRequest) (*dto.MilestoneResponse, error) {
	// Verify project exists
	project, err := s.repos.Project.GetByID(ctx, req.ProjectID)
	if err != nil {
		s.logger.Error("failed to find project", "project_id", req.ProjectID, "error", err)
		return nil, errors.NewNotFoundError("project not found")
	}

	// Convert metadata
	metadata := models.JSONB{}
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, errors.NewValidationError("invalid metadata: " + err.Error())
		}
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, errors.NewValidationError("failed to convert metadata: " + err.Error())
		}
	}

	milestone := &models.ProjectMilestone{
		TenantID:           project.TenantID,
		ProjectID:          req.ProjectID,
		Title:              req.Title,
		Description:        req.Description,
		Status:             models.MilestoneStatusPending,
		StartDate:          req.StartDate,
		DueDate:            req.DueDate,
		PaymentAmount:      req.PaymentAmount,
		PaymentPercentage:  req.PaymentPercentage,
		IsPaymentMilestone: req.IsPaymentMilestone,
		RequiresApproval:   req.RequiresApproval,
		Deliverables:       req.Deliverables,
		ArtisanNotes:       req.ArtisanNotes,
		Metadata:           metadata,
	}

	if err := s.repos.ProjectMilestone.Create(ctx, milestone); err != nil {
		s.logger.Error("failed to create milestone", "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "failed to create milestone", err)
	}

	s.logger.Info("milestone created", "milestone_id", milestone.ID, "project_id", req.ProjectID)

	// Reload with relationships
	created, err := s.repos.ProjectMilestone.GetByID(ctx, milestone.ID)
	if err != nil {
		return dto.ToMilestoneResponse(milestone), nil
	}

	return dto.ToMilestoneResponse(created), nil
}

// GetMilestone retrieves a milestone by ID
func (s *milestoneService) GetMilestone(ctx context.Context, id uuid.UUID) (*dto.MilestoneResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("milestone_id is required")
	}

	milestone, err := s.repos.ProjectMilestone.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get milestone", "milestone_id", id, "error", err)
		return nil, errors.NewNotFoundError("milestone not found")
	}

	return dto.ToMilestoneResponse(milestone), nil
}

// UpdateMilestone updates a milestone
func (s *milestoneService) UpdateMilestone(ctx context.Context, id uuid.UUID, req *dto.UpdateMilestoneRequest) (*dto.MilestoneResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("milestone_id is required")
	}

	// Get existing milestone
	existing, err := s.repos.ProjectMilestone.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find milestone", "milestone_id", id, "error", err)
		return nil, errors.NewNotFoundError("milestone not found")
	}

	// Apply updates
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.StartDate != nil {
		existing.StartDate = req.StartDate
	}
	if req.DueDate != nil {
		existing.DueDate = req.DueDate
	}
	if req.PaymentAmount != nil {
		existing.PaymentAmount = *req.PaymentAmount
	}
	if req.PaymentPercentage != nil {
		existing.PaymentPercentage = *req.PaymentPercentage
	}
	if req.IsPaymentMilestone != nil {
		existing.IsPaymentMilestone = *req.IsPaymentMilestone
	}
	if req.RequiresApproval != nil {
		existing.RequiresApproval = *req.RequiresApproval
	}
	if req.Deliverables != nil {
		existing.Deliverables = req.Deliverables
	}
	if req.AttachmentURLs != nil {
		existing.AttachmentURLs = req.AttachmentURLs
	}
	if req.CompletionProof != nil {
		existing.CompletionProof = req.CompletionProof
	}
	if req.ArtisanNotes != nil {
		existing.ArtisanNotes = *req.ArtisanNotes
	}
	if req.CustomerNotes != nil {
		existing.CustomerNotes = *req.CustomerNotes
	}
	if req.Metadata != nil {
		metadata := models.JSONB{}
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, errors.NewValidationError("invalid metadata: " + err.Error())
		}
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, errors.NewValidationError("failed to convert metadata: " + err.Error())
		}
		existing.Metadata = metadata
	}

	if err := s.repos.ProjectMilestone.Update(ctx, existing); err != nil {
		s.logger.Error("failed to update milestone", "milestone_id", id, "error", err)
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to update milestone", err)
	}

	s.logger.Info("milestone updated", "milestone_id", id)

	// Get updated milestone
	updated, err := s.repos.ProjectMilestone.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "failed to retrieve updated milestone", err)
	}

	return dto.ToMilestoneResponse(updated), nil
}

// DeleteMilestone deletes a milestone
func (s *milestoneService) DeleteMilestone(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewValidationError("milestone_id is required")
	}

	if err := s.repos.ProjectMilestone.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete milestone", "milestone_id", id, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "failed to delete milestone", err)
	}

	s.logger.Info("milestone deleted", "milestone_id", id)
	return nil
}

// ListMilestonesByProject retrieves all milestones for a project
func (s *milestoneService) ListMilestonesByProject(ctx context.Context, projectID uuid.UUID) ([]*dto.MilestoneResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	milestones, err := s.repos.ProjectMilestone.FindByProjectID(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to list milestones", "project_id", projectID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list milestones", err)
	}

	return dto.ToMilestoneResponses(milestones), nil
}

// ListMilestonesByStatus retrieves milestones by status
func (s *milestoneService) ListMilestonesByStatus(ctx context.Context, projectID uuid.UUID, status models.MilestoneStatus) ([]*dto.MilestoneResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	milestones, err := s.repos.ProjectMilestone.FindByStatus(ctx, projectID, status)
	if err != nil {
		s.logger.Error("failed to list milestones by status", "project_id", projectID, "status", status, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list milestones", err)
	}

	return dto.ToMilestoneResponses(milestones), nil
}

// ListOverdueMilestones retrieves overdue milestones
func (s *milestoneService) ListOverdueMilestones(ctx context.Context, projectID uuid.UUID) ([]*dto.MilestoneResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	milestones, err := s.repos.ProjectMilestone.FindOverdueMilestones(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to list overdue milestones", "project_id", projectID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list overdue milestones", err)
	}

	return dto.ToMilestoneResponses(milestones), nil
}

// ListUpcomingMilestones retrieves upcoming milestones
func (s *milestoneService) ListUpcomingMilestones(ctx context.Context, projectID uuid.UUID, days int) ([]*dto.MilestoneResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	milestones, err := s.repos.ProjectMilestone.FindUpcomingMilestones(ctx, projectID, days)
	if err != nil {
		s.logger.Error("failed to list upcoming milestones", "project_id", projectID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list upcoming milestones", err)
	}

	return dto.ToMilestoneResponses(milestones), nil
}

// ListPaymentMilestones retrieves payment milestones
func (s *milestoneService) ListPaymentMilestones(ctx context.Context, projectID uuid.UUID) ([]*dto.MilestoneResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	milestones, err := s.repos.ProjectMilestone.GetPaymentMilestones(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to list payment milestones", "project_id", projectID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list payment milestones", err)
	}

	return dto.ToMilestoneResponses(milestones), nil
}

// UpdateMilestoneStatus updates milestone status
func (s *milestoneService) UpdateMilestoneStatus(ctx context.Context, id uuid.UUID, status models.MilestoneStatus) error {
	if id == uuid.Nil {
		return errors.NewValidationError("milestone_id is required")
	}

	if err := s.repos.ProjectMilestone.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("failed to update milestone status", "milestone_id", id, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "failed to update status", err)
	}

	s.logger.Info("milestone status updated", "milestone_id", id, "status", status)
	return nil
}

// CompleteMilestone marks a milestone as completed
func (s *milestoneService) CompleteMilestone(ctx context.Context, id uuid.UUID) error {
	return s.UpdateMilestoneStatus(ctx, id, models.MilestoneStatusCompleted)
}

// MarkPaymentReceived marks payment as received
func (s *milestoneService) MarkPaymentReceived(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewValidationError("milestone_id is required")
	}

	if err := s.repos.ProjectMilestone.MarkPaymentReceived(ctx, id); err != nil {
		s.logger.Error("failed to mark payment received", "milestone_id", id, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "failed to mark payment received", err)
	}

	s.logger.Info("payment marked as received", "milestone_id", id)
	return nil
}

// ApproveMilestone approves a milestone
func (s *milestoneService) ApproveMilestone(ctx context.Context, id uuid.UUID, req *dto.ApproveMilestoneRequest) error {
	if id == uuid.Nil {
		return errors.NewValidationError("milestone_id is required")
	}

	if err := s.repos.ProjectMilestone.ApproveMilestone(ctx, id); err != nil {
		s.logger.Error("failed to approve milestone", "milestone_id", id, "error", err)
		return errors.NewServiceError("APPROVE_FAILED", "failed to approve milestone", err)
	}

	// Update customer notes if provided
	if req.CustomerNotes != "" {
		milestone, err := s.repos.ProjectMilestone.GetByID(ctx, id)
		if err == nil {
			milestone.CustomerNotes = req.CustomerNotes
			s.repos.ProjectMilestone.Update(ctx, milestone)
		}
	}

	s.logger.Info("milestone approved", "milestone_id", id)
	return nil
}

// RejectMilestone rejects a milestone
func (s *milestoneService) RejectMilestone(ctx context.Context, id uuid.UUID, req *dto.RejectMilestoneRequest) error {
	if id == uuid.Nil {
		return errors.NewValidationError("milestone_id is required")
	}

	if err := s.repos.ProjectMilestone.RejectMilestone(ctx, id, req.RejectionReason); err != nil {
		s.logger.Error("failed to reject milestone", "milestone_id", id, "error", err)
		return errors.NewServiceError("REJECT_FAILED", "failed to reject milestone", err)
	}

	// Update customer notes if provided
	if req.CustomerNotes != "" {
		milestone, err := s.repos.ProjectMilestone.GetByID(ctx, id)
		if err == nil {
			milestone.CustomerNotes = req.CustomerNotes
			s.repos.ProjectMilestone.Update(ctx, milestone)
		}
	}

	s.logger.Info("milestone rejected", "milestone_id", id, "reason", req.RejectionReason)
	return nil
}

// ReorderMilestones reorders milestones
func (s *milestoneService) ReorderMilestones(ctx context.Context, req *dto.ReorderMilestonesRequest) error {
	if req.ProjectID == uuid.Nil {
		return errors.NewValidationError("project_id is required")
	}

	// Convert string keys to UUIDs
	milestoneOrders := make(map[uuid.UUID]int)
	for milestoneIDStr, orderIndex := range req.MilestoneOrders {
		milestoneID, err := uuid.Parse(milestoneIDStr)
		if err != nil {
			return errors.NewValidationError("invalid milestone_id: " + milestoneIDStr)
		}
		milestoneOrders[milestoneID] = orderIndex
	}

	if err := s.repos.ProjectMilestone.ReorderMilestones(ctx, req.ProjectID, milestoneOrders); err != nil {
		s.logger.Error("failed to reorder milestones", "project_id", req.ProjectID, "error", err)
		return errors.NewServiceError("REORDER_FAILED", "failed to reorder milestones", err)
	}

	s.logger.Info("milestones reordered", "project_id", req.ProjectID, "count", len(milestoneOrders))
	return nil
}

// GetMilestoneStats retrieves milestone statistics for a project
func (s *milestoneService) GetMilestoneStats(ctx context.Context, projectID uuid.UUID) (*dto.MilestoneStatsResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	stats, err := s.repos.ProjectMilestone.GetMilestoneStats(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to get milestone stats", "project_id", projectID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "failed to get milestone stats", err)
	}

	return &dto.MilestoneStatsResponse{
		ProjectID:            projectID,
		TotalMilestones:      stats.TotalMilestones,
		PendingMilestones:    stats.PendingMilestones,
		InProgressMilestones: stats.InProgressMilestones,
		CompletedMilestones:  stats.CompletedMilestones,
		OverdueMilestones:    stats.OverdueMilestones,
		TotalPaymentAmount:   stats.TotalPaymentAmount,
		ReceivedPayment:      stats.ReceivedPayment,
		CompletionRate:       stats.CompletionRate,
	}, nil
}
