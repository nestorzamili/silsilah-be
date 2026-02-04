package mocks

import (
	"context"
	"silsilah-keluarga/internal/domain"

	"github.com/stretchr/testify/mock"
)

type NarrativeService struct {
	mock.Mock
}

func (m *NarrativeService) DescribeRelationship(ctx context.Context, path *domain.RelationshipPath, locale string) string {
	args := m.Called(ctx, path, locale)
	return args.String(0)
}
