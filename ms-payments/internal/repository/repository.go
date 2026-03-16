package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slnt/cobooking/ms-payments/internal/domain"
)

type paymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) domain.PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) CreatePayment(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (booking_id, user_id, payment_method, payment_provider, amount, currency, payment_status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
		RETURNING payment_id, created_at
	`
	return r.db.QueryRow(ctx, query, p.BookingID, p.UserID, p.Method, p.Provider, p.Amount, p.Currency).Scan(&p.ID, &p.CreatedAt)
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, paymentID int, status, transactionID string) error {
	query := `UPDATE payments SET payment_status = $1, provider_transaction_id = $2 WHERE payment_id = $3`
	_, err := r.db.Exec(ctx, query, status, transactionID, paymentID)
	return err
}

func (r *paymentRepo) GetPaymentByID(ctx context.Context, paymentID int) (*domain.Payment, error) {
	query := `SELECT payment_id, booking_id, user_id, payment_status, amount FROM payments WHERE payment_id = $1`
	p := &domain.Payment{}
	err := r.db.QueryRow(ctx, query, paymentID).Scan(&p.ID, &p.BookingID, &p.UserID, &p.Status, &p.Amount)
	if err == pgx.ErrNoRows {
		return nil, err
	}
	return p, err
}
