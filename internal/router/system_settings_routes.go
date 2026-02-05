package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupSystemSettingsRoutes sets up system settings routes
func (r *Router) setupSystemSettingsRoutes(api fiber.Router) {
	// Initialize system settings service
	settingsService := service.NewSystemSettingService(r.repos, r.config.Logger)

	// Initialize system settings handler
	settingsHandler := handler.NewSystemSettingsHandler(settingsService)

	// System settings routes group
	settings := api.Group("/settings")

	// Auth middleware configuration

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create setting
	settings.Post("",
		r.RequireAuth(),
		settingsHandler.CreateSetting,
	)

	// List settings
	settings.Get("",
		r.RequireAuth(),
		settingsHandler.ListSettings,
	)

	// Search settings
	settings.Get("/search",
		r.RequireAuth(),
		settingsHandler.SearchSettings,
	)

	// Get setting by ID
	settings.Get("/:id",
		r.RequireAuth(),
		settingsHandler.GetSetting,
	)

	// Update setting
	settings.Put("/:id",
		r.RequireAuth(),
		settingsHandler.UpdateSetting,
	)

	// Delete setting
	settings.Delete("/:id",
		r.RequireAuth(),
		settingsHandler.DeleteSetting,
	)

	// ============================================================================
	// Key-Based Operations
	// ============================================================================

	// Get setting by key
	settings.Get("/key/:key",
		r.RequireAuth(),
		settingsHandler.GetSettingByKey,
	)

	// Delete setting by key
	settings.Delete("/key/:key",
		r.RequireAuth(),
		settingsHandler.DeleteSettingByKey,
	)

	// ============================================================================
	// Category & Group Management
	// ============================================================================

	// Get all categories
	settings.Get("/categories",
		r.RequireAuth(),
		settingsHandler.GetSettingCategories,
	)

	// Get categories with counts
	settings.Get("/categories/stats",
		r.RequireAuth(),
		settingsHandler.GetCategoriesWithCount,
	)

	// Get all groups
	settings.Get("/groups",
		r.RequireAuth(),
		settingsHandler.GetSettingGroups,
	)

	// Get settings by category
	settings.Get("/category/:category",
		r.RequireAuth(),
		settingsHandler.GetSettingsByCategory,
	)

	// Get settings by group
	settings.Get("/group/:group",
		r.RequireAuth(),
		settingsHandler.GetSettingsByGroup,
	)

	// ============================================================================
	// Public/Private Settings
	// ============================================================================

	// Get public settings
	settings.Get("/public",
		r.RequireAuth(),
		settingsHandler.GetPublicSettings,
	)

	// Get private settings
	settings.Get("/private",
		r.RequireAuth(),
		settingsHandler.GetPrivateSettings,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk set settings
	settings.Post("/bulk",
		r.RequireAuth(),
		settingsHandler.BulkSetSettings,
	)

	// Bulk delete settings
	settings.Post("/bulk/delete",
		r.RequireAuth(),
		settingsHandler.BulkDeleteSettings,
	)
}
