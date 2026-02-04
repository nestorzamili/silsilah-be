package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type NotificationRepository struct {
	mock.Mock
}

func (m *NotificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}

func (m *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Notification), args.Error(1)
}

func (m *NotificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool, params domain.PaginationParams) ([]domain.Notification, int64, error) {
	args := m.Called(ctx, userID, unreadOnly, params)
	return args.Get(0).([]domain.Notification), args.Get(1).(int64), args.Error(2)
}

func (m *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
