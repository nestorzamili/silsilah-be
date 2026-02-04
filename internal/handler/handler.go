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
	Audit         *AuditHandler
	Notification  *NotificationHandler
	Dashboard     *DashboardHandler
	Export        *ExportHandler
}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Auth:          NewAuthHandler(services.Auth),
		User:          NewUserHandler(services.User, services.Person),
		Person:        NewPersonHandler(services.Person, services.ChangeRequest),
		Relationship:  NewRelationshipHandler(services.Relationship, services.ChangeRequest),
		Graph:         NewGraphHandler(services.Graph),
		ChangeRequest: NewChangeRequestHandler(services.ChangeRequest),
		Media:         NewMediaHandler(services.Media, services.ChangeRequest),
		Comment:       NewCommentHandler(services.Comment),
		Audit:         NewAuditHandler(services.Audit),
		Notification:  NewNotificationHandler(services.Notification),
		Dashboard:     NewDashboardHandler(services.Dashboard),
		Export:        NewExportHandler(services.Export),
	}
}
