package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	Email                   string     `json:"email" db:"email"`
	PasswordHash            string     `json:"-" db:"password_hash"`
	FullName                string     `json:"full_name" db:"full_name"`
	AvatarURL               *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio                     *string    `json:"bio,omitempty" db:"bio"`
	Role                    string     `json:"role" db:"role"`
	PersonID                *uuid.UUID `json:"person_id" db:"person_id"`
	IsActive                bool       `json:"is_active" db:"is_active"`
	IsEmailVerified         bool       `json:"is_email_verified" db:"is_email_verified"`
	EmailVerificationToken  *string    `json:"-" db:"email_verification_token"`
	EmailVerificationSentAt *time.Time `json:"-" db:"email_verification_sent_at"`
	PasswordResetToken      *string    `json:"-" db:"password_reset_token"`
	PasswordResetExpiresAt  *time.Time `json:"-" db:"password_reset_expires_at"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt               *time.Time `json:"-" db:"deleted_at"`
}

type CreateUserInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2"`
	Role     string `json:"role" validate:"omitempty,oneof=member editor developer"`
}

type UpdateUserInput struct {
	FullName  *string     `json:"full_name,omitempty" validate:"omitempty,min=2"`
	Email     *string     `json:"email,omitempty" validate:"omitempty,email"`
	Password  *string     `json:"password,omitempty" validate:"omitempty,min=8"`
	AvatarURL **string    `json:"avatar_url,omitempty"`
	Bio       **string    `json:"bio,omitempty"`
	Role      *string     `json:"role,omitempty" validate:"omitempty,oneof=member editor developer"`
	PersonID  **uuid.UUID `json:"person_id"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AssignRoleInput struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=member editor developer"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserRole string

const (
	RoleMember    UserRole = "member"
	RoleEditor    UserRole = "editor"
	RoleDeveloper UserRole = "developer"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleMember, RoleEditor, RoleDeveloper:
		return true
	default:
		return false
	}
}

func (u *User) HasRole(requiredRole string) bool {
	switch requiredRole {
	case "developer":
		return u.Role == "developer"
	case "editor":
		return u.Role == "editor" || u.Role == "developer"
	case "member":
		return u.Role == "member" || u.Role == "editor" || u.Role == "developer"
	default:
		return false
	}
}
