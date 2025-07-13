package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit middleware для ограничения частоты запросов
func RateLimit(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Определяем ключ для rate limiting
		var key string
		userID, exists := c.Get("user_id")
		if exists {
			key = fmt.Sprintf("rate_limit:user:%v", userID)
		} else {
			key = fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
		}

		// Устанавливаем лимиты
		limit := int64(100)   // 100 запросов
		window := time.Minute // в минуту

		// Проверяем количество запросов
		pipe := redis.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err := pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		count := incr.Val()
		if count > limit {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

			c.JSON(429, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Превышен лимит запросов",
				},
			})
			c.Abort()
			return
		}

		// Добавляем заголовки с информацией о лимитах
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		c.Next()
	}
}
