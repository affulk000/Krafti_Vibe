package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupSDKRoutes configures all SDK-related routes
func (r *Router) setupSDKRoutes(api fiber.Router) {
	// Initialize SDK service
	sdkService := service.NewSDKService(r.repos, r.config.Logger)

	// Initialize SDK handler
	sdkHandler := handler.NewSDKHandler(sdkService)

	// SDK routes group
	sdk := api.Group("/sdk")

	// Public Routes (no auth required)
	// Key Validation (public endpoint - for SDK runtime validation)
	sdk.Post("/keys/validate", sdkHandler.ValidateSDKKey)

	// SDK Client Routes
	clients := sdk.Group("/clients")
	clients.Use(r.RequireAuth())

	// Client CRUD
	clients.Post("", r.zitadelMW.RequireRole("platform_super_admin"), sdkHandler.CreateSDKClient)
	clients.Get("", sdkHandler.ListSDKClients)
	clients.Get("/:id", sdkHandler.GetSDKClient)
	clients.Put("/:id", r.zitadelMW.RequireRole("platform_super_admin"), sdkHandler.UpdateSDKClient)
	clients.Delete("/:id", r.zitadelMW.RequireRole("platform_super_admin"), sdkHandler.DeleteSDKClient)

	// SDK Key Routes
	keys := sdk.Group("/keys")
	keys.Use(r.RequireAuth())

	// Key Management
	keys.Post("", sdkHandler.CreateSDKKey)
	keys.Get("", sdkHandler.ListSDKKeys)
	keys.Get("/:id", sdkHandler.GetSDKKey)
	keys.Put("/:id", sdkHandler.UpdateSDKKey)

	// Key Status
	keys.Post("/:id/revoke", sdkHandler.RevokeSDKKey)
	keys.Post("/:id/rotate", sdkHandler.RotateSDKKey)

	// Usage Tracking
	sdk.Post("/usage", r.RequireAuth(), sdkHandler.TrackSDKUsage)
	sdk.Get("/usage", r.RequireAuth(), sdkHandler.ListSDKUsage)
}
