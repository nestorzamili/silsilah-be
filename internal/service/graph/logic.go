package graph

import (
	"context"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"

	"github.com/google/uuid"
)

func BFSAncestors(ctx context.Context, relRepo repository.RelationshipRepository, personRepo repository.PersonRepository, startID uuid.UUID, maxDepth int) ([]domain.GraphNode, error) {
	visited := make(map[uuid.UUID]int) 
	visited[startID] = 0
	currentLevel := []uuid.UUID{startID}

	for depth := 0; depth < maxDepth; depth++ {
		if len(currentLevel) == 0 {
			break
		}

		rels, err := relRepo.ListByPeople(ctx, currentLevel)
		if err != nil {
			return nil, err
		}

		nextLevel := []uuid.UUID{}
		for _, r := range rels {
			if r.Type == domain.RelTypeParent {
				isChild := false
				for _, id := range currentLevel {
					if r.PersonA == id {
						isChild = true
						break
					}
				}

				if isChild {
					parentID := r.PersonB
					if _, seen := visited[parentID]; !seen {
						visited[parentID] = depth + 1
						nextLevel = append(nextLevel, parentID)
					}
				}
			}
		}
		currentLevel = nextLevel
	}

	ancestorIDs := make([]uuid.UUID, 0, len(visited))
	for id, d := range visited {
		if id != startID && d > 0 {
			ancestorIDs = append(ancestorIDs, id)
		}
	}

	if len(ancestorIDs) == 0 {
		return []domain.GraphNode{}, nil
	}

	persons, err := personRepo.GetByIDs(ctx, ancestorIDs)
	if err != nil {
		return nil, err
	}

	nodes := make([]domain.GraphNode, len(persons))
	for i, p := range persons {
		gen := visited[p.ID]
		nodes[i] = personToGraphNode(&p, &gen)
	}

	return nodes, nil
}

func BFSDescendants(ctx context.Context, relRepo repository.RelationshipRepository, personRepo repository.PersonRepository, startID uuid.UUID, maxDepth int) ([]domain.GraphNode, error) {
	visited := make(map[uuid.UUID]int)
	visited[startID] = 0
	currentLevel := []uuid.UUID{startID}

	for depth := 0; depth < maxDepth; depth++ {
		if len(currentLevel) == 0 {
			break
		}

		rels, err := relRepo.ListByPeople(ctx, currentLevel)
		if err != nil {
			return nil, err
		}

		nextLevel := []uuid.UUID{}
		for _, r := range rels {
			if r.Type == domain.RelTypeParent {
				isParent := false
				for _, id := range currentLevel {
					if r.PersonB == id {
						isParent = true
						break
					}
				}

				if isParent {
					childID := r.PersonA
					if _, seen := visited[childID]; !seen {
						visited[childID] = depth + 1
						nextLevel = append(nextLevel, childID)
					}
				}
			}
		}
		currentLevel = nextLevel
	}

	descendantIDs := make([]uuid.UUID, 0, len(visited))
	for id, d := range visited {
		if id != startID && d > 0 {
			descendantIDs = append(descendantIDs, id)
		}
	}

	if len(descendantIDs) == 0 {
		return []domain.GraphNode{}, nil
	}

	persons, err := personRepo.GetByIDs(ctx, descendantIDs)
	if err != nil {
		return nil, err
	}

	nodes := make([]domain.GraphNode, len(persons))
	for i, p := range persons {
		gen := visited[p.ID]
		nodes[i] = personToGraphNode(&p, &gen)
	}

	return nodes, nil
}

func BFSShortestPath(ctx context.Context, relRepo repository.RelationshipRepository, startID, targetID uuid.UUID, maxDepth int) ([]uuid.UUID, error) {
	if startID == targetID {
		return []uuid.UUID{startID}, nil
	}

	queue := []uuid.UUID{startID}

	visited := make(map[uuid.UUID]bool)
	visited[startID] = true

	pathParent := make(map[uuid.UUID]uuid.UUID)

	depths := make(map[uuid.UUID]int)
	depths[startID] = 0

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		currentDepth := depths[currentID]
		if currentDepth >= maxDepth {
			continue
		}

		if currentID == targetID {
			break // Found!
		}

		rels, err := relRepo.ListByPerson(ctx, currentID)
		if err != nil {
			return nil, err
		}

		for _, r := range rels {
			var neighborID uuid.UUID

			if r.PersonA == currentID {
				neighborID = r.PersonB
			} else {
				neighborID = r.PersonA
			}

			if !visited[neighborID] {
				visited[neighborID] = true
				pathParent[neighborID] = currentID
				depths[neighborID] = currentDepth + 1
				queue = append(queue, neighborID)

				if neighborID == targetID {
					goto Found
				}
			}
		}
	}

	if !visited[targetID] {
		return nil, nil 
	}

Found:
	path := []uuid.UUID{}
	curr := targetID
	for curr != uuid.Nil {
		path = append([]uuid.UUID{curr}, path...) 
		if curr == startID {
			break
		}
		var ok bool
		curr, ok = pathParent[curr]
		if !ok {
			break 
		}
	}

	return path, nil
}

func GetSiblingsLogic(ctx context.Context, relRepo repository.RelationshipRepository, personRepo repository.PersonRepository, personID uuid.UUID) ([]domain.SiblingInfo, error) {
	rels, err := relRepo.ListByPerson(ctx, personID)
	if err != nil {
		return nil, err
	}

	parentIDs := make([]uuid.UUID, 0)
	for _, r := range rels {
		if r.Type == domain.RelTypeParent && r.PersonA == personID {
			parentIDs = append(parentIDs, r.PersonB)
		}
	}

	if len(parentIDs) == 0 {
		return []domain.SiblingInfo{}, nil
	}

	parentRels, err := relRepo.ListByPeople(ctx, parentIDs)
	if err != nil {
		return nil, err
	}

	siblingMap := make(map[uuid.UUID]int) 

	for _, r := range parentRels {
		if r.Type == domain.RelTypeParent && r.PersonA != personID {
			isMyParent := false
			for _, pid := range parentIDs {
				if r.PersonB == pid {
					isMyParent = true
					break
				}
			}

			if isMyParent {
				siblingMap[r.PersonA]++
			}
		}
	}

	if len(siblingMap) == 0 {
		return []domain.SiblingInfo{}, nil
	}

	siblingIDs := make([]uuid.UUID, 0, len(siblingMap))
	for id := range siblingMap {
		siblingIDs = append(siblingIDs, id)
	}

	siblings, err := personRepo.GetByIDs(ctx, siblingIDs)
	if err != nil {
		return nil, err
	}

	var result []domain.SiblingInfo
	myParentCount := len(parentIDs)

	for _, p := range siblings {
		sharedCount := siblingMap[p.ID]

		sibRels, err := relRepo.ListByPeople(ctx, []uuid.UUID{p.ID})
		if err == nil {
			sibParentCount := 0
			for _, r := range sibRels {
				if r.Type == domain.RelTypeParent && r.PersonA == p.ID {
					sibParentCount++
				}
			}

			typeStr := "HALF"
			if sharedCount == myParentCount && sharedCount == sibParentCount {
				typeStr = "FULL"
			}

			result = append(result, domain.SiblingInfo{
				Person:      p,
				SiblingType: typeStr,
			})
		}
	}

	return result, nil
}
