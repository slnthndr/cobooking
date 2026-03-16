package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/slnt/cobooking/ms-places/internal/domain"
)

// === POSTGRESQL ===
type placeRepo struct {
	db *pgxpool.Pool
}

func NewPlaceRepository(db *pgxpool.Pool) domain.PlaceRepository {
	return &placeRepo{db: db}
}

func (r *placeRepo) GetByID(ctx context.Context, id int) (*domain.Place, error) {
	query := `
		SELECT place_id, place_name, place_type, place_description, 
		       place_location, place_occupancy, place_capacity, rating, 
		       created_at, updated_at
		FROM places WHERE place_id = $1`

	p := &domain.Place{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Type, &p.Description,
		&p.Location, &p.Occupancy, &p.Capacity, &p.Rating,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("place not found")
	}
	return p, err
}

func (r *placeRepo) Search(ctx context.Context, country, city string, limit, offset int) ([]domain.Place, error) {
	// Для упрощения ищем по вхождению строки (LIKE) в place_location
	searchStr := "%" + country + "%" + city + "%"
	if country == "" && city == "" {
		searchStr = "%" // Ищем все, если фильтров нет
	}

	query := `
		SELECT place_id, place_name, place_type, place_description, 
		       place_location, place_occupancy, place_capacity, rating, 
		       created_at, updated_at
		FROM places 
		WHERE place_location ILIKE $1 
		ORDER BY rating DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, searchStr, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	places := make([]domain.Place, 0)
	for rows.Next() {
		var p domain.Place
		err := rows.Scan(
			&p.ID, &p.Name, &p.Type, &p.Description,
			&p.Location, &p.Occupancy, &p.Capacity, &p.Rating,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		places = append(places, p)
	}
	return places, nil
}

// === REDIS (Кэш) ===
type cacheRepo struct {
	rdb *redis.Client
}

func NewCacheRepository(rdb *redis.Client) domain.CacheRepository {
	return &cacheRepo{rdb: rdb}
}

// Чтение из кэша
func (r *cacheRepo) GetPlace(ctx context.Context, placeID int) (*domain.Place, error) {
	key := fmt.Sprintf("place:%d", placeID) // Ключ по ТЗ (стр. 27)
	
	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Кэш промахнулся (Cache miss)
	} else if err != nil {
		return nil, err // Реальная ошибка Redis
	}

	var place domain.Place
	if err := json.Unmarshal([]byte(val), &place); err != nil {
		return nil, err
	}
	return &place, nil
}

// Запись в кэш
func (r *cacheRepo) SetPlace(ctx context.Context, place *domain.Place, ttl time.Duration) error {
	key := fmt.Sprintf("place:%d", place.ID)
	data, err := json.Marshal(place)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, ttl).Err()
}
