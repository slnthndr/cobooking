package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/slnt/cobooking/ms-payments/internal/config"
	delivery "github.com/slnt/cobooking/ms-payments/internal/delivery/http"
	"github.com/slnt/cobooking/ms-payments/internal/repository"
	"github.com/slnt/cobooking/ms-payments/internal/service"
	"github.com/slnt/cobooking/ms-payments/pkg/db"
	"github.com/slnt/cobooking/ms-payments/pkg/mq"
)

func main() {
	cfg := config.LoadConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgPool, err := db.NewPostgresPool(ctx, cfg.PostgresURL)
	if err != nil { log.Fatalf("DB Error: %v", err) }
	defer pgPool.Close()

	amqpConn, amqpChannel, err := mq.NewRabbitMQConn(cfg.RabbitMQURL)
	if err != nil { log.Fatalf("MQ Error: %v", err) }
	defer amqpConn.Close()
	defer amqpChannel.Close()

	repo := repository.NewPaymentRepository(pgPool)
	pub := service.NewEventPublisher(amqpChannel)
	svc := service.NewPaymentService(repo, pub)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	delivery.NewPaymentHandler(r, svc)

	log.Printf("Starting MS Payments on port %s", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, r)
}
