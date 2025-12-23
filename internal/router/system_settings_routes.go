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
		r.zitadelMW.RequireAuth(),
		settingsHandler.CreateSetting,
	)

	// List settings
	settings.Get("",
		r.zitadelMW.RequireAuth(),
		settingsHandler.ListSettings,
	)

	// Search settings
	settings.Get("/search",
		r.zitadelMW.RequireAuth(),
		settingsHandler.SearchSettings,
	)

	// Get setting by ID
	settings.Get("/:id",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetSetting,
	)

	// Update setting
	settings.Put("/:id",
		r.zitadelMW.RequireAuth(),
		settingsHandler.UpdateSetting,
	)

	// Delete setting
	settings.Delete("/:id",
		r.zitadelMW.RequireAuth(),
		settingsHandler.DeleteSetting,
	)

	// ============================================================================
	// Key-Based Operations
	// ============================================================================

	// Get setting by key
	settings.Get("/key/:key",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetSettingByKey,
	)

	// Delete setting by key
	settings.Delete("/key/:key",
		r.zitadelMW.RequireAuth(),
		settingsHandler.DeleteSettingByKey,
	)

	// ============================================================================
	// Category & Group Management
	// ============================================================================

	// Get all categories
	settings.Get("/categories",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetSettingCategories,
	)

	// Get categories with counts
	settings.Get("/categories/stats",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetCategoriesWithCount,
	)

	// Get all groups
	settings.Get("/groups",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetSettingGroups,
	)

	// Get settings by category
	settings.Get("/category/:category",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetSettingsByCategory,
	)

	// Get settings by group
	settings.Get("/group/:group",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetSettingsByGroup,
	)

	// ============================================================================
	// Public/Private Settings
	// ============================================================================

	// Get public settings
	settings.Get("/public",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetPublicSettings,
	)

	// Get private settings
	settings.Get("/private",
		r.zitadelMW.RequireAuth(),
		settingsHandler.GetPrivateSettings,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk set settings
	settings.Post("/bulk",
		r.zitadelMW.RequireAuth(),
		settingsHandler.BulkSetSettings,
	)

	// Bulk delete settings
	settings.Post("/bulk/delete",
		r.zitadelMW.RequireAuth(),
		settingsHandler.BulkDeleteSettings,
	)
}
