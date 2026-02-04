package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"silsilah-keluarga/internal/config"
	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service/email"
)

var (
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrEmailExists              = errors.New("email already registered")
	ErrInvalidToken             = errors.New("invalid or expired token")
	ErrUserNotFound             = errors.New("user not found")
	ErrTokenExpired             = errors.New("password reset token has expired")
	ErrEmailNotVerified         = errors.New("email not verified")
	ErrVerificationTokenExpired = errors.New("email verification token has expired")
)

type Service interface {
	Register(ctx context.Context, input domain.CreateUserInput) (*domain.User, *domain.TokenPair, error)
	Login(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	ValidateAccessToken(token string) (*Claims, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, email string) error
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type service struct {
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	emailService   email.Service
	cfg            *config.Config
}

func NewService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, emailService email.Service, cfg *config.Config) Service {
	return &service{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		emailService:   emailService,
		cfg:            cfg,
	}
}

func (s *service) Register(ctx context.Context, input domain.CreateUserInput) (*domain.User, *domain.TokenPair, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user := &domain.User{
		ID:              uuid.New(),
		Email:           input.Email,
		PasswordHash:    string(hashedPassword),
		FullName:        input.FullName,
		Role:            "member",
		IsActive:        true,
		IsEmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, nil, err
	}
	verificationToken := hex.EncodeToString(tokenBytes)

	now := time.Now()
	if err := s.userRepo.SetEmailVerificationToken(ctx, user.ID, verificationToken, now); err != nil {
		return nil, nil, err
	}

	go func() {
		err := s.emailService.SendEmailVerification(context.Background(), user.Email, user.FullName, verificationToken)
		if err != nil {
			fmt.Printf("Failed to send verification email: %v\n", err)
		}
	}()

	return user, nil, nil
}

func (s *service) Login(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.IsEmailVerified {
		return nil, nil, ErrEmailNotVerified
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *service) generateTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	accessClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTAccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshTokenRaw := uuid.New().String()
	refreshTokenHash := hashToken(refreshTokenRaw)

	session := &repository.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(s.cfg.JWTRefreshExpiry),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenRaw,
		ExpiresIn:    int64(s.cfg.JWTAccessExpiry.Seconds()),
	}, nil
}

func (s *service) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	resetToken := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(time.Hour)

	if err := s.userRepo.SetPasswordResetToken(ctx, user.ID, resetToken, expiresAt); err != nil {
		return err
	}

	go func() {
		err := s.emailService.SendPasswordResetEmail(context.Background(), user.Email, user.FullName, resetToken)
		if err != nil {
			fmt.Printf("Failed to send password reset email: %v\n", err)
		}
	}()

	return nil
}

func (s *service) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.userRepo.GetUserByResetToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrInvalidToken
	}

	if user.PasswordResetExpiresAt != nil && time.Now().After(*user.PasswordResetExpiresAt) {
		return ErrTokenExpired
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if err := s.userRepo.ClearPasswordResetToken(ctx, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *service) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.userRepo.GetUserByEmailVerificationToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrInvalidToken
	}

	if user.EmailVerificationSentAt != nil && time.Now().After(user.EmailVerificationSentAt.Add(24*time.Hour)) {
		return ErrVerificationTokenExpired
	}

	if err := s.userRepo.VerifyEmail(ctx, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *service) ResendVerificationEmail(ctx context.Context, emailAddr string) error {
	user, err := s.userRepo.GetByEmail(ctx, emailAddr)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	if user.IsEmailVerified {
		return nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	verificationToken := hex.EncodeToString(tokenBytes)

	now := time.Now()
	if err := s.userRepo.SetEmailVerificationToken(ctx, user.ID, verificationToken, now); err != nil {
		return err
	}

	go func() {
		err := s.emailService.SendEmailVerification(context.Background(), user.Email, user.FullName, verificationToken)
		if err != nil {
			fmt.Printf("Failed to send verification email: %v\n", err)
		}
	}()

	return nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
