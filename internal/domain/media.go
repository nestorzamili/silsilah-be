package domain

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID          uuid.UUID  `json:"id" db:"media_id"`
	PersonID    *uuid.UUID `json:"person_id,omitempty" db:"person_id"`
	UploadedBy  uuid.UUID  `json:"uploaded_by" db:"uploaded_by"`
	FileName    string     `json:"file_name" db:"file_name"`
	FileSize    int64      `json:"file_size" db:"file_size"`
	MimeType    string     `json:"mime_type" db:"mime_type"`
	StoragePath string     `json:"-" db:"storage_path"`
	URL         string     `json:"url" db:"-"`
	Caption     *string    `json:"caption,omitempty" db:"caption"`
	Status      string     `json:"status" db:"status"`
	TakenAt     *time.Time `json:"taken_at,omitempty" db:"taken_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

type UploadMediaInput struct {
	PersonID *uuid.UUID `json:"person_id,omitempty"`
	Caption  *string    `json:"caption,omitempty" validate:"omitempty,max=500"`
	TakenAt  *time.Time `json:"taken_at,omitempty"`
}
