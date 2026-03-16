package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	PostgresURL string
	RedisAddr   string
	// Добавляем поля для S3 (MinIO)
	MinioEndpoint string
	MinioUser     string
	MinioPassword string
	MinioBucket   string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:        getEnv("PORT", "8082"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://admin:password@localhost:5432/cobooking?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		// Настройки MinIO из нашего docker-compose.yml
		MinioEndpoint: getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioUser:     getEnv("MINIO_USER", "admin"),
		MinioPassword: getEnv("MINIO_PASSWORD", "password123"),
		MinioBucket:   getEnv("MINIO_BUCKET", "workspaces"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
