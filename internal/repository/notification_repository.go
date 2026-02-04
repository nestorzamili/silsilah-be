package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notif *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool, params domain.PaginationParams) ([]domain.Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
}

type notificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, title, message, data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	return r.db.QueryRowxContext(ctx, query,
		notif.ID, notif.UserID, notif.Type, notif.Title, notif.Message, notif.Data,
	).Scan(&notif.CreatedAt)
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	var notif domain.Notification
	query := `SELECT * FROM notifications WHERE notification_id = $1`
	err := r.db.GetContext(ctx, &notif, query, id)
	return &notif, err
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool, params domain.PaginationParams) ([]domain.Notification, int64, error) {
	params.Validate()

	var total int64
	var notifications []domain.Notification

	if unreadOnly {
		countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
		if err := r.db.GetContext(ctx, &total, countQuery, userID); err != nil {
			return nil, 0, err
		}

		query := `
			SELECT * FROM notifications 
			WHERE user_id = $1 AND is_read = false
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		err := r.db.SelectContext(ctx, &notifications, query, userID, params.PageSize, params.Offset())
		return notifications, total, err
	}

	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	if err := r.db.GetContext(ctx, &total, countQuery, userID); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT * FROM notifications 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &notifications, query, userID, params.PageSize, params.Offset())
	return notifications, total, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true, read_at = NOW() WHERE notification_id = $1 AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true, read_at = NOW() WHERE user_id = $1 AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}
