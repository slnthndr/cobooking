package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	RedisAddr     string
	JWTSecret     string
	MsAuthURL     string
	MsPlacesURL   string
	MsBookingURL  string
	MsPaymentsURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:          getEnv("PORT", "8080"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:     "MY_SUPER_SECRET_KEY_123", // <-- Точно такая же строка
		MsAuthURL:     getEnv("MS_AUTH_URL", "http://localhost:8081"),
		MsPlacesURL:   getEnv("MS_PLACES_URL", "http://ms-places:8082"),
		MsBookingURL:  getEnv("MS_BOOKING_URL", "http://ms-booking:8083"),
		MsPaymentsURL: getEnv("MS_PAYMENTS_URL", "http://ms-payments:8085"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
