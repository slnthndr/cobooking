package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/slnt/cobooking/ms-booking/internal/domain"
)

// === 1. RabbitMQ Event Publisher ===
type eventPublisher struct {
	channel *amqp.Channel
}

func NewEventPublisher(ch *amqp.Channel) domain.EventPublisher {
	return &eventPublisher{channel: ch}
}

func (p *eventPublisher) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(ctx,
		"",            // exchange (используем дефолтный)
		"EventsQueue", // routing key (имя очереди по ТЗ)
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		return fmt.Errorf("failed to publish event %s: %w", eventType, err)
	}

	fmt.Printf("[RabbitMQ] Published event: %s\n", string(body))
	return nil
}

// === 2. Booking Service ===
type bookingService struct {
	repo      domain.BookingRepository
	publisher domain.EventPublisher
}

func NewBookingService(repo domain.BookingRepository, publisher domain.EventPublisher) domain.BookingService {
	return &bookingService{repo: repo, publisher: publisher}
}

func (s *bookingService) CreateBooking(ctx context.Context, req domain.CreateBookingRequest) (*domain.Booking, error) {
	// 1. Создаем бронь в БД (с проверкой на двойную бронь внутри транзакции)
	booking, err := s.repo.CreateBookingTx(ctx, req)
	if err != nil {
		return nil, err
	}

	// 2. Отправляем событие bookingCreated (Асинхронно, по ТЗ стр. 39)
	eventPayload := map[string]interface{}{
		"type":      "bookingCreated",
		"bookingId": booking.ID,
		"userId":    booking.UserID,
		"placeId":   booking.PlaceID,
		"startTime": booking.StartTime.Format(time.RFC3339),
		"endTime":   booking.EndTime.Format(time.RFC3339),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Делаем это в горутине, чтобы не задерживать ответ клиенту, если RabbitMQ тормозит
	go func() {
		// Создаем новый контекст, так как контекст HTTP запроса скоро отменится
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := s.publisher.PublishEvent(bgCtx, "bookingCreated", eventPayload)
		if err != nil {
			fmt.Printf("Error publishing event: %v\n", err)
		}
	}()

	return booking, nil
}
