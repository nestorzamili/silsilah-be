package domain

import "github.com/google/uuid"

type GraphNode struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  *string   `json:"last_name,omitempty" db:"last_name"`
	Nickname  *string   `json:"nickname,omitempty" db:"nickname"`
	Gender    Gender    `json:"gender" db:"gender"`
	AvatarURL *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	IsAlive   bool      `json:"is_alive" db:"is_alive"`
	BirthYear *int      `json:"birth_year,omitempty" db:"birth_year"`
	DeathYear *int      `json:"death_year,omitempty" db:"death_year"`
	
	Generation *int `json:"generation,omitempty" db:"generation"`
	X          *float64 `json:"x,omitempty" db:"x"`
	Y          *float64 `json:"y,omitempty" db:"y"`
}

type GraphEdge struct {
	ID       uuid.UUID        `json:"id" db:"id"`
	Source   uuid.UUID        `json:"source" db:"person_a"`
	Target   uuid.UUID        `json:"target" db:"person_b"`
	Type     RelationshipType `json:"type" db:"type"`
	Metadata interface{}      `json:"metadata,omitempty" db:"metadata"`

	IsConsanguineous bool `json:"is_consanguineous,omitempty" db:"is_consanguineous"`
	SpouseOrder      *int `json:"spouse_order,omitempty" db:"spouse_order"` 
	ChildOrder       *int `json:"child_order,omitempty" db:"child_order"`  
}

type FamilyGroup struct {
	ID          string      `json:"id"`
	Parents     []uuid.UUID `json:"parents"`
	Children    []uuid.UUID `json:"children"`
	SpouseOrder int         `json:"spouse_order"`
}

type FamilyGraph struct {
	Nodes  []GraphNode    `json:"nodes"`
	Edges  []GraphEdge    `json:"edges"`
	Groups []FamilyGroup  `json:"groups,omitempty"`
	Stats  *GraphStats    `json:"stats,omitempty"`
}

type GraphStats struct {
	TotalPersons      int `json:"total_persons"`
	TotalRelationships int `json:"total_relationships"`
	MaxGeneration     int `json:"max_generation"`
	LivingPersons     int `json:"living_persons"`
	DeceasedPersons   int `json:"deceased_persons"`
}

type AncestorTree struct {
	RootPerson uuid.UUID   `json:"root_person"`
	Ancestors  []GraphNode `json:"ancestors"`
	Edges      []GraphEdge `json:"edges"`
	MaxDepth   int         `json:"max_depth"`
}

type SplitAncestorTree struct {
	Paternal *AncestorTree `json:"paternal"`
	Maternal *AncestorTree `json:"maternal"`
}

type DescendantTree struct {
	RootPerson  uuid.UUID   `json:"root_person"`
	Descendants []GraphNode `json:"descendants"`
	Edges       []GraphEdge `json:"edges"`
	MaxDepth    int         `json:"max_depth"`
}
