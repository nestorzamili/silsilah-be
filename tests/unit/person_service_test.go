package unit_test

import (
	"context"
	"errors"
	"testing"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/service/person"
	"silsilah-keluarga/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPersonService_Create(t *testing.T) {
	mockPersonRepo := new(mocks.PersonRepository)
	mockRelRepo := new(mocks.RelationshipRepository)
	mockAuditRepo := new(mocks.AuditLogRepository)

	svc := person.NewService(mockPersonRepo, mockRelRepo, mockAuditRepo, nil)
	ctx := context.Background()
	userID := uuid.New()

	input := domain.CreatePersonInput{
		FirstName: "John",
		LastName:  stringPtr("Doe"),
		Gender:    domain.GenderMale,
	}

	t.Run("Success", func(t *testing.T) {
		mockPersonRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.Person) bool {
			return p.FirstName == "John" && *p.LastName == "Doe" && p.CreatedBy == userID
		})).Return(nil).Once()

		mockAuditRepo.On("Create", ctx, mock.MatchedBy(func(log *domain.AuditLog) bool {
			return log.Action == "CREATE" && log.EntityType == "PERSON" && log.UserID == userID
		})).Return(nil).Once()

		person, err := svc.Create(ctx, userID, input)

		assert.NoError(t, err)
		assert.NotNil(t, person)
		assert.Equal(t, "John", person.FirstName)
		assert.Equal(t, "Doe", *person.LastName)

		mockPersonRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("Repo Error", func(t *testing.T) {
		mockPersonRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error")).Once()

		person, err := svc.Create(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, person)
		assert.Equal(t, "db error", err.Error())

		mockPersonRepo.AssertExpectations(t)
	})
}

func TestPersonService_GetByID(t *testing.T) {
	mockPersonRepo := new(mocks.PersonRepository)
	mockRelRepo := new(mocks.RelationshipRepository)
	mockAuditRepo := new(mocks.AuditLogRepository)

	svc := person.NewService(mockPersonRepo, mockRelRepo, mockAuditRepo, nil)
	ctx := context.Background()
	personID := uuid.New()

	t.Run("Found", func(t *testing.T) {
		expectedPerson := &domain.Person{ID: personID, FirstName: "Jane"}
		mockPersonRepo.On("GetByID", ctx, personID).Return(expectedPerson, nil).Once()

		person, err := svc.GetByID(ctx, personID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPerson, person)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockPersonRepo.On("GetByID", ctx, personID).Return(nil, nil).Once()

		person, err := svc.GetByID(ctx, personID)

		assert.ErrorIs(t, err, domain.ErrPersonNotFound)
		assert.Nil(t, person)
	})
}

func stringPtr(s string) *string {
	return &s
}
