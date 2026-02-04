package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service/narrative"
)

type Service interface {
	GetFullGraph(ctx context.Context) (*domain.FamilyGraph, error)
	GetAncestors(ctx context.Context, personID uuid.UUID, maxDepth int) (*domain.AncestorTree, error)
	GetSplitAncestors(ctx context.Context, personID uuid.UUID, maxDepth int) (*domain.SplitAncestorTree, error)
	GetDescendants(ctx context.Context, personID uuid.UUID, maxDepth int) (*domain.DescendantTree, error)
	FindRelationshipPath(ctx context.Context, fromPersonID, toPersonID uuid.UUID, maxDepth int, locale string) (*domain.RelationshipPath, error)
	InvalidateCache(ctx context.Context) error
}

type service struct {
	personRepo repository.PersonRepository
	relRepo    repository.RelationshipRepository
	redis      *redis.Client
	narrative  narrative.Service
}

func NewService(personRepo repository.PersonRepository, relRepo repository.RelationshipRepository, redis *redis.Client, narrative narrative.Service) Service {
	return &service{
		personRepo: personRepo,
		relRepo:    relRepo,
		redis:      redis,
		narrative:  narrative,
	}
}

func (s *service) GetFullGraph(ctx context.Context) (*domain.FamilyGraph, error) {
	cacheKey := "family:graph"

	if s.redis != nil {
		if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
			var graph domain.FamilyGraph
			if json.Unmarshal([]byte(cached), &graph) == nil {
				return &graph, nil
			}
		}
	}

	persons, err := s.personRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	relationships, err := s.relRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	connectedPersonIds := make(map[uuid.UUID]bool)
	for _, r := range relationships {
		connectedPersonIds[r.PersonA] = true
		connectedPersonIds[r.PersonB] = true
	}

	var connectedPersons []domain.Person
	for _, p := range persons {
		if connectedPersonIds[p.ID] {
			connectedPersons = append(connectedPersons, p)
		}
	}

	nodes := make([]domain.GraphNode, len(connectedPersons))
	livingCount := 0
	for i, p := range connectedPersons {
		nodes[i] = personToGraphNode(&p, nil)
		if p.IsAlive {
			livingCount++
		}
	}

	personIdSet := make(map[uuid.UUID]bool)
	for _, p := range connectedPersons {
		personIdSet[p.ID] = true
	}
	edges := buildEdgesForNodes(relationships, personIdSet)
	groups := buildFamilyGroups(relationships, personIdSet)

	graph := &domain.FamilyGraph{
		Nodes:  nodes,
		Edges:  edges,
		Groups: groups,
		Stats: &domain.GraphStats{
			TotalPersons:       len(connectedPersons),
			TotalRelationships: len(relationships),
			LivingPersons:      livingCount,
			DeceasedPersons:    len(connectedPersons) - livingCount,
		},
	}

	if s.redis != nil {
		if graphJSON, err := json.Marshal(graph); err == nil {
			_ = s.redis.Set(ctx, cacheKey, graphJSON, 5*time.Minute).Err()
		}
	}

	return graph, nil
}

func (s *service) GetAncestors(ctx context.Context, personID uuid.UUID, maxDepth int) (*domain.AncestorTree, error) {
	return s.getAncestorTree(ctx, personID, maxDepth)
}

func (s *service) GetSplitAncestors(ctx context.Context, personID uuid.UUID, maxDepth int) (*domain.SplitAncestorTree, error) {
	if maxDepth <= 0 {
		maxDepth = 10
	}

	rels, err := s.relRepo.ListByPerson(ctx, personID)
	if err != nil {
		return nil, err
	}

	var fatherID, motherID *uuid.UUID
	for _, r := range rels {
		if r.Type == domain.RelTypeParent && r.PersonA == personID {
			var meta struct {
				Role string `json:"role"`
			}
			if json.Unmarshal(r.Metadata, &meta) == nil {
				switch meta.Role {
				case "FATHER":
					id := r.PersonB
					fatherID = &id
				case "MOTHER":
					id := r.PersonB
					motherID = &id
				}
			}
		}
	}

	result := &domain.SplitAncestorTree{}

	if fatherID != nil {
		tree, err := s.getAncestorTreeForSide(ctx, personID, *fatherID, maxDepth)
		if err != nil {
			return nil, err
		}
		result.Paternal = tree
	}

	if motherID != nil {
		tree, err := s.getAncestorTreeForSide(ctx, personID, *motherID, maxDepth)
		if err != nil {
			return nil, err
		}
		result.Maternal = tree
	}

	return result, nil
}

func (s *service) getAncestorTree(ctx context.Context, personID uuid.UUID, maxDepth int) (*domain.AncestorTree, error) {
	maxDepth = clampDepth(maxDepth)

	ancestors, err := BFSAncestors(ctx, s.relRepo, s.personRepo, personID, maxDepth)
	if err != nil {
		return nil, err
	}

	rootPerson, err := s.personRepo.GetByID(ctx, personID)
	if err != nil {
		return nil, err
	}

	var allNodes []domain.GraphNode
	if rootPerson != nil {
		g := 0
		rootNode := personToGraphNode(rootPerson, &g)
		allNodes = append(allNodes, rootNode)
	}
	allNodes = append(allNodes, ancestors...)

	invertGenerations(allNodes)

	personIdSet, personIds := createPersonIdSet(allNodes)

	edges, err := s.fetchAndBuildEdges(ctx, personIds, personIdSet)
	if err != nil {
		return nil, err
	}

	return &domain.AncestorTree{
		RootPerson: personID,
		Ancestors:  allNodes,
		Edges:      edges,
		MaxDepth:   maxDepth,
	}, nil
}

func (s *service) getAncestorTreeForSide(ctx context.Context, personID, parentID uuid.UUID, maxDepth int) (*domain.AncestorTree, error) {
	ancestors, err := BFSAncestors(ctx, s.relRepo, s.personRepo, parentID, maxDepth-1)
	if err != nil {
		return nil, err
	}

	for i := range ancestors {
		if ancestors[i].Generation != nil {
			g := *ancestors[i].Generation + 1
			ancestors[i].Generation = &g
		}
	}

	parent, err := s.personRepo.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}

	rootPerson, err := s.personRepo.GetByID(ctx, personID)
	if err != nil {
		return nil, err
	}

	allNodes := []domain.GraphNode{}
	if rootPerson != nil {
		g := 0
		allNodes = append(allNodes, personToGraphNode(rootPerson, &g))
	}
	if parent != nil {
		g := 1
		allNodes = append(allNodes, personToGraphNode(parent, &g))
	}
	allNodes = append(allNodes, ancestors...)

	invertGenerations(allNodes)

	personIdSet, personIds := createPersonIdSet(allNodes)

	edges, err := s.fetchAndBuildEdges(ctx, personIds, personIdSet)
	if err != nil {
		return nil, err
	}

	return &domain.AncestorTree{
		RootPerson: personID,
		Ancestors:  allNodes,
		Edges:      edges,
		MaxDepth:   maxDepth,
	}, nil
}

func (s *service) GetDescendants(ctx context.Context, personID uuid.UUID, maxDepth int) (*domain.DescendantTree, error) {
	maxDepth = clampDepth(maxDepth)

	descendants, err := BFSDescendants(ctx, s.relRepo, s.personRepo, personID, maxDepth)
	if err != nil {
		return nil, err
	}

	rootPerson, err := s.personRepo.GetByID(ctx, personID)
	if err != nil {
		return nil, err
	}

	allNodes := make([]domain.GraphNode, 0, len(descendants)+1)
	if rootPerson != nil {
		g := 0
		allNodes = append(allNodes, personToGraphNode(rootPerson, &g))
	}
	allNodes = append(allNodes, descendants...)

	personIdSet, personIds := createPersonIdSet(allNodes)

	edges, err := s.fetchAndBuildEdges(ctx, personIds, personIdSet)
	if err != nil {
		return nil, err
	}

	return &domain.DescendantTree{
		RootPerson:  personID,
		Descendants: allNodes,
		Edges:       edges,
		MaxDepth:    maxDepth,
	}, nil
}

func (s *service) FindRelationshipPath(ctx context.Context, fromPersonID, toPersonID uuid.UUID, maxDepth int, locale string) (*domain.RelationshipPath, error) {
	if maxDepth <= 0 {
		maxDepth = 20
	}
	
	pathIDs, err := BFSShortestPath(ctx, s.relRepo, fromPersonID, toPersonID, maxDepth)
	if err != nil {
		return nil, err
	}

	if len(pathIDs) == 0 {
		return nil, nil
	}

	degree := len(pathIDs) - 1
	
	pathObj := &domain.RelationshipPath{
		FromPerson:   fromPersonID,
		ToPerson:     toPersonID,
		Path:         pathIDs,
		Relationship: domain.DerivedRelationType("custom"),
		Degree:       degree,
	}

	if s.narrative != nil {
		pathObj.Description = s.narrative.DescribeRelationship(ctx, pathObj, locale)
	} else {
		pathObj.Description = fmt.Sprintf("Related (degree %d)", degree)
		if degree == 0 {
			pathObj.Description = "Self"
		}
	}

	return pathObj, nil
}

func (s *service) InvalidateCache(ctx context.Context) error {
	return s.redis.Del(ctx, "family:graph").Err()
}

func (s *service) fetchAndBuildEdges(ctx context.Context, personIds []uuid.UUID, personIdSet map[uuid.UUID]bool) ([]domain.GraphEdge, error) {
	rels, err := s.relRepo.ListByPeople(ctx, personIds)
	if err != nil {
		return nil, err
	}
	return buildEdgesForNodes(rels, personIdSet), nil
}

func clampDepth(d int) int {
	if d <= 0 {
		return 5
	}
	if d > 20 {
		return 20
	}
	return d
}
