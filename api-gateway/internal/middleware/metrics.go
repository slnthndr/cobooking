package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Инициализируем метрики, которые будет собирать Prometheus
var (
	// Счетчик количества запросов
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP запросов",
		},
		[]string{"method", "path", "status"},
	)

	// Гистограмма времени выполнения запросов (Latency)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Время выполнения HTTP запроса",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// PrometheusMetrics - это middleware, который оборачивает каждый запрос
func PrometheusMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter, чтобы иметь возможность прочитать HTTP Status Code (200, 404, 500)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Передаем запрос дальше (в наши прокси или роуты)
		next.ServeHTTP(ww, r)

		// Считаем, сколько времени занял запрос
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(ww.Status())

		// Записываем данные в Prometheus
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}
