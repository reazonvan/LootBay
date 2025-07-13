package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/gateway"
	"github.com/reazonvan/LootBay/pkg/logger"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем логгер
	logger := logger.NewStructuredLogger("api-gateway", logger.InfoLevel)

	// Создаем API Gateway
	gateway := gateway.NewGateway(cfg, logger)

	// Канал для graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем API Gateway в горутине
	go func() {
		logger.Info("Starting API Gateway", map[string]interface{}{
			"port": cfg.Server.Port,
		})

		if err := gateway.Run(); err != nil {
			logger.Error("Failed to start API Gateway", err, nil)
			os.Exit(1)
		}
	}()

	// Ждем сигнала для graceful shutdown
	<-done
	logger.Info("Shutting down API Gateway...", nil)

	// Здесь можно добавить graceful shutdown логику
	// Например, закрытие соединений с базой данных, кэшем и т.д.
}
