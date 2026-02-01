package domain

import (
	"time"

	"github.com/google/uuid"
)

type Person struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	FirstName  string     `json:"first_name" db:"first_name"`
	LastName   *string    `json:"last_name,omitempty" db:"last_name"`
	Nickname   *string    `json:"nickname,omitempty" db:"nickname"`
	Gender     Gender     `json:"gender" db:"gender"`
	BirthDate  *time.Time `json:"birth_date,omitempty" db:"birth_date"`
	BirthPlace  *string    `json:"birth_place,omitempty" db:"birth_place"`
	DeathDate   *time.Time `json:"death_date,omitempty" db:"death_date"`
	DeathPlace  *string    `json:"death_place,omitempty" db:"death_place"`
	Bio         *string    `json:"bio,omitempty" db:"bio"`
	AvatarURL   *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Occupation  *string    `json:"occupation,omitempty" db:"occupation"`
	Religion    *string    `json:"religion,omitempty" db:"religion"`
	Nationality *string    `json:"nationality,omitempty" db:"nationality"`
	Education   *string    `json:"education,omitempty" db:"education"`
	Phone       *string    `json:"phone,omitempty" db:"phone"`
	Email       *string    `json:"email,omitempty" db:"email"`
	Address     *string    `json:"address,omitempty" db:"address"`
	IsAlive     bool       `json:"is_alive" db:"is_alive"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

type Gender string

const (
	GenderMale    Gender = "MALE"
	GenderFemale  Gender = "FEMALE"
	GenderUnknown Gender = "UNKNOWN"
)

func (g Gender) IsValid() bool {
	switch g {
	case GenderMale, GenderFemale, GenderUnknown:
		return true
	}
	return false
}

type CreatePersonInput struct {
	FirstName   string     `json:"first_name" validate:"required,min=1,max=100"`
	LastName    *string    `json:"last_name,omitempty" validate:"omitempty,max=100"`
	Nickname    *string    `json:"nickname,omitempty" validate:"omitempty,max=50"`
	Gender      Gender     `json:"gender" validate:"required"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
	BirthPlace  *string    `json:"birth_place,omitempty" validate:"omitempty,max=200"`
	DeathDate   *time.Time `json:"death_date,omitempty"`
	DeathPlace  *string    `json:"death_place,omitempty" validate:"omitempty,max=200"`
	Bio         *string    `json:"bio,omitempty" validate:"omitempty,max=2000"`
	Occupation  *string    `json:"occupation,omitempty" validate:"omitempty,max=200"`
	Religion    *string    `json:"religion,omitempty" validate:"omitempty,max=50"`
	Nationality *string    `json:"nationality,omitempty" validate:"omitempty,max=100"`
	Education   *string    `json:"education,omitempty" validate:"omitempty,max=200"`
	Phone       *string    `json:"phone,omitempty" validate:"omitempty,max=20"`
	Email       *string    `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Address     *string    `json:"address,omitempty" validate:"omitempty,max=500"`
	IsAlive     *bool      `json:"is_alive,omitempty"`
}

type UpdatePersonInput struct {
	FirstName   *string     `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName    **string    `json:"last_name,omitempty" validate:"omitempty,max=100"`
	Nickname    **string    `json:"nickname,omitempty" validate:"omitempty,max=50"`
	Gender      *Gender     `json:"gender,omitempty"`
	BirthDate   **time.Time `json:"birth_date,omitempty"`
	BirthPlace  **string    `json:"birth_place,omitempty" validate:"omitempty,max=200"`
	DeathDate   **time.Time `json:"death_date,omitempty"`
	DeathPlace  **string    `json:"death_place,omitempty" validate:"omitempty,max=200"`
	Bio         **string    `json:"bio,omitempty" validate:"omitempty,max=2000"`
	AvatarURL   **string    `json:"avatar_url,omitempty"`
	Occupation  **string    `json:"occupation,omitempty" validate:"omitempty,max=200"`
	Religion    **string    `json:"religion,omitempty" validate:"omitempty,max=50"`
	Nationality **string    `json:"nationality,omitempty" validate:"omitempty,max=100"`
	Education   **string    `json:"education,omitempty" validate:"omitempty,max=200"`
	Phone       **string    `json:"phone,omitempty" validate:"omitempty,max=20"`
	Email       **string    `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Address     **string    `json:"address,omitempty" validate:"omitempty,max=500"`
	IsAlive     *bool       `json:"is_alive,omitempty"`
}

type PersonSearchInput struct {
	Query  string `json:"query" validate:"required,min=2"`
	Limit  int    `json:"limit" validate:"omitempty,min=1,max=50"`
	Offset int    `json:"offset" validate:"omitempty,min=0"`
}

func (p *Person) FullName() string {
	if p.LastName != nil {
		return p.FirstName + " " + *p.LastName
	}
	return p.FirstName
}

type PersonWithRelationships struct {
	Person
	Parents       []ParentInfo       `json:"parents"`
	Spouses       []SpouseInfo       `json:"spouses"`
	Children      []Person           `json:"children"`
	Siblings      []SiblingInfo      `json:"siblings"`
	Relationships []RelationshipInfo `json:"relationships"`
}

type RelationshipInfo struct {
	ID            uuid.UUID `json:"id"`
	Type          string    `json:"type"`
	Role          string    `json:"role,omitempty"`
	RelatedPerson *Person   `json:"related_person,omitempty"`
	Metadata      any       `json:"metadata,omitempty"`
}

type ParentInfo struct {
	Person Person `json:"person"`
	Role   string `json:"role,omitempty"`
}

type SpouseInfo struct {
	Person           Person  `json:"person"`
	IsConsanguineous bool    `json:"is_consanguineous"`
	MarriageDate     *string `json:"marriage_date,omitempty"`
}

type SiblingInfo struct {
	Person     Person `json:"person"`
	SiblingType string `json:"sibling_type"`
}
