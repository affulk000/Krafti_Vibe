package service

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// BookingService defines the interface for booking service operations
type BookingService interface {
	// Core Booking Operations
	CreateBooking(ctx context.Context, req *dto.CreateBookingRequest) (*dto.BookingResponse, error)
	GetBooking(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error)
	UpdateBooking(ctx context.Context, id uuid.UUID, req *dto.UpdateBookingRequest) (*dto.BookingResponse, error)
	DeleteBooking(ctx context.Context, id uuid.UUID) error
	ListBookings(ctx context.Context, filter dto.BookingFilter) (*dto.BookingListResponse, error)

	// Booking Management
	SearchBookings(ctx context.Context, req *dto.BookingSearchRequest) (*dto.BookingListResponse, error)
	GetBookingsByTenant(ctx context.Context, tenantID uuid.UUID, filter dto.BookingFilter) (*dto.BookingListResponse, error)
	GetBookingsByArtisan(ctx context.Context, artisanID uuid.UUID, filter dto.BookingFilter) (*dto.BookingListResponse, error)
	GetBookingsByCustomer(ctx context.Context, customerID uuid.UUID, filter dto.BookingFilter) (*dto.BookingListResponse, error)
	GetBookingsByService(ctx context.Context, serviceID uuid.UUID, filter dto.BookingFilter) (*dto.BookingListResponse, error)

	// Status Management
	ConfirmBooking(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error)
	StartBooking(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error)
	CompleteBooking(ctx context.Context, id uuid.UUID, req *dto.CompleteBookingRequest) (*dto.BookingResponse, error)
	CancelBooking(ctx context.Context, id uuid.UUID, req *dto.CancelBookingRequest) (*dto.BookingResponse, error)
	MarkAsNoShow(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error)
	RescheduleBooking(ctx context.Context, id uuid.UUID, req *dto.RescheduleBookingRequest) (*dto.BookingResponse, error)

	// Scheduling & Availability
	CheckArtisanAvailability(ctx context.Context, req *dto.AvailabilityRequest) (*dto.AvailabilityResponse, error)
	GetAvailableTimeSlots(ctx context.Context, artisanID uuid.UUID, date time.Time, duration int) ([]*dto.TimeSlotResponse, error)
	HasBookingConflicts(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time, excludeBookingID *uuid.UUID) (bool, []*dto.ConflictResponse, error)
	GetArtisanSchedule(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) ([]*dto.BookingResponse, error)

	// Time-based Queries
	GetUpcomingBookings(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.BookingResponse, error)
	GetTodayBookings(ctx context.Context, tenantID uuid.UUID) ([]*dto.BookingResponse, error)
	GetBookingsInDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, filter dto.BookingFilter) (*dto.BookingListResponse, error)
	GetPastDueBookings(ctx context.Context, tenantID uuid.UUID) ([]*dto.BookingResponse, error)
	GetBookingsNeedingReminders(ctx context.Context, hoursAhead int) ([]*dto.BookingResponse, error)

	// Recurring Bookings
	CreateRecurringBookings(ctx context.Context, req *dto.CreateBookingRequest) ([]*dto.BookingResponse, error)
	GetRecurringBookingSeries(ctx context.Context, parentBookingID uuid.UUID) ([]*dto.BookingResponse, error)
	UpdateRecurringSeries(ctx context.Context, parentBookingID uuid.UUID, req *dto.UpdateBookingRequest, updateFuture bool) ([]*dto.BookingResponse, error)
	CancelRecurringSeries(ctx context.Context, parentBookingID uuid.UUID, reason string, cancelFuture bool) error

	// Photo Management
	AddBeforePhotos(ctx context.Context, bookingID uuid.UUID, photoURLs []string) (*dto.BookingResponse, error)
	AddAfterPhotos(ctx context.Context, bookingID uuid.UUID, photoURLs []string) (*dto.BookingResponse, error)

	// Payment Integration
	UpdatePaymentStatus(ctx context.Context, bookingID uuid.UUID, status models.PaymentStatus) (*dto.BookingResponse, error)
	RecordDepositPayment(ctx context.Context, bookingID uuid.UUID, amount float64, paymentIntentID string) (*dto.BookingResponse, error)
	ProcessRefund(ctx context.Context, bookingID uuid.UUID, amount float64, reason string) (*dto.BookingResponse, error)

	// Analytics & Reporting
	GetBookingStats(ctx context.Context, tenantID uuid.UUID) (*dto.BookingStatsResponse, error)
	GetBookingAnalytics(ctx context.Context, filter dto.BookingAnalyticsFilter) (*dto.BookingStatsResponse, error)
	GetArtisanBookingStats(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (*dto.BookingStatsResponse, error)
	GetPopularServices(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]*dto.PopularServiceData, error)
	GetBookingTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.BookingTrendData, error)

	// Bulk Operations
	BulkConfirm(ctx context.Context, bookingIDs []uuid.UUID) ([]*dto.BookingResponse, error)
	BulkCancel(ctx context.Context, bookingIDs []uuid.UUID, reason string) ([]*dto.BookingResponse, error)
	BulkReschedule(ctx context.Context, bookingIDs []uuid.UUID, newStartTime time.Time) ([]*dto.BookingResponse, error)
	BulkUpdateStatus(ctx context.Context, bookingIDs []uuid.UUID, status models.BookingStatus) ([]*dto.BookingResponse, error)

	// Integration Points
	NotifyBookingCreated(ctx context.Context, booking *models.Booking) error
	NotifyBookingUpdated(ctx context.Context, booking *models.Booking, oldStatus models.BookingStatus) error
	NotifyBookingCancelled(ctx context.Context, booking *models.Booking) error
	UpdateCustomerStatistics(ctx context.Context, customerID uuid.UUID, bookingValue float64, loyaltyPoints int) error

	// Health & Monitoring
	HealthCheck(ctx context.Context) error
	GetServiceMetrics(ctx context.Context) map[string]any
}

// bookingService implements BookingService
type bookingService struct {
	repos           *repository.Repositories
	logger          log.AllLogger
	customerService CustomerService
	paymentService  PaymentService
}

// NewBookingService creates a new BookingService instance
func NewBookingService(repos *repository.Repositories, logger log.AllLogger, customerService CustomerService, paymentService PaymentService) BookingService {
	return &bookingService{
		repos:           repos,
		logger:          logger,
		customerService: customerService,
		paymentService:  paymentService,
	}
}

// ============================================================================
// Core Booking Operations
// ============================================================================

// CreateBooking creates a new booking with comprehensive validation and business logic
func (s *bookingService) CreateBooking(ctx context.Context, req *dto.CreateBookingRequest) (*dto.BookingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	// Check artisan availability
	availabilityReq := &dto.AvailabilityRequest{
		ArtisanID: req.ArtisanID,
		Date:      req.StartTime,
		Duration:  req.Duration,
		ServiceID: &req.ServiceID,
	}
	availability, err := s.CheckArtisanAvailability(ctx, availabilityReq)
	if err != nil {
		return nil, errors.NewServiceError("AVAILABILITY_CHECK_FAILED", "failed to check availability", err)
	}
	if !availability.IsAvailable {
		return nil, errors.NewConflictError("artisan is not available for the requested time slot")
	}

	// Fetch service details for pricing
	service, err := s.repos.Service.GetByID(ctx, req.ServiceID)
	if err != nil {
		return nil, errors.NewServiceError("SERVICE_NOT_FOUND", "service not found", err)
	}

	// Calculate pricing with addons
	totalPrice := service.Price
	addonsPrice := float64(0)
	for _, addonID := range req.SelectedAddons {
		addon, err := s.repos.ServiceAddon.GetByID(ctx, addonID)
		if err != nil {
			s.logger.Warn("addon not found", "addon_id", addonID, "error", err)
			continue
		}
		addonsPrice += addon.Price
	}
	totalPrice += addonsPrice

	// Create booking model
	booking := &models.Booking{
		TenantID:          req.TenantID,
		ArtisanID:         req.ArtisanID,
		CustomerID:        req.CustomerID,
		ServiceID:         req.ServiceID,
		StartTime:         req.StartTime,
		EndTime:           req.StartTime.Add(time.Duration(req.Duration) * time.Minute),
		Duration:          req.Duration,
		Status:            models.BookingStatusPending,
		PaymentStatus:     models.PaymentStatusPending,
		BasePrice:         service.Price,
		AddonsPrice:       addonsPrice,
		TotalPrice:        totalPrice,
		DepositPaid:       req.DepositAmount,
		Currency:          service.Currency,
		Notes:             req.Notes,
		CustomerNotes:     req.CustomerNotes,
		SelectedAddons:    req.SelectedAddons,
		ServiceLocation:   req.ServiceLocation,
		IsRecurring:       req.IsRecurring,
		RecurrencePattern: req.RecurrencePattern,
		RecurrenceEndDate: req.RecurrenceEndDate,
		Metadata:          req.Metadata,
	}

	// Auto-confirm if requested
	if req.AutoConfirm {
		booking.Status = models.BookingStatusConfirmed
	}

	// Create in repository
	if err := s.repos.Booking.Create(ctx, booking); err != nil {
		return nil, errors.NewServiceError("BOOKING_CREATE_FAILED", "failed to create booking", err)
	}

	// Handle recurring bookings
	var recurringBookings []*models.Booking
	if req.IsRecurring {
		recurringBookings, err = s.createRecurringBookings(ctx, booking, req)
		if err != nil {
			s.logger.Error("failed to create recurring bookings", "error", err)
			// Continue with the main booking even if recurring creation fails
		}
	}

	// Process deposit payment if required
	if req.RequiresDeposit && req.DepositAmount > 0 && req.PaymentMethodID != "" {
		_, err := s.RecordDepositPayment(ctx, booking.ID, req.DepositAmount, req.PaymentMethodID)
		if err != nil {
			s.logger.Error("failed to process deposit payment", "booking_id", booking.ID, "error", err)
			// Continue with booking creation even if payment fails
		}
	}

	// Send notifications if requested
	if req.SendConfirmationEmail || req.SendConfirmationSMS {
		if err := s.NotifyBookingCreated(ctx, booking); err != nil {
			s.logger.Error("failed to send booking notifications", "booking_id", booking.ID, "error", err)
		}
	}

	// Update customer statistics
	if err := s.UpdateCustomerStatistics(ctx, booking.CustomerID, 0, 0); err != nil {
		s.logger.Error("failed to update customer statistics", "customer_id", booking.CustomerID, "error", err)
	}

	s.logger.Info("booking created", "booking_id", booking.ID, "tenant_id", req.TenantID, "artisan_id", req.ArtisanID, "customer_id", req.CustomerID)

	// Load related entities for response
	if err := s.loadBookingRelations(ctx, booking); err != nil {
		s.logger.Warn("failed to load booking relations", "booking_id", booking.ID, "error", err)
	}

	response := dto.ToBookingResponse(booking)

	// Add recurring booking count to metadata if applicable
	if req.IsRecurring && len(recurringBookings) > 0 {
		if response.Metadata == nil {
			response.Metadata = make(map[string]any)
		}
		response.Metadata["recurring_bookings_created"] = len(recurringBookings)
	}

	return response, nil
}

// GetBooking retrieves a booking by ID with full relations
func (s *bookingService) GetBooking(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("booking ID is required")
	}

	booking, err := s.repos.Booking.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("booking not found")
		}
		return nil, errors.NewServiceError("BOOKING_GET_FAILED", "failed to get booking", err)
	}

	// Load related entities
	if err := s.loadBookingRelations(ctx, booking); err != nil {
		s.logger.Warn("failed to load booking relations", "booking_id", id, "error", err)
	}

	return dto.ToBookingResponse(booking), nil
}

// UpdateBooking updates an existing booking with validation
func (s *bookingService) UpdateBooking(ctx context.Context, id uuid.UUID, req *dto.UpdateBookingRequest) (*dto.BookingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	booking, err := s.repos.Booking.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("booking not found")
		}
		return nil, errors.NewServiceError("BOOKING_GET_FAILED", "failed to get booking", err)
	}

	// Store old status for notifications
	oldStatus := booking.Status

	// Validate status transitions
	if req.Status != nil {
		if err := s.validateStatusTransition(booking.Status, *req.Status); err != nil {
			return nil, errors.NewValidationError("invalid status transition: " + err.Error())
		}
		booking.Status = *req.Status

		// Handle status-specific logic
		switch *req.Status {
		case models.BookingStatusConfirmed:
			// No additional logic needed
		case models.BookingStatusInProgress:
			// Could add start time tracking
		case models.BookingStatusCompleted:
			now := time.Now()
			booking.CompletedAt = &now
		case models.BookingStatusCancelled:
			now := time.Now()
			booking.CancelledAt = &now
			if req.CancellationReason != nil {
				booking.CancellationReason = *req.CancellationReason
			}
		}
	}

	// Update other fields
	if req.Notes != nil {
		booking.Notes = *req.Notes
	}
	if req.CustomerNotes != nil {
		booking.CustomerNotes = *req.CustomerNotes
	}
	if req.InternalNotes != nil {
		booking.InternalNotes = *req.InternalNotes
	}
	if req.ServiceLocation != nil {
		booking.ServiceLocation = req.ServiceLocation
	}
	if req.PaymentStatus != nil {
		booking.PaymentStatus = *req.PaymentStatus
	}
	if len(req.SelectedAddons) > 0 {
		// Recalculate pricing if addons changed
		addonsPrice := float64(0)
		for _, addonID := range req.SelectedAddons {
			addon, err := s.repos.ServiceAddon.GetByID(ctx, addonID)
			if err != nil {
				s.logger.Warn("addon not found", "addon_id", addonID, "error", err)
				continue
			}
			addonsPrice += addon.Price
		}
		booking.SelectedAddons = req.SelectedAddons
		booking.AddonsPrice = addonsPrice
		booking.TotalPrice = booking.BasePrice + addonsPrice
	}
	if req.PaymentIntentID != nil {
		booking.PaymentIntentID = *req.PaymentIntentID
	}
	if req.RefundID != nil {
		booking.RefundID = *req.RefundID
	}
	if len(req.BeforePhotoURLs) > 0 {
		booking.BeforePhotoURLs = req.BeforePhotoURLs
	}
	if len(req.AfterPhotoURLs) > 0 {
		booking.AfterPhotoURLs = req.AfterPhotoURLs
	}
	if req.ReminderSent24h != nil {
		booking.ReminderSent24h = *req.ReminderSent24h
	}
	if req.ReminderSent1h != nil {
		booking.ReminderSent1h = *req.ReminderSent1h
	}
	if req.Metadata != nil {
		if booking.Metadata == nil {
			booking.Metadata = make(map[string]any)
		}
		maps.Copy(booking.Metadata, req.Metadata)
	}

	// Update in repository
	if err := s.repos.Booking.Update(ctx, booking); err != nil {
		return nil, errors.NewServiceError("BOOKING_UPDATE_FAILED", "failed to update booking", err)
	}

	// Send notifications if status changed
	if req.Status != nil && oldStatus != *req.Status {
		if err := s.NotifyBookingUpdated(ctx, booking, oldStatus); err != nil {
			s.logger.Error("failed to send booking update notifications", "booking_id", id, "error", err)
		}

		// Update customer statistics for completed bookings
		if *req.Status == models.BookingStatusCompleted {
			loyaltyPoints := int(booking.TotalPrice / 10) // 1 point per $10 spent
			if err := s.UpdateCustomerStatistics(ctx, booking.CustomerID, booking.TotalPrice, loyaltyPoints); err != nil {
				s.logger.Error("failed to update customer statistics", "customer_id", booking.CustomerID, "error", err)
			}
		}
	}

	s.logger.Info("booking updated", "booking_id", id)

	// Load related entities for response
	if err := s.loadBookingRelations(ctx, booking); err != nil {
		s.logger.Warn("failed to load booking relations", "booking_id", id, "error", err)
	}

	return dto.ToBookingResponse(booking), nil
}

// DeleteBooking soft deletes a booking
func (s *bookingService) DeleteBooking(ctx context.Context, id uuid.UUID) error {
	booking, err := s.repos.Booking.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("booking not found")
		}
		return errors.NewServiceError("BOOKING_GET_FAILED", "failed to get booking", err)
	}

	// Validate that booking can be deleted
	if booking.Status == models.BookingStatusInProgress {
		return errors.NewConflictError("cannot delete booking that is in progress")
	}

	if err := s.repos.Booking.SoftDelete(ctx, id); err != nil {
		return errors.NewServiceError("BOOKING_DELETE_FAILED", "failed to delete booking", err)
	}

	s.logger.Info("booking deleted", "booking_id", id)
	return nil
}

// ListBookings retrieves bookings with filters and pagination
func (s *bookingService) ListBookings(ctx context.Context, filter dto.BookingFilter) (*dto.BookingListResponse, error) {
	if err := filter.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid filter: " + err.Error())
	}

	// Convert DTO filter to repository filter
	repoFilter := s.convertToRepoFilter(filter)

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	bookings, paginationResult, err := s.repos.Booking.FindByFilters(ctx, repoFilter, pagination)
	if err != nil {
		return nil, errors.NewServiceError("BOOKINGS_LIST_FAILED", "failed to list bookings", err)
	}

	// Load related entities for included relations
	for _, booking := range bookings {
		if err := s.loadBookingRelationsSelective(ctx, booking, filter.IncludeRelations); err != nil {
			s.logger.Warn("failed to load booking relations", "booking_id", booking.ID, "error", err)
		}
	}

	// Calculate pagination manually
	totalPages := int((paginationResult.TotalItems + int64(paginationResult.PageSize) - 1) / int64(paginationResult.PageSize))
	hasNext := paginationResult.Page < totalPages
	hasPrevious := paginationResult.Page > 1

	return &dto.BookingListResponse{
		Bookings:    dto.ToBookingResponses(bookings),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}

// ============================================================================
// Status Management
// ============================================================================

// ConfirmBooking confirms a pending booking
func (s *bookingService) ConfirmBooking(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error) {
	updateReq := &dto.UpdateBookingRequest{
		Status: &[]models.BookingStatus{models.BookingStatusConfirmed}[0],
	}
	return s.UpdateBooking(ctx, id, updateReq)
}

// StartBooking starts a confirmed booking
func (s *bookingService) StartBooking(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error) {
	updateReq := &dto.UpdateBookingRequest{
		Status: &[]models.BookingStatus{models.BookingStatusInProgress}[0],
	}
	return s.UpdateBooking(ctx, id, updateReq)
}

// CompleteBooking marks a booking as completed
func (s *bookingService) CompleteBooking(ctx context.Context, id uuid.UUID, req *dto.CompleteBookingRequest) (*dto.BookingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	status := models.BookingStatusCompleted
	updateReq := &dto.UpdateBookingRequest{
		Status:        &status,
		InternalNotes: &req.CompletionNotes,
	}

	if len(req.BeforePhotoURLs) > 0 {
		updateReq.BeforePhotoURLs = req.BeforePhotoURLs
	}
	if len(req.AfterPhotoURLs) > 0 {
		updateReq.AfterPhotoURLs = req.AfterPhotoURLs
	}

	return s.UpdateBooking(ctx, id, updateReq)
}

// CancelBooking cancels a booking with reason and refund processing
func (s *bookingService) CancelBooking(ctx context.Context, id uuid.UUID, req *dto.CancelBookingRequest) (*dto.BookingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	booking, err := s.repos.Booking.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("booking not found")
		}
		return nil, errors.NewServiceError("BOOKING_GET_FAILED", "failed to get booking", err)
	}

	// Check if booking can be cancelled
	if !booking.CanBeCancelled() {
		return nil, errors.NewConflictError("booking cannot be cancelled")
	}

	// Process refund if requested
	if req.RefundRequested && booking.DepositPaid > 0 {
		refundAmount := booking.CalculateRefundAmount()
		if refundAmount > 0 {
			_, err := s.ProcessRefund(ctx, id, refundAmount, req.Reason)
			if err != nil {
				s.logger.Error("failed to process refund", "booking_id", id, "error", err)
				// Continue with cancellation even if refund fails
			}
		}
	}

	// Update booking status
	status := models.BookingStatusCancelled
	updateReq := &dto.UpdateBookingRequest{
		Status:             &status,
		CancellationReason: &req.Reason,
	}

	response, err := s.UpdateBooking(ctx, id, updateReq)
	if err != nil {
		return nil, err
	}

	// Send notifications if requested
	if req.NotifyCustomer || req.NotifyArtisan {
		if err := s.NotifyBookingCancelled(ctx, booking); err != nil {
			s.logger.Error("failed to send cancellation notifications", "booking_id", id, "error", err)
		}
	}

	return response, nil
}

// MarkAsNoShow marks a booking as no-show
func (s *bookingService) MarkAsNoShow(ctx context.Context, id uuid.UUID) (*dto.BookingResponse, error) {
	status := models.BookingStatusNoShow
	updateReq := &dto.UpdateBookingRequest{
		Status: &status,
	}
	return s.UpdateBooking(ctx, id, updateReq)
}

// RescheduleBooking reschedules a booking to a new time
func (s *bookingService) RescheduleBooking(ctx context.Context, id uuid.UUID, req *dto.RescheduleBookingRequest) (*dto.BookingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	booking, err := s.repos.Booking.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("booking not found")
		}
		return nil, errors.NewServiceError("BOOKING_GET_FAILED", "failed to get booking", err)
	}

	// Check if booking can be rescheduled
	if booking.Status == models.BookingStatusCompleted || booking.Status == models.BookingStatusCancelled {
		return nil, errors.NewConflictError("cannot reschedule completed or cancelled booking")
	}

	// Use new duration if provided, otherwise keep existing
	duration := booking.Duration
	if req.NewDuration != nil {
		duration = *req.NewDuration
	}

	// Check availability for new time slot
	availabilityReq := &dto.AvailabilityRequest{
		ArtisanID:        booking.ArtisanID,
		Date:             req.NewStartTime,
		Duration:         duration,
		ExcludeBookingID: &booking.ID,
	}
	availability, err := s.CheckArtisanAvailability(ctx, availabilityReq)
	if err != nil {
		return nil, errors.NewServiceError("AVAILABILITY_CHECK_FAILED", "failed to check availability", err)
	}
	if !availability.IsAvailable {
		return nil, errors.NewConflictError("artisan is not available for the requested time slot")
	}

	// Update booking time and duration
	booking.StartTime = req.NewStartTime
	booking.EndTime = req.NewStartTime.Add(time.Duration(duration) * time.Minute)
	booking.Duration = duration

	// Recalculate pricing if duration changed and service pricing is duration-based
	if req.NewDuration != nil {
		// This would depend on your pricing model
		// For now, we'll keep the same price
	}

	// Add reschedule reason to metadata
	if booking.Metadata == nil {
		booking.Metadata = make(map[string]any)
	}
	if req.Reason != "" {
		booking.Metadata["reschedule_reason"] = req.Reason
	}
	booking.Metadata["reschedule_count"] = getIntFromMetadata(booking.Metadata, "reschedule_count") + 1

	if err := s.repos.Booking.Update(ctx, booking); err != nil {
		return nil, errors.NewServiceError("BOOKING_UPDATE_FAILED", "failed to reschedule booking", err)
	}

	// Send notifications if requested
	if req.NotifyCustomer || req.NotifyArtisan {
		if err := s.NotifyBookingUpdated(ctx, booking, booking.Status); err != nil {
			s.logger.Error("failed to send reschedule notifications", "booking_id", id, "error", err)
		}
	}

	s.logger.Info("booking rescheduled", "booking_id", id, "new_start_time", req.NewStartTime)

	// Load related entities for response
	if err := s.loadBookingRelations(ctx, booking); err != nil {
		s.logger.Warn("failed to load booking relations", "booking_id", id, "error", err)
	}

	return dto.ToBookingResponse(booking), nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// validateStatusTransition validates if a status transition is allowed
func (s *bookingService) validateStatusTransition(from, to models.BookingStatus) error {
	validTransitions := map[models.BookingStatus][]models.BookingStatus{
		models.BookingStatusPending: {
			models.BookingStatusConfirmed,
			models.BookingStatusCancelled,
		},
		models.BookingStatusConfirmed: {
			models.BookingStatusInProgress,
			models.BookingStatusCancelled,
			models.BookingStatusNoShow,
		},
		models.BookingStatusInProgress: {
			models.BookingStatusCompleted,
			models.BookingStatusCancelled,
		},
		models.BookingStatusCompleted: {}, // Final state
		models.BookingStatusCancelled: {}, // Final state
		models.BookingStatusNoShow:    {}, // Final state
	}

	allowed := validTransitions[from]
	if !slices.Contains(allowed, to) {
		return fmt.Errorf("cannot transition from %s to %s", from, to)
	}

	return fmt.Errorf("cannot transition from %s to %s", from, to)
}

// convertToRepoFilter converts DTO filter to repository filter
func (s *bookingService) convertToRepoFilter(filter dto.BookingFilter) repository.BookingFilters {
	var tenantID uuid.UUID
	if filter.TenantID != nil {
		tenantID = *filter.TenantID
	}

	return repository.BookingFilters{
		TenantID:        tenantID,
		ArtisanIDs:      filter.ArtisanIDs,
		CustomerIDs:     filter.CustomerIDs,
		ServiceIDs:      filter.ServiceIDs,
		Statuses:        filter.Statuses,
		PaymentStatuses: filter.PaymentStatuses,
		StartDateFrom:   filter.StartDate,
		StartDateTo:     filter.EndDate,
		MinPrice:        filter.MinAmount,
		MaxPrice:        filter.MaxAmount,
		IsRecurring:     filter.IsRecurring,
	}
}

// loadBookingRelations loads all related entities for a booking
func (s *bookingService) loadBookingRelations(ctx context.Context, booking *models.Booking) error {
	return s.loadBookingRelationsSelective(ctx, booking, []string{"artisan", "customer", "service", "payments", "review"})
}

// loadBookingRelationsSelective loads only specified relations for a booking
func (s *bookingService) loadBookingRelationsSelective(ctx context.Context, booking *models.Booking, relations []string) error {
	if booking == nil || len(relations) == 0 {
		return nil
	}

	for _, relation := range relations {
		var err error
		switch relation {
		case "artisan":
			if booking.ArtisanID != uuid.Nil && booking.Artisan == nil {
				booking.Artisan, err = s.repos.User.GetByID(ctx, booking.ArtisanID)
				if err != nil {
					s.logger.Warn("failed to load artisan", "artisan_id", booking.ArtisanID, "booking_id", booking.ID, "error", err)
				}
			}

		case "customer":
			if booking.CustomerID != uuid.Nil && booking.Customer == nil {
				booking.Customer, err = s.repos.User.GetByID(ctx, booking.CustomerID)
				if err != nil {
					s.logger.Warn("failed to load customer", "customer_id", booking.CustomerID, "booking_id", booking.ID, "error", err)
				}
			}

		case "service":
			if booking.ServiceID != uuid.Nil && booking.Service == nil {
				booking.Service, err = s.repos.Service.GetByID(ctx, booking.ServiceID)
				if err != nil {
					s.logger.Warn("failed to load service", "service_id", booking.ServiceID, "booking_id", booking.ID, "error", err)
				}
			}

		case "payments":
			if booking.Payments == nil {
				paymentsPtr, err := s.repos.Payment.GetByBookingID(ctx, booking.ID)
				if err != nil {
					s.logger.Warn("failed to load payments", "booking_id", booking.ID, "error", err)
				} else if paymentsPtr != nil {
					// Convert []*Payment to []Payment
					booking.Payments = make([]models.Payment, len(paymentsPtr))
					for i, p := range paymentsPtr {
						if p != nil {
							booking.Payments[i] = *p
						}
					}
				}
			}

		case "review":
			if booking.Review == nil {
				review, err := s.repos.Review.FindByBookingID(ctx, booking.ID)
				if err != nil {
					// It's normal for bookings to not have reviews yet
					if !errors.IsNotFoundError(err) {
						s.logger.Warn("failed to load review", "booking_id", booking.ID, "error", err)
					}
				} else {
					booking.Review = review
				}
			}

		case "tenant":
			if booking.TenantID != uuid.Nil && booking.Tenant == nil {
				tenant, err := s.repos.Tenant.GetByID(ctx, booking.TenantID)
				if err != nil {
					s.logger.Warn("failed to load tenant", "tenant_id", booking.TenantID, "booking_id", booking.ID, "error", err)
				} else {
					booking.Tenant = tenant
				}
			}

		case "parent":
			if booking.ParentBookingID != nil && *booking.ParentBookingID != uuid.Nil && booking.ParentBooking == nil {
				parent, err := s.repos.Booking.GetByID(ctx, *booking.ParentBookingID)
				if err != nil {
					s.logger.Warn("failed to load parent booking", "parent_id", *booking.ParentBookingID, "booking_id", booking.ID, "error", err)
				} else {
					booking.ParentBooking = parent
				}
			}

		case "children":
			// Note: Child bookings would typically be loaded via GORM preload
			// or a custom repository method. For now, we skip this relation.
			// To implement, add a FindByParentBookingID method to BookingRepository
			s.logger.Debug("children relation not yet implemented", "booking_id", booking.ID)

		default:
			s.logger.Debug("unknown relation requested", "relation", relation, "booking_id", booking.ID)
		}
	}

	return nil
}

// createRecurringBookings creates recurring bookings based on the parent booking
func (s *bookingService) createRecurringBookings(ctx context.Context, parentBooking *models.Booking, req *dto.CreateBookingRequest) ([]*models.Booking, error) {
	var recurringBookings []*models.Booking

	// Calculate the number of occurrences
	maxOccurrences := 52 // Safety limit
	if req.RecurrenceOccurrences != nil {
		maxOccurrences = *req.RecurrenceOccurrences - 1 // Minus 1 because parent is already created
	}

	currentTime := req.StartTime
	for i := 0; i < maxOccurrences; i++ {
		// Calculate next occurrence time
		switch req.RecurrencePattern {
		case "weekly":
			currentTime = currentTime.AddDate(0, 0, 7)
		case "biweekly":
			currentTime = currentTime.AddDate(0, 0, 14)
		case "monthly":
			currentTime = currentTime.AddDate(0, 1, 0)
		default:
			return recurringBookings, fmt.Errorf("unsupported recurrence pattern: %s", req.RecurrencePattern)
		}

		// Stop if we've passed the end date
		if req.RecurrenceEndDate != nil && currentTime.After(*req.RecurrenceEndDate) {
			break
		}

		// Check availability for this time slot
		availabilityReq := &dto.AvailabilityRequest{
			ArtisanID: req.ArtisanID,
			Date:      currentTime,
			Duration:  req.Duration,
			ServiceID: &req.ServiceID,
		}
		availability, err := s.CheckArtisanAvailability(ctx, availabilityReq)
		if err != nil || !availability.IsAvailable {
			s.logger.Warn("skipping recurring booking due to unavailability", "time", currentTime)
			continue
		}

		// Create recurring booking
		recurringBooking := &models.Booking{
			TenantID:          parentBooking.TenantID,
			ArtisanID:         parentBooking.ArtisanID,
			CustomerID:        parentBooking.CustomerID,
			ServiceID:         parentBooking.ServiceID,
			StartTime:         currentTime,
			EndTime:           currentTime.Add(time.Duration(req.Duration) * time.Minute),
			Duration:          parentBooking.Duration,
			Status:            models.BookingStatusPending,
			PaymentStatus:     models.PaymentStatusPending,
			BasePrice:         parentBooking.BasePrice,
			AddonsPrice:       parentBooking.AddonsPrice,
			TotalPrice:        parentBooking.TotalPrice,
			Currency:          parentBooking.Currency,
			Notes:             parentBooking.Notes,
			CustomerNotes:     parentBooking.CustomerNotes,
			SelectedAddons:    parentBooking.SelectedAddons,
			ServiceLocation:   parentBooking.ServiceLocation,
			IsRecurring:       true,
			RecurrencePattern: parentBooking.RecurrencePattern,
			ParentBookingID:   &parentBooking.ID,
			RecurrenceEndDate: parentBooking.RecurrenceEndDate,
			Metadata:          parentBooking.Metadata,
		}

		if err := s.repos.Booking.Create(ctx, recurringBooking); err != nil {
			s.logger.Error("failed to create recurring booking", "time", currentTime, "error", err)
			continue
		}

		recurringBookings = append(recurringBookings, recurringBooking)
	}

	return recurringBookings, nil
}

// getIntFromMetadata safely gets an int value from metadata
func getIntFromMetadata(metadata map[string]any, key string) int {
	if metadata == nil {
		return 0
	}
	if val, ok := metadata[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return 0
}

// ============================================================================
// Health & Monitoring
// ============================================================================

// HealthCheck performs a health check on the booking service
func (s *bookingService) HealthCheck(ctx context.Context) error {
	// Test database connectivity
	_, err := s.repos.Booking.Count(ctx, map[string]any{})
	if err != nil {
		return fmt.Errorf("booking repository health check failed: %w", err)
	}

	return nil
}

// ============================================================================
// Scheduling and Availability Management
// ============================================================================

// CheckArtisanAvailability checks if an artisan is available at a specific time
func (s *bookingService) CheckArtisanAvailability(ctx context.Context, req *dto.AvailabilityRequest) (*dto.AvailabilityResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get artisan's working hours for the requested date
	workingHours, err := s.getArtisanWorkingHours(ctx, req.ArtisanID, req.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to get working hours: %w", err)
	}

	// Get existing bookings for the artisan on the date
	existingBookings, err := s.getArtisanBookingsForDate(ctx, req.ArtisanID, req.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing bookings: %w", err)
	}

	// Check for conflicts
	conflicts := s.findBookingConflicts(req, existingBookings)

	// Generate time slots
	timeSlots := s.generateAvailableTimeSlots(req, workingHours, existingBookings)

	response := &dto.AvailabilityResponse{
		ArtisanID:    req.ArtisanID,
		Date:         req.Date,
		IsAvailable:  len(conflicts) == 0,
		TimeSlots:    timeSlots,
		WorkingHours: workingHours,
		Conflicts:    conflicts,
	}

	return response, nil
}

// GetAvailableTimeSlots returns available time slots for an artisan on a specific day
func (s *bookingService) GetAvailableTimeSlots(ctx context.Context, artisanID uuid.UUID, date time.Time, duration int) ([]*dto.TimeSlotResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan ID is required")
	}
	if duration < 15 || duration > 480 {
		return nil, errors.NewValidationError("duration must be between 15 minutes and 8 hours")
	}

	// Get working hours
	workingHours, err := s.getArtisanWorkingHours(ctx, artisanID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get working hours: %w", err)
	}

	// Get existing bookings
	existingBookings, err := s.getArtisanBookingsForDate(ctx, artisanID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing bookings: %w", err)
	}

	// Create availability request for slot generation
	req := &dto.AvailabilityRequest{
		ArtisanID: artisanID,
		Date:      date,
		Duration:  duration,
	}

	return s.generateAvailableTimeSlots(req, workingHours, existingBookings), nil
}

// HasBookingConflicts checks if a booking request would conflict with existing bookings
func (s *bookingService) HasBookingConflicts(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time, excludeBookingID *uuid.UUID) (bool, []*dto.ConflictResponse, error) {
	if artisanID == uuid.Nil {
		return false, nil, errors.NewValidationError("artisan ID is required")
	}
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return false, nil, errors.NewValidationError("end time must be after start time")
	}

	// Get bookings for the date range
	dateStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	dateEnd := dateStart.Add(24 * time.Hour)

	existingBookings, err := s.repos.Booking.GetArtisanBookingsInRange(ctx, artisanID, dateStart, dateEnd)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get existing bookings: %w", err)
	}

	// Filter out excluded booking and cancelled/no-show bookings
	var validBookings []*models.Booking
	for _, booking := range existingBookings {
		if excludeBookingID != nil && booking.ID == *excludeBookingID {
			continue
		}
		if booking.Status == models.BookingStatusCancelled || booking.Status == models.BookingStatusNoShow {
			continue
		}
		validBookings = append(validBookings, booking)
	}
	existingBookings = validBookings

	conflicts := make([]*dto.ConflictResponse, 0)

	// Check for time overlaps
	for _, booking := range existingBookings {
		if s.timePeriodsOverlap(startTime, endTime, booking.StartTime, booking.EndTime) {
			conflicts = append(conflicts, &dto.ConflictResponse{
				ConflictType: "booking",
				StartTime:    booking.StartTime,
				EndTime:      booking.EndTime,
				BookingID:    &booking.ID,
				Reason:       fmt.Sprintf("Conflicts with existing booking (Status: %s)", booking.Status),
			})
		}
	}

	return len(conflicts) > 0, conflicts, nil
}

// GetArtisanSchedule returns an artisan's schedule for a date range
func (s *bookingService) GetArtisanSchedule(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) ([]*dto.BookingResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan ID is required")
	}
	if startDate.After(endDate) {
		return nil, errors.NewValidationError("start date cannot be after end date")
	}

	filter := dto.BookingFilter{
		ArtisanIDs:       []uuid.UUID{artisanID},
		StartDate:        &startDate,
		EndDate:          &endDate,
		Statuses:         []models.BookingStatus{models.BookingStatusPending, models.BookingStatusConfirmed, models.BookingStatusInProgress},
		Page:             1,
		PageSize:         100,
		SortBy:           "start_time",
		SortOrder:        "asc",
		IncludeRelations: []string{"customer", "service"},
	}

	response, err := s.ListBookings(ctx, filter)
	if err != nil {
		return nil, err
	}

	return response.Bookings, nil
}

// ============================================================================
// Booking Management Methods
// ============================================================================

// SearchBookings searches bookings by query string
func (s *bookingService) SearchBookings(ctx context.Context, req *dto.BookingSearchRequest) (*dto.BookingListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid search request: " + err.Error())
	}

	pagination := repository.PaginationParams{
		Page:     req.Filters.Page,
		PageSize: req.Filters.PageSize,
	}

	bookings, paginationResult, err := s.repos.Booking.Search(ctx, req.Query, req.TenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("SEARCH_FAILED", "failed to search bookings", err)
	}

	// Load related entities if requested
	for _, booking := range bookings {
		if err := s.loadBookingRelationsSelective(ctx, booking, req.Filters.IncludeRelations); err != nil {
			s.logger.Warn("failed to load booking relations", "booking_id", booking.ID, "error", err)
		}
	}

	totalPages := int((paginationResult.TotalItems + int64(paginationResult.PageSize) - 1) / int64(paginationResult.PageSize))
	hasNext := paginationResult.Page < totalPages
	hasPrevious := paginationResult.Page > 1

	return &dto.BookingListResponse{
		Bookings:    dto.ToBookingResponses(bookings),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}

// GetBookingsByTenant retrieves all bookings for a tenant
func (s *bookingService) GetBookingsByTenant(ctx context.Context, tenantID uuid.UUID, filter dto.BookingFilter) (*dto.BookingListResponse, error) {
	filter.TenantID = &tenantID
	return s.ListBookings(ctx, filter)
}

// GetBookingsByArtisan retrieves all bookings for an artisan
func (s *bookingService) GetBookingsByArtisan(ctx context.Context, artisanID uuid.UUID, filter dto.BookingFilter) (*dto.BookingListResponse, error) {
	filter.ArtisanIDs = []uuid.UUID{artisanID}
	return s.ListBookings(ctx, filter)
}

// GetBookingsByCustomer retrieves all bookings for a customer
func (s *bookingService) GetBookingsByCustomer(ctx context.Context, customerID uuid.UUID, filter dto.BookingFilter) (*dto.BookingListResponse, error) {
	filter.CustomerIDs = []uuid.UUID{customerID}
	return s.ListBookings(ctx, filter)
}

// GetBookingsByService retrieves all bookings for a service
func (s *bookingService) GetBookingsByService(ctx context.Context, serviceID uuid.UUID, filter dto.BookingFilter) (*dto.BookingListResponse, error) {
	filter.ServiceIDs = []uuid.UUID{serviceID}
	return s.ListBookings(ctx, filter)
}

// ============================================================================
// Time-based Query Methods
// ============================================================================

// GetUpcomingBookings returns upcoming bookings within a specified number of days
func (s *bookingService) GetUpcomingBookings(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.BookingResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if days <= 0 {
		return nil, errors.NewValidationError("days must be positive")
	}

	pagination := repository.PaginationParams{
		Page:     1,
		PageSize: 100,
	}

	bookings, _, err := s.repos.Booking.GetUpcomingBookings(ctx, tenantID, days, pagination)
	if err != nil {
		return nil, errors.NewServiceError("QUERY_FAILED", "failed to get upcoming bookings", err)
	}

	// Load related entities
	for _, booking := range bookings {
		if err := s.loadBookingRelations(ctx, booking); err != nil {
			s.logger.Warn("failed to load booking relations", "booking_id", booking.ID, "error", err)
		}
	}

	return dto.ToBookingResponses(bookings), nil
}

// GetTodayBookings returns all bookings for today
func (s *bookingService) GetTodayBookings(ctx context.Context, tenantID uuid.UUID) ([]*dto.BookingResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	bookings, err := s.repos.Booking.GetTodayBookings(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("QUERY_FAILED", "failed to get today's bookings", err)
	}

	// Load related entities
	for _, booking := range bookings {
		if err := s.loadBookingRelations(ctx, booking); err != nil {
			s.logger.Warn("failed to load booking relations", "booking_id", booking.ID, "error", err)
		}
	}

	return dto.ToBookingResponses(bookings), nil
}

// GetBookingsInDateRange returns bookings within a date range
func (s *bookingService) GetBookingsInDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, filter dto.BookingFilter) (*dto.BookingListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if startDate.After(endDate) {
		return nil, errors.NewValidationError("start date cannot be after end date")
	}

	filter.TenantID = &tenantID
	filter.StartDate = &startDate
	filter.EndDate = &endDate

	return s.ListBookings(ctx, filter)
}

// GetPastDueBookings returns bookings that are past their scheduled time
func (s *bookingService) GetPastDueBookings(ctx context.Context, tenantID uuid.UUID) ([]*dto.BookingResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	bookings, err := s.repos.Booking.GetPastDueBookings(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("QUERY_FAILED", "failed to get past due bookings", err)
	}

	// Load related entities
	for _, booking := range bookings {
		if err := s.loadBookingRelations(ctx, booking); err != nil {
			s.logger.Warn("failed to load booking relations", "booking_id", booking.ID, "error", err)
		}
	}

	return dto.ToBookingResponses(bookings), nil
}

// GetBookingsNeedingReminders returns bookings that need reminders
func (s *bookingService) GetBookingsNeedingReminders(ctx context.Context, hoursAhead int) ([]*dto.BookingResponse, error) {
	if hoursAhead <= 0 {
		return nil, errors.NewValidationError("hours ahead must be positive")
	}

	bookings, err := s.repos.Booking.GetBookingsNeedingReminders(ctx, hoursAhead)
	if err != nil {
		return nil, errors.NewServiceError("QUERY_FAILED", "failed to get bookings needing reminders", err)
	}

	// Load related entities
	for _, booking := range bookings {
		if err := s.loadBookingRelations(ctx, booking); err != nil {
			s.logger.Warn("failed to load booking relations", "booking_id", booking.ID, "error", err)
		}
	}

	return dto.ToBookingResponses(bookings), nil
}

// ============================================================================
// Private Helper Methods for Scheduling
// ============================================================================

// getArtisanWorkingHours gets working hours for an artisan on a specific date
func (s *bookingService) getArtisanWorkingHours(ctx context.Context, artisanID uuid.UUID, date time.Time) (*dto.WorkingHoursResponse, error) {
	// In a real implementation, this would fetch from a working hours/schedule table
	// For now, return default business hours
	return &dto.WorkingHoursResponse{
		StartTime: "09:00",
		EndTime:   "17:00",
		TimeZone:  "UTC",
	}, nil
}

// getArtisanBookingsForDate gets all bookings for an artisan on a specific date
func (s *bookingService) getArtisanBookingsForDate(ctx context.Context, artisanID uuid.UUID, date time.Time) ([]*models.Booking, error) {
	bookings, err := s.repos.Booking.GetArtisanBookingsForDate(ctx, artisanID, date)
	if err != nil {
		return nil, err
	}

	// Filter out cancelled and no-show bookings
	validBookings := make([]*models.Booking, 0, len(bookings))
	for _, booking := range bookings {
		if booking.Status != models.BookingStatusCancelled && booking.Status != models.BookingStatusNoShow {
			validBookings = append(validBookings, booking)
		}
	}

	return validBookings, nil
}

// findBookingConflicts finds conflicts with existing bookings
func (s *bookingService) findBookingConflicts(req *dto.AvailabilityRequest, existingBookings []*models.Booking) []*dto.ConflictResponse {
	conflicts := make([]*dto.ConflictResponse, 0)

	requestEnd := req.Date.Add(time.Duration(req.Duration) * time.Minute)

	for _, booking := range existingBookings {
		if req.ExcludeBookingID != nil && booking.ID == *req.ExcludeBookingID {
			continue
		}

		if s.timePeriodsOverlap(req.Date, requestEnd, booking.StartTime, booking.EndTime) {
			conflicts = append(conflicts, &dto.ConflictResponse{
				ConflictType: "booking",
				StartTime:    booking.StartTime,
				EndTime:      booking.EndTime,
				BookingID:    &booking.ID,
				Reason:       fmt.Sprintf("Existing booking (Status: %s)", booking.Status),
			})
		}
	}

	return conflicts
}

// generateAvailableTimeSlots generates available time slots for a day
func (s *bookingService) generateAvailableTimeSlots(req *dto.AvailabilityRequest, workingHours *dto.WorkingHoursResponse, existingBookings []*models.Booking) []*dto.TimeSlotResponse {
	slots := make([]*dto.TimeSlotResponse, 0)

	// Parse working hours
	startTime, _ := time.Parse("15:04", workingHours.StartTime)
	endTime, _ := time.Parse("15:04", workingHours.EndTime)

	// Create start and end times for the requested date
	workStart := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, req.Date.Location())
	workEnd := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, req.Date.Location())

	// Generate 30-minute slots throughout the working day
	slotDuration := 30 * time.Minute
	current := workStart

	for current.Add(time.Duration(req.Duration)*time.Minute).Before(workEnd) ||
		current.Add(time.Duration(req.Duration)*time.Minute).Equal(workEnd) {

		slotEnd := current.Add(time.Duration(req.Duration) * time.Minute)
		available := true
		reason := ""

		// Check if this slot conflicts with existing bookings
		for _, booking := range existingBookings {
			if s.timePeriodsOverlap(current, slotEnd, booking.StartTime, booking.EndTime) {
				available = false
				reason = "Conflict with existing booking"
				break
			}
		}

		slots = append(slots, &dto.TimeSlotResponse{
			StartTime: current,
			EndTime:   slotEnd,
			Duration:  req.Duration,
			Available: available,
			Reason:    reason,
		})

		current = current.Add(slotDuration)
	}

	return slots
}

// timePeriodsOverlap checks if two time periods overlap
func (s *bookingService) timePeriodsOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}

// ============================================================================
// Recurring Booking Methods
// ============================================================================

// CreateRecurringBookings creates a series of recurring bookings
func (s *bookingService) CreateRecurringBookings(ctx context.Context, req *dto.CreateBookingRequest) ([]*dto.BookingResponse, error) {
	if !req.IsRecurring {
		return nil, errors.NewValidationError("request is not for recurring booking")
	}

	// Create the parent booking first
	parentBooking, err := s.CreateBooking(ctx, req)
	if err != nil {
		return nil, err
	}

	// Get the parent booking model to create recurring instances
	booking, err := s.repos.Booking.GetByID(ctx, parentBooking.ID)
	if err != nil {
		return nil, errors.NewServiceError("BOOKING_GET_FAILED", "failed to get parent booking", err)
	}

	// Create recurring instances
	recurringBookings, err := s.createRecurringBookings(ctx, booking, req)
	if err != nil {
		s.logger.Error("failed to create recurring bookings", "parent_id", booking.ID, "error", err)
		return []*dto.BookingResponse{parentBooking}, err
	}

	// Convert all bookings to responses
	responses := []*dto.BookingResponse{parentBooking}
	for _, rb := range recurringBookings {
		if err := s.loadBookingRelations(ctx, rb); err != nil {
			s.logger.Warn("failed to load booking relations", "booking_id", rb.ID, "error", err)
		}
		responses = append(responses, dto.ToBookingResponse(rb))
	}

	return responses, nil
}

// GetRecurringBookingSeries retrieves all bookings in a recurring series
func (s *bookingService) GetRecurringBookingSeries(ctx context.Context, parentBookingID uuid.UUID) ([]*dto.BookingResponse, error) {
	if parentBookingID == uuid.Nil {
		return nil, errors.NewValidationError("parent booking ID is required")
	}

	bookings, err := s.repos.Booking.GetRecurringBookings(ctx, parentBookingID)
	if err != nil {
		return nil, errors.NewServiceError("QUERY_FAILED", "failed to get recurring series", err)
	}

	// Load related entities
	for _, booking := range bookings {
		if err := s.loadBookingRelations(ctx, booking); err != nil {
			s.logger.Warn("failed to load booking relations", "booking_id", booking.ID, "error", err)
		}
	}

	return dto.ToBookingResponses(bookings), nil
}

// UpdateRecurringSeries updates all future bookings in a recurring series
func (s *bookingService) UpdateRecurringSeries(ctx context.Context, parentBookingID uuid.UUID, req *dto.UpdateBookingRequest, updateFuture bool) ([]*dto.BookingResponse, error) {
	if parentBookingID == uuid.Nil {
		return nil, errors.NewValidationError("parent booking ID is required")
	}

	// Get all bookings in the series
	bookings, err := s.repos.Booking.GetRecurringBookings(ctx, parentBookingID)
	if err != nil {
		return nil, errors.NewServiceError("QUERY_FAILED", "failed to get recurring series", err)
	}

	var updatedBookings []*dto.BookingResponse
	now := time.Now()

	for _, booking := range bookings {
		// If updateFuture is true, only update future bookings
		if updateFuture && booking.StartTime.Before(now) {
			continue
		}

		// Skip completed or cancelled bookings
		if booking.Status == models.BookingStatusCompleted || booking.Status == models.BookingStatusCancelled {
			continue
		}

		// Update the booking
		updated, err := s.UpdateBooking(ctx, booking.ID, req)
		if err != nil {
			s.logger.Error("failed to update booking in series", "booking_id", booking.ID, "error", err)
			continue
		}

		updatedBookings = append(updatedBookings, updated)
	}

	return updatedBookings, nil
}

// CancelRecurringSeries cancels all future bookings in a recurring series
func (s *bookingService) CancelRecurringSeries(ctx context.Context, parentBookingID uuid.UUID, reason string, cancelFuture bool) error {
	if parentBookingID == uuid.Nil {
		return errors.NewValidationError("parent booking ID is required")
	}

	if err := s.repos.Booking.CancelRecurringSeries(ctx, parentBookingID, reason); err != nil {
		return errors.NewServiceError("CANCEL_FAILED", "failed to cancel recurring series", err)
	}

	s.logger.Info("recurring series cancelled", "parent_booking_id", parentBookingID)
	return nil
}

// ============================================================================
// Photo Management Methods
// ============================================================================

// AddBeforePhotos adds before photos to a booking
func (s *bookingService) AddBeforePhotos(ctx context.Context, bookingID uuid.UUID, photoURLs []string) (*dto.BookingResponse, error) {
	if bookingID == uuid.Nil {
		return nil, errors.NewValidationError("booking ID is required")
	}
	if len(photoURLs) == 0 {
		return nil, errors.NewValidationError("at least one photo URL is required")
	}

	if err := s.repos.Booking.AddBeforePhotos(ctx, bookingID, photoURLs); err != nil {
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to add before photos", err)
	}

	return s.GetBooking(ctx, bookingID)
}

// AddAfterPhotos adds after photos to a booking
func (s *bookingService) AddAfterPhotos(ctx context.Context, bookingID uuid.UUID, photoURLs []string) (*dto.BookingResponse, error) {
	if bookingID == uuid.Nil {
		return nil, errors.NewValidationError("booking ID is required")
	}
	if len(photoURLs) == 0 {
		return nil, errors.NewValidationError("at least one photo URL is required")
	}

	if err := s.repos.Booking.AddAfterPhotos(ctx, bookingID, photoURLs); err != nil {
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to add after photos", err)
	}

	return s.GetBooking(ctx, bookingID)
}

// ============================================================================
// Payment Integration Methods
// ============================================================================

// UpdatePaymentStatus updates the payment status of a booking
func (s *bookingService) UpdatePaymentStatus(ctx context.Context, bookingID uuid.UUID, status models.PaymentStatus) (*dto.BookingResponse, error) {
	if bookingID == uuid.Nil {
		return nil, errors.NewValidationError("booking ID is required")
	}

	if err := s.repos.Booking.UpdatePaymentStatus(ctx, bookingID, status); err != nil {
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to update payment status", err)
	}

	s.logger.Info("payment status updated", "booking_id", bookingID, "status", status)
	return s.GetBooking(ctx, bookingID)
}

// RecordDepositPayment records a deposit payment for a booking
func (s *bookingService) RecordDepositPayment(ctx context.Context, bookingID uuid.UUID, amount float64, paymentIntentID string) (*dto.BookingResponse, error) {
	if bookingID == uuid.Nil {
		return nil, errors.NewValidationError("booking ID is required")
	}
	if amount <= 0 {
		return nil, errors.NewValidationError("amount must be positive")
	}

	// Record the deposit
	if err := s.repos.Booking.RecordDepositPayment(ctx, bookingID, amount); err != nil {
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to record deposit", err)
	}

	// Update payment intent ID
	if paymentIntentID != "" {
		if err := s.repos.Booking.UpdatePaymentIntent(ctx, bookingID, paymentIntentID); err != nil {
			s.logger.Warn("failed to update payment intent", "booking_id", bookingID, "error", err)
		}
	}

	// Update payment status - deposit has been paid but full payment is pending
	status := models.PaymentStatusPending
	updateReq := &dto.UpdateBookingRequest{
		PaymentStatus: &status,
	}

	return s.UpdateBooking(ctx, bookingID, updateReq)
}

// ProcessRefund processes a refund for a cancelled booking
func (s *bookingService) ProcessRefund(ctx context.Context, bookingID uuid.UUID, amount float64, reason string) (*dto.BookingResponse, error) {
	if bookingID == uuid.Nil {
		return nil, errors.NewValidationError("booking ID is required")
	}
	if amount <= 0 {
		return nil, errors.NewValidationError("amount must be positive")
	}

	booking, err := s.repos.Booking.GetByID(ctx, bookingID)
	if err != nil {
		return nil, errors.NewServiceError("BOOKING_GET_FAILED", "failed to get booking", err)
	}

	// Validate refund amount
	maxRefund := booking.CalculateRefundAmount()
	if amount > maxRefund {
		return nil, errors.NewValidationError(fmt.Sprintf("refund amount exceeds maximum refundable amount (%.2f)", maxRefund))
	}

	// Process refund through payment service
	if s.paymentService != nil {
		// This would integrate with the payment service
		// For now, we'll just update the booking status
		s.logger.Info("processing refund", "booking_id", bookingID, "amount", amount, "reason", reason)
	}

	// Update booking with refund information
	refundStatus := models.PaymentStatusRefunded
	updateReq := &dto.UpdateBookingRequest{
		PaymentStatus: &refundStatus,
	}

	return s.UpdateBooking(ctx, bookingID, updateReq)
}

// ============================================================================
// Payment Integration and Status Management
// ============================================================================

// ============================================================================
// Analytics & Reporting Methods
// ============================================================================

// GetBookingStats retrieves comprehensive booking statistics for a tenant
func (s *bookingService) GetBookingStats(ctx context.Context, tenantID uuid.UUID) (*dto.BookingStatsResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	stats, err := s.repos.Booking.GetBookingStats(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("STATS_FAILED", "failed to get booking stats", err)
	}

	// Convert repository stats to DTO stats
	response := &dto.BookingStatsResponse{
		TotalBookings:       stats.TotalBookings,
		PendingBookings:     stats.PendingBookings,
		ConfirmedBookings:   stats.ConfirmedBookings,
		InProgressBookings:  stats.ByStatus[models.BookingStatusInProgress],
		CompletedBookings:   stats.CompletedBookings,
		CancelledBookings:   stats.CancelledBookings,
		NoShowBookings:      stats.NoShowBookings,
		TotalRevenue:        stats.TotalRevenue,
		AverageBookingValue: stats.AverageBookingValue,
	}

	// Calculate rates
	if stats.TotalBookings > 0 {
		response.CompletionRate = float64(stats.CompletedBookings) / float64(stats.TotalBookings) * 100
		response.CancellationRate = float64(stats.CancelledBookings) / float64(stats.TotalBookings) * 100
		response.NoShowRate = float64(stats.NoShowBookings) / float64(stats.TotalBookings) * 100
	}

	return response, nil
}

// GetBookingAnalytics retrieves detailed analytics based on filters
func (s *bookingService) GetBookingAnalytics(ctx context.Context, filter dto.BookingAnalyticsFilter) (*dto.BookingStatsResponse, error) {
	if err := filter.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid analytics filter: " + err.Error())
	}

	// Get basic stats
	stats, err := s.GetBookingStats(ctx, filter.TenantID)
	if err != nil {
		return nil, err
	}

	// Get trend data
	trends, err := s.repos.Booking.GetBookingTrends(ctx, filter.TenantID, 30)
	if err != nil {
		s.logger.Warn("failed to get booking trends", "error", err)
	} else {
		stats.BookingTrends = make([]dto.BookingTrendData, len(trends))
		for i, trend := range trends {
			stats.BookingTrends[i] = dto.BookingTrendData{
				Date:     trend.Date,
				Bookings: trend.BookingCount,
				Revenue:  trend.Revenue,
			}
		}
	}

	// Get popular services
	popularServices, err := s.repos.Booking.GetPopularServices(ctx, filter.TenantID, 10, filter.StartDate, filter.EndDate)
	if err != nil {
		s.logger.Warn("failed to get popular services", "error", err)
	} else {
		stats.PopularServices = make([]dto.PopularServiceData, len(popularServices))
		for i, svc := range popularServices {
			stats.PopularServices[i] = dto.PopularServiceData{
				ServiceID:     svc.ServiceID,
				ServiceName:   svc.ServiceName,
				BookingCount:  svc.Count,
				TotalRevenue:  svc.Revenue,
				AverageRating: 0, // Would need to calculate from reviews
			}
		}
	}

	return stats, nil
}

// GetArtisanBookingStats retrieves booking statistics for a specific artisan
func (s *bookingService) GetArtisanBookingStats(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (*dto.BookingStatsResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan ID is required")
	}

	stats, err := s.repos.Booking.GetArtisanBookingStats(ctx, artisanID, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("STATS_FAILED", "failed to get artisan stats", err)
	}

	response := &dto.BookingStatsResponse{
		TotalBookings:       stats.TotalBookings,
		CompletedBookings:   stats.CompletedBookings,
		TotalRevenue:        stats.TotalRevenue,
		AverageBookingValue: stats.AverageBookingValue,
		CompletionRate:      stats.UtilizationRate,
	}

	return response, nil
}

// GetPopularServices retrieves most popular services by booking count
func (s *bookingService) GetPopularServices(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]*dto.PopularServiceData, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if limit <= 0 {
		limit = 10
	}

	services, err := s.repos.Booking.GetPopularServices(ctx, tenantID, limit, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("QUERY_FAILED", "failed to get popular services", err)
	}

	result := make([]*dto.PopularServiceData, len(services))
	for i, svc := range services {
		result[i] = &dto.PopularServiceData{
			ServiceID:     svc.ServiceID,
			ServiceName:   svc.ServiceName,
			BookingCount:  svc.Count,
			TotalRevenue:  svc.Revenue,
			AverageRating: 0, // Would need to fetch from reviews
		}
	}

	return result, nil
}

// GetBookingTrends retrieves booking trends over time
func (s *bookingService) GetBookingTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.BookingTrendData, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if days <= 0 {
		days = 30
	}

	trends, err := s.repos.Booking.GetBookingTrends(ctx, tenantID, days)
	if err != nil {
		return nil, errors.NewServiceError("QUERY_FAILED", "failed to get booking trends", err)
	}

	result := make([]*dto.BookingTrendData, len(trends))
	for i, trend := range trends {
		result[i] = &dto.BookingTrendData{
			Date:     trend.Date,
			Bookings: trend.BookingCount,
			Revenue:  trend.Revenue,
		}
	}

	return result, nil
}

// ============================================================================
// Bulk Operations Methods
// ============================================================================

// BulkConfirm confirms multiple bookings
func (s *bookingService) BulkConfirm(ctx context.Context, bookingIDs []uuid.UUID) ([]*dto.BookingResponse, error) {
	if len(bookingIDs) == 0 {
		return nil, errors.NewValidationError("at least one booking ID is required")
	}

	if err := s.repos.Booking.BulkConfirm(ctx, bookingIDs); err != nil {
		return nil, errors.NewServiceError("BULK_UPDATE_FAILED", "failed to bulk confirm bookings", err)
	}

	// Fetch updated bookings
	var responses []*dto.BookingResponse
	for _, id := range bookingIDs {
		booking, err := s.GetBooking(ctx, id)
		if err != nil {
			s.logger.Warn("failed to get booking after bulk confirm", "booking_id", id, "error", err)
			continue
		}
		responses = append(responses, booking)
	}

	s.logger.Info("bookings bulk confirmed", "count", len(bookingIDs))
	return responses, nil
}

// BulkCancel cancels multiple bookings
func (s *bookingService) BulkCancel(ctx context.Context, bookingIDs []uuid.UUID, reason string) ([]*dto.BookingResponse, error) {
	if len(bookingIDs) == 0 {
		return nil, errors.NewValidationError("at least one booking ID is required")
	}
	if reason == "" {
		return nil, errors.NewValidationError("cancellation reason is required")
	}

	if err := s.repos.Booking.BulkCancel(ctx, bookingIDs, reason); err != nil {
		return nil, errors.NewServiceError("BULK_UPDATE_FAILED", "failed to bulk cancel bookings", err)
	}

	// Fetch updated bookings
	var responses []*dto.BookingResponse
	for _, id := range bookingIDs {
		booking, err := s.GetBooking(ctx, id)
		if err != nil {
			s.logger.Warn("failed to get booking after bulk cancel", "booking_id", id, "error", err)
			continue
		}
		responses = append(responses, booking)
	}

	s.logger.Info("bookings bulk cancelled", "count", len(bookingIDs))
	return responses, nil
}

// BulkReschedule reschedules multiple bookings to a new time
func (s *bookingService) BulkReschedule(ctx context.Context, bookingIDs []uuid.UUID, newStartTime time.Time) ([]*dto.BookingResponse, error) {
	if len(bookingIDs) == 0 {
		return nil, errors.NewValidationError("at least one booking ID is required")
	}
	if newStartTime.IsZero() {
		return nil, errors.NewValidationError("new start time is required")
	}

	var responses []*dto.BookingResponse

	// Reschedule each booking individually
	for _, id := range bookingIDs {
		_, err := s.repos.Booking.GetByID(ctx, id)
		if err != nil {
			s.logger.Warn("failed to get booking for reschedule", "booking_id", id, "error", err)
			continue
		}

		rescheduleReq := &dto.RescheduleBookingRequest{
			NewStartTime:   newStartTime,
			Reason:         "Bulk rescheduled",
			NotifyCustomer: true,
			NotifyArtisan:  true,
		}

		rescheduled, err := s.RescheduleBooking(ctx, id, rescheduleReq)
		if err != nil {
			s.logger.Error("failed to reschedule booking", "booking_id", id, "error", err)
			continue
		}

		responses = append(responses, rescheduled)
	}

	s.logger.Info("bookings bulk rescheduled", "count", len(responses))
	return responses, nil
}

// BulkUpdateStatus updates the status of multiple bookings
func (s *bookingService) BulkUpdateStatus(ctx context.Context, bookingIDs []uuid.UUID, status models.BookingStatus) ([]*dto.BookingResponse, error) {
	if len(bookingIDs) == 0 {
		return nil, errors.NewValidationError("at least one booking ID is required")
	}

	var responses []*dto.BookingResponse

	// Update each booking individually to maintain validation
	for _, id := range bookingIDs {
		updateReq := &dto.UpdateBookingRequest{
			Status: &status,
		}

		updated, err := s.UpdateBooking(ctx, id, updateReq)
		if err != nil {
			s.logger.Error("failed to update booking status", "booking_id", id, "error", err)
			continue
		}

		responses = append(responses, updated)
	}

	s.logger.Info("bookings bulk status updated", "count", len(responses), "status", status)
	return responses, nil
}

// ============================================================================
// Notification Integration Methods
// ============================================================================

// NotifyBookingCreated sends notifications when a booking is created
func (s *bookingService) NotifyBookingCreated(ctx context.Context, booking *models.Booking) error {
	// TODO: Integrate with notification service
	s.logger.Info("booking created notification", "booking_id", booking.ID, "customer_id", booking.CustomerID, "artisan_id", booking.ArtisanID)
	return nil
}

// NotifyBookingUpdated sends notifications when a booking is updated
func (s *bookingService) NotifyBookingUpdated(ctx context.Context, booking *models.Booking, oldStatus models.BookingStatus) error {
	// TODO: Integrate with notification service
	s.logger.Info("booking updated notification", "booking_id", booking.ID, "old_status", oldStatus, "new_status", booking.Status)
	return nil
}

// NotifyBookingCancelled sends notifications when a booking is cancelled
func (s *bookingService) NotifyBookingCancelled(ctx context.Context, booking *models.Booking) error {
	// TODO: Integrate with notification service
	s.logger.Info("booking cancelled notification", "booking_id", booking.ID, "reason", booking.CancellationReason)
	return nil
}

// UpdateCustomerStatistics updates customer statistics after a booking event
func (s *bookingService) UpdateCustomerStatistics(ctx context.Context, customerID uuid.UUID, bookingValue float64, loyaltyPoints int) error {
	// TODO: Integrate with customer service to update statistics
	s.logger.Info("updating customer statistics", "customer_id", customerID, "value", bookingValue, "loyalty_points", loyaltyPoints)
	return nil
}

// updateCustomerStats is a helper for updating customer statistics
func (s *bookingService) updateCustomerStats(ctx context.Context, customerID uuid.UUID, eventType string) error {
	return s.UpdateCustomerStatistics(ctx, customerID, 0, 0)
}

// GetServiceMetrics returns service metrics
func (s *bookingService) GetServiceMetrics(ctx context.Context) map[string]interface{} {
	totalBookings, _ := s.repos.Booking.Count(ctx, map[string]any{})

	return map[string]interface{}{
		"total_bookings": totalBookings,
		"service_status": "healthy",
	}
}
