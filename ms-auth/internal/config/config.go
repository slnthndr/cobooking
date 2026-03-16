package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	PostgresURL string
	RedisAddr   string
	JWTSecret   string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:        getEnv("PORT", "8081"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://admin:password@localhost:5432/cobooking?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:   "MY_SUPER_SECRET_KEY_123", // <-- Просто строка, без getEnv
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
