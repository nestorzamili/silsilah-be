package unit_test

import (
	"context"
	"testing"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/service/graph"
	"silsilah-keluarga/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBFSAncestors(t *testing.T) {
	mockPersonRepo := new(mocks.PersonRepository)
	mockRelRepo := new(mocks.RelationshipRepository)

	ctx := context.Background()
	childID := uuid.New()
	fatherID := uuid.New()
	motherID := uuid.New()
	grandfatherID := uuid.New()

	// Setup Relationships
	// Father -> Child
	rel1 := domain.Relationship{
		PersonA: childID,
		PersonB: fatherID,
		Type:    domain.RelTypeParent,
	}
	// Mother -> Child
	rel2 := domain.Relationship{
		PersonA: childID,
		PersonB: motherID,
		Type:    domain.RelTypeParent,
	}
	// Grandfather -> Father
	rel3 := domain.Relationship{
		PersonA: fatherID,
		PersonB: grandfatherID,
		Type:    domain.RelTypeParent,
	}

	// Setup Persons
	father := domain.Person{ID: fatherID, FirstName: "Father"}
	mother := domain.Person{ID: motherID, FirstName: "Mother"}
	grandfather := domain.Person{ID: grandfatherID, FirstName: "Grandfather"}

	t.Run("Should find all ancestors up to max depth", func(t *testing.T) {
		// Level 0 (Child) -> Level 1 (Father, Mother)
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{childID}).Return([]domain.Relationship{rel1, rel2}, nil).Once()

		// Level 1 (Father, Mother) -> Level 2 (Grandfather)
		// Note: ListByPeople argument order is not guaranteed, so we match loosely or ensure logic handles it
		// Here we assume the logic queries for both IDs in the slice
		mockRelRepo.On("ListByPeople", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
			return len(ids) == 2 && (ids[0] == fatherID || ids[0] == motherID)
		})).Return([]domain.Relationship{rel3}, nil).Once()

		// Level 2 (Grandfather) -> Level 3 (Empty)
		mockRelRepo.On("ListByPeople", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
			return len(ids) == 1 && ids[0] == grandfatherID
		})).Return([]domain.Relationship{}, nil).Once()

		// Get Persons
		mockPersonRepo.On("GetByIDs", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
			return len(ids) == 3
		})).Return([]domain.Person{father, mother, grandfather}, nil)

		nodes, err := graph.BFSAncestors(ctx, mockRelRepo, mockPersonRepo, childID, 5)

		assert.NoError(t, err)
		assert.Len(t, nodes, 3)

		// Verify Generations
		for _, node := range nodes {
			if node.ID == fatherID || node.ID == motherID {
				assert.Equal(t, 1, *node.Generation)
			}
			if node.ID == grandfatherID {
				assert.Equal(t, 2, *node.Generation)
			}
		}
	})
}

func TestBFSDescendants(t *testing.T) {
	mockPersonRepo := new(mocks.PersonRepository)
	mockRelRepo := new(mocks.RelationshipRepository)

	ctx := context.Background()
	parentID := uuid.New()
	childID := uuid.New()
	grandchildID := uuid.New()

	// Parent -> Child
	rel1 := domain.Relationship{
		PersonA: childID,
		PersonB: parentID,
		Type:    domain.RelTypeParent,
	}
	// Child -> Grandchild
	rel2 := domain.Relationship{
		PersonA: grandchildID,
		PersonB: childID,
		Type:    domain.RelTypeParent,
	}

	child := domain.Person{ID: childID, FirstName: "Child"}
	grandchild := domain.Person{ID: grandchildID, FirstName: "Grandchild"}

	t.Run("Should find all descendants", func(t *testing.T) {
		// Level 0 (Parent) -> Level 1 (Child)
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{parentID}).Return([]domain.Relationship{rel1}, nil).Once()

		// Level 1 (Child) -> Level 2 (Grandchild)
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{childID}).Return([]domain.Relationship{rel2}, nil).Once()

		// Level 2 (Grandchild) -> Empty
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{grandchildID}).Return([]domain.Relationship{}, nil).Once()

		mockPersonRepo.On("GetByIDs", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
			return len(ids) == 2
		})).Return([]domain.Person{child, grandchild}, nil)

		nodes, err := graph.BFSDescendants(ctx, mockRelRepo, mockPersonRepo, parentID, 5)

		assert.NoError(t, err)
		assert.Len(t, nodes, 2)
	})
}

func TestBFSShortestPath(t *testing.T) {
	mockRelRepo := new(mocks.RelationshipRepository)

	ctx := context.Background()
	p1 := uuid.New()
	p2 := uuid.New()
	p3 := uuid.New()

	// p1 <-> p2 (Spouse)
	rel1 := domain.Relationship{
		PersonA: p1,
		PersonB: p2,
		Type:    domain.RelTypeSpouse,
	}
	// p2 -> p3 (Parent of p3 is p2) - actually rel is Child(p3) -> Parent(p2)
	rel2 := domain.Relationship{
		PersonA: p3,
		PersonB: p2,
		Type:    domain.RelTypeParent,
	}

	t.Run("Should find path between spouse and child", func(t *testing.T) {
		// p1 neighbors
		mockRelRepo.On("ListByPerson", ctx, p1).Return([]domain.Relationship{rel1}, nil).Once()

		// p2 neighbors (found target p3 here if checking neighbors of p2, but we are at p1)
		// BFS Level 0: p1. Neighbors: p2.
		// BFS Level 1: p2. Neighbors: p1, p3.

		mockRelRepo.On("ListByPerson", ctx, p2).Return([]domain.Relationship{rel1, rel2}, nil).Once()

		path, err := graph.BFSShortestPath(ctx, mockRelRepo, p1, p3, 5)

		assert.NoError(t, err)
		assert.Equal(t, []uuid.UUID{p1, p2, p3}, path)
	})
}

func TestGetSiblingsLogic(t *testing.T) {
	mockPersonRepo := new(mocks.PersonRepository)
	mockRelRepo := new(mocks.RelationshipRepository)

	ctx := context.Background()
	meID := uuid.New()
	brotherID := uuid.New()
	dadID := uuid.New()
	momID := uuid.New()

	// Me -> Dad
	rel1 := domain.Relationship{PersonA: meID, PersonB: dadID, Type: domain.RelTypeParent}
	// Me -> Mom
	rel2 := domain.Relationship{PersonA: meID, PersonB: momID, Type: domain.RelTypeParent}

	// Brother -> Dad
	rel3 := domain.Relationship{PersonA: brotherID, PersonB: dadID, Type: domain.RelTypeParent}
	// Brother -> Mom
	rel4 := domain.Relationship{PersonA: brotherID, PersonB: momID, Type: domain.RelTypeParent}

	brother := domain.Person{ID: brotherID, FirstName: "Brother"}

	t.Run("Should return full sibling", func(t *testing.T) {
		// 1. Get parents of 'me'
		mockRelRepo.On("ListByPerson", ctx, meID).Return([]domain.Relationship{rel1, rel2}, nil).Once()

		// 2. Get relationships of parents (to find other children)
		// ListByPeople([]uuid{dadID, momID})
		mockRelRepo.On("ListByPeople", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
			return len(ids) == 2
		})).Return([]domain.Relationship{rel1, rel2, rel3, rel4}, nil).Once()

		// 3. Get sibling details
		mockPersonRepo.On("GetByIDs", ctx, []uuid.UUID{brotherID}).Return([]domain.Person{brother}, nil).Once()

		// 4. Check sibling type (FULL/HALF)
		// Need sibling's parents to compare
		mockRelRepo.On("ListByPeople", ctx, []uuid.UUID{brotherID}).Return([]domain.Relationship{rel3, rel4}, nil).Once()

		siblings, err := graph.GetSiblingsLogic(ctx, mockRelRepo, mockPersonRepo, meID)

		assert.NoError(t, err)
		assert.Len(t, siblings, 1)
		assert.Equal(t, "FULL", siblings[0].SiblingType)
		assert.Equal(t, brotherID, siblings[0].Person.ID)
	})
}
