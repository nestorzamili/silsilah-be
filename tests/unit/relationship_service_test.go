package unit_test

import (
	"context"
	"testing"
	"time"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/service/relationship"
	"silsilah-keluarga/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRelationshipService_Create(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	p1ID := uuid.New()
	p2ID := uuid.New()

	p1 := &domain.Person{ID: p1ID, FirstName: "Parent"}
	p2 := &domain.Person{ID: p2ID, FirstName: "Child"}

	input := domain.CreateRelationshipInput{
		PersonA: p1ID,
		PersonB: p2ID,
		Type:    domain.RelTypeParent,
	}

	setup := func() (relationship.Service, *mocks.PersonRepository, *mocks.RelationshipRepository, *mocks.AuditLogRepository) {
		mockPersonRepo := new(mocks.PersonRepository)
		mockRelRepo := new(mocks.RelationshipRepository)
		mockAuditRepo := new(mocks.AuditLogRepository)
		svc := relationship.NewService(mockRelRepo, mockPersonRepo, mockAuditRepo, nil)
		return svc, mockPersonRepo, mockRelRepo, mockAuditRepo
	}

	t.Run("Success", func(t *testing.T) {
		svc, mockPersonRepo, mockRelRepo, mockAuditRepo := setup()
		// Mock Get Persons
		mockPersonRepo.On("GetByID", ctx, p1ID).Return(p1, nil).Once()
		mockPersonRepo.On("GetByID", ctx, p2ID).Return(p2, nil).Once()

		// Mock Cycle Check (bfsDescendants)
		// Return empty descendants for p1, so p2 is not a descendant
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{p1ID}).Return([]domain.Relationship{}, nil).Once()

		// Mock Create
		mockRelRepo.On("Create", ctx, mock.AnythingOfType("*domain.Relationship")).Return(nil).Once()

		// Mock Audit
		mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

		rel, err := svc.Create(ctx, userID, input)

		assert.NoError(t, err)
		assert.NotNil(t, rel)
		assert.Equal(t, p1ID, rel.PersonA)
		assert.Equal(t, p2ID, rel.PersonB)
	})

	t.Run("Self Relation Error", func(t *testing.T) {
		svc, mockPersonRepo, _, _ := setup()
		inputSelf := domain.CreateRelationshipInput{
			PersonA: p1ID,
			PersonB: p1ID,
			Type:    domain.RelTypeParent,
		}
		
		mockPersonRepo.On("GetByID", ctx, p1ID).Return(p1, nil).Twice()

		rel, err := svc.Create(ctx, userID, inputSelf)

		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.Contains(t, err.Error(), "cannot create relationship with self")
	})

	t.Run("Cycle Error", func(t *testing.T) {
		svc, mockPersonRepo, mockRelRepo, _ := setup()
		
		mockPersonRepo.On("GetByID", ctx, p1ID).Return(p1, nil).Once()
		mockPersonRepo.On("GetByID", ctx, p2ID).Return(p2, nil).Once()

		// Mock BFS: p1 -> p2 (child)
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{p1ID}).Return([]domain.Relationship{
			{PersonA: p2ID, PersonB: p1ID, Type: domain.RelTypeParent},
		}, nil).Once()
		
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{p2ID}).Return([]domain.Relationship{}, nil).Once()
		
		mockPersonRepo.On("GetByIDs", ctx, mock.Anything).Return([]domain.Person{*p2}, nil).Once()

		rel, err := svc.Create(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.Contains(t, err.Error(), "cycle detected")
	})
	
	t.Run("Age Validation Error", func(t *testing.T) {
		svc, mockPersonRepo, mockRelRepo, _ := setup()
		now := time.Now()
		childBirth := now.AddDate(-20, 0, 0)
		parentBirth := now.AddDate(-10, 0, 0) // Parent is 10, Child is 20
		
		p1Young := &domain.Person{ID: p1ID, FirstName: "Child", BirthDate: &childBirth}
		p2Young := &domain.Person{ID: p2ID, FirstName: "Parent", BirthDate: &parentBirth}
		
		mockPersonRepo.On("GetByID", ctx, p1ID).Return(p1Young, nil).Once()
		mockPersonRepo.On("GetByID", ctx, p2ID).Return(p2Young, nil).Once()
		
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{p1ID}).Return([]domain.Relationship{}, nil).Once()

		rel, err := svc.Create(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, rel)
		assert.Contains(t, err.Error(), "parent cannot be younger than child")
	})

	t.Run("Consanguinity Warning", func(t *testing.T) {
		svc, mockPersonRepo, mockRelRepo, mockAuditRepo := setup()
		
		inputSpouse := domain.CreateRelationshipInput{
			PersonA: p1ID,
			PersonB: p2ID,
			Type:    domain.RelTypeSpouse,
		}
		
		mockPersonRepo.On("GetByID", ctx, p1ID).Return(p1, nil).Once()
		mockPersonRepo.On("GetByID", ctx, p2ID).Return(p2, nil).Once()
		
		// Cycle check (validateRelationship)
		// For spouse, it doesn't run bfsDescendants
		// validateRelationship only runs bfsDescendants if RelTypeParent
		
		// Consanguinity Check -> bfsShortestPath(p1, p2)
		mockRelRepo.On("ListByPerson", ctx, p1ID).Return([]domain.Relationship{}, nil).Once()
		
		// Create should succeed
		mockRelRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.Relationship) bool {
			return r.Type == domain.RelTypeSpouse
		})).Return(nil).Once()
		
		mockAuditRepo.On("Create", ctx, mock.Anything).Return(nil).Once()

		rel, err := svc.Create(ctx, userID, inputSpouse)

		assert.NoError(t, err)
		assert.NotNil(t, rel)
		assert.Equal(t, domain.RelTypeSpouse, rel.Type)
	})
}
