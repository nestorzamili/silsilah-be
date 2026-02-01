package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service"
)

type UserHandler struct {
	userService   service.UserService
	personService service.PersonService
}

func NewUserHandler(userService service.UserService, personService service.PersonService) *UserHandler {
	return &UserHandler{
		userService:   userService,
		personService: personService,
	}
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not found")
	}
	return c.JSON(user)
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(c)

	var input domain.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	var body map[string]interface{}
	_ = c.BodyParser(&body)
	if val, ok := body["person_id"]; ok && val == nil {
		nullUUID := (*uuid.UUID)(nil)
		input.PersonID = &nullUUID
	}

	user, err := h.userService.Update(c.Context(), userID, input)
	if err != nil {
		return err
	}

	return c.JSON(user)
}

func (h *UserHandler) AssignRole(c *fiber.Ctx) error {
	current_user := middleware.GetCurrentUser(c)
	if current_user == nil {
		return middleware.Unauthorized("User not found")
	}

	var input domain.AssignRoleInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	if err := h.userService.AssignRole(c.Context(), current_user, input); err != nil {
		if err == service.ErrInsufficientPermissions {
			return middleware.Forbidden("Insufficient permissions to assign roles")
		}
		if err == service.ErrCannotModifySelf {
			return middleware.Forbidden("Cannot modify your own role")
		}
		if err == service.ErrUserNotFound {
			return middleware.NotFound("User not found")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Role assigned successfully",
	})
}

func (h *UserHandler) ListByRole(c *fiber.Ctx) error {
	role := c.Params("role")

	if role != "member" && role != "editor" && role != "developer" {
		return middleware.BadRequest("Invalid role parameter")
	}

	users, err := h.userService.ListByRole(c.Context(), role)
	if err != nil {
		return err
	}

	return c.JSON(users)
}

func (h *UserHandler) GetRoleUsers(c *fiber.Ctx) error {
	result, err := h.userService.GetRoleUsers(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(result)
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	users, err := h.userService.GetAllUsers(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(users)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return middleware.Unauthorized("User not found")
	}

	userID := c.Params("id")
	if userID == "" {
		return middleware.BadRequest("User ID is required")
	}

	if err := h.userService.DeleteUser(c.Context(), currentUser, userID); err != nil {
		if err == service.ErrInsufficientPermissions {
			return middleware.Forbidden("Insufficient permissions to delete user")
		}
		if err == service.ErrCannotModifySelf {
			return middleware.Forbidden("Cannot delete your own account")
		}
		if err == service.ErrUserNotFound {
			return middleware.NotFound("User not found")
		}
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *UserHandler) GetAncestors(c *fiber.Ctx) error {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not found")
	}

	if user.PersonID == nil {
		return c.JSON([]domain.Person{})
	}

	ancestors, err := h.personService.GetAncestors(c.Context(), *user.PersonID)
	if err != nil {
		return err
	}

	return c.JSON(ancestors)
}
