package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type BookingRepository interface {
	BaseRepository[models.Booking]

	// Core Operations
	GetByArtisanID(ctx context.Context, artisanID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetByServiceID(ctx context.Context, serviceID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)

	// Status Operations
	UpdateStatus(ctx context.Context, bookingID uuid.UUID, status models.BookingStatus) error
	ConfirmBooking(ctx context.Context, bookingID uuid.UUID) error
	StartBooking(ctx context.Context, bookingID uuid.UUID) error
	CompleteBooking(ctx context.Context, bookingID uuid.UUID) error
	CancelBooking(ctx context.Context, bookingID uuid.UUID, cancelledBy uuid.UUID, reason string) error
	MarkAsNoShow(ctx context.Context, bookingID uuid.UUID) error

	// Status Queries
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status models.BookingStatus, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetPendingBookings(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetConfirmedBookings(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetCompletedBookings(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetCancelledBookings(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)

	// Time-based Queries
	GetUpcomingBookings(ctx context.Context, tenantID uuid.UUID, days int, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetTodayBookings(ctx context.Context, tenantID uuid.UUID) ([]*models.Booking, error)
	GetBookingsInRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetBookingsForDate(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]*models.Booking, error)
	GetPastDueBookings(ctx context.Context, tenantID uuid.UUID) ([]*models.Booking, error)

	// Artisan Schedule
	GetArtisanBookingsForDate(ctx context.Context, artisanID uuid.UUID, date time.Time) ([]*models.Booking, error)
	GetArtisanBookingsInRange(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) ([]*models.Booking, error)
	GetArtisanAvailableSlots(ctx context.Context, artisanID uuid.UUID, date time.Time, duration int) ([]TimeSlot, error)
	CheckArtisanAvailability(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time) (bool, error)
	HasOverlappingBookings(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time, excludeBookingID *uuid.UUID) (bool, error)

	// Customer Operations
	GetCustomerUpcomingBookings(ctx context.Context, customerID uuid.UUID) ([]*models.Booking, error)
	GetCustomerBookingHistory(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	GetCustomerBookingCount(ctx context.Context, customerID uuid.UUID) (int64, error)

	// Payment Operations
	UpdatePaymentStatus(ctx context.Context, bookingID uuid.UUID, status models.PaymentStatus) error
	RecordDepositPayment(ctx context.Context, bookingID uuid.UUID, amount float64) error
	UpdatePaymentIntent(ctx context.Context, bookingID uuid.UUID, paymentIntentID string) error
	GetUnpaidBookings(ctx context.Context, tenantID uuid.UUID) ([]*models.Booking, error)

	// Recurrence Operations
	CreateRecurringBookings(ctx context.Context, parentBooking *models.Booking, occurrences int) ([]*models.Booking, error)
	GetRecurringBookings(ctx context.Context, parentBookingID uuid.UUID) ([]*models.Booking, error)
	CancelRecurringSeries(ctx context.Context, parentBookingID uuid.UUID, reason string) error

	// Photo Operations
	AddBeforePhotos(ctx context.Context, bookingID uuid.UUID, photoURLs []string) error
	AddAfterPhotos(ctx context.Context, bookingID uuid.UUID, photoURLs []string) error

	// Reminder Operations
	MarkReminderSent(ctx context.Context, bookingID uuid.UUID, reminderType string) error
	GetBookingsNeedingReminders(ctx context.Context, hoursAhead int) ([]*models.Booking, error)

	// Analytics & Reporting
	GetBookingStats(ctx context.Context, tenantID uuid.UUID) (BookingStats, error)
	GetBookingsByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]BookingPeriodData, error)
	GetArtisanBookingStats(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (ArtisanBookingStats, error)
	GetPopularServices(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]ServiceBookingCount, error)
	GetBookingTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]BookingTrend, error)
	GetAverageBookingValue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error)
	GetUtilizationRate(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (float64, error)
	GetCancellationRate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error)
	GetNoShowRate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error)

	// Search & Filter
	Search(ctx context.Context, query string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)
	FindByFilters(ctx context.Context, filters BookingFilters, pagination PaginationParams) ([]*models.Booking, PaginationResult, error)

	// Bulk Operations
	BulkConfirm(ctx context.Context, bookingIDs []uuid.UUID) error
	BulkCancel(ctx context.Context, bookingIDs []uuid.UUID, reason string) error
	BulkUpdatePaymentStatus(ctx context.Context, bookingIDs []uuid.UUID, status models.PaymentStatus) error
}

// TimeSlot represents an available time slot
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// BookingStats represents comprehensive booking statistics
type BookingStats struct {
	TotalBookings       int64                          `json:"total_bookings"`
	PendingBookings     int64                          `json:"pending_bookings"`
	ConfirmedBookings   int64                          `json:"confirmed_bookings"`
	CompletedBookings   int64                          `json:"completed_bookings"`
	CancelledBookings   int64                          `json:"cancelled_bookings"`
	NoShowBookings      int64                          `json:"no_show_bookings"`
	TotalRevenue        float64                        `json:"total_revenue"`
	AverageBookingValue float64                        `json:"average_booking_value"`
	ByStatus            map[models.BookingStatus]int64 `json:"by_status"`
	ThisMonthBookings   int64                          `json:"this_month_bookings"`
	LastMonthBookings   int64                          `json:"last_month_bookings"`
	ThisMonthRevenue    float64                        `json:"this_month_revenue"`
	LastMonthRevenue    float64                        `json:"last_month_revenue"`
}

// BookingPeriodData represents booking data for a period
type BookingPeriodData struct {
	Period         time.Time `json:"period"`
	BookingCount   int64     `json:"booking_count"`
	Revenue        float64   `json:"revenue"`
	CompletedCount int64     `json:"completed_count"`
	AverageValue   float64   `json:"average_value"`
}

// ArtisanBookingStats represents booking statistics for an artisan
type ArtisanBookingStats struct {
	ArtisanID           uuid.UUID                      `json:"artisan_id"`
	TotalBookings       int64                          `json:"total_bookings"`
	CompletedBookings   int64                          `json:"completed_bookings"`
	TotalRevenue        float64                        `json:"total_revenue"`
	AverageBookingValue float64                        `json:"average_booking_value"`
	UtilizationRate     float64                        `json:"utilization_rate"`
	ByStatus            map[models.BookingStatus]int64 `json:"by_status"`
	StartDate           time.Time                      `json:"start_date"`
	EndDate             time.Time                      `json:"end_date"`
}

// ServiceBookingCount represents booking count per service
type ServiceBookingCount struct {
	ServiceID    uuid.UUID `json:"service_id"`
	ServiceName  string    `json:"service_name"`
	Count        int64     `json:"count"`
	Revenue      float64   `json:"revenue"`
	AverageValue float64   `json:"average_value"`
}

// BookingTrend represents booking trends over time
type BookingTrend struct {
	Date           time.Time `json:"date"`
	BookingCount   int64     `json:"booking_count"`
	CompletedCount int64     `json:"completed_count"`
	Revenue        float64   `json:"revenue"`
}

// BookingFilters for advanced filtering
type BookingFilters struct {
	TenantID        uuid.UUID              `json:"tenant_id"`
	ArtisanIDs      []uuid.UUID            `json:"artisan_ids"`
	CustomerIDs     []uuid.UUID            `json:"customer_ids"`
	ServiceIDs      []uuid.UUID            `json:"service_ids"`
	Statuses        []models.BookingStatus `json:"statuses"`
	PaymentStatuses []models.PaymentStatus `json:"payment_statuses"`
	StartDateFrom   *time.Time             `json:"start_date_from"`
	StartDateTo     *time.Time             `json:"start_date_to"`
	MinPrice        *float64               `json:"min_price"`
	MaxPrice        *float64               `json:"max_price"`
	IsRecurring     *bool                  `json:"is_recurring"`
}

type bookingRepository struct {
	BaseRepository[models.Booking]
	db      *gorm.DB
	logger  log.AllLogger
	cache   Cache
	metrics MetricsCollector
}

func NewBookingRepository(db *gorm.DB, config ...RepositoryConfig) BookingRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 2 * time.Minute // Short cache for bookings
	}

	baseRepo := NewBaseRepository[models.Booking](db, cfg)

	return &bookingRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
		metrics:        cfg.Metrics,
	}
}

//------------------------------------------------------------
// Core Operations
//------------------------------------------------------------

func (r *bookingRepository) GetByArtisanID(ctx context.Context, artisanID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Booking{}).Where("artisan_id = ?", artisanID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	var bookings []*models.Booking
	if err := query.
		Preload("Customer").
		Preload("Service").
		Preload("Payments").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time DESC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

func (r *bookingRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Booking{}).Where("customer_id = ?", customerID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	var bookings []*models.Booking
	if err := query.
		Preload("Artisan").
		Preload("Service").
		Preload("Payments").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time DESC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

func (r *bookingRepository) GetByServiceID(ctx context.Context, serviceID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Booking{}).Where("service_id = ?", serviceID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	var bookings []*models.Booking
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Payments").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time DESC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

func (r *bookingRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Booking{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	var bookings []*models.Booking
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Preload("Payments").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time DESC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

//------------------------------------------------------------
// Status Operations
//------------------------------------------------------------

func (r *bookingRepository) UpdateStatus(ctx context.Context, bookingID uuid.UUID, status models.BookingStatus) error {
	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Update("status", status)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "booking not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, bookingID)
	return nil
}

func (r *bookingRepository) ConfirmBooking(ctx context.Context, bookingID uuid.UUID) error {
	return r.UpdateStatus(ctx, bookingID, models.BookingStatusConfirmed)
}

func (r *bookingRepository) StartBooking(ctx context.Context, bookingID uuid.UUID) error {
	return r.UpdateStatus(ctx, bookingID, models.BookingStatusInProgress)
}

func (r *bookingRepository) CompleteBooking(ctx context.Context, bookingID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Updates(map[string]any{
			"status":       models.BookingStatusCompleted,
			"completed_at": &now,
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to complete booking", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "booking not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, bookingID)
	return nil
}

func (r *bookingRepository) CancelBooking(ctx context.Context, bookingID uuid.UUID, cancelledBy uuid.UUID, reason string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Updates(map[string]any{
			"status":              models.BookingStatusCancelled,
			"cancelled_at":        &now,
			"cancelled_by":        &cancelledBy,
			"cancellation_reason": reason,
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to cancel booking", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "booking not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, bookingID)
	return nil
}

func (r *bookingRepository) MarkAsNoShow(ctx context.Context, bookingID uuid.UUID) error {
	return r.UpdateStatus(ctx, bookingID, models.BookingStatusNoShow)
}

//------------------------------------------------------------
// Status Queries
//------------------------------------------------------------

func (r *bookingRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status models.BookingStatus, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND status = ?", tenantID, status)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	var bookings []*models.Booking
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time DESC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

func (r *bookingRepository) GetPendingBookings(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	return r.GetByStatus(ctx, tenantID, models.BookingStatusPending, pagination)
}

func (r *bookingRepository) GetConfirmedBookings(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	return r.GetByStatus(ctx, tenantID, models.BookingStatusConfirmed, pagination)
}

func (r *bookingRepository) GetCompletedBookings(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	countQuery := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.BookingStatusCompleted)
	countQuery = r.applyDateRange(countQuery, "completed_at", startDate, endDate)

	var totalItems int64
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	dataQuery := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.BookingStatusCompleted)
	dataQuery = r.applyDateRange(dataQuery, "completed_at", startDate, endDate)

	var bookings []*models.Booking
	if err := dataQuery.
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("completed_at DESC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

func (r *bookingRepository) GetCancelledBookings(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	return r.GetByStatus(ctx, tenantID, models.BookingStatusCancelled, pagination)
}

//------------------------------------------------------------
// Time-based Queries
//------------------------------------------------------------

func (r *bookingRepository) GetUpcomingBookings(ctx context.Context, tenantID uuid.UUID, days int, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	now := time.Now()
	deadline := now.AddDate(0, 0, days)

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND start_time BETWEEN ? AND ? AND status IN ?",
			tenantID, now, deadline,
			[]models.BookingStatus{models.BookingStatusPending, models.BookingStatusConfirmed})

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	var bookings []*models.Booking
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

func (r *bookingRepository) GetTodayBookings(ctx context.Context, tenantID uuid.UUID) ([]*models.Booking, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var bookings []*models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Where("tenant_id = ? AND start_time >= ? AND start_time < ?", tenantID, startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find today's bookings", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetBookingsInRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	countQuery := r.db.WithContext(ctx).Model(&models.Booking{}).Where("tenant_id = ?", tenantID)
	countQuery = r.applyDateRange(countQuery, "start_time", startDate, endDate)

	var totalItems int64
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	dataQuery := r.db.WithContext(ctx).Model(&models.Booking{}).Where("tenant_id = ?", tenantID)
	dataQuery = r.applyDateRange(dataQuery, "start_time", startDate, endDate)

	var bookings []*models.Booking
	if err := dataQuery.
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

func (r *bookingRepository) GetBookingsForDate(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]*models.Booking, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var bookings []*models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Where("tenant_id = ? AND start_time >= ? AND start_time < ?", tenantID, startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings for date", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetPastDueBookings(ctx context.Context, tenantID uuid.UUID) ([]*models.Booking, error) {
	now := time.Now()

	var bookings []*models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Where("tenant_id = ? AND start_time < ? AND status IN ?",
			tenantID, now,
			[]models.BookingStatus{models.BookingStatusPending, models.BookingStatusConfirmed}).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find past due bookings", err)
	}

	return bookings, nil
}

//------------------------------------------------------------
// Artisan Schedule
//------------------------------------------------------------

func (r *bookingRepository) GetArtisanBookingsForDate(ctx context.Context, artisanID uuid.UUID, date time.Time) ([]*models.Booking, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var bookings []*models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Service").
		Where("artisan_id = ? AND start_time >= ? AND start_time < ?", artisanID, startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find artisan bookings", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetArtisanBookingsInRange(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) ([]*models.Booking, error) {
	query := r.db.WithContext(ctx).Model(&models.Booking{}).Where("artisan_id = ?", artisanID)
	query = r.applyDateRange(query, "start_time", startDate, endDate)

	var bookings []*models.Booking
	if err := query.
		Preload("Customer").
		Preload("Service").
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find artisan bookings", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetArtisanAvailableSlots(ctx context.Context, artisanID uuid.UUID, date time.Time, duration int) ([]TimeSlot, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	bookings, err := r.GetArtisanBookingsForDate(ctx, artisanID, date)
	if err != nil {
		return nil, err
	}

	slots := []TimeSlot{}
	currentTime := startOfDay.Add(9 * time.Hour)
	slotDuration := time.Duration(duration) * time.Minute

	for currentTime.Add(slotDuration).Before(endOfDay) {
		slotEnd := currentTime.Add(slotDuration)

		isAvailable := true
		for _, booking := range bookings {
			if booking.Status == models.BookingStatusCancelled || booking.Status == models.BookingStatusNoShow {
				continue
			}
			if currentTime.Before(booking.EndTime) && slotEnd.After(booking.StartTime) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			slots = append(slots, TimeSlot{
				StartTime: currentTime,
				EndTime:   slotEnd,
			})
		}

		currentTime = currentTime.Add(30 * time.Minute)
	}

	return slots, nil
}

func (r *bookingRepository) CheckArtisanAvailability(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	hasOverlap, err := r.HasOverlappingBookings(ctx, artisanID, startTime, endTime, nil)
	if err != nil {
		return false, err
	}
	return !hasOverlap, nil
}

func (r *bookingRepository) HasOverlappingBookings(ctx context.Context, artisanID uuid.UUID, startTime, endTime time.Time, excludeBookingID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("artisan_id = ? AND status NOT IN ? AND ((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?))",
			artisanID,
			[]models.BookingStatus{models.BookingStatusCancelled, models.BookingStatusNoShow},
			endTime, startTime,
			startTime, endTime,
			startTime, endTime)

	if excludeBookingID != nil {
		query = query.Where("id != ?", *excludeBookingID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, errors.NewRepositoryError("CHECK_FAILED", "failed to check overlap", err)
	}

	return count > 0, nil
}

//------------------------------------------------------------
// Customer Operations
//------------------------------------------------------------

func (r *bookingRepository) GetCustomerUpcomingBookings(ctx context.Context, customerID uuid.UUID) ([]*models.Booking, error) {
	now := time.Now()

	var bookings []*models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Artisan").
		Preload("Service").
		Where("customer_id = ? AND start_time > ? AND status IN ?",
			customerID, now,
			[]models.BookingStatus{models.BookingStatusPending, models.BookingStatusConfirmed}).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find upcoming bookings", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetCustomerBookingHistory(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	return r.GetByCustomerID(ctx, customerID, pagination)
}

func (r *bookingRepository) GetCustomerBookingCount(ctx context.Context, customerID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("customer_id = ?", customerID).
		Count(&count).Error; err != nil {
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}
	return count, nil
}

//------------------------------------------------------------
// Payment Operations
//------------------------------------------------------------

func (r *bookingRepository) UpdatePaymentStatus(ctx context.Context, bookingID uuid.UUID, status models.PaymentStatus) error {
	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Update("payment_status", status)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update payment status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "booking not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, bookingID)
	return nil
}

func (r *bookingRepository) RecordDepositPayment(ctx context.Context, bookingID uuid.UUID, amount float64) error {
	if amount <= 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "amount must be positive", errors.ErrInvalidInput)
	}

	booking, err := r.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	newDeposit := booking.DepositPaid + amount
	if newDeposit > booking.TotalPrice {
		return errors.NewRepositoryError("INVALID_INPUT", "deposit exceeds total price", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Update("deposit_paid", newDeposit)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to record deposit", result.Error)
	}

	r.InvalidateCache(ctx, bookingID)
	return nil
}

func (r *bookingRepository) UpdatePaymentIntent(ctx context.Context, bookingID uuid.UUID, paymentIntentID string) error {
	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Update("payment_intent_id", paymentIntentID)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update payment intent", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "booking not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, bookingID)
	return nil
}

func (r *bookingRepository) GetUnpaidBookings(ctx context.Context, tenantID uuid.UUID) ([]*models.Booking, error) {
	var bookings []*models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Where("tenant_id = ? AND payment_status = ?", tenantID, models.PaymentStatusPending).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find unpaid bookings", err)
	}

	return bookings, nil
}

//------------------------------------------------------------
// Recurrence Operations
//------------------------------------------------------------

func (r *bookingRepository) CreateRecurringBookings(ctx context.Context, parentBooking *models.Booking, occurrences int) ([]*models.Booking, error) {
	if occurrences <= 0 {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "occurrences must be positive", errors.ErrInvalidInput)
	}

	bookings := make([]*models.Booking, 0, occurrences)
	currentStart := parentBooking.StartTime
	currentEnd := parentBooking.EndTime

	var interval time.Duration
	switch parentBooking.RecurrencePattern {
	case "weekly":
		interval = 7 * 24 * time.Hour
	case "biweekly":
		interval = 14 * 24 * time.Hour
	case "monthly":
		interval = 30 * 24 * time.Hour
	default:
		return nil, errors.NewRepositoryError("INVALID_INPUT", "invalid recurrence pattern", errors.ErrInvalidInput)
	}

	for i := 0; i < occurrences; i++ {
		booking := &models.Booking{
			TenantID:          parentBooking.TenantID,
			ArtisanID:         parentBooking.ArtisanID,
			CustomerID:        parentBooking.CustomerID,
			ServiceID:         parentBooking.ServiceID,
			StartTime:         currentStart,
			EndTime:           currentEnd,
			Duration:          parentBooking.Duration,
			Status:            models.BookingStatusPending,
			PaymentStatus:     parentBooking.PaymentStatus,
			BasePrice:         parentBooking.BasePrice,
			AddonsPrice:       parentBooking.AddonsPrice,
			TotalPrice:        parentBooking.TotalPrice,
			Currency:          parentBooking.Currency,
			Notes:             parentBooking.Notes,
			IsRecurring:       true,
			RecurrencePattern: parentBooking.RecurrencePattern,
			ParentBookingID:   &parentBooking.ID,
		}

		if err := r.Create(ctx, booking); err != nil {
			return nil, err
		}

		bookings = append(bookings, booking)
		currentStart = currentStart.Add(interval)
		currentEnd = currentEnd.Add(interval)
	}

	return bookings, nil
}

func (r *bookingRepository) GetRecurringBookings(ctx context.Context, parentBookingID uuid.UUID) ([]*models.Booking, error) {
	var bookings []*models.Booking
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Where("parent_booking_id = ?", parentBookingID).
		Order("start_time ASC").
		Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find recurring bookings", err)
	}

	return bookings, nil
}

func (r *bookingRepository) CancelRecurringSeries(ctx context.Context, parentBookingID uuid.UUID, reason string) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("parent_booking_id = ? AND start_time > ? AND status NOT IN ?",
			parentBookingID, now,
			[]models.BookingStatus{models.BookingStatusCancelled, models.BookingStatusCompleted}).
		Updates(map[string]any{
			"status":              models.BookingStatusCancelled,
			"cancelled_at":        &now,
			"cancellation_reason": reason,
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to cancel recurring series", result.Error)
	}

	return nil
}

//------------------------------------------------------------
// Photo Operations
//------------------------------------------------------------

func (r *bookingRepository) AddBeforePhotos(ctx context.Context, bookingID uuid.UUID, photoURLs []string) error {
	booking, err := r.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	booking.BeforePhotoURLs = append(booking.BeforePhotoURLs, photoURLs...)

	if err := r.Update(ctx, booking); err != nil {
		return err
	}

	r.InvalidateCache(ctx, bookingID)
	return nil
}

func (r *bookingRepository) AddAfterPhotos(ctx context.Context, bookingID uuid.UUID, photoURLs []string) error {
	booking, err := r.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	booking.AfterPhotoURLs = append(booking.AfterPhotoURLs, photoURLs...)

	if err := r.Update(ctx, booking); err != nil {
		return err
	}

	r.InvalidateCache(ctx, bookingID)
	return nil
}

//------------------------------------------------------------
// Reminder Operations
//------------------------------------------------------------

func (r *bookingRepository) MarkReminderSent(ctx context.Context, bookingID uuid.UUID, reminderType string) error {
	updates := map[string]any{}
	switch reminderType {
	case "24h":
		updates["reminder_sent_24h"] = true
	case "1h":
		updates["reminder_sent_1h"] = true
	default:
		return errors.NewRepositoryError("INVALID_INPUT", "invalid reminder type", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark reminder", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "booking not found", errors.ErrNotFound)
	}

	return nil
}

func (r *bookingRepository) GetBookingsNeedingReminders(ctx context.Context, hoursAhead int) ([]*models.Booking, error) {
	now := time.Now()
	reminderTime := now.Add(time.Duration(hoursAhead) * time.Hour)

	var bookings []*models.Booking
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Where("start_time BETWEEN ? AND ? AND status IN ?",
			now, reminderTime,
			[]models.BookingStatus{models.BookingStatusPending, models.BookingStatusConfirmed})

	if hoursAhead == 24 {
		query = query.Where("reminder_sent_24h = ?", false)
	} else if hoursAhead == 1 {
		query = query.Where("reminder_sent_1h = ?", false)
	}

	if err := query.Order("start_time ASC").Find(&bookings).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find bookings needing reminders", err)
	}

	return bookings, nil
}

//------------------------------------------------------------
// Analytics & Reporting
//------------------------------------------------------------

func (r *bookingRepository) GetBookingStats(ctx context.Context, tenantID uuid.UUID) (BookingStats, error) {
	stats := BookingStats{
		ByStatus: make(map[models.BookingStatus]int64),
	}

	now := time.Now()
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := thisMonthStart

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalBookings).Error; err != nil {
		return stats, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get total bookings", err)
	}

	statuses := []models.BookingStatus{
		models.BookingStatusPending,
		models.BookingStatusConfirmed,
		models.BookingStatusCompleted,
		models.BookingStatusCancelled,
		models.BookingStatusNoShow,
	}

	for _, status := range statuses {
		var count int64
		if err := r.db.WithContext(ctx).Model(&models.Booking{}).
			Where("tenant_id = ? AND status = ?", tenantID, status).
			Count(&count).Error; err == nil {
			stats.ByStatus[status] = count
		}

		switch status {
		case models.BookingStatusPending:
			stats.PendingBookings = count
		case models.BookingStatusConfirmed:
			stats.ConfirmedBookings = count
		case models.BookingStatusCompleted:
			stats.CompletedBookings = count
		case models.BookingStatusCancelled:
			stats.CancelledBookings = count
		case models.BookingStatusNoShow:
			stats.NoShowBookings = count
		}
	}

	var totalRevenue float64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Select("COALESCE(SUM(total_price), 0)").
		Where("tenant_id = ? AND status = ?", tenantID, models.BookingStatusCompleted).
		Scan(&totalRevenue).Error; err == nil {
		stats.TotalRevenue = totalRevenue
	}

	if stats.TotalBookings > 0 {
		stats.AverageBookingValue = totalRevenue / float64(stats.TotalBookings)
	}

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND start_time >= ?", tenantID, thisMonthStart).
		Count(&stats.ThisMonthBookings).Error; err == nil {
	}

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND start_time >= ? AND start_time < ?", tenantID, lastMonthStart, lastMonthEnd).
		Count(&stats.LastMonthBookings).Error; err == nil {
	}

	var thisMonthRevenue float64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Select("COALESCE(SUM(total_price), 0)").
		Where("tenant_id = ? AND status = ? AND start_time >= ?", tenantID, models.BookingStatusCompleted, thisMonthStart).
		Scan(&thisMonthRevenue).Error; err == nil {
		stats.ThisMonthRevenue = thisMonthRevenue
	}

	var lastMonthRevenue float64
	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Select("COALESCE(SUM(total_price), 0)").
		Where("tenant_id = ? AND status = ? AND start_time >= ? AND start_time < ?",
			tenantID, models.BookingStatusCompleted, lastMonthStart, lastMonthEnd).
		Scan(&lastMonthRevenue).Error; err == nil {
		stats.LastMonthRevenue = lastMonthRevenue
	}

	return stats, nil
}

func (r *bookingRepository) GetBookingsByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]BookingPeriodData, error) {
	var results []BookingPeriodData

	// Validate groupBy parameter
	validGroupBy := map[string]bool{
		"day":   true,
		"week":  true,
		"month": true,
		"year":  true,
	}

	if !validGroupBy[groupBy] {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "invalid groupBy value", errors.ErrInvalidInput)
	}

	query := fmt.Sprintf(`
	SELECT
		DATE_TRUNC('%s', start_time) AS period,
		COUNT(*) AS booking_count,
		COALESCE(SUM(CASE WHEN status = 'completed' THEN total_price ELSE 0 END), 0) AS revenue,
		COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_count,
		COALESCE(AVG(CASE WHEN status = 'completed' THEN total_price END), 0) AS average_value
	FROM bookings
	WHERE tenant_id = ? AND start_time >= ? AND start_time <= ?
	GROUP BY period
	ORDER BY period ASC
`, groupBy)

	rows, err := r.db.WithContext(ctx).Raw(query, tenantID, startDate, endDate).Rows()
	if err != nil {
		return nil, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get bookings by period", err)
	}
	defer rows.Close()

	for rows.Next() {
		var data BookingPeriodData
		if err := rows.Scan(&data.Period, &data.BookingCount, &data.Revenue, &data.CompletedCount, &data.AverageValue); err != nil {
			continue
		}
		results = append(results, data)
	}

	return results, nil
}

func (r *bookingRepository) GetArtisanBookingStats(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (ArtisanBookingStats, error) {
	stats := ArtisanBookingStats{
		ArtisanID: artisanID,
		StartDate: startDate,
		EndDate:   endDate,
		ByStatus:  make(map[models.BookingStatus]int64),
	}

	query := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("artisan_id = ? AND start_time >= ? AND start_time <= ?", artisanID, startDate, endDate)

	if err := query.Count(&stats.TotalBookings).Error; err != nil {
		return stats, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get total bookings", err)
	}

	if err := query.Where("status = ?", models.BookingStatusCompleted).Count(&stats.CompletedBookings).Error; err != nil {
		return stats, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get completed bookings", err)
	}

	var revenue float64
	if err := query.Select("COALESCE(SUM(total_price), 0)").
		Where("status = ?", models.BookingStatusCompleted).
		Scan(&revenue).Error; err == nil {
		stats.TotalRevenue = revenue
	}

	if stats.CompletedBookings > 0 {
		stats.AverageBookingValue = revenue / float64(stats.CompletedBookings)
	}

	statuses := []models.BookingStatus{
		models.BookingStatusPending,
		models.BookingStatusConfirmed,
		models.BookingStatusCompleted,
		models.BookingStatusCancelled,
		models.BookingStatusNoShow,
	}

	for _, status := range statuses {
		var count int64
		if err := query.Where("status = ?", status).Count(&count).Error; err == nil {
			stats.ByStatus[status] = count
		}
	}

	if stats.TotalBookings > 0 {
		stats.UtilizationRate = float64(stats.CompletedBookings) / float64(stats.TotalBookings) * 100
	}

	return stats, nil
}

func (r *bookingRepository) GetPopularServices(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]ServiceBookingCount, error) {
	var results []ServiceBookingCount

	query := `
	SELECT
		s.id AS service_id,
		s.name AS service_name,
		COUNT(b.id) AS count,
		COALESCE(SUM(CASE WHEN b.status = 'completed' THEN b.total_price ELSE 0 END), 0) AS revenue,
		COALESCE(AVG(CASE WHEN b.status = 'completed' THEN b.total_price END), 0) AS average_value
	FROM bookings b
	INNER JOIN services s ON b.service_id = s.id
	WHERE b.tenant_id = ? AND b.start_time >= ? AND b.start_time <= ?
	GROUP BY s.id, s.name
	ORDER BY count DESC
	LIMIT ?
`

	rows, err := r.db.WithContext(ctx).Raw(query, tenantID, startDate, endDate, limit).Rows()
	if err != nil {
		return nil, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get popular services", err)
	}
	defer rows.Close()

	for rows.Next() {
		var data ServiceBookingCount
		if err := rows.Scan(&data.ServiceID, &data.ServiceName, &data.Count, &data.Revenue, &data.AverageValue); err != nil {
			continue
		}
		results = append(results, data)
	}

	return results, nil
}

func (r *bookingRepository) GetBookingTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]BookingTrend, error) {
	var results []BookingTrend

	startDate := time.Now().AddDate(0, 0, -days)

	query := `
	SELECT
		DATE(start_time) AS date,
		COUNT(*) AS booking_count,
		COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_count,
		COALESCE(SUM(CASE WHEN status = 'completed' THEN total_price ELSE 0 END), 0) AS revenue
	FROM bookings
	WHERE tenant_id = ? AND start_time >= ?
	GROUP BY DATE(start_time)
	ORDER BY date ASC
`

	rows, err := r.db.WithContext(ctx).Raw(query, tenantID, startDate).Rows()
	if err != nil {
		return nil, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get booking trends", err)
	}
	defer rows.Close()

	for rows.Next() {
		var trend BookingTrend
		if err := rows.Scan(&trend.Date, &trend.BookingCount, &trend.CompletedCount, &trend.Revenue); err != nil {
			continue
		}
		results = append(results, trend)
	}

	return results, nil
}

func (r *bookingRepository) GetAverageBookingValue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var avgValue float64
	err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Select("COALESCE(AVG(total_price), 0)").
		Where("tenant_id = ? AND status = ? AND start_time >= ? AND start_time <= ?",
			tenantID, models.BookingStatusCompleted, startDate, endDate).
		Scan(&avgValue).Error

	if err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get average booking value", err)
	}

	return avgValue, nil
}

func (r *bookingRepository) GetUtilizationRate(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var completed, total int64

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("artisan_id = ? AND start_time >= ? AND start_time <= ? AND status = ?",
			artisanID, startDate, endDate, models.BookingStatusCompleted).
		Count(&completed).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count completed bookings", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("artisan_id = ? AND start_time >= ? AND start_time <= ?",
			artisanID, startDate, endDate).
		Count(&total).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count total bookings", err)
	}

	if total == 0 {
		return 0, nil
	}

	return float64(completed) / float64(total) * 100, nil
}

func (r *bookingRepository) GetCancellationRate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var cancelled, total int64

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND start_time >= ? AND start_time <= ? AND status = ?",
			tenantID, startDate, endDate, models.BookingStatusCancelled).
		Count(&cancelled).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count cancelled bookings", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND start_time >= ? AND start_time <= ?",
			tenantID, startDate, endDate).
		Count(&total).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count total bookings", err)
	}

	if total == 0 {
		return 0, nil
	}

	return float64(cancelled) / float64(total) * 100, nil
}

func (r *bookingRepository) GetNoShowRate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var noShow, total int64

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND start_time >= ? AND start_time <= ? AND status = ?",
			tenantID, startDate, endDate, models.BookingStatusNoShow).
		Count(&noShow).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count no-show bookings", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Where("tenant_id = ? AND start_time >= ? AND start_time <= ?",
			tenantID, startDate, endDate).
		Count(&total).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count total bookings", err)
	}

	if total == 0 {
		return 0, nil
	}

	return float64(noShow) / float64(total) * 100, nil
}

//------------------------------------------------------------
// Search & Filter
//------------------------------------------------------------

func (r *bookingRepository) Search(ctx context.Context, query string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	pagination.Validate()

	like := fmt.Sprintf("%%%s%%", strings.TrimSpace(query))

	countQuery := r.db.WithContext(ctx).Model(&models.Booking{}).Where("tenant_id = ?", tenantID)
	if query != "" {
		countQuery = countQuery.Where("notes ILIKE ? OR customer_notes ILIKE ? OR internal_notes ILIKE ?", like, like, like)
	}

	var totalItems int64
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	dataQuery := r.db.WithContext(ctx).Model(&models.Booking{}).Where("tenant_id = ?", tenantID)
	if query != "" {
		dataQuery = dataQuery.Where("notes ILIKE ? OR customer_notes ILIKE ? OR internal_notes ILIKE ?", like, like, like)
	}

	var bookings []*models.Booking
	if err := dataQuery.
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time DESC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search bookings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

func (r *bookingRepository) FindByFilters(ctx context.Context, filters BookingFilters, pagination PaginationParams) ([]*models.Booking, PaginationResult, error) {
	if filters.TenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id is required", errors.ErrInvalidInput)
	}

	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.Booking{})
	query = r.applyBookingFilters(query, filters)

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count bookings", err)
	}

	var bookings []*models.Booking
	if err := r.applyBookingFilters(r.db.WithContext(ctx).Model(&models.Booking{}), filters).
		Preload("Customer").
		Preload("Artisan").
		Preload("Service").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("start_time DESC").
		Find(&bookings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to apply filters", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return bookings, paginationResult, nil
}

//------------------------------------------------------------
// Bulk Operations
//------------------------------------------------------------

func (r *bookingRepository) BulkConfirm(ctx context.Context, bookingIDs []uuid.UUID) error {
	if len(bookingIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id IN ?", bookingIDs).
		Update("status", models.BookingStatusConfirmed)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk confirm bookings", result.Error)
	}

	for _, id := range bookingIDs {
		r.InvalidateCache(ctx, id)
	}

	return nil
}

func (r *bookingRepository) BulkCancel(ctx context.Context, bookingIDs []uuid.UUID, reason string) error {
	if len(bookingIDs) == 0 {
		return nil
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id IN ?", bookingIDs).
		Updates(map[string]any{
			"status":              models.BookingStatusCancelled,
			"cancelled_at":        &now,
			"cancellation_reason": reason,
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk cancel bookings", result.Error)
	}

	for _, id := range bookingIDs {
		r.InvalidateCache(ctx, id)
	}

	return nil
}

func (r *bookingRepository) BulkUpdatePaymentStatus(ctx context.Context, bookingIDs []uuid.UUID, status models.PaymentStatus) error {
	if len(bookingIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Where("id IN ?", bookingIDs).
		Update("payment_status", status)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update payment status", result.Error)
	}

	for _, id := range bookingIDs {
		r.InvalidateCache(ctx, id)
	}

	return nil
}

//------------------------------------------------------------
// Helper Methods
//------------------------------------------------------------

func (r *bookingRepository) applyDateRange(query *gorm.DB, column string, startDate, endDate time.Time) *gorm.DB {
	if !startDate.IsZero() {
		query = query.Where(fmt.Sprintf("%s >= ?", column), startDate)
	}
	if !endDate.IsZero() {
		query = query.Where(fmt.Sprintf("%s <= ?", column), endDate)
	}
	return query
}

func (r *bookingRepository) applyBookingFilters(query *gorm.DB, filters BookingFilters) *gorm.DB {
	query = query.Where("tenant_id = ?", filters.TenantID)

	if len(filters.ArtisanIDs) > 0 {
		query = query.Where("artisan_id IN ?", filters.ArtisanIDs)
	}

	if len(filters.CustomerIDs) > 0 {
		query = query.Where("customer_id IN ?", filters.CustomerIDs)
	}

	if len(filters.ServiceIDs) > 0 {
		query = query.Where("service_id IN ?", filters.ServiceIDs)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if len(filters.PaymentStatuses) > 0 {
		query = query.Where("payment_status IN ?", filters.PaymentStatuses)
	}

	if filters.StartDateFrom != nil {
		query = query.Where("start_time >= ?", *filters.StartDateFrom)
	}

	if filters.StartDateTo != nil {
		query = query.Where("start_time <= ?", *filters.StartDateTo)
	}

	if filters.MinPrice != nil {
		query = query.Where("total_price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("total_price <= ?", *filters.MaxPrice)
	}

	if filters.IsRecurring != nil {
		query = query.Where("is_recurring = ?", *filters.IsRecurring)
	}

	return query
}

