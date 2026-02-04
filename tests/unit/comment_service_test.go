package unit_test

import (
	"context"
	"testing"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/service/comment"
	"silsilah-keluarga/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCommentService_Create(t *testing.T) {
	mockRepo := new(mocks.CommentRepository)
	mockNotifSvc := new(mocks.NotificationService)
	svc := comment.NewService(mockRepo, nil) // Redis nil
	svc.SetNotificationService(mockNotifSvc)

	ctx := context.Background()
	personID := uuid.New()
	userID := uuid.New()
	input := domain.CreateCommentInput{
		Content: "Test comment",
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.MatchedBy(func(c *domain.Comment) bool {
			return c.PersonID == personID && c.UserID == userID && c.Content == input.Content
		})).Return(nil).Once()

		mockNotifSvc.On("NotifyNewComment", mock.MatchedBy(func(c context.Context) bool { return true }), mock.AnythingOfType("uuid.UUID"), userID).Return(nil).Maybe()

		c, err := svc.Create(ctx, personID, userID, input)

		assert.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, input.Content, c.Content)
		mockRepo.AssertExpectations(t)
	})
}

func TestCommentService_Update(t *testing.T) {
	mockRepo := new(mocks.CommentRepository)
	svc := comment.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	commentID := uuid.New()
	
	existingComment := &domain.Comment{
		ID:     commentID,
		UserID: userID,
		Content: "Original",
	}

	input := domain.UpdateCommentInput{
		Content: "Updated",
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, commentID).Return(existingComment, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(c *domain.Comment) bool {
			return c.ID == commentID && c.Content == "Updated"
		})).Return(nil).Once()

		c, err := svc.Update(ctx, userID, commentID, input)

		assert.NoError(t, err)
		assert.Equal(t, "Updated", c.Content)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Permission Error", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, commentID).Return(existingComment, nil).Once()

		c, err := svc.Update(ctx, otherUserID, commentID, input)

		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "insufficient permissions")
	})
}

func TestCommentService_Delete(t *testing.T) {
	mockRepo := new(mocks.CommentRepository)
	svc := comment.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	commentID := uuid.New()
	
	existingComment := &domain.Comment{
		ID:     commentID,
		UserID: userID,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, commentID).Return(existingComment, nil).Once()
		mockRepo.On("Delete", ctx, commentID).Return(nil).Once()

		err := svc.Delete(ctx, userID, commentID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Permission Error", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, commentID).Return(existingComment, nil).Once()

		err := svc.Delete(ctx, otherUserID, commentID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient permissions")
	})
}
