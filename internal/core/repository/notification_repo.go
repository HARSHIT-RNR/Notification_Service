package repository

import (
	"context"

	"notification-service/internal/adapters/database/db"
)

// Template represents a parsed email template.
type Template struct {
	Name     string
	Subject  string
	BodyHTML string
	BodyText string
}

// TemplateRepository is the port for fetching notification templates.
type TemplateRepository interface {
	GetTemplate(ctx context.Context, name string) (*Template, error)
}

// NotificationLogRepository is the port for logging notification attempts.
type NotificationLogRepository interface {
	CreateLog(ctx context.Context, params db.CreateNotificationLogParams) error
}
