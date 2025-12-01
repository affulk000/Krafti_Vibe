package repository

import (
	"context"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectUpdateRepository defines the interface for project update repository operations
type ProjectUpdateRepository interface {
	BaseRepository[models.ProjectUpdate]

	// Query Operations
	FindByProjectID(ctx context.Context, projectID uuid.UUID, pagination PaginationParams) ([]*models.ProjectUpdate, PaginationResult, error)
	FindByType(ctx context.Context, projectID uuid.UUID, updateType models.UpdateType, pagination PaginationParams) ([]*models.ProjectUpdate, PaginationResult, error)
	FindCustomerVisible(ctx context.Context, projectID uuid.UUID, pagination PaginationParams) ([]*models.ProjectUpdate, PaginationResult, error)
	GetLatestUpdate(ctx context.Context, projectID uuid.UUID) (*models.ProjectUpdate, error)
}

// projectUpdateRepository implements ProjectUpdateRepository
type projectUpdateRepository struct {
	BaseRepository[models.ProjectUpdate]
	db     *gorm.DB
	logger log.AllLogger
}

// NewProjectUpdateRepository creates a new project update repository
func NewProjectUpdateRepository(db *gorm.DB, config ...RepositoryConfig) ProjectUpdateRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.ProjectUpdate](db, cfg)

	return &projectUpdateRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
	}
}

// FindByProjectID retrieves all updates for a project
func (r *projectUpdateRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, pagination PaginationParams) ([]*models.ProjectUpdate, PaginationResult, error) {
	if projectID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.ProjectUpdate{}).
		Where("project_id = ?", projectID).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count project updates", "project_id", projectID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count updates", err)
	}

	var updates []*models.ProjectUpdate
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Find(&updates).Error; err != nil {
		r.logger.Error("failed to find project updates", "project_id", projectID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find updates", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return updates, paginationResult, nil
}

// FindByType retrieves updates by type
func (r *projectUpdateRepository) FindByType(ctx context.Context, projectID uuid.UUID, updateType models.UpdateType, pagination PaginationParams) ([]*models.ProjectUpdate, PaginationResult, error) {
	if projectID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.ProjectUpdate{}).
		Where("project_id = ? AND type = ?", projectID, updateType).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count project updates by type", "project_id", projectID, "type", updateType, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count updates", err)
	}

	var updates []*models.ProjectUpdate
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ? AND type = ?", projectID, updateType).
		Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Find(&updates).Error; err != nil {
		r.logger.Error("failed to find project updates by type", "project_id", projectID, "type", updateType, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find updates", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return updates, paginationResult, nil
}

// FindCustomerVisible retrieves customer-visible updates
func (r *projectUpdateRepository) FindCustomerVisible(ctx context.Context, projectID uuid.UUID, pagination PaginationParams) ([]*models.ProjectUpdate, PaginationResult, error) {
	if projectID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.ProjectUpdate{}).
		Where("project_id = ? AND visible_to_customer = ?", projectID, true).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count customer-visible updates", "project_id", projectID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count updates", err)
	}

	var updates []*models.ProjectUpdate
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ? AND visible_to_customer = ?", projectID, true).
		Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Find(&updates).Error; err != nil {
		r.logger.Error("failed to find customer-visible updates", "project_id", projectID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find updates", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return updates, paginationResult, nil
}

// GetLatestUpdate retrieves the latest update for a project
func (r *projectUpdateRepository) GetLatestUpdate(ctx context.Context, projectID uuid.UUID) (*models.ProjectUpdate, error) {
	if projectID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "project_id cannot be nil", errors.ErrInvalidInput)
	}

	var update models.ProjectUpdate
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		First(&update).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "no updates found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get latest update", "project_id", projectID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to get latest update", err)
	}

	return &update, nil
}
