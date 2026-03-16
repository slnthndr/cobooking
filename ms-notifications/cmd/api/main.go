package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/slnt/cobooking/ms-notifications/internal/config"
	"github.com/slnt/cobooking/ms-notifications/internal/consumer"
	"github.com/slnt/cobooking/ms-notifications/internal/repository"
	"github.com/slnt/cobooking/ms-notifications/pkg/db"
	"github.com/slnt/cobooking/ms-notifications/pkg/mq"
)

func main() {
	cfg := config.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Подключаем PostgreSQL
	pgPool, err := db.NewPostgresPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer pgPool.Close()

	// 2. Подключаем RabbitMQ
	amqpConn, amqpChannel, err := mq.NewRabbitMQConn(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to rabbitmq: %v", err)
	}
	defer amqpConn.Close()
	defer amqpChannel.Close()

	// 3. Инициализируем зависимости
	repo := repository.NewNotificationRepo(pgPool)
	eventConsumer := consumer.NewEventConsumer(amqpChannel, repo)

	// 4. ЗАПУСКАЕМ СЛУШАТЕЛЬ ОЧЕРЕДИ В ФОНЕ (Горутина)
	go eventConsumer.StartConsuming()

	// 5. Запускаем простой HTTP сервер (например, для проверки Health или GET-запросов)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong from MS Notifications"))
	})

	log.Printf("Starting MS Notifications on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
