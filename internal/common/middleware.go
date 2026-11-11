package common

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics — общий набор HTTP-метрик. Регистрируем один раз.
type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer, namespace string) *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
	}
	reg.MustRegister(m.RequestsTotal, m.RequestDuration)
	return m
}

// statusRecorder перехватывает статус ответа.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Observe оборачивает хендлер: логирует, генерирует request_id, пишет метрики.
func (m *Metrics) Observe(log *slog.Logger, route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := uuid.NewString()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next(rec, r)

		dur := time.Since(start)
		statusStr := strconv.Itoa(rec.status)

		m.RequestsTotal.WithLabelValues(r.Method, route, statusStr).Inc()
		m.RequestDuration.WithLabelValues(r.Method, route, statusStr).Observe(dur.Seconds())

		log.Info("http_request",
			"request_id", reqID,
			"method", r.Method,
			"route", route,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", dur.Milliseconds(),
			"remote", r.RemoteAddr,
		)
	}
}
