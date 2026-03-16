package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slnt/cobooking/ms-booking/internal/domain"
)

type bookingRepo struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) domain.BookingRepository {
	return &bookingRepo{db: db}
}

// CreateBookingTx выполняет проверку слота и создание брони в одной транзакции
func (r *bookingRepo) CreateBookingTx(ctx context.Context, req domain.CreateBookingRequest) (*domain.Booking, error) {
	// 1. Начинаем транзакцию
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	// Если что-то пойдет не так, откатываем изменения
	defer tx.Rollback(ctx)

	// 2. Проверяем пересечения (Защита от двойного бронирования)
	// Смотрим, есть ли активные брони на это место, которые пересекаются по времени
	overlapQuery := `
		SELECT COUNT(*) FROM bookings 
		WHERE place_id = $1 
		AND booking_status IN ('created', 'reserved', 'paid') 
		AND booking_start_time < $3 
		AND booking_end_time > $2
	`
	var count int
	err = tx.QueryRow(ctx, overlapQuery, req.PlaceID, req.StartTime, req.EndTime).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, errors.New("SLOT_ALREADY_RESERVED") // Как написано на стр. 8 ТЗ
	}

	// 3. Если слот свободен, создаем бронь
	insertQuery := `
		INSERT INTO bookings (user_id, place_id, booking_start_time, booking_end_time, booking_status)
		VALUES ($1, $2, $3, $4, 'created')
		RETURNING booking_id, created_at
	`
	b := &domain.Booking{
		UserID:    req.UserID,
		PlaceID:   req.PlaceID,
		Status:    "created",
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	err = tx.QueryRow(ctx, insertQuery, req.UserID, req.PlaceID, req.StartTime, req.EndTime).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 4. Фиксируем транзакцию (сохраняем в БД)
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return b, nil
}
