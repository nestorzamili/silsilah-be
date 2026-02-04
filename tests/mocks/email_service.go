package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type EmailService struct {
	mock.Mock
}

func (m *EmailService) SendRegistrationEmail(ctx context.Context, toEmail, fullName string) error {
	args := m.Called(ctx, toEmail, fullName)
	return args.Error(0)
}

func (m *EmailService) SendEmailVerification(ctx context.Context, toEmail, fullName, verificationToken string) error {
	args := m.Called(ctx, toEmail, fullName, verificationToken)
	return args.Error(0)
}

func (m *EmailService) SendPasswordResetEmail(ctx context.Context, toEmail, fullName, resetToken string) error {
	args := m.Called(ctx, toEmail, fullName, resetToken)
	return args.Error(0)
}

func (m *EmailService) SendChangeRequestEmail(ctx context.Context, toEmail, recipientName, requesterName, action, entityType string) error {
	args := m.Called(ctx, toEmail, recipientName, requesterName, action, entityType)
	return args.Error(0)
}

func (m *EmailService) SendChangeStatusEmail(ctx context.Context, toEmail, recipientName, action, entityType, status, reviewerName string) error {
	args := m.Called(ctx, toEmail, recipientName, action, entityType, status, reviewerName)
	return args.Error(0)
}

func (m *EmailService) SendNewCommentEmail(ctx context.Context, toEmail, recipientName, authorName, personName string) error {
	args := m.Called(ctx, toEmail, recipientName, authorName, personName)
	return args.Error(0)
}
