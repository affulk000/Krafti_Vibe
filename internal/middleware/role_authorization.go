package middleware

import (
	"Krafti_Vibe/internal/domain/models"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
)

// Role authorization middleware helpers
// These work with both Zitadel roles (platform) and database roles (tenant)

// Platform-level role requirements (from Zitadel)
func (m *ZitadelAuthMiddleware) RequirePlatformSuperAdmin() fiber.Handler {
	return m.RequireAuth(authorization.WithRole(string(models.UserRolePlatformSuperAdmin)))
}

func (m *ZitadelAuthMiddleware) RequirePlatformAdmin() fiber.Handler {
	return m.RequireAuth(authorization.WithRole(string(models.UserRolePlatformAdmin)))
}

func (m *ZitadelAuthMiddleware) RequirePlatformSupport() fiber.Handler {
	return m.RequireAuth(authorization.WithRole(string(models.UserRolePlatformSupport)))
}

// RequireAnyPlatformRole requires any platform-level role
// NOTE: This expects RequireAuth() to have already been called (usually at group level)
// It only performs the platform role check, not authentication
func (m *ZitadelAuthMiddleware) RequireAnyPlatformRole() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if user is platform user
		// (auth should have already been done by group-level middleware)
		dbUser := c.Locals("db_user")
		if dbUser == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Platform access required",
				},
			})
		}

		user, ok := dbUser.(*models.User)
		if !ok || !user.IsPlatformUser {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Platform access required",
				},
			})
		}

		return c.Next()
	}
}

// Tenant-level role requirements (from database)
// These check the database user's role, not Zitadel roles

func RequireTenantOwner() fiber.Handler {
	return requireDatabaseRole(models.UserRoleTenantOwner)
}

func RequireTenantAdmin() fiber.Handler {
	return requireDatabaseRole(models.UserRoleTenantAdmin)
}

func RequireArtisan() fiber.Handler {
	return requireDatabaseRole(models.UserRoleArtisan)
}

func RequireTeamMember() fiber.Handler {
	return requireDatabaseRole(models.UserRoleTeamMember)
}

// RequireTenantOwnerOrAdmin allows either tenant owner or admin
func RequireTenantOwnerOrAdmin() fiber.Handler {
	return requireAnyDatabaseRole(models.UserRoleTenantOwner, models.UserRoleTenantAdmin)
}

// RequireArtisanOrTeamMember allows artisan or team member
func RequireArtisanOrTeamMember() fiber.Handler {
	return requireAnyDatabaseRole(models.UserRoleArtisan, models.UserRoleTeamMember)
}

// Helper function to require a specific database role
func requireDatabaseRole(role models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dbUser := c.Locals("db_user")
		if dbUser == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "User not synchronized",
				},
			})
		}

		user, ok := dbUser.(*models.User)
		if !ok || user == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid user context",
				},
			})
		}

		// Platform users have access to everything
		if user.IsPlatformUser {
			return c.Next()
		}

		// Check if user has the required role
		if user.Role != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "You do not have permission to access this resource",
				},
			})
		}

		return c.Next()
	}
}

// Helper function to require any of the specified database roles
func requireAnyDatabaseRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dbUser := c.Locals("db_user")
		if dbUser == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "User not synchronized",
				},
			})
		}

		user, ok := dbUser.(*models.User)
		if !ok || user == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid user context",
				},
			})
		}

		// Platform users have access to everything
		if user.IsPlatformUser {
			return c.Next()
		}

		// Check if user has any of the required roles
		if !slices.Contains(roles, user.Role) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "You do not have permission to access this resource",
				},
			})
		}

		return c.Next()
	}
}

// RequireSelfOrAdmin ensures user can only access their own data unless they're an admin
func RequireSelfOrAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		dbUser := c.Locals("db_user")
		if dbUser == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "User not synchronized",
				},
			})
		}

		user, ok := dbUser.(*models.User)
		if !ok || user == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid user context",
				},
			})
		}

		// Platform users can access any user's data
		if user.IsPlatformUser {
			return c.Next()
		}

		// Tenant owners and admins can access any user in their tenant
		if user.Role == models.UserRoleTenantOwner || user.Role == models.UserRoleTenantAdmin {
			// TODO: Add tenant validation here
			return c.Next()
		}

		// Regular users can only access their own data
		userID := c.Params("id")
		if userID != user.ID.String() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "You can only access your own data",
				},
			})
		}

		return c.Next()
	}
}

// GetDatabaseUser helper to extract database user from context
func GetDatabaseUser(c *fiber.Ctx) (*models.User, bool) {
	dbUser := c.Locals("db_user")
	if dbUser == nil {
		return nil, false
	}
	user, ok := dbUser.(*models.User)
	if !ok || user == nil {
		return nil, false
	}
	return user, true
}

// IsPlatformUser checks if the current user is a platform user
func IsPlatformUser(c *fiber.Ctx) bool {
	user, ok := GetDatabaseUser(c)
	if !ok {
		return false
	}
	return user.IsPlatformUser
}

// HasDatabaseRole checks if user has a specific database role
func HasDatabaseRole(c *fiber.Ctx, role models.UserRole) bool {
	user, ok := GetDatabaseUser(c)
	if !ok {
		return false
	}
	// Platform users have all permissions
	if user.IsPlatformUser {
		return true
	}
	return user.Role == role
}

// HasAnyDatabaseRole checks if user has any of the specified database roles
func HasAnyDatabaseRole(c *fiber.Ctx, roles ...models.UserRole) bool {
	user, ok := GetDatabaseUser(c)
	if !ok {
		return false
	}
	// Platform users have all permissions
	if user.IsPlatformUser {
		return true
	}
	return slices.Contains(roles, user.Role)
}
