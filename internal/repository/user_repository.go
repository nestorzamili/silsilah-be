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

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	AssignRole(ctx context.Context, userID uuid.UUID, role string) error
	ListByRole(ctx context.Context, role string) ([]domain.User, error)
	GetByRoles(ctx context.Context, roles []domain.UserRole) ([]domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	SetPasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetUserByResetToken(ctx context.Context, token string) (*domain.User, error)
	ClearPasswordResetToken(ctx context.Context, userID uuid.UUID) error
	SetEmailVerificationToken(ctx context.Context, userID uuid.UUID, token string, sentAt time.Time) error
	GetUserByEmailVerificationToken(ctx context.Context, token string) (*domain.User, error)
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (user_id, email, password_hash, full_name, avatar_url, bio, role, is_active, is_email_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FullName,
		user.AvatarURL, user.Bio, user.Role, user.IsActive, user.IsEmailVerified,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE user_id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET email = :email, password_hash = :password_hash, full_name = :full_name, 
			avatar_url = :avatar_url, bio = :bio, role = :role, 
			linked_person_id = :linked_person_id, updated_at = NOW()
		WHERE user_id = :user_id AND deleted_at IS NULL`

	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE user_id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	err := r.db.GetContext(ctx, &exists, query, email)
	return exists, err
}

func (r *userRepository) AssignRole(ctx context.Context, userID uuid.UUID, role string) error {
	query := `
		UPDATE users 
		SET role = $2, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	var updatedAt time.Time
	err := r.db.QueryRowxContext(ctx, query, userID, role).Scan(&updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("user not found")
	}
	return err
}

func (r *userRepository) ListByRole(ctx context.Context, role string) ([]domain.User, error) {
	var users []domain.User
	query := `SELECT * FROM users WHERE role = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &users, query, role)
	return users, err
}

func (r *userRepository) GetByRoles(ctx context.Context, roles []domain.UserRole) ([]domain.User, error) {
	if len(roles) == 0 {
		return []domain.User{}, nil
	}

	var users []domain.User
	query := `SELECT * FROM users WHERE role = ANY($1) AND deleted_at IS NULL ORDER BY created_at DESC`

	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = string(role)
	}

	err := r.db.SelectContext(ctx, &users, query, roleStrings)
	return users, err
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	query := `SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &users, query)
	return users, err
}

func (r *userRepository) SetPasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		UPDATE users 
		SET password_reset_token = $2, password_reset_expires_at = $3, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	var updatedAt time.Time
	err := r.db.QueryRowxContext(ctx, query, userID, token, expiresAt).Scan(&updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("user not found")
	}
	return err
}

func (r *userRepository) GetUserByResetToken(ctx context.Context, token string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE password_reset_token = $1 AND password_reset_expires_at > NOW() AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &user, query, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ClearPasswordResetToken(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET password_reset_token = NULL, password_reset_expires_at = NULL, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	var updatedAt time.Time
	err := r.db.QueryRowxContext(ctx, query, userID).Scan(&updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("user not found")
	}
	return err
}

func (r *userRepository) SetEmailVerificationToken(ctx context.Context, userID uuid.UUID, token string, sentAt time.Time) error {
	query := `
		UPDATE users 
		SET email_verification_token = $2, email_verification_sent_at = $3, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	var updatedAt time.Time
	err := r.db.QueryRowxContext(ctx, query, userID, token, sentAt).Scan(&updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("user not found")
	}
	return err
}

func (r *userRepository) GetUserByEmailVerificationToken(ctx context.Context, token string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE email_verification_token = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &user, query, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET is_email_verified = TRUE, email_verification_token = NULL, email_verification_sent_at = NULL, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	var updatedAt time.Time
	err := r.db.QueryRowxContext(ctx, query, userID).Scan(&updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("user not found")
	}
	return err
}
