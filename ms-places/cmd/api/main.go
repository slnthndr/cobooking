package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/slnt/cobooking/ms-places/internal/config"
	delivery "github.com/slnt/cobooking/ms-places/internal/delivery/http"
	"github.com/slnt/cobooking/ms-places/internal/repository"
	"github.com/slnt/cobooking/ms-places/internal/service"
	"github.com/slnt/cobooking/ms-places/pkg/db"
	"github.com/slnt/cobooking/ms-places/pkg/storage"
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

	// 2. Подключаем Redis
	redisClient, err := db.NewRedisClient(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisClient.Close()

	// 3. Подключаем S3 (MinIO)
	s3Storage, err := storage.NewMinioStorage(
		cfg.MinioEndpoint,
		cfg.MinioUser,
		cfg.MinioPassword,
		cfg.MinioBucket,
	)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}

	// 4. DI (Внедрение зависимостей)
	placeRepo := repository.NewPlaceRepository(pgPool)
	cacheRepo := repository.NewCacheRepository(redisClient)
	placeService := service.NewPlaceService(placeRepo, cacheRepo)

	// 5. Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Передаем S3 в Хэндлер вместе с сервисом
	delivery.NewPlaceHandler(r, placeService, s3Storage)

	log.Printf("Starting MS Places on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
