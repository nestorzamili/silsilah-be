package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type PersonRepository interface {
	Create(ctx context.Context, person *domain.Person) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error)
	Update(ctx context.Context, person *domain.Person) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params domain.PaginationParams) ([]domain.Person, int64, error)
	Search(ctx context.Context, query string, limit int) ([]domain.Person, error)
	GetAll(ctx context.Context) ([]domain.Person, error)
}

type personRepository struct {
	db *sqlx.DB
}

func NewPersonRepository(db *sqlx.DB) PersonRepository {
	return &personRepository{db: db}
}

func (r *personRepository) Create(ctx context.Context, person *domain.Person) error {
	query := `
		INSERT INTO persons (id, first_name, last_name, nickname, gender, 
			birth_date, birth_place, death_date, death_place, bio, avatar_url, 
			occupation, religion, nationality, education, phone, email, address,
			is_alive, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		person.ID, person.FirstName, person.LastName, person.Nickname,
		person.Gender, person.BirthDate, person.BirthPlace, person.DeathDate,
		person.DeathPlace, person.Bio, person.AvatarURL,
		person.Occupation, person.Religion, person.Nationality, person.Education,
		person.Phone, person.Email, person.Address,
		person.IsAlive, person.CreatedBy,
	).Scan(&person.CreatedAt, &person.UpdatedAt)
}

func (r *personRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	var person domain.Person
	query := `SELECT * FROM persons WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &person, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *personRepository) Update(ctx context.Context, person *domain.Person) error {
	query := `
		UPDATE persons 
		SET first_name = $2, last_name = $3, nickname = $4, gender = $5,
			birth_date = $6, birth_place = $7, death_date = $8, death_place = $9,
			bio = $10, avatar_url = $11, occupation = $12, religion = $13,
			nationality = $14, education = $15, phone = $16, email = $17,
			address = $18, is_alive = $19, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	return r.db.QueryRowxContext(ctx, query,
		person.ID, person.FirstName, person.LastName, person.Nickname, person.Gender,
		person.BirthDate, person.BirthPlace, person.DeathDate, person.DeathPlace,
		person.Bio, person.AvatarURL, person.Occupation, person.Religion,
		person.Nationality, person.Education, person.Phone, person.Email,
		person.Address, person.IsAlive,
	).Scan(&person.UpdatedAt)
}

func (r *personRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE persons SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *personRepository) List(ctx context.Context, params domain.PaginationParams) ([]domain.Person, int64, error) {
	params.Validate()

	var total int64
	countQuery := `SELECT COUNT(*) FROM persons WHERE deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT * FROM persons 
		WHERE deleted_at IS NULL
		ORDER BY first_name, last_name
		LIMIT $1 OFFSET $2`

	var persons []domain.Person
	err := r.db.SelectContext(ctx, &persons, query, params.PageSize, params.Offset())
	return persons, total, err
}

func (r *personRepository) Search(ctx context.Context, query string, limit int) ([]domain.Person, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	sqlQuery := `
		SELECT * FROM persons 
		WHERE deleted_at IS NULL
			AND (
				(first_name || ' ' || COALESCE(last_name, '')) ILIKE '%' || $1 || '%'
				OR nickname ILIKE '%' || $1 || '%'
			)
		ORDER BY first_name, last_name
		LIMIT $2`

	var persons []domain.Person
	err := r.db.SelectContext(ctx, &persons, sqlQuery, query, limit)
	return persons, err
}

func (r *personRepository) GetAll(ctx context.Context) ([]domain.Person, error) {
	query := `SELECT * FROM persons WHERE deleted_at IS NULL`

	var persons []domain.Person
	err := r.db.SelectContext(ctx, &persons, query)
	return persons, err
}
