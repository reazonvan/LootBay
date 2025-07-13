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
	cfg, err := config.LoadConfigForService(".", "payment-service")
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
			"service":   "payment-service",
		})
	})

	// Инициализация репозиториев
	paymentRepo := repository.NewPaymentRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Инициализация JWT Manager
	jwtManager := auth.NewJWTManager(&cfg.JWT, redisClient)

	// Инициализация сервисов
	paymentService := service.NewPaymentService(paymentRepo, orderRepo, transactionRepo, userRepo, log)
	stripeService := service.NewStripeService(cfg.Payment.Stripe.SecretKey, log)
	paypalService := service.NewPaypalService(cfg.Payment.PayPal.ClientID, cfg.Payment.PayPal.ClientSecret, log)

	// Инициализация хендлеров
	paymentHandler := api.NewPaymentHandler(paymentService, stripeService, paypalService, log)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Публичные маршруты для webhooks
		v1.POST("/payments/webhook/stripe", paymentHandler.StripeWebhook)
		v1.POST("/payments/webhook/paypal", paymentHandler.PaypalWebhook)

		// Защищенные маршруты
		auth := v1.Group("")
		auth.Use(middleware.Auth(jwtManager))
		{
			// Управление платежами
			auth.POST("/payments", paymentHandler.CreatePayment)
			auth.GET("/payments/:id", paymentHandler.GetPayment)
			auth.GET("/payments", paymentHandler.GetPayments)
			auth.POST("/payments/:id/capture", paymentHandler.CapturePayment)
			auth.POST("/payments/:id/refund", paymentHandler.RefundPayment)

			// Управление балансом
			auth.POST("/balance/deposit", paymentHandler.DepositToBalance)
			auth.POST("/balance/withdraw", paymentHandler.WithdrawFromBalance)
			auth.GET("/balance/transactions", paymentHandler.GetTransactions)
		}
	}

	// Создание HTTP сервера
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cfg.Server.Host, "8084"),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: int(cfg.Server.MaxBodySize),
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Info("Starting Payment Service", "address", server.Addr)
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
