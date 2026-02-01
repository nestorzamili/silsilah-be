package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

type ChangeRequestService interface {
	Create(ctx context.Context, userID uuid.UUID, input domain.CreateChangeRequestInput) (*domain.ChangeRequest, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error)
	List(ctx context.Context, status *domain.ChangeRequestStatus, params domain.PaginationParams) (domain.PaginatedResponse[domain.ChangeRequest], error)
	Approve(ctx context.Context, id, reviewerID uuid.UUID, note *string) error
	Reject(ctx context.Context, id, reviewerID uuid.UUID, note *string) error
}

type changeRequestService struct {
	crRepo       repository.ChangeRequestRepository
	notifRepo    repository.NotificationRepository
	userRepo     repository.UserRepository
	personRepo   repository.PersonRepository
	relRepo      repository.RelationshipRepository
	mediaRepo    repository.MediaRepository
	auditRepo    repository.AuditLogRepository
	personSvc    PersonService
	relSvc       RelationshipService
	mediaSvc     MediaService
}

func NewChangeRequestService(
	crRepo repository.ChangeRequestRepository,
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	personRepo repository.PersonRepository,
	relRepo repository.RelationshipRepository,
	mediaRepo repository.MediaRepository,
	auditRepo repository.AuditLogRepository,
	personSvc PersonService,
	relSvc RelationshipService,
	mediaSvc MediaService,
) ChangeRequestService {
	return &changeRequestService{
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

func (s *changeRequestService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateChangeRequestInput) (*domain.ChangeRequest, error) {
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

func (s *changeRequestService) validatePayload(input domain.CreateChangeRequestInput) error {
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

func (s *changeRequestService) notifyReviewers(ctx context.Context, cr *domain.ChangeRequest) {
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

func (s *changeRequestService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error) {
	return s.crRepo.GetByID(ctx, id)
}

func (s *changeRequestService) List(ctx context.Context, status *domain.ChangeRequestStatus, params domain.PaginationParams) (domain.PaginatedResponse[domain.ChangeRequest], error) {
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

func (s *changeRequestService) Approve(ctx context.Context, id, reviewerID uuid.UUID, note *string) error {
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

	s.notifyRequester(ctx, cr, domain.StatusApproved, note)
	s.logAudit(ctx, reviewerID, "APPROVE_CHANGE_REQUEST", cr)

	return nil
}

func (s *changeRequestService) Reject(ctx context.Context, id, reviewerID uuid.UUID, note *string) error {
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

	s.notifyRequester(ctx, cr, domain.StatusRejected, note)
	s.logAudit(ctx, reviewerID, "REJECT_CHANGE_REQUEST", cr)

	return nil
}

func (s *changeRequestService) validateReview(ctx context.Context, cr *domain.ChangeRequest, reviewerID uuid.UUID) error {
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

func (s *changeRequestService) executeChange(ctx context.Context, cr *domain.ChangeRequest) error {
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

func (s *changeRequestService) executePersonChange(ctx context.Context, cr *domain.ChangeRequest) error {
	switch cr.Action {
	case domain.ActionCreate:
		var person domain.Person
		if err := json.Unmarshal(cr.Payload, &person); err != nil {
			return err
		}
		person.ID = uuid.New()
		person.CreatedBy = cr.RequestedBy
		return s.personRepo.Create(ctx, &person)

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

func (s *changeRequestService) executeRelationshipChange(ctx context.Context, cr *domain.ChangeRequest) error {
	switch cr.Action {
	case domain.ActionCreate:
		var input struct {
			PersonA  uuid.UUID              `json:"person_a"`
			PersonB  uuid.UUID              `json:"person_b"`
			Type     domain.RelationshipType `json:"type"`
			Metadata json.RawMessage        `json:"metadata,omitempty"`
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
		return s.relRepo.Create(ctx, rel)

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

func (s *changeRequestService) executeMediaChange(ctx context.Context, cr *domain.ChangeRequest) error {
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

func (s *changeRequestService) notifyRequester(ctx context.Context, cr *domain.ChangeRequest, status domain.ChangeRequestStatus, note *string) {
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

func (s *changeRequestService) logAudit(ctx context.Context, reviewerID uuid.UUID, action string, cr *domain.ChangeRequest) {
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

	_ = s.auditRepo.Create(ctx, audit)
}
