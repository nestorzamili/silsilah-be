package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	User          UserRepository
	Person        PersonRepository
	Relationship  RelationshipRepository
	ChangeRequest ChangeRequestRepository
	Media         MediaRepository
	Comment       CommentRepository
	AuditLog      AuditLogRepository
	Notification  NotificationRepository
	Session       SessionRepository
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		User:          NewUserRepository(db),
		Person:        NewPersonRepository(db),
		Relationship:  NewRelationshipRepository(db),
		ChangeRequest: NewChangeRequestRepository(db),
		Media:         NewMediaRepository(db),
		Comment:       NewCommentRepository(db),
		AuditLog:      NewAuditLogRepository(db),
		Notification:  NewNotificationRepository(db),
		Session:       NewSessionRepository(db),
	}
}
