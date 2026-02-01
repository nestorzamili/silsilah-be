package service

import (
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"

	"silsilah-keluarga/internal/config"
	"silsilah-keluarga/internal/repository"
)

type Services struct {
	Auth          AuthService
	User          UserService
	Person        PersonService
	Relationship  RelationshipService
	Graph         GraphService
	ChangeRequest ChangeRequestService
	Media         MediaService
	Comment       CommentService
	Email         EmailService
	Audit         AuditService
	Notification  NotificationService
}

func NewServices(repos *repository.Repositories, redis *redis.Client, minioClient *minio.Client, cfg *config.Config) *Services {
	emailService := NewEmailService(cfg)
	authService := NewAuthService(repos.User, repos.Session, emailService, cfg)
	personService := NewPersonService(repos.Person, repos.Relationship, repos.AuditLog, redis)
	relationshipService := NewRelationshipService(repos.Relationship, repos.Person, repos.AuditLog, redis)
	mediaService := NewMediaService(repos.Media, minioClient, cfg)
	notificationService := NewNotificationService(repos.Notification, repos.User, repos.ChangeRequest, repos.Comment, repos.Person)

	return &Services{
		Auth:          authService,
		User:          NewUserService(repos.User),
		Person:        personService,
		Relationship:  relationshipService,
		Graph:         NewGraphService(repos.Person, repos.Relationship, redis),
		ChangeRequest: NewChangeRequestService(
			repos.ChangeRequest,
			repos.Notification,
			repos.User,
			repos.Person,
			repos.Relationship,
			repos.Media,
			repos.AuditLog,
			personService,
			relationshipService,
			mediaService,
		),
		Media:        mediaService,
		Comment:      NewCommentService(repos.Comment, redis),
		Email:        emailService,
		Audit:        NewAuditService(repos.AuditLog),
		Notification: notificationService,
	}
}
