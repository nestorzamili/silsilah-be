package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/service"
)

const (
	UserContextKey   = "user"
	UserIDContextKey = "user_id"
)

func AuthRequired(authService service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "Missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "Invalid authorization header format",
			})
		}

		token := parts[1]
		claims, err := authService.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "Invalid or expired token",
			})
		}

		user, err := authService.GetUserByID(c.Context(), claims.UserID)
		if err != nil || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "User not found",
			})
		}

		c.Locals(UserContextKey, user)
		c.Locals(UserIDContextKey, user.ID)

		return c.Next()
	}
}

func GetCurrentUser(c *fiber.Ctx) *domain.User {
	user, ok := c.Locals(UserContextKey).(*domain.User)
	if !ok {
		return nil
	}
	return user
}

func GetCurrentUserID(c *fiber.Ctx) uuid.UUID {
	userID, ok := c.Locals(UserIDContextKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}
