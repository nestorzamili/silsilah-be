package service

import (
	"context"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

type AuditService interface {
	GetRecentActivities(ctx context.Context, limit int) ([]domain.AuditLog, error)
}

type auditService struct {
	auditRepo repository.AuditLogRepository
}

func NewAuditService(auditRepo repository.AuditLogRepository) AuditService {
	return &auditService{
		auditRepo: auditRepo,
	}
}

func (s *auditService) GetRecentActivities(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	params := domain.PaginationParams{
		Page:     1,
		PageSize: limit,
	}
	
	logs, _, err := s.auditRepo.List(ctx, params)
	return logs, err
}
