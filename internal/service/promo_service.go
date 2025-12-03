package service

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// PromoCodeService defines the interface for promo code service operations
type PromoCodeService interface {
	// CRUD Operations
	CreatePromoCode(ctx context.Context, tenantID *uuid.UUID, req *dto.CreatePromoCodeRequest) (*dto.PromoCodeResponse, error)
	GetPromoCode(ctx context.Context, id uuid.UUID) (*dto.PromoCodeResponse, error)
	GetPromoCodeByCode(ctx context.Context, code string) (*dto.PromoCodeResponse, error)
	UpdatePromoCode(ctx context.Context, id uuid.UUID, req *dto.UpdatePromoCodeRequest) (*dto.PromoCodeResponse, error)
	DeletePromoCode(ctx context.Context, id uuid.UUID) error

	// Promo Code Management
	ListPromoCodes(ctx context.Context, filter *dto.PromoCodeFilter) (*dto.PromoCodeListResponse, error)
	GetActivePromoCodes(ctx context.Context, tenantID *uuid.UUID, page, pageSize int) (*dto.PromoCodeListResponse, error)
	GetExpiredPromoCodes(ctx context.Context, tenantID uuid.UUID) ([]*dto.PromoCodeResponse, error)
	GetExpiringPromoCodes(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.PromoCodeResponse, error)
	SearchPromoCodes(ctx context.Context, query string, tenantID *uuid.UUID, page, pageSize int) (*dto.PromoCodeListResponse, error)

	// Validation & Usage
	ValidatePromoCode(ctx context.Context, req *dto.ValidatePromoCodeRequest, tenantID *uuid.UUID, userID uuid.UUID) (*dto.PromoCodeValidationResponse, error)
	ApplyPromoCode(ctx context.Context, code string, tenantID *uuid.UUID, userID uuid.UUID, amount float64, serviceID *uuid.UUID, artisanID *uuid.UUID) (*dto.PromoCodeValidationResponse, error)
	IncrementUsage(ctx context.Context, promoCodeID uuid.UUID) error
	CanUserUsePromoCode(ctx context.Context, promoCodeID, userID uuid.UUID) (bool, error)

	// Status Operations
	ActivatePromoCode(ctx context.Context, id uuid.UUID) error
	DeactivatePromoCode(ctx context.Context, id uuid.UUID) error
	BulkActivate(ctx context.Context, ids []uuid.UUID) error
	BulkDeactivate(ctx context.Context, ids []uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error

	// Service/Artisan Specific
	GetValidPromoCodesForService(ctx context.Context, serviceID uuid.UUID) ([]*dto.PromoCodeResponse, error)
	GetValidPromoCodesForArtisan(ctx context.Context, artisanID uuid.UUID) ([]*dto.PromoCodeResponse, error)

	// Analytics & Statistics
	GetPromoCodeStats(ctx context.Context, promoCodeID uuid.UUID) (*dto.PromoCodeStatsResponse, error)
	GetTopPerformingPromoCodes(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]repository.PromoCodePerformance, error)
}

// promoCodeService implements PromoCodeService
type promoCodeService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewPromoCodeService creates a new PromoCodeService instance
func NewPromoCodeService(repos *repository.Repositories, logger log.AllLogger) PromoCodeService {
	return &promoCodeService{
		repos:  repos,
		logger: logger,
	}
}

// CreatePromoCode creates a new promo code
func (s *promoCodeService) CreatePromoCode(ctx context.Context, tenantID *uuid.UUID, req *dto.CreatePromoCodeRequest) (*dto.PromoCodeResponse, error) {
	s.logger.Info("creating promo code", "code", req.Code, "tenant_id", tenantID)

	// Validate code is uppercase
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, errors.NewValidationError("Promo code cannot be empty")
	}

	// Check if code already exists
	existing, err := s.repos.PromoCode.GetByCode(ctx, code)
	if err == nil && existing != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("Promo code '%s' already exists", code))
	}

	// Validate dates
	if req.ExpiresAt != nil && req.ExpiresAt.Before(req.StartsAt) {
		return nil, errors.NewValidationError("Expiration date must be after start date")
	}

	// Validate discount value
	if req.Type == models.DiscountTypePercentage && req.Value > 100 {
		return nil, errors.NewValidationError("Percentage discount cannot exceed 100%")
	}

	if req.Value <= 0 {
		return nil, errors.NewValidationError("Discount value must be greater than 0")
	}

	// Create promo code model
	promoCode := &models.PromoCode{
		TenantID:           tenantID,
		Code:               code,
		Description:        req.Description,
		Type:               req.Type,
		Value:              req.Value,
		MaxDiscount:        req.MaxDiscount,
		MinOrderAmount:     req.MinOrderAmount,
		StartsAt:           req.StartsAt,
		ExpiresAt:          req.ExpiresAt,
		MaxUses:            req.MaxUses,
		UsedCount:          0,
		MaxUsesPerUser:     req.MaxUsesPerUser,
		ApplicableServices: req.ApplicableServices,
		ApplicableArtisans: req.ApplicableArtisans,
		IsActive:           req.IsActive,
		Metadata:           req.Metadata,
	}

	// Save to database
	if err := s.repos.PromoCode.Create(ctx, promoCode); err != nil {
		s.logger.Error("failed to create promo code", "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "Failed to create promo code", err)
	}

	s.logger.Info("promo code created", "promo_code_id", promoCode.ID, "code", code)

	// Load with relationships
	created, err := s.repos.PromoCode.GetByID(ctx, promoCode.ID)
	if err != nil {
		s.logger.Error("failed to load promo code", "promo_code_id", promoCode.ID, "error", err)
		return dto.ToPromoCodeResponse(promoCode), nil
	}

	return dto.ToPromoCodeResponse(created), nil
}

// GetPromoCode retrieves a promo code by ID
func (s *promoCodeService) GetPromoCode(ctx context.Context, id uuid.UUID) (*dto.PromoCodeResponse, error) {
	s.logger.Info("retrieving promo code", "promo_code_id", id)

	promoCode, err := s.repos.PromoCode.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("promo code not found", "promo_code_id", id, "error", err)
		return nil, errors.NewNotFoundError("promo code")
	}

	return dto.ToPromoCodeResponse(promoCode), nil
}

// GetPromoCodeByCode retrieves a promo code by code
func (s *promoCodeService) GetPromoCodeByCode(ctx context.Context, code string) (*dto.PromoCodeResponse, error) {
	s.logger.Info("retrieving promo code by code", "code", code)

	if code == "" {
		return nil, errors.NewValidationError("Code cannot be empty")
	}

	promoCode, err := s.repos.PromoCode.GetByCode(ctx, code)
	if err != nil {
		s.logger.Error("promo code not found", "code", code, "error", err)
		return nil, errors.NewNotFoundError("promo code")
	}

	return dto.ToPromoCodeResponse(promoCode), nil
}

// UpdatePromoCode updates a promo code
func (s *promoCodeService) UpdatePromoCode(ctx context.Context, id uuid.UUID, req *dto.UpdatePromoCodeRequest) (*dto.PromoCodeResponse, error) {
	s.logger.Info("updating promo code", "promo_code_id", id)

	// Get existing promo code
	promoCode, err := s.repos.PromoCode.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("promo code not found", "promo_code_id", id, "error", err)
		return nil, errors.NewNotFoundError("promo code")
	}

	// Update fields
	if req.Description != nil {
		promoCode.Description = *req.Description
	}

	if req.MaxDiscount != nil {
		promoCode.MaxDiscount = *req.MaxDiscount
	}

	if req.MinOrderAmount != nil {
		promoCode.MinOrderAmount = *req.MinOrderAmount
	}

	if req.ExpiresAt != nil {
		promoCode.ExpiresAt = req.ExpiresAt
	}

	if req.MaxUses != nil {
		promoCode.MaxUses = *req.MaxUses
	}

	if req.MaxUsesPerUser != nil {
		promoCode.MaxUsesPerUser = *req.MaxUsesPerUser
	}

	if req.ApplicableServices != nil {
		promoCode.ApplicableServices = req.ApplicableServices
	}

	if req.ApplicableArtisans != nil {
		promoCode.ApplicableArtisans = req.ApplicableArtisans
	}

	if req.IsActive != nil {
		promoCode.IsActive = *req.IsActive
	}

	if req.Metadata != nil {
		if promoCode.Metadata == nil {
			promoCode.Metadata = make(models.JSONB)
		}
		maps.Copy(promoCode.Metadata, req.Metadata)
	}

	// Save changes
	if err := s.repos.PromoCode.Update(ctx, promoCode); err != nil {
		s.logger.Error("failed to update promo code", "promo_code_id", id, "error", err)
		return nil, errors.NewServiceError("UPDATE_FAILED", "Failed to update promo code", err)
	}

	s.logger.Info("promo code updated", "promo_code_id", id)
	return dto.ToPromoCodeResponse(promoCode), nil
}

// DeletePromoCode deletes a promo code
func (s *promoCodeService) DeletePromoCode(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("deleting promo code", "promo_code_id", id)

	// Verify promo code exists
	_, err := s.repos.PromoCode.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("promo code not found", "promo_code_id", id, "error", err)
		return errors.NewNotFoundError("promo code")
	}

	if err := s.repos.PromoCode.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete promo code", "promo_code_id", id, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete promo code", err)
	}

	s.logger.Info("promo code deleted", "promo_code_id", id)
	return nil
}

// ListPromoCodes retrieves promo codes with filters
func (s *promoCodeService) ListPromoCodes(ctx context.Context, filter *dto.PromoCodeFilter) (*dto.PromoCodeListResponse, error) {
	s.logger.Info("listing promo codes")

	// Build repository filters
	repoFilters := repository.PromoCodeFilters{
		TenantID: filter.TenantID,
		Type:     filter.Type,
		IsActive: filter.IsActive,
	}

	if filter.ServiceID != nil {
		repoFilters.ServiceIDs = []uuid.UUID{*filter.ServiceID}
	}

	if filter.ArtisanID != nil {
		repoFilters.ArtisanIDs = []uuid.UUID{*filter.ArtisanID}
	}

	// Set defaults
	page := max(1, filter.Page)

	pageSize := min(100, max(1, filter.PageSize))

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	promoCodes, paginationResult, err := s.repos.PromoCode.FindByFilters(ctx, repoFilters, pagination)
	if err != nil {
		s.logger.Error("failed to list promo codes", "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list promo codes", err)
	}

	return &dto.PromoCodeListResponse{
		PromoCodes:  dto.ToPromoCodeResponses(promoCodes),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetActivePromoCodes retrieves active promo codes
func (s *promoCodeService) GetActivePromoCodes(ctx context.Context, tenantID *uuid.UUID, page, pageSize int) (*dto.PromoCodeListResponse, error) {
	s.logger.Info("getting active promo codes", "tenant_id", tenantID)

	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	promoCodes, paginationResult, err := s.repos.PromoCode.GetActivePromoCodes(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to get active promo codes", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get active promo codes", err)
	}

	return &dto.PromoCodeListResponse{
		PromoCodes:  dto.ToPromoCodeResponses(promoCodes),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetExpiredPromoCodes retrieves expired promo codes
func (s *promoCodeService) GetExpiredPromoCodes(ctx context.Context, tenantID uuid.UUID) ([]*dto.PromoCodeResponse, error) {
	s.logger.Info("getting expired promo codes", "tenant_id", tenantID)

	promoCodes, err := s.repos.PromoCode.GetExpiredPromoCodes(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get expired promo codes", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get expired promo codes", err)
	}

	return dto.ToPromoCodeResponses(promoCodes), nil
}

// GetExpiringPromoCodes retrieves promo codes expiring soon
func (s *promoCodeService) GetExpiringPromoCodes(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.PromoCodeResponse, error) {
	s.logger.Info("getting expiring promo codes", "tenant_id", tenantID, "days", days)

	if days <= 0 {
		days = 7
	}

	promoCodes, err := s.repos.PromoCode.GetExpiringPromoCodes(ctx, tenantID, days)
	if err != nil {
		s.logger.Error("failed to get expiring promo codes", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get expiring promo codes", err)
	}

	return dto.ToPromoCodeResponses(promoCodes), nil
}

// SearchPromoCodes searches promo codes by code or description
func (s *promoCodeService) SearchPromoCodes(ctx context.Context, query string, tenantID *uuid.UUID, page, pageSize int) (*dto.PromoCodeListResponse, error) {
	s.logger.Info("searching promo codes", "query", query, "tenant_id", tenantID)

	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	promoCodes, paginationResult, err := s.repos.PromoCode.Search(ctx, query, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to search promo codes", "error", err)
		return nil, errors.NewServiceError("SEARCH_FAILED", "Failed to search promo codes", err)
	}

	return &dto.PromoCodeListResponse{
		PromoCodes:  dto.ToPromoCodeResponses(promoCodes),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ValidatePromoCode validates a promo code without applying it
func (s *promoCodeService) ValidatePromoCode(ctx context.Context, req *dto.ValidatePromoCodeRequest, tenantID *uuid.UUID, userID uuid.UUID) (*dto.PromoCodeValidationResponse, error) {
	s.logger.Info("validating promo code", "code", req.Code, "tenant_id", tenantID, "user_id", userID)

	return s.ApplyPromoCode(ctx, req.Code, tenantID, userID, req.Amount, req.ServiceID, req.ArtisanID)
}

// ApplyPromoCode validates and calculates discount for a promo code
func (s *promoCodeService) ApplyPromoCode(ctx context.Context, code string, tenantID *uuid.UUID, userID uuid.UUID, amount float64, serviceID *uuid.UUID, artisanID *uuid.UUID) (*dto.PromoCodeValidationResponse, error) {
	s.logger.Info("applying promo code", "code", code, "amount", amount)

	if code == "" {
		return nil, errors.NewValidationError("Promo code cannot be empty")
	}

	if amount <= 0 {
		return nil, errors.NewValidationError("Amount must be greater than 0")
	}

	// Validate promo code
	promoCode, discountAmount, err := s.repos.PromoCode.ValidateCode(ctx, code, tenantID, amount, serviceID, artisanID)
	if err != nil {
		s.logger.Error("promo code validation failed", "code", code, "error", err)
		return &dto.PromoCodeValidationResponse{
			IsValid:        false,
			PromoCode:      nil,
			DiscountAmount: 0,
			FinalAmount:    amount,
			Message:        err.Error(),
		}, nil
	}

	// Check if user can use this promo code
	canUse, err := s.repos.PromoCode.CanUserUsePromoCode(ctx, promoCode.ID, userID)
	if err != nil {
		s.logger.Error("failed to check user promo code usage", "promo_code_id", promoCode.ID, "user_id", userID, "error", err)
		return nil, errors.NewServiceError("CHECK_FAILED", "Failed to check promo code usage", err)
	}

	if !canUse {
		return &dto.PromoCodeValidationResponse{
			IsValid:        false,
			PromoCode:      nil,
			DiscountAmount: 0,
			FinalAmount:    amount,
			Message:        "You have already used this promo code the maximum number of times",
		}, nil
	}

	finalAmount := amount - discountAmount
	if finalAmount < 0 {
		finalAmount = 0
	}

	return &dto.PromoCodeValidationResponse{
		IsValid:        true,
		PromoCode:      dto.ToPromoCodeResponse(promoCode),
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
		Message:        fmt.Sprintf("Promo code applied successfully! You save %.2f", discountAmount),
	}, nil
}

// IncrementUsage increments the usage count of a promo code
func (s *promoCodeService) IncrementUsage(ctx context.Context, promoCodeID uuid.UUID) error {
	s.logger.Info("incrementing promo code usage", "promo_code_id", promoCodeID)

	if err := s.repos.PromoCode.IncrementUsage(ctx, promoCodeID); err != nil {
		s.logger.Error("failed to increment usage", "promo_code_id", promoCodeID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to increment usage", err)
	}

	return nil
}

// CanUserUsePromoCode checks if a user can use a promo code
func (s *promoCodeService) CanUserUsePromoCode(ctx context.Context, promoCodeID, userID uuid.UUID) (bool, error) {
	s.logger.Info("checking if user can use promo code", "promo_code_id", promoCodeID, "user_id", userID)

	canUse, err := s.repos.PromoCode.CanUserUsePromoCode(ctx, promoCodeID, userID)
	if err != nil {
		s.logger.Error("failed to check user promo code usage", "error", err)
		return false, errors.NewServiceError("CHECK_FAILED", "Failed to check promo code usage", err)
	}

	return canUse, nil
}

// ActivatePromoCode activates a promo code
func (s *promoCodeService) ActivatePromoCode(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("activating promo code", "promo_code_id", id)

	if err := s.repos.PromoCode.Activate(ctx, id); err != nil {
		s.logger.Error("failed to activate promo code", "promo_code_id", id, "error", err)
		return errors.NewServiceError("ACTIVATE_FAILED", "Failed to activate promo code", err)
	}

	s.logger.Info("promo code activated", "promo_code_id", id)
	return nil
}

// DeactivatePromoCode deactivates a promo code
func (s *promoCodeService) DeactivatePromoCode(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("deactivating promo code", "promo_code_id", id)

	if err := s.repos.PromoCode.Deactivate(ctx, id); err != nil {
		s.logger.Error("failed to deactivate promo code", "promo_code_id", id, "error", err)
		return errors.NewServiceError("DEACTIVATE_FAILED", "Failed to deactivate promo code", err)
	}

	s.logger.Info("promo code deactivated", "promo_code_id", id)
	return nil
}

// BulkActivate activates multiple promo codes
func (s *promoCodeService) BulkActivate(ctx context.Context, ids []uuid.UUID) error {
	s.logger.Info("bulk activating promo codes", "count", len(ids))

	if len(ids) == 0 {
		return errors.NewValidationError("No promo code IDs provided")
	}

	if err := s.repos.PromoCode.BulkActivate(ctx, ids); err != nil {
		s.logger.Error("failed to bulk activate promo codes", "error", err)
		return errors.NewServiceError("BULK_ACTIVATE_FAILED", "Failed to bulk activate promo codes", err)
	}

	s.logger.Info("promo codes bulk activated", "count", len(ids))
	return nil
}

// BulkDeactivate deactivates multiple promo codes
func (s *promoCodeService) BulkDeactivate(ctx context.Context, ids []uuid.UUID) error {
	s.logger.Info("bulk deactivating promo codes", "count", len(ids))

	if len(ids) == 0 {
		return errors.NewValidationError("No promo code IDs provided")
	}

	if err := s.repos.PromoCode.BulkDeactivate(ctx, ids); err != nil {
		s.logger.Error("failed to bulk deactivate promo codes", "error", err)
		return errors.NewServiceError("BULK_DEACTIVATE_FAILED", "Failed to bulk deactivate promo codes", err)
	}

	s.logger.Info("promo codes bulk deactivated", "count", len(ids))
	return nil
}

// BulkDelete deletes multiple promo codes
func (s *promoCodeService) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	s.logger.Info("bulk deleting promo codes", "count", len(ids))

	if len(ids) == 0 {
		return errors.NewValidationError("No promo code IDs provided")
	}

	if err := s.repos.PromoCode.BulkDelete(ctx, ids); err != nil {
		s.logger.Error("failed to bulk delete promo codes", "error", err)
		return errors.NewServiceError("BULK_DELETE_FAILED", "Failed to bulk delete promo codes", err)
	}

	s.logger.Info("promo codes bulk deleted", "count", len(ids))
	return nil
}

// GetValidPromoCodesForService retrieves valid promo codes for a service
func (s *promoCodeService) GetValidPromoCodesForService(ctx context.Context, serviceID uuid.UUID) ([]*dto.PromoCodeResponse, error) {
	s.logger.Info("getting valid promo codes for service", "service_id", serviceID)

	promoCodes, err := s.repos.PromoCode.GetValidPromoCodesForService(ctx, serviceID)
	if err != nil {
		s.logger.Error("failed to get valid promo codes for service", "service_id", serviceID, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get valid promo codes", err)
	}

	return dto.ToPromoCodeResponses(promoCodes), nil
}

// GetValidPromoCodesForArtisan retrieves valid promo codes for an artisan
func (s *promoCodeService) GetValidPromoCodesForArtisan(ctx context.Context, artisanID uuid.UUID) ([]*dto.PromoCodeResponse, error) {
	s.logger.Info("getting valid promo codes for artisan", "artisan_id", artisanID)

	promoCodes, err := s.repos.PromoCode.GetValidPromoCodesForArtisan(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get valid promo codes for artisan", "artisan_id", artisanID, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get valid promo codes", err)
	}

	return dto.ToPromoCodeResponses(promoCodes), nil
}

// GetPromoCodeStats retrieves statistics for a promo code
func (s *promoCodeService) GetPromoCodeStats(ctx context.Context, promoCodeID uuid.UUID) (*dto.PromoCodeStatsResponse, error) {
	s.logger.Info("getting promo code stats", "promo_code_id", promoCodeID)

	stats, err := s.repos.PromoCode.GetPromoCodeStats(ctx, promoCodeID)
	if err != nil {
		s.logger.Error("failed to get promo code stats", "promo_code_id", promoCodeID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get promo code stats", err)
	}

	return &dto.PromoCodeStatsResponse{
		PromoCodeID:     stats.PromoCodeID,
		Code:            stats.Code,
		TotalUses:       stats.TotalUses,
		UniqueUsers:     stats.UniqueUsers,
		TotalDiscount:   stats.TotalDiscount,
		AverageDiscount: stats.AverageDiscount,
		TotalRevenue:    stats.TotalRevenue,
		ConversionRate:  stats.ConversionRate,
		RemainingUses:   stats.RemainingUses,
		DaysUntilExpiry: stats.DaysUntilExpiry,
	}, nil
}

// GetTopPerformingPromoCodes retrieves top performing promo codes
func (s *promoCodeService) GetTopPerformingPromoCodes(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]repository.PromoCodePerformance, error) {
	s.logger.Info("getting top performing promo codes", "tenant_id", tenantID, "limit", limit)

	if limit <= 0 {
		limit = 10
	}

	performance, err := s.repos.PromoCode.GetTopPerformingPromoCodes(ctx, tenantID, limit, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get top performing promo codes", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get top performing promo codes", err)
	}

	return performance, nil
}
