package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectTaskRepository defines the interface for project task repository operations
type ProjectTaskRepository interface {
	BaseRepository[models.ProjectTask]

	// Task Queries
	FindByProjectID(ctx context.Context, projectID uuid.UUID, pagination PaginationParams) ([]*models.ProjectTask, PaginationResult, error)
	FindByMilestoneID(ctx context.Context, milestoneID uuid.UUID) ([]*models.ProjectTask, error)
	FindByAssignedUser(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.ProjectTask, PaginationResult, error)
	FindByStatus(ctx context.Context, projectID uuid.UUID, status models.TaskStatus) ([]*models.ProjectTask, error)
	FindByPriority(ctx context.Context, projectID uuid.UUID, priority models.TaskPriority) ([]*models.ProjectTask, error)
	FindOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]*models.ProjectTask, error)
	FindBlockedTasks(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectTask, error)

	// Status Management
	UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status models.TaskStatus) error
	CompleteTask(ctx context.Context, taskID uuid.UUID) error
	BlockTask(ctx context.Context, taskID uuid.UUID, reason string) error
	UnblockTask(ctx context.Context, taskID uuid.UUID) error

	// Assignment Management
	AssignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	UnassignTask(ctx context.Context, taskID uuid.UUID) error

	// Time Tracking
	LogTimeEntry(ctx context.Context, taskID uuid.UUID, hours float64) error
	GetTaskTimeTracking(ctx context.Context, taskID uuid.UUID) (TimeTrackingSummary, error)
	UpdateTrackedHours(ctx context.Context, taskID uuid.UUID, hours float64) error
	UpdateActualCost(ctx context.Context, taskID uuid.UUID, cost float64) error

	// Dependencies
	AddDependency(ctx context.Context, taskID, dependsOnTaskID uuid.UUID) error
	RemoveDependency(ctx context.Context, taskID, dependsOnTaskID uuid.UUID) error
	GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*models.ProjectTask, error)
	GetBlockedBy(ctx context.Context, taskID uuid.UUID) ([]*models.ProjectTask, error)

	// Checklist Management
	UpdateChecklist(ctx context.Context, taskID uuid.UUID, checklist []models.ChecklistItem) error
	CompleteChecklistItem(ctx context.Context, taskID uuid.UUID, itemID string) error

	// Analytics
	GetTaskStats(ctx context.Context, projectID uuid.UUID) (TaskStats, error)
	GetUserTaskStats(ctx context.Context, userID uuid.UUID) (UserTaskStats, error)
	GetTaskVelocity(ctx context.Context, projectID uuid.UUID, days int) (float64, error)

	// Bulk Operations
	BulkUpdateStatus(ctx context.Context, taskIDs []uuid.UUID, status models.TaskStatus) error
	BulkAssign(ctx context.Context, taskIDs []uuid.UUID, userID uuid.UUID) error
	ReorderTasks(ctx context.Context, projectID uuid.UUID, taskOrders map[uuid.UUID]int) error
}

// TimeTrackingSummary represents time tracking summary for a task
type TimeTrackingSummary struct {
	TaskID          uuid.UUID `json:"task_id"`
	EstimatedHours  float64   `json:"estimated_hours"`
	TrackedHours    float64   `json:"tracked_hours"`
	RemainingHours  float64   `json:"remaining_hours"`
	PercentComplete float64   `json:"percent_complete"`
	IsOverEstimate  bool      `json:"is_over_estimate"`
}

// TaskStats represents task statistics for a project
type TaskStats struct {
	TotalTasks          int64                         `json:"total_tasks"`
	TodoTasks           int64                         `json:"todo_tasks"`
	InProgressTasks     int64                         `json:"in_progress_tasks"`
	BlockedTasks        int64                         `json:"blocked_tasks"`
	ReviewTasks         int64                         `json:"review_tasks"`
	CompletedTasks      int64                         `json:"completed_tasks"`
	ByStatus            map[models.TaskStatus]int64   `json:"by_status"`
	ByPriority          map[models.TaskPriority]int64 `json:"by_priority"`
	OverdueTasks        int64                         `json:"overdue_tasks"`
	UnassignedTasks     int64                         `json:"unassigned_tasks"`
	TotalEstimatedHours float64                       `json:"total_estimated_hours"`
	TotalTrackedHours   float64                       `json:"total_tracked_hours"`
	AvgCompletionTime   float64                       `json:"avg_completion_time_days"`
	CompletionRate      float64                       `json:"completion_rate"`
}

// UserTaskStats represents task statistics for a user
type UserTaskStats struct {
	TotalAssigned   int64   `json:"total_assigned"`
	Completed       int64   `json:"completed"`
	InProgress      int64   `json:"in_progress"`
	Overdue         int64   `json:"overdue"`
	CompletionRate  float64 `json:"completion_rate"`
	AvgHoursPerTask float64 `json:"avg_hours_per_task"`
}

// projectTaskRepository implements ProjectTaskRepository
type projectTaskRepository struct {
	BaseRepository[models.ProjectTask]
	db     *gorm.DB
	logger log.AllLogger
	cache  Cache
}

// NewProjectTaskRepository creates a new ProjectTaskRepository instance
func NewProjectTaskRepository(db *gorm.DB, config ...RepositoryConfig) ProjectTaskRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.ProjectTask](db, cfg)

	return &projectTaskRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
	}
}

// FindByProjectID retrieves all tasks for a project
func (r *projectTaskRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, pagination PaginationParams) ([]*models.ProjectTask, PaginationResult, error) {
	if projectID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ?", projectID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count tasks", err)
	}

	var tasks []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Preload("AssignedTo").
		Preload("Milestone").
		Where("project_id = ?", projectID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("order_index ASC, created_at ASC").
		Find(&tasks).Error; err != nil {
		r.logger.Error("failed to find tasks", "project_id", projectID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find tasks", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return tasks, paginationResult, nil
}

// FindByMilestoneID retrieves all tasks for a milestone
func (r *projectTaskRepository) FindByMilestoneID(ctx context.Context, milestoneID uuid.UUID) ([]*models.ProjectTask, error) {
	if milestoneID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	var tasks []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Preload("AssignedTo").
		Where("milestone_id = ?", milestoneID).
		Order("order_index ASC").
		Find(&tasks).Error; err != nil {
		r.logger.Error("failed to find tasks by milestone", "milestone_id", milestoneID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find tasks", err)
	}

	return tasks, nil
}

// FindByAssignedUser retrieves all tasks assigned to a user
func (r *projectTaskRepository) FindByAssignedUser(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.ProjectTask, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("assigned_to_id = ?", userID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count tasks", err)
	}

	var tasks []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Milestone").
		Where("assigned_to_id = ?", userID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("priority DESC, due_date ASC").
		Find(&tasks).Error; err != nil {
		r.logger.Error("failed to find tasks by assigned user", "user_id", userID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find tasks", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return tasks, paginationResult, nil
}

// FindByStatus retrieves tasks by status
func (r *projectTaskRepository) FindByStatus(ctx context.Context, projectID uuid.UUID, status models.TaskStatus) ([]*models.ProjectTask, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var tasks []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Preload("AssignedTo").
		Where("project_id = ? AND status = ?", projectID, status).
		Order("order_index ASC").
		Find(&tasks).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find tasks by status", err)
	}

	return tasks, nil
}

// FindByPriority retrieves tasks by priority
func (r *projectTaskRepository) FindByPriority(ctx context.Context, projectID uuid.UUID, priority models.TaskPriority) ([]*models.ProjectTask, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var tasks []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Preload("AssignedTo").
		Where("project_id = ? AND priority = ?", projectID, priority).
		Order("due_date ASC").
		Find(&tasks).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find tasks by priority", err)
	}

	return tasks, nil
}

// FindOverdueTasks retrieves all overdue tasks for a tenant
func (r *projectTaskRepository) FindOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]*models.ProjectTask, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var tasks []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("AssignedTo").
		Where("tenant_id = ? AND status NOT IN (?) AND due_date < ? AND due_date IS NOT NULL",
			tenantID, models.TaskStatusDone, time.Now()).
		Order("due_date ASC").
		Find(&tasks).Error; err != nil {
		r.logger.Error("failed to find overdue tasks", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find overdue tasks", err)
	}

	return tasks, nil
}

// FindBlockedTasks retrieves all blocked tasks for a project
func (r *projectTaskRepository) FindBlockedTasks(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectTask, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var tasks []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Preload("AssignedTo").
		Where("project_id = ? AND status = ?", projectID, models.TaskStatusBlocked).
		Find(&tasks).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find blocked tasks", err)
	}

	return tasks, nil
}

// UpdateTaskStatus updates the status of a task
func (r *projectTaskRepository) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status models.TaskStatus) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.TaskStatusDone {
		updates["completed_at"] = time.Now()
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ?", taskID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to update task status", "task_id", taskID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update task status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Update project progress
	var task models.ProjectTask
	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err == nil {
		// This would call the project repository's UpdateProgress method
		// For now, we'll just invalidate cache
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	return nil
}

// CompleteTask marks a task as completed
func (r *projectTaskRepository) CompleteTask(ctx context.Context, taskID uuid.UUID) error {
	return r.UpdateTaskStatus(ctx, taskID, models.TaskStatusDone)
}

// BlockTask marks a task as blocked with a reason
func (r *projectTaskRepository) BlockTask(ctx context.Context, taskID uuid.UUID, reason string) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status":       models.TaskStatusBlocked,
		"block_reason": reason,
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ?", taskID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to block task", "task_id", taskID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to block task", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// UnblockTask removes the blocked status from a task
func (r *projectTaskRepository) UnblockTask(ctx context.Context, taskID uuid.UUID) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status":       models.TaskStatusTodo,
		"block_reason": "",
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ? AND status = ?", taskID, models.TaskStatusBlocked).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to unblock task", "task_id", taskID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to unblock task", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found or not blocked", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// AssignTask assigns a task to a user
func (r *projectTaskRepository) AssignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}
	if userID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ?", taskID).
		Update("assigned_to_id", userID)

	if result.Error != nil {
		r.logger.Error("failed to assign task", "task_id", taskID, "user_id", userID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to assign task", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// UnassignTask removes the assignment from a task
func (r *projectTaskRepository) UnassignTask(ctx context.Context, taskID uuid.UUID) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ?", taskID).
		Update("assigned_to_id", nil)

	if result.Error != nil {
		r.logger.Error("failed to unassign task", "task_id", taskID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to unassign task", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// UpdateTrackedHours updates the tracked hours for a task
func (r *projectTaskRepository) UpdateTrackedHours(ctx context.Context, taskID uuid.UUID, hours float64) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ?", taskID).
		Update("tracked_hours", hours)

	if result.Error != nil {
		r.logger.Error("failed to update tracked hours", "task_id", taskID, "hours", hours, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update tracked hours", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// UpdateActualCost updates the actual cost for a task
func (r *projectTaskRepository) UpdateActualCost(ctx context.Context, taskID uuid.UUID, cost float64) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ?", taskID).
		Update("actual_cost", cost)

	if result.Error != nil {
		r.logger.Error("failed to update actual cost", "task_id", taskID, "cost", cost, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update actual cost", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// LogTimeEntry adds tracked hours to a task
func (r *projectTaskRepository) LogTimeEntry(ctx context.Context, taskID uuid.UUID, hours float64) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}
	if hours <= 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "hours must be positive", errors.ErrInvalidInput)
	}

	// Increment tracked hours
	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ?", taskID).
		Update("tracked_hours", gorm.Expr("tracked_hours + ?", hours))

	if result.Error != nil {
		r.logger.Error("failed to log time entry", "task_id", taskID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to log time entry", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// GetTaskTimeTracking retrieves time tracking summary for a task
func (r *projectTaskRepository) GetTaskTimeTracking(ctx context.Context, taskID uuid.UUID) (TimeTrackingSummary, error) {
	if taskID == uuid.Nil {
		return TimeTrackingSummary{}, errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	var task models.ProjectTask
	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return TimeTrackingSummary{}, errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
		}
		return TimeTrackingSummary{}, errors.NewRepositoryError("FIND_FAILED", "failed to find task", err)
	}

	summary := TimeTrackingSummary{
		TaskID:         taskID,
		EstimatedHours: task.EstimatedHours,
		TrackedHours:   task.TrackedHours,
	}

	if task.EstimatedHours > 0 {
		summary.RemainingHours = task.EstimatedHours - task.TrackedHours
		summary.PercentComplete = (task.TrackedHours / task.EstimatedHours) * 100
		summary.IsOverEstimate = task.TrackedHours > task.EstimatedHours
	}

	return summary, nil
}

// AddDependency adds a dependency relationship between tasks
func (r *projectTaskRepository) AddDependency(ctx context.Context, taskID, dependsOnTaskID uuid.UUID) error {
	if taskID == uuid.Nil || dependsOnTaskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task IDs cannot be nil", errors.ErrInvalidInput)
	}

	// Get current task
	var task models.ProjectTask
	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Add to depends_on array if not already present
	dependsOn := task.DependsOn
	for _, id := range dependsOn {
		if id == dependsOnTaskID {
			return nil // Already exists
		}
	}
	dependsOn = append(dependsOn, dependsOnTaskID)

	if err := r.db.WithContext(ctx).
		Model(&task).
		Update("depends_on", dependsOn).Error; err != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to add dependency", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// RemoveDependency removes a dependency relationship
func (r *projectTaskRepository) RemoveDependency(ctx context.Context, taskID, dependsOnTaskID uuid.UUID) error {
	if taskID == uuid.Nil || dependsOnTaskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task IDs cannot be nil", errors.ErrInvalidInput)
	}

	// Get current task
	var task models.ProjectTask
	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Remove from depends_on array
	var newDependsOn []uuid.UUID
	for _, id := range task.DependsOn {
		if id != dependsOnTaskID {
			newDependsOn = append(newDependsOn, id)
		}
	}

	if err := r.db.WithContext(ctx).
		Model(&task).
		Update("depends_on", newDependsOn).Error; err != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to remove dependency", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// GetDependencies retrieves all tasks that a task depends on
func (r *projectTaskRepository) GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*models.ProjectTask, error) {
	if taskID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	var task models.ProjectTask
	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return nil, errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	if len(task.DependsOn) == 0 {
		return []*models.ProjectTask{}, nil
	}

	var dependencies []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Where("id IN ?", task.DependsOn).
		Find(&dependencies).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find dependencies", err)
	}

	return dependencies, nil
}

// GetBlockedBy retrieves all tasks that block a task
func (r *projectTaskRepository) GetBlockedBy(ctx context.Context, taskID uuid.UUID) ([]*models.ProjectTask, error) {
	if taskID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	var task models.ProjectTask
	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return nil, errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	if len(task.BlockedBy) == 0 {
		return []*models.ProjectTask{}, nil
	}

	var blockers []*models.ProjectTask
	if err := r.db.WithContext(ctx).
		Where("id IN ?", task.BlockedBy).
		Find(&blockers).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find blocking tasks", err)
	}

	return blockers, nil
}

// UpdateChecklist updates the entire checklist for a task
func (r *projectTaskRepository) UpdateChecklist(ctx context.Context, taskID uuid.UUID, checklist []models.ChecklistItem) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id = ?", taskID).
		Update("checklist", checklist)

	if result.Error != nil {
		r.logger.Error("failed to update checklist", "task_id", taskID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update checklist", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// CompleteChecklistItem marks a specific checklist item as completed
func (r *projectTaskRepository) CompleteChecklistItem(ctx context.Context, taskID uuid.UUID, itemID string) error {
	if taskID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "task_id cannot be nil", errors.ErrInvalidInput)
	}

	var task models.ProjectTask
	if err := r.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return errors.NewRepositoryError("NOT_FOUND", "task not found", errors.ErrNotFound)
	}

	// Find and update the checklist item
	found := false
	for i := range task.Checklist {
		if task.Checklist[i].ID == itemID {
			task.Checklist[i].IsCompleted = true
			found = true
			break
		}
	}

	if !found {
		return errors.NewRepositoryError("NOT_FOUND", "checklist item not found", errors.ErrNotFound)
	}

	if err := r.db.WithContext(ctx).
		Model(&task).
		Update("checklist", task.Checklist).Error; err != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to complete checklist item", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// GetTaskStats retrieves statistics for tasks in a project
func (r *projectTaskRepository) GetTaskStats(ctx context.Context, projectID uuid.UUID) (TaskStats, error) {
	if projectID == uuid.Nil {
		return TaskStats{}, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := TaskStats{
		ByStatus:   make(map[models.TaskStatus]int64),
		ByPriority: make(map[models.TaskPriority]int64),
	}

	// Total tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ?", projectID).
		Count(&stats.TotalTasks)

	// Todo tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status = ?", projectID, models.TaskStatusTodo).
		Count(&stats.TodoTasks)

	// In-progress tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status = ?", projectID, models.TaskStatusInProgress).
		Count(&stats.InProgressTasks)

	// Review tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status = ?", projectID, models.TaskStatusReview).
		Count(&stats.ReviewTasks)

	// Completed tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status = ?", projectID, models.TaskStatusDone).
		Count(&stats.CompletedTasks)

	// By status
	type StatusCount struct {
		Status models.TaskStatus
		Count  int64
	}
	var statusCounts []StatusCount
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Select("status, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// By priority
	type PriorityCount struct {
		Priority models.TaskPriority
		Count    int64
	}
	var priorityCounts []PriorityCount
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Select("priority, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("priority").
		Scan(&priorityCounts)

	for _, pc := range priorityCounts {
		stats.ByPriority[pc.Priority] = pc.Count
	}

	// Overdue tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status != ? AND due_date < ? AND due_date IS NOT NULL",
			projectID, models.TaskStatusDone, time.Now()).
		Count(&stats.OverdueTasks)

	// Blocked tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status = ?", projectID, models.TaskStatusBlocked).
		Count(&stats.BlockedTasks)

	// Unassigned tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND assigned_to_id IS NULL", projectID).
		Count(&stats.UnassignedTasks)

	// Total estimated hours
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(SUM(estimated_hours), 0)").
		Scan(&stats.TotalEstimatedHours)

	// Total tracked hours
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(SUM(tracked_hours), 0)").
		Scan(&stats.TotalTrackedHours)

	// Average completion time
	var avgDays float64
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Select("AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400)").
		Where("project_id = ? AND status = ? AND completed_at IS NOT NULL", projectID, models.TaskStatusDone).
		Scan(&avgDays)
	stats.AvgCompletionTime = avgDays

	// Completion rate
	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}

// GetUserTaskStats retrieves task statistics for a user
func (r *projectTaskRepository) GetUserTaskStats(ctx context.Context, userID uuid.UUID) (UserTaskStats, error) {
	if userID == uuid.Nil {
		return UserTaskStats{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := UserTaskStats{}

	// Total assigned
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("assigned_to_id = ?", userID).
		Count(&stats.TotalAssigned)

	// Completed
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("assigned_to_id = ? AND status = ?", userID, models.TaskStatusDone).
		Count(&stats.Completed)

	// In progress
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("assigned_to_id = ? AND status = ?", userID, models.TaskStatusInProgress).
		Count(&stats.InProgress)

	// Overdue
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("assigned_to_id = ? AND status != ? AND due_date < ? AND due_date IS NOT NULL",
			userID, models.TaskStatusDone, time.Now()).
		Count(&stats.Overdue)

	// Completion rate
	if stats.TotalAssigned > 0 {
		stats.CompletionRate = float64(stats.Completed) / float64(stats.TotalAssigned) * 100
	}

	// Average hours per task
	var totalHours float64
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Select("SUM(tracked_hours)").
		Where("assigned_to_id = ?", userID).
		Scan(&totalHours)

	if stats.Completed > 0 {
		stats.AvgHoursPerTask = totalHours / float64(stats.Completed)
	}

	return stats, nil
}

// GetTaskVelocity calculates task completion velocity over a period
func (r *projectTaskRepository) GetTaskVelocity(ctx context.Context, projectID uuid.UUID, days int) (float64, error) {
	if projectID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}
	if days <= 0 {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "days must be positive", errors.ErrInvalidInput)
	}

	startDate := time.Now().AddDate(0, 0, -days)

	var completedCount int64
	if err := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status = ? AND completed_at >= ?",
			projectID, models.TaskStatusDone, startDate).
		Count(&completedCount).Error; err != nil {
		return 0, errors.NewRepositoryError("QUERY_FAILED", "failed to calculate velocity", err)
	}

	velocity := float64(completedCount) / float64(days)
	return velocity, nil
}

// BulkUpdateStatus updates the status of multiple tasks
func (r *projectTaskRepository) BulkUpdateStatus(ctx context.Context, taskIDs []uuid.UUID, status models.TaskStatus) error {
	if len(taskIDs) == 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "task IDs cannot be empty", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.TaskStatusDone {
		updates["completed_at"] = time.Now()
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id IN (?)", taskIDs).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to bulk update task status", "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update task status", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	return nil
}

// BulkAssign assigns multiple tasks to a user
func (r *projectTaskRepository) BulkAssign(ctx context.Context, taskIDs []uuid.UUID, userID uuid.UUID) error {
	if len(taskIDs) == 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "task IDs cannot be empty", errors.ErrInvalidInput)
	}
	if userID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("id IN (?)", taskIDs).
		Update("assigned_to_id", userID)

	if result.Error != nil {
		r.logger.Error("failed to bulk assign tasks", "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk assign tasks", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}

// ReorderTasks updates the order of tasks
func (r *projectTaskRepository) ReorderTasks(ctx context.Context, projectID uuid.UUID, taskOrders map[uuid.UUID]int) error {
	if projectID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}
	if len(taskOrders) == 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "task orders cannot be empty", errors.ErrInvalidInput)
	}

	// Use transaction to ensure atomicity
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for taskID, orderIndex := range taskOrders {
		if err := tx.Model(&models.ProjectTask{}).
			Where("id = ? AND project_id = ?", taskID, projectID).
			Update("order_index", orderIndex).Error; err != nil {
			tx.Rollback()
			return errors.NewRepositoryError("UPDATE_FAILED", "failed to reorder tasks", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return errors.NewRepositoryError("COMMIT_FAILED", "failed to commit transaction", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_tasks:*")
	}

	return nil
}
