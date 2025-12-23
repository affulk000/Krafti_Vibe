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
		r.zitadelMW.RequireAuth(),
		reportHandler.CreateReport,
	)

	// List reports
	reports.Get("",
		r.zitadelMW.RequireAuth(),
		reportHandler.ListReports,
	)

	// Search reports
	reports.Get("/search",
		r.zitadelMW.RequireAuth(),
		reportHandler.SearchReports,
	)

	// Get report by ID
	reports.Get("/:id",
		r.zitadelMW.RequireAuth(),
		reportHandler.GetReport,
	)

	// Update report
	reports.Put("/:id",
		r.zitadelMW.RequireAuth(),
		reportHandler.UpdateReport,
	)

	// Delete report
	reports.Delete("/:id",
		r.zitadelMW.RequireAuth(),
		reportHandler.DeleteReport,
	)

	// ============================================================================
	// Status-Based Queries
	// ============================================================================

	// Get pending reports
	reports.Get("/pending",
		r.zitadelMW.RequireAuth(),
		reportHandler.GetPendingReports,
	)

	// Get scheduled reports
	reports.Get("/scheduled",
		r.zitadelMW.RequireAuth(),
		reportHandler.GetScheduledReports,
	)

	// Get failed reports
	reports.Get("/failed",
		r.zitadelMW.RequireAuth(),
		reportHandler.GetFailedReports,
	)

	// ============================================================================
	// Status Management (System/Internal Operations)
	// ============================================================================

	// Mark as generating
	reports.Post("/:id/generating",
		r.zitadelMW.RequireAuth(),
		reportHandler.MarkAsGenerating,
	)

	// Mark as completed
	reports.Post("/:id/completed",
		r.zitadelMW.RequireAuth(),
		reportHandler.MarkAsCompleted,
	)

	// Mark as failed
	reports.Post("/:id/failed",
		r.zitadelMW.RequireAuth(),
		reportHandler.MarkAsFailed,
	)

	// Retry failed report
	reports.Post("/:id/retry",
		r.zitadelMW.RequireAuth(),
		reportHandler.RetryFailedReport,
	)

	// ============================================================================
	// Schedule Management
	// ============================================================================

	// Enable schedule
	reports.Post("/:id/schedule/enable",
		r.zitadelMW.RequireAuth(),
		reportHandler.EnableSchedule,
	)

	// Disable schedule
	reports.Post("/:id/schedule/disable",
		r.zitadelMW.RequireAuth(),
		reportHandler.DisableSchedule,
	)

	// Update schedule cron
	reports.Put("/:id/schedule/cron",
		r.zitadelMW.RequireAuth(),
		reportHandler.UpdateScheduleCron,
	)

	// ============================================================================
	// Statistics & Analytics
	// ============================================================================

	// Get report statistics
	reports.Get("/stats",
		r.zitadelMW.RequireAuth(),
		reportHandler.GetReportStats,
	)

	// Get report type usage
	reports.Get("/analytics/type-usage",
		r.zitadelMW.RequireAuth(),
		reportHandler.GetReportTypeUsage,
	)

	// Get user report activity
	reports.Get("/analytics/user-activity",
		r.zitadelMW.RequireAuth(),
		reportHandler.GetUserReportActivity,
	)

	// Get generation metrics
	reports.Get("/analytics/generation-metrics",
		r.zitadelMW.RequireAuth(),
		reportHandler.GetReportGenerationMetrics,
	)

	// ============================================================================
	// Cleanup Operations (Admin/Manage access required)
	// ============================================================================

	// Delete old reports
	reports.Delete("/cleanup/old",
		r.zitadelMW.RequireAuth(),
		reportHandler.DeleteOldReports,
	)

	// Delete failed reports
	reports.Delete("/cleanup/failed",
		r.zitadelMW.RequireAuth(),
		reportHandler.DeleteFailedReports,
	)
}
