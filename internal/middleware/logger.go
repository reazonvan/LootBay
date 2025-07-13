package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/reazonvan/LootBay/pkg/logger"
)

// Logger middleware для логирования запросов
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Генерируем trace ID для запроса
		traceID := uuid.New().String()
		c.Set("trace_id", traceID)

		// Добавляем trace ID в контекст
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "trace_id", traceID)
		c.Request = c.Request.WithContext(ctx)

		// Начинаем отсчет времени
		start := time.Now()

		// Обрабатываем запрос
		c.Next()

		// Логируем результат
		duration := time.Since(start)
		log.Info("Request processed",
			"trace_id", traceID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration.String(),
			"ip", c.ClientIP(),
			"user_agent", c.GetHeader("User-Agent"),
			"error", c.Errors.String(),
		)
	}
}

// Recovery middleware для обработки паник
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID, _ := c.Get("trace_id")
				log.Error("Panic recovered",
					"trace_id", traceID,
					"error", err,
					"path", c.Request.URL.Path,
				)

				c.JSON(500, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "Внутренняя ошибка сервера",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
