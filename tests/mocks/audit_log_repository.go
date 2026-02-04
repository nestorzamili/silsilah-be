package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type AuditLogRepository struct {
	mock.Mock
}

func (m *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *AuditLogRepository) List(ctx context.Context, params domain.PaginationParams) ([]domain.AuditLog, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]domain.AuditLog), args.Get(1).(int64), args.Error(2)
}

func (m *AuditLogRepository) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, params domain.PaginationParams) ([]domain.AuditLog, int64, error) {
	args := m.Called(ctx, entityType, entityID, params)
	return args.Get(0).([]domain.AuditLog), args.Get(1).(int64), args.Error(2)
}
