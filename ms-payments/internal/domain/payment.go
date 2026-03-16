package domain

import (
	"context"
	"time"
)

type Payment struct {
	ID        int       `json:"paymentId"`
	BookingID int       `json:"bookingId"`
	UserID    int       `json:"userId"`
	Status    string    `json:"paymentStatus"`
	Method    string    `json:"paymentMethod"`
	Provider  string    `json:"paymentProvider"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"createdAt"`
}

type PayRequest struct {
	BookingID     int     `json:"bookingId"`
	UserID        int     `json:"userId"`
	PaymentMethod string  `json:"paymentMethod"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

type WebhookPayload struct {
	PaymentID     int    `json:"paymentId"`
	TransactionID string `json:"transactionId"`
	Status        string `json:"paymentStatus"`
}

type PaymentRepository interface {
	CreatePayment(ctx context.Context, p *Payment) error
	UpdateStatus(ctx context.Context, paymentID int, status, transactionID string) error
	GetPaymentByID(ctx context.Context, paymentID int) (*Payment, error)
}

type EventPublisher interface {
	PublishEvent(ctx context.Context, eventType string, payload interface{}) error
}

type PaymentService interface {
	Pay(ctx context.Context, req PayRequest) (*Payment, error)
	ProcessWebhook(ctx context.Context, payload WebhookPayload) error
}
