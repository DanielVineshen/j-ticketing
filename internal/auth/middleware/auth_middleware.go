// FILE: internal/auth/middleware/auth_middleware.go
package middleware

import (
	"j-ticketing/internal/auth/jwt"

	"github.com/gofiber/fiber/v2"
)

// Protected returns a middleware that protects routes by verifying JWT tokens
func Protected(jwtService jwt.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized - No token provided",
			})
		}

		// Extract token from header
		tokenString := jwtService.ExtractTokenFromHeader(authHeader)
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized - Invalid token format",
			})
		}

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized - " + err.Error(),
			})
		}

		// Add claims to context for use in handlers
		c.Locals("userId", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("userType", claims.UserType)
		c.Locals("role", claims.Role)
		c.Locals("roles", claims.Roles)

		// Continue to the route handler
		return c.Next()
	}
}

// HasRole returns a middleware that restricts access based on user's role
func HasRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check context for user's role
		userRole := c.Locals("role")
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized - Role information missing",
			})
		}

		// Convert role to string and compare with required role
		if userRole.(string) != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Forbidden - Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// HasAnyRole returns a middleware that restricts access based on if user has any of the specified roles
func HasAnyRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check context for user's role
		userRole := c.Locals("role")
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized - Role information missing",
			})
		}

		// Convert role to string and check if it's in the required roles
		userRoleStr := userRole.(string)
		for _, role := range roles {
			if userRoleStr == role {
				return c.Next()
			}
		}

		// Check if user has roles array
		if userRoles, ok := c.Locals("roles").([]string); ok {
			for _, userRole := range userRoles {
				for _, requiredRole := range roles {
					if userRole == requiredRole {
						return c.Next()
					}
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden - Insufficient permissions",
		})
	}
}
