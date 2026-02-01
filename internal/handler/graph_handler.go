package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service"
)

type GraphHandler struct {
	graphService service.GraphService
}

func NewGraphHandler(graphService service.GraphService) *GraphHandler {
	return &GraphHandler{graphService: graphService}
}

func (h *GraphHandler) GetFullGraph(c *fiber.Ctx) error {
	graph, err := h.graphService.GetFullGraph(c.Context())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(graph)
}

func (h *GraphHandler) GetAncestors(c *fiber.Ctx) error {
	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	maxDepth := c.QueryInt("max_depth", 10)

	ancestors, err := h.graphService.GetAncestors(c.Context(), personID, maxDepth)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ancestors)
}

func (h *GraphHandler) GetSplitAncestors(c *fiber.Ctx) error {
	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	maxDepth := c.QueryInt("max_depth", 10)

	splitAncestors, err := h.graphService.GetSplitAncestors(c.Context(), personID, maxDepth)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(splitAncestors)
}

func (h *GraphHandler) GetDescendants(c *fiber.Ctx) error {
	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	maxDepth := c.QueryInt("max_depth", 10)

	descendants, err := h.graphService.GetDescendants(c.Context(), personID, maxDepth)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(descendants)
}

func (h *GraphHandler) FindRelationshipPath(c *fiber.Ctx) error {
	fromIDStr := c.Query("from")
	toIDStr := c.Query("to")

	fromID, err := uuid.Parse(fromIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid 'from' person ID")
	}

	toID, err := uuid.Parse(toIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid 'to' person ID")
	}

	path, err := h.graphService.FindRelationshipPath(c.Context(), fromID, toID)
	if err != nil {
		return err
	}

	if path == nil {
		return middleware.NotFound("No relationship path found")
	}

	return c.Status(fiber.StatusOK).JSON(path)
}
