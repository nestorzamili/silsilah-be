package handler

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service/changerequest"
	"silsilah-keluarga/internal/service/relationship"
)

type RelationshipHandler struct {
	relService relationship.Service
	crService  changerequest.Service
}

func NewRelationshipHandler(relService relationship.Service, crService changerequest.Service) *RelationshipHandler {
	return &RelationshipHandler{
		relService: relService,
		crService:  crService,
	}
}

func (h *RelationshipHandler) Create(c *fiber.Ctx) error {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	var input domain.CreateRelationshipInput
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
			EntityType:    domain.EntityRelationship,
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

	rel, err := h.relService.Create(c.Context(), user.ID, input)
	if err != nil {
		switch err {
		case relationship.ErrSelfRelation:
			return middleware.BadRequest("Cannot create relationship with self")
		case relationship.ErrInvalidRelationType:
			return middleware.BadRequest("Invalid relationship type")
		case domain.ErrPersonNotFound:
			return middleware.NotFound("One or both persons not found")
		case relationship.ErrDuplicateRelationship:
			return middleware.Conflict("Relationship already exists between these two persons")
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
		if err == relationship.ErrRelationshipNotFound {
			return middleware.NotFound("Relationship not found")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(rel)
}

func (h *RelationshipHandler) Update(c *fiber.Ctx) error {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	relIDStr := c.Params("relationshipId")
	relID, err := uuid.Parse(relIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid relationship ID")
	}

	var input domain.UpdateRelationshipInput
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
			EntityType:    domain.EntityRelationship,
			EntityID:      &relID,
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

	rel, err := h.relService.Update(c.Context(), user.ID, relID, input)
	if err != nil {
		if err == relationship.ErrRelationshipNotFound {
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
			EntityType:    domain.EntityRelationship,
			EntityID:      &relID,
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

	if err := h.relService.Delete(c.Context(), relID); err != nil {
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}
