package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetCurrentUser(c)
		if user == nil {
			return Unauthorized("User not found")
		}

		if !user.HasRole(requiredRole) {
			return Forbidden("Insufficient permissions for this operation")
		}

		return c.Next()
	}
}

func RequireAnyRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetCurrentUser(c)
		if user == nil {
			return Unauthorized("User not found")
		}

		hasPermission := false
		for _, role := range roles {
			if user.HasRole(role) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return Forbidden("Insufficient permissions for this operation")
		}

		return c.Next()
	}
}

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetCurrentUser(c)
		if user == nil {
			return Unauthorized("User not found")
		}

		if !hasPermission(user.Role, permission) {
			return Forbidden("Insufficient permissions for this operation")
		}

		return c.Next()
	}
}

func hasPermission(role, permission string) bool {
	permissions := map[string]map[string]bool{
		"member": {
			"view_family_tree":    true,
			"view_person":         true,
			"add_comment":         true,
			"view_media":          true,
			"view_comments":       true,
			"update_comment":      true,
			"delete_comment":      true,
			"upload_media":        true,
			"delete_media":        true,
		},
		"editor": {
			"view_family_tree":    true,
			"view_person":         true,
			"add_comment":         true,
			"view_media":          true,
			"view_comments":       true,
			"create_person":       true,
			"update_person":       true,
			"delete_person":       true,
			"create_relationship": true,
			"update_relationship": true,
			"delete_relationship": true,
			"upload_media":        true,
			"delete_media":        true,
			"update_comment":      true,
			"delete_comment":      true,
		},
		"developer": {
			"view_family_tree":    true,
			"view_person":         true,
			"add_comment":         true,
			"view_media":          true,
			"view_comments":       true,
			"create_person":       true,
			"update_person":       true,
			"delete_person":       true,
			"create_relationship": true,
			"update_relationship": true,
			"delete_relationship": true,
			"upload_media":        true,
			"delete_media":        true,
			"update_comment":      true,
			"delete_comment":      true,
			"assign_roles":        true,
			"view_audit_logs":     true,
			"system_configuration": true,
		},
	}

	if rolePermissions, exists := permissions[role]; exists {
		return rolePermissions[permission]
	}
	return false
}

func GetCurrentUserRole(c *fiber.Ctx) string {
	user := GetCurrentUser(c)
	if user == nil {
		return ""
	}
	return user.Role
}

func IsDeveloper(c *fiber.Ctx) bool {
	return GetCurrentUserRole(c) == "developer"
}

func IsEditor(c *fiber.Ctx) bool {
	role := GetCurrentUserRole(c)
	return role == "editor" || role == "developer"
}

func IsMember(c *fiber.Ctx) bool {
	role := GetCurrentUserRole(c)
	return role == "member" || role == "editor" || role == "developer"
}