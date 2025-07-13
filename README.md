# LootBay - Современный игровой маркетплейс

LootBay - это микросервисная платформа для торговли игровыми активами, вдохновленная функциональностью FunPay.

## 🏗️ Архитектура

Проект построен на микросервисной архитектуре с использованием Go и следующих технологий:

- **API Gateway**: Единая точка входа для всех клиентских запросов
- **Микросервисы**: Независимые сервисы для каждой доменной области
- **PostgreSQL**: Основная база данных
- **Redis**: Кэширование и сессии
- **MongoDB**: Хранение чатов (в разработке)
- **RabbitMQ**: Асинхронная обработка сообщений
- **Docker**: Контейнеризация
- **Kubernetes**: Оркестрация (в разработке)

## 📦 Микросервисы

### ✅ Реализованные

#### User Service (Порт: 8081)
- Регистрация и аутентификация пользователей
- JWT токены (Access + Refresh)
- Управление профилями
- Сброс пароля
- Email верификация

### 🚧 В разработке

#### Product Service (Порт: 8082)
- CRUD операции с товарами
- Категоризация и поиск
- Автоподнятие лотов
- Система автовыдачи

#### Order Service (Порт: 8083)
- Создание и обработка заказов
- Escrow система
- Управление состояниями заказов
- Обработка споров

#### Payment Service (Порт: 8084)
- Интеграция с Stripe и PayPal
- Управление балансом пользователей
- Обработка webhook'ов
- Возвраты и возмещения

#### Chat Service (Порт: 8085)
- WebSocket соединения
- Чат между покупателем и продавцом
- История сообщений
- Автоответчики

#### Notification Service (Порт: 8086)
- Email уведомления
- Telegram уведомления
- Push уведомления
- Шаблоны сообщений

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.21+
- Docker и Docker Compose
- PostgreSQL 15+
- Redis 7+
- Make

### Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/username/lootbay.git
cd lootbay
```

2. Настройте переменные окружения:
```bash
cp .env.example .env
# Отредактируйте .env файл под ваши настройки
```

3. Установите зависимости:
```bash
make deps
```

4. Запустите инфраструктуру:
```bash
make up
```

5. Примените миграции:
```bash
make migrate-up
```

6. Запустите API Gateway:
```bash
make monitor-gateway
```

7. Запустите frontend (в отдельном терминале):
```bash
cd web
npm install
npm run dev
```

8. Откройте приложение:
- Frontend: http://localhost:3000
- API Gateway: http://localhost:8080
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090

## 🛠️ Разработка

### Полезные команды

```bash
# Тестирование
make test              # Все тесты
make test-unit         # Unit тесты
make test-integration  # Интеграционные тесты
make test-coverage     # Покрытие кода
make test-benchmark    # Benchmark тесты

# Миграции
make migrate-up        # Применить миграции
make migrate-down      # Откатить миграции
make migrate-status    # Статус миграций
make migrate-create    # Создать миграцию

# Мониторинг
make monitor-health    # Проверить здоровье
make monitor-metrics   # Метрики Prometheus
make monitor-gateway   # Запустить API Gateway

# Разработка
make build            # Собрать все сервисы
make clean            # Очистить сборки
make lint             # Проверить код
make fmt              # Форматирование кода
make help             # Показать все команды
```

### Структура проекта

```
lootbay/
├── cmd/                    # Точки входа для сервисов
│   ├── api-gateway/
│   ├── user-service/
│   └── ...
├── internal/              # Внутренние пакеты
│   ├── api/              # HTTP хендлеры
│   ├── auth/             # Аутентификация
│   ├── config/           # Конфигурация
│   ├── database/         # Подключения к БД
│   ├── middleware/       # Middleware
│   ├── models/           # Модели данных
│   ├── repository/       # Репозитории
│   └── service/          # Бизнес-логика
├── pkg/                   # Переиспользуемые пакеты
│   ├── errors/           # Обработка ошибок
│   ├── logger/           # Логирование
│   ├── response/         # HTTP ответы
│   └── validation/       # Валидация
├── docker/                # Docker файлы
├── k8s/                   # Kubernetes манифесты
├── migrations/            # Миграции БД
├── tests/                 # Тесты
└── web/                   # Frontend (в разработке)
```

## 📊 API Документация

API документация доступна по адресу: http://localhost:8080/swagger (в разработке)

### API Gateway

API Gateway предоставляет единую точку входа для всех микросервисов с дополнительными возможностями:

- **Rate Limiting**: Ограничение количества запросов
- **Circuit Breaker**: Автоматическое отключение недоступных сервисов
- **Authentication**: Централизованная аутентификация
- **Request/Response Logging**: Логирование всех запросов
- **Health Checks**: Проверка здоровья всех сервисов
- **API Versioning**: Версионирование API

### Эндпоинты

#### Публичные эндпоинты
- `POST /api/v1/auth/register` - Регистрация
- `POST /api/v1/auth/login` - Вход
- `GET /api/v1/products` - Список товаров
- `GET /api/v1/products/:id` - Детали товара

#### Защищенные эндпоинты (требуют JWT токен)
- `GET /api/v1/users/profile` - Профиль пользователя
- `PUT /api/v1/users/profile` - Обновление профиля
- `POST /api/v1/products` - Создание товара
- `GET /api/v1/orders` - Список заказов
- `POST /api/v1/orders` - Создание заказа
- `GET /api/v1/chats` - Список чатов

#### Административные эндпоинты (требуют роль admin)
- `GET /api/v1/admin/users` - Список всех пользователей
- `GET /api/v1/admin/orders` - Список всех заказов
- `GET /api/v1/admin/payments` - Список всех платежей

### Примеры запросов

#### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "SecurePass123!"
  }'
```

#### Вход в систему
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "user@example.com",
    "password": "SecurePass123!"
  }'
```

## 🔐 Безопасность

- **JWT токены** для аутентификации (Access + Refresh)
- **Bcrypt** для хеширования паролей
- **Rate limiting** для защиты от DDoS (10 запросов/сек на сервис)
- **Circuit Breaker** для защиты от каскадных отказов
- **CORS** настройки
- **Security headers** (HSTS, CSP, X-Frame-Options)
- **Input validation** и sanitization
- **Environment variables** для всех секретов
- **Structured logging** без чувствительных данных

## 📈 Мониторинг

### Метрики и Observability

- **Prometheus**: Метрики (http://localhost:9090)
  - HTTP метрики (запросы, ответы, latency)
  - Database метрики (connections, queries)
  - Cache метрики (hits, misses, operations)
  - Business метрики (users, orders, payments)
  - System метрики (CPU, memory, disk)

- **Grafana**: Дашборды (http://localhost:3000)
  - API Gateway дашборд
  - Database performance дашборд
  - Cache performance дашборд
  - Business metrics дашборд

- **Structured Logging**: JSON логи с correlation IDs
- **Health Checks**: Проверка здоровья всех сервисов
- **Distributed Tracing**: Jaeger (в разработке)

### Кэширование

- **Redis Cache**: Кэширование часто запрашиваемых данных
- **Cache Invalidation**: Автоматическая инвалидация связанных ключей
- **Distributed Locking**: Блокировки для критических операций
- **TTL Management**: Управление временем жизни кэша

## 🗺️ Roadmap

### Фаза 1: Core Services ✅ (частично)
- [x] User Service - аутентификация и профили
- [ ] Product Service - управление товарами
- [ ] Order Service - Escrow система

### Фаза 2: Advanced Features 🚧
- [ ] Payment Service - интеграция платежей
- [ ] Chat Service - WebSocket чат
- [ ] Notification Service - уведомления

### Фаза 3: Frontend & Testing 📅
- [ ] Web Application - React/Next.js
- [ ] Testing Suite - полное покрытие тестами
- [ ] Deployment - production готовность

## 🤝 Вклад в проект

Мы приветствуем вклад в развитие проекта! Пожалуйста, ознакомьтесь с [CONTRIBUTING.md](CONTRIBUTING.md) перед отправкой PR.

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](LICENSE) для деталей. 