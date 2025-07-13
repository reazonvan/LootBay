package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics структура для хранения метрик
type Metrics struct {
	// HTTP метрики
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// База данных метрики
	dbConnectionsTotal *prometheus.CounterVec
	dbQueryDuration    *prometheus.HistogramVec
	dbConnectionsInUse *prometheus.GaugeVec

	// Кэш метрики
	cacheHitsTotal   *prometheus.CounterVec
	cacheMissesTotal *prometheus.CounterVec
	cacheSize        *prometheus.GaugeVec

	// Бизнес метрики
	userRegistrationsTotal *prometheus.CounterVec
	userLoginsTotal        *prometheus.CounterVec
	ordersCreatedTotal     *prometheus.CounterVec
	paymentsProcessedTotal *prometheus.CounterVec

	// Системные метрики
	goroutinesTotal *prometheus.GaugeVec
	memoryUsage     *prometheus.GaugeVec
	cpuUsage        *prometheus.GaugeVec

	// Ошибки
	errorsTotal *prometheus.CounterVec
}

// NewMetrics создает новый экземпляр метрик
func NewMetrics(serviceName string) *Metrics {
	return &Metrics{
		// HTTP метрики
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "path", "status_code"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "path"},
		),
		httpRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
			[]string{"service"},
		),

		// База данных метрики
		dbConnectionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_connections_total",
				Help: "Total number of database connections",
			},
			[]string{"service", "database", "operation"},
		),
		dbQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "database", "operation"},
		),
		dbConnectionsInUse: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "db_connections_in_use",
				Help: "Current number of database connections in use",
			},
			[]string{"service", "database"},
		),

		// Кэш метрики
		cacheHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"service", "cache_type"},
		),
		cacheMissesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"service", "cache_type"},
		),
		cacheSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cache_size",
				Help: "Current cache size",
			},
			[]string{"service", "cache_type"},
		),

		// Бизнес метрики
		userRegistrationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "user_registrations_total",
				Help: "Total number of user registrations",
			},
			[]string{"service", "status"},
		),
		userLoginsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "user_logins_total",
				Help: "Total number of user logins",
			},
			[]string{"service", "status"},
		),
		ordersCreatedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_created_total",
				Help: "Total number of orders created",
			},
			[]string{"service", "status"},
		),
		paymentsProcessedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payments_processed_total",
				Help: "Total number of payments processed",
			},
			[]string{"service", "payment_method", "status"},
		),

		// Системные метрики
		goroutinesTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "goroutines_total",
				Help: "Current number of goroutines",
			},
			[]string{"service"},
		),
		memoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
			[]string{"service", "type"},
		),
		cpuUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
			[]string{"service"},
		),

		// Ошибки
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors",
			},
			[]string{"service", "type", "code"},
		),
	}
}

// RecordHTTPRequest записывает метрики HTTP запроса
func (m *Metrics) RecordHTTPRequest(service, method, path string, statusCode int, duration time.Duration) {
	m.httpRequestsTotal.WithLabelValues(service, method, path, fmt.Sprintf("%d", statusCode)).Inc()
	m.httpRequestDuration.WithLabelValues(service, method, path).Observe(duration.Seconds())
}

// RecordHTTPRequestInFlight увеличивает счетчик активных запросов
func (m *Metrics) RecordHTTPRequestInFlight(service string, delta float64) {
	m.httpRequestsInFlight.WithLabelValues(service).Add(delta)
}

// RecordDatabaseConnection записывает метрики подключения к БД
func (m *Metrics) RecordDatabaseConnection(service, database, operation string, duration time.Duration) {
	m.dbConnectionsTotal.WithLabelValues(service, database, operation).Inc()
	m.dbQueryDuration.WithLabelValues(service, database, operation).Observe(duration.Seconds())
}

// RecordDatabaseConnectionsInUse записывает количество активных подключений к БД
func (m *Metrics) RecordDatabaseConnectionsInUse(service, database string, count float64) {
	m.dbConnectionsInUse.WithLabelValues(service, database).Set(count)
}

// RecordCacheHit записывает попадание в кэш
func (m *Metrics) RecordCacheHit(service, cacheType string) {
	m.cacheHitsTotal.WithLabelValues(service, cacheType).Inc()
}

// RecordCacheMiss записывает промах кэша
func (m *Metrics) RecordCacheMiss(service, cacheType string) {
	m.cacheMissesTotal.WithLabelValues(service, cacheType).Inc()
}

// RecordCacheSize записывает размер кэша
func (m *Metrics) RecordCacheSize(service, cacheType string, size float64) {
	m.cacheSize.WithLabelValues(service, cacheType).Set(size)
}

// RecordUserRegistration записывает регистрацию пользователя
func (m *Metrics) RecordUserRegistration(service, status string) {
	m.userRegistrationsTotal.WithLabelValues(service, status).Inc()
}

// RecordUserLogin записывает вход пользователя
func (m *Metrics) RecordUserLogin(service, status string) {
	m.userLoginsTotal.WithLabelValues(service, status).Inc()
}

// RecordOrderCreated записывает создание заказа
func (m *Metrics) RecordOrderCreated(service, status string) {
	m.ordersCreatedTotal.WithLabelValues(service, status).Inc()
}

// RecordPaymentProcessed записывает обработку платежа
func (m *Metrics) RecordPaymentProcessed(service, paymentMethod, status string) {
	m.paymentsProcessedTotal.WithLabelValues(service, paymentMethod, status).Inc()
}

// RecordError записывает ошибку
func (m *Metrics) RecordError(service, errorType, code string) {
	m.errorsTotal.WithLabelValues(service, errorType, code).Inc()
}

// RecordSystemMetrics записывает системные метрики
func (m *Metrics) RecordSystemMetrics(service string, goroutines int, memoryBytes uint64, cpuPercent float64) {
	m.goroutinesTotal.WithLabelValues(service).Set(float64(goroutines))
	m.memoryUsage.WithLabelValues(service, "heap").Set(float64(memoryBytes))
	m.cpuUsage.WithLabelValues(service).Set(cpuPercent)
}

// StartMetricsCollector запускает сборщик метрик
func (m *Metrics) StartMetricsCollector(ctx context.Context, serviceName string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Собираем системные метрики каждые 30 секунд
			// В реальном приложении здесь можно добавить сбор метрик через runtime/debug
			// и runtime.ReadMemStats
		}
	}
}

// MetricsMiddleware middleware для автоматического сбора HTTP метрик
func MetricsMiddleware(metrics *Metrics, serviceName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Увеличиваем счетчик активных запросов
			metrics.RecordHTTPRequestInFlight(serviceName, 1)
			defer metrics.RecordHTTPRequestInFlight(serviceName, -1)

			// Оборачиваем ResponseWriter для получения статус кода
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(wrappedWriter, r)

			// Записываем метрики
			duration := time.Since(start)
			metrics.RecordHTTPRequest(serviceName, r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
		})
	}
}

// responseWriter обертка для ResponseWriter для получения статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
