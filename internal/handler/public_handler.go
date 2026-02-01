package handler

import (
	"silsilah-keluarga/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PublicHandler struct {
	personService service.PersonService
	graphService  service.GraphService
}

func NewPublicHandler(
	personService service.PersonService,
	graphService service.GraphService,
) *PublicHandler {
	return &PublicHandler{
		personService: personService,
		graphService:  graphService,
	}
}

func (h *PublicHandler) GetGraph(c *fiber.Ctx) error {
	graph, err := h.graphService.GetFullGraph(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch graph")
	}

	return c.Status(fiber.StatusOK).JSON(graph)
}

func (h *PublicHandler) GetPerson(c *fiber.Ctx) error {
	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid person ID")
	}

	person, err := h.personService.GetByIDWithRelationships(c.Context(), personID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Person not found")
	}

	return c.Status(fiber.StatusOK).JSON(person)
}
