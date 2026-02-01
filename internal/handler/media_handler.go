package handler

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service"
)

type MediaHandler struct {
	mediaService         service.MediaService
	changeRequestService service.ChangeRequestService
}

func NewMediaHandler(mediaService service.MediaService, changeRequestService service.ChangeRequestService) *MediaHandler {
	return &MediaHandler{
		mediaService:         mediaService,
		changeRequestService: changeRequestService,
	}
}

func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(c)
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return middleware.BadRequest("File is required")
	}

	if file.Size > 10*1024*1024 {
		return middleware.BadRequest("File size must be less than 10MB")
	}

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	var personID *uuid.UUID
	if pidStr := c.FormValue("person_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err == nil {
			personID = &pid
		}
	}

	var caption *string
	if cap := c.FormValue("caption"); cap != "" {
		caption = &cap
	}

	fileReader, err := file.Open()
	if err != nil {
		return middleware.BadRequest("Failed to read file")
	}
	defer fileReader.Close()

	status := "active"
	if currentUser.Role == "member" {
		status = "pending"
	}

	media, err := h.mediaService.Upload(c.Context(), userID, personID, caption, file.Filename, file.Size, mimeType, fileReader, status)
	if err != nil {
		return err
	}

	if status == "pending" {
		var requesterNote *string
		if note := c.FormValue("requester_note"); note != "" {
			requesterNote = &note
		}

		input := domain.CreateChangeRequestInput{
			EntityType:    domain.EntityMedia,
			EntityID:      &media.ID,
			Action:        domain.ActionCreate,
			Payload:       json.RawMessage("{}"),
			RequesterNote: requesterNote,
		}

		request, err := h.changeRequestService.Create(c.Context(), userID, input)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":        "Upload request submitted for approval",
			"change_request": request,
			"media":          media,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(media)
}

func (h *MediaHandler) List(c *fiber.Ctx) error {
	params := getPaginationParams(c)

	var personID *uuid.UUID
	if pidStr := c.Query("person_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err == nil {
			personID = &pid
		}
	}

	result, err := h.mediaService.List(c.Context(), personID, params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *MediaHandler) Get(c *fiber.Ctx) error {
	mediaIDStr := c.Params("mediaId")
	mediaID, err := uuid.Parse(mediaIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid media ID")
	}

	media, err := h.mediaService.GetByID(c.Context(), mediaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return middleware.NotFound("Media not found")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(media)
}

func (h *MediaHandler) Delete(c *fiber.Ctx) error {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return middleware.Unauthorized("User not authenticated")
	}

	mediaIDStr := c.Params("mediaId")
	mediaID, err := uuid.Parse(mediaIDStr)
	if err != nil {
		return middleware.BadRequest("Invalid media ID")
	}

	media, err := h.mediaService.GetByID(c.Context(), mediaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return middleware.NotFound("Media not found")
		}
		return err
	}

	if currentUser.Role == "member" {
		var requesterNote *string
		requesterNoteStr := c.Query("requester_note")
		if requesterNoteStr != "" {
			requesterNote = &requesterNoteStr
		}

		input := domain.CreateChangeRequestInput{
			EntityType:    domain.EntityMedia,
			EntityID:      &mediaID,
			Action:        domain.ActionDelete,
			Payload:       json.RawMessage("{}"),
			RequesterNote: requesterNote,
		}

		request, err := h.changeRequestService.Create(c.Context(), currentUser.ID, input)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":        "Delete request submitted for approval",
			"change_request": request,
		})
	}

	if currentUser.Role != "editor" && currentUser.Role != "developer" {
		if currentUser.ID != media.UploadedBy {
			return middleware.Forbidden("You don't have permission to delete this media")
		}
	}

	if err := h.mediaService.Delete(c.Context(), mediaID); err != nil {
		return err
	}

	return c.Status(fiber.StatusNoContent).SendString("")
}
