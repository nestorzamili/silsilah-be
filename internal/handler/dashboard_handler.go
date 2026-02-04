package handler

import (
	"github.com/gofiber/fiber/v2"

	"silsilah-keluarga/internal/service/dashboard"
)

type DashboardHandler struct {
	dashboardService dashboard.Service
}

func NewDashboardHandler(dashboardService dashboard.Service) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.dashboardService.GetStats(c.Context())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}
