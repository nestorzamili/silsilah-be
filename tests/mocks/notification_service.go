package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type NotificationService struct {
	mock.Mock
}

func (m *NotificationService) Create(ctx context.Context, notif *domain.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}

func (m *NotificationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Notification), args.Error(1)
}

func (m *NotificationService) List(ctx context.Context, userID uuid.UUID, unreadOnly bool, params domain.PaginationParams) (domain.PaginatedResponse[domain.Notification], error) {
	args := m.Called(ctx, userID, unreadOnly, params)
	return args.Get(0).(domain.PaginatedResponse[domain.Notification]), args.Error(1)
}

func (m *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *NotificationService) NotifyChangeRequest(ctx context.Context, changeRequestID uuid.UUID, requesterID uuid.UUID) error {
	args := m.Called(ctx, changeRequestID, requesterID)
	return args.Error(0)
}

func (m *NotificationService) NotifyChangeApproved(ctx context.Context, changeRequestID uuid.UUID, reviewerID uuid.UUID) error {
	args := m.Called(ctx, changeRequestID, reviewerID)
	return args.Error(0)
}

func (m *NotificationService) NotifyChangeRejected(ctx context.Context, changeRequestID uuid.UUID, reviewerID uuid.UUID) error {
	args := m.Called(ctx, changeRequestID, reviewerID)
	return args.Error(0)
}

func (m *NotificationService) NotifyNewComment(ctx context.Context, commentID uuid.UUID, authorID uuid.UUID) error {
	args := m.Called(ctx, commentID, authorID)
	return args.Error(0)
}

func (m *NotificationService) NotifyPersonAdded(ctx context.Context, personID uuid.UUID, addedBy uuid.UUID) error {
	args := m.Called(ctx, personID, addedBy)
	return args.Error(0)
}

func (m *NotificationService) NotifyRelationshipAdded(ctx context.Context, relationshipID uuid.UUID, addedBy uuid.UUID) error {
	args := m.Called(ctx, relationshipID, addedBy)
	return args.Error(0)
}
