package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reazonvan/LootBay/internal/api"
	"github.com/reazonvan/LootBay/internal/auth"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/database"
	"github.com/reazonvan/LootBay/internal/middleware"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/logger"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfigForService(".", "product-service")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация логгера
	logger := logger.NewLogger(cfg.Server.Mode == "debug")

	// Инициализация базы данных
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}

	// Инициализация Redis
	redisClient, err := database.NewRedisConnection(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", "error", err)
	}

	// Установка режима Gin
	gin.SetMode(cfg.Server.Mode)

	// Создание роутера
	router := gin.New()

	// Добавление middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(redisClient))
	router.Use(middleware.Security())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "product-service",
		})
	})

	// Инициализация JWT Manager
	jwtManager := auth.NewJWTManager(&cfg.JWT, redisClient)

	// Инициализация репозиториев
	productRepo := repository.NewProductRepository(db)
	gameRepo := repository.NewGameRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// Инициализация сервисов
	productService := service.NewProductService(productRepo, gameRepo, categoryRepo, logger)
	gameService := service.NewGameService(gameRepo, logger)
	categoryService := service.NewCategoryService(categoryRepo, logger)

	// Инициализация хендлеров
	productHandler := api.NewProductHandler(productService, logger)
	gameHandler := api.NewGameHandler(gameService, logger)
	categoryHandler := api.NewCategoryHandler(categoryService, logger)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Публичные маршруты для продуктов
		v1.GET("/products", productHandler.GetProducts)
		v1.GET("/products/:id", productHandler.GetProduct)
		v1.GET("/products/search", productHandler.SearchProducts)

		// Публичные маршруты для игр
		v1.GET("/games", gameHandler.GetGames)
		v1.GET("/games/:slug", gameHandler.GetGame)

		// Публичные маршруты для категорий
		v1.GET("/categories", categoryHandler.GetCategories)
		v1.GET("/categories/:slug", categoryHandler.GetCategory)

		// Защищенные маршруты
		auth := v1.Group("")
		auth.Use(middleware.Auth(jwtManager))
		{
			// Управление продуктами (только для продавцов)
			auth.POST("/products", productHandler.CreateProduct)
			auth.PUT("/products/:id", productHandler.UpdateProduct)
			auth.DELETE("/products/:id", productHandler.DeleteProduct)

			// Управление играми (только для админов)
			auth.POST("/games", gameHandler.CreateGame)
			auth.PUT("/games/:id", gameHandler.UpdateGame)
			auth.DELETE("/games/:id", gameHandler.DeleteGame)

			// Управление категориями (только для админов)
			auth.POST("/categories", categoryHandler.CreateCategory)
			auth.PUT("/categories/:id", categoryHandler.UpdateCategory)
			auth.DELETE("/categories/:id", categoryHandler.DeleteCategory)
		}
	}

	// Создание HTTP сервера
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cfg.Server.Host, "8082"),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: int(cfg.Server.MaxBodySize),
	}

	// Запуск сервера в отдельной горутине
	go func() {
		logger.Info("Starting Product Service", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Закрытие соединений с базами данных
	if err := database.ClosePostgres(db); err != nil {
		logger.Error("Error closing Postgres connection", "error", err)
	}

	if err := database.CloseRedis(redisClient); err != nil {
		logger.Error("Error closing Redis connection", "error", err)
	}

	// Graceful shutdown сервера
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
