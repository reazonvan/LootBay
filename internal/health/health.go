package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/reazonvan/LootBay/internal/database"
	"github.com/reazonvan/LootBay/pkg/logger"
)

// Status статус проверки здоровья
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// HealthCheck интерфейс для health check
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthResult
}

// HealthResult результат проверки здоровья
type HealthResult struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HealthChecker основной checker здоровья
type HealthChecker struct {
	checks []HealthCheck
	logger *logger.StructuredLogger
	mu     sync.RWMutex
}

// NewHealthChecker создает новый health checker
func NewHealthChecker(logger *logger.StructuredLogger) *HealthChecker {
	return &HealthChecker{
		checks: make([]HealthCheck, 0),
		logger: logger,
	}
}

// AddCheck добавляет проверку здоровья
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks = append(hc.checks, check)
}

// Check выполняет все проверки здоровья
func (hc *HealthChecker) Check(ctx context.Context) HealthResponse {
	hc.mu.RLock()
	checks := make([]HealthCheck, len(hc.checks))
	copy(checks, hc.checks)
	hc.mu.RUnlock()

	results := make([]HealthResult, 0, len(checks))
	overallStatus := StatusHealthy

	// Выполняем все проверки параллельно
	var wg sync.WaitGroup
	resultChan := make(chan HealthResult, len(checks))

	for _, check := range checks {
		wg.Add(1)
		go func(c HealthCheck) {
			defer wg.Done()
			result := c.Check(ctx)
			resultChan <- result
		}(check)
	}

	// Ждем завершения всех проверок
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Собираем результаты
	for result := range resultChan {
		results = append(results, result)

		// Определяем общий статус
		switch result.Status {
		case StatusUnhealthy:
			overallStatus = StatusUnhealthy
		case StatusDegraded:
			if overallStatus != StatusUnhealthy {
				overallStatus = StatusDegraded
			}
		}
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    results,
	}

	// Логируем результат
	hc.logger.Info("Health check completed", map[string]interface{}{
		"overall_status":  overallStatus,
		"checks_count":    len(results),
		"unhealthy_count": countByStatus(results, StatusUnhealthy),
		"degraded_count":  countByStatus(results, StatusDegraded),
	})

	return response
}

// HealthResponse ответ с результатами проверки здоровья
type HealthResponse struct {
	Status    Status         `json:"status"`
	Timestamp time.Time      `json:"timestamp"`
	Checks    []HealthResult `json:"checks"`
}

// countByStatus подсчитывает количество проверок по статусу
func countByStatus(results []HealthResult, status Status) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}

// DatabaseHealthCheck проверка здоровья базы данных
type DatabaseHealthCheck struct {
	db     *database.PostgresDB
	logger *logger.StructuredLogger
}

// NewDatabaseHealthCheck создает проверку здоровья БД
func NewDatabaseHealthCheck(db *database.PostgresDB, logger *logger.StructuredLogger) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{
		db:     db,
		logger: logger,
	}
}

// Name возвращает название проверки
func (d *DatabaseHealthCheck) Name() string {
	return "database"
}

// Check выполняет проверку БД
func (d *DatabaseHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()

	// Проверяем подключение к БД
	var result HealthResult
	result.Name = d.Name()
	result.Timestamp = time.Now()

	// Выполняем простой запрос для проверки подключения
	var version string
	err := d.db.DB.Raw("SELECT version()").Scan(&version).Error

	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Database connection failed: %v", err)
		d.logger.Error("Database health check failed", err, map[string]interface{}{
			"duration": time.Since(start).String(),
		})
	} else {
		result.Status = StatusHealthy
		result.Message = "Database connection is healthy"
		result.Details = map[string]interface{}{
			"version":  version,
			"duration": time.Since(start).String(),
		}
		d.logger.Debug("Database health check passed", map[string]interface{}{
			"duration": time.Since(start).String(),
		})
	}

	return result
}

// RedisHealthCheck проверка здоровья Redis
type RedisHealthCheck struct {
	redis  *database.RedisDB
	logger *logger.StructuredLogger
}

// NewRedisHealthCheck создает проверку здоровья Redis
func NewRedisHealthCheck(redis *database.RedisDB, logger *logger.StructuredLogger) *RedisHealthCheck {
	return &RedisHealthCheck{
		redis:  redis,
		logger: logger,
	}
}

// Name возвращает название проверки
func (r *RedisHealthCheck) Name() string {
	return "redis"
}

// Check выполняет проверку Redis
func (r *RedisHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()

	var result HealthResult
	result.Name = r.Name()
	result.Timestamp = time.Now()

	// Проверяем подключение к Redis
	_, err := r.redis.Client.Ping(ctx).Result()

	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Redis connection failed: %v", err)
		r.logger.Error("Redis health check failed", err, map[string]interface{}{
			"duration": time.Since(start).String(),
		})
	} else {
		result.Status = StatusHealthy
		result.Message = "Redis connection is healthy"
		result.Details = map[string]interface{}{
			"duration": time.Since(start).String(),
		}
		r.logger.Debug("Redis health check passed", map[string]interface{}{
			"duration": time.Since(start).String(),
		})
	}

	return result
}

// SystemHealthCheck проверка системных ресурсов
type SystemHealthCheck struct {
	logger *logger.StructuredLogger
}

// NewSystemHealthCheck создает проверку системных ресурсов
func NewSystemHealthCheck(logger *logger.StructuredLogger) *SystemHealthCheck {
	return &SystemHealthCheck{
		logger: logger,
	}
}

// Name возвращает название проверки
func (s *SystemHealthCheck) Name() string {
	return "system"
}

// Check выполняет проверку системных ресурсов
func (s *SystemHealthCheck) Check(ctx context.Context) HealthResult {
	var result HealthResult
	result.Name = s.Name()
	result.Timestamp = time.Now()

	// В реальном приложении здесь можно добавить проверки:
	// - Использование памяти
	// - Использование CPU
	// - Количество горутин
	// - Свободное место на диске

	result.Status = StatusHealthy
	result.Message = "System resources are healthy"
	result.Details = map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		// Можно добавить другие метрики
	}

	return result
}

// HealthHandler HTTP handler для health checks
func HealthHandler(checker *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Устанавливаем timeout для проверок
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		response := checker.Check(ctx)

		// Определяем HTTP статус код
		var statusCode int
		switch response.Status {
		case StatusHealthy:
			statusCode = http.StatusOK
		case StatusDegraded:
			statusCode = http.StatusOK // 200, но с предупреждением
		case StatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		}

		// Устанавливаем заголовки
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		// Кодируем ответ
		if err := json.NewEncoder(w).Encode(response); err != nil {
			checker.logger.Error("Failed to encode health response", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// ReadinessHandler handler для проверки готовности сервиса
func ReadinessHandler(checker *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		response := checker.Check(ctx)

		// Для readiness проверяем только критичные сервисы
		criticalChecks := []string{"database", "redis"}
		allCriticalHealthy := true

		for _, check := range response.Checks {
			for _, critical := range criticalChecks {
				if check.Name == critical && check.Status == StatusUnhealthy {
					allCriticalHealthy = false
					break
				}
			}
		}

		if allCriticalHealthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready"}`))
		}
	}
}

// LivenessHandler handler для проверки жизнеспособности сервиса
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"alive"}`))
	}
}
