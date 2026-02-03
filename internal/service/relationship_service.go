package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

var (
	ErrRelationshipNotFound   = errors.New("relationship not found")
	ErrSelfRelation           = errors.New("cannot create relationship with self")
	ErrInvalidRelationType    = errors.New("invalid relationship type")
	ErrDuplicateRelationship  = errors.New("relationship already exists")
	ErrDuplicateParentRole    = errors.New("person already has a parent with this role")
)

type RelationshipService interface {
	Create(ctx context.Context, userID uuid.UUID, input domain.CreateRelationshipInput) (*domain.Relationship, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error)
	Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input domain.UpdateRelationshipInput) (*domain.Relationship, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error)
	ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error)
}

type relationshipService struct {
	relRepo    repository.RelationshipRepository
	personRepo repository.PersonRepository
	auditRepo  repository.AuditLogRepository
	redis      *redis.Client
}

func NewRelationshipService(relRepo repository.RelationshipRepository, personRepo repository.PersonRepository, auditRepo repository.AuditLogRepository, redis *redis.Client) RelationshipService {
	return &relationshipService{
		relRepo:    relRepo,
		personRepo: personRepo,
		auditRepo:  auditRepo,
		redis:      redis,
	}
}

func (s *relationshipService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateRelationshipInput) (*domain.Relationship, error) {
	if input.PersonA == input.PersonB {
		return nil, ErrSelfRelation
	}

	if !input.Type.IsValid() {
		return nil, ErrInvalidRelationType
	}

	personA, err := s.personRepo.GetByID(ctx, input.PersonA)
	if err != nil {
		return nil, err
	}
	if personA == nil {
		return nil, ErrPersonNotFound
	}

	personB, err := s.personRepo.GetByID(ctx, input.PersonB)
	if err != nil {
		return nil, err
	}
	if personB == nil {
		return nil, ErrPersonNotFound
	}

	if input.Type == domain.RelTypeParent && input.Metadata != nil {
		var parentMeta domain.ParentMetadata
		if err := json.Unmarshal(input.Metadata, &parentMeta); err == nil && parentMeta.Role.IsValid() {
			existingParents, err := s.relRepo.ListByPerson(ctx, input.PersonB)
			if err != nil {
				return nil, err
			}
			
			for _, rel := range existingParents {
				if rel.Type == domain.RelTypeParent && rel.PersonB == input.PersonB {
					var existingMeta domain.ParentMetadata
					if err := json.Unmarshal(rel.Metadata, &existingMeta); err == nil {
						if existingMeta.Role == parentMeta.Role {
							return nil, ErrDuplicateParentRole
						}
					}
				}
			}
		}
	}

	var metadata json.RawMessage
	var spouseOrder *int
	var childOrder *int

	if input.Type == domain.RelTypeSpouse {
		if input.SpouseOrder != nil && *input.SpouseOrder > 0 {
			spouseOrder = input.SpouseOrder
		} else {
			existingRels, err := s.relRepo.ListByPerson(ctx, input.PersonA)
			if err != nil {
				return nil, err
			}

			spouseCount := 0
			for _, rel := range existingRels {
				if rel.Type == domain.RelTypeSpouse {
					spouseCount++
				}
			}
			order := spouseCount + 1
			spouseOrder = &order
		}

		consanguinity, err := s.relRepo.CalculateConsanguinity(ctx, input.PersonA, input.PersonB)
		if err == nil && consanguinity != nil {
			var existingMeta domain.SpouseMetadata
			if input.Metadata != nil {
				_ = json.Unmarshal(input.Metadata, &existingMeta)
			}
			existingMeta.IsConsanguineous = consanguinity.IsConsanguineous
			existingMeta.ConsanguinityDegree = consanguinity.ConsanguinityDegree
			existingMeta.CommonAncestors = consanguinity.CommonAncestors

			metadata, _ = json.Marshal(existingMeta)
		} else if input.Metadata != nil {
			metadata = input.Metadata
		} else {
			metadata = json.RawMessage(`{}`)
		}
	} else if input.Type == domain.RelTypeParent {
		if input.ChildOrder != nil && *input.ChildOrder > 0 {
			childOrder = input.ChildOrder
		} else {
			existingRels, err := s.relRepo.ListByPerson(ctx, input.PersonB)
			if err != nil {
				return nil, err
			}

			childCount := 0
			for _, rel := range existingRels {
				if rel.Type == domain.RelTypeParent && rel.PersonB == input.PersonB {
					childCount++
				}
			}
			order := childCount + 1 
			childOrder = &order
		}

		if input.Metadata != nil {
			metadata = input.Metadata
		} else {
			metadata = json.RawMessage(`{}`)
		}
	} else if input.Metadata != nil {
		metadata = input.Metadata
	} else {
		metadata = json.RawMessage(`{}`)
	}

	rel := &domain.Relationship{
		ID:          uuid.New(),
		PersonA:     input.PersonA,
		PersonB:     input.PersonB,
		Type:        input.Type,
		Metadata:    metadata,
		SpouseOrder: spouseOrder,
		ChildOrder:  childOrder,
		CreatedBy:   userID,
	}

	if err := s.relRepo.Create(ctx, rel); err != nil {
		if strings.Contains(err.Error(), "uk_relationship_pair") || strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrDuplicateRelationship
		}
		return nil, err
	}

	if s.redis != nil {
		_ = s.redis.Del(ctx, "family:graph").Err()
	}

	// Create enriched audit log data with person names
	personAName := personA.FirstName
	if personA.LastName != nil {
		personAName += " " + *personA.LastName
	}
	personBName := personB.FirstName
	if personB.LastName != nil {
		personBName += " " + *personB.LastName
	}
	
	auditData := map[string]interface{}{
		"id":            rel.ID,
		"person_a":      rel.PersonA,
		"person_a_name": personAName,
		"person_b":      rel.PersonB,
		"person_b_name": personBName,
		"type":          rel.Type,
		"metadata":      rel.Metadata,
		"spouse_order":  rel.SpouseOrder,
		"child_order":   rel.ChildOrder,
	}

	_ = repository.CreateAuditLog(s.auditRepo, ctx, domain.CreateAuditLogInput{
		UserID:     userID,
		Action:     "CREATE",
		EntityType: "RELATIONSHIP",
		EntityID:   rel.ID,
		NewValue:   auditData,
	})

	return rel, nil
}

func (s *relationshipService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error) {
	rel, err := s.relRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrRelationshipNotFound
	}
	return rel, nil
}

func (s *relationshipService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input domain.UpdateRelationshipInput) (*domain.Relationship, error) {
	rel, err := s.relRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrRelationshipNotFound
	}

	oldValue := *rel

	if input.Metadata != nil {
		rel.Metadata = input.Metadata
	}
	if input.SpouseOrder != nil {
		rel.SpouseOrder = input.SpouseOrder
	}
	if input.ChildOrder != nil {
		rel.ChildOrder = input.ChildOrder
	}

	if err := s.relRepo.Update(ctx, rel); err != nil {
		return nil, err
	}

	if s.redis != nil {
		_ = s.redis.Del(ctx, "family:graph").Err()
	}

	_ = repository.CreateAuditLog(s.auditRepo, ctx, domain.CreateAuditLogInput{
		UserID:     userID,
		Action:     "UPDATE",
		EntityType: "RELATIONSHIP",
		EntityID:   rel.ID,
		OldValue:   oldValue,
		NewValue:   rel,
	})

	return rel, nil
}

func (s *relationshipService) Delete(ctx context.Context, id uuid.UUID) error {
	if s.redis != nil {
		_ = s.redis.Del(ctx, "family:graph").Err()
	}
	return s.relRepo.Delete(ctx, id)
}

func (s *relationshipService) List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error) {
	return s.relRepo.List(ctx, relType)
}

func (s *relationshipService) ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error) {
	return s.relRepo.ListByPerson(ctx, personID)
}
