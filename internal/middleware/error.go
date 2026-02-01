package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id,omitempty"`
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal server error"
	errorCode := "INTERNAL_ERROR"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
		
		switch code {
		case fiber.StatusBadRequest:
			errorCode = "BAD_REQUEST"
		case fiber.StatusUnauthorized:
			errorCode = "UNAUTHORIZED"
		case fiber.StatusForbidden:
			errorCode = "FORBIDDEN"
		case fiber.StatusNotFound:
			errorCode = "NOT_FOUND"
		case fiber.StatusConflict:
			errorCode = "CONFLICT"
		case fiber.StatusUnprocessableEntity:
			errorCode = "VALIDATION_ERROR"
		}
	}

	traceID := uuid.New().String()[:8]

	return c.Status(code).JSON(ErrorResponse{
		Code:    errorCode,
		Message: message,
		TraceID: traceID,
	})
}

func NewError(code int, message string) *fiber.Error {
	return fiber.NewError(code, message)
}

func BadRequest(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusBadRequest, message)
}

func Unauthorized(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusUnauthorized, message)
}

func Forbidden(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusForbidden, message)
}

func NotFound(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusNotFound, message)
}

func Conflict(message string) *fiber.Error {
	return fiber.NewError(fiber.StatusConflict, message)
}
