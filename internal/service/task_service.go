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

// TaskService defines the interface for task operations
type TaskService interface {
	// CRUD Operations
	CreateTask(ctx context.Context, req *dto.CreateTaskRequest) (*dto.TaskResponse, error)
	GetTask(ctx context.Context, id uuid.UUID) (*dto.TaskResponse, error)
	UpdateTask(ctx context.Context, id uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error

	// Query Operations
	ListTasksByProject(ctx context.Context, projectID uuid.UUID, page, pageSize int) (*dto.TaskListResponse, error)
	ListTasksByMilestone(ctx context.Context, milestoneID uuid.UUID) ([]*dto.TaskResponse, error)
	ListTasksByAssignee(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.TaskListResponse, error)
	ListTasksByStatus(ctx context.Context, projectID uuid.UUID, status models.TaskStatus) ([]*dto.TaskResponse, error)
	ListOverdueTasks(ctx context.Context, projectID uuid.UUID) ([]*dto.TaskResponse, error)

	// Status Operations
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, status models.TaskStatus) error
	CompleteTask(ctx context.Context, id uuid.UUID) error
	ReopenTask(ctx context.Context, id uuid.UUID) error

	// Assignment Operations
	AssignTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	UnassignTask(ctx context.Context, id uuid.UUID) error

	// Time & Cost Tracking
	UpdateTrackedHours(ctx context.Context, id uuid.UUID, hours float64) error
	UpdateActualCost(ctx context.Context, id uuid.UUID, cost float64) error

	// Order Operations
	ReorderTasks(ctx context.Context, req *dto.ReorderTasksRequest) error

	// Statistics
	GetTaskStats(ctx context.Context, projectID uuid.UUID) (*dto.TaskStatsResponse, error)
	GetUserTaskStats(ctx context.Context, userID uuid.UUID) (*dto.TaskStatsResponse, error)
}

// taskService implements TaskService
type taskService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewTaskService creates a new task service
func NewTaskService(repos *repository.Repositories, logger log.AllLogger) TaskService {
	return &taskService{
		repos:  repos,
		logger: logger,
	}
}

// CreateTask creates a new task
func (s *taskService) CreateTask(ctx context.Context, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	// Verify project exists
	project, err := s.repos.Project.GetByID(ctx, req.ProjectID)
	if err != nil {
		s.logger.Error("failed to find project", "project_id", req.ProjectID, "error", err)
		return nil, errors.NewNotFoundError("project not found")
	}

	// Verify milestone if provided
	if req.MilestoneID != nil {
		milestone, err := s.repos.ProjectMilestone.GetByID(ctx, *req.MilestoneID)
		if err != nil {
			s.logger.Error("failed to find milestone", "milestone_id", *req.MilestoneID, "error", err)
			return nil, errors.NewNotFoundError("milestone not found")
		}
		if milestone.ProjectID != req.ProjectID {
			return nil, errors.NewValidationError("milestone does not belong to project")
		}
	}

	// Verify assignee if provided
	if req.AssignedToID != nil {
		_, err := s.repos.User.GetByID(ctx, *req.AssignedToID)
		if err != nil {
			s.logger.Error("failed to find user", "user_id", *req.AssignedToID, "error", err)
			return nil, errors.NewNotFoundError("user not found")
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

	task := &models.ProjectTask{
		TenantID:       project.TenantID,
		ProjectID:      req.ProjectID,
		MilestoneID:    req.MilestoneID,
		AssignedToID:   req.AssignedToID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         models.TaskStatusTodo,
		Priority:       req.Priority,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		EstimatedHours: req.EstimatedHours,
		EstimatedCost:  req.EstimatedCost,
		DependsOn:      req.DependsOn,
		Checklist:      req.Checklist,
		Metadata:       metadata,
	}

	if err := s.repos.ProjectTask.Create(ctx, task); err != nil {
		s.logger.Error("failed to create task", "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "failed to create task", err)
	}

	s.logger.Info("task created", "task_id", task.ID, "project_id", req.ProjectID)

	// Reload with relationships
	created, err := s.repos.ProjectTask.GetByID(ctx, task.ID)
	if err != nil {
		return dto.ToTaskResponse(task), nil
	}

	return dto.ToTaskResponse(created), nil
}

// GetTask retrieves a task by ID
func (s *taskService) GetTask(ctx context.Context, id uuid.UUID) (*dto.TaskResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("task_id is required")
	}

	task, err := s.repos.ProjectTask.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get task", "task_id", id, "error", err)
		return nil, errors.NewNotFoundError("task not found")
	}

	return dto.ToTaskResponse(task), nil
}

// UpdateTask updates a task
func (s *taskService) UpdateTask(ctx context.Context, id uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("task_id is required")
	}

	// Get existing task
	existing, err := s.repos.ProjectTask.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find task", "task_id", id, "error", err)
		return nil, errors.NewNotFoundError("task not found")
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
	if req.Priority != nil {
		existing.Priority = *req.Priority
	}
	if req.AssignedToID != nil {
		existing.AssignedToID = req.AssignedToID
	}
	if req.MilestoneID != nil {
		existing.MilestoneID = req.MilestoneID
	}
	if req.StartDate != nil {
		existing.StartDate = req.StartDate
	}
	if req.DueDate != nil {
		existing.DueDate = req.DueDate
	}
	if req.EstimatedHours != nil {
		existing.EstimatedHours = *req.EstimatedHours
	}
	if req.EstimatedCost != nil {
		existing.EstimatedCost = *req.EstimatedCost
	}
	if req.TrackedHours != nil {
		existing.TrackedHours = *req.TrackedHours
	}
	if req.ActualCost != nil {
		existing.ActualCost = *req.ActualCost
	}
	if req.DependsOn != nil {
		existing.DependsOn = req.DependsOn
	}
	if req.BlockedBy != nil {
		existing.BlockedBy = req.BlockedBy
	}
	if req.BlockReason != nil {
		existing.BlockReason = *req.BlockReason
	}
	if req.Checklist != nil {
		existing.Checklist = req.Checklist
	}
	if req.AttachmentURLs != nil {
		existing.AttachmentURLs = req.AttachmentURLs
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

	if err := s.repos.ProjectTask.Update(ctx, existing); err != nil {
		s.logger.Error("failed to update task", "task_id", id, "error", err)
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to update task", err)
	}

	s.logger.Info("task updated", "task_id", id)

	// Get updated task
	updated, err := s.repos.ProjectTask.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "failed to retrieve updated task", err)
	}

	return dto.ToTaskResponse(updated), nil
}

// DeleteTask deletes a task
func (s *taskService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewValidationError("task_id is required")
	}

	if err := s.repos.ProjectTask.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete task", "task_id", id, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "failed to delete task", err)
	}

	s.logger.Info("task deleted", "task_id", id)
	return nil
}

// ListTasksByProject retrieves all tasks for a project
func (s *taskService) ListTasksByProject(ctx context.Context, projectID uuid.UUID, page, pageSize int) (*dto.TaskListResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	tasks, paginationResult, err := s.repos.ProjectTask.FindByProjectID(ctx, projectID, pagination)
	if err != nil {
		s.logger.Error("failed to list tasks", "project_id", projectID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list tasks", err)
	}

	return &dto.TaskListResponse{
		Tasks:       dto.ToTaskResponses(tasks),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListTasksByMilestone retrieves all tasks for a milestone
func (s *taskService) ListTasksByMilestone(ctx context.Context, milestoneID uuid.UUID) ([]*dto.TaskResponse, error) {
	if milestoneID == uuid.Nil {
		return nil, errors.NewValidationError("milestone_id is required")
	}

	tasks, err := s.repos.ProjectTask.FindByMilestoneID(ctx, milestoneID)
	if err != nil {
		s.logger.Error("failed to list tasks by milestone", "milestone_id", milestoneID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list tasks", err)
	}

	return dto.ToTaskResponses(tasks), nil
}

// ListTasksByAssignee retrieves all tasks assigned to a user
func (s *taskService) ListTasksByAssignee(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.TaskListResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id is required")
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	tasks, paginationResult, err := s.repos.ProjectTask.FindByAssignedUser(ctx, userID, pagination)
	if err != nil {
		s.logger.Error("failed to list tasks by assignee", "user_id", userID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list tasks", err)
	}

	return &dto.TaskListResponse{
		Tasks:       dto.ToTaskResponses(tasks),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListTasksByStatus retrieves tasks by status
func (s *taskService) ListTasksByStatus(ctx context.Context, projectID uuid.UUID, status models.TaskStatus) ([]*dto.TaskResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	tasks, err := s.repos.ProjectTask.FindByStatus(ctx, projectID, status)
	if err != nil {
		s.logger.Error("failed to list tasks by status", "project_id", projectID, "status", status, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list tasks", err)
	}

	return dto.ToTaskResponses(tasks), nil
}

// ListOverdueTasks retrieves overdue tasks
func (s *taskService) ListOverdueTasks(ctx context.Context, projectID uuid.UUID) ([]*dto.TaskResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	tasks, err := s.repos.ProjectTask.FindOverdueTasks(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to list overdue tasks", "project_id", projectID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "failed to list overdue tasks", err)
	}

	return dto.ToTaskResponses(tasks), nil
}

// UpdateTaskStatus updates task status
func (s *taskService) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status models.TaskStatus) error {
	if id == uuid.Nil {
		return errors.NewValidationError("task_id is required")
	}

	if err := s.repos.ProjectTask.UpdateTaskStatus(ctx, id, status); err != nil {
		s.logger.Error("failed to update task status", "task_id", id, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "failed to update status", err)
	}

	s.logger.Info("task status updated", "task_id", id, "status", status)
	return nil
}

// CompleteTask marks a task as completed
func (s *taskService) CompleteTask(ctx context.Context, id uuid.UUID) error {
	return s.UpdateTaskStatus(ctx, id, models.TaskStatusDone)
}

// ReopenTask reopens a completed task
func (s *taskService) ReopenTask(ctx context.Context, id uuid.UUID) error {
	return s.UpdateTaskStatus(ctx, id, models.TaskStatusTodo)
}

// AssignTask assigns a task to a user
func (s *taskService) AssignTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if id == uuid.Nil || userID == uuid.Nil {
		return errors.NewValidationError("task_id and user_id are required")
	}

	if err := s.repos.ProjectTask.AssignTask(ctx, id, userID); err != nil {
		s.logger.Error("failed to assign task", "task_id", id, "user_id", userID, "error", err)
		return errors.NewServiceError("ASSIGN_FAILED", "failed to assign task", err)
	}

	s.logger.Info("task assigned", "task_id", id, "user_id", userID)
	return nil
}

// UnassignTask removes task assignment
func (s *taskService) UnassignTask(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewValidationError("task_id is required")
	}

	if err := s.repos.ProjectTask.UnassignTask(ctx, id); err != nil {
		s.logger.Error("failed to unassign task", "task_id", id, "error", err)
		return errors.NewServiceError("UNASSIGN_FAILED", "failed to unassign task", err)
	}

	s.logger.Info("task unassigned", "task_id", id)
	return nil
}

// UpdateTrackedHours updates tracked hours for a task
func (s *taskService) UpdateTrackedHours(ctx context.Context, id uuid.UUID, hours float64) error {
	if id == uuid.Nil {
		return errors.NewValidationError("task_id is required")
	}

	if hours < 0 {
		return errors.NewValidationError("hours cannot be negative")
	}

	if err := s.repos.ProjectTask.UpdateTrackedHours(ctx, id, hours); err != nil {
		s.logger.Error("failed to update tracked hours", "task_id", id, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "failed to update tracked hours", err)
	}

	s.logger.Info("tracked hours updated", "task_id", id, "hours", hours)
	return nil
}

// UpdateActualCost updates actual cost for a task
func (s *taskService) UpdateActualCost(ctx context.Context, id uuid.UUID, cost float64) error {
	if id == uuid.Nil {
		return errors.NewValidationError("task_id is required")
	}

	if cost < 0 {
		return errors.NewValidationError("cost cannot be negative")
	}

	if err := s.repos.ProjectTask.UpdateActualCost(ctx, id, cost); err != nil {
		s.logger.Error("failed to update actual cost", "task_id", id, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "failed to update actual cost", err)
	}

	s.logger.Info("actual cost updated", "task_id", id, "cost", cost)
	return nil
}

// ReorderTasks reorders tasks
func (s *taskService) ReorderTasks(ctx context.Context, req *dto.ReorderTasksRequest) error {
	if req.ProjectID == uuid.Nil {
		return errors.NewValidationError("project_id is required")
	}

	// Convert string keys to UUIDs
	taskOrders := make(map[uuid.UUID]int)
	for taskIDStr, orderIndex := range req.TaskOrders {
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return errors.NewValidationError("invalid task_id: " + taskIDStr)
		}
		taskOrders[taskID] = orderIndex
	}

	if err := s.repos.ProjectTask.ReorderTasks(ctx, req.ProjectID, taskOrders); err != nil {
		s.logger.Error("failed to reorder tasks", "project_id", req.ProjectID, "error", err)
		return errors.NewServiceError("REORDER_FAILED", "failed to reorder tasks", err)
	}

	s.logger.Info("tasks reordered", "project_id", req.ProjectID, "count", len(taskOrders))
	return nil
}

// GetTaskStats retrieves task statistics for a project
func (s *taskService) GetTaskStats(ctx context.Context, projectID uuid.UUID) (*dto.TaskStatsResponse, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewValidationError("project_id is required")
	}

	stats, err := s.repos.ProjectTask.GetTaskStats(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to get task stats", "project_id", projectID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "failed to get task stats", err)
	}

	return &dto.TaskStatsResponse{
		ProjectID:           projectID,
		TotalTasks:          stats.TotalTasks,
		TodoTasks:           stats.TodoTasks,
		InProgressTasks:     stats.InProgressTasks,
		BlockedTasks:        stats.BlockedTasks,
		ReviewTasks:         stats.ReviewTasks,
		CompletedTasks:      stats.CompletedTasks,
		OverdueTasks:        stats.OverdueTasks,
		TotalEstimatedHours: stats.TotalEstimatedHours,
		TotalTrackedHours:   stats.TotalTrackedHours,
		CompletionRate:      stats.CompletionRate,
	}, nil
}

// GetUserTaskStats retrieves task statistics for a user
func (s *taskService) GetUserTaskStats(ctx context.Context, userID uuid.UUID) (*dto.TaskStatsResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id is required")
	}

	stats, err := s.repos.ProjectTask.GetUserTaskStats(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user task stats", "user_id", userID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "failed to get user task stats", err)
	}

	return &dto.TaskStatsResponse{
		UserID:            userID,
		TotalTasks:        stats.TotalAssigned,
		InProgressTasks:   stats.InProgress,
		CompletedTasks:    stats.Completed,
		OverdueTasks:      stats.Overdue,
		CompletionRate:    stats.CompletionRate,
		TotalTrackedHours: stats.AvgHoursPerTask * float64(stats.TotalAssigned),
	}, nil
}
