package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
)

func personToGraphNode(p *domain.Person, generation *int) domain.GraphNode {
	var birthYear, deathYear *int
	if p.BirthDate != nil {
		y := p.BirthDate.Year()
		birthYear = &y
	}
	if p.DeathDate != nil {
		y := p.DeathDate.Year()
		deathYear = &y
	}

	return domain.GraphNode{
		ID:         p.ID,
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Nickname:   p.Nickname,
		Gender:     p.Gender,
		AvatarURL:  p.AvatarURL,
		IsAlive:    p.IsAlive,
		BirthYear:  birthYear,
		DeathYear:  deathYear,
		Generation: generation,
	}
}

func invertGenerations(nodes []domain.GraphNode) {
	maxGen := 0
	for _, n := range nodes {
		if n.Generation != nil && *n.Generation > maxGen {
			maxGen = *n.Generation
		}
	}
	
	for i := range nodes {
		if nodes[i].Generation != nil {
			inverted := maxGen - *nodes[i].Generation
			nodes[i].Generation = &inverted
		}
	}
}

func buildEdgesForNodes(relationships []domain.Relationship, personIdSet map[uuid.UUID]bool) []domain.GraphEdge {
	edges := make([]domain.GraphEdge, 0, len(relationships))
	
	for _, r := range relationships {
		if !personIdSet[r.PersonA] || !personIdSet[r.PersonB] {
			continue
		}

		var isConsanguineous bool
		if r.Type == domain.RelTypeSpouse && r.Metadata != nil {
			var meta domain.SpouseMetadata
			if json.Unmarshal(r.Metadata, &meta) == nil {
				isConsanguineous = meta.IsConsanguineous
			}
		}

		edges = append(edges, domain.GraphEdge{
			ID:               r.ID,
			Source:           r.PersonA,
			Target:           r.PersonB,
			Type:             r.Type,
			IsConsanguineous: isConsanguineous,
			SpouseOrder:      r.SpouseOrder,
			ChildOrder:       r.ChildOrder,
		})
	}

	return edges
}

func createPersonIdSet(nodes []domain.GraphNode) (map[uuid.UUID]bool, []uuid.UUID) {
	personIdSet := make(map[uuid.UUID]bool, len(nodes))
	personIds := make([]uuid.UUID, 0, len(nodes))
	
	for _, n := range nodes {
		personIdSet[n.ID] = true
		personIds = append(personIds, n.ID)
	}
	
	return personIdSet, personIds
}

func clampDepth(depth int) int {
	if depth <= 0 {
		return 10
	}
	if depth > 20 {
		return 20
	}
	return depth
}

func (s *graphService) fetchAndBuildEdges(ctx context.Context, personIds []uuid.UUID, personIdSet map[uuid.UUID]bool) ([]domain.GraphEdge, error) {
	allRels, err := s.relRepo.ListByPeople(ctx, personIds)
	if err != nil {
		return nil, err
	}
	return buildEdgesForNodes(allRels, personIdSet), nil
}

func buildFamilyGroups(relationships []domain.Relationship, personIdSet map[uuid.UUID]bool) []domain.FamilyGroup {
	childToParents := make(map[uuid.UUID][]uuid.UUID)
	spouseOrderMap := make(map[string]int)
	childOrderMap := make(map[uuid.UUID]int)

	for _, r := range relationships {
		if r.Type != domain.RelTypeParent {
			continue
		}
		if !personIdSet[r.PersonA] || !personIdSet[r.PersonB] {
			continue
		}

		childID := r.PersonA
		parentID := r.PersonB

		childToParents[childID] = append(childToParents[childID], parentID)

		if r.ChildOrder != nil {
			childOrderMap[childID] = *r.ChildOrder
		}
	}

	for _, r := range relationships {
		if r.Type != domain.RelTypeSpouse {
			continue
		}
		if !personIdSet[r.PersonA] || !personIdSet[r.PersonB] {
			continue
		}

		key := makeParentPairKey(r.PersonA, r.PersonB)
		if r.SpouseOrder != nil {
			spouseOrderMap[key] = *r.SpouseOrder
		} else {
			if _, exists := spouseOrderMap[key]; !exists {
				spouseOrderMap[key] = 1
			}
		}
	}

	parentPairToChildren := make(map[string][]uuid.UUID)
	parentPairToParents := make(map[string][]uuid.UUID)

	for childID, parents := range childToParents {
		key := makeParentPairKeyFromSlice(parents)
		parentPairToChildren[key] = append(parentPairToChildren[key], childID)
		if _, exists := parentPairToParents[key]; !exists {
			parentPairToParents[key] = parents
		}
	}

	var groups []domain.FamilyGroup
	processedPairs := make(map[string]bool)

	for key, children := range parentPairToChildren {
		if processedPairs[key] {
			continue
		}
		processedPairs[key] = true

		parents := parentPairToParents[key]

		sortedChildren := make([]uuid.UUID, len(children))
		copy(sortedChildren, children)
		sortChildrenByOrder(sortedChildren, childOrderMap)

		spouseOrder := 1
		if len(parents) == 2 {
			pairKey := makeParentPairKey(parents[0], parents[1])
			if order, exists := spouseOrderMap[pairKey]; exists {
				spouseOrder = order
			}
		}

		groups = append(groups, domain.FamilyGroup{
			ID:          "family-" + key,
			Parents:     parents,
			Children:    sortedChildren,
			SpouseOrder: spouseOrder,
		})
	}

	sortFamilyGroupsBySpouseOrder(groups)

	return groups
}

func makeParentPairKey(a, b uuid.UUID) string {
	if a.String() < b.String() {
		return a.String() + "-" + b.String()
	}
	return b.String() + "-" + a.String()
}

func makeParentPairKeyFromSlice(parents []uuid.UUID) string {
	if len(parents) == 0 {
		return ""
	}
	if len(parents) == 1 {
		return parents[0].String()
	}
	return makeParentPairKey(parents[0], parents[1])
}

func sortChildrenByOrder(children []uuid.UUID, orderMap map[uuid.UUID]int) {
	for i := 0; i < len(children)-1; i++ {
		for j := i + 1; j < len(children); j++ {
			orderI := orderMap[children[i]]
			orderJ := orderMap[children[j]]
			if orderI == 0 {
				orderI = 999
			}
			if orderJ == 0 {
				orderJ = 999
			}
			if orderI > orderJ {
				children[i], children[j] = children[j], children[i]
			}
		}
	}
}

func sortFamilyGroupsBySpouseOrder(groups []domain.FamilyGroup) {
	for i := 0; i < len(groups)-1; i++ {
		for j := i + 1; j < len(groups); j++ {
			if groups[i].SpouseOrder > groups[j].SpouseOrder {
				groups[i], groups[j] = groups[j], groups[i]
			}
		}
	}
}
