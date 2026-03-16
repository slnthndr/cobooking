package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // У нас без пароля в docker-compose
		DB:       0,
	})

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	log.Println("Connected to Redis successfully")
	return client, nil
}
