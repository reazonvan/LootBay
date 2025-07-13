package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/middleware"
	"github.com/reazonvan/LootBay/pkg/logger"
	"golang.org/x/time/rate"
)

// ServiceInfo информация о микросервисе
type ServiceInfo struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	Port    string `json:"port"`
	Version string `json:"version"`
	Health  string `json:"health"`
}

// Gateway основной API Gateway
type Gateway struct {
	config          *config.Config
	logger          *logger.StructuredLogger
	router          *gin.Engine
	services        map[string]*ServiceInfo
	rateLimiters    map[string]*rate.Limiter
	circuitBreakers map[string]*CircuitBreaker
}

// NewGateway создает новый API Gateway
func NewGateway(cfg *config.Config, log *logger.StructuredLogger) *Gateway {
	gateway := &Gateway{
		config:          cfg,
		logger:          log,
		router:          gin.New(),
		services:        make(map[string]*ServiceInfo),
		rateLimiters:    make(map[string]*rate.Limiter),
		circuitBreakers: make(map[string]*CircuitBreaker),
	}

	// Инициализация сервисов
	gateway.initServices()

	// Настройка middleware
	gateway.setupMiddleware()

	// Настройка роутов
	gateway.setupRoutes()

	return gateway
}

// initServices инициализирует информацию о микросервисах
func (g *Gateway) initServices() {
	g.services = map[string]*ServiceInfo{
		"user-service": {
			Name:    "user-service",
			Host:    "localhost",
			Port:    "8081",
			Version: "v1",
			Health:  "/health",
		},
		"product-service": {
			Name:    "product-service",
			Host:    "localhost",
			Port:    "8082",
			Version: "v1",
			Health:  "/health",
		},
		"order-service": {
			Name:    "order-service",
			Host:    "localhost",
			Port:    "8083",
			Version: "v1",
			Health:  "/health",
		},
		"payment-service": {
			Name:    "payment-service",
			Host:    "localhost",
			Port:    "8084",
			Version: "v1",
			Health:  "/health",
		},
		"chat-service": {
			Name:    "chat-service",
			Host:    "localhost",
			Port:    "8085",
			Version: "v1",
			Health:  "/health",
		},
		"notification-service": {
			Name:    "notification-service",
			Host:    "localhost",
			Port:    "8086",
			Version: "v1",
			Health:  "/health",
		},
	}

	// Инициализация rate limiters (10 запросов в секунду на сервис)
	for serviceName := range g.services {
		g.rateLimiters[serviceName] = rate.NewLimiter(rate.Limit(10), 30)
		g.circuitBreakers[serviceName] = NewCircuitBreaker(5, 10*time.Second, 30*time.Second)
	}
}

// setupMiddleware настраивает middleware
func (g *Gateway) setupMiddleware() {
	// Recovery middleware
	g.router.Use(gin.Recovery())

	// CORS middleware
	g.router.Use(middleware.CORS())

	// Request logging
	g.router.Use(g.requestLogger())

	// Rate limiting
	g.router.Use(g.rateLimiter())

	// Authentication (для защищенных эндпоинтов)
	g.router.Use(g.authMiddleware())
}

// setupRoutes настраивает роуты API Gateway
func (g *Gateway) setupRoutes() {
	// Health check
	g.router.GET("/health", g.healthCheck)

	// API версионирование
	v1 := g.router.Group("/api/v1")
	{
		// Публичные эндпоинты
		public := v1.Group("")
		{
			public.POST("/auth/register", g.proxyToService("user-service", "/auth/register"))
			public.POST("/auth/login", g.proxyToService("user-service", "/auth/login"))
			public.GET("/products", g.proxyToService("product-service", "/products"))
			public.GET("/products/:id", g.proxyToService("product-service", "/products/:id"))
		}

		// Защищенные эндпоинты
		protected := v1.Group("")
		protected.Use(g.requireAuth())
		{
			// User endpoints
			protected.GET("/users/profile", g.proxyToService("user-service", "/users/profile"))
			protected.PUT("/users/profile", g.proxyToService("user-service", "/users/profile"))

			// Product endpoints
			protected.POST("/products", g.proxyToService("product-service", "/products"))
			protected.PUT("/products/:id", g.proxyToService("product-service", "/products/:id"))
			protected.DELETE("/products/:id", g.proxyToService("product-service", "/products/:id"))

			// Order endpoints
			protected.GET("/orders", g.proxyToService("order-service", "/orders"))
			protected.POST("/orders", g.proxyToService("order-service", "/orders"))
			protected.GET("/orders/:id", g.proxyToService("order-service", "/orders/:id"))

			// Payment endpoints
			protected.POST("/payments", g.proxyToService("payment-service", "/payments"))
			protected.GET("/payments/:id", g.proxyToService("payment-service", "/payments/:id"))

			// Chat endpoints
			protected.GET("/chats", g.proxyToService("chat-service", "/chats"))
			protected.POST("/chats", g.proxyToService("chat-service", "/chats"))
			protected.GET("/chats/:id/messages", g.proxyToService("chat-service", "/chats/:id/messages"))
			protected.POST("/chats/:id/messages", g.proxyToService("chat-service", "/chats/:id/messages"))
		}
	}

	// WebSocket endpoints
	g.router.GET("/ws/chat", g.proxyWebSocket("chat-service", "/ws/chat"))

	// Admin endpoints
	admin := v1.Group("/admin")
	admin.Use(g.requireAdmin())
	{
		admin.GET("/users", g.proxyToService("user-service", "/admin/users"))
		admin.GET("/orders", g.proxyToService("order-service", "/admin/orders"))
		admin.GET("/payments", g.proxyToService("payment-service", "/admin/payments"))
	}
}

// healthCheck проверка здоровья API Gateway
func (g *Gateway) healthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"services":  g.getServicesHealth(),
	}

	c.JSON(http.StatusOK, health)
}

// getServicesHealth проверяет здоровье всех сервисов
func (g *Gateway) getServicesHealth() map[string]string {
	health := make(map[string]string)

	for serviceName, service := range g.services {
		// Проверяем circuit breaker
		if g.circuitBreakers[serviceName].IsOpen() {
			health[serviceName] = "circuit_open"
			continue
		}

		// Проверяем доступность сервиса
		url := fmt.Sprintf("http://%s:%s%s", service.Host, service.Port, service.Health)
		resp, err := http.Get(url)
		if err != nil {
			health[serviceName] = "unavailable"
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			health[serviceName] = "healthy"
		} else {
			health[serviceName] = "unhealthy"
		}
	}

	return health
}

// proxyToService проксирует запрос к микросервису
func (g *Gateway) proxyToService(serviceName, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Проверяем circuit breaker
		if !g.circuitBreakers[serviceName].Allow() {
			g.logger.Error("Circuit breaker is open", fmt.Errorf("service %s is unavailable", serviceName), map[string]interface{}{
				"service": serviceName,
				"path":    path,
			})
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service temporarily unavailable"})
			return
		}

		// Получаем информацию о сервисе
		service, exists := g.services[serviceName]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Строим URL для проксирования
		targetURL := fmt.Sprintf("http://%s:%s%s", service.Host, service.Port, path)

		// Создаем HTTP клиент с таймаутом
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		// Создаем запрос
		req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			g.logger.Error("Failed to create request", err, map[string]interface{}{
				"service": serviceName,
				"path":    path,
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Копируем заголовки
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Добавляем заголовки для трейсинга
		req.Header.Set("X-Correlation-ID", c.GetString("correlation_id"))
		req.Header.Set("X-User-ID", c.GetString("user_id"))

		// Выполняем запрос
		resp, err := client.Do(req)
		if err != nil {
			g.circuitBreakers[serviceName].RecordFailure()
			g.logger.Error("Failed to proxy request", err, map[string]interface{}{
				"service": serviceName,
				"path":    path,
				"url":     targetURL,
			})
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
			return
		}
		defer resp.Body.Close()

		// Копируем заголовки ответа
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Копируем статус код
		c.Status(resp.StatusCode)

		// Копируем тело ответа
		if _, err := c.Writer.Write([]byte{}); err != nil {
			g.logger.Error("Failed to copy response body", err, map[string]interface{}{
				"service": serviceName,
				"path":    path,
			})
		}

		// Записываем метрики
		duration := time.Since(start)
		g.logger.Info("Request proxied", map[string]interface{}{
			"service":  serviceName,
			"path":     path,
			"duration": duration.String(),
			"status":   resp.StatusCode,
		})

		// Записываем успешный запрос в circuit breaker
		g.circuitBreakers[serviceName].RecordSuccess()
	}
}

// proxyWebSocket проксирует WebSocket соединения
func (g *Gateway) proxyWebSocket(serviceName, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем информацию о сервисе
		service, exists := g.services[serviceName]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Строим URL для WebSocket
		_ = fmt.Sprintf("ws://%s:%s%s", service.Host, service.Port, path)

		// Здесь должна быть логика проксирования WebSocket
		// Для простоты возвращаем ошибку
		c.JSON(http.StatusNotImplemented, gin.H{"error": "WebSocket proxying not implemented"})
	}
}

// requestLogger middleware для логирования запросов
func (g *Gateway) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Генерируем correlation ID
		correlationID := logger.GenerateCorrelationID()
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		// Добавляем correlation ID в контекст
		ctx := logger.WithCorrelationID(c.Request.Context(), correlationID)
		c.Request = c.Request.WithContext(ctx)

		// Обрабатываем запрос
		c.Next()

		// Логируем запрос
		duration := time.Since(start)
		g.logger.LogHTTPRequest(ctx, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration, map[string]interface{}{
			"user_agent": c.Request.UserAgent(),
			"remote_ip":  c.ClientIP(),
		})
	}
}

// rateLimiter middleware для rate limiting
func (g *Gateway) rateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Определяем сервис по пути
		serviceName := g.getServiceFromPath(c.Request.URL.Path)
		if serviceName == "" {
			c.Next()
			return
		}

		// Проверяем rate limiter
		limiter, exists := g.rateLimiters[serviceName]
		if !exists {
			c.Next()
			return
		}

		if !limiter.Allow() {
			g.logger.Warn("Rate limit exceeded", map[string]interface{}{
				"service": serviceName,
				"ip":      c.ClientIP(),
			})
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// authMiddleware middleware для аутентификации
func (g *Gateway) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем публичные эндпоинты
		if g.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Проверяем JWT токен
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Убираем префикс "Bearer "
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		// Валидируем токен (здесь должна быть реальная валидация)
		userID, err := g.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Добавляем user ID в контекст
		c.Set("user_id", userID)
		c.Header("X-User-ID", userID)

		// Добавляем user ID в контекст запроса
		ctx := logger.WithUserID(c.Request.Context(), userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// requireAuth middleware для обязательной аутентификации
func (g *Gateway) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// requireAdmin middleware для административных эндпоинтов
func (g *Gateway) requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем роль пользователя (здесь должна быть реальная проверка)
		userRole := c.GetString("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// getServiceFromPath определяет сервис по пути запроса
func (g *Gateway) getServiceFromPath(path string) string {
	// Простая логика определения сервиса по пути
	if strings.Contains(path, "/auth/") || strings.Contains(path, "/users/") {
		return "user-service"
	}
	if strings.Contains(path, "/products") {
		return "product-service"
	}
	if strings.Contains(path, "/orders") {
		return "order-service"
	}
	if strings.Contains(path, "/payments") {
		return "payment-service"
	}
	if strings.Contains(path, "/chats") || strings.Contains(path, "/ws/") {
		return "chat-service"
	}
	return ""
}

// isPublicEndpoint проверяет, является ли эндпоинт публичным
func (g *Gateway) isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/products",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// validateToken валидирует JWT токен
func (g *Gateway) validateToken(token string) (string, error) {
	// TODO: Инициализировать JWTManager в Gateway
	// Временно возвращаем ошибку для безопасности
	return "", errors.New("token validation not implemented")
}

// Run запускает API Gateway
func (g *Gateway) Run() error {
	addr := fmt.Sprintf("%s:%s", g.config.Server.Host, g.config.Server.Port)
	g.logger.Info("Starting API Gateway", map[string]interface{}{
		"address": addr,
		"mode":    g.config.Server.Mode,
	})

	return g.router.Run(addr)
}
