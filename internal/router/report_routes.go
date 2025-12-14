package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
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
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create report
	reports.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.CreateReport,
	)

	// List reports
	reports.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.ListReports,
	)

	// Search reports
	reports.Get("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.SearchReports,
	)

	// Get report by ID
	reports.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.GetReport,
	)

	// Update report
	reports.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.UpdateReport,
	)

	// Delete report
	reports.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportExport),
		reportHandler.DeleteReport,
	)

	// ============================================================================
	// Status-Based Queries
	// ============================================================================

	// Get pending reports
	reports.Get("/pending",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.GetPendingReports,
	)

	// Get scheduled reports
	reports.Get("/scheduled",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.GetScheduledReports,
	)

	// Get failed reports
	reports.Get("/failed",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.GetFailedReports,
	)

	// ============================================================================
	// Status Management (System/Internal Operations)
	// ============================================================================

	// Mark as generating
	reports.Post("/:id/generating",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.MarkAsGenerating,
	)

	// Mark as completed
	reports.Post("/:id/completed",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.MarkAsCompleted,
	)

	// Mark as failed
	reports.Post("/:id/failed",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.MarkAsFailed,
	)

	// Retry failed report
	reports.Post("/:id/retry",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.RetryFailedReport,
	)

	// ============================================================================
	// Schedule Management
	// ============================================================================

	// Enable schedule
	reports.Post("/:id/schedule/enable",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.EnableSchedule,
	)

	// Disable schedule
	reports.Post("/:id/schedule/disable",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.DisableSchedule,
	)

	// Update schedule cron
	reports.Put("/:id/schedule/cron",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportGenerate),
		reportHandler.UpdateScheduleCron,
	)

	// ============================================================================
	// Statistics & Analytics
	// ============================================================================

	// Get report statistics
	reports.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.GetReportStats,
	)

	// Get report type usage
	reports.Get("/analytics/type-usage",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.GetReportTypeUsage,
	)

	// Get user report activity
	reports.Get("/analytics/user-activity",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.GetUserReportActivity,
	)

	// Get generation metrics
	reports.Get("/analytics/generation-metrics",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportRead),
		reportHandler.GetReportGenerationMetrics,
	)

	// ============================================================================
	// Cleanup Operations (Admin/Manage access required)
	// ============================================================================

	// Delete old reports
	reports.Delete("/cleanup/old",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportExport),
		reportHandler.DeleteOldReports,
	)

	// Delete failed reports
	reports.Delete("/cleanup/failed",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReportExport),
		reportHandler.DeleteFailedReports,
	)
}
