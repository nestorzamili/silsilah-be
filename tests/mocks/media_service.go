package mocks

import (
	"context"
	"io"
	"silsilah-keluarga/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MediaService struct {
	mock.Mock
}

func (m *MediaService) Upload(ctx context.Context, userID uuid.UUID, personID *uuid.UUID, caption *string, fileName string, fileSize int64, mimeType string, reader io.Reader, status string) (*domain.Media, error) {
	args := m.Called(ctx, userID, personID, caption, fileName, fileSize, mimeType, reader, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Media), args.Error(1)
}

func (m *MediaService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Media), args.Error(1)
}

func (m *MediaService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MediaService) List(ctx context.Context, personID *uuid.UUID, params domain.PaginationParams) (domain.PaginatedResponse[domain.Media], error) {
	args := m.Called(ctx, personID, params)
	return args.Get(0).(domain.PaginatedResponse[domain.Media]), args.Error(1)
}

func (m *MediaService) Approve(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
