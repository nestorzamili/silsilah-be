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
