package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

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

	tp, err := tracing.InitTracer("api-gateway", "jaeger:4318")
	if err == nil && tp != nil {
		defer tp.Shutdown(context.Background())
		log.Println("OpenTelemetry Tracing enabled")
	}

	r := chi.NewRouter()

	r.Use(otelchi.Middleware("api-gateway", otelchi.WithChiRoutes(r)))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.PrometheusMetrics)

	r.Handle("/metrics", promhttp.Handler())

	r.Get("/health", jsonStatus("ok"))
	r.Get("/ready", jsonStatus("ready"))

	// proxies
	authProxy := proxy.New(cfg.MsAuthURL)
	placesProxy := proxy.New(cfg.MsPlacesURL)
	bookingProxy := proxy.New(cfg.MsBookingURL)
	paymentsProxy := proxy.New(cfg.MsPaymentsURL)

	// PUBLIC
	r.Group(func(r chi.Router) {
		r.Post("/api/v1/users", authProxy.ServeHTTP)
		r.Post("/api/v1/auth/login", authProxy.ServeHTTP)

		r.Get("/api/v1/places/{placeId}", placesProxy.ServeHTTP)
		r.Get("/api/v1/search", placesProxy.ServeHTTP)

		r.Post("/api/v1/payments/webhook", paymentsProxy.ServeHTTP)

		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong"))
		})
	})

	// PROTECTED
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.JWTAuth(cfg.JWTSecret))
		r.Use(InjectUserHeader)

		r.Get("/api/v1/auth/check", func(w http.ResponseWriter, r *http.Request) {
			userId := getUserID(r.Context())
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"authenticated":true,"userId":"` + userId + `"}`))
		})

		r.Delete("/api/v1/users/{userId}", authProxy.ServeHTTP)
		r.Post("/api/v1/bookings/{placeId}", bookingProxy.ServeHTTP)
		r.Post("/api/v1/payments/pay", paymentsProxy.ServeHTTP)
		r.Post("/api/v1/places/{placeId}/image", placesProxy.ServeHTTP)
	})

	log.Printf("API Gateway running on :%s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}

func InjectUserHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := getUserID(r.Context())
		r.Header.Set("X-User-Id", userId)
		next.ServeHTTP(w, r)
	})
}

func getUserID(ctx context.Context) string {
	v := ctx.Value("userId")

	switch val := v.(type) {
	case int:
		return strconv.Itoa(val)
	case float64:
		return strconv.Itoa(int(val))
	case string:
		return val
	default:
		return ""
	}
}

func jsonStatus(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"` + status + `"}`))
	}
}
