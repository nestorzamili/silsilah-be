package graph

import (
	"silsilah-keluarga/internal/domain"

	"github.com/google/uuid"
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

	node := domain.GraphNode{
		ID:        p.ID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Gender:    p.Gender,
		BirthYear: birthYear,
		DeathYear: deathYear,
		IsAlive:   p.IsAlive,
	}
	if generation != nil {
		node.Generation = generation
	}
	return node
}

func buildEdgesForNodes(rels []domain.Relationship, personIdSet map[uuid.UUID]bool) []domain.GraphEdge {
	var edges []domain.GraphEdge
	for _, r := range rels {
		if personIdSet[r.PersonA] && personIdSet[r.PersonB] {
			edges = append(edges, domain.GraphEdge{
				Source: r.PersonA,
				Target: r.PersonB,
				Type:   r.Type,
			})
		}
	}
	return edges
}

func createPersonIdSet(nodes []domain.GraphNode) (map[uuid.UUID]bool, []uuid.UUID) {
	personIdSet := make(map[uuid.UUID]bool)
	var personIds []uuid.UUID
	for _, n := range nodes {
		if !personIdSet[n.ID] {
			personIdSet[n.ID] = true
			personIds = append(personIds, n.ID)
		}
	}
	return personIdSet, personIds
}

func invertGenerations(nodes []domain.GraphNode) {
	for i := range nodes {
		if nodes[i].Generation != nil {
			g := -(*nodes[i].Generation)
			nodes[i].Generation = &g
		}
	}
}

func buildFamilyGroups(rels []domain.Relationship, personIdSet map[uuid.UUID]bool) []domain.FamilyGroup {
	// Placeholder implementation
	return []domain.FamilyGroup{}
}
