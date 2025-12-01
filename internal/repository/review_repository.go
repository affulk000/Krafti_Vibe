package repository

import (
	"Krafti_Vibe/internal/domain/models"
	errs "Krafti_Vibe/internal/pkg/errors"
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

// ReviewRepository handles review-related database operations
type ReviewRepository struct {
	BaseRepository[models.Review]
	db     *gorm.DB
	logger log.AllLogger
}

func NewReviewRepository(db *gorm.DB, logger log.AllLogger) *ReviewRepository {
	return &ReviewRepository{
		NewBaseRepository[models.Review](db),
		db,
		logger,
	}
}

// FindByTenantID retrieves reviews for a tenant
func (r *ReviewRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if err := query.Model(&models.Review{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Preload("Artisan").
		Preload("Customer").
		Preload("Service").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		r.logger.Error("failed to retrieve review for a tenant", err)
		return nil, 0, err
	}

	r.logger.Info("successfully retrieved reviews for a tenant")
	return reviews, total, nil
}

// FindByArtisanID retrieves reviews for an artisan
func (r *ReviewRepository) FindByArtisanID(ctx context.Context, artisanID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	if err := r.db.WithContext(ctx).
		Where("artisan_id = ? AND is_published = ?", artisanID, true).
		Preload("Customer").
		Preload("Service").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

// FindByCustomerID retrieves reviews by a customer
func (r *ReviewRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	if err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Preload("Artisan").
		Preload("Service").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		r.logger.Error("failed to retrieve reviews by customer", err)
		return nil, err
	}

	r.logger.Info("successfully retrieved reviews by customer")
	return reviews, nil
}

// FindByBookingID retrieves review for a specific booking
func (r *ReviewRepository) FindByBookingID(ctx context.Context, bookingID uuid.UUID) (*models.Review, error) {
	var review models.Review
	if err := r.db.WithContext(ctx).
		Where("booking_id = ?", bookingID).
		Preload("Artisan").
		Preload("Customer").
		Preload("Service").
		First(&review).Error; err != nil {
		r.logger.Error("failed to retrieve review by booking ID", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Info("review not found")
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	r.logger.Info("successfully retrieved review by booking ID")
	return &review, nil
}

// GetAverageRating calculates average rating for an artisan
func (r *ReviewRepository) GetAverageRating(ctx context.Context, artisanID uuid.UUID) (float64, error) {
	var avgRating float64
	if err := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("artisan_id = ? AND is_published = ?", artisanID, true).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&avgRating).Error; err != nil {
		return 0, err
	}
	return avgRating, nil
}

// GetRatingDistribution gets rating distribution for an artisan
func (r *ReviewRepository) GetRatingDistribution(ctx context.Context, artisanID uuid.UUID) (map[int]int64, error) {
	var results []struct {
		Rating int
		Count  int64
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Select("rating, COUNT(*) as count").
		Where("artisan_id = ? AND is_published = ?", artisanID, true).
		Group("rating").
		Scan(&results).Error; err != nil {
		r.logger.Errorf("failed to get rating distribution for an artisan %s:", err)
		return nil, err
	}

	distribution := make(map[int]int64)
	for _, r := range results {
		distribution[r.Rating] = r.Count
	}

	r.logger.Infof("rating distribution for artisan %s: %v", artisanID, distribution)
	return distribution, nil
}

// FindPendingModeration retrieves reviews pending moderation
func (r *ReviewRepository) FindPendingModeration(ctx context.Context, tenantID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_published = ? AND moderated_at IS NULL", tenantID, false).
		Preload("Artisan").
		Preload("Customer").
		Preload("Service").
		Order("created_at ASC").
		Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

// FindFlagged retrieves flagged reviews
func (r *ReviewRepository) FindFlagged(ctx context.Context, tenantID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_flagged = ?", tenantID, true).
		Preload("Artisan").
		Preload("Customer").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		r.logger.Errorf("failed to find flagged reviews for tenant %s: %v", tenantID, err)
		return nil, err
	}
	return reviews, nil
}

// UpdatePublishStatus updates review publish status
func (r *ReviewRepository) UpdatePublishStatus(ctx context.Context, reviewID uuid.UUID, isPublished bool) error {
	return r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("id = ?", reviewID).
		Update("is_published", isPublished).Error
}

// AddResponse adds artisan response to a review
func (r *ReviewRepository) AddResponse(ctx context.Context, reviewID uuid.UUID, responseText string, respondedBy uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("id = ?", reviewID).
		Updates(map[string]any{
			"response_text": responseText,
			"responsed_at":  &now,
			"responsed_by":  &respondedBy,
		}).Error
}

// IncrementHelpfulCount increments helpful vote count
func (r *ReviewRepository) IncrementHelpfulCount(ctx context.Context, reviewID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("id = ?", reviewID).
		UpdateColumn("helpful_count", gorm.Expr("helpful_count + ?", 1)).Error
}

// IncrementNotHelpfulCount increments not helpful vote count
func (r *ReviewRepository) IncrementNotHelpfulCount(ctx context.Context, reviewID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Review{}).
		Where("id = ?", reviewID).
		UpdateColumn("not_helpful_count", gorm.Expr("not_helpful_count + ?", 1)).Error
}
