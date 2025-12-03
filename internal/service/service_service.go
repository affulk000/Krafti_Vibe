package service

import (
	"context"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/repository/types"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// ServiceService handles business logic for services
type ServiceService interface {
	// CRUD Operations
	CreateService(ctx context.Context, req *dto.CreateServiceRequest) (*dto.ServiceResponse, error)
	GetServiceByID(ctx context.Context, serviceID uuid.UUID) (*dto.ServiceResponse, error)
	UpdateService(ctx context.Context, serviceID uuid.UUID, req *dto.UpdateServiceRequest) (*dto.ServiceResponse, error)
	DeleteService(ctx context.Context, serviceID uuid.UUID) error

	// Query Operations
	ListServices(ctx context.Context, req *dto.ListServicesRequest) (*dto.ListServicesResponse, error)
	ListServicesByCategory(ctx context.Context, tenantID uuid.UUID, category models.ServiceCategory, page, pageSize int) (*dto.ListServicesResponse, error)
	ListServicesByArtisan(ctx context.Context, artisanID uuid.UUID) ([]*dto.ServiceResponse, error)
	ListActiveServices(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.ListServicesResponse, error)
	ListOrganizationServices(ctx context.Context, tenantID uuid.UUID) ([]*dto.ServiceResponse, error)

	// Search & Discovery
	SearchServices(ctx context.Context, tenantID uuid.UUID, query string, page, pageSize int) (*dto.ListServicesResponse, error)
	FindServicesByTags(ctx context.Context, tenantID uuid.UUID, tags []string, page, pageSize int) (*dto.ListServicesResponse, error)
	FindServicesByPriceRange(ctx context.Context, tenantID uuid.UUID, minPrice, maxPrice float64, page, pageSize int) (*dto.ListServicesResponse, error)
	GetPopularServices(ctx context.Context, tenantID uuid.UUID, limit int) ([]*dto.ServiceResponse, error)
	GetRecommendedServices(ctx context.Context, tenantID, customerID uuid.UUID, limit int) ([]*dto.ServiceResponse, error)

	// Availability Management
	ActivateService(ctx context.Context, serviceID uuid.UUID) error
	DeactivateService(ctx context.Context, serviceID uuid.UUID) error
	ToggleServiceAvailability(ctx context.Context, serviceID uuid.UUID, isActive bool) error

	// Pricing Management
	UpdateServicePrice(ctx context.Context, serviceID uuid.UUID, newPrice float64) error
	UpdateServiceDeposit(ctx context.Context, serviceID uuid.UUID, depositAmount float64, requiresDeposit bool) error
	BulkUpdatePrices(ctx context.Context, req *dto.BulkPriceUpdateRequest) error

	// Addon Management
	AddServiceAddon(ctx context.Context, serviceID, addonID uuid.UUID) error
	RemoveServiceAddon(ctx context.Context, serviceID, addonID uuid.UUID) error
	GetServiceAddons(ctx context.Context, serviceID uuid.UUID) ([]*dto.ServiceAddonResponse, error)

	// Analytics & Statistics
	GetServiceStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.ServiceStatistics, error)
	GetCategoryStatistics(ctx context.Context, tenantID uuid.UUID) ([]*dto.CategoryStatisticsResponse, error)
	GetServicePerformance(ctx context.Context, serviceID uuid.UUID) (*dto.ServicePerformanceResponse, error)
	GetServiceRevenue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*dto.ServiceRevenueResponse, error)
	GetServiceTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.ServiceTrendResponse, error)

	// Bulk Operations
	BulkActivateServices(ctx context.Context, serviceIDs []uuid.UUID) error
	BulkDeactivateServices(ctx context.Context, serviceIDs []uuid.UUID) error
	BulkUpdateCategory(ctx context.Context, serviceIDs []uuid.UUID, category models.ServiceCategory) error
	BulkDeleteServices(ctx context.Context, serviceIDs []uuid.UUID) error

	// Advanced Filtering
	FilterServices(ctx context.Context, tenantID uuid.UUID, filters *dto.ServiceFilterRequest) (*dto.ListServicesResponse, error)
	GetCategoriesWithCount(ctx context.Context, tenantID uuid.UUID) ([]*dto.CategoryCountResponse, error)
}

// serviceService implements the ServiceService interface
type serviceService struct {
	serviceRepo repository.ServiceRepository
	tenantRepo  repository.TenantRepository
	userRepo    repository.UserRepository
	logger      log.AllLogger
}

// NewServiceService creates a new service service instance
func NewServiceService(
	serviceRepo repository.ServiceRepository,
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
	logger log.AllLogger,
) ServiceService {
	return &serviceService{
		serviceRepo: serviceRepo,
		tenantRepo:  tenantRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// CreateService creates a new service
func (s *serviceService) CreateService(ctx context.Context, req *dto.CreateServiceRequest) (*dto.ServiceResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid service data")
	}

	// Verify tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		s.logger.Error("failed to find tenant", "tenant_id", req.TenantID, "error", err)
		return nil, errors.NewNotFoundError("tenant not found")
	}

	// If artisan-specific, verify artisan exists and belongs to tenant
	if req.ArtisanID != nil {
		artisan, err := s.userRepo.GetByID(ctx, *req.ArtisanID)
		if err != nil {
			s.logger.Error("failed to find artisan", "artisan_id", *req.ArtisanID, "error", err)
			return nil, errors.NewNotFoundError("artisan not found")
		}

		// Verify artisan belongs to the tenant
		if (*artisan).TenantID == nil || *(*artisan).TenantID != (*tenant).ID {
			return nil, errors.NewForbiddenError("artisan does not belong to this tenant")
		}

		// Verify user has artisan role
		if (*artisan).Role != models.UserRoleArtisan {
			return nil, errors.NewValidationError("user is not an artisan")
		}
	}

	// Create service model
	service := &models.Service{
		TenantID:        req.TenantID,
		ArtisanID:       req.ArtisanID,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		Price:           req.Price,
		Currency:        req.Currency,
		DepositAmount:   req.DepositAmount,
		DurationMinutes: req.DurationMinutes,
		BufferMinutes:   req.BufferMinutes,
		IsActive:        req.IsActive,
		MaxBookingsDay:  req.MaxBookingsDay,
		ImageURL:        req.ImageURL,
		RequiresDeposit: req.RequiresDeposit,
		Tags:            req.Tags,
		Metadata:        req.Metadata,
	}

	// Create service
	if err := s.serviceRepo.Create(ctx, service); err != nil {
		s.logger.Error("failed to create service", "error", err)
		return nil, errors.NewInternalError("failed to create service", err)
	}

	s.logger.Info("service created successfully", "service_id", service.ID, "tenant_id", req.TenantID)

	return s.toServiceResponse(service), nil
}

// GetServiceByID retrieves a service by ID
func (s *serviceService) GetServiceByID(ctx context.Context, serviceID uuid.UUID) (*dto.ServiceResponse, error) {
	if serviceID == uuid.Nil {
		return nil, errors.NewValidationError("service ID is required")
	}

	service, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		s.logger.Error("failed to find service", "service_id", serviceID, "error", err)
		return nil, errors.NewNotFoundError("service not found")
	}

	return s.toServiceResponse(service), nil
}

// UpdateService updates an existing service
func (s *serviceService) UpdateService(ctx context.Context, serviceID uuid.UUID, req *dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
	if serviceID == uuid.Nil {
		return nil, errors.NewValidationError("service ID is required")
	}

	// Get existing service
	service, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		s.logger.Error("failed to find service", "service_id", serviceID, "error", err)
		return nil, errors.NewNotFoundError("service not found")
	}

	// Update fields if provided
	if req.Name != nil {
		(*service).Name = *req.Name
	}
	if req.Description != nil {
		(*service).Description = *req.Description
	}
	if req.Category != nil {
		(*service).Category = *req.Category
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, errors.NewValidationError("price cannot be negative")
		}
		(*service).Price = *req.Price
	}
	if req.Currency != nil {
		(*service).Currency = *req.Currency
	}
	if req.DepositAmount != nil {
		if *req.DepositAmount < 0 {
			return nil, errors.NewValidationError("deposit amount cannot be negative")
		}
		(*service).DepositAmount = *req.DepositAmount
	}
	if req.DurationMinutes != nil {
		if *req.DurationMinutes < 5 {
			return nil, errors.NewValidationError("duration must be at least 5 minutes")
		}
		(*service).DurationMinutes = *req.DurationMinutes
	}
	if req.BufferMinutes != nil {
		(*service).BufferMinutes = *req.BufferMinutes
	}
	if req.IsActive != nil {
		(*service).IsActive = *req.IsActive
	}
	if req.MaxBookingsDay != nil {
		(*service).MaxBookingsDay = *req.MaxBookingsDay
	}
	if req.ImageURL != nil {
		(*service).ImageURL = *req.ImageURL
	}
	if req.RequiresDeposit != nil {
		(*service).RequiresDeposit = *req.RequiresDeposit
	}
	if req.Tags != nil {
		(*service).Tags = req.Tags
	}
	if req.Metadata != nil {
		(*service).Metadata = req.Metadata
	}

	// Update service
	if err := s.serviceRepo.Update(ctx, service); err != nil {
		s.logger.Error("failed to update service", "service_id", serviceID, "error", err)
		return nil, errors.NewInternalError("failed to update service", err)
	}

	s.logger.Info("service updated successfully", "service_id", serviceID)

	return s.toServiceResponse(service), nil
}

// DeleteService deletes a service
func (s *serviceService) DeleteService(ctx context.Context, serviceID uuid.UUID) error {
	if serviceID == uuid.Nil {
		return errors.NewValidationError("service ID is required")
	}

	// Check if service exists
	service, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return errors.NewNotFoundError("service not found")
	}

	// Soft delete
	if err := s.serviceRepo.Delete(ctx, serviceID); err != nil {
		s.logger.Error("failed to delete service", "service_id", serviceID, "error", err)
		return errors.NewInternalError("failed to delete service", err)
	}

	s.logger.Info("service deleted successfully", "service_id", serviceID, "tenant_id", (*service).TenantID)

	return nil
}

// ListServices lists services with pagination
func (s *serviceService) ListServices(ctx context.Context, req *dto.ListServicesRequest) (*dto.ListServicesResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request")
	}

	pagination := repository.PaginationParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	services, paginationResult, err := s.serviceRepo.FindByTenantID(ctx, req.TenantID, pagination)
	if err != nil {
		s.logger.Error("failed to list services", "tenant_id", req.TenantID, "error", err)
		return nil, errors.NewInternalError("failed to list services", err)
	}

	return &dto.ListServicesResponse{
		Services:   s.toServiceResponseList(services),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalItems: paginationResult.TotalItems,
		TotalPages: paginationResult.TotalPages,
	}, nil
}

// ListServicesByCategory lists services by category
func (s *serviceService) ListServicesByCategory(ctx context.Context, tenantID uuid.UUID, category models.ServiceCategory, page, pageSize int) (*dto.ListServicesResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	services, paginationResult, err := s.serviceRepo.FindByCategory(ctx, tenantID, category, pagination)
	if err != nil {
		s.logger.Error("failed to list services by category", "tenant_id", tenantID, "category", category, "error", err)
		return nil, errors.NewInternalError("failed to list services", err)
	}

	return &dto.ListServicesResponse{
		Services:   s.toServiceResponseList(services),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalItems: paginationResult.TotalItems,
		TotalPages: paginationResult.TotalPages,
	}, nil
}

// ListServicesByArtisan lists services by artisan
func (s *serviceService) ListServicesByArtisan(ctx context.Context, artisanID uuid.UUID) ([]*dto.ServiceResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan ID is required")
	}

	services, err := s.serviceRepo.FindByArtisanID(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to list services by artisan", "artisan_id", artisanID, "error", err)
		return nil, errors.NewInternalError("failed to list services", err)
	}

	return s.toServiceResponseList(services), nil
}

// ListActiveServices lists only active services
func (s *serviceService) ListActiveServices(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.ListServicesResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	services, paginationResult, err := s.serviceRepo.FindActiveServices(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to list active services", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to list active services", err)
	}

	return &dto.ListServicesResponse{
		Services:   s.toServiceResponseList(services),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalItems: paginationResult.TotalItems,
		TotalPages: paginationResult.TotalPages,
	}, nil
}

// ListOrganizationServices lists organization-wide services
func (s *serviceService) ListOrganizationServices(ctx context.Context, tenantID uuid.UUID) ([]*dto.ServiceResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	services, err := s.serviceRepo.FindOrganizationWideServices(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to list organization services", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to list organization services", err)
	}

	return s.toServiceResponseList(services), nil
}

// SearchServices searches for services
func (s *serviceService) SearchServices(ctx context.Context, tenantID uuid.UUID, query string, page, pageSize int) (*dto.ListServicesResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	services, paginationResult, err := s.serviceRepo.Search(ctx, tenantID, query, pagination)
	if err != nil {
		s.logger.Error("failed to search services", "tenant_id", tenantID, "query", query, "error", err)
		return nil, errors.NewInternalError("failed to search services", err)
	}

	return &dto.ListServicesResponse{
		Services:   s.toServiceResponseList(services),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalItems: paginationResult.TotalItems,
		TotalPages: paginationResult.TotalPages,
	}, nil
}

// FindServicesByTags finds services by tags
func (s *serviceService) FindServicesByTags(ctx context.Context, tenantID uuid.UUID, tags []string, page, pageSize int) (*dto.ListServicesResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	if len(tags) == 0 {
		return nil, errors.NewValidationError("at least one tag is required")
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	services, paginationResult, err := s.serviceRepo.FindByTags(ctx, tenantID, tags, pagination)
	if err != nil {
		s.logger.Error("failed to find services by tags", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to find services", err)
	}

	return &dto.ListServicesResponse{
		Services:   s.toServiceResponseList(services),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalItems: paginationResult.TotalItems,
		TotalPages: paginationResult.TotalPages,
	}, nil
}

// FindServicesByPriceRange finds services within a price range
func (s *serviceService) FindServicesByPriceRange(ctx context.Context, tenantID uuid.UUID, minPrice, maxPrice float64, page, pageSize int) (*dto.ListServicesResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	if minPrice < 0 || maxPrice < 0 {
		return nil, errors.NewValidationError("prices cannot be negative")
	}

	if minPrice > maxPrice {
		return nil, errors.NewValidationError("minimum price cannot exceed maximum price")
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	services, paginationResult, err := s.serviceRepo.FindByPriceRange(ctx, tenantID, minPrice, maxPrice, pagination)
	if err != nil {
		s.logger.Error("failed to find services by price range", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to find services", err)
	}

	return &dto.ListServicesResponse{
		Services:   s.toServiceResponseList(services),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalItems: paginationResult.TotalItems,
		TotalPages: paginationResult.TotalPages,
	}, nil
}

// GetPopularServices gets popular services
func (s *serviceService) GetPopularServices(ctx context.Context, tenantID uuid.UUID, limit int) ([]*dto.ServiceResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	if limit <= 0 {
		limit = 10
	}

	services, err := s.serviceRepo.FindPopularServices(ctx, tenantID, limit)
	if err != nil {
		s.logger.Error("failed to get popular services", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to get popular services", err)
	}

	return s.toServiceResponseList(services), nil
}

// GetRecommendedServices gets recommended services for a customer
func (s *serviceService) GetRecommendedServices(ctx context.Context, tenantID, customerID uuid.UUID, limit int) ([]*dto.ServiceResponse, error) {
	if tenantID == uuid.Nil || customerID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID and customer ID are required")
	}

	if limit <= 0 {
		limit = 10
	}

	services, err := s.serviceRepo.FindRecommendedServices(ctx, tenantID, customerID, limit)
	if err != nil {
		s.logger.Error("failed to get recommended services", "tenant_id", tenantID, "customer_id", customerID, "error", err)
		return nil, errors.NewInternalError("failed to get recommended services", err)
	}

	return s.toServiceResponseList(services), nil
}

// ActivateService activates a service
func (s *serviceService) ActivateService(ctx context.Context, serviceID uuid.UUID) error {
	if serviceID == uuid.Nil {
		return errors.NewValidationError("service ID is required")
	}

	if err := s.serviceRepo.ActivateService(ctx, serviceID); err != nil {
		s.logger.Error("failed to activate service", "service_id", serviceID, "error", err)
		return errors.NewInternalError("failed to activate service", err)
	}

	s.logger.Info("service activated", "service_id", serviceID)
	return nil
}

// DeactivateService deactivates a service
func (s *serviceService) DeactivateService(ctx context.Context, serviceID uuid.UUID) error {
	if serviceID == uuid.Nil {
		return errors.NewValidationError("service ID is required")
	}

	if err := s.serviceRepo.DeactivateService(ctx, serviceID); err != nil {
		s.logger.Error("failed to deactivate service", "service_id", serviceID, "error", err)
		return errors.NewInternalError("failed to deactivate service", err)
	}

	s.logger.Info("service deactivated", "service_id", serviceID)
	return nil
}

// ToggleServiceAvailability toggles service availability
func (s *serviceService) ToggleServiceAvailability(ctx context.Context, serviceID uuid.UUID, isActive bool) error {
	if serviceID == uuid.Nil {
		return errors.NewValidationError("service ID is required")
	}

	if err := s.serviceRepo.UpdateAvailability(ctx, serviceID, isActive); err != nil {
		s.logger.Error("failed to toggle service availability", "service_id", serviceID, "error", err)
		return errors.NewInternalError("failed to toggle service availability", err)
	}

	s.logger.Info("service availability toggled", "service_id", serviceID, "is_active", isActive)
	return nil
}

// UpdateServicePrice updates service price
func (s *serviceService) UpdateServicePrice(ctx context.Context, serviceID uuid.UUID, newPrice float64) error {
	if serviceID == uuid.Nil {
		return errors.NewValidationError("service ID is required")
	}

	if newPrice < 0 {
		return errors.NewValidationError("price cannot be negative")
	}

	if err := s.serviceRepo.UpdatePrice(ctx, serviceID, newPrice); err != nil {
		s.logger.Error("failed to update service price", "service_id", serviceID, "error", err)
		return errors.NewInternalError("failed to update service price", err)
	}

	s.logger.Info("service price updated", "service_id", serviceID, "new_price", newPrice)
	return nil
}

// UpdateServiceDeposit updates service deposit
func (s *serviceService) UpdateServiceDeposit(ctx context.Context, serviceID uuid.UUID, depositAmount float64, requiresDeposit bool) error {
	if serviceID == uuid.Nil {
		return errors.NewValidationError("service ID is required")
	}

	if depositAmount < 0 {
		return errors.NewValidationError("deposit amount cannot be negative")
	}

	if err := s.serviceRepo.UpdateDeposit(ctx, serviceID, depositAmount); err != nil {
		s.logger.Error("failed to update service deposit", "service_id", serviceID, "error", err)
		return errors.NewInternalError("failed to update service deposit", err)
	}

	s.logger.Info("service deposit updated", "service_id", serviceID, "deposit_amount", depositAmount)
	return nil
}

// BulkUpdatePrices updates prices for multiple services
func (s *serviceService) BulkUpdatePrices(ctx context.Context, req *dto.BulkPriceUpdateRequest) error {
	if err := req.Validate(); err != nil {
		return errors.NewValidationError("invalid request")
	}

	if err := s.serviceRepo.BulkUpdatePrices(ctx, req.ServiceIDs, req.Adjustment, req.IsPercentage); err != nil {
		s.logger.Error("failed to bulk update prices", "error", err)
		return errors.NewInternalError("failed to bulk update prices", err)
	}

	s.logger.Info("bulk price update completed", "service_count", len(req.ServiceIDs))
	return nil
}

// AddServiceAddon adds an addon to a service
func (s *serviceService) AddServiceAddon(ctx context.Context, serviceID, addonID uuid.UUID) error {
	if serviceID == uuid.Nil || addonID == uuid.Nil {
		return errors.NewValidationError("service ID and addon ID are required")
	}

	if err := s.serviceRepo.AddServiceAddon(ctx, serviceID, addonID); err != nil {
		s.logger.Error("failed to add service addon", "service_id", serviceID, "addon_id", addonID, "error", err)
		return errors.NewInternalError("failed to add service addon", err)
	}

	s.logger.Info("service addon added", "service_id", serviceID, "addon_id", addonID)
	return nil
}

// RemoveServiceAddon removes an addon from a service
func (s *serviceService) RemoveServiceAddon(ctx context.Context, serviceID, addonID uuid.UUID) error {
	if serviceID == uuid.Nil || addonID == uuid.Nil {
		return errors.NewValidationError("service ID and addon ID are required")
	}

	if err := s.serviceRepo.RemoveServiceAddon(ctx, serviceID, addonID); err != nil {
		s.logger.Error("failed to remove service addon", "service_id", serviceID, "addon_id", addonID, "error", err)
		return errors.NewInternalError("failed to remove service addon", err)
	}

	s.logger.Info("service addon removed", "service_id", serviceID, "addon_id", addonID)
	return nil
}

// GetServiceAddons gets all addons for a service
func (s *serviceService) GetServiceAddons(ctx context.Context, serviceID uuid.UUID) ([]*dto.ServiceAddonResponse, error) {
	if serviceID == uuid.Nil {
		return nil, errors.NewValidationError("service ID is required")
	}

	addons, err := s.serviceRepo.GetServiceAddons(ctx, serviceID)
	if err != nil {
		s.logger.Error("failed to get service addons", "service_id", serviceID, "error", err)
		return nil, errors.NewInternalError("failed to get service addons", err)
	}

	responses := make([]*dto.ServiceAddonResponse, len(addons))
	for i, addon := range addons {
		responses[i] = &dto.ServiceAddonResponse{
			ID:          addon.ID,
			TenantID:    addon.TenantID,
			Name:        addon.Name,
			Description: addon.Description,
			Price:       addon.Price,
			IsActive:    addon.IsActive,
			CreatedAt:   addon.CreatedAt,
			UpdatedAt:   addon.UpdatedAt,
		}
	}

	return responses, nil
}

// GetServiceStatistics gets service statistics for a tenant
func (s *serviceService) GetServiceStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.ServiceStatistics, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	stats, err := s.serviceRepo.GetServiceStats(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get service statistics", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to get service statistics", err)
	}

	return &dto.ServiceStatistics{
		TotalServices:        stats.TotalServices,
		ActiveServices:       stats.ActiveServices,
		InactiveServices:     stats.InactiveServices,
		ByCategory:           stats.ByCategory,
		AverageDuration:      stats.AverageDuration,
		AveragePrice:         stats.AveragePrice,
		TotalBookings:        stats.TotalBookings,
		TotalRevenue:         stats.TotalRevenue,
		ServicesWithDeposit:  stats.ServicesWithDeposit,
		MostPopularCategory:  stats.MostPopularCategory,
		HighestPricedService: stats.HighestPricedService,
		MostBookedService:    stats.MostBookedService,
	}, nil
}

// GetCategoryStatistics gets statistics per category
func (s *serviceService) GetCategoryStatistics(ctx context.Context, tenantID uuid.UUID) ([]*dto.CategoryStatisticsResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	stats, err := s.serviceRepo.GetCategoryStats(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get category statistics", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to get category statistics", err)
	}

	responses := make([]*dto.CategoryStatisticsResponse, len(stats))
	for i, stat := range stats {
		responses[i] = &dto.CategoryStatisticsResponse{
			Category:      stat.Category,
			ServiceCount:  stat.ServiceCount,
			TotalBookings: stat.TotalBookings,
			TotalRevenue:  stat.TotalRevenue,
			AveragePrice:  stat.AveragePrice,
			ActiveCount:   stat.ActiveCount,
		}
	}

	return responses, nil
}

// GetServicePerformance gets performance metrics for a service
func (s *serviceService) GetServicePerformance(ctx context.Context, serviceID uuid.UUID) (*dto.ServicePerformanceResponse, error) {
	if serviceID == uuid.Nil {
		return nil, errors.NewValidationError("service ID is required")
	}

	perf, err := s.serviceRepo.GetServicePerformance(ctx, serviceID)
	if err != nil {
		s.logger.Error("failed to get service performance", "service_id", serviceID, "error", err)
		return nil, errors.NewInternalError("failed to get service performance", err)
	}

	return &dto.ServicePerformanceResponse{
		ServiceID:         perf.ServiceID,
		ServiceName:       perf.ServiceName,
		TotalBookings:     perf.TotalBookings,
		CompletedBookings: perf.CompletedBookings,
		CancelledBookings: perf.CancelledBookings,
		TotalRevenue:      perf.TotalRevenue,
		AverageRating:     perf.AverageRating,
		ReviewCount:       perf.ReviewCount,
		CompletionRate:    perf.CompletionRate,
		CancellationRate:  perf.CancellationRate,
		BookingsThisMonth: perf.BookingsThisMonth,
		RevenueThisMonth:  perf.RevenueThisMonth,
		PopularityScore:   perf.PopularityScore,
	}, nil
}

// GetServiceRevenue gets revenue by service
func (s *serviceService) GetServiceRevenue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*dto.ServiceRevenueResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	if startDate.After(endDate) {
		return nil, errors.NewValidationError("start date must be before end date")
	}

	revenues, err := s.serviceRepo.GetRevenueByService(ctx, tenantID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get service revenue", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to get service revenue", err)
	}

	responses := make([]*dto.ServiceRevenueResponse, len(revenues))
	for i, rev := range revenues {
		responses[i] = &dto.ServiceRevenueResponse{
			ServiceID:    rev.ServiceID,
			ServiceName:  rev.ServiceName,
			Category:     rev.Category,
			Bookings:     rev.Bookings,
			Revenue:      rev.Revenue,
			AveragePrice: rev.AveragePrice,
		}
	}

	return responses, nil
}

// GetServiceTrends gets booking trends
func (s *serviceService) GetServiceTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.ServiceTrendResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	if days <= 0 {
		days = 30
	}

	trends, err := s.serviceRepo.GetServiceBookingTrends(ctx, tenantID, days)
	if err != nil {
		s.logger.Error("failed to get service trends", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to get service trends", err)
	}

	responses := make([]*dto.ServiceTrendResponse, len(trends))
	for i, trend := range trends {
		responses[i] = &dto.ServiceTrendResponse{
			Date:        trend.Date,
			ServiceID:   trend.ServiceID,
			ServiceName: trend.ServiceName,
			Bookings:    trend.Bookings,
			Revenue:     trend.Revenue,
		}
	}

	return responses, nil
}

// BulkActivateServices activates multiple services
func (s *serviceService) BulkActivateServices(ctx context.Context, serviceIDs []uuid.UUID) error {
	if len(serviceIDs) == 0 {
		return errors.NewValidationError("at least one service ID is required")
	}

	if err := s.serviceRepo.BulkActivate(ctx, serviceIDs); err != nil {
		s.logger.Error("failed to bulk activate services", "error", err)
		return errors.NewInternalError("failed to bulk activate services", err)
	}

	s.logger.Info("bulk activate completed", "service_count", len(serviceIDs))
	return nil
}

// BulkDeactivateServices deactivates multiple services
func (s *serviceService) BulkDeactivateServices(ctx context.Context, serviceIDs []uuid.UUID) error {
	if len(serviceIDs) == 0 {
		return errors.NewValidationError("at least one service ID is required")
	}

	if err := s.serviceRepo.BulkDeactivate(ctx, serviceIDs); err != nil {
		s.logger.Error("failed to bulk deactivate services", "error", err)
		return errors.NewInternalError("failed to bulk deactivate services", err)
	}

	s.logger.Info("bulk deactivate completed", "service_count", len(serviceIDs))
	return nil
}

// BulkUpdateCategory updates category for multiple services
func (s *serviceService) BulkUpdateCategory(ctx context.Context, serviceIDs []uuid.UUID, category models.ServiceCategory) error {
	if len(serviceIDs) == 0 {
		return errors.NewValidationError("at least one service ID is required")
	}

	if err := s.serviceRepo.BulkUpdateCategory(ctx, serviceIDs, category); err != nil {
		s.logger.Error("failed to bulk update category", "error", err)
		return errors.NewInternalError("failed to bulk update category", err)
	}

	s.logger.Info("bulk category update completed", "service_count", len(serviceIDs), "category", category)
	return nil
}

// BulkDeleteServices deletes multiple services
func (s *serviceService) BulkDeleteServices(ctx context.Context, serviceIDs []uuid.UUID) error {
	if len(serviceIDs) == 0 {
		return errors.NewValidationError("at least one service ID is required")
	}

	if err := s.serviceRepo.BulkDelete(ctx, serviceIDs); err != nil {
		s.logger.Error("failed to bulk delete services", "error", err)
		return errors.NewInternalError("failed to bulk delete services", err)
	}

	s.logger.Info("bulk delete completed", "service_count", len(serviceIDs))
	return nil
}

// FilterServices filters services with advanced criteria
func (s *serviceService) FilterServices(ctx context.Context, tenantID uuid.UUID, filters *dto.ServiceFilterRequest) (*dto.ListServicesResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	if err := filters.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid filters")
	}

	repoFilters := types.ServiceFilters{
		Categories:      filters.Categories,
		MinPrice:        filters.MinPrice,
		MaxPrice:        filters.MaxPrice,
		MinDuration:     filters.MinDuration,
		MaxDuration:     filters.MaxDuration,
		IsActive:        filters.IsActive,
		ArtisanID:       filters.ArtisanID,
		Tags:            filters.Tags,
		RequiresDeposit: filters.RequiresDeposit,
	}

	pagination := repository.PaginationParams{
		Page:     filters.Page,
		PageSize: filters.PageSize,
	}

	services, paginationResult, err := s.serviceRepo.FindByFilters(ctx, tenantID, repoFilters, pagination)
	if err != nil {
		s.logger.Error("failed to filter services", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to filter services", err)
	}

	return &dto.ListServicesResponse{
		Services:   s.toServiceResponseList(services),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalItems: paginationResult.TotalItems,
		TotalPages: paginationResult.TotalPages,
	}, nil
}

// GetCategoriesWithCount gets categories with service counts
func (s *serviceService) GetCategoriesWithCount(ctx context.Context, tenantID uuid.UUID) ([]*dto.CategoryCountResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	counts, err := s.serviceRepo.GetCategoriesWithCount(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get categories with count", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to get categories with count", err)
	}

	responses := make([]*dto.CategoryCountResponse, len(counts))
	for i, count := range counts {
		responses[i] = &dto.CategoryCountResponse{
			Category: count.Category,
			Count:    count.Count,
		}
	}

	return responses, nil
}

// Helper methods

func (s *serviceService) toServiceResponse(service *models.Service) *dto.ServiceResponse {
	if service == nil {
		return nil
	}

	resp := &dto.ServiceResponse{
		ID:              service.ID,
		TenantID:        service.TenantID,
		ArtisanID:       service.ArtisanID,
		Name:            service.Name,
		Description:     service.Description,
		Category:        service.Category,
		Price:           service.Price,
		Currency:        service.Currency,
		DepositAmount:   service.DepositAmount,
		DurationMinutes: service.DurationMinutes,
		BufferMinutes:   service.BufferMinutes,
		TotalDuration:   service.GetTotalDuration(),
		IsActive:        service.IsActive,
		MaxBookingsDay:  service.MaxBookingsDay,
		ImageURL:        service.ImageURL,
		RequiresDeposit: service.RequiresDeposit,
		Tags:            service.Tags,
		Metadata:        service.Metadata,
		CreatedAt:       service.CreatedAt,
		UpdatedAt:       service.UpdatedAt,
	}

	if service.Artisan != nil {
		resp.ArtisanName = service.Artisan.FullName()
	}

	return resp
}

func (s *serviceService) toServiceResponseList(services []*models.Service) []*dto.ServiceResponse {
	responses := make([]*dto.ServiceResponse, len(services))
	for i, service := range services {
		responses[i] = s.toServiceResponse(service)
	}
	return responses
}
