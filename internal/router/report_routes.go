package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupReportRoutes sets up report routes
func (r *Router) setupReportRoutes(api fiber.Router) {
	// Initialize report service
	reportService := service.NewReportService(r.repos, r.config.Logger)

	// Initialize report handler
	reportHandler := handler.NewReportHandler(reportService)

	// Report routes group
	reports := api.Group("/reports")

	// Auth middleware configuration

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create report
	reports.Post("",
		r.RequireAuth(),
		reportHandler.CreateReport,
	)

	// List reports
	reports.Get("",
		r.RequireAuth(),
		reportHandler.ListReports,
	)

	// Search reports
	reports.Get("/search",
		r.RequireAuth(),
		reportHandler.SearchReports,
	)

	// Get report by ID
	reports.Get("/:id",
		r.RequireAuth(),
		reportHandler.GetReport,
	)

	// Update report
	reports.Put("/:id",
		r.RequireAuth(),
		reportHandler.UpdateReport,
	)

	// Delete report
	reports.Delete("/:id",
		r.RequireAuth(),
		reportHandler.DeleteReport,
	)

	// ============================================================================
	// Status-Based Queries
	// ============================================================================

	// Get pending reports
	reports.Get("/pending",
		r.RequireAuth(),
		reportHandler.GetPendingReports,
	)

	// Get scheduled reports
	reports.Get("/scheduled",
		r.RequireAuth(),
		reportHandler.GetScheduledReports,
	)

	// Get failed reports
	reports.Get("/failed",
		r.RequireAuth(),
		reportHandler.GetFailedReports,
	)

	// ============================================================================
	// Status Management (System/Internal Operations)
	// ============================================================================

	// Mark as generating
	reports.Post("/:id/generating",
		r.RequireAuth(),
		reportHandler.MarkAsGenerating,
	)

	// Mark as completed
	reports.Post("/:id/completed",
		r.RequireAuth(),
		reportHandler.MarkAsCompleted,
	)

	// Mark as failed
	reports.Post("/:id/failed",
		r.RequireAuth(),
		reportHandler.MarkAsFailed,
	)

	// Retry failed report
	reports.Post("/:id/retry",
		r.RequireAuth(),
		reportHandler.RetryFailedReport,
	)

	// ============================================================================
	// Schedule Management
	// ============================================================================

	// Enable schedule
	reports.Post("/:id/schedule/enable",
		r.RequireAuth(),
		reportHandler.EnableSchedule,
	)

	// Disable schedule
	reports.Post("/:id/schedule/disable",
		r.RequireAuth(),
		reportHandler.DisableSchedule,
	)

	// Update schedule cron
	reports.Put("/:id/schedule/cron",
		r.RequireAuth(),
		reportHandler.UpdateScheduleCron,
	)

	// ============================================================================
	// Statistics & Analytics
	// ============================================================================

	// Get report statistics
	reports.Get("/stats",
		r.RequireAuth(),
		reportHandler.GetReportStats,
	)

	// Get report type usage
	reports.Get("/analytics/type-usage",
		r.RequireAuth(),
		reportHandler.GetReportTypeUsage,
	)

	// Get user report activity
	reports.Get("/analytics/user-activity",
		r.RequireAuth(),
		reportHandler.GetUserReportActivity,
	)

	// Get generation metrics
	reports.Get("/analytics/generation-metrics",
		r.RequireAuth(),
		reportHandler.GetReportGenerationMetrics,
	)

	// ============================================================================
	// Cleanup Operations (Admin/Manage access required)
	// ============================================================================

	// Delete old reports
	reports.Delete("/cleanup/old",
		r.RequireAuth(),
		reportHandler.DeleteOldReports,
	)

	// Delete failed reports
	reports.Delete("/cleanup/failed",
		r.RequireAuth(),
		reportHandler.DeleteFailedReports,
	)
}
