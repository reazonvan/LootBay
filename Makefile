.PHONY: help build run test clean docker-build docker-up docker-down migrate

# Переменные
GO=go
GOFLAGS=-v
DOCKER_COMPOSE=docker-compose

# Цвета для вывода
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Показать эту справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Сборка
build: ## Собрать все сервисы
	@echo "$(GREEN)Building services...$(NC)"
	$(GO) build $(GOFLAGS) -o bin/api-gateway ./cmd/api-gateway
	$(GO) build $(GOFLAGS) -o bin/user-service ./cmd/user-service
	@echo "$(GREEN)Build complete!$(NC)"

build-gateway: ## Собрать API Gateway
	$(GO) build $(GOFLAGS) -o bin/api-gateway ./cmd/api-gateway

build-user: ## Собрать User Service
	$(GO) build $(GOFLAGS) -o bin/user-service ./cmd/user-service

# Запуск
run-gateway: ## Запустить API Gateway
	$(GO) run ./cmd/api-gateway/main.go

run-user: ## Запустить User Service
	$(GO) run ./cmd/user-service/main.go

# Docker
docker-build: ## Собрать Docker образы
	@echo "$(GREEN)Building Docker images...$(NC)"
	$(DOCKER_COMPOSE) build

docker-up: ## Запустить все сервисы в Docker
	@echo "$(GREEN)Starting services...$(NC)"
	$(DOCKER_COMPOSE) up -d

docker-down: ## Остановить все сервисы в Docker
	@echo "$(RED)Stopping services...$(NC)"
	$(DOCKER_COMPOSE) down

docker-logs: ## Показать логи всех сервисов
	$(DOCKER_COMPOSE) logs -f

docker-ps: ## Показать статус контейнеров
	$(DOCKER_COMPOSE) ps

# База данных
migrate-up: ## Применить миграции
	@echo "$(GREEN)Applying migrations...$(NC)"
	$(GO) run ./cmd/migrate/main.go up

migrate-down: ## Откатить миграции
	@echo "$(RED)Rolling back migrations...$(NC)"
	$(GO) run ./cmd/migrate/main.go down

migrate-status: ## Показать статус миграций
	@echo "$(GREEN)Migration status...$(NC)"
	$(GO) run ./cmd/migrate/main.go status

migrate-create: ## Создать новую миграцию
	@echo "$(GREEN)Creating new migration...$(NC)"
	@read -p "Enter migration name: " name; \
	$(GO) run ./cmd/migrate/main.go create $$name

# Тестирование
test: ## Запустить тесты
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test ./... -v -cover

test-unit: ## Запустить unit тесты
	@echo "$(GREEN)Running unit tests...$(NC)"
	$(GO) test ./internal/service -v -cover

test-integration: ## Запустить интеграционные тесты
	@echo "$(GREEN)Running integration tests...$(NC)"
	$(GO) test ./internal/service -v -cover -tags=integration

test-coverage: ## Запустить тесты с покрытием
	$(GO) test ./... -v -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

test-benchmark: ## Запустить benchmark тесты
	@echo "$(GREEN)Running benchmark tests...$(NC)"
	$(GO) test ./... -bench=. -benchmem

# Линтинг
lint: ## Запустить линтер
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run ./...

# Утилиты
clean: ## Очистить бинарные файлы и кэш
	@echo "$(RED)Cleaning...$(NC)"
	rm -rf bin/
	$(GO) clean -cache

deps: ## Загрузить зависимости
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy

fmt: ## Форматировать код
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...

# Разработка
dev: ## Запустить в режиме разработки
	@echo "$(GREEN)Starting development environment...$(NC)"
	$(DOCKER_COMPOSE) up postgres redis mongodb rabbitmq -d
	@echo "$(GREEN)Waiting for services to start...$(NC)"
	@sleep 5
	@echo "$(GREEN)Development environment ready!$(NC)"

dev-stop: ## Остановить режим разработки
	@echo "$(RED)Stopping development environment...$(NC)"
	$(DOCKER_COMPOSE) stop postgres redis mongodb rabbitmq

# Генерация
gen-swagger: ## Сгенерировать Swagger документацию
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	swag init -g cmd/api-gateway/main.go -o docs/swagger

# Мониторинг
monitor: ## Открыть Grafana
	@echo "$(GREEN)Opening Grafana...$(NC)"
	open http://localhost:3000

monitor-metrics: ## Показать метрики Prometheus
	@echo "$(GREEN)Prometheus metrics...$(NC)"
	curl http://localhost:9090/metrics

monitor-health: ## Проверить здоровье сервисов
	@echo "$(GREEN)Health check...$(NC)"
	curl http://localhost:8080/health

monitor-gateway: ## Запустить API Gateway
	@echo "$(GREEN)Starting API Gateway...$(NC)"
	$(GO) run ./cmd/gateway/main.go

# Быстрые команды
up: docker-up ## Быстрый запуск всех сервисов

down: docker-down ## Быстрая остановка всех сервисов

restart: down up ## Перезапуск всех сервисов

logs: docker-logs ## Показать логи

# Быстрый запуск для разработки
fast-start: ## Быстрый запуск минимальной системы
	@echo "🚀 Быстрый запуск LootBay..."
	-@pkill -f "api-gateway" 2>/dev/null
	-@pkill -f "next dev" 2>/dev/null
	-@docker compose -f docker-compose.dev.yml down 2>/dev/null
	@docker compose -f docker-compose.dev.yml up -d
	@echo "⏳ Ожидание баз данных..."
	@sleep 5
	@echo "🔨 Сборка API Gateway..."
	@go build -o bin/api-gateway ./cmd/api-gateway
	@echo "🌐 Запуск API Gateway..."
	@./bin/api-gateway &
	@echo "⚛️  Запуск фронтенда..."
	@cd web && npm run dev &
	@echo "✅ Система запущена!"
	@echo "🌐 API: http://localhost:8080/health"
	@echo "⚛️  Web: http://localhost:3000"

fast-stop: ## Остановить все процессы разработки
	@echo "⏹️  Остановка процессов..."
	-@pkill -f "api-gateway" 2>/dev/null
	-@pkill -f "next dev" 2>/dev/null
	-@docker compose -f docker-compose.dev.yml down 2>/dev/null
	@echo "✅ Остановлено"

status: ## Показать статус системы
	@echo "📊 Статус системы:"
	@echo "API Gateway:" && curl -s http://localhost:8080/health 2>/dev/null || echo "❌ Недоступен"
	@echo "Frontend:" && curl -s http://localhost:3000 2>/dev/null | head -1 || echo "❌ Недоступен"
	@echo "Процессы:" && ps aux | grep -E "(api-gateway|next dev)" | grep -v grep || echo "Нет запущенных процессов" 