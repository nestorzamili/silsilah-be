package export

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service/graph"
)

type Service interface {
	ExportJSON(ctx context.Context, userID, personID uuid.UUID) (any, error)
	ExportGEDCOM(ctx context.Context, userID, rootID uuid.UUID) (string, error)
}

type service struct {
	personRepo repository.PersonRepository
	relRepo    repository.RelationshipRepository
	auditRepo  repository.AuditLogRepository
	graphSvc   graph.Service
}

func NewService(personRepo repository.PersonRepository, relRepo repository.RelationshipRepository, auditRepo repository.AuditLogRepository, graphSvc graph.Service) Service {
	return &service{
		personRepo: personRepo,
		relRepo:    relRepo,
		auditRepo:  auditRepo,
		graphSvc:   graphSvc,
	}
}

func (s *service) ExportJSON(ctx context.Context, userID, personID uuid.UUID) (any, error) {
	graphData, err := s.graphSvc.GetFullGraph(ctx)
	if err != nil {
		return nil, err
	}
	return graphData, nil
}

func (s *service) ExportGEDCOM(ctx context.Context, userID, rootID uuid.UUID) (string, error) {
	return "", errors.New("GEDCOM export not implemented yet")
}
