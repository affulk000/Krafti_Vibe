package service

import (
	"context"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// CustomerService defines the interface for customer service operations
type CustomerService interface {
	// Core Operations
	CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CustomerResponse, error)
	GetCustomer(ctx context.Context, id uuid.UUID) (*dto.CustomerResponse, error)
	GetCustomerByUserID(ctx context.Context, userID uuid.UUID) (*dto.CustomerResponse, error)
	UpdateCustomer(ctx context.Context, id uuid.UUID, req *dto.UpdateCustomerRequest) (*dto.CustomerResponse, error)
	DeleteCustomer(ctx context.Context, id uuid.UUID) error
	ListCustomers(ctx context.Context, filter dto.CustomerFilter) (*dto.CustomerListResponse, error)

	// Customer Management
	SearchCustomers(ctx context.Context, req *dto.CustomerSearchRequest) (*dto.CustomerListResponse, error)
	GetCustomersByTenant(ctx context.Context, tenantID uuid.UUID, filter dto.CustomerFilter) (*dto.CustomerListResponse, error)
	GetActiveCustomers(ctx context.Context, tenantID uuid.UUID, filter dto.CustomerFilter) (*dto.CustomerListResponse, error)

	// Loyalty & Rewards
	UpdateLoyaltyPoints(ctx context.Context, customerID uuid.UUID, req *dto.UpdateLoyaltyPointsRequest) (*dto.CustomerResponse, error)
	GetLoyaltyPointsHistory(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*dto.LoyaltyPointsHistoryResponse, error)
	GetTopCustomers(ctx context.Context, tenantID uuid.UUID, criteria string, limit int) ([]*dto.CustomerResponse, error)

	// Preferences & Settings
	AddPreferredArtisan(ctx context.Context, customerID uuid.UUID, req *dto.AddPreferredArtisanRequest) (*dto.CustomerResponse, error)
	RemovePreferredArtisan(ctx context.Context, customerID, artisanID uuid.UUID) (*dto.CustomerResponse, error)
	UpdateNotificationPreferences(ctx context.Context, customerID uuid.UUID, req *dto.UpdateNotificationPreferencesRequest) (*dto.CustomerResponse, error)
	UpdatePrimaryLocation(ctx context.Context, customerID uuid.UUID, location *models.Location) (*dto.CustomerResponse, error)

	// Analytics & Reporting
	GetCustomerStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.CustomerStatsResponse, error)
	GetCustomerAnalytics(ctx context.Context, filter dto.CustomerAnalyticsFilter) (*dto.CustomerAnalyticsResponse, error)
	GetCustomerSegments(ctx context.Context, tenantID uuid.UUID) ([]*dto.CustomerSegmentResponse, error)
	GetCustomerRetentionAnalysis(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error)

	// Booking Integration
	RecordBookingCompletion(ctx context.Context, customerID uuid.UUID, bookingValue float64, loyaltyPoints int) error
	RecordBookingCancellation(ctx context.Context, customerID uuid.UUID) error
	UpdateCustomerSpending(ctx context.Context, customerID uuid.UUID, amount float64) error

	// Health & Monitoring
	HealthCheck(ctx context.Context) error
	GetServiceMetrics(ctx context.Context) map[string]interface{}
}

// customerService implements CustomerService
type customerService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewCustomerService creates a new CustomerService instance
func NewCustomerService(repos *repository.Repositories, logger log.AllLogger) CustomerService {
	return &customerService{
		repos:  repos,
		logger: logger,
	}
}

// ============================================================================
// Core Operations
// ============================================================================

// CreateCustomer creates a new customer profile
func (s *customerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CustomerResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	// Check if customer already exists for this user
	existingCustomer, err := s.repos.Customer.GetByUserID(ctx, req.UserID)
	if err != nil && !errors.IsNotFoundError(err) {
		return nil, errors.NewServiceError("CUSTOMER_CHECK_FAILED", "failed to check existing customer", err)
	}
	if existingCustomer != nil {
		return nil, errors.NewConflictError("customer profile already exists for user")
	}

	// Create customer model
	customer := &models.Customer{
		UserID:             req.UserID,
		TenantID:           req.TenantID,
		Notes:              req.Notes,
		PreferredArtisans:  req.PreferredArtisans,
		LoyaltyPoints:      0,
		TotalSpent:         0,
		TotalBookings:      0,
		CancelledBookings:  0,
		CompletedBookings:  0,
		EmailNotifications: req.EmailNotifications,
		SMSNotifications:   req.SMSNotifications,
		PushNotifications:  req.PushNotifications,
		Metadata:           req.Metadata,
	}

	if req.PrimaryLocation != nil {
		customer.PrimaryLocation = *req.PrimaryLocation
	}

	// Create in repository
	if err := s.repos.Customer.Create(ctx, customer); err != nil {
		return nil, errors.NewServiceError("CUSTOMER_CREATE_FAILED", "failed to create customer", err)
	}

	// User data is preloaded by repository

	s.logger.Info("customer created", "customer_id", customer.ID, "user_id", req.UserID, "tenant_id", req.TenantID)

	return dto.ToCustomerResponse(customer), nil
}

// GetCustomer retrieves a customer by ID
func (s *customerService) GetCustomer(ctx context.Context, id uuid.UUID) (*dto.CustomerResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("customer ID is required")
	}

	customer, err := s.repos.Customer.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("customer not found")
		}
		return nil, errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	// User data is preloaded by repository

	return dto.ToCustomerResponse(customer), nil
}

// GetCustomerByUserID retrieves a customer by user ID
func (s *customerService) GetCustomerByUserID(ctx context.Context, userID uuid.UUID) (*dto.CustomerResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user ID is required")
	}

	customer, err := s.repos.Customer.GetByUserID(ctx, userID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("customer not found")
		}
		return nil, errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	// User data is preloaded by repository

	return dto.ToCustomerResponse(customer), nil
}

// UpdateCustomer updates an existing customer profile
func (s *customerService) UpdateCustomer(ctx context.Context, id uuid.UUID, req *dto.UpdateCustomerRequest) (*dto.CustomerResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	customer, err := s.repos.Customer.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("customer not found")
		}
		return nil, errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	// Update fields
	if req.Notes != nil {
		customer.Notes = *req.Notes
	}
	if len(req.PreferredArtisans) > 0 {
		customer.PreferredArtisans = req.PreferredArtisans
	}
	if req.PrimaryLocation != nil {
		customer.PrimaryLocation = *req.PrimaryLocation
	}
	if req.DefaultPaymentMethodID != nil {
		customer.DefaultPaymentMethodID = *req.DefaultPaymentMethodID
	}
	if req.EmailNotifications != nil {
		customer.EmailNotifications = *req.EmailNotifications
	}
	if req.SMSNotifications != nil {
		customer.SMSNotifications = *req.SMSNotifications
	}
	if req.PushNotifications != nil {
		customer.PushNotifications = *req.PushNotifications
	}
	if req.Metadata != nil {
		customer.Metadata = req.Metadata
	}

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return nil, errors.NewServiceError("CUSTOMER_UPDATE_FAILED", "failed to update customer", err)
	}

	// User data is preloaded by repository

	s.logger.Info("customer updated", "customer_id", id)

	return dto.ToCustomerResponse(customer), nil
}

// DeleteCustomer soft deletes a customer
func (s *customerService) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	if err := s.repos.Customer.SoftDelete(ctx, id); err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("customer not found")
		}
		return errors.NewServiceError("CUSTOMER_DELETE_FAILED", "failed to delete customer", err)
	}

	s.logger.Info("customer deleted", "customer_id", id)
	return nil
}

// ListCustomers retrieves customers with filters and pagination
func (s *customerService) ListCustomers(ctx context.Context, filter dto.CustomerFilter) (*dto.CustomerListResponse, error) {
	if err := filter.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid filter: " + err.Error())
	}

	// Convert DTO filter to repository filter
	repoFilter := s.convertToRepoFilter(filter)

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	customers, paginationResult, err := s.repos.Customer.FindByFilters(ctx, repoFilter, pagination)
	if err != nil {
		return nil, errors.NewServiceError("CUSTOMERS_LIST_FAILED", "failed to list customers", err)
	}

	// User data is preloaded by repository

	// Calculate pagination manually
	totalPages := int((paginationResult.TotalItems + int64(paginationResult.PageSize) - 1) / int64(paginationResult.PageSize))
	hasNext := paginationResult.Page < totalPages
	hasPrevious := paginationResult.Page > 1

	return &dto.CustomerListResponse{
		Customers:   dto.ToCustomerResponses(customers),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}

// ============================================================================
// Customer Management
// ============================================================================

// SearchCustomers searches customers by name or email
func (s *customerService) SearchCustomers(ctx context.Context, req *dto.CustomerSearchRequest) (*dto.CustomerListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	customers, paginationResult, err := s.repos.Customer.Search(
		ctx,
		req.Query,
		req.TenantID,
		repository.PaginationParams{
			Page:     req.Filters.Page,
			PageSize: req.Filters.PageSize,
		},
	)
	if err != nil {
		return nil, errors.NewServiceError("CUSTOMER_SEARCH_FAILED", "failed to search customers", err)
	}

	// User data is preloaded by repository

	// Calculate pagination manually
	totalPages := int((paginationResult.TotalItems + int64(paginationResult.PageSize) - 1) / int64(paginationResult.PageSize))
	hasNext := paginationResult.Page < totalPages
	hasPrevious := paginationResult.Page > 1

	return &dto.CustomerListResponse{
		Customers:   dto.ToCustomerResponses(customers),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}

// GetCustomersByTenant retrieves customers for a specific tenant
func (s *customerService) GetCustomersByTenant(ctx context.Context, tenantID uuid.UUID, filter dto.CustomerFilter) (*dto.CustomerListResponse, error) {
	filter.TenantID = &tenantID
	return s.ListCustomers(ctx, filter)
}

// GetActiveCustomers retrieves active customers (with recent bookings)
func (s *customerService) GetActiveCustomers(ctx context.Context, tenantID uuid.UUID, filter dto.CustomerFilter) (*dto.CustomerListResponse, error) {
	// Set filter for customers with bookings in last 90 days
	lastBookingAfter := time.Now().AddDate(0, 0, -90)
	filter.TenantID = &tenantID
	filter.LastBookingAfter = &lastBookingAfter

	return s.ListCustomers(ctx, filter)
}

// ============================================================================
// Loyalty & Rewards
// ============================================================================

// UpdateLoyaltyPoints updates customer's loyalty points
func (s *customerService) UpdateLoyaltyPoints(ctx context.Context, customerID uuid.UUID, req *dto.UpdateLoyaltyPointsRequest) (*dto.CustomerResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("customer not found")
		}
		return nil, errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	// Update points based on operation
	switch req.Operation {
	case "add":
		customer.LoyaltyPoints += req.Points
	case "subtract":
		customer.LoyaltyPoints -= req.Points
		if customer.LoyaltyPoints < 0 {
			customer.LoyaltyPoints = 0
		}
	case "set":
		customer.LoyaltyPoints = req.Points
	}

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return nil, errors.NewServiceError("LOYALTY_POINTS_UPDATE_FAILED", "failed to update loyalty points", err)
	}

	// Note: Loyalty points history tracking would be implemented separately if needed

	s.logger.Info("loyalty points updated", "customer_id", customerID, "operation", req.Operation, "points", req.Points)

	// User data is preloaded by repository

	return dto.ToCustomerResponse(customer), nil
}

// GetLoyaltyPointsHistory retrieves loyalty points history for a customer
func (s *customerService) GetLoyaltyPointsHistory(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*dto.LoyaltyPointsHistoryResponse, error) {
	// This functionality would require a separate loyalty points history table
	// For now, return empty history
	return []*dto.LoyaltyPointsHistoryResponse{}, nil
}

// GetTopCustomers retrieves top customers by specified criteria
func (s *customerService) GetTopCustomers(ctx context.Context, tenantID uuid.UUID, criteria string, limit int) ([]*dto.CustomerResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var customers []*models.Customer
	var err error

	switch criteria {
	case "spending":
		customers, err = s.repos.Customer.GetTopSpendingCustomers(ctx, tenantID, limit)
	case "bookings":
		customers, err = s.repos.Customer.GetMostActiveCustomers(ctx, tenantID, limit)
	case "loyalty":
		customers, err = s.repos.Customer.GetTopLoyaltyCustomers(ctx, tenantID, limit)
	default:
		customers, err = s.repos.Customer.GetTopSpendingCustomers(ctx, tenantID, limit)
	}

	if err != nil {
		return nil, errors.NewServiceError("TOP_CUSTOMERS_GET_FAILED", "failed to get top customers", err)
	}

	// User data is preloaded by repository

	return dto.ToCustomerResponses(customers), nil
}

// ============================================================================
// Preferences & Settings
// ============================================================================

// AddPreferredArtisan adds an artisan to customer's preferred list
func (s *customerService) AddPreferredArtisan(ctx context.Context, customerID uuid.UUID, req *dto.AddPreferredArtisanRequest) (*dto.CustomerResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("customer not found")
		}
		return nil, errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	// Check if artisan is already in preferred list
	for _, artisanID := range customer.PreferredArtisans {
		if artisanID == req.ArtisanID {
			return dto.ToCustomerResponse(customer), nil // Already preferred
		}
	}

	// Add artisan to preferred list
	customer.PreferredArtisans = append(customer.PreferredArtisans, req.ArtisanID)

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return nil, errors.NewServiceError("PREFERRED_ARTISAN_ADD_FAILED", "failed to add preferred artisan", err)
	}

	s.logger.Info("preferred artisan added", "customer_id", customerID, "artisan_id", req.ArtisanID)

	// User data is preloaded by repository

	return dto.ToCustomerResponse(customer), nil
}

// RemovePreferredArtisan removes an artisan from customer's preferred list
func (s *customerService) RemovePreferredArtisan(ctx context.Context, customerID, artisanID uuid.UUID) (*dto.CustomerResponse, error) {
	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("customer not found")
		}
		return nil, errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	// Remove artisan from preferred list
	newPreferred := make([]uuid.UUID, 0, len(customer.PreferredArtisans))
	for _, id := range customer.PreferredArtisans {
		if id != artisanID {
			newPreferred = append(newPreferred, id)
		}
	}
	customer.PreferredArtisans = newPreferred

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return nil, errors.NewServiceError("PREFERRED_ARTISAN_REMOVE_FAILED", "failed to remove preferred artisan", err)
	}

	s.logger.Info("preferred artisan removed", "customer_id", customerID, "artisan_id", artisanID)

	return dto.ToCustomerResponse(customer), nil
}

// UpdateNotificationPreferences updates customer's notification preferences
func (s *customerService) UpdateNotificationPreferences(ctx context.Context, customerID uuid.UUID, req *dto.UpdateNotificationPreferencesRequest) (*dto.CustomerResponse, error) {
	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("customer not found")
		}
		return nil, errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	// Update notification preferences
	if req.EmailNotifications != nil {
		customer.EmailNotifications = *req.EmailNotifications
	}
	if req.SMSNotifications != nil {
		customer.SMSNotifications = *req.SMSNotifications
	}
	if req.PushNotifications != nil {
		customer.PushNotifications = *req.PushNotifications
	}

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_PREFERENCES_UPDATE_FAILED", "failed to update notification preferences", err)
	}

	s.logger.Info("notification preferences updated", "customer_id", customerID)

	// User data is preloaded by repository

	return dto.ToCustomerResponse(customer), nil
}

// UpdatePrimaryLocation updates customer's primary location
func (s *customerService) UpdatePrimaryLocation(ctx context.Context, customerID uuid.UUID, location *models.Location) (*dto.CustomerResponse, error) {
	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("customer not found")
		}
		return nil, errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	customer.PrimaryLocation = *location

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return nil, errors.NewServiceError("PRIMARY_LOCATION_UPDATE_FAILED", "failed to update primary location", err)
	}

	s.logger.Info("primary location updated", "customer_id", customerID)

	// User data is preloaded by repository

	return dto.ToCustomerResponse(customer), nil
}

// ============================================================================
// Analytics & Reporting
// ============================================================================

// GetCustomerStats retrieves customer statistics for a tenant
func (s *customerService) GetCustomerStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.CustomerStatsResponse, error) {
	analytics, err := s.repos.Customer.GetCustomerAnalytics(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("CUSTOMER_STATS_GET_FAILED", "failed to get customer statistics", err)
	}

	// Get top customers
	topBySpending, _ := s.repos.Customer.GetTopSpendingCustomers(ctx, tenantID, 5)
	topByBookings, _ := s.repos.Customer.GetMostActiveCustomers(ctx, tenantID, 5)

	// User data is preloaded by repository

	response := &dto.CustomerStatsResponse{
		TotalCustomers:             analytics.TotalCustomers,
		NewCustomers:               analytics.NewCustomers,
		ActiveCustomers:            analytics.ActiveCustomers,
		InactiveCustomers:          analytics.ChurnedCustomers,
		ByLoyaltyTier:              analytics.ByLoyaltyTier,
		TotalRevenue:               0, // Would need to be calculated separately
		AverageSpentPerCustomer:    analytics.AverageLifetimeValue,
		HighestSpendingCustomer:    0, // Would need to be calculated separately
		TotalBookings:              0, // Would need to be calculated separately
		AverageBookingsPerCustomer: analytics.AverageBookings,
		BookingCompletionRate:      0, // Would need to be calculated separately
		BookingCancelRate:          0, // Would need to be calculated separately
		CustomerGrowthRate:         0, // Would need to be calculated separately
		RevenueGrowthRate:          0, // Would need to be calculated separately
		CustomerRetentionRate:      analytics.RetentionRate,
		TopCustomersBySpending:     dto.ToCustomerResponses(topBySpending),
		TopCustomersByBookings:     dto.ToCustomerResponses(topByBookings),
	}

	return response, nil
}

// GetCustomerAnalytics retrieves customer analytics data
func (s *customerService) GetCustomerAnalytics(ctx context.Context, filter dto.CustomerAnalyticsFilter) (*dto.CustomerAnalyticsResponse, error) {
	if err := filter.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid filter: " + err.Error())
	}

	analytics, err := s.repos.Customer.GetCustomerAnalytics(ctx, filter.TenantID, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, errors.NewServiceError("CUSTOMER_ANALYTICS_GET_FAILED", "failed to get customer analytics", err)
	}

	response := &dto.CustomerAnalyticsResponse{
		Period:     fmt.Sprintf("%s to %s", filter.StartDate.Format("2006-01-02"), filter.EndDate.Format("2006-01-02")),
		TenantID:   filter.TenantID,
		GroupBy:    filter.GroupBy,
		DataPoints: []dto.CustomerAnalyticsDataPoint{},
		Summary:    &dto.CustomerAnalyticsSummary{},
	}

	// Convert analytics data to response format
	response.DataPoints = []dto.CustomerAnalyticsDataPoint{
		{
			Date:            filter.StartDate,
			NewCustomers:    analytics.NewCustomers,
			ActiveCustomers: analytics.ActiveCustomers,
			Revenue:         0, // Would need to be calculated separately
			Bookings:        0, // Would need to be calculated separately
			LoyaltyPoints:   0, // Would need to be calculated separately
		},
	}

	response.Summary = &dto.CustomerAnalyticsSummary{
		TotalNewCustomers:     analytics.NewCustomers,
		TotalRevenue:          0, // Would need to be calculated separately
		TotalBookings:         0, // Would need to be calculated separately
		AverageBookingValue:   0, // Would need to be calculated separately
		CustomerGrowthRate:    0, // Would need to be calculated separately
		RevenueGrowthRate:     0, // Would need to be calculated separately
		CustomerRetentionRate: analytics.RetentionRate,
	}

	return response, nil
}

// GetCustomerSegments retrieves customer segments for a tenant
func (s *customerService) GetCustomerSegments(ctx context.Context, tenantID uuid.UUID) ([]*dto.CustomerSegmentResponse, error) {
	// This method would implement customer segmentation logic
	// For now, return basic segments based on loyalty tiers
	segments := []*dto.CustomerSegmentResponse{
		{
			SegmentName:     "High Value",
			Description:     "Customers with high lifetime value",
			CustomerCount:   0,
			Percentage:      0,
			TotalRevenue:    0,
			AverageSpending: 0,
		},
		{
			SegmentName:     "Regular",
			Description:     "Regular customers",
			CustomerCount:   0,
			Percentage:      0,
			TotalRevenue:    0,
			AverageSpending: 0,
		},
		{
			SegmentName:     "At Risk",
			Description:     "Customers who haven't made recent bookings",
			CustomerCount:   0,
			Percentage:      0,
			TotalRevenue:    0,
			AverageSpending: 0,
		},
	}

	return segments, nil
}

// GetCustomerRetentionAnalysis retrieves customer retention analysis
func (s *customerService) GetCustomerRetentionAnalysis(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Get retention rate using existing method
	retentionRate, err := s.repos.Customer.GetCustomerRetentionRate(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("CUSTOMER_RETENTION_GET_FAILED", "failed to get customer retention analysis", err)
	}

	analysis := map[string]interface{}{
		"period":         fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"retention_rate": retentionRate,
		"churn_rate":     100 - retentionRate,
	}

	return analysis, nil
}

// ============================================================================
// Booking Integration
// ============================================================================

// RecordBookingCompletion records a completed booking for a customer
func (s *customerService) RecordBookingCompletion(ctx context.Context, customerID uuid.UUID, bookingValue float64, loyaltyPoints int) error {
	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		return errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	// Update customer statistics
	customer.TotalBookings++
	customer.CompletedBookings++
	customer.TotalSpent += bookingValue
	customer.LoyaltyPoints += loyaltyPoints

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return errors.NewServiceError("BOOKING_COMPLETION_RECORD_FAILED", "failed to record booking completion", err)
	}

	// Record loyalty points if earned
	if loyaltyPoints > 0 {
		if err := s.repos.Customer.AddLoyaltyPoints(ctx, customerID, loyaltyPoints); err != nil {
			s.logger.Error("failed to add loyalty points", "customer_id", customerID, "error", err)
		}
	}

	s.logger.Info("booking completion recorded", "customer_id", customerID, "value", bookingValue, "points", loyaltyPoints)
	return nil
}

// RecordBookingCancellation records a cancelled booking for a customer
func (s *customerService) RecordBookingCancellation(ctx context.Context, customerID uuid.UUID) error {
	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		return errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	customer.TotalBookings++
	customer.CancelledBookings++

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return errors.NewServiceError("BOOKING_CANCELLATION_RECORD_FAILED", "failed to record booking cancellation", err)
	}

	s.logger.Info("booking cancellation recorded", "customer_id", customerID)
	return nil
}

// UpdateCustomerSpending updates customer's total spending
func (s *customerService) UpdateCustomerSpending(ctx context.Context, customerID uuid.UUID, amount float64) error {
	customer, err := s.repos.Customer.GetByID(ctx, customerID)
	if err != nil {
		return errors.NewServiceError("CUSTOMER_GET_FAILED", "failed to get customer", err)
	}

	customer.TotalSpent += amount

	if err := s.repos.Customer.Update(ctx, customer); err != nil {
		return errors.NewServiceError("SPENDING_UPDATE_FAILED", "failed to update customer spending", err)
	}

	s.logger.Info("customer spending updated", "customer_id", customerID, "amount", amount)
	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// convertToRepoFilter converts DTO filter to repository filter
func (s *customerService) convertToRepoFilter(filter dto.CustomerFilter) repository.CustomerFilters {
	var tenantID uuid.UUID
	if filter.TenantID != nil {
		tenantID = *filter.TenantID
	}
	
	return repository.CustomerFilters{
		TenantID:            tenantID,
		LoyaltyTiers:        filter.LoyaltyTiers,
		MinTotalSpent:       filter.MinTotalSpent,
		MaxTotalSpent:       filter.MaxTotalSpent,
		MinLoyaltyPoints:    filter.MinLoyaltyPoints,
		MaxLoyaltyPoints:    filter.MaxLoyaltyPoints,
		MinBookings:         filter.MinBookings,
		MaxBookings:         filter.MaxBookings,
		PreferredArtisanIDs: filter.PreferredArtisans,
		CreatedAfter:        filter.CreatedAfter,
		CreatedBefore:       filter.CreatedBefore,
	}
}

// ============================================================================
// Health & Monitoring
// ============================================================================

// HealthCheck performs a health check on the customer service
func (s *customerService) HealthCheck(ctx context.Context) error {
	// Test database connectivity
	_, err := s.repos.Customer.Count(ctx, map[string]any{})
	if err != nil {
		return fmt.Errorf("customer repository health check failed: %w", err)
	}

	return nil
}

// GetServiceMetrics returns service metrics
func (s *customerService) GetServiceMetrics(ctx context.Context) map[string]interface{} {
	totalCustomers, _ := s.repos.Customer.Count(ctx, map[string]any{})

	return map[string]interface{}{
		"total_customers": totalCustomers,
		"service_status":  "healthy",
	}
}
