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
		r.zitadelMW.RequireAuth(),
		invitationHandler.CreateInvitation,
	)

	// List invitations
	invitations.Get("",
		r.zitadelMW.RequireAuth(),
		invitationHandler.ListInvitations,
	)

	// Get invitation by ID
	invitations.Get("/:id",
		r.zitadelMW.RequireAuth(),
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
		r.zitadelMW.RequireAuth(),
		invitationHandler.GetPendingInvitations,
	)

	// Get invitations by email
	invitations.Get("/by-email",
		r.zitadelMW.RequireAuth(),
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
		r.zitadelMW.RequireAuth(),
		invitationHandler.RevokeInvitation,
	)

	// Resend invitation
	invitations.Post("/:id/resend",
		r.zitadelMW.RequireAuth(),
		invitationHandler.ResendInvitation,
	)

	// ============================================================================
	// Cleanup Operations
	// ============================================================================

	// Delete expired invitations
	invitations.Delete("/cleanup/expired",
		r.zitadelMW.RequireAuth(),
		invitationHandler.DeleteExpiredInvitations,
	)
}
