package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
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

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ========================================================================
	// Public Routes (no auth required)
	// ========================================================================

	// Key Validation (public endpoint - for SDK runtime validation)
	sdk.Post("/keys/validate", sdkHandler.ValidateSDKKey)

	// ========================================================================
	// SDK Client Routes
	// ========================================================================
	clients := sdk.Group("/clients")

	// Client CRUD
	clients.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		sdkHandler.CreateSDKClient,
	)

	clients.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		sdkHandler.ListSDKClients,
	)

	clients.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		sdkHandler.GetSDKClient,
	)

	clients.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		sdkHandler.UpdateSDKClient,
	)

	clients.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		sdkHandler.DeleteSDKClient,
	)

	// ========================================================================
	// SDK Key Routes
	// ========================================================================
	keys := sdk.Group("/keys")

	// Key CRUD
	keys.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		sdkHandler.CreateSDKKey,
	)

	keys.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		sdkHandler.ListSDKKeys,
	)

	keys.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		sdkHandler.GetSDKKey,
	)

	keys.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		sdkHandler.UpdateSDKKey,
	)

	// Key Management
	keys.Post("/:id/revoke",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		sdkHandler.RevokeSDKKey,
	)

	keys.Post("/:id/rotate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		sdkHandler.RotateSDKKey,
	)

	// ========================================================================
	// SDK Usage Routes
	// ========================================================================
	usage := sdk.Group("/usage")

	// Usage tracking and analytics
	usage.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		sdkHandler.TrackSDKUsage,
	)

	usage.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		sdkHandler.ListSDKUsage,
	)

	usage.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		sdkHandler.GetSDKUsageStats,
	)
}
