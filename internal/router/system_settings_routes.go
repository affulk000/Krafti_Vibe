package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
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
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create setting
	settings.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsWrite),
		settingsHandler.CreateSetting,
	)

	// List settings
	settings.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.ListSettings,
	)

	// Search settings
	settings.Get("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.SearchSettings,
	)

	// Get setting by ID
	settings.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.GetSetting,
	)

	// Update setting
	settings.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsWrite),
		settingsHandler.UpdateSetting,
	)

	// Delete setting
	settings.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsDelete),
		settingsHandler.DeleteSetting,
	)

	// ============================================================================
	// Key-Based Operations
	// ============================================================================

	// Get setting by key
	settings.Get("/key/:key",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.GetSettingByKey,
	)

	// Delete setting by key
	settings.Delete("/key/:key",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsDelete),
		settingsHandler.DeleteSettingByKey,
	)

	// ============================================================================
	// Category & Group Management
	// ============================================================================

	// Get all categories
	settings.Get("/categories",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.GetSettingCategories,
	)

	// Get categories with counts
	settings.Get("/categories/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.GetCategoriesWithCount,
	)

	// Get all groups
	settings.Get("/groups",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.GetSettingGroups,
	)

	// Get settings by category
	settings.Get("/category/:category",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.GetSettingsByCategory,
	)

	// Get settings by group
	settings.Get("/group/:group",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.GetSettingsByGroup,
	)

	// ============================================================================
	// Public/Private Settings
	// ============================================================================

	// Get public settings
	settings.Get("/public",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsRead),
		settingsHandler.GetPublicSettings,
	)

	// Get private settings
	settings.Get("/private",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsManage),
		settingsHandler.GetPrivateSettings,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk set settings
	settings.Post("/bulk",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsWrite),
		settingsHandler.BulkSetSettings,
	)

	// Bulk delete settings
	settings.Post("/bulk/delete",
		authMiddleware,
		middleware.RequireScopes(r.scopes.SettingsDelete),
		settingsHandler.BulkDeleteSettings,
	)
}
