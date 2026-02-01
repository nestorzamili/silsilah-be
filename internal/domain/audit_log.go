package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	UserID     uuid.UUID       `json:"user_id" db:"user_id"`
	UserName   *string         `json:"user_name,omitempty" db:"user_name"`
	Action     string          `json:"action" db:"action"`
	EntityType string          `json:"entity_type" db:"entity_type"`
	EntityID   uuid.UUID       `json:"entity_id" db:"entity_id"`
	OldValue   json.RawMessage `json:"old_value,omitempty" db:"old_value"`
	NewValue   json.RawMessage `json:"new_value,omitempty" db:"new_value"`
	IPAddress  *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  *string         `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

type CreateAuditLogInput struct {
	UserID     uuid.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	OldValue   interface{}
	NewValue   interface{}
	IPAddress  *string
	UserAgent  *string
}
