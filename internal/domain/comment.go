package domain

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID  `json:"id" db:"comment_id"`
	PersonID  uuid.UUID  `json:"person_id" db:"person_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	ParentID  *uuid.UUID `json:"parent_id" db:"parent_id"`
	Content   string     `json:"content" db:"content"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`

	User *CommentUser `json:"user,omitempty"`
	Replies []Comment `json:"replies,omitempty"`
}

type CommentUser struct {
	ID        uuid.UUID `json:"id" db:"user_id"`
	FullName  string    `json:"full_name" db:"user_full_name"`
	AvatarURL *string   `json:"avatar_url" db:"user_avatar_url"`
}

type CreateCommentInput struct {
	ParentID *uuid.UUID `json:"parent_id"`
	Content  string     `json:"content" validate:"required,min=1,max=2000"`
}

type UpdateCommentInput struct {
	Content string `json:"content" validate:"required,min=1,max=2000"`
}
