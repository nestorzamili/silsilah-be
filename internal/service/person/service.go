package person

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service/graph"
	"silsilah-keluarga/internal/service/notification"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, input domain.CreatePersonInput) (*domain.Person, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error)
	GetByIDWithRelationships(ctx context.Context, personID uuid.UUID) (*domain.PersonWithRelationships, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, input domain.UpdatePersonInput) (*domain.Person, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params domain.PaginationParams) (domain.PaginatedResponse[domain.Person], error)
	Search(ctx context.Context, query string, limit int) ([]domain.Person, error)
	GetAncestors(ctx context.Context, personID uuid.UUID) ([]domain.Person, error)
	SetNotificationService(notifSvc notification.Service)
}

type service struct {
	personRepo       repository.PersonRepository
	relationshipRepo repository.RelationshipRepository
	auditRepo        repository.AuditLogRepository
	redis            *redis.Client
	notifSvc         notification.Service
}

func NewService(personRepo repository.PersonRepository, relationshipRepo repository.RelationshipRepository, auditRepo repository.AuditLogRepository, redis *redis.Client) Service {
	return &service{
		personRepo:       personRepo,
		relationshipRepo: relationshipRepo,
		auditRepo:        auditRepo,
		redis:            redis,
	}
}

func (s *service) SetNotificationService(notifSvc notification.Service) {
	s.notifSvc = notifSvc
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, input domain.CreatePersonInput) (*domain.Person, error) {
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

	if s.notifSvc != nil {
		go func() {
			_ = s.notifSvc.NotifyPersonAdded(context.Background(), person.ID, userID)
		}()
	}

	return person, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	person, err := s.personRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, domain.ErrPersonNotFound
	}
	return person, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, input domain.UpdatePersonInput) (*domain.Person, error) {
	person, err := s.personRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, domain.ErrPersonNotFound
	}

	oldPerson := *person

	if input.FirstName != nil {
		person.FirstName = *input.FirstName
	}
	if input.LastName.Set {
		person.LastName = input.LastName.Value
	}
	if input.Nickname.Set {
		person.Nickname = input.Nickname.Value
	}
	if input.Gender.Set {
		if input.Gender.Value != nil {
			person.Gender = *input.Gender.Value
		} else {
			person.Gender = domain.GenderUnknown
		}
	}
	if input.BirthDate.Set {
		person.BirthDate = input.BirthDate.Value
	}
	if input.BirthPlace.Set {
		person.BirthPlace = input.BirthPlace.Value
	}
	if input.DeathDate.Set {
		person.DeathDate = input.DeathDate.Value
	}
	if input.DeathPlace.Set {
		person.DeathPlace = input.DeathPlace.Value
	}
	if input.Bio.Set {
		person.Bio = input.Bio.Value
	}
	if input.AvatarURL.Set {
		person.AvatarURL = input.AvatarURL.Value
	}
	if input.Occupation.Set {
		person.Occupation = input.Occupation.Value
	}
	if input.Religion.Set {
		person.Religion = input.Religion.Value
	}
	if input.Nationality.Set {
		person.Nationality = input.Nationality.Value
	}
	if input.Education.Set {
		person.Education = input.Education.Value
	}
	if input.Phone.Set {
		person.Phone = input.Phone.Value
	}
	if input.Email.Set {
		person.Email = input.Email.Value
	}
	if input.Address.Set {
		person.Address = input.Address.Value
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

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.personRepo.Delete(ctx, id); err != nil {
		return err
	}

	if s.redis != nil {
		_ = s.redis.Del(ctx, "family:graph").Err()
	}

	return nil
}

func (s *service) List(ctx context.Context, params domain.PaginationParams) (domain.PaginatedResponse[domain.Person], error) {
	persons, total, err := s.personRepo.List(ctx, params)
	if err != nil {
		return domain.PaginatedResponse[domain.Person]{}, err
	}

	return domain.NewPaginatedResponse(persons, params.Page, params.PageSize, total), nil
}

func (s *service) Search(ctx context.Context, query string, limit int) ([]domain.Person, error) {
	return s.personRepo.Search(ctx, query, limit)
}

func (s *service) GetByIDWithRelationships(ctx context.Context, personID uuid.UUID) (*domain.PersonWithRelationships, error) {
	person, err := s.personRepo.GetByID(ctx, personID)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, domain.ErrPersonNotFound
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

	siblings, err := graph.GetSiblingsLogic(ctx, s.relationshipRepo, s.personRepo, personID)
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

func (s *service) GetAncestors(ctx context.Context, personID uuid.UUID) ([]domain.Person, error) {
	return s.getAncestorsRecursive(ctx, personID, 0, make(map[uuid.UUID]bool))
}

func (s *service) getAncestorsRecursive(ctx context.Context, personID uuid.UUID, depth int, visited map[uuid.UUID]bool) ([]domain.Person, error) {
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
