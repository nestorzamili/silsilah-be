package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service/export"
)

type ExportHandler struct {
	exportSvc export.Service
}

func NewExportHandler(exportSvc export.Service) *ExportHandler {
	return &ExportHandler{exportSvc: exportSvc}
}

func (h *ExportHandler) ExportJSON(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized("User not authenticated")
	}

	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	data, err := h.exportSvc.ExportJSON(c.Context(), userID, personID)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("family_tree_%s.json", time.Now().Format("20060102_150405"))
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "application/json")

	return c.JSON(data)
}

func (h *ExportHandler) ExportGEDCOM(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	rootIDStr := c.Query("root_id")
	if rootIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "root_id is required")
	}
	rootID, err := uuid.Parse(rootIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid root_id")
	}

	gedcomData, err := h.exportSvc.ExportGEDCOM(c.Context(), userID, rootID)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("family_tree_%s.ged", time.Now().Format("20060102_150405"))
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "text/plain")

	return c.SendString(gedcomData)
}
