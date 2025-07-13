-- Создание системы ролей
-- +migrate Up

-- Таблица ролей
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    level INTEGER NOT NULL DEFAULT 0, -- уровень роли (0 - USER, 50 - ADMIN, 100 - OWNER)
    is_system BOOLEAN DEFAULT FALSE, -- системная роль (нельзя удалить)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица разрешений
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL, -- users, products, orders, etc.
    action VARCHAR(50) NOT NULL, -- create, read, update, delete, manage
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Связь ролей и разрешений (многие ко многим)
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Связь пользователей и ролей (многие ко многим)
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Индексы для производительности
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_level ON roles(level);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Добавляем триггеры для обновления updated_at
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON permissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Создаем базовые роли
INSERT INTO roles (name, description, level, is_system) VALUES
('USER', 'Обычный пользователь', 0, TRUE),
('ADMIN_SUPPORT', 'Администратор техподдержки', 50, TRUE),
('ADMIN_MODERATION', 'Администратор модерации', 50, TRUE),
('ADMIN_GAMES', 'Администратор игр и категорий', 50, TRUE),
('OWNER', 'Владелец сайта', 100, TRUE);

-- Создаем базовые разрешения
INSERT INTO permissions (name, description, resource, action, is_system) VALUES
-- Пользователи
('users.read', 'Просмотр пользователей', 'users', 'read', TRUE),
('users.update', 'Обновление пользователей', 'users', 'update', TRUE),
('users.delete', 'Удаление пользователей', 'users', 'delete', TRUE),
('users.manage', 'Управление пользователями', 'users', 'manage', TRUE),

-- Роли (только для OWNER)
('roles.read', 'Просмотр ролей', 'roles', 'read', TRUE),
('roles.create', 'Создание ролей', 'roles', 'create', TRUE),
('roles.update', 'Обновление ролей', 'roles', 'update', TRUE),
('roles.delete', 'Удаление ролей', 'roles', 'delete', TRUE),
('roles.assign', 'Назначение ролей пользователям', 'roles', 'assign', TRUE),

-- Продукты
('products.read', 'Просмотр продуктов', 'products', 'read', TRUE),
('products.create', 'Создание продуктов', 'products', 'create', TRUE),
('products.update', 'Обновление продуктов', 'products', 'update', TRUE),
('products.delete', 'Удаление продуктов', 'products', 'delete', TRUE),
('products.moderate', 'Модерация продуктов', 'products', 'moderate', TRUE),

-- Игры
('games.read', 'Просмотр игр', 'games', 'read', TRUE),
('games.create', 'Создание игр', 'games', 'create', TRUE),
('games.update', 'Обновление игр', 'games', 'update', TRUE),
('games.delete', 'Удаление игр', 'games', 'delete', TRUE),

-- Категории
('categories.read', 'Просмотр категорий', 'categories', 'read', TRUE),
('categories.create', 'Создание категорий', 'categories', 'create', TRUE),
('categories.update', 'Обновление категорий', 'categories', 'update', TRUE),
('categories.delete', 'Удаление категорий', 'categories', 'delete', TRUE),

-- Заказы
('orders.read', 'Просмотр заказов', 'orders', 'read', TRUE),
('orders.update', 'Обновление заказов', 'orders', 'update', TRUE),
('orders.cancel', 'Отмена заказов', 'orders', 'cancel', TRUE),

-- Модерация
('moderation.reports', 'Обработка жалоб', 'moderation', 'reports', TRUE),
('moderation.bans', 'Блокировка пользователей', 'moderation', 'bans', TRUE),

-- Техподдержка
('support.tickets', 'Обработка тикетов поддержки', 'support', 'tickets', TRUE),
('support.chat', 'Чат техподдержки', 'support', 'chat', TRUE);

-- Назначаем разрешения для роли USER (базовые права)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'USER' AND p.name IN ('products.read', 'games.read', 'categories.read');

-- Назначаем разрешения для роли ADMIN_SUPPORT
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'ADMIN_SUPPORT' AND p.name IN (
    'users.read', 'users.update',
    'products.read', 'games.read', 'categories.read',
    'orders.read', 'orders.update',
    'support.tickets', 'support.chat'
);

-- Назначаем разрешения для роли ADMIN_MODERATION
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'ADMIN_MODERATION' AND p.name IN (
    'users.read', 'users.update',
    'products.read', 'products.moderate', 'products.delete',
    'games.read', 'categories.read',
    'orders.read', 'orders.cancel',
    'moderation.reports', 'moderation.bans'
);

-- Назначаем разрешения для роли ADMIN_GAMES
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'ADMIN_GAMES' AND p.name IN (
    'products.read', 'products.create', 'products.update', 'products.delete',
    'games.read', 'games.create', 'games.update', 'games.delete',
    'categories.read', 'categories.create', 'categories.update', 'categories.delete'
);

-- Назначаем ВСЕ разрешения для роли OWNER
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'OWNER';

-- Назначаем роль USER всем существующим пользователям
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE r.name = 'USER';

-- +migrate Down

-- Удаляем таблицы в обратном порядке
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles; 