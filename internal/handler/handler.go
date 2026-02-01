package handler

import "silsilah-keluarga/internal/service"

type Handlers struct {
	Auth          *AuthHandler
	User          *UserHandler
	Person        *PersonHandler
	Relationship  *RelationshipHandler
	Graph         *GraphHandler
	ChangeRequest *ChangeRequestHandler
	Media         *MediaHandler
	Comment       *CommentHandler
	Public        *PublicHandler
	Audit         *AuditHandler
	Notification  *NotificationHandler
}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Auth:          NewAuthHandler(services.Auth),
		User:          NewUserHandler(services.User, services.Person),
		Person:        NewPersonHandler(services.Person),
		Relationship:  NewRelationshipHandler(services.Relationship),
		Graph:         NewGraphHandler(services.Graph),
		ChangeRequest: NewChangeRequestHandler(services.ChangeRequest),
		Media:         NewMediaHandler(services.Media, services.ChangeRequest),
		Comment:       NewCommentHandler(services.Comment),
		Public:        NewPublicHandler(services.Person, services.Graph),
		Audit:         NewAuditHandler(services.Audit),
		Notification:  NewNotificationHandler(services.Notification),
	}
}
