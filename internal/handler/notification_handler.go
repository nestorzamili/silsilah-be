package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service/notification"
)

type NotificationHandler struct {
	notifService notification.Service
}

func NewNotificationHandler(notifService notification.Service) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	unreadOnly := c.Query("unread_only") == "true"

	params := getPaginationParams(c)
	if page > 0 {
		params.Page = page
	}
	if pageSize > 0 && pageSize <= 100 {
		params.PageSize = pageSize
	}

	result, err := h.notifService.List(c.Context(), userID, unreadOnly, params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	count, err := h.notifService.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"count": count,
	})
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	notifIDStr := c.Params("id")
	notifID, err := uuid.Parse(notifIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid notification ID")
	}

	if err := h.notifService.MarkAsRead(c.Context(), notifID); err != nil {
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	if err := h.notifService.MarkAllAsRead(c.Context(), userID); err != nil {
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}
