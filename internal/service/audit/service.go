package audit

import (
	"context"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

type Service interface {
	GetRecentActivities(ctx context.Context, limit int) ([]domain.AuditLog, error)
}

type service struct {
	auditRepo repository.AuditLogRepository
}

func NewService(auditRepo repository.AuditLogRepository) Service {
	return &service{
		auditRepo: auditRepo,
	}
}

func (s *service) GetRecentActivities(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	params := domain.PaginationParams{
		Page:     1,
		PageSize: limit,
	}

	logs, _, err := s.auditRepo.List(ctx, params)
	return logs, err
}
