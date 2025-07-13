package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/reazonvan/LootBay/internal/middleware"
)

var serviceMap = map[string]string{
	"/api/v1/auth":          "http://user-service:8081",
	"/api/v1/users":         "http://user-service:8081",
	"/api/v1/config":        "http://user-service:8081",
	"/api/v1/products":      "http://product-service:8082",
	"/api/v1/orders":        "http://order-service:8083",
	"/api/v1/payments":      "http://payment-service:8084",
	"/api/v1/balance":       "http://payment-service:8084",
	"/api/v1/chat":          "http://chat-service:8085",
	"/api/v1/conversations": "http://chat-service:8085",
	"/api/v1/notifications": "http://notification-service:8086",
	"/api/v1/games":         "http://product-service:8082",
	"/api/v1/categories":    "http://product-service:8082",
	"/health":               "http://user-service:8081",
}

func main() {
	router := gin.New()

	// Настройка CORS в первую очередь
	router.Use(middleware.CORS())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Отдельный health-check для самого шлюза
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Регистрируем прокси для каждого сервиса
	for prefix, target := range serviceMap {
		targetURL, err := url.Parse(target)
		if err != nil {
			log.Fatalf("Error parsing target URL for prefix %s: %v", prefix, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Настраиваем Director
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.Host = targetURL.Host
		}

		// Создаем группу маршрутов для префикса
		group := router.Group(prefix)
		group.Any("/*proxyPath", func(c *gin.Context) {
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	log.Println("API Gateway is running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run API Gateway: %v", err)
	}
}
