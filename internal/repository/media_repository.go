package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type MediaRepository interface {
	Create(ctx context.Context, media *domain.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error)
	Update(ctx context.Context, media *domain.Media) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, personID *uuid.UUID, params domain.PaginationParams) ([]domain.Media, int64, error)
}

type mediaRepository struct {
	db *sqlx.DB
}

func NewMediaRepository(db *sqlx.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *domain.Media) error {
	query := `
		INSERT INTO media (media_id, person_id, uploaded_by, file_name, file_size, mime_type, storage_path, caption, status, taken_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at`

	return r.db.QueryRowxContext(ctx, query,
		media.ID, media.PersonID, media.UploadedBy,
		media.FileName, media.FileSize, media.MimeType, media.StoragePath,
		media.Caption, media.Status, media.TakenAt,
	).Scan(&media.CreatedAt)
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	var media domain.Media
	query := `SELECT * FROM media WHERE media_id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &media, query, id)
	return &media, err
}

func (r *mediaRepository) Update(ctx context.Context, media *domain.Media) error {
	query := `
		UPDATE media 
		SET status = $1, caption = $2, person_id = $3
		WHERE media_id = $4 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, media.Status, media.Caption, media.PersonID, media.ID)
	return err
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE media SET deleted_at = NOW() WHERE media_id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *mediaRepository) List(ctx context.Context, personID *uuid.UUID, params domain.PaginationParams) ([]domain.Media, int64, error) {
	params.Validate()

	var total int64
	var mediaList []domain.Media

	if personID != nil {
		countQuery := `SELECT COUNT(*) FROM media WHERE person_id = $1 AND status = 'active' AND deleted_at IS NULL`
		if err := r.db.GetContext(ctx, &total, countQuery, *personID); err != nil {
			return nil, 0, err
		}

		query := `
			SELECT * FROM media 
			WHERE person_id = $1 AND status = 'active' AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		err := r.db.SelectContext(ctx, &mediaList, query, *personID, params.PageSize, params.Offset())
		return mediaList, total, err
	}

	countQuery := `SELECT COUNT(*) FROM media WHERE status = 'active' AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT * FROM media 
		WHERE status = 'active' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &mediaList, query, params.PageSize, params.Offset())
	return mediaList, total, err
}
