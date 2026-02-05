package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// setupUserRoutes configures all user-related routes with authentication
// This function is called by the main router setup
func (r *Router) setupUserRoutes(api fiber.Router) {
	// Initialize user service with repositories and logger
	userService := service.NewUserService(r.repos, r.config.Logger)

	// Initialize user handler
	userHandler := handler.NewUserHandler(userService)

	// Create users group
	users := api.Group("/users")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		// Use provided zap logger or create a default one
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger, _ = zap.NewProduction()
		}

		// Apply global rate limit to all user routes
		users.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Public routes (no authentication required)
	users.Post("/password-reset", userHandler.ResetPassword)
	users.Post("/password-reset/confirm", userHandler.ConfirmPasswordReset)

	// Protected routes (authentication required)
	// Base auth middleware validates token - applied to all routes below
	users.Use(r.RequireAuth())

	// Analytics - must be registered before /:id routes
	users.Get("/stats", middleware.RequireTenantOwnerOrAdmin(), userHandler.GetUserStats)
	users.Get("/registration-stats", middleware.RequireTenantOwnerOrAdmin(), userHandler.GetRegistrationStats)
	users.Get("/growth", middleware.RequireTenantOwnerOrAdmin(), userHandler.GetUserGrowth)

	// User Queries - must be registered before /:id routes
	users.Get("/", middleware.RequireTenantOwnerOrAdmin(), userHandler.ListUsers)
	users.Get("/search", middleware.RequireTenantOwnerOrAdmin(), userHandler.SearchUsers)
	users.Get("/by-role/:role", middleware.RequireTenantOwnerOrAdmin(), userHandler.GetUsersByRole)
	users.Get("/active", middleware.RequireTenantOwnerOrAdmin(), userHandler.GetActiveUsers)
	users.Get("/recently-active", middleware.RequireTenantOwnerOrAdmin(), userHandler.GetRecentlyActive)
	users.Get("/locked", middleware.RequireTenantOwnerOrAdmin(), userHandler.GetLockedUsers)
	users.Get("/marked-for-deletion", middleware.RequireTenantOwnerOrAdmin(), userHandler.GetUsersMarkedForDeletion)

	// CRUD Operations - parameterized routes registered after specific routes
	users.Post("/", middleware.RequireTenantOwnerOrAdmin(), userHandler.CreateUser)
	users.Get("/:id", middleware.RequireSelfOrAdmin(), userHandler.GetUser)
	users.Put("/:id", middleware.RequireSelfOrAdmin(), userHandler.UpdateUser)
	users.Delete("/:id", middleware.RequireTenantOwnerOrAdmin(), userHandler.DeleteUser)

	// Authentication & Security
	users.Put("/:id/password", middleware.RequireSelfOrAdmin(), userHandler.UpdatePassword)
	users.Post("/:id/verify-email", middleware.RequireTenantOwnerOrAdmin(), userHandler.VerifyEmail)
	users.Post("/:id/verify-phone", middleware.RequireTenantOwnerOrAdmin(), userHandler.VerifyPhone)
	users.Post("/:id/unlock", middleware.RequireTenantOwnerOrAdmin(), userHandler.UnlockUser)

	// MFA Management - self or admin
	users.Post("/:id/mfa/setup", middleware.RequireSelfOrAdmin(), userHandler.SetupMFA)
	users.Post("/:id/mfa/enable", middleware.RequireSelfOrAdmin(), userHandler.EnableMFA)
	users.Post("/:id/mfa/disable", middleware.RequireSelfOrAdmin(), userHandler.DisableMFA)
	users.Post("/:id/mfa/verify", middleware.RequireSelfOrAdmin(), userHandler.VerifyMFA)

	// Role & Status - tenant owner/admin only
	users.Put("/:id/role", middleware.RequireTenantOwnerOrAdmin(), userHandler.UpdateRole)
	users.Put("/:id/status", middleware.RequireTenantOwnerOrAdmin(), userHandler.UpdateStatus)
	users.Post("/:id/activate", middleware.RequireTenantOwnerOrAdmin(), userHandler.ActivateUser)
	users.Post("/:id/deactivate", middleware.RequireTenantOwnerOrAdmin(), userHandler.DeactivateUser)
	users.Post("/:id/suspend", middleware.RequireTenantOwnerOrAdmin(), userHandler.SuspendUser)

	// Profile Management - self or admin
	users.Put("/:id/profile", middleware.RequireSelfOrAdmin(), userHandler.UpdateProfile)
	users.Put("/:id/avatar", middleware.RequireSelfOrAdmin(), userHandler.UpdateAvatar)
	users.Put("/:id/preferences", middleware.RequireSelfOrAdmin(), userHandler.UpdatePreferences)

	// Compliance & GDPR - self for acceptance, admin for deletion
	users.Post("/:id/accept-terms", middleware.RequireSelfOrAdmin(), userHandler.AcceptTerms)
	users.Post("/:id/accept-privacy", middleware.RequireSelfOrAdmin(), userHandler.AcceptPrivacyPolicy)
	users.Put("/:id/consent", middleware.RequireSelfOrAdmin(), userHandler.UpdateConsent)
	users.Post("/:id/mark-for-deletion", middleware.RequireTenantOwnerOrAdmin(), userHandler.MarkForDeletion)
	users.Delete("/:id/permanent", middleware.RequireTenantOwnerOrAdmin(), userHandler.PermanentlyDeleteUser)

	r.config.Logger.Info("user routes registered successfully with authentication")
}

// UserRoutesInfo provides information about registered user routes
func UserRoutesInfo() map[string][]string {
	return map[string][]string{
		"Public Routes": {
			"POST /api/v1/users/password-reset",
			"POST /api/v1/users/password-reset/confirm",
		},
		"CRUD Operations": {
			"POST   /api/v1/users",
			"GET    /api/v1/users/:id",
			"PUT    /api/v1/users/:id",
			"DELETE /api/v1/users/:id",
		},
		"User Queries": {
			"GET /api/v1/users",
			"GET /api/v1/users/search",
			"GET /api/v1/users/by-role/:role",
			"GET /api/v1/users/active",
			"GET /api/v1/users/recently-active",
			"GET /api/v1/users/locked",
			"GET /api/v1/users/marked-for-deletion",
		},
		"Authentication & Security": {
			"PUT  /api/v1/users/:id/password",
			"POST /api/v1/users/:id/verify-email",
			"POST /api/v1/users/:id/verify-phone",
			"POST /api/v1/users/:id/unlock",
		},
		"MFA Management": {
			"POST /api/v1/users/:id/mfa/setup",
			"POST /api/v1/users/:id/mfa/enable",
			"POST /api/v1/users/:id/mfa/disable",
			"POST /api/v1/users/:id/mfa/verify",
		},
		"Role & Status": {
			"PUT  /api/v1/users/:id/role",
			"PUT  /api/v1/users/:id/status",
			"POST /api/v1/users/:id/activate",
			"POST /api/v1/users/:id/deactivate",
			"POST /api/v1/users/:id/suspend",
		},
		"Profile Management": {
			"PUT /api/v1/users/:id/profile",
			"PUT /api/v1/users/:id/avatar",
			"PUT /api/v1/users/:id/preferences",
		},
		"Compliance & GDPR": {
			"POST   /api/v1/users/:id/accept-terms",
			"POST   /api/v1/users/:id/accept-privacy",
			"PUT    /api/v1/users/:id/consent",
			"POST   /api/v1/users/:id/mark-for-deletion",
			"DELETE /api/v1/users/:id/permanent",
		},
		"Analytics": {
			"GET /api/v1/users/stats",
			"GET /api/v1/users/registration-stats",
			"GET /api/v1/users/growth",
		},
	}
}
