package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

var (
	ErrPersonNotFound = errors.New("person not found")
)

type PersonService interface {
	Create(ctx context.Context, userID uuid.UUID, input domain.CreatePersonInput) (*domain.Person, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error)
	GetByIDWithRelationships(ctx context.Context, personID uuid.UUID) (*domain.PersonWithRelationships, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, input domain.UpdatePersonInput) (*domain.Person, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params domain.PaginationParams) (domain.PaginatedResponse[domain.Person], error)
	Search(ctx context.Context, query string, limit int) ([]domain.Person, error)
	GetAncestors(ctx context.Context, personID uuid.UUID) ([]domain.Person, error)
}

type personService struct {
	personRepo       repository.PersonRepository
	relationshipRepo repository.RelationshipRepository
	auditRepo        repository.AuditLogRepository
	redis            *redis.Client
}

func NewPersonService(personRepo repository.PersonRepository, relationshipRepo repository.RelationshipRepository, auditRepo repository.AuditLogRepository, redis *redis.Client) PersonService {
	return &personService{
		personRepo:       personRepo,
		relationshipRepo: relationshipRepo,
		auditRepo:        auditRepo,
		redis:            redis,
	}
}

func (s *personService) Create(ctx context.Context, userID uuid.UUID, input domain.CreatePersonInput) (*domain.Person, error) {
	isAlive := true
	if input.IsAlive != nil {
		isAlive = *input.IsAlive
	}

	person := &domain.Person{
		ID:          uuid.New(),
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Nickname:    input.Nickname,
		Gender:      input.Gender,
		BirthDate:   input.BirthDate,
		BirthPlace:  input.BirthPlace,
		DeathDate:   input.DeathDate,
		DeathPlace:  input.DeathPlace,
		Bio:         input.Bio,
		Occupation:  input.Occupation,
		Religion:    input.Religion,
		Nationality: input.Nationality,
		Education:   input.Education,
		Phone:       input.Phone,
		Email:       input.Email,
		Address:     input.Address,
		IsAlive:     isAlive,
		CreatedBy:   userID,
	}

	if err := s.personRepo.Create(ctx, person); err != nil {
		return nil, err
	}

	_ = repository.CreateAuditLog(s.auditRepo, ctx, domain.CreateAuditLogInput{
		UserID:     userID,
		Action:     "CREATE",
		EntityType: "PERSON",
		EntityID:   person.ID,
		NewValue:   person,
	})

	if s.redis != nil {
		_ = s.redis.Del(ctx, "family:graph").Err()
	}

	return person, nil
}

func (s *personService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	person, err := s.personRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}
	return person, nil
}

func (s *personService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, input domain.UpdatePersonInput) (*domain.Person, error) {
	person, err := s.personRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}

	oldPerson := *person

	if input.FirstName != nil {
		person.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		person.LastName = *input.LastName
	}
	if input.Nickname != nil {
		person.Nickname = *input.Nickname
	}
	if input.Gender != nil {
		person.Gender = *input.Gender
	}
	if input.BirthDate != nil {
		person.BirthDate = *input.BirthDate
	}
	if input.BirthPlace != nil {
		person.BirthPlace = *input.BirthPlace
	}
	if input.DeathDate != nil {
		person.DeathDate = *input.DeathDate
	}
	if input.DeathPlace != nil {
		person.DeathPlace = *input.DeathPlace
	}
	if input.Bio != nil {
		person.Bio = *input.Bio
	}
	if input.AvatarURL != nil {
		person.AvatarURL = *input.AvatarURL
	}
	if input.Occupation != nil {
		person.Occupation = *input.Occupation
	}
	if input.Religion != nil {
		person.Religion = *input.Religion
	}
	if input.Nationality != nil {
		person.Nationality = *input.Nationality
	}
	if input.Education != nil {
		person.Education = *input.Education
	}
	if input.Phone != nil {
		person.Phone = *input.Phone
	}
	if input.Email != nil {
		person.Email = *input.Email
	}
	if input.Address != nil {
		person.Address = *input.Address
	}
	if input.IsAlive != nil {
		person.IsAlive = *input.IsAlive
	}

	if err := s.personRepo.Update(ctx, person); err != nil {
		return nil, err
	}

	_ = repository.CreateAuditLog(s.auditRepo, ctx, domain.CreateAuditLogInput{
		UserID:     userID,
		Action:     "UPDATE",
		EntityType: "PERSON",
		EntityID:   person.ID,
		OldValue:   oldPerson,
		NewValue:   *person,
	})

	if s.redis != nil {
		_ = s.redis.Del(ctx, "family:graph").Err()
	}

	return person, nil
}

func (s *personService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.personRepo.Delete(ctx, id); err != nil {
		return err
	}

	if s.redis != nil {
		_ = s.redis.Del(ctx, "family:graph").Err()
	}

	return nil
}

func (s *personService) List(ctx context.Context, params domain.PaginationParams) (domain.PaginatedResponse[domain.Person], error) {
	persons, total, err := s.personRepo.List(ctx, params)
	if err != nil {
		return domain.PaginatedResponse[domain.Person]{}, err
	}

	return domain.NewPaginatedResponse(persons, params.Page, params.PageSize, total), nil
}

func (s *personService) Search(ctx context.Context, query string, limit int) ([]domain.Person, error) {
	return s.personRepo.Search(ctx, query, limit)
}

func (s *personService) GetByIDWithRelationships(ctx context.Context, personID uuid.UUID) (*domain.PersonWithRelationships, error) {
	person, err := s.personRepo.GetByID(ctx, personID)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}

	result := &domain.PersonWithRelationships{
		Person:        *person,
		Parents:       []domain.ParentInfo{},
		Spouses:       []domain.SpouseInfo{},
		Children:      []domain.Person{},
		Siblings:      []domain.SiblingInfo{},
		Relationships: []domain.RelationshipInfo{},
	}

	relationships, err := s.relationshipRepo.GetByPerson(ctx, personID)
	if err != nil {
		return result, nil
	}

	for _, rel := range relationships {
		switch rel.Type {
		case domain.RelTypeParent:
			if rel.PersonA == personID {
				if p, err := s.personRepo.GetByID(ctx, rel.PersonB); err == nil && p != nil {
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
				if p, err := s.personRepo.GetByID(ctx, rel.PersonA); err == nil && p != nil {
					result.Children = append(result.Children, *p)
					result.Relationships = append(result.Relationships, domain.RelationshipInfo{
						ID:            rel.ID,
						Type:          "CHILD",
						RelatedPerson: p,
					})
				}
			}
		case domain.RelTypeSpouse:
			otherID := rel.PersonA
			if rel.PersonA == personID {
				otherID = rel.PersonB
			}
			if p, err := s.personRepo.GetByID(ctx, otherID); err == nil && p != nil {
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
	}

	siblings, err := s.relationshipRepo.GetSiblings(ctx, personID)
	if err == nil {
		result.Siblings = siblings
		for _, sib := range siblings {
			result.Relationships = append(result.Relationships, domain.RelationshipInfo{
				ID:            uuid.Nil,
				Type:          "SIBLING",
				Role:          sib.SiblingType,
				RelatedPerson: &sib.Person,
			})
		}
	}

	return result, nil
}

func (s *personService) GetAncestors(ctx context.Context, personID uuid.UUID) ([]domain.Person, error) {
	return s.getAncestorsRecursive(ctx, personID, 0, make(map[uuid.UUID]bool))
}

func (s *personService) getAncestorsRecursive(ctx context.Context, personID uuid.UUID, depth int, visited map[uuid.UUID]bool) ([]domain.Person, error) {
	if depth >= 10 || visited[personID] {
		return nil, nil
	}
	visited[personID] = true

	relationships, err := s.relationshipRepo.GetByPerson(ctx, personID)
	if err != nil {
		return nil, err
	}

	var ancestors []domain.Person
	for _, rel := range relationships {
		if rel.Type == domain.RelTypeParent && rel.PersonA == personID {
			parent, err := s.personRepo.GetByID(ctx, rel.PersonB)
			if err != nil || parent == nil {
				continue
			}

			ancestors = append(ancestors, *parent)
			
			grandparents, err := s.getAncestorsRecursive(ctx, rel.PersonB, depth+1, visited)
			if err == nil {
				ancestors = append(ancestors, grandparents...)
			}
		}
	}

	return ancestors, nil
}
