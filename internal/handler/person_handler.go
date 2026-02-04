package handler

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service/changerequest"
	"silsilah-keluarga/internal/service/person"
)

type PersonHandler struct {
	personService person.Service
	crService     changerequest.Service
}

func NewPersonHandler(personService person.Service, crService changerequest.Service) *PersonHandler {
	return &PersonHandler{
		personService: personService,
		crService:     crService,
	}
}

func (h *PersonHandler) Create(c *fiber.Ctx) error {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	var input domain.CreatePersonInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	if user.Role == string(domain.RoleMember) {
		payload, err := json.Marshal(input)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to marshal payload")
		}

		var requesterNote *string
		if note := c.Query("requester_note"); note != "" {
			requesterNote = &note
		}

		crInput := domain.CreateChangeRequestInput{
			EntityType:    domain.EntityPerson,
			Action:        domain.ActionCreate,
			Payload:       payload,
			RequesterNote: requesterNote,
		}

		cr, err := h.crService.Create(c.Context(), user.ID, crInput)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":        "Create request submitted for approval",
			"change_request": cr,
		})
	}

	person, err := h.personService.Create(c.Context(), user.ID, input)
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
		if errors.Is(err, domain.ErrPersonNotFound) {
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

	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	var input domain.UpdatePersonInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	if user.Role == string(domain.RoleMember) {
		payload, err := json.Marshal(input)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to marshal payload")
		}

		var requesterNote *string
		if note := c.Query("requester_note"); note != "" {
			requesterNote = &note
		}

		crInput := domain.CreateChangeRequestInput{
			EntityType:    domain.EntityPerson,
			EntityID:      &personID,
			Action:        domain.ActionUpdate,
			Payload:       payload,
			RequesterNote: requesterNote,
		}

		cr, err := h.crService.Create(c.Context(), user.ID, crInput)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":        "Update request submitted for approval",
			"change_request": cr,
		})
	}

	person, err := h.personService.Update(c.Context(), user.ID, personID, input)
	if err != nil {
		if errors.Is(err, domain.ErrPersonNotFound) {
			return middleware.NotFound("Person not found")
		}
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

	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	if user.Role == string(domain.RoleMember) {
		payload := json.RawMessage("{}")

		var requesterNote *string
		if note := c.Query("requester_note"); note != "" {
			requesterNote = &note
		}

		crInput := domain.CreateChangeRequestInput{
			EntityType:    domain.EntityPerson,
			EntityID:      &personID,
			Action:        domain.ActionDelete,
			Payload:       payload,
			RequesterNote: requesterNote,
		}

		cr, err := h.crService.Create(c.Context(), user.ID, crInput)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":        "Delete request submitted for approval",
			"change_request": cr,
		})
	}

	if err := h.personService.Delete(c.Context(), personID); err != nil {
		if errors.Is(err, domain.ErrPersonNotFound) {
			return middleware.NotFound("Person not found")
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
