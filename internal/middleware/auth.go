package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/service"
)

const (
	UserContextKey      = "user"
	UserIDContextKey    = "user_id"
	IPAddressContextKey = "ip_address"
	UserAgentContextKey = "user_agent"
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

func GetRealIP(c *fiber.Ctx) string {
	if cfIP := c.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	if realIP := c.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	if xff := c.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	return c.IP()
}

func GetUserAgent(c *fiber.Ctx) string {
	return c.Get("User-Agent")
}

func RequestInfo() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(IPAddressContextKey, GetRealIP(c))
		c.Locals(UserAgentContextKey, GetUserAgent(c))
		return c.Next()
	}
}

func GetIPAddress(c *fiber.Ctx) string {
	ip, ok := c.Locals(IPAddressContextKey).(string)
	if !ok {
		return GetRealIP(c)
	}
	return ip
}

func GetUserAgentFromContext(c *fiber.Ctx) string {
	ua, ok := c.Locals(UserAgentContextKey).(string)
	if !ok {
		return GetUserAgent(c)
	}
	return ua
}
