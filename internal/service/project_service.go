package service

import (
	"context"
	"encoding/json"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// ProjectServiceEnhanced defines the enhanced project service interface
type ProjectService interface {
	// CRUD Operations
	CreateProject(ctx context.Context, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error)
	GetProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)
	GetProjectWithDetails(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)
	UpdateProject(ctx context.Context, id uuid.UUID, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error

	// Query Operations
	ListProjects(ctx context.Context, filter dto.ProjectFilter) (*dto.ProjectListResponse, error)
	SearchProjects(ctx context.Context, tenantID uuid.UUID, query string, pagination repository.PaginationParams) (*dto.ProjectListResponse, error)
	GetProjectsByArtisan(ctx context.Context, artisanID uuid.UUID, pagination repository.PaginationParams) (*dto.ProjectListResponse, error)
	GetProjectsByCustomer(ctx context.Context, customerID uuid.UUID, pagination repository.PaginationParams) (*dto.ProjectListResponse, error)
	GetProjectsByStatus(ctx context.Context, tenantID uuid.UUID, status models.ProjectStatus, pagination repository.PaginationParams) (*dto.ProjectListResponse, error)
	GetOverdueProjects(ctx context.Context, tenantID uuid.UUID) ([]*dto.ProjectResponse, error)
	GetActiveProjects(ctx context.Context, tenantID uuid.UUID) ([]*dto.ProjectResponse, error)

	// Status Management
	StartProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)
	PauseProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)
	CompleteProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)
	CancelProject(ctx context.Context, id uuid.UUID, reason string) (*dto.ProjectResponse, error)
	ResumeProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error)

	// Progress Management
	UpdateProgress(ctx context.Context, id uuid.UUID) error
	RecalculateAllProgress(ctx context.Context, tenantID uuid.UUID) error

	// Analytics & Statistics
	GetProjectStats(ctx context.Context, tenantID uuid.UUID) (*dto.ProjectStatsResponse, error)
	GetArtisanProjectStats(ctx context.Context, artisanID uuid.UUID) (*dto.ArtisanProjectStatsResponse, error)
	GetCustomerProjectStats(ctx context.Context, customerID uuid.UUID) (*dto.CustomerProjectStatsResponse, error)
	GetProjectHealth(ctx context.Context, id uuid.UUID) (*dto.ProjectHealthResponse, error)
	GetProjectTimeline(ctx context.Context, id uuid.UUID) (*dto.ProjectTimelineResponse, error)

	// Dashboard
	GetArtisanDashboard(ctx context.Context, artisanID uuid.UUID) (*dto.ArtisanDashboardResponse, error)
	GetTenantDashboard(ctx context.Context, tenantID uuid.UUID) (*dto.TenantProjectDashboardResponse, error)

	// Bulk Operations
	BulkUpdateStatus(ctx context.Context, req *dto.BulkProjectUpdateRequest) error
	BulkAssignArtisan(ctx context.Context, req *dto.BulkProjectUpdateRequest) error
	ArchiveCompletedProjects(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error

	// Health Check
	HealthCheck(ctx context.Context) error
}

// projectServiceEnhanced implements ProjectServiceEnhanced
type projectService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewProjectServiceEnhanced creates a new ProjectServiceEnhanced instance
func NewProjectService(repos *repository.Repositories, logger log.AllLogger) ProjectService {
	return &projectService{
		repos:  repos,
		logger: logger,
	}
}

// CreateProject creates a new project
func (s *projectService) CreateProject(ctx context.Context, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
	if err := req.Validate(); err != nil {
		s.logger.Warn("invalid create project request", "error", err)
		return nil, errors.NewValidationError(err.Error())
	}

	// Verify artisan exists
	artisan, err := s.repos.Artisan.GetByID(ctx, req.ArtisanID)
	if err != nil {
		s.logger.Error("failed to find artisan", "artisan_id", req.ArtisanID, "error", err)
		return nil, errors.NewNotFoundError("artisan not found")
	}

	// Verify tenant matches
	if artisan.TenantID != req.TenantID {
		s.logger.Warn("artisan tenant mismatch", "artisan_tenant", artisan.TenantID, "request_tenant", req.TenantID)
		return nil, errors.NewValidationError("artisan does not belong to tenant")
	}

	// Verify customer if provided
	if req.CustomerID != nil {
		customer, err := s.repos.Customer.GetByID(ctx, *req.CustomerID)
		if err != nil {
			s.logger.Error("failed to find customer", "customer_id", *req.CustomerID, "error", err)
			return nil, errors.NewNotFoundError("customer not found")
		}
		if customer.TenantID != req.TenantID {
			s.logger.Warn("customer tenant mismatch", "customer_tenant", customer.TenantID, "request_tenant", req.TenantID)
			return nil, errors.NewValidationError("customer does not belong to tenant")
		}
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

	// Set default currency
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	project := &models.Project{
		TenantID:     req.TenantID,
		ArtisanID:    req.ArtisanID,
		CustomerID:   req.CustomerID,
		Title:        req.Title,
		Description:  req.Description,
		Status:       models.ProjectStatusPlanned,
		Priority:     req.Priority,
		StartDate:    req.StartDate,
		DueDate:      req.DueDate,
		BudgetAmount: req.BudgetAmount,
		Currency:     currency,
		Tags:         req.Tags,
		Metadata:     metadata,
	}

	if err := s.repos.Project.Create(ctx, project); err != nil {
		s.logger.Error("failed to create project", "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "failed to create project", err)
	}

	s.logger.Info("project created", "project_id", project.ID, "tenant_id", req.TenantID)

	// Reload with relationships
	created, err := s.repos.Project.GetProjectWithFullDetails(ctx, project.ID)
	if err != nil {
		return dto.ToProjectResponse(project), nil
	}

	return dto.ToProjectResponse(created), nil
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	project, err := s.repos.Project.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get project", "project_id", id, "error", err)
		return nil, errors.NewNotFoundError("project not found")
	}

	return dto.ToProjectResponse(project), nil
}

// GetProjectWithDetails retrieves a project with full details
func (s *projectService) GetProjectWithDetails(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	project, err := s.repos.Project.GetProjectWithFullDetails(ctx, id)
	if err != nil {
		s.logger.Error("failed to get project with details", "project_id", id, "error", err)
		return nil, errors.NewNotFoundError("project not found")
	}

	return dto.ToProjectResponse(project), nil
}

// UpdateProject updates a project
func (s *projectService) UpdateProject(ctx context.Context, id uuid.UUID, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	// Get existing project
	existing, err := s.repos.Project.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find project", "project_id", id, "error", err)
		return nil, errors.NewNotFoundError("project not found")
	}

	// Apply updates
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Priority != nil {
		existing.Priority = *req.Priority
	}
	if req.StartDate != nil {
		existing.StartDate = req.StartDate
	}
	if req.DueDate != nil {
		// Validate date range
		startDate := existing.StartDate
		if req.StartDate != nil {
			startDate = req.StartDate
		}
		if startDate != nil && req.DueDate.Before(*startDate) {
			return nil, errors.NewValidationError("due_date must be after start_date")
		}
		existing.DueDate = req.DueDate
	}
	if req.BudgetAmount != nil {
		if *req.BudgetAmount < 0 {
			return nil, errors.NewValidationError("budget_amount cannot be negative")
		}
		existing.BudgetAmount = *req.BudgetAmount
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
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

	if err := s.repos.Project.Update(ctx, existing); err != nil {
		s.logger.Error("failed to update project", "project_id", id, "error", err)
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to update project", err)
	}

	s.logger.Info("project updated", "project_id", id)

	// Get updated project
	updated, err := s.repos.Project.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "failed to retrieve updated project", err)
	}

	return dto.ToProjectResponse(updated), nil
}

// DeleteProject deletes a project
func (s *projectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewValidationError("project_id is required")
	}

	if err := s.repos.Project.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete project", "project_id", id, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "failed to delete project", err)
	}

	s.logger.Info("project deleted", "project_id", id)
	return nil
}

// ListProjects retrieves a paginated list of projects
func (s *projectService) ListProjects(ctx context.Context, filter dto.ProjectFilter) (*dto.ProjectListResponse, error) {
	// Platform admins can query across all tenants (tenant_id can be nil)
	// Tenant users will have their tenant_id set by the handler

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}
	pagination.Validate()

	var projects []*models.Project
	var paginationResult repository.PaginationResult
	var err error

	// Apply filters
	if filter.ArtisanID != nil {
		projects, paginationResult, err = s.repos.Project.FindByArtisanID(ctx, *filter.ArtisanID, pagination)
	} else if filter.CustomerID != nil {
		projects, paginationResult, err = s.repos.Project.FindByCustomerID(ctx, *filter.CustomerID, pagination)
	} else if filter.TenantID != uuid.Nil {
		// Tenant-specific queries
		if filter.Status != nil {
			projects, paginationResult, err = s.repos.Project.FindByStatus(ctx, filter.TenantID, *filter.Status, pagination)
		} else if filter.Priority != nil {
			projects, paginationResult, err = s.repos.Project.FindByPriority(ctx, filter.TenantID, *filter.Priority, pagination)
		} else if len(filter.Tags) > 0 {
			projects, paginationResult, err = s.repos.Project.FindByTags(ctx, filter.TenantID, filter.Tags, pagination)
		} else if filter.FromDate != nil && filter.ToDate != nil {
			projects, paginationResult, err = s.repos.Project.FindByDateRange(ctx, filter.TenantID, *filter.FromDate, *filter.ToDate, pagination)
		} else {
			projects, paginationResult, err = s.repos.Project.FindByTenantID(ctx, filter.TenantID, pagination)
		}
	} else {
		// Platform admin querying all projects across all tenants
		// This would require a new repository method or use an existing one
		// For now, return empty list as this feature needs to be implemented at repo level
		s.logger.Warn("platform admin query for all projects not yet implemented at repository level")
		projects = []*models.Project{}
		paginationResult = repository.PaginationResult{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: 0,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}
	}

	if err != nil {
		s.logger.Error("failed to list projects", "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list projects", err)
	}

	return &dto.ProjectListResponse{
		Projects:    dto.ToProjectResponses(projects),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// SearchProjects searches projects
func (s *projectService) SearchProjects(ctx context.Context, tenantID uuid.UUID, query string, pagination repository.PaginationParams) (*dto.ProjectListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}

	pagination.Validate()

	projects, paginationResult, err := s.repos.Project.Search(ctx, tenantID, query, pagination)
	if err != nil {
		s.logger.Error("failed to search projects", "error", err)
		return nil, errors.NewServiceError("SEARCH_FAILED", "failed to search projects", err)
	}

	return &dto.ProjectListResponse{
		Projects:    dto.ToProjectResponses(projects),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetProjectsByArtisan retrieves projects by artisan
func (s *projectService) GetProjectsByArtisan(ctx context.Context, artisanID uuid.UUID, pagination repository.PaginationParams) (*dto.ProjectListResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan_id is required")
	}

	pagination.Validate()

	projects, paginationResult, err := s.repos.Project.FindByArtisanID(ctx, artisanID, pagination)
	if err != nil {
		s.logger.Error("failed to get projects by artisan", "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "failed to get projects", err)
	}

	return &dto.ProjectListResponse{
		Projects:    dto.ToProjectResponses(projects),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetProjectsByCustomer retrieves projects by customer
func (s *projectService) GetProjectsByCustomer(ctx context.Context, customerID uuid.UUID, pagination repository.PaginationParams) (*dto.ProjectListResponse, error) {
	if customerID == uuid.Nil {
		return nil, errors.NewValidationError("customer_id is required")
	}

	pagination.Validate()

	projects, paginationResult, err := s.repos.Project.FindByCustomerID(ctx, customerID, pagination)
	if err != nil {
		s.logger.Error("failed to get projects by customer", "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "failed to get projects", err)
	}

	return &dto.ProjectListResponse{
		Projects:    dto.ToProjectResponses(projects),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetProjectsByStatus retrieves projects by status
func (s *projectService) GetProjectsByStatus(ctx context.Context, tenantID uuid.UUID, status models.ProjectStatus, pagination repository.PaginationParams) (*dto.ProjectListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}

	pagination.Validate()

	projects, paginationResult, err := s.repos.Project.FindByStatus(ctx, tenantID, status, pagination)
	if err != nil {
		s.logger.Error("failed to get projects by status", "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "failed to get projects", err)
	}

	return &dto.ProjectListResponse{
		Projects:    dto.ToProjectResponses(projects),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetOverdueProjects retrieves overdue projects
func (s *projectService) GetOverdueProjects(ctx context.Context, tenantID uuid.UUID) ([]*dto.ProjectResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}

	projects, err := s.repos.Project.FindOverdueProjects(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get overdue projects", "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "failed to get overdue projects", err)
	}

	return dto.ToProjectResponses(projects), nil
}

// GetActiveProjects retrieves active projects
func (s *projectService) GetActiveProjects(ctx context.Context, tenantID uuid.UUID) ([]*dto.ProjectResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}

	projects, err := s.repos.Project.FindActiveProjects(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get active projects", "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "failed to get active projects", err)
	}

	return dto.ToProjectResponses(projects), nil
}

// StartProject starts a project
func (s *projectService) StartProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	if err := s.repos.Project.StartProject(ctx, id); err != nil {
		s.logger.Error("failed to start project", "project_id", id, "error", err)
		return nil, errors.NewServiceError("START_FAILED", "failed to start project", err)
	}

	s.logger.Info("project started", "project_id", id)

	project, err := s.repos.Project.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "failed to retrieve project", err)
	}

	return dto.ToProjectResponse(project), nil
}

// PauseProject pauses a project
func (s *projectService) PauseProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	if err := s.repos.Project.PauseProject(ctx, id); err != nil {
		s.logger.Error("failed to pause project", "project_id", id, "error", err)
		return nil, errors.NewServiceError("PAUSE_FAILED", "failed to pause project", err)
	}

	s.logger.Info("project paused", "project_id", id)

	project, err := s.repos.Project.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "failed to retrieve project", err)
	}

	return dto.ToProjectResponse(project), nil
}

// CompleteProject completes a project
func (s *projectService) CompleteProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	if err := s.repos.Project.CompleteProject(ctx, id); err != nil {
		s.logger.Error("failed to complete project", "project_id", id, "error", err)
		return nil, errors.NewServiceError("COMPLETE_FAILED", "failed to complete project", err)
	}

	s.logger.Info("project completed", "project_id", id)

	project, err := s.repos.Project.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "failed to retrieve project", err)
	}

	return dto.ToProjectResponse(project), nil
}

// CancelProject cancels a project
func (s *projectService) CancelProject(ctx context.Context, id uuid.UUID, reason string) (*dto.ProjectResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	if err := s.repos.Project.CancelProject(ctx, id, reason); err != nil {
		s.logger.Error("failed to cancel project", "project_id", id, "error", err)
		return nil, errors.NewServiceError("CANCEL_FAILED", "failed to cancel project", err)
	}

	s.logger.Info("project cancelled", "project_id", id, "reason", reason)

	project, err := s.repos.Project.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "failed to retrieve project", err)
	}

	return dto.ToProjectResponse(project), nil
}

// ResumeProject resumes a paused project
func (s *projectService) ResumeProject(ctx context.Context, id uuid.UUID) (*dto.ProjectResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	if err := s.repos.Project.ResumeProject(ctx, id); err != nil {
		s.logger.Error("failed to resume project", "project_id", id, "error", err)
		return nil, errors.NewServiceError("RESUME_FAILED", "failed to resume project", err)
	}

	s.logger.Info("project resumed", "project_id", id)

	project, err := s.repos.Project.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "failed to retrieve project", err)
	}

	return dto.ToProjectResponse(project), nil
}

// UpdateProgress updates project progress
func (s *projectService) UpdateProgress(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewValidationError("project_id is required")
	}

	if err := s.repos.Project.UpdateProgress(ctx, id); err != nil {
		s.logger.Error("failed to update project progress", "project_id", id, "error", err)
		return errors.NewServiceError("UPDATE_PROGRESS_FAILED", "failed to update progress", err)
	}

	s.logger.Info("project progress updated", "project_id", id)
	return nil
}

// RecalculateAllProgress recalculates progress for all projects
func (s *projectService) RecalculateAllProgress(ctx context.Context, tenantID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return errors.NewValidationError("tenant_id is required")
	}

	if err := s.repos.Project.RecalculateAllProgress(ctx, tenantID); err != nil {
		s.logger.Error("failed to recalculate all progress", "tenant_id", tenantID, "error", err)
		return errors.NewServiceError("RECALCULATE_FAILED", "failed to recalculate progress", err)
	}

	s.logger.Info("all project progress recalculated", "tenant_id", tenantID)
	return nil
}

// GetProjectStats retrieves project statistics
func (s *projectService) GetProjectStats(ctx context.Context, tenantID uuid.UUID) (*dto.ProjectStatsResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}

	stats, err := s.repos.Project.GetProjectStats(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get project stats", "tenant_id", tenantID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "failed to get project stats", err)
	}

	return dto.ToProjectStatsResponse(stats), nil
}

// GetArtisanProjectStats retrieves artisan project statistics
func (s *projectService) GetArtisanProjectStats(ctx context.Context, artisanID uuid.UUID) (*dto.ArtisanProjectStatsResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan_id is required")
	}

	stats, err := s.repos.Project.GetArtisanProjectStats(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get artisan project stats", "artisan_id", artisanID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "failed to get artisan project stats", err)
	}

	return dto.ToArtisanProjectStatsResponse(artisanID, stats), nil
}

// GetCustomerProjectStats retrieves customer project statistics
func (s *projectService) GetCustomerProjectStats(ctx context.Context, customerID uuid.UUID) (*dto.CustomerProjectStatsResponse, error) {
	if customerID == uuid.Nil {
		return nil, errors.NewValidationError("customer_id is required")
	}

	stats, err := s.repos.Project.GetCustomerProjectStats(ctx, customerID)
	if err != nil {
		s.logger.Error("failed to get customer project stats", "customer_id", customerID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "failed to get customer project stats", err)
	}

	return dto.ToCustomerProjectStatsResponse(customerID, stats), nil
}

// GetProjectHealth retrieves project health metrics
func (s *projectService) GetProjectHealth(ctx context.Context, id uuid.UUID) (*dto.ProjectHealthResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	health, err := s.repos.Project.GetProjectHealth(ctx, id)
	if err != nil {
		s.logger.Error("failed to get project health", "project_id", id, "error", err)
		return nil, errors.NewServiceError("HEALTH_FAILED", "failed to get project health", err)
	}

	return dto.ToProjectHealthResponse(health), nil
}

// GetProjectTimeline retrieves project timeline
func (s *projectService) GetProjectTimeline(ctx context.Context, id uuid.UUID) (*dto.ProjectTimelineResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	events, err := s.repos.Project.GetProjectTimeline(ctx, id)
	if err != nil {
		s.logger.Error("failed to get project timeline", "project_id", id, "error", err)
		return nil, errors.NewServiceError("TIMELINE_FAILED", "failed to get project timeline", err)
	}

	return &dto.ProjectTimelineResponse{
		Events: dto.ToTimelineEventResponses(events),
	}, nil
}

// GetArtisanDashboard retrieves artisan dashboard
func (s *projectService) GetArtisanDashboard(ctx context.Context, artisanID uuid.UUID) (*dto.ArtisanDashboardResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan_id is required")
	}

	dashboard, err := s.repos.Project.GetArtisanDashboard(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get artisan dashboard", "artisan_id", artisanID, "error", err)
		return nil, errors.NewServiceError("DASHBOARD_FAILED", "failed to get artisan dashboard", err)
	}

	return dto.ToArtisanDashboardResponse(dashboard, artisanID), nil
}

// GetTenantDashboard retrieves tenant dashboard
func (s *projectService) GetTenantDashboard(ctx context.Context, tenantID uuid.UUID) (*dto.TenantProjectDashboardResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}

	dashboard, err := s.repos.Project.GetTenantDashboard(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get tenant dashboard", "tenant_id", tenantID, "error", err)
		return nil, errors.NewServiceError("DASHBOARD_FAILED", "failed to get tenant dashboard", err)
	}

	return dto.ToTenantProjectDashboardResponse(dashboard), nil
}

// BulkUpdateStatus updates status for multiple projects
func (s *projectService) BulkUpdateStatus(ctx context.Context, req *dto.BulkProjectUpdateRequest) error {
	if len(req.ProjectIDs) == 0 {
		return errors.NewValidationError("project_ids is required")
	}
	if req.Status == nil {
		return errors.NewValidationError("status is required")
	}

	if err := s.repos.Project.BulkUpdateStatus(ctx, req.ProjectIDs, *req.Status); err != nil {
		s.logger.Error("failed to bulk update project status", "error", err)
		return errors.NewServiceError("BULK_UPDATE_FAILED", "failed to bulk update status", err)
	}

	s.logger.Info("bulk updated project status", "count", len(req.ProjectIDs), "status", *req.Status)
	return nil
}

// BulkAssignArtisan assigns multiple projects to an artisan
func (s *projectService) BulkAssignArtisan(ctx context.Context, req *dto.BulkProjectUpdateRequest) error {
	if len(req.ProjectIDs) == 0 {
		return errors.NewValidationError("project_ids is required")
	}
	if req.ArtisanID == nil {
		return errors.NewValidationError("artisan_id is required")
	}

	if err := s.repos.Project.BulkAssignArtisan(ctx, req.ProjectIDs, *req.ArtisanID); err != nil {
		s.logger.Error("failed to bulk assign artisan", "error", err)
		return errors.NewServiceError("BULK_ASSIGN_FAILED", "failed to bulk assign artisan", err)
	}

	s.logger.Info("bulk assigned artisan to projects", "count", len(req.ProjectIDs), "artisan_id", *req.ArtisanID)
	return nil
}

// ArchiveCompletedProjects archives completed projects
func (s *projectService) ArchiveCompletedProjects(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error {
	if tenantID == uuid.Nil {
		return errors.NewValidationError("tenant_id is required")
	}

	if err := s.repos.Project.ArchiveCompletedProjects(ctx, tenantID, olderThan); err != nil {
		s.logger.Error("failed to archive completed projects", "tenant_id", tenantID, "error", err)
		return errors.NewServiceError("ARCHIVE_FAILED", "failed to archive projects", err)
	}

	s.logger.Info("archived completed projects", "tenant_id", tenantID, "older_than", olderThan)
	return nil
}

// HealthCheck checks service health
func (s *projectService) HealthCheck(ctx context.Context) error {
	return nil
}
