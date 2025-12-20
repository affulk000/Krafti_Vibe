package middleware

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetContext retrieves the authentication context from Fiber locals
// This is an alias for GetAuthContext for backward compatibility
func GetContext(c *fiber.Ctx) (*AuthContext, bool) {
	return GetAuthContext(c)
}

// GetUserID retrieves the user ID from the authentication context
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	authCtx, ok := GetAuthContext(c)
	if !ok {
		return uuid.Nil, false
	}
	return authCtx.UserID, true
}

// MustGetUserID retrieves the user ID or panics if not found
func MustGetUserID(c *fiber.Ctx) uuid.UUID {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user ID not found - did you forget to use the auth middleware?")
	}
	return userID
}

// HasRole checks if the user has a specific role
func HasRole(c *fiber.Ctx, role string) bool {
	authCtx, ok := GetAuthContext(c)
	if !ok {
		return false
	}
	return slices.Contains(authCtx.Roles, role)
}

// HasAnyRole checks if the user has any of the specified roles
func HasAnyRole(c *fiber.Ctx, roles ...string) bool {
	authCtx, ok := GetAuthContext(c)
	if !ok {
		return false
	}
	for _, userRole := range authCtx.Roles {
		if slices.Contains(roles, userRole) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the user has all of the specified roles
func HasAllRoles(c *fiber.Ctx, roles ...string) bool {
	authCtx, ok := GetAuthContext(c)
	if !ok {
		return false
	}
	for _, role := range roles {
		if !slices.Contains(authCtx.Roles, role) {
			return false
		}
	}
	return true
}

// HasScope checks if the user has a specific scope
func HasScope(c *fiber.Ctx, scope string) bool {
	authCtx, ok := GetAuthContext(c)
	if !ok {
		return false
	}
	return slices.Contains(authCtx.Scopes, scope)
}
