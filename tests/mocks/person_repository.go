package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type PersonRepository struct {
	mock.Mock
}

func (m *PersonRepository) Create(ctx context.Context, person *domain.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *PersonRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Person), args.Error(1)
}

func (m *PersonRepository) Update(ctx context.Context, person *domain.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *PersonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *PersonRepository) List(ctx context.Context, params domain.PaginationParams) ([]domain.Person, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]domain.Person), args.Get(1).(int64), args.Error(2)
}

func (m *PersonRepository) Search(ctx context.Context, query string, limit int) ([]domain.Person, error) {
	args := m.Called(ctx, query, limit)
	return args.Get(0).([]domain.Person), args.Error(1)
}

func (m *PersonRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Person, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]domain.Person), args.Error(1)
}

func (m *PersonRepository) GetAll(ctx context.Context) ([]domain.Person, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Person), args.Error(1)
}

func (m *PersonRepository) CountAll(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *PersonRepository) CountLiving(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *PersonRepository) CountOrphans(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *PersonRepository) GetLastActivityAt(ctx context.Context) (*time.Time, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}
