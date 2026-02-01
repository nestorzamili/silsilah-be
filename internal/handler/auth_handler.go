package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input domain.CreateUserInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	user, _, err := h.authService.Register(c.Context(), input)
	if err != nil {
		if err == service.ErrEmailExists {
			return middleware.Conflict("Email already registered")
		}
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":    user,
		"message": "Registration successful. Please check your email for verification.",
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input domain.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	user, tokens, err := h.authService.Login(c.Context(), input)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return middleware.Unauthorized("Invalid email or password")
		}
		if err == service.ErrEmailNotVerified {
			return middleware.Forbidden("Email not verified. Please verify your email first.")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":          user,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	tokens, err := h.authService.RefreshToken(c.Context(), input.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			return middleware.Unauthorized("Invalid refresh token")
		}
		if err == service.ErrUserNotFound {
			return middleware.Unauthorized("User not found")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	if err := h.authService.RequestPasswordReset(c.Context(), input.Email); err != nil {
		if strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "connection") {
			return middleware.NewError(fiber.StatusInternalServerError, "Failed to process request")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If the email exists, a reset link has been sent",
	})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var input struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	if err := h.authService.ResetPassword(c.Context(), input.Token, input.NewPassword); err != nil {
		if err == service.ErrInvalidToken || err == service.ErrTokenExpired {
			return middleware.BadRequest("Invalid or expired reset token")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password has been reset successfully",
	})
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return middleware.BadRequest("Verification token is required")
	}

	if err := h.authService.VerifyEmail(c.Context(), token); err != nil {
		if err == service.ErrInvalidToken || err == service.ErrVerificationTokenExpired {
			return middleware.BadRequest("Invalid or expired verification token")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email verified successfully",
	})
}

func (h *AuthHandler) ResendVerificationEmail(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return middleware.BadRequest("Invalid request body")
	}

	if err := h.authService.ResendVerificationEmail(c.Context(), input.Email); err != nil {
		if strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "connection") {
			return middleware.NewError(fiber.StatusInternalServerError, "Failed to process request")
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If the email exists and is not verified, a verification email has been sent",
	})
}
