package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

type NotificationService interface {
	Create(ctx context.Context, notif *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	List(ctx context.Context, userID uuid.UUID, unreadOnly bool, params domain.PaginationParams) (domain.PaginatedResponse[domain.Notification], error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)

	NotifyChangeRequest(ctx context.Context, changeRequestID uuid.UUID, requesterID uuid.UUID) error
	NotifyChangeApproved(ctx context.Context, changeRequestID uuid.UUID, reviewerID uuid.UUID) error
	NotifyChangeRejected(ctx context.Context, changeRequestID uuid.UUID, reviewerID uuid.UUID) error
	NotifyNewComment(ctx context.Context, commentID uuid.UUID, authorID uuid.UUID) error
}

type notificationService struct {
	notifRepo repository.NotificationRepository
	userRepo  repository.UserRepository
	crRepo    repository.ChangeRequestRepository
	commentRepo repository.CommentRepository
	personRepo repository.PersonRepository
}

func NewNotificationService(
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	crRepo repository.ChangeRequestRepository,
	commentRepo repository.CommentRepository,
	personRepo repository.PersonRepository,
) NotificationService {
	return &notificationService{
		notifRepo:   notifRepo,
		userRepo:    userRepo,
		crRepo:      crRepo,
		commentRepo: commentRepo,
		personRepo:  personRepo,
	}
}

func (s *notificationService) Create(ctx context.Context, notif *domain.Notification) error {
	return s.notifRepo.Create(ctx, notif)
}

func (s *notificationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return s.notifRepo.GetByID(ctx, id)
}

func (s *notificationService) List(ctx context.Context, userID uuid.UUID, unreadOnly bool, params domain.PaginationParams) (domain.PaginatedResponse[domain.Notification], error) {
	notifications, total, err := s.notifRepo.ListByUser(ctx, userID, unreadOnly, params)
	if err != nil {
		return domain.PaginatedResponse[domain.Notification]{}, err
	}

	return domain.NewPaginatedResponse(notifications, params.Page, params.PageSize, total), nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.notifRepo.MarkAsRead(ctx, id)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notifRepo.CountUnread(ctx, userID)
}


func (s *notificationService) NotifyChangeRequest(ctx context.Context, crID uuid.UUID, requesterID uuid.UUID) error {
	cr, err := s.crRepo.GetByID(ctx, crID)
	if err != nil {
		return fmt.Errorf("failed to get change request: %w", err)
	}

	requester, err := s.userRepo.GetByID(ctx, requesterID)
	if err != nil {
		return fmt.Errorf("failed to get requester: %w", err)
	}

	recipients, err := s.userRepo.GetByRoles(ctx, []domain.UserRole{domain.RoleEditor, domain.RoleDeveloper})
	if err != nil {
		return fmt.Errorf("failed to get reviewers: %w", err)
	}

	for _, user := range recipients {
		if user.ID == requesterID {
			continue
		}

		dataMap := map[string]string{
			"change_request_id": crID.String(),
			"entity_type":       string(cr.EntityType),
		}
		if cr.EntityID != nil {
			dataMap["entity_id"] = cr.EntityID.String()
		}
		data, _ := json.Marshal(dataMap)

		notif := &domain.Notification{
			ID:      uuid.New(),
			UserID:  user.ID,
			Type:    domain.NotifChangeRequest,
			Title:   "Permintaan Perubahan Baru",
			Message: fmt.Sprintf("%s mengajukan %s %s", requester.FullName, getActionLabel(string(cr.Action)), getEntityLabel(string(cr.EntityType))),
			Data:    json.RawMessage(data),
		}

		if err := s.notifRepo.Create(ctx, notif); err != nil {
			fmt.Printf("Failed to create notification for user %s: %v\n", user.ID, err)
		}
	}

	return nil
}

func (s *notificationService) NotifyChangeApproved(ctx context.Context, crID uuid.UUID, reviewerID uuid.UUID) error {
	cr, err := s.crRepo.GetByID(ctx, crID)
	if err != nil {
		return fmt.Errorf("failed to get change request: %w", err)
	}

	reviewer, err := s.userRepo.GetByID(ctx, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to get reviewer: %w", err)
	}

	if cr.RequestedBy == reviewerID {
		return nil
	}

	dataMap := map[string]string{
		"change_request_id": crID.String(),
		"entity_type":       string(cr.EntityType),
	}
	if cr.EntityID != nil {
		dataMap["entity_id"] = cr.EntityID.String()
	}
	data, _ := json.Marshal(dataMap)

	notif := &domain.Notification{
		ID:      uuid.New(),
		UserID:  cr.RequestedBy,
		Type:    domain.NotifChangeApproved,
		Title:   "Permintaan Disetujui",
		Message: fmt.Sprintf("%s menyetujui permintaan %s %s Anda", reviewer.FullName, getActionLabel(string(cr.Action)), getEntityLabel(string(cr.EntityType))),
		Data:    json.RawMessage(data),
	}

	if err := s.notifRepo.Create(ctx, notif); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (s *notificationService) NotifyChangeRejected(ctx context.Context, crID uuid.UUID, reviewerID uuid.UUID) error {
	cr, err := s.crRepo.GetByID(ctx, crID)
	if err != nil {
		return fmt.Errorf("failed to get change request: %w", err)
	}

	reviewer, err := s.userRepo.GetByID(ctx, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to get reviewer: %w", err)
	}

	if cr.RequestedBy == reviewerID {
		return nil
	}

	dataMap := map[string]string{
		"change_request_id": crID.String(),
		"entity_type":       string(cr.EntityType),
	}
	if cr.EntityID != nil {
		dataMap["entity_id"] = cr.EntityID.String()
	}
	data, _ := json.Marshal(dataMap)

	notif := &domain.Notification{
		ID:      uuid.New(),
		UserID:  cr.RequestedBy,
		Type:    domain.NotifChangeRejected,
		Title:   "Permintaan Ditolak",
		Message: fmt.Sprintf("%s menolak permintaan %s %s Anda", reviewer.FullName, getActionLabel(string(cr.Action)), getEntityLabel(string(cr.EntityType))),
		Data:    json.RawMessage(data),
	}

	if err := s.notifRepo.Create(ctx, notif); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (s *notificationService) NotifyNewComment(ctx context.Context, commentID uuid.UUID, authorID uuid.UUID) error {
	// TODO: Implement comment notifications
	return nil
}

// Helper functions

func getActionLabel(action string) string {
	switch action {
	case "CREATE":
		return "tambah"
	case "UPDATE":
		return "ubah"
	case "DELETE":
		return "hapus"
	default:
		return action
	}
}

func getEntityLabel(entityType string) string {
	switch entityType {
	case "PERSON":
		return "anggota"
	case "RELATIONSHIP":
		return "hubungan"
	case "MEDIA":
		return "media"
	default:
		return entityType
	}
}
