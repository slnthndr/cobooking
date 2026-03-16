package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/slnt/cobooking/ms-booking/internal/config"
	delivery "github.com/slnt/cobooking/ms-booking/internal/delivery/http"
	"github.com/slnt/cobooking/ms-booking/internal/repository"
	"github.com/slnt/cobooking/ms-booking/internal/service"
	"github.com/slnt/cobooking/ms-booking/pkg/db"
	"github.com/slnt/cobooking/ms-booking/pkg/mq"
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

	// 3. DI (Внедрение зависимостей)
	repo := repository.NewBookingRepository(pgPool)
	publisher := service.NewEventPublisher(amqpChannel)
	bookingService := service.NewBookingService(repo, publisher)

	// 4. Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Подключаем Хэндлеры
	delivery.NewBookingHandler(r, bookingService)

	log.Printf("Starting MS Booking on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
