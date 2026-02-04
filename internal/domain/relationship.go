package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Relationship struct {
	ID          uuid.UUID        `json:"id" db:"relationship_id"`
	PersonA     uuid.UUID        `json:"person_a" db:"person_a"`
	PersonB     uuid.UUID        `json:"person_b" db:"person_b"`
	Type        RelationshipType `json:"type" db:"type"`
	Metadata    json.RawMessage  `json:"metadata,omitempty" db:"metadata"`
	CreatedBy   uuid.UUID        `json:"created_by" db:"created_by"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time       `json:"-" db:"deleted_at"`

	PersonAData *Person `json:"person_a_data,omitempty" db:"-"`
	PersonBData *Person `json:"person_b_data,omitempty" db:"-"`
}

type RelationshipType string

const (
	RelTypeParent RelationshipType = "PARENT"
	RelTypeSpouse RelationshipType = "SPOUSE"
)

func (r RelationshipType) IsValid() bool {
	switch r {
	case RelTypeParent, RelTypeSpouse:
		return true
	}
	return false
}

type SpouseMetadata struct {
	MarriageDate        *time.Time  `json:"marriage_date,omitempty"`
	MarriagePlace       *string     `json:"marriage_place,omitempty"`
	DivorceDate         *time.Time  `json:"divorce_date,omitempty"`
	IsConsanguineous    bool        `json:"is_consanguineous"`
	ConsanguinityDegree *int        `json:"consanguinity_degree,omitempty"`
	CommonAncestors     []uuid.UUID `json:"common_ancestors,omitempty"`
}

type ParentRole string

const (
	ParentRoleFather ParentRole = "FATHER"
	ParentRoleMother ParentRole = "MOTHER"
)

func (r ParentRole) IsValid() bool {
	switch r {
	case ParentRoleFather, ParentRoleMother:
		return true
	}
	return false
}

type ParentMetadata struct {
	Role ParentRole `json:"role"`
}

type CreateRelationshipInput struct {
	PersonA     uuid.UUID        `json:"person_a" validate:"required"`
	PersonB     uuid.UUID        `json:"person_b" validate:"required"`
	Type        RelationshipType `json:"type" validate:"required"`
	Metadata    json.RawMessage  `json:"metadata,omitempty"`
}

type UpdateRelationshipInput struct {
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

type DerivedRelationType string

const (
	DerivedChild      DerivedRelationType = "CHILD"
	DerivedSibling    DerivedRelationType = "SIBLING"
	DerivedGrandparent DerivedRelationType = "GRANDPARENT"
	DerivedGrandchild DerivedRelationType = "GRANDCHILD"
	DerivedUncleAunt  DerivedRelationType = "UNCLE_AUNT"
	DerivedCousin     DerivedRelationType = "COUSIN"
	DerivedNephewNiece DerivedRelationType = "NEPHEW_NIECE"
)

type RelationshipPath struct {
	FromPerson   uuid.UUID           `json:"from_person"`
	ToPerson     uuid.UUID           `json:"to_person"`
	Path         []uuid.UUID         `json:"path"`
	Relationship DerivedRelationType `json:"relationship"`
	Description  string              `json:"description"`
	Degree       int                 `json:"degree"`
}
