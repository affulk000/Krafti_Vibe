package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupTenantInvitationRoutes sets up tenant invitation routes
func (r *Router) setupTenantInvitationRoutes(api fiber.Router) {
	// Initialize tenant invitation service
	invitationService := service.NewTenantInvitationService(r.repos, r.config.ZapLogger)

	// Initialize tenant invitation handler
	invitationHandler := handler.NewTenantInvitationHandler(invitationService)

	// Tenant invitation routes group
	invitations := api.Group("/invitations")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Invitation Operations
	// ============================================================================

	// Create invitation
	invitations.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		invitationHandler.CreateInvitation,
	)

	// List invitations
	invitations.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		invitationHandler.ListInvitations,
	)

	// Get invitation by ID
	invitations.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		invitationHandler.GetInvitation,
	)

	// ============================================================================
	// Token-Based Operations
	// ============================================================================

	// Get invitation by token (no auth required - user needs token to accept)
	invitations.Get("/token/:token",
		invitationHandler.GetInvitationByToken,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// Get pending invitations
	invitations.Get("/pending",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		invitationHandler.GetPendingInvitations,
	)

	// Get invitations by email
	invitations.Get("/by-email",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		invitationHandler.GetInvitationsByEmail,
	)

	// ============================================================================
	// Invitation Actions
	// ============================================================================

	// Accept invitation (no auth required - token-based)
	invitations.Post("/accept",
		invitationHandler.AcceptInvitation,
	)

	// Revoke invitation
	invitations.Post("/:id/revoke",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		invitationHandler.RevokeInvitation,
	)

	// Resend invitation
	invitations.Post("/:id/resend",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantManage),
		invitationHandler.ResendInvitation,
	)

	// ============================================================================
	// Cleanup Operations
	// ============================================================================

	// Delete expired invitations
	invitations.Delete("/cleanup/expired",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantAdmin),
		invitationHandler.DeleteExpiredInvitations,
	)
}
