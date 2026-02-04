package relationship

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service/graph"
	"silsilah-keluarga/internal/service/notification"
)

var (
	ErrRelationshipNotFound  = errors.New("relationship not found")
	ErrSelfRelation          = errors.New("cannot create relationship with self")
	ErrInvalidRelationType   = errors.New("invalid relationship type")
	ErrDuplicateRelationship = errors.New("relationship already exists")
	ErrDuplicateParentRole   = errors.New("person already has a parent with this role")
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, input domain.CreateRelationshipInput) (*domain.Relationship, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error)
	Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input domain.UpdateRelationshipInput) (*domain.Relationship, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error)
	ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error)
	SetNotificationService(notifSvc notification.Service)
}

type service struct {
	relRepo    repository.RelationshipRepository
	personRepo repository.PersonRepository
	auditRepo  repository.AuditLogRepository
	redis      *redis.Client
	notifSvc   notification.Service
}

func NewService(relRepo repository.RelationshipRepository, personRepo repository.PersonRepository, auditRepo repository.AuditLogRepository, redis *redis.Client) Service {
	return &service{
		relRepo:    relRepo,
		personRepo: personRepo,
		auditRepo:  auditRepo,
		redis:      redis,
	}
}

func (s *service) SetNotificationService(notifSvc notification.Service) {
	s.notifSvc = notifSvc
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, input domain.CreateRelationshipInput) (*domain.Relationship, error) {
	personA, err := s.personRepo.GetByID(ctx, input.PersonA)
	if err != nil {
		return nil, err
	}
	if personA == nil {
		return nil, domain.ErrPersonNotFound
	}

	personB, err := s.personRepo.GetByID(ctx, input.PersonB)
	if err != nil {
		return nil, err
	}
	if personB == nil {
		return nil, domain.ErrPersonNotFound
	}

	if err := s.validateRelationship(ctx, personA, personB, input.Type); err != nil {
		return nil, err
	}

	if input.Type == domain.RelTypeSpouse {
		path, err := graph.BFSShortestPath(ctx, s.relRepo, personA.ID, personB.ID, 10) // Max depth 10
		if err == nil && len(path) > 0 {
			degree := len(path) - 1
			isConsanguineous := true

			var meta domain.SpouseMetadata
			if len(input.Metadata) > 0 {
				_ = json.Unmarshal(input.Metadata, &meta)
			}

			meta.IsConsanguineous = isConsanguineous
			meta.ConsanguinityDegree = &degree

			metaBytes, _ := json.Marshal(meta)
			input.Metadata = metaBytes
		}
	}

	rel := &domain.Relationship{
		ID:        uuid.New(),
		PersonA:   input.PersonA,
		PersonB:   input.PersonB,
		Type:      input.Type,
		Metadata:  input.Metadata,
		CreatedBy: userID,
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

	_ = repository.CreateAuditLog(s.auditRepo, ctx, domain.CreateAuditLogInput{
		UserID:     userID,
		Action:     "CREATE_RELATIONSHIP",
		EntityType: "RELATIONSHIP",
		EntityID:   rel.ID,
		NewValue: map[string]interface{}{
			"person_a":    personA.FirstName + " " + stringPtrValue(personA.LastName),
			"person_b":    personB.FirstName + " " + stringPtrValue(personB.LastName),
			"type":        rel.Type,
			"person_a_id": rel.PersonA,
			"person_b_id": rel.PersonB,
		},
	})

	if s.notifSvc != nil {
		go func() {
			_ = s.notifSvc.NotifyRelationshipAdded(context.Background(), rel.ID, userID)
		}()
	}

	return rel, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error) {
	rel, err := s.relRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrRelationshipNotFound
	}
	return rel, nil
}

func (s *service) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input domain.UpdateRelationshipInput) (*domain.Relationship, error) {
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

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	if s.redis != nil {
		_ = s.redis.Del(ctx, "family:graph").Err()
	}
	return s.relRepo.Delete(ctx, id)
}

func (s *service) List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error) {
	return s.relRepo.List(ctx, relType)
}

func (s *service) ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error) {
	return s.relRepo.ListByPerson(ctx, personID)
}

func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *service) validateRelationship(ctx context.Context, personA, personB *domain.Person, relType domain.RelationshipType) error {
	if personA.ID == personB.ID {
		return errors.New("cannot create relationship with self")
	}

	if relType == domain.RelTypeParent {
		descendants, err := graph.BFSDescendants(ctx, s.relRepo, s.personRepo, personA.ID, 100)
		if err == nil {
			for _, d := range descendants {
				if d.ID == personB.ID {
					return errors.New("cycle detected: cannot make a descendant a parent")
				}
			}
		}

		if personA.BirthDate != nil && personB.BirthDate != nil {
			if personB.BirthDate.After(*personA.BirthDate) {
				return errors.New("invalid relationship: parent cannot be younger than child")
			}
		}
	}

	return nil
}
