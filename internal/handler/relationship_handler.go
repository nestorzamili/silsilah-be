package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service"
)

type RelationshipHandler struct {
	relService service.RelationshipService
}

func NewRelationshipHandler(relService service.RelationshipService) *RelationshipHandler {
	return &RelationshipHandler{relService: relService}
}

func (h *RelationshipHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(c)

	var input domain.CreateRelationshipInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	rel, err := h.relService.Create(c.Context(), userID, input)
	if err != nil {
		switch err {
		case service.ErrSelfRelation:
			return middleware.BadRequest("Cannot create relationship with self")
		case service.ErrInvalidRelationType:
			return middleware.BadRequest("Invalid relationship type")
		case service.ErrPersonNotFound:
			return middleware.NotFound("One or both persons not found")
		case service.ErrDuplicateRelationship:
			return middleware.Conflict("Relationship already exists between these two persons")
		case service.ErrDuplicateParentRole:
			return middleware.Conflict("Person already has a parent with this role (father/mother)")
		}
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(rel)
}

func (h *RelationshipHandler) List(c *fiber.Ctx) error {
	var relType *domain.RelationshipType
	if t := c.Query("type"); t != "" {
		rt := domain.RelationshipType(t)
		relType = &rt
	}

	relationships, err := h.relService.List(c.Context(), relType)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(relationships)
}

func (h *RelationshipHandler) Get(c *fiber.Ctx) error {
	relIDStr := c.Params("relationshipId")
	relID, err := uuid.Parse(relIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid relationship ID")
	}

	rel, err := h.relService.GetByID(c.Context(), relID)
	if err != nil {
		if err == service.ErrRelationshipNotFound {
			return middleware.NotFound("Relationship not found")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(rel)
}

func (h *RelationshipHandler) Update(c *fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(c)

	relIDStr := c.Params("relationshipId")
	relID, err := uuid.Parse(relIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid relationship ID")
	}

	var input domain.UpdateRelationshipInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	rel, err := h.relService.Update(c.Context(), userID, relID, input)
	if err != nil {
		if err == service.ErrRelationshipNotFound {
			return middleware.NotFound("Relationship not found")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(rel)
}

func (h *RelationshipHandler) Delete(c *fiber.Ctx) error {
	relIDStr := c.Params("relationshipId")
	relID, err := uuid.Parse(relIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid relationship ID")
	}

	if err := h.relService.Delete(c.Context(), relID); err != nil {
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}
