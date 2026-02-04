package dashboard

import (
	"context"
	"encoding/json"
	"time"

	"silsilah-keluarga/internal/repository"

	"github.com/redis/go-redis/v9"
)

type Stats struct {
	TotalPeople           int64      `json:"total_people"`
	TotalRelationships    int64      `json:"total_relationships"`
	LivingPeople          int64      `json:"living_people"`
	DeceasedPeople        int64      `json:"deceased_people"`
	OrphanNodes           int64      `json:"orphan_nodes"`
	PendingChangeRequests int64      `json:"pending_change_requests"`
	LastActivityAt        *time.Time `json:"last_activity_at"`
}

type Service interface {
	GetStats(ctx context.Context) (*Stats, error)
}

type service struct {
	personRepo repository.PersonRepository
	relRepo    repository.RelationshipRepository
	crRepo     repository.ChangeRequestRepository
	redis      *redis.Client
}

func NewService(personRepo repository.PersonRepository, relRepo repository.RelationshipRepository, crRepo repository.ChangeRequestRepository, redis *redis.Client) Service {
	return &service{
		personRepo: personRepo,
		relRepo:    relRepo,
		crRepo:     crRepo,
		redis:      redis,
	}
}

func (s *service) GetStats(ctx context.Context) (*Stats, error) {
	cacheKey := "dashboard:stats"

	if s.redis != nil {
		if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
			var stats Stats
			if json.Unmarshal([]byte(cached), &stats) == nil {
				return &stats, nil
			}
		}
	}

	totalPeople, err := s.personRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	livingPeople, err := s.personRepo.CountLiving(ctx)
	if err != nil {
		return nil, err
	}

	totalRels, err := s.relRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	orphanCount, err := s.personRepo.CountOrphans(ctx)
	if err != nil {
		return nil, err
	}

	pendingCRs, err := s.crRepo.CountPending(ctx)
	if err != nil {
		return nil, err
	}

	lastPersonActivity, _ := s.personRepo.GetLastActivityAt(ctx)
	lastRelActivity, _ := s.relRepo.GetLastActivityAt(ctx)

	var lastActivity *time.Time
	if lastPersonActivity != nil {
		lastActivity = lastPersonActivity
	}
	if lastRelActivity != nil {
		if lastActivity == nil || lastRelActivity.After(*lastActivity) {
			lastActivity = lastRelActivity
		}
	}

	stats := &Stats{
		TotalPeople:           totalPeople,
		TotalRelationships:    totalRels,
		LivingPeople:          livingPeople,
		DeceasedPeople:        totalPeople - livingPeople,
		OrphanNodes:           orphanCount,
		PendingChangeRequests: pendingCRs,
		LastActivityAt:        lastActivity,
	}

	if s.redis != nil {
		if statsJSON, err := json.Marshal(stats); err == nil {
			_ = s.redis.Set(ctx, cacheKey, statsJSON, 5*time.Minute).Err()
		}
	}

	return stats, nil
}
