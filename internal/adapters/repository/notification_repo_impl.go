package database

import (
	"context"
	"database/sql"
	"notification-service/internal/adapters/database/db"
)

type NotificationLogRepo struct {
	db *db.Queries
}

func NewNotificationLogRepo(conn *sql.DB) *NotificationLogRepo {
	return &NotificationLogRepo{
		db: db.New(conn),
	}
}

func (r *NotificationLogRepo) CreateLog(ctx context.Context, params db.CreateNotificationLogParams) error {
	return r.db.CreateNotificationLog(ctx, params)
}
