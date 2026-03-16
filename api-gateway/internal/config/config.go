package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	RedisAddr string
	JWTSecret string
	MsAuthURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:      getEnv("PORT", "8080"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret: "MY_SUPER_SECRET_KEY_123", // <-- Точно такая же строка
		MsAuthURL: getEnv("MS_AUTH_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
