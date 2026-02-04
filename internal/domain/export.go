package domain

import (
	"time"

	"github.com/google/uuid"
)

type FamilyTreeExport struct {
	RootPersonID  uuid.UUID      `json:"root_person_id"`
	ExportedAt    time.Time      `json:"exported_at"`
	People        []Person       `json:"people"`
	Relationships []Relationship `json:"relationships"`
}
