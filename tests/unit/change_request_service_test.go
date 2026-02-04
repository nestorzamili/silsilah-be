package unit_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/service/changerequest"
	"silsilah-keluarga/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChangeRequestService_Create(t *testing.T) {
	mockCRRepo := new(mocks.ChangeRequestRepository)
	mockNotifRepo := new(mocks.NotificationRepository)
	mockUserRepo := new(mocks.UserRepository)
	mockPersonRepo := new(mocks.PersonRepository)
	mockRelRepo := new(mocks.RelationshipRepository)
	mockMediaRepo := new(mocks.MediaRepository)
	mockAuditRepo := new(mocks.AuditLogRepository)
	mockNotifSvc := new(mocks.NotificationService)

	svc := changerequest.NewService(
		mockCRRepo, mockNotifRepo, mockUserRepo, mockPersonRepo, mockRelRepo, mockMediaRepo, mockAuditRepo,
		nil, nil, nil, // Dependent services not needed for Create logic
	)
	svc.SetNotificationService(mockNotifSvc)

	ctx := context.Background()
	userID := uuid.New()

	payload := map[string]interface{}{
		"first_name": "New",
		"last_name":  "Person",
	}
	payloadBytes, _ := json.Marshal(payload)

	input := domain.CreateChangeRequestInput{
		EntityType:    domain.EntityPerson,
		Action:        domain.ActionCreate,
		Payload:       payloadBytes,
		RequesterNote: stringPtr("Please add"),
	}

	t.Run("Success", func(t *testing.T) {
		mockCRRepo.On("Create", ctx, mock.MatchedBy(func(cr *domain.ChangeRequest) bool {
			return cr.RequestedBy == userID && cr.Status == domain.StatusPending
		})).Return(nil).Once()

		// IMPORTANT: The service code ignores errors from NotifyChangeRequest.
		mockNotifSvc.On("NotifyChangeRequest", mock.MatchedBy(func(c context.Context) bool { return true }), mock.AnythingOfType("uuid.UUID"), userID).Return(nil).Maybe()

		cr, err := svc.Create(ctx, userID, input)

		assert.NoError(t, err)
		assert.NotNil(t, cr)
		assert.Equal(t, domain.StatusPending, cr.Status)

		mockCRRepo.AssertExpectations(t)
		mockNotifSvc.AssertExpectations(t)
	})

	t.Run("Validation Error - Invalid Payload", func(t *testing.T) {
		invalidInput := input
		invalidInput.Payload = []byte("invalid json")

		cr, err := svc.Create(ctx, userID, invalidInput)

		assert.Error(t, err)
		assert.Nil(t, cr)
		assert.Contains(t, err.Error(), "invalid JSON")
	})

	t.Run("Validation Error - Missing EntityID for Update", func(t *testing.T) {
		updateInput := input
		updateInput.Action = domain.ActionUpdate
		updateInput.EntityID = nil // Should fail

		cr, err := svc.Create(ctx, userID, updateInput)

		assert.Error(t, err)
		assert.Nil(t, cr)
		assert.Contains(t, err.Error(), "entity_id required")
	})
}

func TestChangeRequestService_Approve(t *testing.T) {
	mockCRRepo := new(mocks.ChangeRequestRepository)
	mockUserRepo := new(mocks.UserRepository)
	mockAuditRepo := new(mocks.AuditLogRepository)
	mockPersonRepo := new(mocks.PersonRepository)
	mockNotifSvc := new(mocks.NotificationService)

	svc := changerequest.NewService(
		mockCRRepo, nil, mockUserRepo, mockPersonRepo, nil, nil, mockAuditRepo,
		nil, nil, nil,
	)
	svc.SetNotificationService(mockNotifSvc)

	ctx := context.Background()
	crID := uuid.New()
	reviewerID := uuid.New()
	requesterID := uuid.New()

	cr := &domain.ChangeRequest{
		ID:          crID,
		RequestedBy: requesterID,
		Status:      domain.StatusPending,
		EntityType:  domain.EntityPerson,
		Action:      domain.ActionCreate,
		Payload:     []byte(`{"first_name":"John"}`),
	}

	reviewer := &domain.User{
		ID:   reviewerID,
		Role: string(domain.RoleEditor),
	}

	t.Run("Success", func(t *testing.T) {
		// 1. Get CR
		mockCRRepo.On("GetByID", ctx, crID).Return(cr, nil).Once()

		// 2. Validate Reviewer
		mockUserRepo.On("GetByID", ctx, reviewerID).Return(reviewer, nil).Once()

		// 3. Execute Change (Mock PersonRepo Create)
		mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil).Once()

		// 4. Notify Person Added (triggered by executePersonChange)
		// Similar to above, match any context, use Maybe because it's async
		mockNotifSvc.On("NotifyPersonAdded", mock.MatchedBy(func(c context.Context) bool { return true }), mock.AnythingOfType("uuid.UUID"), requesterID).Return(nil).Maybe()

		// 5. Update Status
		mockCRRepo.On("UpdateStatus", ctx, crID, domain.StatusApproved, reviewerID, mock.Anything).Return(nil).Once()

		// 6. Notify Requester
		// Async call
		mockNotifSvc.On("NotifyChangeApproved", mock.MatchedBy(func(c context.Context) bool { return true }), crID, reviewerID).Return(nil).Maybe()

		// 7. Audit Log
		mockAuditRepo.On("Create", ctx, mock.MatchedBy(func(log *domain.AuditLog) bool {
			return log.Action == "APPROVE_CHANGE_REQUEST" && log.UserID == reviewerID
		})).Return(nil).Once()

		err := svc.Approve(ctx, crID, reviewerID, nil, nil)

		// Give a tiny bit of time for async calls to potentially happen before we assert (optional, but helps with Maybe calls if we wanted to verify they happened)
		time.Sleep(10 * time.Millisecond)

		assert.NoError(t, err)
		mockCRRepo.AssertExpectations(t)
		mockPersonRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("Self Review Error", func(t *testing.T) {
		mockCRRepo.On("GetByID", ctx, crID).Return(cr, nil).Once()

		// Reviewer == Requester
		err := svc.Approve(ctx, crID, requesterID, nil, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot review own change request")
	})
}
