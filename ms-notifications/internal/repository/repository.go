package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepo struct {
	db *pgxpool.Pool
}

func NewNotificationRepo(db *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{db: db}
}

// Записываем факт отправки уведомления в базу
func (r *NotificationRepo) CreateNotification(ctx context.Context, userID, bookingID int, nType, channel, message string) error {
	query := `
		INSERT INTO notifications (user_id, booking_id, type, channel, status, message, sent_at)
		VALUES ($1, $2, $3, $4, 'sent', $5, $6)
	`
	_, err := r.db.Exec(ctx, query, userID, bookingID, nType, channel, message, time.Now())
	return err
}
