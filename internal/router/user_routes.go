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
	// Base auth middleware validates token
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// CRUD Operations - require user:manage scope
	users.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.CreateUser,
	)

	users.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.GetUser,
	)

	users.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.UpdateUser,
	)

	users.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserDelete),
		userHandler.DeleteUser,
	)

	// User Queries - require user:read scope
	users.Get("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.ListUsers,
	)

	users.Get("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.SearchUsers,
	)

	users.Get("/by-role/:role",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.GetUsersByRole,
	)

	users.Get("/active",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.GetActiveUsers,
	)

	users.Get("/recently-active",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.GetRecentlyActive,
	)

	users.Get("/locked",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.GetLockedUsers,
	)

	users.Get("/marked-for-deletion",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.GetUsersMarkedForDeletion,
	)

	// Authentication & Security - require user:manage scope
	users.Put("/:id/password",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.UpdatePassword,
	)

	users.Post("/:id/verify-email",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.VerifyEmail,
	)

	users.Post("/:id/verify-phone",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.VerifyPhone,
	)

	users.Post("/:id/unlock",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.UnlockUser,
	)

	// MFA Management - require user:write scope
	users.Post("/:id/mfa/setup",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.SetupMFA,
	)

	users.Post("/:id/mfa/enable",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.EnableMFA,
	)

	users.Post("/:id/mfa/disable",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.DisableMFA,
	)

	users.Post("/:id/mfa/verify",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.VerifyMFA,
	)

	// Role & Status - require user:manage scope
	users.Put("/:id/role",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.UpdateRole,
	)

	users.Put("/:id/status",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.UpdateStatus,
	)

	users.Post("/:id/activate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.ActivateUser,
	)

	users.Post("/:id/deactivate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.DeactivateUser,
	)

	users.Post("/:id/suspend",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.SuspendUser,
	)

	// Profile Management - require user:write scope
	users.Put("/:id/profile",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.UpdateProfile,
	)

	users.Put("/:id/avatar",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.UpdateAvatar,
	)

	users.Put("/:id/preferences",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.UpdatePreferences,
	)

	// Compliance & GDPR - require user:write scope
	users.Post("/:id/accept-terms",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.AcceptTerms,
	)

	users.Post("/:id/accept-privacy",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.AcceptPrivacyPolicy,
	)

	users.Put("/:id/consent",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserWrite),
		userHandler.UpdateConsent,
	)

	users.Post("/:id/mark-for-deletion",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserManage),
		userHandler.MarkForDeletion,
	)

	users.Delete("/:id/permanent",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserDelete),
		userHandler.PermanentlyDeleteUser,
	)

	// Analytics - require admin or reporting permissions
	users.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.GetUserStats,
	)

	users.Get("/registration-stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.GetRegistrationStats,
	)

	users.Get("/growth",
		authMiddleware,
		middleware.RequireScopes(r.scopes.UserRead),
		userHandler.GetUserGrowth,
	)

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
