package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type CommentRepository struct {
	mock.Mock
}

func (m *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Comment), args.Error(1)
}

func (m *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *CommentRepository) ListByPerson(ctx context.Context, personID uuid.UUID, params domain.PaginationParams) ([]domain.Comment, int64, error) {
	args := m.Called(ctx, personID, params)
	return args.Get(0).([]domain.Comment), args.Get(1).(int64), args.Error(2)
}
