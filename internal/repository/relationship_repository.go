package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type RelationshipRepository interface {
	Create(ctx context.Context, rel *domain.Relationship) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error)
	Update(ctx context.Context, rel *domain.Relationship) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error)
	ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error)
	GetByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error)
	GetAll(ctx context.Context) ([]domain.Relationship, error)
	ListByPeople(ctx context.Context, personIDs []uuid.UUID) ([]domain.Relationship, error)
	CountAll(ctx context.Context) (int64, error)
	GetLastActivityAt(ctx context.Context) (*time.Time, error)
}

type relationshipRepository struct {
	db *sqlx.DB
}

func NewRelationshipRepository(db *sqlx.DB) RelationshipRepository {
	return &relationshipRepository{db: db}
}

func (r *relationshipRepository) Create(ctx context.Context, rel *domain.Relationship) error {
	query := `
		INSERT INTO relationships (relationship_id, person_a, person_b, type, metadata, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		rel.ID, rel.PersonA, rel.PersonB, rel.Type, rel.Metadata, rel.CreatedBy,
	).Scan(&rel.CreatedAt, &rel.UpdatedAt)
}

func (r *relationshipRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error) {
	var rel domain.Relationship
	query := `
		SELECT relationship_id, person_a, person_b, type, metadata, created_by, created_at, updated_at, deleted_at 
		FROM relationships 
		WHERE relationship_id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &rel, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *relationshipRepository) Update(ctx context.Context, rel *domain.Relationship) error {
	query := `
		UPDATE relationships 
		SET metadata = $2, updated_at = NOW()
		WHERE relationship_id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	return r.db.QueryRowxContext(ctx, query,
		rel.ID, rel.Metadata,
	).Scan(&rel.UpdatedAt)
}

func (r *relationshipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE relationships SET deleted_at = NOW() WHERE relationship_id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *relationshipRepository) List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error) {
	var relationships []domain.Relationship
	var err error

	query := `
		SELECT relationship_id, person_a, person_b, type, metadata, created_by, created_at, updated_at, deleted_at 
		FROM relationships 
		WHERE deleted_at IS NULL`

	if relType != nil {
		query += ` AND type = $1`
		err = r.db.SelectContext(ctx, &relationships, query, *relType)
	} else {
		err = r.db.SelectContext(ctx, &relationships, query)
	}

	return relationships, err
}

func (r *relationshipRepository) ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error) {
	query := `
		SELECT relationship_id, person_a, person_b, type, metadata, created_by, created_at, updated_at, deleted_at 
		FROM relationships 
		WHERE (person_a = $1 OR person_b = $1) AND deleted_at IS NULL`

	var relationships []domain.Relationship
	err := r.db.SelectContext(ctx, &relationships, query, personID)
	return relationships, err
}

func (r *relationshipRepository) GetByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error) {
	return r.ListByPerson(ctx, personID)
}

func (r *relationshipRepository) GetAll(ctx context.Context) ([]domain.Relationship, error) {
	query := `
		SELECT relationship_id, person_a, person_b, type, metadata, created_by, created_at, updated_at, deleted_at 
		FROM relationships 
		WHERE deleted_at IS NULL`

	var relationships []domain.Relationship
	err := r.db.SelectContext(ctx, &relationships, query)
	return relationships, err
}

func (r *relationshipRepository) ListByPeople(ctx context.Context, personIDs []uuid.UUID) ([]domain.Relationship, error) {
	if len(personIDs) == 0 {
		return []domain.Relationship{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT relationship_id, person_a, person_b, type, metadata, created_by, created_at, updated_at, deleted_at 
		FROM relationships 
		WHERE (person_a IN (?) OR person_b IN (?))
		AND deleted_at IS NULL`, personIDs, personIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var rels []domain.Relationship
	err = r.db.SelectContext(ctx, &rels, query, args...)
	return rels, err
}

func (r *relationshipRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM relationships WHERE deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *relationshipRepository) GetLastActivityAt(ctx context.Context) (*time.Time, error) {
	var t *time.Time
	query := `SELECT MAX(updated_at) FROM relationships`
	err := r.db.GetContext(ctx, &t, query)
	return t, err
}
