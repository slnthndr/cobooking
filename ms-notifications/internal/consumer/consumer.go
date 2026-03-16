package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/slnt/cobooking/ms-notifications/internal/repository"
)

type EventConsumer struct {
	channel *amqp.Channel
	repo    *repository.NotificationRepo
}

func NewEventConsumer(ch *amqp.Channel, repo *repository.NotificationRepo) *EventConsumer {
	return &EventConsumer{channel: ch, repo: repo}
}

// Структура, которую мы ожидаем получить из RabbitMQ
type BaseEvent struct {
	Type      string `json:"type"`
	BookingID int    `json:"bookingId"`
	UserID    int    `json:"userId"`
}

func (c *EventConsumer) StartConsuming() {
	// Подписываемся на очередь "EventsQueue"
	msgs, err := c.channel.Consume(
		"EventsQueue", // Очередь
		"",            // Имя консьюмера
		true,          // auto-ack (автоматически подтверждаем получение)
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("[*] Waiting for events in EventsQueue. To exit press CTRL+C")

	// Читаем канал в бесконечном цикле
	for d := range msgs {
		var event BaseEvent
		if err := json.Unmarshal(d.Body, &event); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			continue
		}

		fmt.Printf("--> Received event: %s for BookingID: %d\n", event.Type, event.BookingID)

		// Обрабатываем конкретные события (ТЗ стр. 39)
		switch event.Type {
		case "bookingCreated":
			c.handleBookingCreated(event)
		case "paymentCompleted":
			fmt.Println("Payment completed logic here...")
		default:
			fmt.Println("Unknown event type, skipping...")
		}
	}
}

func (c *EventConsumer) handleBookingCreated(event BaseEvent) {
	msg := fmt.Sprintf("Dear User %d, your booking #%d has been successfully created!", event.UserID, event.BookingID)
	
	// Эмулируем отправку Email/Push
	fmt.Println("\n========================================")
	fmt.Printf("📩 SENDING EMAIL TO USER ID: %d\n", event.UserID)
	fmt.Printf("📝 MESSAGE: %s\n", msg)
	fmt.Println("========================================\n")

	// Сохраняем в БД (ТЗ требует тип "bookingConfirmed" для уведомлений о созданной брони)
	err := c.repo.CreateNotification(context.Background(), event.UserID, event.BookingID, "bookingConfirmed", "email", msg)
	if err != nil {
		log.Printf("Failed to save notification to DB: %v", err)
	}
}
