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
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Booking Operations
	// ============================================================================

	// Create booking (authenticated, requires booking:write scope)
	bookings.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.CreateBooking,
	)

	// Get booking by ID (authenticated, requires booking:read scope)
	bookings.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetBooking,
	)

	// Update booking (authenticated, requires booking:write scope)
	bookings.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.UpdateBooking,
	)

	// Delete booking (authenticated, requires booking:write scope)
	bookings.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.DeleteBooking,
	)

	// List bookings (authenticated, requires booking:read scope)
	bookings.Get("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.ListBookings,
	)

	// Search bookings (authenticated, requires booking:read scope)
	bookings.Post("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.SearchBookings,
	)

	// ============================================================================
	// Artisan & Customer Specific Queries
	// ============================================================================

	// Get bookings by artisan (authenticated, requires booking:read scope)
	bookings.Get("/artisan/:artisan_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetBookingsByArtisan,
	)

	// Get artisan schedule (authenticated, requires booking:read scope)
	bookings.Get("/artisan/:artisan_id/schedule",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetArtisanSchedule,
	)

	// Get artisan booking stats (authenticated, requires booking:read scope)
	bookings.Get("/artisan/:artisan_id/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetArtisanBookingStats,
	)

	// Get bookings by customer (authenticated, requires booking:read scope)
	bookings.Get("/customer/:customer_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetBookingsByCustomer,
	)

	// ============================================================================
	// Status Management Operations
	// ============================================================================

	// Confirm booking (authenticated, requires booking:write scope)
	bookings.Post("/:id/confirm",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.ConfirmBooking,
	)

	// Start booking (authenticated, requires booking:write scope)
	bookings.Post("/:id/start",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.StartBooking,
	)

	// Complete booking (authenticated, requires booking:write scope)
	bookings.Post("/:id/complete",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.CompleteBooking,
	)

	// Cancel booking (authenticated, requires booking:write scope)
	bookings.Post("/:id/cancel",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.CancelBooking,
	)

	// Mark as no-show (authenticated, requires booking:write scope)
	bookings.Post("/:id/no-show",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.MarkAsNoShow,
	)

	// Reschedule booking (authenticated, requires booking:write scope)
	bookings.Post("/:id/reschedule",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.RescheduleBooking,
	)

	// ============================================================================
	// Scheduling & Availability
	// ============================================================================

	// Check artisan availability (authenticated, requires booking:read scope)
	bookings.Post("/check-availability",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.CheckArtisanAvailability,
	)

	// Get available time slots (authenticated, requires booking:read scope)
	bookings.Get("/available-slots",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetAvailableTimeSlots,
	)

	// ============================================================================
	// Time-based Queries
	// ============================================================================

	// Get upcoming bookings (authenticated, requires booking:read scope)
	bookings.Get("/upcoming",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetUpcomingBookings,
	)

	// Get today's bookings (authenticated, requires booking:read scope)
	bookings.Get("/today",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetTodayBookings,
	)

	// Get bookings in date range (authenticated, requires booking:read scope)
	bookings.Get("/date-range",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetBookingsInDateRange,
	)

	// ============================================================================
	// Photo Management
	// ============================================================================

	// Add before photos (authenticated, requires booking:write scope)
	bookings.Post("/:id/photos/before",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.AddBeforePhotos,
	)

	// Add after photos (authenticated, requires booking:write scope)
	bookings.Post("/:id/photos/after",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.AddAfterPhotos,
	)

	// ============================================================================
	// Analytics & Reporting
	// ============================================================================

	// Get booking statistics (authenticated, requires booking:read scope)
	bookings.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingRead),
		bookingHandler.GetBookingStats,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk confirm bookings (authenticated, requires booking:write scope)
	bookings.Post("/bulk/confirm",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.BulkConfirm,
	)

	// Bulk cancel bookings (authenticated, requires booking:write scope)
	bookings.Post("/bulk/cancel",
		authMiddleware,
		middleware.RequireScopes(r.scopes.BookingWrite),
		bookingHandler.BulkCancel,
	)
}
