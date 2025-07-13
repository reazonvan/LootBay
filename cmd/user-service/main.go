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
	cfg, err := config.LoadConfigForService(".", "user-service")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация логгера
	appLogger := logger.NewLogger(cfg.Server.Mode == "debug")

	// Инициализация базы данных
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}

	// Инициализация Redis
	redisClient, err := database.NewRedisConnection(&cfg.Redis)
	if err != nil {
		appLogger.Fatal("Failed to connect to Redis", "error", err)
	}

	// Установка режима Gin
	gin.SetMode(cfg.Server.Mode)

	// Создание роутера
	router := gin.New()

	// Добавление middleware
	router.Use(middleware.Logger(appLogger))
	router.Use(middleware.Recovery(appLogger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(redisClient))
	router.Use(middleware.Security())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "debug health endpoint works - роуты работают!"})
	})

	// Debug endpoint в корне
	router.GET("/debug", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "debug root endpoint works"})
	})

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	// Инициализация сервисов
	roleService := service.NewRoleService(roleRepo, userRepo, appLogger)
	userService := service.NewUserService(userRepo, roleService, &cfg.JWT, redisClient, appLogger)

	// Инициализация JWT менеджера
	jwtManager := auth.NewJWTManager(&cfg.JWT, redisClient)

	// Инициализация хендлеров
	userHandler := api.NewUserHandler(userService, jwtManager, appLogger)
	// roleHandler := api.NewRoleHandler(roleService, appLogger)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Публичные маршруты
		v1.POST("/auth/register", userHandler.Register)
		v1.POST("/auth/login", userHandler.Login)
		v1.POST("/auth/refresh", userHandler.RefreshToken)
		v1.POST("/auth/forgot-password", userHandler.ForgotPassword)
		v1.POST("/auth/reset-password", userHandler.ResetPassword)
		v1.GET("/auth/verify-email", userHandler.VerifyEmail)

		// Конфигурация (например, политика паролей)
		v1.GET("/config/password-policy", userHandler.PasswordPolicy)

		// Debug: простой роут для проверки
		v1.GET("/debug-test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "debug endpoint works"})
		})

		// Роли маршруты (только для OWNER)
		v1.GET("/roles/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "roles test endpoint works"})
		})
		v1.GET("/roles", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "roles endpoint works"})
		})

		// Получение публичной информации о пользователе
		v1.GET("/users/:id", userHandler.GetUser)

		// Защищенные маршруты
		authRoutes := v1.Group("")
		authRoutes.Use(middleware.Auth(jwtManager))
		{
			authRoutes.POST("/auth/logout", userHandler.Logout)
			authRoutes.GET("/users/profile", userHandler.GetProfile)
			authRoutes.PUT("/users/profile", userHandler.UpdateProfile)
		}
	}

	// Создание HTTP сервера
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: int(cfg.Server.MaxBodySize),
	}

	// Запуск сервера в отдельной горутине
	go func() {
		appLogger.Info("Starting User Service", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Закрытие соединений с базами данных
	if err := database.ClosePostgres(db); err != nil {
		appLogger.Error("Error closing Postgres connection", "error", err)
	}

	if err := database.CloseRedis(redisClient); err != nil {
		appLogger.Error("Error closing Redis connection", "error", err)
	}

	// Graceful shutdown сервера
	if err := server.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server exited")
}
