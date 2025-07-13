# 🚀 Быстрый запуск LootBay

## Оптимизированный запуск (< 30 секунд)

### 1. Быстрый старт
```bash
make fast-start
```

### 2. Проверка статуса
```bash
make status
```

### 3. Остановка
```bash
make fast-stop
```

## Что запускается

✅ **PostgreSQL** (порт 5432) - основная база данных  
✅ **Redis** (порт 6379) - кэширование  
✅ **API Gateway** (порт 8080) - прокси к микросервисам  
✅ **Next.js Frontend** (порт 3000) - веб-интерфейс  

## Доступ

- **🌐 Веб-интерфейс**: http://localhost:3000
- **🔌 API**: http://localhost:8080
- **💚 Health check**: http://localhost:8080/health

## Особенности оптимизации

### Backend
- Минимальный docker-compose.dev.yml (только PostgreSQL + Redis)
- Отключены тяжелые сервисы (MongoDB, RabbitMQ)
- API Gateway работает как прокси

### Frontend  
- Отключены source maps в development
- Кэширование babel
- Оптимизированная webpack конфигурация
- Отключен React Strict Mode
- Статус API в реальном времени

## Альтернативные команды

```bash
# Полный запуск всех сервисов
make docker-up

# Только базы данных
docker compose -f docker-compose.dev.yml up -d

# Ручной запуск
./scripts/dev-start.sh
```

## Время запуска

- **Базы данных**: ~5 сек
- **API Gateway**: ~2 сек  
- **Frontend**: ~15 сек
- **Общее время**: ~22 сек

## Troubleshooting

```bash
# Если порты заняты
make fast-stop
pkill -f "api-gateway|next dev"

# Очистка Docker
docker compose -f docker-compose.dev.yml down -v

# Пересборка
go build -o bin/api-gateway ./cmd/api-gateway
cd web && npm run build
``` 