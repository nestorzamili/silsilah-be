package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"silsilah-keluarga/internal/domain"
)

type CommonAncestorResult struct {
	AncestorID  uuid.UUID `db:"common_ancestor_id"`
	DepthFromA  int       `db:"depth_from_a"`
	DepthFromB  int       `db:"depth_from_b"`
	TotalDegree int       `db:"total_degree"`
}

type RelationshipRepository interface {
	Create(ctx context.Context, rel *domain.Relationship) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error)
	Update(ctx context.Context, rel *domain.Relationship) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error)
	ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error)
	GetByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error)
	GetAll(ctx context.Context) ([]domain.Relationship, error)
	GetAncestors(ctx context.Context, personID uuid.UUID, maxDepth int) ([]domain.GraphNode, error)
	GetDescendants(ctx context.Context, personID uuid.UUID, maxDepth int) ([]domain.GraphNode, error)
	FindCommonAncestors(ctx context.Context, personA, personB uuid.UUID) ([]CommonAncestorResult, error)
	CalculateConsanguinity(ctx context.Context, personA, personB uuid.UUID) (*domain.SpouseMetadata, error)
	GetSiblings(ctx context.Context, personID uuid.UUID) ([]domain.SiblingInfo, error)
	ListByPeople(ctx context.Context, personIDs []uuid.UUID) ([]domain.Relationship, error)
}

type relationshipRepository struct {
	db *sqlx.DB
}

func NewRelationshipRepository(db *sqlx.DB) RelationshipRepository {
	return &relationshipRepository{db: db}
}

func (r *relationshipRepository) Create(ctx context.Context, rel *domain.Relationship) error {
	query := `
		INSERT INTO relationships (id, person_a, person_b, type, metadata, spouse_order, child_order, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		rel.ID, rel.PersonA, rel.PersonB, rel.Type, rel.Metadata, rel.SpouseOrder, rel.ChildOrder, rel.CreatedBy,
	).Scan(&rel.CreatedAt, &rel.UpdatedAt)
}

func (r *relationshipRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Relationship, error) {
	var rel domain.Relationship
	query := `SELECT * FROM relationships WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &rel, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *relationshipRepository) Update(ctx context.Context, rel *domain.Relationship) error {
	query := `
		UPDATE relationships 
		SET metadata = $2, spouse_order = $3, child_order = $4, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	return r.db.QueryRowxContext(ctx, query,
		rel.ID, rel.Metadata, rel.SpouseOrder, rel.ChildOrder,
	).Scan(&rel.UpdatedAt)
}

func (r *relationshipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE relationships SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *relationshipRepository) List(ctx context.Context, relType *domain.RelationshipType) ([]domain.Relationship, error) {
	var relationships []domain.Relationship
	var err error

	if relType != nil {
		query := `SELECT * FROM relationships WHERE type = $1 AND deleted_at IS NULL`
		err = r.db.SelectContext(ctx, &relationships, query, *relType)
	} else {
		query := `SELECT * FROM relationships WHERE deleted_at IS NULL`
		err = r.db.SelectContext(ctx, &relationships, query)
	}

	return relationships, err
}

func (r *relationshipRepository) ListByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error) {
	query := `
		SELECT * FROM relationships 
		WHERE (person_a = $1 OR person_b = $1) AND deleted_at IS NULL`

	var relationships []domain.Relationship
	err := r.db.SelectContext(ctx, &relationships, query, personID)
	return relationships, err
}

func (r *relationshipRepository) GetByPerson(ctx context.Context, personID uuid.UUID) ([]domain.Relationship, error) {
	return r.ListByPerson(ctx, personID)
}

func (r *relationshipRepository) GetAll(ctx context.Context) ([]domain.Relationship, error) {
	query := `SELECT * FROM relationships WHERE deleted_at IS NULL`

	var relationships []domain.Relationship
	err := r.db.SelectContext(ctx, &relationships, query)
	return relationships, err
}

func (r *relationshipRepository) GetAncestors(ctx context.Context, personID uuid.UUID, maxDepth int) ([]domain.GraphNode, error) {
	query := `
		SELECT p.id, p.first_name, p.last_name, p.gender, p.avatar_url, p.is_alive,
			EXTRACT(YEAR FROM p.birth_date)::int as birth_year,
			EXTRACT(YEAR FROM p.death_date)::int as death_year,
			a.depth as generation
		FROM get_ancestors($1, $2) a
		INNER JOIN persons p ON a.ancestor_id = p.id
		WHERE p.deleted_at IS NULL`

	var nodes []domain.GraphNode
	err := r.db.SelectContext(ctx, &nodes, query, personID, maxDepth)
	return nodes, err
}

func (r *relationshipRepository) GetDescendants(ctx context.Context, personID uuid.UUID, maxDepth int) ([]domain.GraphNode, error) {
	query := `
		SELECT p.id, p.first_name, p.last_name, p.gender, p.avatar_url, p.is_alive,
			EXTRACT(YEAR FROM p.birth_date)::int as birth_year,
			EXTRACT(YEAR FROM p.death_date)::int as death_year,
			d.depth as generation
		FROM get_descendants($1, $2) d
		INNER JOIN persons p ON d.descendant_id = p.id
		WHERE p.deleted_at IS NULL`

	var nodes []domain.GraphNode
	err := r.db.SelectContext(ctx, &nodes, query, personID, maxDepth)
	return nodes, err
}

func (r *relationshipRepository) FindCommonAncestors(ctx context.Context, personA, personB uuid.UUID) ([]CommonAncestorResult, error) {
	query := `SELECT * FROM find_common_ancestors($1, $2)`

	var results []CommonAncestorResult
	err := r.db.SelectContext(ctx, &results, query, personA, personB)
	return results, err
}

func (r *relationshipRepository) CalculateConsanguinity(ctx context.Context, personA, personB uuid.UUID) (*domain.SpouseMetadata, error) {
	query := `SELECT * FROM calculate_consanguinity($1, $2)`

	var result struct {
		IsConsanguineous        bool        `db:"is_consanguineous"`
		Degree                  *int        `db:"degree"`
		ClosestCommonAncestors  []uuid.UUID `db:"closest_common_ancestors"`
		RelationshipDescription string      `db:"relationship_description"`
	}

	err := r.db.GetContext(ctx, &result, query, personA, personB)
	if err != nil {
		return nil, err
	}

	metadata := &domain.SpouseMetadata{
		IsConsanguineous:    result.IsConsanguineous,
		ConsanguinityDegree: result.Degree,
		CommonAncestors:     result.ClosestCommonAncestors,
	}

	return metadata, nil
}

func (r *relationshipRepository) GetSiblings(ctx context.Context, personID uuid.UUID) ([]domain.SiblingInfo, error) {
	query := `
		WITH my_parents AS (
			SELECT person_b AS parent_id FROM relationships
			WHERE person_a = $1 AND type = 'PARENT' AND deleted_at IS NULL
		),
		siblings_with_shared_parents AS (
			SELECT 
				r.person_a AS sibling_id,
				COUNT(DISTINCT r.person_b) AS shared_parent_count,
				(SELECT COUNT(*) FROM my_parents) AS my_parent_count,
				COUNT(DISTINCT r2.person_b) AS sibling_parent_count
			FROM relationships r
			INNER JOIN my_parents mp ON r.person_b = mp.parent_id
			LEFT JOIN relationships r2 ON r2.person_a = r.person_a AND r2.type = 'PARENT' AND r2.deleted_at IS NULL
			WHERE r.type = 'PARENT' AND r.deleted_at IS NULL AND r.person_a != $1
			GROUP BY r.person_a
		)
		SELECT 
			p.id, p.first_name, p.last_name, p.nickname, p.gender, 
			p.birth_date, p.birth_place, p.death_date, p.death_place,
			p.bio, p.avatar_url, p.is_alive, p.created_by, p.created_at, p.updated_at,
			CASE 
				WHEN s.shared_parent_count = s.my_parent_count AND s.shared_parent_count = s.sibling_parent_count 
				THEN 'FULL' 
				ELSE 'HALF' 
			END AS sibling_type
		FROM siblings_with_shared_parents s
		INNER JOIN persons p ON s.sibling_id = p.id
		WHERE p.deleted_at IS NULL`

	rows, err := r.db.QueryxContext(ctx, query, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var siblings []domain.SiblingInfo
	for rows.Next() {
		var person domain.Person
		var siblingType string
		err := rows.Scan(
			&person.ID, &person.FirstName, &person.LastName, &person.Nickname, &person.Gender,
			&person.BirthDate, &person.BirthPlace, &person.DeathDate, &person.DeathPlace,
			&person.Bio, &person.AvatarURL, &person.IsAlive, &person.CreatedBy, &person.CreatedAt, &person.UpdatedAt,
			&siblingType,
		)
		if err != nil {
			return nil, err
		}
		siblings = append(siblings, domain.SiblingInfo{
			Person:      person,
			SiblingType: siblingType,
		})
	}

	return siblings, nil
}

func (r *relationshipRepository) ListByPeople(ctx context.Context, personIDs []uuid.UUID) ([]domain.Relationship, error) {
	if len(personIDs) == 0 {
		return []domain.Relationship{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT * FROM relationships 
		WHERE person_a IN (?) AND person_b IN (?) AND deleted_at IS NULL`, 
		personIDs, personIDs)
	
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var relationships []domain.Relationship
	err = r.db.SelectContext(ctx, &relationships, query, args...)
	return relationships, err
}
