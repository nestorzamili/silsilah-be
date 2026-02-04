package repository

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Session struct {
	ID        uuid.UUID  `db:"session_id"`
	UserID    uuid.UUID  `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	UserAgent *string    `db:"user_agent"`
	IPAddress *net.IP    `db:"ip_address"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (session_id, user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	return r.db.QueryRowxContext(ctx, query,
		session.ID, session.UserID, session.TokenHash, session.UserAgent, session.IPAddress, session.ExpiresAt,
	).Scan(&session.CreatedAt)
}

func (r *sessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	var session Session
	query := `SELECT * FROM sessions WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`

	err := r.db.GetContext(ctx, &session, query, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error) {
	var sessions []*Session
	query := `SELECT * FROM sessions WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW() ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET revoked_at = NOW() WHERE session_id = $1 AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW() OR revoked_at IS NOT NULL`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
