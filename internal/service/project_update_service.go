package service

import (
	"context"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// ProjectUpdateService defines the interface for project update operations
type ProjectUpdateService interface {
	// CRUD operations
	CreateProjectUpdate(ctx context.Context, req *dto.CreateProjectUpdateRequest) (*dto.ProjectUpdateResponse, error)
	GetProjectUpdate(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.ProjectUpdateResponse, error)
	UpdateProjectUpdate(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateProjectUpdateRequest) (*dto.ProjectUpdateResponse, error)
	DeleteProjectUpdate(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// Query operations
	ListProjectUpdates(ctx context.Context, filter *dto.ProjectUpdateFilter) (*dto.ProjectUpdateListResponse, error)
	ListByType(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID, updateType models.UpdateType, page, pageSize int) (*dto.ProjectUpdateListResponse, error)
	ListCustomerVisible(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID, page, pageSize int) (*dto.ProjectUpdateListResponse, error)
	GetLatestUpdate(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID) (*dto.ProjectUpdateResponse, error)
}

// projectUpdateService implements ProjectUpdateService
type projectUpdateService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewProjectUpdateService creates a new instance of ProjectUpdateService
func NewProjectUpdateService(repos *repository.Repositories, logger log.AllLogger) ProjectUpdateService {
	return &projectUpdateService{
		repos:  repos,
		logger: logger,
	}
}

// CreateProjectUpdate creates a new project update
func (s *projectUpdateService) CreateProjectUpdate(ctx context.Context, req *dto.CreateProjectUpdateRequest) (*dto.ProjectUpdateResponse, error) {
	// Verify project exists and user has access
	project, err := s.repos.Project.GetByID(ctx, req.ProjectID)
	if err != nil {
		s.logger.Error("failed to get project", "project_id", req.ProjectID, "error", err)
		return nil, errors.NewNotFoundError("project")
	}

	// Verify user exists
	user, err := s.repos.User.GetByID(ctx, req.UserID)
	if err != nil {
		s.logger.Error("failed to get user", "user_id", req.UserID, "error", err)
		return nil, errors.NewNotFoundError("user")
	}

	// Verify user belongs to same tenant (handle nullable TenantID)
	if user.TenantID != nil && *user.TenantID != project.TenantID {
		return nil, errors.NewValidationError("User does not belong to project tenant")
	}

	// Create the update
	update := &models.ProjectUpdate{
		TenantID:          project.TenantID,
		ProjectID:         req.ProjectID,
		UserID:            req.UserID,
		Type:              req.Type,
		Title:             req.Title,
		Description:       req.Description,
		VisibleToCustomer: req.VisibleToCustomer,
		AttachmentURLs:    req.AttachmentURLs,
		Metadata:          req.Metadata,
	}

	if err := s.repos.ProjectUpdate.Create(ctx, update); err != nil {
		s.logger.Error("failed to create project update", "error", err)
		return nil, errors.NewRepositoryError("CREATE_FAILED", "Failed to create project update", err)
	}

	// Load user relationship
	update.User = user

	s.logger.Info("project update created", "update_id", update.ID, "project_id", req.ProjectID, "type", req.Type)
	return dto.ToProjectUpdateResponse(update), nil
}

// GetProjectUpdate retrieves a project update by ID
func (s *projectUpdateService) GetProjectUpdate(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.ProjectUpdateResponse, error) {
	update, err := s.repos.ProjectUpdate.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get project update", "id", id, "error", err)
		return nil, errors.NewNotFoundError("project_update")
	}

	// Verify tenant access
	if update.TenantID != tenantID {
		return nil, errors.NewNotFoundError("project_update")
	}

	// Load relationships
	if update.UserID != uuid.Nil {
		user, err := s.repos.User.GetByID(ctx, update.UserID)
		if err != nil {
			s.logger.Warn("failed to load user", "user_id", update.UserID, "error", err)
		} else {
			update.User = user
		}
	}

	return dto.ToProjectUpdateResponse(update), nil
}

// UpdateProjectUpdate updates an existing project update
func (s *projectUpdateService) UpdateProjectUpdate(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateProjectUpdateRequest) (*dto.ProjectUpdateResponse, error) {
	// Get existing update
	update, err := s.repos.ProjectUpdate.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get project update", "id", id, "error", err)
		return nil, errors.NewNotFoundError("project_update")
	}

	// Verify tenant access
	if update.TenantID != tenantID {
		return nil, errors.NewNotFoundError("project_update")
	}

	// Update fields
	if req.Title != nil {
		update.Title = *req.Title
	}
	if req.Description != nil {
		update.Description = *req.Description
	}
	if req.VisibleToCustomer != nil {
		update.VisibleToCustomer = *req.VisibleToCustomer
	}
	if req.AttachmentURLs != nil {
		update.AttachmentURLs = req.AttachmentURLs
	}
	if req.Metadata != nil {
		update.Metadata = req.Metadata
	}

	if err := s.repos.ProjectUpdate.Update(ctx, update); err != nil {
		s.logger.Error("failed to update project update", "id", id, "error", err)
		return nil, errors.NewRepositoryError("UPDATE_FAILED", "Failed to update project update", err)
	}

	// Load relationships
	if update.UserID != uuid.Nil {
		user, err := s.repos.User.GetByID(ctx, update.UserID)
		if err != nil {
			s.logger.Warn("failed to load user", "user_id", update.UserID, "error", err)
		} else {
			update.User = user
		}
	}

	s.logger.Info("project update updated", "id", id)
	return dto.ToProjectUpdateResponse(update), nil
}

// DeleteProjectUpdate deletes a project update
func (s *projectUpdateService) DeleteProjectUpdate(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// Get existing update
	update, err := s.repos.ProjectUpdate.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get project update", "id", id, "error", err)
		return errors.NewNotFoundError("project_update")
	}

	// Verify tenant access
	if update.TenantID != tenantID {
		return errors.NewNotFoundError("project_update")
	}

	if err := s.repos.ProjectUpdate.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete project update", "id", id, "error", err)
		return errors.NewRepositoryError("DELETE_FAILED", "Failed to delete project update", err)
	}

	s.logger.Info("project update deleted", "id", id)
	return nil
}

// ListProjectUpdates lists project updates with filtering and pagination
func (s *projectUpdateService) ListProjectUpdates(ctx context.Context, filter *dto.ProjectUpdateFilter) (*dto.ProjectUpdateListResponse, error) {
	// Verify project exists
	_, err := s.repos.Project.GetByID(ctx, filter.ProjectID)
	if err != nil {
		s.logger.Error("failed to get project", "project_id", filter.ProjectID, "error", err)
		return nil, errors.NewNotFoundError("project")
	}

	// Set default pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	var updates []*models.ProjectUpdate
	var paginationResult repository.PaginationResult

	// Use appropriate repository method based on filters
	if filter.Type != nil {
		updates, paginationResult, err = s.repos.ProjectUpdate.FindByType(ctx, filter.ProjectID, *filter.Type, pagination)
	} else if filter.VisibleToCustomer != nil && *filter.VisibleToCustomer {
		updates, paginationResult, err = s.repos.ProjectUpdate.FindCustomerVisible(ctx, filter.ProjectID, pagination)
	} else {
		updates, paginationResult, err = s.repos.ProjectUpdate.FindByProjectID(ctx, filter.ProjectID, pagination)
	}

	if err != nil {
		s.logger.Error("failed to list project updates", "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "Failed to list project updates", err)
	}

	return &dto.ProjectUpdateListResponse{
		Updates:     dto.ToProjectUpdateResponses(updates),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListByType lists project updates by type
func (s *projectUpdateService) ListByType(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID, updateType models.UpdateType, page, pageSize int) (*dto.ProjectUpdateListResponse, error) {
	filter := &dto.ProjectUpdateFilter{
		ProjectID: projectID,
		Type:      &updateType,
		Page:      page,
		PageSize:  pageSize,
	}
	return s.ListProjectUpdates(ctx, filter)
}

// ListCustomerVisible lists customer-visible project updates
func (s *projectUpdateService) ListCustomerVisible(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID, page, pageSize int) (*dto.ProjectUpdateListResponse, error) {
	visibleToCustomer := true
	filter := &dto.ProjectUpdateFilter{
		ProjectID:         projectID,
		VisibleToCustomer: &visibleToCustomer,
		Page:              page,
		PageSize:          pageSize,
	}
	return s.ListProjectUpdates(ctx, filter)
}

// GetLatestUpdate retrieves the most recent update for a project
func (s *projectUpdateService) GetLatestUpdate(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID) (*dto.ProjectUpdateResponse, error) {
	// Verify project exists
	project, err := s.repos.Project.GetByID(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to get project", "project_id", projectID, "error", err)
		return nil, errors.NewNotFoundError("project")
	}

	// Verify tenant access
	if project.TenantID != tenantID {
		return nil, errors.NewNotFoundError("project")
	}

	update, err := s.repos.ProjectUpdate.GetLatestUpdate(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to get latest update", "project_id", projectID, "error", err)
		return nil, errors.NewNotFoundError("project_update")
	}

	// Load relationships
	if update.UserID != uuid.Nil {
		user, err := s.repos.User.GetByID(ctx, update.UserID)
		if err != nil {
			s.logger.Warn("failed to load user", "user_id", update.UserID, "error", err)
		} else {
			update.User = user
		}
	}

	return dto.ToProjectUpdateResponse(update), nil
}
