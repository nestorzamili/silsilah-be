package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	Update(ctx context.Context, event *domain.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Event, error)
	ListByRelationship(ctx context.Context, relationshipID uuid.UUID) ([]domain.Event, error)
}

type eventRepository struct {
	db *sqlx.DB
}

func NewEventRepository(db *sqlx.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(ctx context.Context, event *domain.Event) error {
	query := `
		INSERT INTO events (event_id, person_id, relationship_id, type, title, date, place, description, metadata, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		event.ID, event.PersonID, event.RelationshipID, event.Type, event.Title, event.Date, event.Place, event.Description, event.Metadata, event.CreatedBy,
	).Scan(&event.CreatedAt, &event.UpdatedAt)
}

func (r *eventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	var event domain.Event
	query := `SELECT * FROM events WHERE event_id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &event, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) Update(ctx context.Context, event *domain.Event) error {
	query := `
		UPDATE events 
		SET type = $2, title = $3, date = $4, place = $5, description = $6, metadata = $7, updated_at = NOW()
		WHERE event_id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	return r.db.QueryRowxContext(ctx, query,
		event.ID, event.Type, event.Title, event.Date, event.Place, event.Description, event.Metadata,
	).Scan(&event.UpdatedAt)
}

func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE events SET deleted_at = NOW() WHERE event_id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *eventRepository) ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Event, error) {
	query := `SELECT * FROM events WHERE person_id = $1 AND deleted_at IS NULL ORDER BY date ASC`
	var events []domain.Event
	err := r.db.SelectContext(ctx, &events, query, personID)
	return events, err
}

func (r *eventRepository) ListByRelationship(ctx context.Context, relationshipID uuid.UUID) ([]domain.Event, error) {
	query := `SELECT * FROM events WHERE relationship_id = $1 AND deleted_at IS NULL ORDER BY date ASC`
	var events []domain.Event
	err := r.db.SelectContext(ctx, &events, query, relationshipID)
	return events, err
}
