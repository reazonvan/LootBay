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
	cfg, err := config.LoadConfigForService(".", "notification-service")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация логгера
	log := logger.NewLogger(cfg.Server.Mode == "debug")

	// Инициализация базы данных
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}

	// Инициализация Redis
	redisClient, err := database.NewRedisConnection(&cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis", "error", err)
	}

	// Инициализация JWT менеджера
	jwtManager := auth.NewJWTManager(&cfg.JWT, redisClient)

	// Установка режима Gin
	gin.SetMode(cfg.Server.Mode)

	// Создание роутера
	router := gin.New()

	// Добавление middleware
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery(log))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(redisClient))
	router.Use(middleware.Security())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "notification-service",
		})
	})

	// Инициализация репозиториев
	notificationRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Инициализация сервисов
	emailService := service.NewEmailService(&cfg.Email, log)
	telegramService := service.NewTelegramService(&cfg.Telegram, log)
	pushService := service.NewPushService(log)
	notificationService := service.NewNotificationService(
		notificationRepo,
		userRepo,
		emailService,
		telegramService,
		pushService,
		*log,
	)

	// Инициализация хендлеров
	notificationHandler := api.NewNotificationHandler(notificationService, log)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Защищенные маршруты
		authRoutes := v1.Group("")
		authRoutes.Use(middleware.Auth(jwtManager))
		{
			// Управление уведомлениями
			authRoutes.GET("/notifications", notificationHandler.GetNotifications)
			authRoutes.PUT("/notifications/:id/read", notificationHandler.MarkAsRead)
			authRoutes.DELETE("/notifications/:id", notificationHandler.DeleteNotification)
			authRoutes.POST("/notifications/send", notificationHandler.SendNotification)

			// Настройки уведомлений
			authRoutes.GET("/notifications/settings", notificationHandler.GetSettings)
			authRoutes.PUT("/notifications/settings", notificationHandler.UpdateSettings)
		}
	}

	// Создание HTTP сервера
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cfg.Server.Host, "8086"),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: int(cfg.Server.MaxBodySize),
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Info("Starting Notification Service", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Закрытие соединений с базами данных
	if err := database.ClosePostgres(db); err != nil {
		log.Error("Error closing Postgres connection", "error", err)
	}

	if err := database.CloseRedis(redisClient); err != nil {
		log.Error("Error closing Redis connection", "error", err)
	}

	// Graceful shutdown сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited")
}
