package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventTypeBirth    EventType = "BIRTH"
	EventTypeDeath    EventType = "DEATH"
	EventTypeMarriage EventType = "MARRIAGE"
	EventTypeDivorce  EventType = "DIVORCE"
	EventTypeOther    EventType = "OTHER"
)

func (e EventType) IsValid() bool {
	switch e {
	case EventTypeBirth, EventTypeDeath, EventTypeMarriage, EventTypeDivorce, EventTypeOther:
		return true
	}
	return false
}

type Event struct {
	ID             uuid.UUID       `json:"id" db:"event_id"`
	PersonID       *uuid.UUID      `json:"person_id" db:"person_id"`
	RelationshipID *uuid.UUID      `json:"relationship_id,omitempty" db:"relationship_id"`
	Type           EventType       `json:"type" db:"type"`
	Title          string          `json:"title" db:"title"`
	Date           *time.Time      `json:"date,omitempty" db:"date"`
	Place          *string         `json:"place,omitempty" db:"place"`
	Description    *string         `json:"description,omitempty" db:"description"`
	Metadata       json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedBy      uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time      `json:"-" db:"deleted_at"`
}

type CreateEventInput struct {
	PersonID       *uuid.UUID      `json:"person_id" validate:"required_without=RelationshipID"`
	RelationshipID *uuid.UUID      `json:"relationship_id" validate:"required_without=PersonID"`
	Type           EventType       `json:"type" validate:"required"`
	Title          string          `json:"title" validate:"required,max=100"`
	Date           *time.Time      `json:"date"`
	Place          *string         `json:"place" validate:"omitempty,max=200"`
	Description    *string         `json:"description"`
	Metadata       json.RawMessage `json:"metadata"`
}

type UpdateEventInput struct {
	Type        *EventType      `json:"type" validate:"omitempty"`
	Title       *string         `json:"title" validate:"omitempty,max=100"`
	Date        NullableTime    `json:"date"`
	Place       NullableString  `json:"place" validate:"omitempty,max=200"`
	Description NullableString  `json:"description"`
	Metadata    json.RawMessage `json:"metadata"`
}
