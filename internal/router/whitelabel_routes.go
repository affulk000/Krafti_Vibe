package router

import (
	"Krafti_Vibe/internal/handler"
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
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.GetMyWhiteLabel,
	)

	// Create whitelabel configuration
	whitelabel.Post("",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.CreateWhiteLabel,
	)

	// Get whitelabel by ID
	whitelabel.Get("/:id",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.GetWhiteLabel,
	)

	// Update whitelabel configuration
	whitelabel.Put("/:id",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.UpdateWhiteLabel,
	)

	// Delete whitelabel configuration
	whitelabel.Delete("/:id",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.DeleteWhiteLabel,
	)

	// ============================================================================
	// Partial Update Routes
	// ============================================================================

	// Update color scheme only
	whitelabel.Put("/colors",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.UpdateColorScheme,
	)

	// Update branding assets only
	whitelabel.Put("/branding",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.UpdateBranding,
	)

	// Update custom domain only
	whitelabel.Put("/domain",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.UpdateDomain,
	)

	// ============================================================================
	// Activation Routes
	// ============================================================================

	// Activate whitelabel
	whitelabel.Post("/activate",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.ActivateWhiteLabel,
	)

	// Deactivate whitelabel
	whitelabel.Post("/deactivate",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.DeactivateWhiteLabel,
	)

	// ============================================================================
	// Utility Routes
	// ============================================================================

	// Check domain availability
	whitelabel.Get("/check-domain",
		r.zitadelMW.RequireAuth(),
		whiteLabelHandler.CheckDomainAvailability,
	)
}
