package helpers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

func BuildPersonWithRelationships(
	ctx context.Context,
	personRepo repository.PersonRepository,
	relRepo repository.RelationshipRepository,
	person *domain.Person,
) (*domain.PersonWithRelationships, error) {
	
	result := &domain.PersonWithRelationships{
		Person:        *person,
		Parents:       []domain.ParentInfo{},
		Spouses:       []domain.SpouseInfo{},
		Children:      []domain.Person{},
		Siblings:      []domain.SiblingInfo{},
		Relationships: []domain.RelationshipInfo{},
	}

	relationships, err := relRepo.GetByPerson(ctx, person.ID)
	if err != nil {
		return result, nil
	}

	for _, rel := range relationships {
		switch rel.Type {
		case domain.RelTypeParent:
			processParentRelationship(ctx, personRepo, rel, person.ID, result)
		case domain.RelTypeSpouse:
			processSpouseRelationship(ctx, personRepo, rel, person.ID, result)
		}
	}

	return result, nil
}

func processParentRelationship(
	ctx context.Context,
	personRepo repository.PersonRepository,
	rel domain.Relationship,
	personID uuid.UUID,
	result *domain.PersonWithRelationships,
) {
	if rel.PersonA == personID {
		if p, err := personRepo.GetByID(ctx, rel.PersonB); err == nil && p != nil {
			role := "PARENT"
			if len(rel.Metadata) > 0 {
				var meta domain.ParentMetadata
				if err := json.Unmarshal(rel.Metadata, &meta); err == nil && meta.Role.IsValid() {
					role = string(meta.Role)
				}
			}
			
			result.Parents = append(result.Parents, domain.ParentInfo{
				Person: *p,
				Role:   role,
			})
			result.Relationships = append(result.Relationships, domain.RelationshipInfo{
				ID:            rel.ID,
				Type:          "PARENT",
				Role:          role,
				RelatedPerson: p,
			})
		}
	} else {
		if p, err := personRepo.GetByID(ctx, rel.PersonA); err == nil && p != nil {
			result.Children = append(result.Children, *p)
			result.Relationships = append(result.Relationships, domain.RelationshipInfo{
				ID:            rel.ID,
				Type:          "CHILD",
				RelatedPerson: p,
			})
		}
	}
}

func processSpouseRelationship(
	ctx context.Context,
	personRepo repository.PersonRepository,
	rel domain.Relationship,
	personID uuid.UUID,
	result *domain.PersonWithRelationships,
) {
	otherID := rel.PersonA
	if rel.PersonA == personID {
		otherID = rel.PersonB
	}
	if p, err := personRepo.GetByID(ctx, otherID); err == nil && p != nil {
		isConsanguineous := false
		var metadata any
		if len(rel.Metadata) > 0 {
			var meta domain.SpouseMetadata
			if err := json.Unmarshal(rel.Metadata, &meta); err == nil {
				isConsanguineous = meta.IsConsanguineous
				metadata = meta
			}
		}
		result.Spouses = append(result.Spouses, domain.SpouseInfo{
			Person:           *p,
			IsConsanguineous: isConsanguineous,
		})
		result.Relationships = append(result.Relationships, domain.RelationshipInfo{
			ID:            rel.ID,
			Type:          "SPOUSE",
			RelatedPerson: p,
			Metadata:      metadata,
		})
	}
}
