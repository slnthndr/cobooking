package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	PostgresURL string
	RabbitMQURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:        getEnv("PORT", "8083"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://admin:password@localhost:5432/cobooking?sslmode=disable"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://admin:password@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
