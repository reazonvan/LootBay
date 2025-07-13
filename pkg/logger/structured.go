package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// StructuredLogger структурированный логгер
type StructuredLogger struct {
	logger *zap.Logger
}

// LogLevel уровень логирования
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// LogEntry структура записи лога
type LogEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	Level         string                 `json:"level"`
	Message       string                 `json:"message"`
	Service       string                 `json:"service"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	Method        string                 `json:"method,omitempty"`
	Path          string                 `json:"path,omitempty"`
	StatusCode    int                    `json:"status_code,omitempty"`
	Duration      string                 `json:"duration,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Stack         string                 `json:"stack,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
}

// NewStructuredLogger создает новый структурированный логгер
func NewStructuredLogger(serviceName string, level LogLevel) *StructuredLogger {
	// Настройка уровня логирования
	var zapLevel zapcore.Level
	switch level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	case FatalLevel:
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Конфигурация логгера
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stack"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Добавляем поля по умолчанию
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
		"version": os.Getenv("APP_VERSION"),
		"env":     os.Getenv("ENV"),
	}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	return &StructuredLogger{
		logger: logger,
	}
}

// WithContext создает логгер с контекстом
func (l *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	fields := []zap.Field{}

	// Добавляем correlation ID из контекста
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		fields = append(fields, zap.String("correlation_id", correlationID.(string)))
	}

	// Добавляем user ID из контекста
	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.String("user_id", userID.(string)))
	}

	// Добавляем request ID из контекста
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	return &StructuredLogger{
		logger: l.logger.With(fields...),
	}
}

// WithFields создает логгер с дополнительными полями
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return &StructuredLogger{
		logger: l.logger.With(zapFields...),
	}
}

// Debug логирует сообщение уровня debug
func (l *StructuredLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.logWithFields(l.logger.Debug, msg, fields...)
}

// Info логирует сообщение уровня info
func (l *StructuredLogger) Info(msg string, fields ...map[string]interface{}) {
	l.logWithFields(l.logger.Info, msg, fields...)
}

// Warn логирует сообщение уровня warn
func (l *StructuredLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.logWithFields(l.logger.Warn, msg, fields...)
}

// Error логирует сообщение уровня error
func (l *StructuredLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	if err != nil {
		if fields == nil {
			fields = []map[string]interface{}{}
		}
		fields = append(fields, map[string]interface{}{
			"error": err.Error(),
		})
	}
	l.logWithFields(l.logger.Error, msg, fields...)
}

// Fatal логирует сообщение уровня fatal и завершает программу
func (l *StructuredLogger) Fatal(msg string, fields ...map[string]interface{}) {
	l.logWithFields(l.logger.Fatal, msg, fields...)
}

// logWithFields вспомогательный метод для логирования с полями
func (l *StructuredLogger) logWithFields(logFunc func(string, ...zap.Field), msg string, fields ...map[string]interface{}) {
	zapFields := make([]zap.Field, 0)

	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			zapFields = append(zapFields, zap.Any(key, value))
		}
	}

	logFunc(msg, zapFields...)
}

// LogHTTPRequest логирует HTTP запрос
func (l *StructuredLogger) LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, fields ...map[string]interface{}) {
	requestFields := map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration":    duration.String(),
	}

	// Добавляем дополнительные поля
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			requestFields[key] = value
		}
	}

	l.WithContext(ctx).Info("HTTP Request", requestFields)
}

// LogDatabaseQuery логирует запрос к базе данных
func (l *StructuredLogger) LogDatabaseQuery(ctx context.Context, query string, duration time.Duration, rowsAffected int64, err error) {
	fields := map[string]interface{}{
		"query":         query,
		"duration":      duration.String(),
		"rows_affected": rowsAffected,
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithContext(ctx).Error("Database Query Error", err, fields)
	} else {
		l.WithContext(ctx).Debug("Database Query", fields)
	}
}

// LogCacheOperation логирует операции с кэшем
func (l *StructuredLogger) LogCacheOperation(ctx context.Context, operation, key string, duration time.Duration, hit bool, err error) {
	fields := map[string]interface{}{
		"operation": operation,
		"key":       key,
		"duration":  duration.String(),
		"hit":       hit,
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithContext(ctx).Error("Cache Operation Error", err, fields)
	} else {
		l.WithContext(ctx).Debug("Cache Operation", fields)
	}
}

// LogBusinessEvent логирует бизнес-события
func (l *StructuredLogger) LogBusinessEvent(ctx context.Context, eventType, eventName string, data map[string]interface{}) {
	fields := map[string]interface{}{
		"event_type": eventType,
		"event_name": eventName,
	}

	if data != nil {
		for key, value := range data {
			fields[key] = value
		}
	}

	l.WithContext(ctx).Info("Business Event", fields)
}

// LogSecurityEvent логирует события безопасности
func (l *StructuredLogger) LogSecurityEvent(ctx context.Context, eventType, description string, severity string, data map[string]interface{}) {
	fields := map[string]interface{}{
		"event_type":  eventType,
		"description": description,
		"severity":    severity,
	}

	if data != nil {
		for key, value := range data {
			fields[key] = value
		}
	}

	l.WithContext(ctx).Warn("Security Event", fields)
}

// GenerateCorrelationID генерирует новый correlation ID
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// WithCorrelationID добавляет correlation ID в контекст
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, "correlation_id", correlationID)
}

// WithUserID добавляет user ID в контекст
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}

// WithRequestID добавляет request ID в контекст
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

// Sync синхронизирует буферы логгера
func (l *StructuredLogger) Sync() error {
	return l.logger.Sync()
}
