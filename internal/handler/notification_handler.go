package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(c)

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

	result, err := h.notificationService.List(c.Context(), userID, unreadOnly, params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(c)

	count, err := h.notificationService.GetUnreadCount(c.Context(), userID)
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

	if err := h.notificationService.MarkAsRead(c.Context(), notifID); err != nil {
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(c)

	if err := h.notificationService.MarkAllAsRead(c.Context(), userID); err != nil {
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}
