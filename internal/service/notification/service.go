package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service/email"
)

type Service interface {
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
	NotifyPersonAdded(ctx context.Context, personID uuid.UUID, addedBy uuid.UUID) error
	NotifyRelationshipAdded(ctx context.Context, relationshipID uuid.UUID, addedBy uuid.UUID) error
}

type service struct {
	notifRepo        repository.NotificationRepository
	userRepo         repository.UserRepository
	crRepo           repository.ChangeRequestRepository
	commentRepo      repository.CommentRepository
	personRepo       repository.PersonRepository
	relationshipRepo repository.RelationshipRepository
	emailSvc         email.Service
}

func NewService(
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	crRepo repository.ChangeRequestRepository,
	commentRepo repository.CommentRepository,
	personRepo repository.PersonRepository,
	relationshipRepo repository.RelationshipRepository,
	emailSvc email.Service,
) Service {
	return &service{
		notifRepo:        notifRepo,
		userRepo:         userRepo,
		crRepo:           crRepo,
		commentRepo:      commentRepo,
		personRepo:       personRepo,
		relationshipRepo: relationshipRepo,
		emailSvc:         emailSvc,
	}
}

func (s *service) Create(ctx context.Context, notif *domain.Notification) error {
	return s.notifRepo.Create(ctx, notif)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return s.notifRepo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, userID uuid.UUID, unreadOnly bool, params domain.PaginationParams) (domain.PaginatedResponse[domain.Notification], error) {
	notifications, total, err := s.notifRepo.ListByUser(ctx, userID, unreadOnly, params)
	if err != nil {
		return domain.PaginatedResponse[domain.Notification]{}, err
	}

	return domain.NewPaginatedResponse(notifications, params.Page, params.PageSize, total), nil
}

func (s *service) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.notifRepo.MarkAsRead(ctx, id)
}

func (s *service) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

func (s *service) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notifRepo.CountUnread(ctx, userID)
}

func (s *service) NotifyChangeRequest(ctx context.Context, crID uuid.UUID, requesterID uuid.UUID) error {
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

		// Send email
		if s.emailSvc != nil && user.Email != "" {
			go func(toEmail, recipientName, requesterName, action, entityType string) {
				// Create a new context for async operation
				ctx := context.Background()
				_ = s.emailSvc.SendChangeRequestEmail(ctx, toEmail, recipientName, requesterName, action, entityType)
			}(user.Email, user.FullName, requester.FullName, getActionLabel(string(cr.Action)), getEntityLabel(string(cr.EntityType)))
		}
	}

	return nil
}

func (s *service) NotifyChangeApproved(ctx context.Context, crID uuid.UUID, reviewerID uuid.UUID) error {
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

	// Send email
	if s.emailSvc != nil {
		if requester, err := s.userRepo.GetByID(ctx, cr.RequestedBy); err == nil && requester.Email != "" {
			go func(toEmail, recipientName, action, entityType, status, reviewerName string) {
				ctx := context.Background()
				_ = s.emailSvc.SendChangeStatusEmail(ctx, toEmail, recipientName, action, entityType, status, reviewerName)
			}(requester.Email, requester.FullName, getActionLabel(string(cr.Action)), getEntityLabel(string(cr.EntityType)), "DISETUJUI", reviewer.FullName)
		}
	}

	return nil
}

func (s *service) NotifyChangeRejected(ctx context.Context, crID uuid.UUID, reviewerID uuid.UUID) error {
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

	// Send email
	if s.emailSvc != nil {
		if requester, err := s.userRepo.GetByID(ctx, cr.RequestedBy); err == nil && requester.Email != "" {
			go func(toEmail, recipientName, action, entityType, status, reviewerName string) {
				ctx := context.Background()
				_ = s.emailSvc.SendChangeStatusEmail(ctx, toEmail, recipientName, action, entityType, status, reviewerName)
			}(requester.Email, requester.FullName, getActionLabel(string(cr.Action)), getEntityLabel(string(cr.EntityType)), "DITOLAK", reviewer.FullName)
		}
	}

	return nil
}

func (s *service) NotifyNewComment(ctx context.Context, commentID uuid.UUID, authorID uuid.UUID) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return fmt.Errorf("failed to get author: %w", err)
	}

	person, err := s.personRepo.GetByID(ctx, comment.PersonID)
	if err != nil {
		return fmt.Errorf("failed to get person: %w", err)
	}

	recipients, err := s.userRepo.GetByRoles(ctx, []domain.UserRole{domain.RoleEditor, domain.RoleDeveloper})
	if err != nil {
		return err
	}

	for _, user := range recipients {
		if user.ID == authorID {
			continue
		}

		dataMap := map[string]string{
			"comment_id": commentID.String(),
			"person_id":  comment.PersonID.String(),
		}
		data, _ := json.Marshal(dataMap)

		notif := &domain.Notification{
			ID:      uuid.New(),
			UserID:  user.ID,
			Type:    domain.NotifNewComment,
			Title:   "Komentar Baru",
			Message: fmt.Sprintf("%s mengomentari profil %s", author.FullName, person.FirstName),
			Data:    json.RawMessage(data),
		}

		_ = s.notifRepo.Create(ctx, notif)
	}

	return nil
}

func (s *service) NotifyPersonAdded(ctx context.Context, personID uuid.UUID, addedBy uuid.UUID) error {
	person, err := s.personRepo.GetByID(ctx, personID)
	if err != nil {
		return fmt.Errorf("failed to get person: %w", err)
	}

	adder, err := s.userRepo.GetByID(ctx, addedBy)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	recipients, err := s.userRepo.GetByRoles(ctx, []domain.UserRole{domain.RoleEditor, domain.RoleDeveloper})
	if err != nil {
		return err
	}

	for _, user := range recipients {
		if user.ID == addedBy {
			continue
		}

		dataMap := map[string]string{
			"person_id": personID.String(),
		}
		data, _ := json.Marshal(dataMap)

		notif := &domain.Notification{
			ID:      uuid.New(),
			UserID:  user.ID,
			Type:    domain.NotifPersonAdded,
			Title:   "Anggota Keluarga Baru",
			Message: fmt.Sprintf("%s menambahkan anggota baru: %s", adder.FullName, person.FirstName),
			Data:    json.RawMessage(data),
		}

		_ = s.notifRepo.Create(ctx, notif)
	}

	return nil
}

func (s *service) NotifyRelationshipAdded(ctx context.Context, relationshipID uuid.UUID, addedBy uuid.UUID) error {
	rel, err := s.relationshipRepo.GetByID(ctx, relationshipID)
	if err != nil {
		return fmt.Errorf("failed to get relationship: %w", err)
	}

	adder, err := s.userRepo.GetByID(ctx, addedBy)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	personA, _ := s.personRepo.GetByID(ctx, rel.PersonA)
	personB, _ := s.personRepo.GetByID(ctx, rel.PersonB)

	recipients, err := s.userRepo.GetByRoles(ctx, []domain.UserRole{domain.RoleEditor, domain.RoleDeveloper})
	if err != nil {
		return err
	}

	for _, user := range recipients {
		if user.ID == addedBy {
			continue
		}

		dataMap := map[string]string{
			"relationship_id": relationshipID.String(),
		}
		data, _ := json.Marshal(dataMap)

		msg := fmt.Sprintf("%s menambahkan hubungan baru", adder.FullName)
		if personA != nil && personB != nil {
			msg = fmt.Sprintf("%s menghubungkan %s dan %s (%s)", adder.FullName, personA.FirstName, personB.FirstName, rel.Type)
		}

		notif := &domain.Notification{
			ID:      uuid.New(),
			UserID:  user.ID,
			Type:    domain.NotifRelationshipAdded,
			Title:   "Hubungan Keluarga Baru",
			Message: msg,
			Data:    json.RawMessage(data),
		}

		_ = s.notifRepo.Create(ctx, notif)
	}

	return nil
}


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
