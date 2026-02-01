package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

var (
	ErrInsufficientPermissions   = errors.New("insufficient permissions")
	ErrCannotModifySelf          = errors.New("cannot modify your own role")
)

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateUserInput) (*domain.User, error)
	AssignRole(ctx context.Context, currentUser *domain.User, input domain.AssignRoleInput) error
	ListByRole(ctx context.Context, role string) ([]domain.User, error)
	GetRoleUsers(ctx context.Context) (map[string][]domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	DeleteUser(ctx context.Context, currentUser *domain.User, userID string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateUserInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if input.FullName != nil {
		user.FullName = *input.FullName
	}
	if input.Email != nil && *input.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, *input.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("email already in use")
		}
		user.Email = *input.Email
	}
	if input.Password != nil && *input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hashedPassword)
	}
	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.PersonID != nil {
		user.PersonID = *input.PersonID
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	updatedUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return user, nil
	}

	return updatedUser, nil
}

func (s *userService) AssignRole(ctx context.Context, currentUser *domain.User, input domain.AssignRoleInput) error {
	if !currentUser.HasRole("developer") {
		return ErrInsufficientPermissions
	}

	if currentUser.ID == input.UserID {
		return ErrCannotModifySelf
	}

	targetUser, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return err
	}
	if targetUser == nil {
		return ErrUserNotFound
	}

	if targetUser.Role == "developer" && currentUser.Role != "developer" {
		return ErrInsufficientPermissions
	}

	return s.userRepo.AssignRole(ctx, input.UserID, input.Role)
}

func (s *userService) ListByRole(ctx context.Context, role string) ([]domain.User, error) {
	return s.userRepo.ListByRole(ctx, role)
}

func (s *userService) GetRoleUsers(ctx context.Context) (map[string][]domain.User, error) {
	roles := []string{"member", "editor", "developer"}
	result := make(map[string][]domain.User)

	for _, role := range roles {
		users, err := s.userRepo.ListByRole(ctx, role)
		if err != nil {
			return nil, err
		}
		result[role] = users
	}

	return result, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.userRepo.GetAllUsers(ctx)
}

func (s *userService) DeleteUser(ctx context.Context, currentUser *domain.User, userID string) error {
	if !currentUser.HasRole("developer") {
		return ErrInsufficientPermissions
	}

	if currentUser.ID.String() == userID {
		return ErrCannotModifySelf
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}
