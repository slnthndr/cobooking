package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"

	"github.com/slnt/cobooking/api-gateway/internal/config"
	customMiddleware "github.com/slnt/cobooking/api-gateway/internal/middleware"
	"github.com/slnt/cobooking/api-gateway/internal/proxy"
	"github.com/slnt/cobooking/api-gateway/internal/tracing"
)

func main() {
	cfg := config.LoadConfig()

	// === ИНИЦИАЛИЗАЦИЯ JAEGER TRACING ===
	tp, err := tracing.InitTracer("api-gateway", "jaeger:4318") // "jaeger" - имя контейнера
	if err == nil {
		defer tp.Shutdown(context.Background())
		log.Println("OpenTelemetry Tracing is enabled")
	}

	// 1. Создаем роутер (ОДИН РАЗ!)
	r := chi.NewRouter()

	// 2. Добавляем Middleware для Трассировки (Jaeger)
	r.Use(otelchi.Middleware("api-gateway", otelchi.WithChiRoutes(r)))

	// 3. Стандартные Middleware и Метрики (Prometheus)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.PrometheusMetrics)

	// === 4. OBSERVABILITY & HEALTH CHECKS ===
	r.Handle("/metrics", promhttp.Handler())

	// ... дальше всё остается как было (r.Get("/health"... и настройка прокси) ...

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// === PROXIES ===
	authProxy := proxy.New(cfg.MsAuthURL)
	placesProxy := proxy.New("http://ms-places:8082")
	bookingProxy := proxy.New("http://ms-booking:8083")
	paymentsProxy := proxy.New("http://ms-payments:8085")

	// === PUBLIC ROUTES ===
	r.Group(func(r chi.Router) {
		r.Post("/api/v1/users", authProxy.ServeHTTP)
		r.Post("/api/v1/auth/login", authProxy.ServeHTTP)

		r.Get("/api/v1/places/{placeId}", placesProxy.ServeHTTP)
		r.Get("/api/v1/search", placesProxy.ServeHTTP)

		r.Post("/api/v1/payments/webhook", paymentsProxy.ServeHTTP)

		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong from API Gateway"))
		})
	})

	// === PROTECTED ROUTES ===
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.JWTAuth(cfg.JWTSecret))

		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userId := getAnyAsString(r.Context().Value("userId"))
				r.Header.Set("X-User-Id", userId)
				next.ServeHTTP(w, r)
			})
		})

		r.Get("/api/v1/auth/check", func(w http.ResponseWriter, r *http.Request) {
			userId := r.Context().Value("userId")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"authenticated":true,"userId":"` + getAnyAsString(userId) + `"}`))
		})

		r.Post("/api/v1/bookings/{placeId}", bookingProxy.ServeHTTP)
		r.Post("/api/v1/payments/pay", paymentsProxy.ServeHTTP)
		r.Post("/api/v1/places/{placeId}/image", placesProxy.ServeHTTP)
	}) // ВАЖНО — закрыли group

	log.Printf("API Gateway is running on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("API Gateway stopped: %v", err)
	}
}

// helper
func getAnyAsString(v any) string {
	switch val := v.(type) {
	case float64:
		return string(rune(int(val) + '0'))
	case string:
		return val
	default:
		return "unknown"
	}
}
