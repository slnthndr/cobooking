package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/slnt/cobooking/ms-auth/internal/config"
	delivery "github.com/slnt/cobooking/ms-auth/internal/delivery/http"
	"github.com/slnt/cobooking/ms-auth/internal/repository"
	"github.com/slnt/cobooking/ms-auth/internal/service"
	"github.com/slnt/cobooking/ms-auth/pkg/db"
)

func main() {
	cfg := config.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgPool, err := db.NewPostgresPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer pgPool.Close()

	redisClient, err := db.NewRedisClient(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisClient.Close()

	// Инициализация слоев (Dependency Injection)
	userRepo := repository.NewUserRepository(pgPool)
	tokenRepo := repository.NewTokenRepository(redisClient)
	authService := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret)

	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong from MS Auth"))
	})

	// Подключаем хэндлеры
	delivery.NewAuthHandler(r, authService)

	log.Printf("Starting MS Auth on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
