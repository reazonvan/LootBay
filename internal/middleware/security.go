package middleware

import (
	"github.com/gin-gonic/gin"
)

// Security middleware для установки заголовков безопасности
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Защита от clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Защита от XSS
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy (более строгая)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")

		// Strict Transport Security (для HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
