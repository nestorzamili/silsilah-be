package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type ChangeRequestRepository interface {
	Create(ctx context.Context, req *domain.ChangeRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error)
	List(ctx context.Context, status *domain.ChangeRequestStatus, params domain.PaginationParams) ([]domain.ChangeRequest, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ChangeRequestStatus, reviewedBy uuid.UUID, note *string) error
}

type changeRequestRepository struct {
	db *sqlx.DB
}

func NewChangeRequestRepository(db *sqlx.DB) ChangeRequestRepository {
	return &changeRequestRepository{db: db}
}

func (r *changeRequestRepository) Create(ctx context.Context, req *domain.ChangeRequest) error {
	query := `
		INSERT INTO change_requests (id, requested_by, entity_type, entity_id, action, payload, requester_note, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		req.ID, req.RequestedBy, req.EntityType, req.EntityID,
		req.Action, req.Payload, req.RequesterNote, req.Status,
	).Scan(&req.CreatedAt, &req.UpdatedAt)
}

func (r *changeRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error) {
	var req domain.ChangeRequest
	query := `SELECT * FROM change_requests WHERE id = $1`
	err := r.db.GetContext(ctx, &req, query, id)
	return &req, err
}

func (r *changeRequestRepository) List(ctx context.Context, status *domain.ChangeRequestStatus, params domain.PaginationParams) ([]domain.ChangeRequest, int64, error) {
	params.Validate()

	var total int64
	var requests []domain.ChangeRequest

	if status != nil {
		countQuery := `SELECT COUNT(*) FROM change_requests WHERE status = $1`
		if err := r.db.GetContext(ctx, &total, countQuery, *status); err != nil {
			return nil, 0, err
		}

		query := `
			SELECT * FROM change_requests 
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		err := r.db.SelectContext(ctx, &requests, query, *status, params.PageSize, params.Offset())
		return requests, total, err
	}

	countQuery := `SELECT COUNT(*) FROM change_requests`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT * FROM change_requests 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &requests, query, params.PageSize, params.Offset())
	return requests, total, err
}

func (r *changeRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ChangeRequestStatus, reviewedBy uuid.UUID, note *string) error {
	query := `
		UPDATE change_requests 
		SET status = $2, reviewed_by = $3, reviewed_at = NOW(), review_note = $4, updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status, reviewedBy, note)
	return err
}
