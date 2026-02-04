package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service/comment"
)

type CommentHandler struct {
	commentService comment.Service
}

func NewCommentHandler(commentService comment.Service) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

func (h *CommentHandler) Create(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	var input domain.CreateCommentInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	comment, err := h.commentService.Create(c.Context(), personID, userID, input)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *CommentHandler) List(c *fiber.Ctx) error {
	personIDStr := c.Params("personId")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid person ID")
	}

	params := getPaginationParams(c)

	result, err := h.commentService.ListByPerson(c.Context(), personID, params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *CommentHandler) Update(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}
	commentIDStr := c.Params("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid comment ID")
	}

	var input domain.UpdateCommentInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	comment, err := h.commentService.Update(c.Context(), userID, commentID, input)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(comment)
}

func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}
	commentIDStr := c.Params("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid comment ID")
	}

	if err := h.commentService.Delete(c.Context(), userID, commentID); err != nil {
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}
