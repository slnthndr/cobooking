package service

import (
	"context"
	"fmt"
	"time"

	"github.com/slnt/cobooking/ms-places/internal/domain"
)

type placeService struct {
	placeRepo domain.PlaceRepository
	cacheRepo domain.CacheRepository
}

func NewPlaceService(placeRepo domain.PlaceRepository, cacheRepo domain.CacheRepository) domain.PlaceService {
	return &placeService{placeRepo: placeRepo, cacheRepo: cacheRepo}
}

func (s *placeService) GetPlace(ctx context.Context, id int) (*domain.Place, error) {
	// 1. Пытаемся взять из кэша (Redis)
	cachedPlace, err := s.cacheRepo.GetPlace(ctx, id)
	if err == nil && cachedPlace != nil {
		fmt.Printf("[CACHE HIT] Place ID: %d\n", id) // Для наглядности в логах
		return cachedPlace, nil
	}

	fmt.Printf("[CACHE MISS] Place ID: %d. Fetching from DB...\n", id)

	// 2. Если в кэше пусто, идем в PostgreSQL
	place, err := s.placeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем в кэш на 5 минут (как по ТЗ, стр. 27: "TTL: 5-10 минут")
	go func() {
		// Асинхронно сохраняем в кэш, чтобы не тормозить отдачу ответа пользователю
		_ = s.cacheRepo.SetPlace(context.Background(), place, 5*time.Minute)
	}()

	return place, nil
}

func (s *placeService) SearchPlaces(ctx context.Context, country, city string, limit, offset int) ([]domain.Place, error) {
	// Для поиска кэширование обычно сложнее (инвалидация), поэтому пока просто идем в БД
	return s.placeRepo.Search(ctx, country, city, limit, offset)
}
