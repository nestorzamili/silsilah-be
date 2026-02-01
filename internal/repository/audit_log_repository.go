package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	List(ctx context.Context, params domain.PaginationParams) ([]domain.AuditLog, int64, error)
	ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, params domain.PaginationParams) ([]domain.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *sqlx.DB
}

func NewAuditLogRepository(db *sqlx.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, old_value, new_value, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at`

	return r.db.QueryRowxContext(ctx, query,
		log.ID, log.UserID, log.Action, log.EntityType, log.EntityID,
		log.OldValue, log.NewValue, log.IPAddress, log.UserAgent,
	).Scan(&log.CreatedAt)
}

func (r *auditLogRepository) List(ctx context.Context, params domain.PaginationParams) ([]domain.AuditLog, int64, error) {
	params.Validate()

	var total int64
	countQuery := `SELECT COUNT(*) FROM audit_logs`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			al.*,
			u.full_name as user_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		ORDER BY al.created_at DESC
		LIMIT $1 OFFSET $2`

	var logs []domain.AuditLog
	err := r.db.SelectContext(ctx, &logs, query, params.PageSize, params.Offset())
	return logs, total, err
}

func (r *auditLogRepository) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, params domain.PaginationParams) ([]domain.AuditLog, int64, error) {
	params.Validate()

	var total int64
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE entity_type = $1 AND entity_id = $2`
	if err := r.db.GetContext(ctx, &total, countQuery, entityType, entityID); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT * FROM audit_logs 
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	var logs []domain.AuditLog
	err := r.db.SelectContext(ctx, &logs, query, entityType, entityID, params.PageSize, params.Offset())
	return logs, total, err
}

func CreateAuditLog(repo AuditLogRepository, ctx context.Context, input domain.CreateAuditLogInput) error {
	oldValueJSON, _ := json.Marshal(input.OldValue)
	newValueJSON, _ := json.Marshal(input.NewValue)

	log := &domain.AuditLog{
		ID:         uuid.New(),
		UserID:     input.UserID,
		Action:     input.Action,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		OldValue:   oldValueJSON,
		NewValue:   newValueJSON,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
	}

	return repo.Create(ctx, log)
}
