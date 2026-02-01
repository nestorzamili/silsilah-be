package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"silsilah-keluarga/internal/config"
	"silsilah-keluarga/internal/domain"
	"silsilah-keluarga/internal/repository"
)

type MediaService interface {
	Upload(ctx context.Context, userID uuid.UUID, personID *uuid.UUID, caption *string, fileName string, fileSize int64, mimeType string, reader io.Reader, status string) (*domain.Media, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, personID *uuid.UUID, params domain.PaginationParams) (domain.PaginatedResponse[domain.Media], error)
	Approve(ctx context.Context, id uuid.UUID) error
}

type mediaService struct {
	mediaRepo   repository.MediaRepository
	minioClient *minio.Client
	cfg         *config.Config
}

func NewMediaService(mediaRepo repository.MediaRepository, minioClient *minio.Client, cfg *config.Config) MediaService {
	return &mediaService{
		mediaRepo:   mediaRepo,
		minioClient: minioClient,
		cfg:         cfg,
	}
}

func (s *mediaService) Upload(ctx context.Context, userID uuid.UUID, personID *uuid.UUID, caption *string, fileName string, fileSize int64, mimeType string, reader io.Reader, status string) (*domain.Media, error) {
	mediaID := uuid.New()
	storagePath := fmt.Sprintf("media/%s/%s", time.Now().Format("2006/01"), mediaID.String())

	_, err := s.minioClient.PutObject(ctx, s.cfg.MinIOBucket, storagePath, reader, fileSize, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	media := &domain.Media{
		ID:          mediaID,
		PersonID:    personID,
		UploadedBy:  userID,
		FileName:    fileName,
		FileSize:    fileSize,
		MimeType:    mimeType,
		StoragePath: storagePath,
		Caption:     caption,
		Status:      status,
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		_ = s.minioClient.RemoveObject(ctx, s.cfg.MinIOBucket, storagePath, minio.RemoveObjectOptions{})
		return nil, err
	}

	media.URL = s.getPublicURL(storagePath)
	return media, nil
}

func (s *mediaService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	media, err := s.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	media.URL = s.getPublicURL(media.StoragePath)
	return media, nil
}

func (s *mediaService) Delete(ctx context.Context, id uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.mediaRepo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.minioClient.RemoveObject(ctx, s.cfg.MinIOBucket, media.StoragePath, minio.RemoveObjectOptions{})
	return nil
}

func (s *mediaService) List(ctx context.Context, personID *uuid.UUID, params domain.PaginationParams) (domain.PaginatedResponse[domain.Media], error) {
	mediaList, total, err := s.mediaRepo.List(ctx, personID, params)
	if err != nil {
		return domain.PaginatedResponse[domain.Media]{}, err
	}

	for i := range mediaList {
		mediaList[i].URL = s.getPublicURL(mediaList[i].StoragePath)
	}

	return domain.NewPaginatedResponse(mediaList, params.Page, params.PageSize, total), nil
}

func (s *mediaService) getPublicURL(storagePath string) string {
	scheme := "http"
	if s.cfg.MinIOPublicUseSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.cfg.MinIOPublicEndpoint, s.cfg.MinIOBucket, url.PathEscape(storagePath))
}

func (s *mediaService) Approve(ctx context.Context, id uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if media.Status != "pending" {
		return fmt.Errorf("media is not pending approval")
	}

	media.Status = "active"
	return s.mediaRepo.Update(ctx, media)
}
