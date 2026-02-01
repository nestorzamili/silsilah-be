package handler

import (
	"github.com/gofiber/fiber/v2"

	"silsilah-keluarga/internal/service"
)

type AuditHandler struct {
	auditService service.AuditService
}

func NewAuditHandler(auditService service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

func (h *AuditHandler) GetRecentActivities(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}

	logs, err := h.auditService.GetRecentActivities(c.Context(), limit)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(logs)
}
