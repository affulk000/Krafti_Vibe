package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupBookingRoutes(api fiber.Router) {
	// Initialize dependent services
	customerService := service.NewCustomerService(r.repos, r.config.Logger)
	paymentService := service.NewPaymentService(r.repos, r.config.Logger)

	// Initialize booking service with dependencies
	bookingService := service.NewBookingService(r.repos, r.config.Logger, customerService, paymentService)
	bookingHandler := handler.NewBookingHandler(bookingService)

	// Create bookings group
	bookings := api.Group("/bookings")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		bookings.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration
	bookings.Use(r.RequireAuth())

	// ============================================================================
	// Core Booking Operations
	// ============================================================================

	// Create booking - any authenticated user can create a booking
	bookings.Post("/",
		bookingHandler.CreateBooking,
	)

	// Get booking by ID - owner (customer/artisan) or tenant owner/admin
	bookings.Get("/:id",
		bookingHandler.GetBooking,
	)

	// Update booking - owner (customer/artisan) or tenant owner/admin
	bookings.Put("/:id",
		bookingHandler.UpdateBooking,
	)

	// Delete booking - tenant owner/admin only
	bookings.Delete("/:id",
		middleware.RequireTenantOwnerOrAdmin(),
		bookingHandler.DeleteBooking,
	)

	// List bookings - tenant owner/admin only (filters by tenant)
	bookings.Get("/",
		middleware.RequireTenantOwnerOrAdmin(),
		bookingHandler.ListBookings,
	)

	// Search bookings - tenant owner/admin only
	bookings.Post("/search",
		middleware.RequireTenantOwnerOrAdmin(),
		bookingHandler.SearchBookings,
	)

	// ============================================================================
	// Artisan & Customer Specific Queries
	// ============================================================================

	// Get bookings by artisan - artisan (self) or tenant owner/admin
	bookings.Get("/artisan/:artisan_id",
		middleware.RequireArtisanOrTeamMember(),
		bookingHandler.GetBookingsByArtisan,
	)

	// Get artisan schedule - artisan (self) or tenant owner/admin
	bookings.Get("/artisan/:artisan_id/schedule",
		middleware.RequireArtisanOrTeamMember(),
		bookingHandler.GetArtisanSchedule,
	)

	// Get artisan booking stats - artisan (self) or tenant owner/admin
	bookings.Get("/artisan/:artisan_id/stats",
		middleware.RequireArtisanOrTeamMember(),
		bookingHandler.GetArtisanBookingStats,
	)

	// Get bookings by customer - customer (self) or tenant owner/admin
	bookings.Get("/customer/:customer_id",
		bookingHandler.GetBookingsByCustomer,
	)

	// ============================================================================
	// Status Management Operations
	// ============================================================================

	// Confirm booking - artisan or tenant owner/admin
	bookings.Post("/:id/confirm",
		middleware.RequireArtisanOrTeamMember(),
		bookingHandler.ConfirmBooking,
	)

	// Start booking - artisan or tenant owner/admin
	bookings.Post("/:id/start",
		middleware.RequireArtisanOrTeamMember(),
		bookingHandler.StartBooking,
	)

	// Complete booking - artisan or tenant owner/admin
	bookings.Post("/:id/complete",
		middleware.RequireArtisanOrTeamMember(),
		bookingHandler.CompleteBooking,
	)

	// Cancel booking - customer, artisan, or tenant owner/admin
	bookings.Post("/:id/cancel",
		bookingHandler.CancelBooking,
	)

	// Mark as no-show - artisan or tenant owner/admin
	bookings.Post("/:id/no-show",
		middleware.RequireArtisanOrTeamMember(),
		bookingHandler.MarkAsNoShow,
	)

	// Reschedule booking - customer or artisan
	bookings.Post("/:id/reschedule",
		bookingHandler.RescheduleBooking,
	)

	// ============================================================================
	// Scheduling & Availability
	// ============================================================================

	// Check artisan availability - any authenticated user
	bookings.Post("/check-availability",
		bookingHandler.CheckArtisanAvailability,
	)

	// Get available time slots - any authenticated user
	bookings.Get("/available-slots",
		bookingHandler.GetAvailableTimeSlots,
	)

	// ============================================================================
	// Time-based Queries
	// ============================================================================

	// Get upcoming bookings - tenant owner/admin
	bookings.Get("/upcoming",
		middleware.RequireTenantOwnerOrAdmin(),
		bookingHandler.GetUpcomingBookings,
	)

	// Get today's bookings - tenant owner/admin or artisan
	bookings.Get("/today",
		middleware.RequireArtisanOrTeamMember(),
		bookingHandler.GetTodayBookings,
	)

	// Get bookings in date range - tenant owner/admin
	bookings.Get("/date-range",
		middleware.RequireTenantOwnerOrAdmin(),
		bookingHandler.GetBookingsInDateRange,
	)

	// ============================================================================
	// Photo Management
	// ============================================================================

	// Add before photos - artisan only
	bookings.Post("/:id/photos/before",
		middleware.RequireArtisan(),
		bookingHandler.AddBeforePhotos,
	)

	// Add after photos - artisan only
	bookings.Post("/:id/photos/after",
		middleware.RequireArtisan(),
		bookingHandler.AddAfterPhotos,
	)

	// ============================================================================
	// Analytics & Reporting
	// ============================================================================

	// Get booking statistics - tenant owner/admin only
	bookings.Get("/stats",
		middleware.RequireTenantOwnerOrAdmin(),
		bookingHandler.GetBookingStats,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk confirm bookings - tenant owner/admin only
	bookings.Post("/bulk/confirm",
		middleware.RequireTenantOwnerOrAdmin(),
		bookingHandler.BulkConfirm,
	)

	// Bulk cancel bookings - tenant owner/admin only
	bookings.Post("/bulk/cancel",
		middleware.RequireTenantOwnerOrAdmin(),
		bookingHandler.BulkCancel,
	)
}
