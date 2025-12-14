package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupWhiteLabelRoutes sets up whitelabel routes
func (r *Router) setupWhiteLabelRoutes(api fiber.Router) {
	// Initialize whitelabel service
	whiteLabelService := service.NewWhiteLabelService(r.repos, r.config.Logger)

	// Initialize whitelabel handler
	whiteLabelHandler := handler.NewWhiteLabelHandler(whiteLabelService)

	// WhiteLabel routes group
	whitelabel := api.Group("/whitelabel")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Public Routes (no auth required)
	// ============================================================================

	// Get public whitelabel by tenant ID
	whitelabel.Get("/public",
		whiteLabelHandler.GetPublicWhiteLabel,
	)

	// Get public whitelabel by custom domain
	whitelabel.Get("/domain",
		whiteLabelHandler.GetPublicWhiteLabelByDomain,
	)

	// ============================================================================
	// Protected Routes (authentication required)
	// ============================================================================

	// Get my whitelabel configuration
	whitelabel.Get("/me",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		whiteLabelHandler.GetMyWhiteLabel,
	)

	// Create whitelabel configuration
	whitelabel.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		whiteLabelHandler.CreateWhiteLabel,
	)

	// Get whitelabel by ID
	whitelabel.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		whiteLabelHandler.GetWhiteLabel,
	)

	// Update whitelabel configuration
	whitelabel.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		whiteLabelHandler.UpdateWhiteLabel,
	)

	// Delete whitelabel configuration
	whitelabel.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		whiteLabelHandler.DeleteWhiteLabel,
	)

	// ============================================================================
	// Partial Update Routes
	// ============================================================================

	// Update color scheme only
	whitelabel.Put("/colors",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		whiteLabelHandler.UpdateColorScheme,
	)

	// Update branding assets only
	whitelabel.Put("/branding",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		whiteLabelHandler.UpdateBranding,
	)

	// Update custom domain only
	whitelabel.Put("/domain",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		whiteLabelHandler.UpdateDomain,
	)

	// ============================================================================
	// Activation Routes
	// ============================================================================

	// Activate whitelabel
	whitelabel.Post("/activate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		whiteLabelHandler.ActivateWhiteLabel,
	)

	// Deactivate whitelabel
	whitelabel.Post("/deactivate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		whiteLabelHandler.DeactivateWhiteLabel,
	)

	// ============================================================================
	// Utility Routes
	// ============================================================================

	// Check domain availability
	whitelabel.Get("/check-domain",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		whiteLabelHandler.CheckDomainAvailability,
	)
}
