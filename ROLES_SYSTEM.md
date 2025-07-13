# Система ролей LootBay

## Обзор

Система ролей LootBay предоставляет гибкий механизм управления правами доступа с использованием ролей и разрешений.

## Архитектура

### Основные компоненты:
- **Роли** (`roles`) - группы пользователей с определенными правами
- **Разрешения** (`permissions`) - конкретные права доступа к ресурсам
- **Связи** (`role_permissions`, `user_roles`) - многие-ко-многим связи

### Базовые роли:
- **OWNER** (уровень 100) - владелец сайта, все права
- **ADMIN_GAMES** (уровень 50) - управление играми и категориями
- **ADMIN_MODERATION** (уровень 50) - модерация контента
- **ADMIN_SUPPORT** (уровень 50) - техподдержка
- **USER** (уровень 0) - обычный пользователь

## Установка и настройка

### 1. Выполнить миграции
```bash
make migrate
```

### 2. Инициализировать базовые роли
```bash
go run cmd/init-roles/main.go
```

### 3. Назначить владельца
Скрипт запросит email существующего пользователя для назначения роли OWNER.

## API Endpoints

### Роли
- `GET /api/v1/admin/roles` - получить все роли
- `POST /api/v1/admin/roles` - создать роль (только OWNER)
- `GET /api/v1/admin/roles/{id}` - получить роль по ID
- `PUT /api/v1/admin/roles/{id}` - обновить роль (только OWNER)
- `DELETE /api/v1/admin/roles/{id}` - удалить роль (только OWNER)

### Разрешения
- `GET /api/v1/admin/permissions` - получить все разрешения
- `POST /api/v1/admin/permissions` - создать разрешение (только OWNER)
- `GET /api/v1/admin/permissions/{id}` - получить разрешение по ID
- `PUT /api/v1/admin/permissions/{id}` - обновить разрешение (только OWNER)
- `DELETE /api/v1/admin/permissions/{id}` - удалить разрешение (только OWNER)

### Управление пользователями
- `POST /api/v1/admin/users/roles` - назначить роль пользователю (только OWNER)
- `DELETE /api/v1/admin/users/{user_id}/roles/{role_id}` - удалить роль (только OWNER)
- `GET /api/v1/admin/users/{user_id}/roles` - получить роли пользователя
- `GET /api/v1/admin/users/{user_id}/permissions` - получить разрешения пользователя

## Middleware

### Базовые middleware:
- `Auth()` - проверка JWT токена
- `OwnerOnly()` - только для владельца
- `AdminOnly()` - для любых админов
- `GamesAdminOnly()` - для админов игр
- `ModerationAdminOnly()` - для админов модерации
- `SupportAdminOnly()` - для админов техподдержки

### Гибкие middleware:
- `RequireRole("ROLE_NAME")` - требует конкретную роль
- `RequireAnyRole("ROLE1", "ROLE2")` - требует любую из ролей
- `RequirePermission("permission.name")` - требует разрешение
- `RequireAnyPermission("perm1", "perm2")` - требует любое разрешение

## Примеры использования

### Проверка роли в коде
```go
// Проверка роли
isOwner, err := roleService.IsOwner(userID)
if err != nil {
    return err
}
if !isOwner {
    return errors.ErrForbidden
}

// Проверка разрешения
hasPermission, err := roleService.HasPermission(userID, "games.create")
if err != nil {
    return err
}
```

### Использование middleware в маршрутах
```go
// Только для владельца
router.POST("/admin/roles", middleware.OwnerOnly(), roleHandler.CreateRole)

// Для админов игр
router.POST("/admin/games", middleware.GamesAdminOnly(), gameHandler.CreateGame)

// Гибкая проверка
router.GET("/admin/reports", middleware.RequireAnyRole("ADMIN_MODERATION", "OWNER"), reportHandler.GetReports)
```

### JWT токены
JWT токены автоматически включают:
- `roles` - массив ролей пользователя
- `permissions` - массив разрешений пользователя

## Базовые разрешения

### Пользователи
- `users.read` - просмотр пользователей
- `users.update` - обновление пользователей
- `users.delete` - удаление пользователей
- `users.manage` - управление пользователями

### Роли (только OWNER)
- `roles.read` - просмотр ролей
- `roles.create` - создание ролей
- `roles.update` - обновление ролей
- `roles.delete` - удаление ролей
- `roles.assign` - назначение ролей

### Игры
- `games.read` - просмотр игр
- `games.create` - создание игр
- `games.update` - обновление игр
- `games.delete` - удаление игр

### Категории
- `categories.read` - просмотр категорий
- `categories.create` - создание категорий
- `categories.update` - обновление категорий
- `categories.delete` - удаление категорий

## Безопасность

### Ограничения:
- Системные роли нельзя изменять или удалять
- Роль USER нельзя удалить у пользователя
- Только OWNER может управлять ролями
- Аккаунты пользователей постоянны (нельзя удалить)

### Лучшие практики:
1. Используйте принцип минимальных привилегий
2. Регулярно проверяйте назначенные роли
3. Логируйте все изменения ролей
4. Используйте разрешения для тонкой настройки доступа

## Структура базы данных

```sql
-- Роли
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    level INTEGER NOT NULL DEFAULT 0,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Разрешения
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Связи
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id),
    role_id UUID REFERENCES roles(id),
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);
```

## Часто задаваемые вопросы

### Как добавить новую роль?
Только владелец может создавать новые роли через API или админ-панель.

### Как назначить роль пользователю?
Используйте API endpoint `POST /api/v1/admin/users/roles` с правами владельца.

### Можно ли изменить системные роли?
Нет, системные роли защищены от изменений и удаления.

### Как работает иерархия ролей?
Роли имеют уровни (level), где OWNER=100, ADMIN=50, USER=0. Высший уровень имеет больше прав.

## Миграция с старой системы

Если у вас была старая система с полем `is_seller`:
1. Выполните миграцию `003_roles_system.sql`
2. Все пользователи автоматически получат роль USER
3. Поле `is_seller` сохраняется для обратной совместимости
4. Назначьте роли админам через скрипт инициализации 