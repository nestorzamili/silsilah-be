package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	UserID    uuid.UUID        `json:"user_id" db:"user_id"`
	Type      NotificationType `json:"type" db:"type"`
	Title     string           `json:"title" db:"title"`
	Message   string           `json:"message" db:"message"`
	Data      json.RawMessage  `json:"data,omitempty" db:"data"`
	IsRead    bool             `json:"is_read" db:"is_read"`
	ReadAt    *time.Time       `json:"read_at,omitempty" db:"read_at"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}

type NotificationType string

const (
	NotifChangeRequest     NotificationType = "CHANGE_REQUEST"
	NotifChangeApproved    NotificationType = "CHANGE_APPROVED"
	NotifChangeRejected    NotificationType = "CHANGE_REJECTED"
	NotifNewComment        NotificationType = "NEW_COMMENT"
	NotifPersonAdded       NotificationType = "PERSON_ADDED"
	NotifRelationshipAdded NotificationType = "RELATIONSHIP_ADDED"
)
