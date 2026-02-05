package router

import (
	"Krafti_Vibe/internal/handler"
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

	// ============================================================================
	// Core Invitation Operations
	// ============================================================================

	// Create invitation
	invitations.Post("",
		r.RequireAuth(),
		invitationHandler.CreateInvitation,
	)

	// List invitations
	invitations.Get("",
		r.RequireAuth(),
		invitationHandler.ListInvitations,
	)

	// Get invitation by ID
	invitations.Get("/:id",
		r.RequireAuth(),
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
		r.RequireAuth(),
		invitationHandler.GetPendingInvitations,
	)

	// Get invitations by email
	invitations.Get("/by-email",
		r.RequireAuth(),
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
		r.RequireAuth(),
		invitationHandler.RevokeInvitation,
	)

	// Resend invitation
	invitations.Post("/:id/resend",
		r.RequireAuth(),
		invitationHandler.ResendInvitation,
	)

	// ============================================================================
	// Cleanup Operations
	// ============================================================================

	// Delete expired invitations
	invitations.Delete("/cleanup/expired",
		r.RequireAuth(),
		invitationHandler.DeleteExpiredInvitations,
	)
}
