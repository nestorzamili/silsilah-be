package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MediaRepository struct {
	mock.Mock
}

func (m *MediaRepository) Create(ctx context.Context, media *domain.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Media), args.Error(1)
}

func (m *MediaRepository) Update(ctx context.Context, media *domain.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MediaRepository) List(ctx context.Context, personID *uuid.UUID, params domain.PaginationParams) ([]domain.Media, int64, error) {
	args := m.Called(ctx, personID, params)
	return args.Get(0).([]domain.Media), args.Get(1).(int64), args.Error(2)
}
