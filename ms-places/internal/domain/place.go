package domain

import (
	"context"
	"time"
)

// Модель рабочего места (совпадает с БД)
type Place struct {
	ID               int       `json:"placeId"`
	Name             string    `json:"placeName"`
	Type             string    `json:"type"` // workTable, meetingRoom и т.д.
	Description      *string    `json:"placeDescription"`
	Location         string    `json:"placeLocation"`
	Occupancy        int       `json:"placeOccupancy"`
	Capacity         int       `json:"placeCapacity"`
	Rating           float64   `json:"rating"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Интерфейсы (контракты)
type PlaceRepository interface {
	GetByID(ctx context.Context, id int) (*Place, error)
	Search(ctx context.Context, country, city string, limit, offset int) ([]Place, error)
}

// Интерфейс для кэша Redis (Паттерн Cache-Aside)
type CacheRepository interface {
	GetPlace(ctx context.Context, placeID int) (*Place, error)
	SetPlace(ctx context.Context, place *Place, ttl time.Duration) error
}

type PlaceService interface {
	GetPlace(ctx context.Context, id int) (*Place, error)
	SearchPlaces(ctx context.Context, country, city string, limit, offset int) ([]Place, error)
}
