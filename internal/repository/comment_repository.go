package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	Update(ctx context.Context, comment *domain.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPerson(ctx context.Context, personID uuid.UUID, params domain.PaginationParams) ([]domain.Comment, int64, error)
}

type commentRepository struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	query := `
		INSERT INTO comments (comment_id, person_id, user_id, parent_id, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		comment.ID, comment.PersonID, comment.UserID, comment.ParentID, comment.Content,
	).Scan(&comment.CreatedAt, &comment.UpdatedAt)
}

func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	var comment domain.Comment
	query := `SELECT * FROM comments WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &comment, query, id)
	return &comment, err
}

func (r *commentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	query := `
		UPDATE comments 
		SET content = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	return r.db.QueryRowxContext(ctx, query,
		comment.ID, comment.Content,
	).Scan(&comment.UpdatedAt)
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE comments SET deleted_at = NOW() WHERE comment_id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *commentRepository) ListByPerson(ctx context.Context, personID uuid.UUID, params domain.PaginationParams) ([]domain.Comment, int64, error) {
	params.Validate()

	var total int64
	countQuery := `SELECT COUNT(*) FROM comments WHERE person_id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &total, countQuery, personID); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			c.comment_id as id, c.person_id, c.user_id, c.parent_id, c.content, c.created_at, c.updated_at,
			u.user_id as user_id, u.full_name as user_full_name, u.avatar_url as user_avatar_url
		FROM comments c
		INNER JOIN users u ON c.user_id = u.user_id
		WHERE c.person_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryxContext(ctx, query, personID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		var user domain.CommentUser
		err := rows.Scan(
			&c.ID, &c.PersonID, &c.UserID, &c.ParentID, &c.Content, &c.CreatedAt, &c.UpdatedAt,
			&user.ID, &user.FullName, &user.AvatarURL,
		)
		if err != nil {
			return nil, 0, err
		}
		c.User = &user
		comments = append(comments, c)
	}

	return comments, total, rows.Err()
}
