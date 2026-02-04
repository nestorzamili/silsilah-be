package service

import (
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"

	"silsilah-keluarga/internal/config"
	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service/audit"
	"silsilah-keluarga/internal/service/auth"
	"silsilah-keluarga/internal/service/changerequest"
	"silsilah-keluarga/internal/service/comment"
	"silsilah-keluarga/internal/service/dashboard"
	"silsilah-keluarga/internal/service/email"
	"silsilah-keluarga/internal/service/export"
	"silsilah-keluarga/internal/service/graph"
	"silsilah-keluarga/internal/service/media"
	"silsilah-keluarga/internal/service/narrative"
	"silsilah-keluarga/internal/service/notification"
	"silsilah-keluarga/internal/service/person"
	"silsilah-keluarga/internal/service/relationship"
	"silsilah-keluarga/internal/service/user"
)

type Services struct {
	Auth          auth.Service
	User          user.Service
	Person        person.Service
	Relationship  relationship.Service
	Graph         graph.Service
	ChangeRequest changerequest.Service
	Media         media.Service
	Comment       comment.Service
	Email         email.Service
	Audit         audit.Service
	Notification  notification.Service
	Dashboard     dashboard.Service
	Narrative     narrative.Service
	Export        export.Service
}

func NewServices(repos *repository.Repositories, redis *redis.Client, minioClient *minio.Client, cfg *config.Config) *Services {
	emailService := email.NewService(cfg)
	authService := auth.NewService(repos.User, repos.Session, emailService, cfg)
	personService := person.NewService(repos.Person, repos.Relationship, repos.AuditLog, redis)
	auditService := audit.NewService(repos.AuditLog)
	relationshipService := relationship.NewService(repos.Relationship, repos.Person, repos.AuditLog, redis)
	mediaService := media.NewService(repos.Media, minioClient, cfg)
	narrativeService := narrative.NewService(repos.Person, repos.Relationship)
	graphService := graph.NewService(repos.Person, repos.Relationship, redis, narrativeService)
	commentService := comment.NewService(repos.Comment, redis)
	notificationService := notification.NewService(repos.Notification, repos.User, repos.ChangeRequest, repos.Comment, repos.Person, repos.Relationship, emailService)
	commentService.SetNotificationService(notificationService)
	personService.SetNotificationService(notificationService)
	relationshipService.SetNotificationService(notificationService)

	changeRequestService := changerequest.NewService(
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
	)
	changeRequestService.SetNotificationService(notificationService)

	dashboardService := dashboard.NewService(repos.Person, repos.Relationship, repos.ChangeRequest, redis)
	exportService := export.NewService(repos.Person, repos.Relationship, repos.AuditLog, graphService)
	userService := user.NewService(repos.User)

	return &Services{
		Auth:          authService,
		User:          userService,
		Person:        personService,
		Relationship:  relationshipService,
		Graph:         graphService,
		ChangeRequest: changeRequestService,
		Media:         mediaService,
		Comment:       commentService,
		Email:         emailService,
		Audit:         auditService,
		Notification:  notificationService,
		Dashboard:     dashboardService,
		Narrative:     narrativeService,
		Export:        exportService,
	}
}
