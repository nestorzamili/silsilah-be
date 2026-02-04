package narrative

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/pkg/i18n"
	"silsilah-keluarga/internal/repository"
)

type Service interface {
	DescribeRelationship(ctx context.Context, path *domain.RelationshipPath, locale string) string
}

type service struct {
	personRepo repository.PersonRepository
	relRepo    repository.RelationshipRepository
}

func NewService(personRepo repository.PersonRepository, relRepo repository.RelationshipRepository) Service {
	return &service{
		personRepo: personRepo,
		relRepo:    relRepo,
	}
}

func (s *service) DescribeRelationship(ctx context.Context, path *domain.RelationshipPath, locale string) string {
	if path == nil || len(path.Path) == 0 {
		return ""
	}

	var nameA, nameB string
	if pA, err := s.personRepo.GetByID(ctx, path.FromPerson); err == nil && pA != nil {
		nameA = pA.FirstName
	} else {
		nameA = "Unknown"
	}

	if pB, err := s.personRepo.GetByID(ctx, path.ToPerson); err == nil && pB != nil {
		nameB = pB.FirstName
	} else {
		nameB = "Unknown"
	}

	if path.Degree == 0 {
		template := i18n.Translate(locale, "SELF")
		result := strings.ReplaceAll(template, "{A}", nameA)
		result = strings.ReplaceAll(result, "{B}", nameB)
		return result
	}

	relKey := string(path.Relationship)

	if pA, err := s.personRepo.GetByID(ctx, path.FromPerson); err == nil && pA != nil {
		genderMap := map[domain.DerivedRelationType]map[domain.Gender]string{
			domain.DerivedChild: {
				domain.GenderMale:   "SON",
				domain.GenderFemale: "DAUGHTER",
			},
			domain.DerivedSibling: {
				domain.GenderMale:   "BROTHER",
				domain.GenderFemale: "SISTER",
			},
			domain.DerivedGrandparent: {
				domain.GenderMale:   "GRANDFATHER",
				domain.GenderFemale: "GRANDMOTHER",
			},
			domain.DerivedGrandchild: {
				domain.GenderMale:   "GRANDSON",
				domain.GenderFemale: "GRANDDAUGHTER",
			},
			domain.DerivedUncleAunt: {
				domain.GenderMale:   "UNCLE",
				domain.GenderFemale: "AUNT",
			},
			domain.DerivedNephewNiece: {
				domain.GenderMale:   "NEPHEW",
				domain.GenderFemale: "NIECE",
			},
			domain.DerivedCousin: {
				domain.GenderMale:   "COUSIN",
				domain.GenderFemale: "COUSIN",
			},
		}

		if genderRels, ok := genderMap[path.Relationship]; ok {
			if val, ok := genderRels[pA.Gender]; ok {
				relKey = val
			}
		}
	}

	if relKey == "" {
		relKey = "RELATED"
	}

	relName := i18n.Translate(locale, relKey)
	if relName == relKey {
		relKey = "RELATED"
	}

	var template string
	if relKey == "RELATED" {
		template = i18n.Translate(locale, "RELATED")
	} else {
		template = i18n.Translate(locale, "RELATED_FULL")
	}

	removedStr := ""
	if path.Relationship == domain.DerivedCousin {
	}

	lineageStr := ""
	if len(path.Path) >= 2 {
		rootID := path.Path[0]
		nextID := path.Path[1]

		rels, err := s.relRepo.ListByPeople(ctx, []uuid.UUID{rootID})
		if err == nil {
			for _, r := range rels {
				if r.Type == domain.RelTypeParent && r.PersonA == rootID && r.PersonB == nextID {
					if pNext, err := s.personRepo.GetByID(ctx, nextID); err == nil && pNext != nil {
						switch pNext.Gender {
							case domain.GenderMale:
								lineageStr = i18n.Translate(locale, "LINEAGE_PATERNAL")
							case domain.GenderFemale:
								lineageStr = i18n.Translate(locale, "LINEAGE_MATERNAL")
						}
					}
					break
				}
			}
		}
	}
	if lineageStr == "" {
		lineageStr = i18n.Translate(locale, "LINEAGE_MIXED")
	}

	result := strings.ReplaceAll(template, "{A}", nameA)
	result = strings.ReplaceAll(result, "{B}", nameB)
	result = strings.ReplaceAll(result, "{relationship}", relName)
	result = strings.ReplaceAll(result, "{degree}", fmt.Sprintf("%d", path.Degree))
	result = strings.ReplaceAll(result, "{removed_str}", removedStr)
	result = strings.ReplaceAll(result, "{lineage}", lineageStr)

	result = strings.ReplaceAll(result, "  ", " ")

	return result
}
