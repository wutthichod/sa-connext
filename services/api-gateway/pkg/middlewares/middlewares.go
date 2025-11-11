package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/shared/auth"
	"github.com/wutthichod/sa-connext/shared/config"
)

// JWTMiddleware checks for JWT in cookie, Authorization header, or query parameter and validates it
func JWTMiddleware(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Get JWT from cookie first, then fallback to Authorization header, then query parameter
		token := c.Cookies("token")
		if token == "" {
			// Try Authorization header as fallback
			authHeader := c.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}
		}
		if token == "" {
			// Try query parameter (useful for WebSocket connections)
			token = c.Query("token")
		}

		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token")
		}

		// 2. Validate token
		claims, err := auth.ValidateToken(cfg.JWT().Token, token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		// 3. Store user info in Locals for next handlers
		c.Locals("userID", claims.UserID)

		// 4. Continue to next handler
		return c.Next()
	}
}
