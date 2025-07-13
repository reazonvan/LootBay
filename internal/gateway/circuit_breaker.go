package gateway

import (
	"sync"
	"time"
)

// CircuitState состояние circuit breaker
type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

// CircuitBreaker реализация circuit breaker паттерна
type CircuitBreaker struct {
	mu sync.RWMutex

	// Конфигурация
	failureThreshold int           // Количество ошибок для открытия
	timeout          time.Duration // Время ожидания перед попыткой восстановления
	resetTimeout     time.Duration // Время для полного сброса счетчиков

	// Состояние
	state CircuitState

	// Счетчики
	failureCount int
	successCount int

	// Временные метки
	lastFailureTime time.Time
	lastStateChange time.Time
}

// NewCircuitBreaker создает новый circuit breaker
func NewCircuitBreaker(failureThreshold int, timeout, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		timeout:          timeout,
		resetTimeout:     resetTimeout,
		state:            Closed,
		lastStateChange:  time.Now(),
	}
}

// Allow проверяет, можно ли выполнить запрос
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case Closed:
		return true
	case Open:
		// Проверяем, прошло ли достаточно времени для попытки восстановления
		if time.Since(cb.lastFailureTime) >= cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = HalfOpen
			cb.lastStateChange = time.Now()
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case HalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess записывает успешный запрос
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case Closed:
		// Сбрасываем счетчик ошибок при успешных запросах
		if time.Since(cb.lastStateChange) >= cb.resetTimeout {
			cb.failureCount = 0
		}
	case HalfOpen:
		cb.successCount++
		// Если достаточно успешных запросов, закрываем circuit breaker
		if cb.successCount >= cb.failureThreshold {
			cb.state = Closed
			cb.lastStateChange = time.Now()
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// RecordFailure записывает неудачный запрос
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case Closed:
		// Если достигли порога ошибок, открываем circuit breaker
		if cb.failureCount >= cb.failureThreshold {
			cb.state = Open
			cb.lastStateChange = time.Now()
		}
	case HalfOpen:
		// При ошибке в half-open состоянии, возвращаемся в open
		cb.state = Open
		cb.lastStateChange = time.Now()
		cb.successCount = 0
	}
}

// IsOpen проверяет, открыт ли circuit breaker
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == Open
}

// IsClosed проверяет, закрыт ли circuit breaker
func (cb *CircuitBreaker) IsClosed() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == Closed
}

// IsHalfOpen проверяет, находится ли circuit breaker в half-open состоянии
func (cb *CircuitBreaker) IsHalfOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == HalfOpen
}

// GetState возвращает текущее состояние
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats возвращает статистику circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state,
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_failure":      cb.lastFailureTime,
		"last_state_change": cb.lastStateChange,
	}
}

// ForceOpen принудительно открывает circuit breaker
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = Open
	cb.lastStateChange = time.Now()
}

// ForceClose принудительно закрывает circuit breaker
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = Closed
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
}

// Reset сбрасывает circuit breaker в начальное состояние
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = Closed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastFailureTime = time.Time{}
	cb.lastStateChange = time.Now()
}
