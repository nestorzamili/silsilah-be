package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service"
)

type PersonHandler struct {
	personService service.PersonService
}

func NewPersonHandler(personService service.PersonService) *PersonHandler {
	return &PersonHandler{personService: personService}
}

func (h *PersonHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(c)

	var input domain.CreatePersonInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	person, err := h.personService.Create(c.Context(), userID, input)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(person)
}

func (h *PersonHandler) List(c *fiber.Ctx) error {
	params := getPaginationParams(c)

	result, err := h.personService.List(c.Context(), params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *PersonHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	limit := c.QueryInt("limit", 10)

	if len(query) < 2 {
		return middleware.BadRequest("Search query must be at least 2 characters")
	}

	persons, err := h.personService.Search(c.Context(), query, limit)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(persons)
}

func (h *PersonHandler) Get(c *fiber.Ctx) error {
	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	person, err := h.personService.GetByIDWithRelationships(c.Context(), personID)
	if err != nil {
		if err == service.ErrPersonNotFound {
			return middleware.NotFound("Person not found")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(person)
}

func (h *PersonHandler) Update(c *fiber.Ctx) error {
	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	userID := middleware.GetCurrentUserID(c)

	var input domain.UpdatePersonInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	person, err := h.personService.Update(c.Context(), personID, userID, input)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(person)
}

func (h *PersonHandler) Delete(c *fiber.Ctx) error {
	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	if err := h.personService.Delete(c.Context(), personID); err != nil {
		if err == service.ErrPersonNotFound {
			return middleware.NotFound("Person not found")
		}
		if strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "constraint") {
			return middleware.NewError(fiber.StatusInternalServerError, "Failed to delete person")
		}
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}

func getPaginationParams(c *fiber.Ctx) domain.PaginationParams {
	params := domain.DefaultPagination()

	if page := c.QueryInt("page", 1); page > 0 {
		params.Page = page
	}
	if pageSize := c.QueryInt("page_size", 20); pageSize > 0 {
		params.PageSize = pageSize
	}

	params.Validate()
	return params
}
