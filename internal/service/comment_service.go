package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

type CommentService interface {
	Create(ctx context.Context, personID, userID uuid.UUID, input domain.CreateCommentInput) (*domain.Comment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input domain.UpdateCommentInput) (*domain.Comment, error)
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	ListByPerson(ctx context.Context, personID uuid.UUID, params domain.PaginationParams) (domain.PaginatedResponse[domain.Comment], error)
}

type commentService struct {
	commentRepo repository.CommentRepository
	redis       *redis.Client
}

func NewCommentService(commentRepo repository.CommentRepository, redis *redis.Client) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		redis:       redis,
	}
}

func (s *commentService) Create(ctx context.Context, personID, userID uuid.UUID, input domain.CreateCommentInput) (*domain.Comment, error) {
	comment := &domain.Comment{
		ID:       uuid.New(),
		PersonID: personID,
		UserID:   userID,
		Content:  input.Content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	if s.redis != nil {
		cachePattern := fmt.Sprintf("comments:%s:*", personID)
		keys, _ := s.redis.Keys(ctx, cachePattern).Result()
		if len(keys) > 0 {
			_ = s.redis.Del(ctx, keys...).Err()
		}
	}

	return comment, nil
}

func (s *commentService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	return s.commentRepo.GetByID(ctx, id)
}

func (s *commentService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input domain.UpdateCommentInput) (*domain.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, errors.New("comment not found")
	}

	if comment.UserID != userID {
		return nil, errors.New("insufficient permissions to edit this comment")
	}

	comment.Content = input.Content

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	if s.redis != nil {
		cachePattern := fmt.Sprintf("comments:%s:*", comment.PersonID)
		keys, _ := s.redis.Keys(ctx, cachePattern).Result()
		if len(keys) > 0 {
			_ = s.redis.Del(ctx, keys...).Err()
		}
	}

	return comment, nil
}

func (s *commentService) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if comment == nil {
		return errors.New("comment not found")
	}

	if comment.UserID != userID {
		return errors.New("insufficient permissions to delete this comment")
	}
	if s.redis != nil {
		cachePattern := fmt.Sprintf("comments:%s:*", comment.PersonID)
		keys, _ := s.redis.Keys(ctx, cachePattern).Result()
		if len(keys) > 0 {
			_ = s.redis.Del(ctx, keys...).Err()
		}
	}

	return s.commentRepo.Delete(ctx, id)
}

func (s *commentService) ListByPerson(ctx context.Context, personID uuid.UUID, params domain.PaginationParams) (domain.PaginatedResponse[domain.Comment], error) {
	cacheKey := fmt.Sprintf("comments:%s:page:%d:size:%d", personID, params.Page, params.PageSize)

	if s.redis != nil {
		if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
			var result domain.PaginatedResponse[domain.Comment]
			if json.Unmarshal([]byte(cached), &result) == nil {
				return result, nil
			}
		}
	}

	comments, total, err := s.commentRepo.ListByPerson(ctx, personID, params)
	if err != nil {
		return domain.PaginatedResponse[domain.Comment]{}, err
	}

	result := domain.NewPaginatedResponse(comments, params.Page, params.PageSize, total)

	if s.redis != nil {
		if resultJSON, err := json.Marshal(result); err == nil {
			_ = s.redis.Set(ctx, cacheKey, resultJSON, 5*time.Minute).Err()
		}
	}

	return result, nil
}
