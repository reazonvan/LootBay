package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/reazonvan/LootBay/internal/auth"
	"github.com/reazonvan/LootBay/pkg/response"
)

// Auth middleware для проверки JWT токена
func Auth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// Проверяем формат Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		token := parts[1]

		// Валидируем токен
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// TODO: Проверить сессию в Redis

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("is_seller", claims.IsSeller)
		c.Set("token", token)

		c.Next()
	}
}

// AdminOnly middleware для проверки прав администратора
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Реализовать проверку прав администратора
		isSeller, exists := c.Get("is_seller")
		if !exists || !isSeller.(bool) {
			response.Forbidden(c)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SellerOnly middleware для проверки прав продавца
func SellerOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isSeller, exists := c.Get("is_seller")
		if !exists || !isSeller.(bool) {
			response.Forbidden(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
