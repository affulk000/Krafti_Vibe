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

// ProjectMilestoneRepository defines the interface for project milestone repository operations
type ProjectMilestoneRepository interface {
	BaseRepository[models.ProjectMilestone]

	// Milestone Queries
	FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMilestone, error)
	FindByStatus(ctx context.Context, projectID uuid.UUID, status models.MilestoneStatus) ([]*models.ProjectMilestone, error)
	FindOverdueMilestones(ctx context.Context, tenantID uuid.UUID) ([]*models.ProjectMilestone, error)
	FindUpcomingMilestones(ctx context.Context, projectID uuid.UUID, days int) ([]*models.ProjectMilestone, error)
	FindPendingApproval(ctx context.Context, tenantID uuid.UUID) ([]*models.ProjectMilestone, error)
	FindPaymentMilestones(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMilestone, error)
	GetPaymentMilestones(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMilestone, error)
	FindUnpaidMilestones(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMilestone, error)

	// Status Management
	UpdateStatus(ctx context.Context, milestoneID uuid.UUID, status models.MilestoneStatus) error
	StartMilestone(ctx context.Context, milestoneID uuid.UUID) error
	CompleteMilestone(ctx context.Context, milestoneID uuid.UUID) error
	CancelMilestone(ctx context.Context, milestoneID uuid.UUID) error

	// Approval Management
	SubmitForApproval(ctx context.Context, milestoneID uuid.UUID) error
	ApproveMilestone(ctx context.Context, milestoneID uuid.UUID) error
	RejectMilestone(ctx context.Context, milestoneID uuid.UUID, reason string) error

	// Payment Management
	MarkPaymentReceived(ctx context.Context, milestoneID uuid.UUID) error
	GetPaymentSummary(ctx context.Context, projectID uuid.UUID) (PaymentSummary, error)

	// Analytics
	GetMilestoneStats(ctx context.Context, projectID uuid.UUID) (MilestoneStats, error)
	GetMilestoneProgress(ctx context.Context, projectID uuid.UUID) (float64, error)

	// Ordering
	ReorderMilestones(ctx context.Context, projectID uuid.UUID, milestoneOrders map[uuid.UUID]int) error
}

// PaymentSummary represents payment summary for a project
type PaymentSummary struct {
	ProjectID         uuid.UUID `json:"project_id"`
	TotalMilestones   int64     `json:"total_milestones"`
	PaymentMilestones int64     `json:"payment_milestones"`
	TotalAmount       float64   `json:"total_amount"`
	PaidAmount        float64   `json:"paid_amount"`
	PendingAmount     float64   `json:"pending_amount"`
	PaidMilestones    int64     `json:"paid_milestones"`
	UnpaidMilestones  int64     `json:"unpaid_milestones"`
	PaymentPercentage float64   `json:"payment_percentage"`
}

// MilestoneStats represents milestone statistics
type MilestoneStats struct {
	TotalMilestones       int64                            `json:"total_milestones"`
	PendingMilestones     int64                            `json:"pending_milestones"`
	InProgressMilestones  int64                            `json:"in_progress_milestones"`
	CompletedMilestones   int64                            `json:"completed_milestones"`
	ByStatus              map[models.MilestoneStatus]int64 `json:"by_status"`
	OverdueMilestones     int64                            `json:"overdue_milestones"`
	PendingApproval       int64                            `json:"pending_approval"`
	TotalPaymentAmount    float64                          `json:"total_payment_amount"`
	ReceivedPayment       float64                          `json:"received_payment"`
	CompletionRate        float64                          `json:"completion_rate"`
	AverageCompletionDays float64                          `json:"average_completion_days"`
}

// projectMilestoneRepository implements ProjectMilestoneRepository
type projectMilestoneRepository struct {
	BaseRepository[models.ProjectMilestone]
	db     *gorm.DB
	logger log.AllLogger
	cache  Cache
}

// NewProjectMilestoneRepository creates a new ProjectMilestoneRepository instance
func NewProjectMilestoneRepository(db *gorm.DB, config ...RepositoryConfig) ProjectMilestoneRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.ProjectMilestone](db, cfg)

	return &projectMilestoneRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
	}
}

// FindByProjectID retrieves all milestones for a project
func (r *projectMilestoneRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMilestone, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var milestones []*models.ProjectMilestone
	if err := r.db.WithContext(ctx).
		Preload("Tasks").
		Where("project_id = ?", projectID).
		Order("order_index ASC, due_date ASC").
		Find(&milestones).Error; err != nil {
		r.logger.Error("failed to find milestones", "project_id", projectID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find milestones", err)
	}

	return milestones, nil
}

// FindByStatus retrieves milestones by status
func (r *projectMilestoneRepository) FindByStatus(ctx context.Context, projectID uuid.UUID, status models.MilestoneStatus) ([]*models.ProjectMilestone, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var milestones []*models.ProjectMilestone
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND status = ?", projectID, status).
		Order("order_index ASC").
		Find(&milestones).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find milestones by status", err)
	}

	return milestones, nil
}

// FindOverdueMilestones retrieves all overdue milestones for a tenant
func (r *projectMilestoneRepository) FindOverdueMilestones(ctx context.Context, tenantID uuid.UUID) ([]*models.ProjectMilestone, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var milestones []*models.ProjectMilestone
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Where("tenant_id = ? AND status NOT IN (?, ?) AND due_date < ? AND due_date IS NOT NULL",
			tenantID, models.MilestoneStatusCompleted, models.MilestoneStatusCancelled, time.Now()).
		Order("due_date ASC").
		Find(&milestones).Error; err != nil {
		r.logger.Error("failed to find overdue milestones", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find overdue milestones", err)
	}

	return milestones, nil
}

// FindPendingApproval retrieves milestones pending customer approval
func (r *projectMilestoneRepository) FindPendingApproval(ctx context.Context, tenantID uuid.UUID) ([]*models.ProjectMilestone, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var milestones []*models.ProjectMilestone
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Where("tenant_id = ? AND requires_approval = ? AND approved_by_customer = ? AND status = ?",
			tenantID, true, false, models.MilestoneStatusCompleted).
		Order("completed_at ASC").
		Find(&milestones).Error; err != nil {
		r.logger.Error("failed to find pending approval milestones", "tenant_id", tenantID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find pending approval milestones", err)
	}

	return milestones, nil
}

// FindPaymentMilestones retrieves all payment milestones for a project
func (r *projectMilestoneRepository) FindPaymentMilestones(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMilestone, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var milestones []*models.ProjectMilestone
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND is_payment_milestone = ?", projectID, true).
		Order("order_index ASC").
		Find(&milestones).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find payment milestones", err)
	}

	return milestones, nil
}

// FindUpcomingMilestones retrieves milestones due within the specified number of days
func (r *projectMilestoneRepository) FindUpcomingMilestones(ctx context.Context, projectID uuid.UUID, days int) ([]*models.ProjectMilestone, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var milestones []*models.ProjectMilestone
	endDate := time.Now().AddDate(0, 0, days)

	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND status NOT IN (?, ?) AND due_date >= ? AND due_date <= ?",
			projectID, models.MilestoneStatusCompleted, models.MilestoneStatusCancelled, time.Now(), endDate).
		Order("due_date ASC").
		Find(&milestones).Error; err != nil {
		r.logger.Error("failed to find upcoming milestones", "project_id", projectID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find upcoming milestones", err)
	}

	return milestones, nil
}

// GetPaymentMilestones is an alias for FindPaymentMilestones
func (r *projectMilestoneRepository) GetPaymentMilestones(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMilestone, error) {
	return r.FindPaymentMilestones(ctx, projectID)
}

// FindUnpaidMilestones retrieves unpaid milestones for a project
func (r *projectMilestoneRepository) FindUnpaidMilestones(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMilestone, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var milestones []*models.ProjectMilestone
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND is_payment_milestone = ? AND payment_received = ?",
			projectID, true, false).
		Order("order_index ASC").
		Find(&milestones).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find unpaid milestones", err)
	}

	return milestones, nil
}

// UpdateStatus updates the status of a milestone
func (r *projectMilestoneRepository) UpdateStatus(ctx context.Context, milestoneID uuid.UUID, status models.MilestoneStatus) error {
	if milestoneID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": status,
	}

	// Set completed_at if status is completed
	if status == models.MilestoneStatusCompleted {
		updates["completed_at"] = time.Now()
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("id = ?", milestoneID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to update milestone status", "milestone_id", milestoneID, "status", status, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update milestone status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "milestone not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_milestones:*")
	}

	return nil
}

// StartMilestone transitions milestone to in_progress status
func (r *projectMilestoneRepository) StartMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	if milestoneID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": models.MilestoneStatusInProgress,
	}

	// Set start date if not already set
	var milestone models.ProjectMilestone
	if err := r.db.WithContext(ctx).First(&milestone, milestoneID).Error; err == nil {
		if milestone.StartDate == nil {
			now := time.Now()
			updates["start_date"] = now
		}
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("id = ?", milestoneID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to start milestone", "milestone_id", milestoneID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to start milestone", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "milestone not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_milestones:*")
	}

	return nil
}

// CompleteMilestone marks a milestone as completed
func (r *projectMilestoneRepository) CompleteMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	if milestoneID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.MilestoneStatusCompleted,
		"completed_at": now,
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("id = ?", milestoneID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to complete milestone", "milestone_id", milestoneID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to complete milestone", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "milestone not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_milestones:*")
	}

	return nil
}

// CancelMilestone cancels a milestone
func (r *projectMilestoneRepository) CancelMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	if milestoneID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("id = ?", milestoneID).
		Update("status", models.MilestoneStatusCancelled)

	if result.Error != nil {
		r.logger.Error("failed to cancel milestone", "milestone_id", milestoneID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to cancel milestone", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "milestone not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_milestones:*")
	}

	return nil
}

// SubmitForApproval submits milestone for customer approval
func (r *projectMilestoneRepository) SubmitForApproval(ctx context.Context, milestoneID uuid.UUID) error {
	if milestoneID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	var milestone models.ProjectMilestone
	if err := r.db.WithContext(ctx).First(&milestone, milestoneID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "milestone not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("FIND_FAILED", "failed to find milestone", err)
	}

	if !milestone.RequiresApproval {
		return errors.NewRepositoryError("INVALID_OPERATION", "milestone does not require approval", errors.ErrInvalidInput)
	}

	if milestone.Status != models.MilestoneStatusCompleted {
		return errors.NewRepositoryError("INVALID_STATUS", "milestone must be completed before submission", errors.ErrInvalidInput)
	}

	// Set to pending approval (completed but not approved)
	// The status remains completed, we just track approval state separately
	return nil
}

// ApproveMilestone approves a milestone
func (r *projectMilestoneRepository) ApproveMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	if milestoneID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"approved_by_customer": true,
		"approved_at":          now,
		"rejection_reason":     "",
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("id = ?", milestoneID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to approve milestone", "milestone_id", milestoneID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to approve milestone", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "milestone not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_milestones:*")
	}

	return nil
}

// RejectMilestone rejects a milestone with a reason
func (r *projectMilestoneRepository) RejectMilestone(ctx context.Context, milestoneID uuid.UUID, reason string) error {
	if milestoneID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"approved_by_customer": false,
		"rejection_reason":     reason,
		"status":               models.MilestoneStatusInProgress, // Return to in progress
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("id = ?", milestoneID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to reject milestone", "milestone_id", milestoneID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to reject milestone", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "milestone not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_milestones:*")
	}

	return nil
}

// MarkPaymentReceived marks a milestone payment as received
func (r *projectMilestoneRepository) MarkPaymentReceived(ctx context.Context, milestoneID uuid.UUID) error {
	if milestoneID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "milestone_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"payment_received":    true,
		"payment_received_at": now,
	}

	result := r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("id = ? AND is_payment_milestone = ?", milestoneID, true).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to mark payment received", "milestone_id", milestoneID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark payment received", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "payment milestone not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_milestones:*")
	}

	return nil
}

// GetPaymentSummary retrieves payment summary for a project
func (r *projectMilestoneRepository) GetPaymentSummary(ctx context.Context, projectID uuid.UUID) (PaymentSummary, error) {
	if projectID == uuid.Nil {
		return PaymentSummary{}, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	summary := PaymentSummary{
		ProjectID: projectID,
	}

	// Total milestones
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ?", projectID).
		Count(&summary.TotalMilestones)

	// Payment milestones
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND is_payment_milestone = ?", projectID, true).
		Count(&summary.PaymentMilestones)

	// Total amount
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND is_payment_milestone = ?", projectID, true).
		Select("SUM(payment_amount)").
		Scan(&summary.TotalAmount)

	// Paid amount
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND is_payment_milestone = ? AND payment_received = ?",
			projectID, true, true).
		Select("SUM(payment_amount)").
		Scan(&summary.PaidAmount)

	// Pending amount
	summary.PendingAmount = summary.TotalAmount - summary.PaidAmount

	// Paid milestones count
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND is_payment_milestone = ? AND payment_received = ?",
			projectID, true, true).
		Count(&summary.PaidMilestones)

	// Unpaid milestones count
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND is_payment_milestone = ? AND payment_received = ?",
			projectID, true, false).
		Count(&summary.UnpaidMilestones)

	// Payment percentage
	if summary.TotalAmount > 0 {
		summary.PaymentPercentage = (summary.PaidAmount / summary.TotalAmount) * 100
	}

	return summary, nil
}

// GetMilestoneStats retrieves milestone statistics for a project
func (r *projectMilestoneRepository) GetMilestoneStats(ctx context.Context, projectID uuid.UUID) (MilestoneStats, error) {
	if projectID == uuid.Nil {
		return MilestoneStats{}, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := MilestoneStats{
		ByStatus: make(map[models.MilestoneStatus]int64),
	}

	// Total milestones
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ?", projectID).
		Count(&stats.TotalMilestones)

	// Pending milestones
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND status = ?", projectID, models.MilestoneStatusPending).
		Count(&stats.PendingMilestones)

	// In-progress milestones
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND status = ?", projectID, models.MilestoneStatusInProgress).
		Count(&stats.InProgressMilestones)

	// Completed milestones
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND status = ?", projectID, models.MilestoneStatusCompleted).
		Count(&stats.CompletedMilestones)

	// Milestones by status
	type StatusCount struct {
		Status models.MilestoneStatus
		Count  int64
	}
	var statusCounts []StatusCount
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Select("status, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// Overdue milestones
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND status NOT IN (?, ?) AND due_date < ?",
			projectID, models.MilestoneStatusCompleted, models.MilestoneStatusCancelled, time.Now()).
		Count(&stats.OverdueMilestones)

	// Pending approval
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND requires_approval = ? AND approved_by_customer = ? AND status = ?",
			projectID, true, false, models.MilestoneStatusCompleted).
		Count(&stats.PendingApproval)

	// Total payment amount
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND is_payment_milestone = ?", projectID, true).
		Select("COALESCE(SUM(payment_amount), 0)").
		Scan(&stats.TotalPaymentAmount)

	// Received payment
	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND is_payment_milestone = ? AND payment_received = ?", projectID, true, true).
		Select("COALESCE(SUM(payment_amount), 0)").
		Scan(&stats.ReceivedPayment)

	// Completion rate
	if stats.TotalMilestones > 0 {
		stats.CompletionRate = (float64(stats.CompletedMilestones) / float64(stats.TotalMilestones)) * 100
	}

	// Average completion time
	r.db.WithContext(ctx).Raw(`
		SELECT AVG(EXTRACT(DAY FROM (completed_at - start_date)))
		FROM project_milestones
		WHERE project_id = ?
			AND status = ?
			AND start_date IS NOT NULL
			AND completed_at IS NOT NULL
	`, projectID, models.MilestoneStatusCompleted).Scan(&stats.AverageCompletionDays)

	return stats, nil
}

// GetMilestoneProgress calculates overall milestone completion percentage
func (r *projectMilestoneRepository) GetMilestoneProgress(ctx context.Context, projectID uuid.UUID) (float64, error) {
	if projectID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var total, completed int64

	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ?", projectID).
		Count(&total)

	r.db.WithContext(ctx).
		Model(&models.ProjectMilestone{}).
		Where("project_id = ? AND status = ?", projectID, models.MilestoneStatusCompleted).
		Count(&completed)

	if total == 0 {
		return 0, nil
	}

	progress := (float64(completed) / float64(total)) * 100
	return progress, nil
}

// ReorderMilestones updates the order_index for multiple milestones
func (r *projectMilestoneRepository) ReorderMilestones(ctx context.Context, projectID uuid.UUID, milestoneOrders map[uuid.UUID]int) error {
	if projectID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}
	if len(milestoneOrders) == 0 {
		return nil
	}

	// Update each milestone's order_index
	for milestoneID, orderIndex := range milestoneOrders {
		if err := r.db.WithContext(ctx).
			Model(&models.ProjectMilestone{}).
			Where("id = ? AND project_id = ?", milestoneID, projectID).
			Update("order_index", orderIndex).Error; err != nil {
			r.logger.Warn("failed to update milestone order", "milestone_id", milestoneID, "error", err)
		}
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:project_milestones:*")
	}

	r.logger.Info("reordered milestones", "project_id", projectID, "count", len(milestoneOrders))
	return nil
}
