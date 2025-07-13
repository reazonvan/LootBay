# Настройка GitHub репозитория для LootBay

## Шаги для пушу на GitHub

### 1. Создание репозитория на GitHub

1. Перейдите на [GitHub](https://github.com)
2. Нажмите "New repository"
3. Название: `LootBay`
4. Описание: `Игровой маркетплейс на микросервисной архитектуре`
5. Выберите "Public" или "Private"
6. **НЕ** инициализируйте с README (у нас уже есть)
7. Нажмите "Create repository"

### 2. Подключение локального репозитория

```bash
# Добавить удаленный репозиторий
git remote add origin https://github.com/YOUR_USERNAME/LootBay.git

# Переименовать ветку в main (если нужно)
git branch -M main

# Первый пуш
git push -u origin main
```

### 3. Настройка GitHub Secrets (для CI/CD)

В настройках репозитория (Settings → Secrets and variables → Actions):

- `DOCKER_USERNAME` - ваш Docker Hub логин
- `DOCKER_PASSWORD` - ваш Docker Hub токен

### 4. Настройка GitHub Pages (опционально)

1. Settings → Pages
2. Source: "Deploy from a branch"
3. Branch: `main`
4. Folder: `/docs`

### 5. Настройка веток

```bash
# Создать develop ветку
git checkout -b develop
git push -u origin develop

# Защитить main ветку
# Settings → Branches → Add rule
# Branch name pattern: main
# Require pull request reviews before merging: ✓
# Require status checks to pass before merging: ✓
```

### 6. Настройка Issues и Projects

1. Включить Issues в настройках репозитория
2. Создать Project board для управления задачами
3. Настроить Labels для Issues

### 7. Настройка Wiki (опционально)

1. Включить Wiki в настройках
2. Добавить документацию по API
3. Добавить руководства по развертыванию

## Готово! 🚀

Ваш проект LootBay теперь готов для разработки на GitHub с полным набором инструментов:

- ✅ CI/CD pipeline
- ✅ Автоматические тесты
- ✅ Сборка Docker образов
- ✅ Обновление зависимостей
- ✅ Шаблоны Issues и PR
- ✅ Документация
- ✅ Система ролей
- ✅ Безопасность

## Следующие шаги

1. Создайте первый Issue для планирования разработки
2. Настройте Project board
3. Пригласите контрибьюторов
4. Начните разработку новых функций! 