package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/resend/resend-go/v3"

	"silsilah-keluarga/internal/config"
)

type Service interface {
	SendRegistrationEmail(ctx context.Context, toEmail, fullName string) error
	SendEmailVerification(ctx context.Context, toEmail, fullName, verificationToken string) error
	SendPasswordResetEmail(ctx context.Context, toEmail, fullName, resetToken string) error
	SendChangeRequestEmail(ctx context.Context, toEmail, recipientName, requesterName, action, entityType string) error
	SendChangeStatusEmail(ctx context.Context, toEmail, recipientName, action, entityType, status, reviewerName string) error
	SendNewCommentEmail(ctx context.Context, toEmail, recipientName, authorName, personName string) error
}

type service struct {
	client       *resend.Client
	config       *config.Config
	templatePath string
}

func NewService(cfg *config.Config) Service {
	client := resend.NewClient(cfg.ResendAPIKey)
	templatePath := "internal/service/templates/email"
	return &service{
		client:       client,
		config:       cfg,
		templatePath: templatePath,
	}
}

func (s *service) sendEmail(toEmail, subject, templateName string, data interface{}) error {
	tmpl, err := template.ParseFiles(
		filepath.Join(s.templatePath, "layout.html"),
		filepath.Join(s.templatePath, templateName),
	)
	if err != nil {
		return fmt.Errorf("failed to parse email templates: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Silsilah Keluarga <%s>", s.config.FromEmail),
		To:      []string{toEmail},
		Html:    body.String(),
		Subject: subject,
	}

	_, err = s.client.Emails.Send(params)
	return err
}

func (s *service) SendRegistrationEmail(ctx context.Context, toEmail, fullName string) error {
	data := struct {
		Title string
		Name  string
		Link  string
	}{
		Title: "Selamat Datang di Silsilah Keluarga",
		Name:  fullName,
		Link:  fmt.Sprintf("http://%s/login", s.config.Domain),
	}
	return s.sendEmail(toEmail, "Selamat Datang di Silsilah Keluarga!", "registration.html", data)
}

func (s *service) SendEmailVerification(ctx context.Context, toEmail, fullName, verificationToken string) error {
	data := struct {
		Title string
		Name  string
		Link  string
	}{
		Title: "Verifikasi Email - Silsilah Keluarga",
		Name:  fullName,
		Link:  fmt.Sprintf("https://%s/verify-email?token=%s", s.config.Domain, verificationToken),
	}
	return s.sendEmail(toEmail, "Verifikasi Email - Silsilah Keluarga", "verification.html", data)
}

func (s *service) SendPasswordResetEmail(ctx context.Context, toEmail, fullName, resetToken string) error {
	data := struct {
		Title string
		Name  string
		Link  string
	}{
		Title: "Reset Kata Sandi - Silsilah Keluarga",
		Name:  fullName,
		Link:  fmt.Sprintf("https://%s/reset-password?token=%s", s.config.Domain, resetToken),
	}
	return s.sendEmail(toEmail, "Permintaan Reset Kata Sandi - Silsilah Keluarga", "reset_password.html", data)
}

func (s *service) SendChangeRequestEmail(ctx context.Context, toEmail, recipientName, requesterName, action, entityType string) error {
	data := struct {
		Title         string
		Name          string
		RequesterName string
		Action        string
		EntityType    string
	}{
		Title:         "Permintaan Perubahan Baru",
		Name:          recipientName,
		RequesterName: requesterName,
		Action:        action,
		EntityType:    entityType,
	}
	return s.sendEmail(toEmail, "Permintaan Perubahan Baru - Silsilah Keluarga", "change_request.html", data)
}

func (s *service) SendChangeStatusEmail(ctx context.Context, toEmail, recipientName, action, entityType, status, reviewerName string) error {
	color := "#10b981" 
	if status == "Ditolak" || status == "REJECTED" {
		color = "#ef4444" 
	}

	data := struct {
		Title        string
		Name         string
		Action       string
		EntityType   string
		Status       string
		ReviewerName string
		Color        string
	}{
		Title:        fmt.Sprintf("Permintaan %s", status),
		Name:         recipientName,
		Action:       action,
		EntityType:   entityType,
		Status:       status,
		ReviewerName: reviewerName,
		Color:        color,
	}
	return s.sendEmail(toEmail, fmt.Sprintf("Permintaan %s - Silsilah Keluarga", status), "change_status.html", data)
}

func (s *service) SendNewCommentEmail(ctx context.Context, toEmail, recipientName, authorName, personName string) error {
	data := struct {
		Title      string
		Name       string
		AuthorName string
		PersonName string
	}{
		Title:      "Komentar Baru",
		Name:       recipientName,
		AuthorName: authorName,
		PersonName: personName,
	}
	return s.sendEmail(toEmail, fmt.Sprintf("Komentar Baru di Profil %s - Silsilah Keluarga", personName), "new_comment.html", data)
}
