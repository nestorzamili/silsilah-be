package changerequest

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service/media"
	"silsilah-keluarga/internal/service/notification"
	"silsilah-keluarga/internal/service/person"
	"silsilah-keluarga/internal/service/relationship"
)

type RequestMeta struct {
	IPAddress string
	UserAgent string
}

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, input domain.CreateChangeRequestInput) (*domain.ChangeRequest, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error)
	List(ctx context.Context, status *domain.ChangeRequestStatus, params domain.PaginationParams) (domain.PaginatedResponse[domain.ChangeRequest], error)
	Approve(ctx context.Context, id, reviewerID uuid.UUID, note *string, meta *RequestMeta) error
	Reject(ctx context.Context, id, reviewerID uuid.UUID, note *string, meta *RequestMeta) error
	SetNotificationService(notifSvc notification.Service)
}

type service struct {
	crRepo     repository.ChangeRequestRepository
	notifRepo  repository.NotificationRepository
	userRepo   repository.UserRepository
	personRepo repository.PersonRepository
	relRepo    repository.RelationshipRepository
	mediaRepo  repository.MediaRepository
	auditRepo  repository.AuditLogRepository
	personSvc  person.Service
	relSvc     relationship.Service
	mediaSvc   media.Service
	notifSvc   notification.Service
}

func NewService(
	crRepo repository.ChangeRequestRepository,
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	personRepo repository.PersonRepository,
	relRepo repository.RelationshipRepository,
	mediaRepo repository.MediaRepository,
	auditRepo repository.AuditLogRepository,
	personSvc person.Service,
	relSvc relationship.Service,
	mediaSvc media.Service,
) Service {
	return &service{
		crRepo:     crRepo,
		notifRepo:  notifRepo,
		userRepo:   userRepo,
		personRepo: personRepo,
		relRepo:    relRepo,
		mediaRepo:  mediaRepo,
		auditRepo:  auditRepo,
		personSvc:  personSvc,
		relSvc:     relSvc,
		mediaSvc:   mediaSvc,
	}
}

func (s *service) SetNotificationService(notifSvc notification.Service) {
	s.notifSvc = notifSvc
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, input domain.CreateChangeRequestInput) (*domain.ChangeRequest, error) {
	if err := s.validatePayload(input); err != nil {
		return nil, err
	}

	cr := &domain.ChangeRequest{
		ID:            uuid.New(),
		RequestedBy:   userID,
		EntityType:    input.EntityType,
		EntityID:      input.EntityID,
		Action:        input.Action,
		Payload:       input.Payload,
		RequesterNote: input.RequesterNote,
		Status:        domain.StatusPending,
	}

	if err := s.crRepo.Create(ctx, cr); err != nil {
		return nil, err
	}

	s.notifyReviewers(ctx, cr)

	return cr, nil
}

func (s *service) validatePayload(input domain.CreateChangeRequestInput) error {
	if !json.Valid(input.Payload) {
		return errors.New("invalid JSON payload")
	}

	if input.Action == domain.ActionUpdate || input.Action == domain.ActionDelete {
		if input.EntityID == nil {
			return errors.New("entity_id required for update/delete actions")
		}
	}

	if input.Action == domain.ActionCreate && input.EntityID != nil {
		return errors.New("entity_id must be null for create actions")
	}

	return nil
}

func (s *service) notifyReviewers(ctx context.Context, cr *domain.ChangeRequest) {
	if s.notifSvc != nil {
		go func() {
			_ = s.notifSvc.NotifyChangeRequest(context.Background(), cr.ID, cr.RequestedBy)
		}()
		return
	}

	reviewers, err := s.userRepo.GetByRoles(ctx, []domain.UserRole{domain.RoleEditor, domain.RoleDeveloper})
	if err != nil {
		return
	}

	for _, reviewer := range reviewers {
		if reviewer.ID == cr.RequestedBy {
			continue
		}

		notif := &domain.Notification{
			ID:      uuid.New(),
			UserID:  reviewer.ID,
			Type:    domain.NotifChangeRequest,
			Title:   "New Change Request",
			Message: "A new change request requires your review",
			Data:    json.RawMessage(`{"change_request_id":"` + cr.ID.String() + `"}`),
		}

		_ = s.notifRepo.Create(ctx, notif)
	}
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error) {
	return s.crRepo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, status *domain.ChangeRequestStatus, params domain.PaginationParams) (domain.PaginatedResponse[domain.ChangeRequest], error) {
	requests, total, err := s.crRepo.List(ctx, status, params)
	if err != nil {
		return domain.PaginatedResponse[domain.ChangeRequest]{}, err
	}

	for i := range requests {
		if requester, err := s.userRepo.GetByID(ctx, requests[i].RequestedBy); err == nil {
			requests[i].Requester = requester
		}

		if requests[i].ReviewedBy != nil {
			if reviewer, err := s.userRepo.GetByID(ctx, *requests[i].ReviewedBy); err == nil {
				requests[i].Reviewer = reviewer
			}
		}
	}

	return domain.NewPaginatedResponse(requests, params.Page, params.PageSize, total), nil
}

func (s *service) Approve(ctx context.Context, id, reviewerID uuid.UUID, note *string, meta *RequestMeta) error {
	cr, err := s.crRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.validateReview(ctx, cr, reviewerID); err != nil {
		return err
	}

	if err := s.executeChange(ctx, cr); err != nil {
		return err
	}

	if err := s.crRepo.UpdateStatus(ctx, id, domain.StatusApproved, reviewerID, note); err != nil {
		return err
	}

	s.notifyRequester(ctx, cr, domain.StatusApproved, reviewerID, note)
	s.logAudit(ctx, reviewerID, "APPROVE_CHANGE_REQUEST", cr, meta)

	return nil
}

func (s *service) Reject(ctx context.Context, id, reviewerID uuid.UUID, note *string, meta *RequestMeta) error {
	cr, err := s.crRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.validateReview(ctx, cr, reviewerID); err != nil {
		return err
	}

	if err := s.crRepo.UpdateStatus(ctx, id, domain.StatusRejected, reviewerID, note); err != nil {
		return err
	}

	s.notifyRequester(ctx, cr, domain.StatusRejected, reviewerID, note)
	s.logAudit(ctx, reviewerID, "REJECT_CHANGE_REQUEST", cr, meta)

	return nil
}

func (s *service) validateReview(ctx context.Context, cr *domain.ChangeRequest, reviewerID uuid.UUID) error {
	if cr.Status != domain.StatusPending {
		return errors.New("change request is not pending")
	}

	if cr.RequestedBy == reviewerID {
		return errors.New("cannot review own change request")
	}

	reviewer, err := s.userRepo.GetByID(ctx, reviewerID)
	if err != nil {
		return err
	}

	if reviewer.Role != string(domain.RoleEditor) && reviewer.Role != string(domain.RoleDeveloper) {
		return errors.New("insufficient permissions to review change request")
	}

	if reviewer.DeletedAt != nil {
		return errors.New("reviewer account is inactive")
	}

	return nil
}

func (s *service) executeChange(ctx context.Context, cr *domain.ChangeRequest) error {
	switch cr.EntityType {
	case domain.EntityPerson:
		return s.executePersonChange(ctx, cr)
	case domain.EntityRelationship:
		return s.executeRelationshipChange(ctx, cr)
	case domain.EntityMedia:
		return s.executeMediaChange(ctx, cr)
	default:
		return errors.New("unknown entity type")
	}
}

func (s *service) executePersonChange(ctx context.Context, cr *domain.ChangeRequest) error {
	switch cr.Action {
	case domain.ActionCreate:
		var person domain.Person
		if err := json.Unmarshal(cr.Payload, &person); err != nil {
			return err
		}
		person.ID = uuid.New()
		person.CreatedBy = cr.RequestedBy
		if err := s.personRepo.Create(ctx, &person); err != nil {
			return err
		}

		if s.notifSvc != nil {
			go func() {
				_ = s.notifSvc.NotifyPersonAdded(context.Background(), person.ID, cr.RequestedBy)
			}()
		}
		return nil

	case domain.ActionUpdate:
		if cr.EntityID == nil {
			return errors.New("entity_id required for update")
		}
		var updates domain.Person
		if err := json.Unmarshal(cr.Payload, &updates); err != nil {
			return err
		}
		existing, err := s.personRepo.GetByID(ctx, *cr.EntityID)
		if err != nil {
			return err
		}
		if existing.DeletedAt != nil {
			return errors.New("cannot update deleted person")
		}
		updates.ID = *cr.EntityID
		return s.personRepo.Update(ctx, &updates)

	case domain.ActionDelete:
		if cr.EntityID == nil {
			return errors.New("entity_id required for delete")
		}
		return s.personRepo.Delete(ctx, *cr.EntityID)

	default:
		return errors.New("unknown action")
	}
}

func (s *service) executeRelationshipChange(ctx context.Context, cr *domain.ChangeRequest) error {
	switch cr.Action {
	case domain.ActionCreate:
		var input struct {
			PersonA  uuid.UUID               `json:"person_a"`
			PersonB  uuid.UUID               `json:"person_b"`
			Type     domain.RelationshipType `json:"type"`
			Metadata json.RawMessage         `json:"metadata,omitempty"`
		}
		if err := json.Unmarshal(cr.Payload, &input); err != nil {
			return err
		}
		rel := &domain.Relationship{
			ID:        uuid.New(),
			PersonA:   input.PersonA,
			PersonB:   input.PersonB,
			Type:      input.Type,
			Metadata:  input.Metadata,
			CreatedBy: cr.RequestedBy,
		}
		if err := s.relRepo.Create(ctx, rel); err != nil {
			return err
		}

		if s.notifSvc != nil {
			go func() {
				_ = s.notifSvc.NotifyRelationshipAdded(context.Background(), rel.ID, cr.RequestedBy)
			}()
		}
		return nil

	case domain.ActionUpdate:
		if cr.EntityID == nil {
			return errors.New("entity_id required for update")
		}
		var updates domain.Relationship
		if err := json.Unmarshal(cr.Payload, &updates); err != nil {
			return err
		}
		existing, err := s.relRepo.GetByID(ctx, *cr.EntityID)
		if err != nil {
			return err
		}
		if existing.DeletedAt != nil {
			return errors.New("cannot update deleted relationship")
		}
		updates.ID = *cr.EntityID
		return s.relRepo.Update(ctx, &updates)

	case domain.ActionDelete:
		if cr.EntityID == nil {
			return errors.New("entity_id required for delete")
		}
		return s.relRepo.Delete(ctx, *cr.EntityID)

	default:
		return errors.New("unknown action")
	}
}

func (s *service) executeMediaChange(ctx context.Context, cr *domain.ChangeRequest) error {
	switch cr.Action {
	case domain.ActionCreate:
		if cr.EntityID == nil {
			return errors.New("entity_id required for media approval")
		}
		return s.mediaSvc.Approve(ctx, *cr.EntityID)

	case domain.ActionDelete:
		if cr.EntityID == nil {
			return errors.New("entity_id required for delete")
		}
		return s.mediaRepo.Delete(ctx, *cr.EntityID)

	default:
		return errors.New("unsupported action for media")
	}
}

func (s *service) notifyRequester(ctx context.Context, cr *domain.ChangeRequest, status domain.ChangeRequestStatus, reviewerID uuid.UUID, note *string) {
	if s.notifSvc != nil {
		go func() {
			if status == domain.StatusApproved {
				_ = s.notifSvc.NotifyChangeApproved(context.Background(), cr.ID, reviewerID)
			} else {
				_ = s.notifSvc.NotifyChangeRejected(context.Background(), cr.ID, reviewerID)
			}
		}()
		return
	}

	var title, message string
	if status == domain.StatusApproved {
		title = "Change Request Approved"
		message = "Your change request has been approved"
	} else {
		title = "Change Request Rejected"
		message = "Your change request has been rejected"
	}

	if note != nil && *note != "" {
		message += ": " + *note
	}

	notif := &domain.Notification{
		ID:      uuid.New(),
		UserID:  cr.RequestedBy,
		Type:    domain.NotifChangeApproved,
		Title:   title,
		Message: message,
		Data:    json.RawMessage(`{"change_request_id":"` + cr.ID.String() + `"}`),
	}

	if status == domain.StatusRejected {
		notif.Type = domain.NotifChangeRejected
	}

	_ = s.notifRepo.Create(ctx, notif)
}

func (s *service) logAudit(ctx context.Context, reviewerID uuid.UUID, action string, cr *domain.ChangeRequest, meta *RequestMeta) {
	var entityID uuid.UUID
	if cr.EntityID != nil {
		entityID = *cr.EntityID
	} else {
		entityID = cr.ID
	}

	audit := &domain.AuditLog{
		ID:         uuid.New(),
		UserID:     reviewerID,
		Action:     action,
		EntityType: string(cr.EntityType),
		EntityID:   entityID,
		OldValue:   json.RawMessage(`{"status":"PENDING"}`),
		NewValue:   json.RawMessage(`{"status":"` + string(cr.Status) + `"}`),
		CreatedAt:  time.Now(),
	}

	if meta != nil {
		if meta.IPAddress != "" {
			audit.IPAddress = &meta.IPAddress
		}
		if meta.UserAgent != "" {
			audit.UserAgent = &meta.UserAgent
		}
	}

	_ = s.auditRepo.Create(ctx, audit)
}
