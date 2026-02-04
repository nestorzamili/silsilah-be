package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service/changerequest"
)

type ChangeRequestHandler struct {
	crService changerequest.Service
}

func NewChangeRequestHandler(crService changerequest.Service) *ChangeRequestHandler {
	return &ChangeRequestHandler{crService: crService}
}

func (h *ChangeRequestHandler) Create(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	var input domain.CreateChangeRequestInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	cr, err := h.crService.Create(c.Context(), userID, input)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(cr)
}

func (h *ChangeRequestHandler) List(c *fiber.Ctx) error {
	params := getPaginationParams(c)

	var status *domain.ChangeRequestStatus
	if s := c.Query("status"); s != "" {
		st := domain.ChangeRequestStatus(s)
		status = &st
	}

	result, err := h.crService.List(c.Context(), status, params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *ChangeRequestHandler) Get(c *fiber.Ctx) error {
	requestIDStr := c.Params("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid request ID")
	}

	cr, err := h.crService.GetByID(c.Context(), requestID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(cr)
}

func (h *ChangeRequestHandler) Approve(c *fiber.Ctx) error {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	if user.Role != string(domain.RoleEditor) && user.Role != string(domain.RoleDeveloper) {
		return middleware.Forbidden("Insufficient permissions")
	}

	requestIDStr := c.Params("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid request ID")
	}

	var input domain.ReviewChangeRequestInput
	_ = c.BodyParser(&input)

	meta := &changerequest.RequestMeta{
		IPAddress: middleware.GetIPAddress(c),
		UserAgent: middleware.GetUserAgentFromContext(c),
	}

	if err := h.crService.Approve(c.Context(), requestID, user.ID, input.Note, meta); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Change request approved"})
}

func (h *ChangeRequestHandler) Reject(c *fiber.Ctx) error {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	if user.Role != string(domain.RoleEditor) && user.Role != string(domain.RoleDeveloper) {
		return middleware.Forbidden("Insufficient permissions")
	}

	requestIDStr := c.Params("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid request ID")
	}

	var input domain.ReviewChangeRequestInput
	_ = c.BodyParser(&input)

	meta := &changerequest.RequestMeta{
		IPAddress: middleware.GetIPAddress(c),
		UserAgent: middleware.GetUserAgentFromContext(c),
	}

	if err := h.crService.Reject(c.Context(), requestID, user.ID, input.Note, meta); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Change request rejected"})
}
