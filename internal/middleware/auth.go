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
		c.Set("user_roles", claims.Roles)
		c.Set("user_permissions", claims.Permissions)
		c.Set("token", token)

		c.Next()
	}
}

// AdminOnly middleware для проверки прав администратора (обновленный)
func AdminOnly() gin.HandlerFunc {
	return RequireAnyRole("ADMIN_SUPPORT", "ADMIN_MODERATION", "ADMIN_GAMES", "OWNER")
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

// RequireRole middleware для проверки конкретной роли
func RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("user_roles")
		if !exists {
			response.Forbidden(c)
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			response.Forbidden(c)
			c.Abort()
			return
		}

		// Проверяем, есть ли нужная роль
		for _, role := range userRoles {
			if role == roleName {
				c.Next()
				return
			}
		}

		response.Forbidden(c)
		c.Abort()
	}
}

// RequireAnyRole middleware для проверки любой из ролей
func RequireAnyRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("user_roles")
		if !exists {
			response.Forbidden(c)
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			response.Forbidden(c)
			c.Abort()
			return
		}

		// Проверяем, есть ли любая из нужных ролей
		for _, userRole := range userRoles {
			for _, requiredRole := range roleNames {
				if userRole == requiredRole {
					c.Next()
					return
				}
			}
		}

		response.Forbidden(c)
		c.Abort()
	}
}

// RequirePermission middleware для проверки конкретного разрешения
func RequirePermission(permissionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("user_permissions")
		if !exists {
			response.Forbidden(c)
			c.Abort()
			return
		}

		userPermissions, ok := permissions.([]string)
		if !ok {
			response.Forbidden(c)
			c.Abort()
			return
		}

		// Проверяем, есть ли нужное разрешение
		for _, permission := range userPermissions {
			if permission == permissionName {
				c.Next()
				return
			}
		}

		response.Forbidden(c)
		c.Abort()
	}
}

// RequireAnyPermission middleware для проверки любого из разрешений
func RequireAnyPermission(permissionNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("user_permissions")
		if !exists {
			response.Forbidden(c)
			c.Abort()
			return
		}

		userPermissions, ok := permissions.([]string)
		if !ok {
			response.Forbidden(c)
			c.Abort()
			return
		}

		// Проверяем, есть ли любое из нужных разрешений
		for _, userPermission := range userPermissions {
			for _, requiredPermission := range permissionNames {
				if userPermission == requiredPermission {
					c.Next()
					return
				}
			}
		}

		response.Forbidden(c)
		c.Abort()
	}
}

// OwnerOnly middleware для проверки роли владельца
func OwnerOnly() gin.HandlerFunc {
	return RequireRole("OWNER")
}

// GamesAdminOnly middleware для проверки прав на управление играми
func GamesAdminOnly() gin.HandlerFunc {
	return RequireAnyRole("ADMIN_GAMES", "OWNER")
}

// ModerationAdminOnly middleware для проверки прав на модерацию
func ModerationAdminOnly() gin.HandlerFunc {
	return RequireAnyRole("ADMIN_MODERATION", "OWNER")
}

// SupportAdminOnly middleware для проверки прав на техподдержку
func SupportAdminOnly() gin.HandlerFunc {
	return RequireAnyRole("ADMIN_SUPPORT", "OWNER")
}
