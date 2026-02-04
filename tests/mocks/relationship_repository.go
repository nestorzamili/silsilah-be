package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type RelationshipRepository struct {
	mock.Mock
}

func (m *RelationshipRepository) Create(ctx context.Context, rel *domain.Relationship) error {
	args := m.Called(ctx, rel)
	return args.Error(0)
}

func (m *RelationshipRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Relationship), args.Error(1)
}

func (m *RelationshipRepository) Update(ctx context.Context, rel *domain.Relationship) error {
	args := m.Called(ctx, rel)
	return args.Error(0)
}

func (m *RelationshipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *RelationshipRepository) List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error) {
	args := m.Called(ctx, relType)
	return args.Get(0).([]domain.Relationship), args.Error(1)
}

func (m *RelationshipRepository) ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error) {
	args := m.Called(ctx, personID)
	return args.Get(0).([]domain.Relationship), args.Error(1)
}

func (m *RelationshipRepository) GetByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error) {
	args := m.Called(ctx, personID)
	return args.Get(0).([]domain.Relationship), args.Error(1)
}

func (m *RelationshipRepository) GetAll(ctx context.Context) ([]domain.Relationship, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Relationship), args.Error(1)
}

func (m *RelationshipRepository) ListByPeople(ctx context.Context, personIDs []uuid.UUID) ([]domain.Relationship, error) {
	args := m.Called(ctx, personIDs)
	return args.Get(0).([]domain.Relationship), args.Error(1)
}

func (m *RelationshipRepository) CountAll(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RelationshipRepository) GetLastActivityAt(ctx context.Context) (*time.Time, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}
