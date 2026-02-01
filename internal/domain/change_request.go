package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ChangeRequest struct {
	ID            uuid.UUID           `json:"id" db:"id"`
	RequestedBy   uuid.UUID           `json:"requested_by" db:"requested_by"`
	EntityType    EntityType          `json:"entity_type" db:"entity_type"`
	EntityID      *uuid.UUID          `json:"entity_id,omitempty" db:"entity_id"`
	Action        ChangeAction        `json:"action" db:"action"`
	Payload       json.RawMessage     `json:"payload" db:"payload"`
	RequesterNote *string             `json:"requester_note,omitempty" db:"requester_note"`
	Status        ChangeRequestStatus `json:"status" db:"status"`
	ReviewedBy    *uuid.UUID          `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt    *time.Time          `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewNote    *string             `json:"review_note,omitempty" db:"review_note"`
	CreatedAt     time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at" db:"updated_at"`

	Requester *User `json:"requester,omitempty" db:"-"`
	Reviewer  *User `json:"reviewer,omitempty" db:"-"`
}

type EntityType string

const (
	EntityPerson       EntityType = "PERSON"
	EntityRelationship EntityType = "RELATIONSHIP"
	EntityMedia        EntityType = "MEDIA"
)

type ChangeAction string

const (
	ActionCreate ChangeAction = "CREATE"
	ActionUpdate ChangeAction = "UPDATE"
	ActionDelete ChangeAction = "DELETE"
)

type ChangeRequestStatus string

const (
	StatusPending  ChangeRequestStatus = "PENDING"
	StatusApproved ChangeRequestStatus = "APPROVED"
	StatusRejected ChangeRequestStatus = "REJECTED"
)

type CreateChangeRequestInput struct {
	EntityType    EntityType      `json:"entity_type" validate:"required"`
	EntityID      *uuid.UUID      `json:"entity_id,omitempty"`
	Action        ChangeAction    `json:"action" validate:"required"`
	Payload       json.RawMessage `json:"payload" validate:"required"`
	RequesterNote *string         `json:"requester_note,omitempty" validate:"omitempty,max=500"`
}

type ReviewChangeRequestInput struct {
	Note *string `json:"note,omitempty" validate:"omitempty,max=500"`
}
