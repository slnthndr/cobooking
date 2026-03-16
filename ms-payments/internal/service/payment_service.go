package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/slnt/cobooking/ms-payments/internal/domain"
)

type eventPublisher struct {
	channel *amqp.Channel
}

func NewEventPublisher(ch *amqp.Channel) domain.EventPublisher {
	return &eventPublisher{channel: ch}
}

func (p *eventPublisher) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
	body, _ := json.Marshal(payload)
	err := p.channel.PublishWithContext(ctx, "", "EventsQueue", false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
	if err == nil {
		fmt.Printf("[RabbitMQ] Published: %s\n", string(body))
	}
	return err
}

type paymentService struct {
	repo      domain.PaymentRepository
	publisher domain.EventPublisher
}

func NewPaymentService(repo domain.PaymentRepository, pub domain.EventPublisher) domain.PaymentService {
	return &paymentService{repo: repo, publisher: pub}
}

func (s *paymentService) Pay(ctx context.Context, req domain.PayRequest) (*domain.Payment, error) {
	p := &domain.Payment{
		BookingID: req.BookingID,
		UserID:    req.UserID,
		Method:    req.PaymentMethod,
		Provider:  "bank",
		Amount:    req.Amount,
		Currency:  req.Currency,
		Status:    "pending",
	}

	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *paymentService) ProcessWebhook(ctx context.Context, payload domain.WebhookPayload) error {
	if err := s.repo.UpdateStatus(ctx, payload.PaymentID, payload.Status, payload.TransactionID); err != nil {
		return err
	}

	if payload.Status == "completed" || payload.Status == "paid" {
		payment, err := s.repo.GetPaymentByID(ctx, payload.PaymentID)
		if err == nil {
			event := map[string]interface{}{
				"type":      "paymentCompleted",
				"paymentId": payment.ID,
				"bookingId": payment.BookingID,
				"userId":    payment.UserID,
				"timestamp": time.Now().Format(time.RFC3339),
			}
			_ = s.publisher.PublishEvent(context.Background(), "paymentCompleted", event)
		}
	}
	return nil
}
