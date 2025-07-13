# Архитектура LootBay

## Обзор

LootBay - это современный игровой маркетплейс, построенный на микросервисной архитектуре с использованием Go и PostgreSQL. Проект вдохновлен функциональностью FunPay и предоставляет платформу для торговли игровыми активами.

## Архитектурные принципы

### 1. Микросервисная архитектура
- Каждый сервис отвечает за свою доменную область
- Независимое развертывание и масштабирование
- Отказоустойчивость через изоляцию сервисов

### 2. Event-Driven Architecture
- Асинхронная обработка через RabbitMQ
- Event sourcing для критических операций
- Eventual consistency между сервисами

### 3. CQRS (Command Query Responsibility Segregation)
- Разделение команд и запросов
- Оптимизированные модели для чтения и записи
- Кэширование через Redis

## Компоненты системы

### API Gateway
**Порт:** 8080  
**Назначение:** Единая точка входа для всех клиентских запросов

**Функции:**
- Маршрутизация запросов к микросервисам
- Аутентификация и авторизация
- Rate limiting
- Логирование и мониторинг
- CORS handling

### User Service
**Порт:** 8081  
**Назначение:** Управление пользователями и аутентификация

**Функции:**
- Регистрация и аутентификация пользователей
- Управление профилями
- JWT токены
- Двухфакторная аутентификация
- Telegram интеграция
- **Постоянность аккаунтов** (аккаунты не могут быть удалены)

### Product Service
**Порт:** 8082  
**Назначение:** Управление товарами и услугами

**Функции:**
- CRUD операции с товарами
- Категоризация и поиск
- Автоподнятие лотов
- Система автовыдачи
- Управление изображениями

### Order Service
**Порт:** 8083  
**Назначение:** Обработка заказов и транзакций

**Функции:**
- Создание и обработка заказов
- Escrow система
- Управление состояниями заказов
- Автоматическое завершение
- Обработка споров

### Payment Service
**Порт:** 8084  
**Назначение:** Обработка платежей

**Функции:**
- Интеграция с платежными системами (Stripe, PayPal)
- Управление балансом пользователей
- Обработка webhook'ов
- Возвраты и возмещения
- Комиссии и сборы

### Chat Service
**Порт:** 8085  
**Назначение:** Система сообщений

**Функции:**
- WebSocket соединения
- Чат между покупателем и продавцом
- История сообщений
- Автоответчики
- Уведомления о новых сообщениях

### Notification Service
**Порт:** 8086  
**Назначение:** Отправка уведомлений

**Функции:**
- Email уведомления
- Telegram уведомления
- Push уведомления
- Шаблоны сообщений
- Очередь отправки

## Хранилища данных

### PostgreSQL (Основная БД)
**Назначение:** Хранение основных данных приложения

**Таблицы:**
- `users` - Пользователи
- `games` - Игры
- `categories` - Категории товаров
- `products` - Товары и услуги
- `orders` - Заказы
- `payments` - Платежи
- `reviews` - Отзывы
- `transactions` - Транзакции баланса
- `notifications` - Уведомления
- `user_sessions` - Сессии пользователей
- `auto_responses` - Автоответы

### Redis
**Назначение:** Кэширование и сессии

**Использование:**
- Кэш для часто запрашиваемых данных
- Сессии пользователей
- Rate limiting counters
- Temporary data storage

### MongoDB
**Назначение:** Хранение сообщений чата

**Коллекции:**
- `messages` - Сообщения
- `conversations` - Диалоги
- `message_read_status` - Статус прочтения

### RabbitMQ
**Назначение:** Message broker для асинхронной обработки

**Очереди:**
- `notifications` - Уведомления
- `order_events` - События заказов
- `payment_events` - События платежей
- `user_events` - События пользователей

## Безопасность

### Аутентификация
- JWT токены (Access + Refresh)
- Telegram OAuth
- Двухфакторная аутентификация

### Авторизация
- Role-based access control (RBAC)
- Middleware для проверки прав
- API key authentication для внешних интеграций

### Защита данных
- Шифрование паролей (bcrypt)
- HTTPS only в продакшене
- Input validation и sanitization
- SQL injection protection
- XSS protection

### Rate Limiting
- Per-user rate limits
- Per-IP rate limits
- Different limits for different endpoints

## Мониторинг и логирование

### Метрики (Prometheus)
- HTTP request metrics
- Database connection metrics
- Queue metrics
- Business metrics (orders, payments, etc.)

### Логирование
- Structured logging (JSON)
- Different log levels
- Request tracing
- Error tracking

### Трейсинг (Jaeger)
- Distributed tracing
- Request flow visualization
- Performance bottleneck identification

### Дашборды (Grafana)
- System metrics
- Business metrics
- Alerting rules
- Custom dashboards

## Развертывание

### Docker
- Multi-stage builds
- Optimized images
- Health checks
- Environment-specific configs

### Kubernetes
- Deployment manifests
- Services and ingress
- ConfigMaps and Secrets
- Horizontal Pod Autoscaling
- Rolling updates

### CI/CD
- GitHub Actions workflow
- Automated testing
- Security scanning
- Automated deployment

## Масштабирование

### Горизонтальное масштабирование
- Stateless microservices
- Load balancing
- Database read replicas
- Caching strategies

### Вертикальное масштабирование
- Resource optimization
- Connection pooling
- Query optimization
- Index tuning

### Кэширование
- Application-level caching (Redis)
- Database query caching
- CDN for static assets
- Browser caching headers

## Отказоустойчивость

### Circuit Breaker Pattern
- Protection against cascading failures
- Graceful degradation
- Automatic recovery

### Retry Logic
- Exponential backoff
- Dead letter queues
- Idempotent operations

### Health Checks
- Liveness probes
- Readiness probes
- Dependency health monitoring

### Backup и Recovery
- Automated database backups
- Point-in-time recovery
- Disaster recovery procedures

## API Design

### RESTful API
- Resource-based URLs
- HTTP methods semantics
- Status codes
- Content negotiation

### GraphQL (Future)
- Flexible data fetching
- Type safety
- Real-time subscriptions

### Versioning
- URL versioning (/api/v1/)
- Backward compatibility
- Deprecation strategy

### Documentation
- OpenAPI/Swagger specs
- Interactive documentation
- Code examples
- SDKs for popular languages

## Особенности реализации

### Escrow система
1. Покупатель оплачивает заказ
2. Деньги блокируются в системе
3. Продавец выполняет обязательства
4. Покупатель подтверждает получение
5. Деньги переводятся продавцу

### Автоподнятие лотов
- Configurable intervals
- Smart scheduling
- Rate limiting
- Analytics tracking

### Автовыдача товаров
- JSON-based configuration
- Template system
- Delivery confirmation
- Error handling

### Telegram интеграция
- Bot для уведомлений
- Управление через команды
- WebApp для полного функционала
- OAuth авторизация

## Производительность

### Цели производительности
- Response time < 200ms для 95% запросов
- Throughput > 1000 RPS
- Uptime > 99.9%
- Database queries < 50ms

### Оптимизации
- Database indexing
- Query optimization
- Connection pooling
- Caching strategies
- Lazy loading
- Pagination

### Load Testing
- Stress testing с Apache Bench
- Performance monitoring
- Capacity planning
- Bottleneck identification

## Безопасность данных

### Шифрование
- Data at rest encryption
- Data in transit encryption
- Key rotation policies

### Аудит
- Access logging
- Change tracking
- Security event monitoring
- Compliance reporting

### Privacy
- GDPR compliance
- Data minimization
- Account permanence (accounts cannot be deleted)
- Privacy by design

## Roadmap

### Phase 1 (MVP)
- [x] Basic user management
- [x] Product catalog
- [x] Order processing
- [x] Payment integration
- [x] Basic notifications

### Phase 2 (Enhanced Features)
- [ ] Advanced search
- [ ] Recommendation system
- [ ] Mobile app
- [ ] Advanced analytics
- [ ] Multi-language support

### Phase 3 (Enterprise Features)
- [ ] White-label solutions
- [ ] Advanced fraud detection
- [ ] Machine learning integration
- [ ] Blockchain integration
- [ ] Global payment methods 