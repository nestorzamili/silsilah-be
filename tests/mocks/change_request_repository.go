package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type ChangeRequestRepository struct {
	mock.Mock
}

func (m *ChangeRequestRepository) Create(ctx context.Context, req *domain.ChangeRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *ChangeRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChangeRequest), args.Error(1)
}

func (m *ChangeRequestRepository) List(ctx context.Context, status *domain.ChangeRequestStatus, params domain.PaginationParams) ([]domain.ChangeRequest, int64, error) {
	args := m.Called(ctx, status, params)
	return args.Get(0).([]domain.ChangeRequest), args.Get(1).(int64), args.Error(2)
}

func (m *ChangeRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ChangeRequestStatus, reviewedBy uuid.UUID, note *string) error {
	args := m.Called(ctx, id, status, reviewedBy, note)
	return args.Error(0)
}

func (m *ChangeRequestRepository) CountPending(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
