package repository

import (
	"context"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectRepository defines the interface for project repository operations
type ProjectRepository interface {
	BaseRepository[models.Project]

	// Project Queries
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Project, PaginationResult, error)
	FindByArtisanID(ctx context.Context, artisanID uuid.UUID, pagination PaginationParams) ([]*models.Project, PaginationResult, error)
	FindByCustomerID(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Project, PaginationResult, error)
	FindByStatus(ctx context.Context, tenantID uuid.UUID, status models.ProjectStatus, pagination PaginationParams) ([]*models.Project, PaginationResult, error)
	FindByPriority(ctx context.Context, tenantID uuid.UUID, priority models.ProjectPriority, pagination PaginationParams) ([]*models.Project, PaginationResult, error)
	FindOverdueProjects(ctx context.Context, tenantID uuid.UUID) ([]*models.Project, error)
	FindActiveProjects(ctx context.Context, tenantID uuid.UUID) ([]*models.Project, error)

	// Search & Filter
	Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.Project, PaginationResult, error)
	FindByTags(ctx context.Context, tenantID uuid.UUID, tags []string, pagination PaginationParams) ([]*models.Project, PaginationResult, error)
	FindByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Project, PaginationResult, error)

	// Progress Management
	UpdateProgress(ctx context.Context, projectID uuid.UUID) error
	RecalculateAllProgress(ctx context.Context, tenantID uuid.UUID) error
	GetProjectWithFullDetails(ctx context.Context, projectID uuid.UUID) (*models.Project, error)

	// Status Transitions
	StartProject(ctx context.Context, projectID uuid.UUID) error
	PauseProject(ctx context.Context, projectID uuid.UUID) error
	CompleteProject(ctx context.Context, projectID uuid.UUID) error
	CancelProject(ctx context.Context, projectID uuid.UUID, reason string) error
	ResumeProject(ctx context.Context, projectID uuid.UUID) error

	// Analytics & Statistics
	GetProjectStats(ctx context.Context, tenantID uuid.UUID) (ProjectStats, error)
	GetArtisanProjectStats(ctx context.Context, artisanID uuid.UUID) (ArtisanProjectStats, error)
	GetCustomerProjectStats(ctx context.Context, customerID uuid.UUID) (CustomerProjectStats, error)
	GetProjectTimeline(ctx context.Context, projectID uuid.UUID) ([]TimelineEvent, error)
	GetProjectHealth(ctx context.Context, projectID uuid.UUID) (ProjectHealth, error)

	// Bulk Operations
	BulkUpdateStatus(ctx context.Context, projectIDs []uuid.UUID, status models.ProjectStatus) error
	BulkAssignArtisan(ctx context.Context, projectIDs []uuid.UUID, artisanID uuid.UUID) error
	ArchiveCompletedProjects(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error

	// Dashboard Queries
	GetArtisanDashboard(ctx context.Context, artisanID uuid.UUID) (ArtisanDashboard, error)
	GetTenantDashboard(ctx context.Context, tenantID uuid.UUID) (TenantProjectDashboard, error)
}

// ProjectStats represents project statistics
type ProjectStats struct {
	TotalProjects          int64                            `json:"total_projects"`
	ActiveProjects         int64                            `json:"active_projects"`
	CompletedProjects      int64                            `json:"completed_projects"`
	OnHoldProjects         int64                            `json:"on_hold_projects"`
	CancelledProjects      int64                            `json:"cancelled_projects"`
	OverdueProjects        int64                            `json:"overdue_projects"`
	ByStatus               map[models.ProjectStatus]int64   `json:"by_status"`
	ByPriority             map[models.ProjectPriority]int64 `json:"by_priority"`
	AverageProgress        float64                          `json:"average_progress"`
	TotalBudget            float64                          `json:"total_budget"`
	OnTimeProjects         int64                            `json:"on_time_projects"`
	CompletionRate         float64                          `json:"completion_rate"`
	AverageTasksPerProject float64                          `json:"average_tasks_per_project"`
}

// ArtisanProjectStats represents artisan-specific project statistics
type ArtisanProjectStats struct {
	TotalProjects          int64   `json:"total_projects"`
	ActiveProjects         int64   `json:"active_projects"`
	CompletedProjects      int64   `json:"completed_projects"`
	AverageProgress        float64 `json:"average_progress"`
	TotalRevenue           float64 `json:"total_revenue"`
	OnTimeDeliveryRate     float64 `json:"on_time_delivery_rate"`
	CustomerSatisfaction   float64 `json:"customer_satisfaction"`
	AverageProjectDuration float64 `json:"average_project_duration_days"`
}

// CustomerProjectStats represents customer-specific project statistics
type CustomerProjectStats struct {
	TotalProjects      int64   `json:"total_projects"`
	ActiveProjects     int64   `json:"active_projects"`
	CompletedProjects  int64   `json:"completed_projects"`
	TotalSpent         float64 `json:"total_spent"`
	AverageProjectCost float64 `json:"average_project_cost"`
}

// TimelineEvent represents a project timeline event
type TimelineEvent struct {
	Date        time.Time  `json:"date"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	UserName    string     `json:"user_name,omitempty"`
}

// ProjectHealth represents project health metrics
type ProjectHealth struct {
	ProjectID          uuid.UUID `json:"project_id"`
	HealthScore        int       `json:"health_score"` // 0-100
	IsOnTrack          bool      `json:"is_on_track"`
	IsOverBudget       bool      `json:"is_over_budget"`
	IsOverdue          bool      `json:"is_overdue"`
	BlockedTasksCount  int       `json:"blocked_tasks_count"`
	OverdueTasksCount  int       `json:"overdue_tasks_count"`
	CompletionVelocity float64   `json:"completion_velocity"` // tasks per day
	RiskLevel          string    `json:"risk_level"`          // low, medium, high
	Recommendations    []string  `json:"recommendations"`
}

// ArtisanDashboard represents artisan dashboard data
type ArtisanDashboard struct {
	ActiveProjects    []*models.Project   `json:"active_projects"`
	UpcomingDeadlines []*models.Project   `json:"upcoming_deadlines"`
	OverdueProjects   []*models.Project   `json:"overdue_projects"`
	RecentlyCompleted []*models.Project   `json:"recently_completed"`
	Statistics        ArtisanProjectStats `json:"statistics"`
	TasksSummary      TasksSummary        `json:"tasks_summary"`
}

// TenantProjectDashboard represents tenant dashboard data
type TenantProjectDashboard struct {
	Statistics           ProjectStats         `json:"statistics"`
	ActiveProjects       []*models.Project    `json:"active_projects"`
	HighPriorityProjects []*models.Project    `json:"high_priority_projects"`
	OverdueProjects      []*models.Project    `json:"overdue_projects"`
	RecentActivity       []TimelineEvent      `json:"recent_activity"`
	TopArtisans          []ArtisanPerformance `json:"top_artisans"`
}

// TasksSummary represents task summary
type TasksSummary struct {
	TotalTasks      int64 `json:"total_tasks"`
	CompletedTasks  int64 `json:"completed_tasks"`
	OverdueTasks    int64 `json:"overdue_tasks"`
	BlockedTasks    int64 `json:"blocked_tasks"`
	InProgressTasks int64 `json:"in_progress_tasks"`
}

// ArtisanPerformance represents artisan performance metrics
type ArtisanPerformance struct {
	ArtisanID         uuid.UUID `json:"artisan_id"`
	ArtisanName       string    `json:"artisan_name"`
	ActiveProjects    int       `json:"active_projects"`
	CompletedProjects int       `json:"completed_projects"`
	AverageProgress   float64   `json:"average_progress"`
	OnTimeRate        float64   `json:"on_time_rate"`
}

// projectRepository implements ProjectRepository
type projectRepository struct {
	BaseRepository[models.Project]
	db     *gorm.DB
	logger log.AllLogger
	cache  Cache
}

// NewProjectRepository creates a new ProjectRepository instance
func NewProjectRepository(db *gorm.DB, config ...RepositoryConfig) ProjectRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.Project](db, cfg)

	return &projectRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
	}
}

// FindByTenantID retrieves all projects for a tenant
func (r *projectRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Project, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count projects", err)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Preload("Tasks").
		Preload("Milestones").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		r.logger.Error("failed to find projects", "tenant_id", tenantID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find projects", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return projects, paginationResult, nil
}

// FindByArtisanID retrieves all projects for an artisan
func (r *projectRepository) FindByArtisanID(ctx context.Context, artisanID uuid.UUID, pagination PaginationParams) ([]*models.Project, PaginationResult, error) {
	if artisanID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("artisan_id = ?", artisanID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count projects", err)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Tasks", "assigned_to_id = ?", artisanID).
		Preload("Milestones").
		Where("artisan_id = ?", artisanID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		r.logger.Error("failed to find projects", "artisan_id", artisanID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find projects", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return projects, paginationResult, nil
}

// FindByCustomerID retrieves all projects for a customer
func (r *projectRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Project, PaginationResult, error) {
	if customerID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "customer_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("customer_id = ?", customerID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count projects", err)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Milestones").
		Preload("Updates", "visible_to_customer = ?", true).
		Where("customer_id = ?", customerID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		r.logger.Error("failed to find projects", "customer_id", customerID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find projects", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return projects, paginationResult, nil
}

// FindByStatus retrieves projects by status
func (r *projectRepository) FindByStatus(ctx context.Context, tenantID uuid.UUID, status models.ProjectStatus, pagination PaginationParams) ([]*models.Project, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count projects", err)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find projects", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return projects, paginationResult, nil
}

// FindByPriority retrieves projects by priority
func (r *projectRepository) FindByPriority(ctx context.Context, tenantID uuid.UUID, priority models.ProjectPriority, pagination PaginationParams) ([]*models.Project, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND priority = ?", tenantID, priority).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count projects", err)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND priority = ?", tenantID, priority).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find projects", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return projects, paginationResult, nil
}

// FindOverdueProjects retrieves overdue projects
func (r *projectRepository) FindOverdueProjects(ctx context.Context, tenantID uuid.UUID) ([]*models.Project, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND status NOT IN (?, ?) AND due_date < ? AND due_date IS NOT NULL",
			tenantID, models.ProjectStatusCompleted, models.ProjectStatusCancelled, time.Now()).
		Order("due_date ASC").
		Find(&projects).Error; err != nil {
		r.logger.Error("failed to find overdue projects", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find overdue projects", err)
	}

	return projects, nil
}

// FindActiveProjects retrieves active projects
func (r *projectRepository) FindActiveProjects(ctx context.Context, tenantID uuid.UUID) ([]*models.Project, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND status IN (?, ?)",
			tenantID, models.ProjectStatusPlanned, models.ProjectStatusInProgress).
		Order("priority DESC, created_at DESC").
		Find(&projects).Error; err != nil {
		r.logger.Error("failed to find active projects", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find active projects", err)
	}

	return projects, nil
}

// Search searches projects by title or description
func (r *projectRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, pagination PaginationParams) ([]*models.Project, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND (title ILIKE ? OR description ILIKE ?)", tenantID, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count projects", err)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND (title ILIKE ? OR description ILIKE ?)", tenantID, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search projects", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return projects, paginationResult, nil
}

// FindByTags retrieves projects by tags
func (r *projectRepository) FindByTags(ctx context.Context, tenantID uuid.UUID, tags []string, pagination PaginationParams) ([]*models.Project, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}
	if len(tags) == 0 {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tags cannot be empty", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND tags && ?", tenantID, tags).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count projects", err)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND tags && ?", tenantID, tags).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find projects by tags", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return projects, paginationResult, nil
}

// FindByDateRange retrieves projects within a date range
func (r *projectRepository) FindByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Project, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND ((start_date BETWEEN ? AND ?) OR (due_date BETWEEN ? AND ?))",
			tenantID, startDate, endDate, startDate, endDate).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count projects", err)
	}

	var projects []*models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND ((start_date BETWEEN ? AND ?) OR (due_date BETWEEN ? AND ?))",
			tenantID, startDate, endDate, startDate, endDate).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_date ASC").
		Find(&projects).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find projects", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return projects, paginationResult, nil
}

// UpdateProgress recalculates and updates project progress
func (r *projectRepository) UpdateProgress(ctx context.Context, projectID uuid.UUID) error {
	if projectID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	// Get task statistics
	var stats struct {
		Total     int64
		Completed int64
		Overdue   int64
		Blocked   int64
	}

	// Total tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ?", projectID).
		Count(&stats.Total)

	// Completed tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status = ?", projectID, models.TaskStatusDone).
		Count(&stats.Completed)

	// Overdue tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status NOT IN (?) AND due_date < ? AND due_date IS NOT NULL",
			projectID, models.TaskStatusDone, time.Now()).
		Count(&stats.Overdue)

	// Blocked tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Where("project_id = ? AND status = ?", projectID, models.TaskStatusBlocked).
		Count(&stats.Blocked)

	// Calculate progress percentage
	progressPercent := 0
	if stats.Total > 0 {
		progressPercent = int((stats.Completed * 100) / stats.Total)
	}

	// Update project
	updates := map[string]interface{}{
		"progress_percent":     progressPercent,
		"tasks_total":          stats.Total,
		"tasks_completed":      stats.Completed,
		"tasks_overdue":        stats.Overdue,
		"active_blocked_tasks": stats.Blocked,
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ?", projectID).
		Updates(updates).Error; err != nil {
		r.logger.Error("failed to update project progress", "project_id", projectID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update project progress", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	return nil
}

// RecalculateAllProgress recalculates progress for all projects in a tenant
func (r *projectRepository) RecalculateAllProgress(ctx context.Context, tenantID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var projectIDs []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status NOT IN (?)", tenantID, models.ProjectStatusCompleted).
		Pluck("id", &projectIDs).Error; err != nil {
		return errors.NewRepositoryError("QUERY_FAILED", "failed to get project IDs", err)
	}

	for _, projectID := range projectIDs {
		if err := r.UpdateProgress(ctx, projectID); err != nil {
			r.logger.Warn("failed to update progress", "project_id", projectID, "error", err)
		}
	}

	r.logger.Info("recalculated progress for all projects", "tenant_id", tenantID, "count", len(projectIDs))
	return nil
}

// GetProjectWithFullDetails retrieves a project with all relationships
func (r *projectRepository) GetProjectWithFullDetails(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var project models.Project
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).
		Preload("Tasks.AssignedTo").
		Preload("Milestones", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).
		Preload("Milestones.Tasks").
		Preload("Updates", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("Updates.User").
		First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "project not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get project with details", "project_id", projectID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to get project", err)
	}

	return &project, nil
}

// StartProject transitions project to in_progress status
func (r *projectRepository) StartProject(ctx context.Context, projectID uuid.UUID) error {
	if projectID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var project models.Project
	if err := r.db.WithContext(ctx).First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "project not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find project", err)
	}

	if project.Status == models.ProjectStatusInProgress {
		return nil // Already in progress
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status": models.ProjectStatusInProgress,
	}

	// Set start date if not already set
	if project.StartDate == nil {
		updates["start_date"] = now
	}

	if err := r.db.WithContext(ctx).
		Model(&project).
		Updates(updates).Error; err != nil {
		r.logger.Error("failed to start project", "project_id", projectID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to start project", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	return nil
}

// PauseProject transitions project to on_hold status
func (r *projectRepository) PauseProject(ctx context.Context, projectID uuid.UUID) error {
	if projectID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ?", projectID).
		Update("status", models.ProjectStatusOnHold).Error; err != nil {
		r.logger.Error("failed to pause project", "project_id", projectID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to pause project", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	return nil
}

// CompleteProject transitions project to completed status
func (r *projectRepository) CompleteProject(ctx context.Context, projectID uuid.UUID) error {
	if projectID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":           models.ProjectStatusCompleted,
		"completed_at":     now,
		"progress_percent": 100,
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ?", projectID).
		Updates(updates).Error; err != nil {
		r.logger.Error("failed to complete project", "project_id", projectID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to complete project", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	return nil
}

// CancelProject transitions project to cancelled status
func (r *projectRepository) CancelProject(ctx context.Context, projectID uuid.UUID, reason string) error {
	if projectID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": models.ProjectStatusCancelled,
	}

	// Store cancellation reason in metadata
	if reason != "" {
		var project models.Project
		if err := r.db.WithContext(ctx).First(&project, projectID).Error; err == nil {
			metadata := project.Metadata
			if metadata == nil {
				metadata = make(models.JSONB)
			}
			metadata["cancellation_reason"] = reason
			metadata["cancelled_at"] = time.Now()
			updates["metadata"] = metadata
		}
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ?", projectID).
		Updates(updates).Error; err != nil {
		r.logger.Error("failed to cancel project", "project_id", projectID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to cancel project", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	return nil
}

// ResumeProject resumes a paused project
func (r *projectRepository) ResumeProject(ctx context.Context, projectID uuid.UUID) error {
	if projectID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var project models.Project
	if err := r.db.WithContext(ctx).First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "project not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find project", err)
	}

	if project.Status != models.ProjectStatusOnHold {
		return errors.NewRepositoryError("INVALID_STATUS", "can only resume projects on hold", errors.ErrInvalidInput)
	}

	if err := r.db.WithContext(ctx).
		Model(&project).
		Update("status", models.ProjectStatusInProgress).Error; err != nil {
		r.logger.Error("failed to resume project", "project_id", projectID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to resume project", err)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	return nil
}

// GetProjectStats retrieves comprehensive project statistics for a tenant
func (r *projectRepository) GetProjectStats(ctx context.Context, tenantID uuid.UUID) (ProjectStats, error) {
	if tenantID == uuid.Nil {
		return ProjectStats{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := ProjectStats{
		ByStatus:   make(map[models.ProjectStatus]int64),
		ByPriority: make(map[models.ProjectPriority]int64),
	}

	// Total projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalProjects)

	// Active projects (planned or in progress)
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status IN (?)",
			tenantID, []models.ProjectStatus{models.ProjectStatusPlanned, models.ProjectStatusInProgress}).
		Count(&stats.ActiveProjects)

	// Completed projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.ProjectStatusCompleted).
		Count(&stats.CompletedProjects)

	// On hold projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.ProjectStatusOnHold).
		Count(&stats.OnHoldProjects)

	// Cancelled projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.ProjectStatusCancelled).
		Count(&stats.CancelledProjects)

	// Overdue projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status NOT IN (?, ?) AND due_date < ? AND due_date IS NOT NULL",
			tenantID, models.ProjectStatusCompleted, models.ProjectStatusCancelled, time.Now()).
		Count(&stats.OverdueProjects)

	// Projects by status
	type StatusCount struct {
		Status models.ProjectStatus
		Count  int64
	}
	var statusCounts []StatusCount
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// Projects by priority
	type PriorityCount struct {
		Priority models.ProjectPriority
		Count    int64
	}
	var priorityCounts []PriorityCount
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Select("priority, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("priority").
		Scan(&priorityCounts)

	for _, pc := range priorityCounts {
		stats.ByPriority[pc.Priority] = pc.Count
	}

	// Average progress
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status NOT IN (?)",
			tenantID, []models.ProjectStatus{models.ProjectStatusCompleted, models.ProjectStatusCancelled}).
		Select("AVG(progress_percent)").
		Scan(&stats.AverageProgress)

	// Total budget
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ?", tenantID).
		Select("SUM(budget_amount)").
		Scan(&stats.TotalBudget)

	// On-time projects (completed before or on due date)
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND status = ? AND completed_at <= due_date AND due_date IS NOT NULL",
			tenantID, models.ProjectStatusCompleted).
		Count(&stats.OnTimeProjects)

	// Completion rate
	if stats.TotalProjects > 0 {
		stats.CompletionRate = (float64(stats.CompletedProjects) / float64(stats.TotalProjects)) * 100
	}

	// Average tasks per project
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("tenant_id = ? AND tasks_total > 0", tenantID).
		Select("AVG(tasks_total)").
		Scan(&stats.AverageTasksPerProject)

	return stats, nil
}

// GetArtisanProjectStats retrieves artisan-specific project statistics
func (r *projectRepository) GetArtisanProjectStats(ctx context.Context, artisanID uuid.UUID) (ArtisanProjectStats, error) {
	if artisanID == uuid.Nil {
		return ArtisanProjectStats{}, errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := ArtisanProjectStats{}

	// Total projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("artisan_id = ?", artisanID).
		Count(&stats.TotalProjects)

	// Active projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("artisan_id = ? AND status IN (?)",
			artisanID, []models.ProjectStatus{models.ProjectStatusPlanned, models.ProjectStatusInProgress}).
		Count(&stats.ActiveProjects)

	// Completed projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("artisan_id = ? AND status = ?", artisanID, models.ProjectStatusCompleted).
		Count(&stats.CompletedProjects)

	// Average progress
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("artisan_id = ? AND status NOT IN (?)",
			artisanID, []models.ProjectStatus{models.ProjectStatusCompleted, models.ProjectStatusCancelled}).
		Select("AVG(progress_percent)").
		Scan(&stats.AverageProgress)

	// Total revenue (completed projects)
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("artisan_id = ? AND status = ?", artisanID, models.ProjectStatusCompleted).
		Select("SUM(budget_amount)").
		Scan(&stats.TotalRevenue)

	// On-time delivery rate
	var onTimeCount, totalCompleted int64
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("artisan_id = ? AND status = ?", artisanID, models.ProjectStatusCompleted).
		Count(&totalCompleted)

	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("artisan_id = ? AND status = ? AND completed_at <= due_date AND due_date IS NOT NULL",
			artisanID, models.ProjectStatusCompleted).
		Count(&onTimeCount)

	if totalCompleted > 0 {
		stats.OnTimeDeliveryRate = (float64(onTimeCount) / float64(totalCompleted)) * 100
	}

	// Average project duration (completed projects)
	r.db.WithContext(ctx).Raw(`
		SELECT AVG(EXTRACT(DAY FROM (completed_at - start_date)))
		FROM projects
		WHERE artisan_id = ?
			AND status = ?
			AND start_date IS NOT NULL
			AND completed_at IS NOT NULL
	`, artisanID, models.ProjectStatusCompleted).Scan(&stats.AverageProjectDuration)

	return stats, nil
}

// GetCustomerProjectStats retrieves customer-specific project statistics
func (r *projectRepository) GetCustomerProjectStats(ctx context.Context, customerID uuid.UUID) (CustomerProjectStats, error) {
	if customerID == uuid.Nil {
		return CustomerProjectStats{}, errors.NewRepositoryError("INVALID_INPUT", "customer_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := CustomerProjectStats{}

	// Total projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("customer_id = ?", customerID).
		Count(&stats.TotalProjects)

	// Active projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("customer_id = ? AND status IN (?)",
			customerID, []models.ProjectStatus{models.ProjectStatusPlanned, models.ProjectStatusInProgress}).
		Count(&stats.ActiveProjects)

	// Completed projects
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("customer_id = ? AND status = ?", customerID, models.ProjectStatusCompleted).
		Count(&stats.CompletedProjects)

	// Total spent
	r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("customer_id = ?", customerID).
		Select("SUM(budget_amount)").
		Scan(&stats.TotalSpent)

	// Average project cost
	if stats.TotalProjects > 0 {
		stats.AverageProjectCost = stats.TotalSpent / float64(stats.TotalProjects)
	}

	return stats, nil
}

// GetProjectTimeline retrieves timeline events for a project
func (r *projectRepository) GetProjectTimeline(ctx context.Context, projectID uuid.UUID) ([]TimelineEvent, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var events []TimelineEvent

	// Get project updates
	var updates []models.ProjectUpdate
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(50).
		Find(&updates).Error; err == nil {
		for _, update := range updates {
			userName := "System"
			if update.User != nil {
				userName = update.User.FirstName + " " + update.User.LastName
			}

			events = append(events, TimelineEvent{
				Date:        update.CreatedAt,
				Type:        string(update.Type),
				Title:       update.Title,
				Description: update.Description,
				UserID:      &update.UserID,
				UserName:    userName,
			})
		}
	}

	// Get milestone completions
	var milestones []models.ProjectMilestone
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND completed_at IS NOT NULL", projectID).
		Order("completed_at DESC").
		Find(&milestones).Error; err == nil {
		for _, milestone := range milestones {
			events = append(events, TimelineEvent{
				Date:        *milestone.CompletedAt,
				Type:        "milestone_completed",
				Title:       "Milestone Completed: " + milestone.Title,
				Description: milestone.Description,
			})
		}
	}

	// Sort events by date (most recent first)
	// In production, use sort.Slice

	return events, nil
}

// GetProjectHealth calculates project health metrics
func (r *projectRepository) GetProjectHealth(ctx context.Context, projectID uuid.UUID) (ProjectHealth, error) {
	if projectID == uuid.Nil {
		return ProjectHealth{}, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var project models.Project
	if err := r.db.WithContext(ctx).First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ProjectHealth{}, errors.NewRepositoryError("NOT_FOUND", "project not found", errors.ErrNotFound)
		}
		return ProjectHealth{}, errors.NewRepositoryError("FIND_FAILED", "failed to find project", err)
	}

	health := ProjectHealth{
		ProjectID:       projectID,
		IsOnTrack:       true,
		RiskLevel:       "low",
		Recommendations: []string{},
	}

	// Check if overdue
	if project.DueDate != nil && time.Now().After(*project.DueDate) &&
		project.Status != models.ProjectStatusCompleted {
		health.IsOverdue = true
		health.IsOnTrack = false
		health.RiskLevel = "high"
		health.Recommendations = append(health.Recommendations, "Project is overdue. Consider reallocating resources.")
	}

	// Check blocked tasks
	health.BlockedTasksCount = project.ActiveBlockedTasks
	if health.BlockedTasksCount > 0 {
		health.IsOnTrack = false
		if health.RiskLevel != "high" {
			health.RiskLevel = "medium"
		}
		health.Recommendations = append(health.Recommendations, "Resolve blocked tasks to improve velocity.")
	}

	// Check overdue tasks
	health.OverdueTasksCount = project.TasksOverdue
	if health.OverdueTasksCount > 3 {
		health.IsOnTrack = false
		health.RiskLevel = "high"
		health.Recommendations = append(health.Recommendations, "Multiple overdue tasks detected. Review task assignments.")
	}

	// Calculate completion velocity (tasks per day)
	if project.StartDate != nil && project.TasksCompleted > 0 {
		daysSinceStart := time.Since(*project.StartDate).Hours() / 24
		if daysSinceStart > 0 {
			health.CompletionVelocity = float64(project.TasksCompleted) / daysSinceStart
		}
	}

	// Calculate health score (0-100)
	healthScore := 100

	// Deduct for overdue
	if health.IsOverdue {
		healthScore -= 30
	}

	// Deduct for blocked tasks
	healthScore -= health.BlockedTasksCount * 10
	if healthScore < 0 {
		healthScore = 0
	}

	// Deduct for low progress
	if project.DueDate != nil && project.StartDate != nil {
		totalDuration := project.DueDate.Sub(*project.StartDate).Hours() / 24
		elapsed := time.Since(*project.StartDate).Hours() / 24
		expectedProgress := (elapsed / totalDuration) * 100

		progressGap := expectedProgress - float64(project.ProgressPercent)
		if progressGap > 20 {
			healthScore -= int(progressGap / 2)
			health.IsOnTrack = false
			health.Recommendations = append(health.Recommendations, "Project is behind schedule. Consider adding resources.")
		}
	}

	// Adjust for overdue tasks
	healthScore -= health.OverdueTasksCount * 5

	// Ensure health score is in range
	healthScore = max(0, min(healthScore, 100))

	health.HealthScore = healthScore

	// Set risk level based on health score
	if healthScore >= 80 {
		health.RiskLevel = "low"
	} else if healthScore >= 60 {
		health.RiskLevel = "medium"
	} else {
		health.RiskLevel = "high"
	}

	return health, nil
}

// BulkUpdateStatus updates status for multiple projects
func (r *projectRepository) BulkUpdateStatus(ctx context.Context, projectIDs []uuid.UUID, status models.ProjectStatus) error {
	if len(projectIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id IN ?", projectIDs).
		Update("status", status)

	if result.Error != nil {
		r.logger.Error("failed to bulk update project status", "count", len(projectIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update status", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	r.logger.Info("bulk updated project status", "count", result.RowsAffected, "status", status)
	return nil
}

// BulkAssignArtisan assigns multiple projects to an artisan
func (r *projectRepository) BulkAssignArtisan(ctx context.Context, projectIDs []uuid.UUID, artisanID uuid.UUID) error {
	if len(projectIDs) == 0 {
		return nil
	}
	if artisanID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id IN ?", projectIDs).
		Update("artisan_id", artisanID)

	if result.Error != nil {
		r.logger.Error("failed to bulk assign artisan", "count", len(projectIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk assign artisan", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	r.logger.Info("bulk assigned artisan to projects", "count", result.RowsAffected, "artisan_id", artisanID)
	return nil
}

// ArchiveCompletedProjects archives (soft deletes) completed projects older than specified duration
func (r *projectRepository) ArchiveCompletedProjects(ctx context.Context, tenantID uuid.UUID, olderThan time.Duration) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	cutoffDate := time.Now().Add(-olderThan)

	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ? AND completed_at < ?",
			tenantID, models.ProjectStatusCompleted, cutoffDate).
		Delete(&models.Project{})

	if result.Error != nil {
		r.logger.Error("failed to archive completed projects", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to archive projects", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:projects:*")
	}

	r.logger.Info("archived completed projects", "tenant_id", tenantID, "count", result.RowsAffected, "older_than", olderThan)
	return nil
}

// GetArtisanDashboard retrieves comprehensive dashboard data for an artisan
func (r *projectRepository) GetArtisanDashboard(ctx context.Context, artisanID uuid.UUID) (ArtisanDashboard, error) {
	if artisanID == uuid.Nil {
		return ArtisanDashboard{}, errors.NewRepositoryError("INVALID_INPUT", "artisan_id cannot be nil", errors.ErrInvalidInput)
	}

	dashboard := ArtisanDashboard{}

	// Get statistics
	stats, err := r.GetArtisanProjectStats(ctx, artisanID)
	if err != nil {
		r.logger.Warn("failed to get artisan stats", "artisan_id", artisanID, "error", err)
	}
	dashboard.Statistics = stats

	// Active projects
	r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Tasks", "assigned_to_id = ?", artisanID).
		Where("artisan_id = ? AND status IN (?)",
			artisanID, []models.ProjectStatus{models.ProjectStatusPlanned, models.ProjectStatusInProgress}).
		Order("priority DESC, due_date ASC").
		Limit(10).
		Find(&dashboard.ActiveProjects)

	// Upcoming deadlines (next 7 days)
	sevenDaysFromNow := time.Now().AddDate(0, 0, 7)
	r.db.WithContext(ctx).
		Preload("Customer").
		Where("artisan_id = ? AND status IN (?) AND due_date BETWEEN ? AND ?",
			artisanID,
			[]models.ProjectStatus{models.ProjectStatusPlanned, models.ProjectStatusInProgress},
			time.Now(), sevenDaysFromNow).
		Order("due_date ASC").
		Limit(5).
		Find(&dashboard.UpcomingDeadlines)

	// Overdue projects
	r.db.WithContext(ctx).
		Preload("Customer").
		Where("artisan_id = ? AND status IN (?) AND due_date < ?",
			artisanID,
			[]models.ProjectStatus{models.ProjectStatusPlanned, models.ProjectStatusInProgress},
			time.Now()).
		Order("due_date ASC").
		Find(&dashboard.OverdueProjects)

	// Recently completed (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	r.db.WithContext(ctx).
		Preload("Customer").
		Where("artisan_id = ? AND status = ? AND completed_at >= ?",
			artisanID, models.ProjectStatusCompleted, thirtyDaysAgo).
		Order("completed_at DESC").
		Limit(5).
		Find(&dashboard.RecentlyCompleted)

	// Tasks summary
	var tasksSummary TasksSummary

	// Total tasks assigned to artisan
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Joins("JOIN projects ON projects.id = project_tasks.project_id").
		Where("projects.artisan_id = ? OR project_tasks.assigned_to_id = ?", artisanID, artisanID).
		Count(&tasksSummary.TotalTasks)

	// Completed tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Joins("JOIN projects ON projects.id = project_tasks.project_id").
		Where("(projects.artisan_id = ? OR project_tasks.assigned_to_id = ?) AND project_tasks.status = ?",
			artisanID, artisanID, models.TaskStatusDone).
		Count(&tasksSummary.CompletedTasks)

	// In progress tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Joins("JOIN projects ON projects.id = project_tasks.project_id").
		Where("(projects.artisan_id = ? OR project_tasks.assigned_to_id = ?) AND project_tasks.status = ?",
			artisanID, artisanID, models.TaskStatusInProgress).
		Count(&tasksSummary.InProgressTasks)

	// Overdue tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Joins("JOIN projects ON projects.id = project_tasks.project_id").
		Where("(projects.artisan_id = ? OR project_tasks.assigned_to_id = ?) AND project_tasks.status NOT IN (?) AND project_tasks.due_date < ?",
			artisanID, artisanID, models.TaskStatusDone, time.Now()).
		Count(&tasksSummary.OverdueTasks)

	// Blocked tasks
	r.db.WithContext(ctx).
		Model(&models.ProjectTask{}).
		Joins("JOIN projects ON projects.id = project_tasks.project_id").
		Where("(projects.artisan_id = ? OR project_tasks.assigned_to_id = ?) AND project_tasks.status = ?",
			artisanID, artisanID, models.TaskStatusBlocked).
		Count(&tasksSummary.BlockedTasks)

	dashboard.TasksSummary = tasksSummary

	return dashboard, nil
}

// GetTenantDashboard retrieves comprehensive dashboard data for a tenant
func (r *projectRepository) GetTenantDashboard(ctx context.Context, tenantID uuid.UUID) (TenantProjectDashboard, error) {
	if tenantID == uuid.Nil {
		return TenantProjectDashboard{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	dashboard := TenantProjectDashboard{}

	// Get statistics
	stats, err := r.GetProjectStats(ctx, tenantID)
	if err != nil {
		r.logger.Warn("failed to get project stats", "tenant_id", tenantID, "error", err)
	}
	dashboard.Statistics = stats

	// Active projects
	r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND status IN (?)",
			tenantID, []models.ProjectStatus{models.ProjectStatusPlanned, models.ProjectStatusInProgress}).
		Order("priority DESC, created_at DESC").
		Limit(10).
		Find(&dashboard.ActiveProjects)

	// High priority projects
	r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND priority = ? AND status NOT IN (?)",
			tenantID, models.ProjectPriorityHigh,
			[]models.ProjectStatus{models.ProjectStatusCompleted, models.ProjectStatusCancelled}).
		Order("due_date ASC").
		Limit(5).
		Find(&dashboard.HighPriorityProjects)

	// Overdue projects
	r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Customer").
		Where("tenant_id = ? AND status NOT IN (?) AND due_date < ?",
			tenantID,
			[]models.ProjectStatus{models.ProjectStatusCompleted, models.ProjectStatusCancelled},
			time.Now()).
		Order("due_date ASC").
		Find(&dashboard.OverdueProjects)

	// Recent activity (from project updates)
	var updates []models.ProjectUpdate
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(20).
		Find(&updates).Error; err == nil {

		for _, update := range updates {
			userName := "System"
			if update.User != nil {
				userName = update.User.FirstName + " " + update.User.LastName
			}

			projectTitle := ""
			if update.Project != nil {
				projectTitle = update.Project.Title
			}

			dashboard.RecentActivity = append(dashboard.RecentActivity, TimelineEvent{
				Date:        update.CreatedAt,
				Type:        string(update.Type),
				Title:       update.Title,
				Description: projectTitle + ": " + update.Description,
				UserID:      &update.UserID,
				UserName:    userName,
			})
		}
	}

	// Top artisans by performance
	type ArtisanStats struct {
		ArtisanID         uuid.UUID
		ActiveProjects    int64
		CompletedProjects int64
		AverageProgress   float64
		OnTimeCount       int64
		TotalCompleted    int64
	}

	var artisanStats []ArtisanStats
	r.db.WithContext(ctx).Raw(`
		SELECT
			artisan_id,
			COUNT(CASE WHEN status IN ('planned', 'in_progress') THEN 1 END) as active_projects,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_projects,
			AVG(CASE WHEN status NOT IN ('completed', 'cancelled') THEN progress_percent ELSE NULL END) as average_progress,
			COUNT(CASE WHEN status = 'completed' AND completed_at <= due_date THEN 1 END) as on_time_count,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as total_completed
		FROM projects
		WHERE tenant_id = ?
		GROUP BY artisan_id
		HAVING COUNT(*) > 0
		ORDER BY active_projects DESC, average_progress DESC
		LIMIT 5
	`, tenantID).Scan(&artisanStats)

	for _, stat := range artisanStats {
		// Get artisan name
		var artisan models.Artisan
		var artisanName string
		if err := r.db.WithContext(ctx).
			Preload("User").
			First(&artisan, stat.ArtisanID).Error; err == nil && artisan.User != nil {
			artisanName = artisan.User.FirstName + " " + artisan.User.LastName
		}

		onTimeRate := 0.0
		if stat.TotalCompleted > 0 {
			onTimeRate = (float64(stat.OnTimeCount) / float64(stat.TotalCompleted)) * 100
		}

		dashboard.TopArtisans = append(dashboard.TopArtisans, ArtisanPerformance{
			ArtisanID:         stat.ArtisanID,
			ArtisanName:       artisanName,
			ActiveProjects:    int(stat.ActiveProjects),
			CompletedProjects: int(stat.CompletedProjects),
			AverageProgress:   stat.AverageProgress,
			OnTimeRate:        onTimeRate,
		})
	}

	return dashboard, nil
}
