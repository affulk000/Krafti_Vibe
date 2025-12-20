package repository

import (
	"context"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WhiteLabelRepository defines the interface for whitelabel data operations
type WhiteLabelRepository interface {
	// CRUD operations
	Create(ctx context.Context, whitelabel *models.WhiteLabel) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.WhiteLabel, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.WhiteLabel, error)
	GetByCustomDomain(ctx context.Context, domain string) (*models.WhiteLabel, error)
	Update(ctx context.Context, whitelabel *models.WhiteLabel) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Activation
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error

	// Domain operations
	CheckDomainAvailability(ctx context.Context, domain string, excludeTenantID uuid.UUID) (bool, error)
}

type whiteLabelRepository struct {
	db *gorm.DB
}

// NewWhiteLabelRepository creates a new whitelabel repository
func NewWhiteLabelRepository(db *gorm.DB) WhiteLabelRepository {
	return &whiteLabelRepository{
		db: db,
	}
}

func (r *whiteLabelRepository) Create(ctx context.Context, whitelabel *models.WhiteLabel) error {
	return r.db.WithContext(ctx).Create(whitelabel).Error
}

func (r *whiteLabelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WhiteLabel, error) {
	var whitelabel models.WhiteLabel
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("id = ?", id).
		First(&whitelabel).Error
	if err != nil {
		return nil, err
	}
	return &whitelabel, nil
}

func (r *whiteLabelRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.WhiteLabel, error) {
	var whitelabel models.WhiteLabel
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("tenant_id = ?", tenantID).
		First(&whitelabel).Error
	if err != nil {
		return nil, err
	}
	return &whitelabel, nil
}

func (r *whiteLabelRepository) GetByCustomDomain(ctx context.Context, domain string) (*models.WhiteLabel, error) {
	var whitelabel models.WhiteLabel
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("custom_domain = ? AND custom_domain_enabled = ? AND is_active = ?", domain, true, true).
		First(&whitelabel).Error
	if err != nil {
		return nil, err
	}
	return &whitelabel, nil
}

func (r *whiteLabelRepository) Update(ctx context.Context, whitelabel *models.WhiteLabel) error {
	return r.db.WithContext(ctx).Save(whitelabel).Error
}

func (r *whiteLabelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.WhiteLabel{}, id).Error
}

func (r *whiteLabelRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.WhiteLabel{}).
		Where("id = ?", id).
		Update("is_active", true).Error
}

func (r *whiteLabelRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.WhiteLabel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *whiteLabelRepository) CheckDomainAvailability(ctx context.Context, domain string, excludeTenantID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.WhiteLabel{}).
		Where("custom_domain = ? AND tenant_id != ?", domain, excludeTenantID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
