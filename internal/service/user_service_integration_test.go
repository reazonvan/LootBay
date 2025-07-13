package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/database"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/internal/testutils"
	"github.com/reazonvan/LootBay/pkg/errors"
	"github.com/reazonvan/LootBay/pkg/logger"
)

// UserServiceIntegrationTestSuite набор интеграционных тестов для UserService
type UserServiceIntegrationTestSuite struct {
	suite.Suite
	containers  *testutils.TestContainers
	db          *database.PostgresDB
	userRepo    repository.UserRepository
	userService UserService
	ctx         context.Context
	cancel      context.CancelFunc
}

// SetupSuite настройка тестового окружения
func (suite *UserServiceIntegrationTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 2*time.Minute)

	// Запуск тестовых контейнеров
	containers, err := testutils.StartTestContainers(suite.ctx)
	assert.NoError(suite.T(), err)
	suite.containers = containers

	// Получение информации о подключении
	connInfo, err := containers.GetContainerConnectionInfo(suite.ctx)
	assert.NoError(suite.T(), err)

	// Установка переменных окружения для тестов
	os.Setenv("DATABASE_HOST", connInfo["POSTGRES_HOST"])
	os.Setenv("DATABASE_PORT", connInfo["POSTGRES_PORT"])
	os.Setenv("DATABASE_USER", "lootbay")
	os.Setenv("DATABASE_PASSWORD", "password")
	os.Setenv("DATABASE_NAME", "lootbay_test")
	os.Setenv("DATABASE_SSL_MODE", "disable")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-for-integration-tests")

	// Инициализация базы данных
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:         connInfo["POSTGRES_HOST"],
			Port:         connInfo["POSTGRES_PORT"],
			User:         "lootbay",
			Password:     "password",
			DBName:       "lootbay_test",
			SSLMode:      "disable",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			AutoMigrate:  true,
		},
		JWT: config.JWTConfig{
			SecretKey:       "test-secret-key-for-integration-tests",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 168 * time.Hour,
		},
	}

	// Подключение к базе данных
	db, err := database.NewPostgresDB(&cfg.Database)
	assert.NoError(suite.T(), err)
	suite.db = db

	// Инициализация репозитория
	suite.userRepo = repository.NewUserRepository(db.DB)

	// Инициализация сервиса
	log := logger.NewLogger(true)
	suite.userService = NewUserService(suite.userRepo, &cfg.JWT, nil, log)
}

// TearDownSuite очистка тестового окружения
func (suite *UserServiceIntegrationTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}

	if suite.containers != nil {
		err := suite.containers.StopTestContainers(context.Background())
		assert.NoError(suite.T(), err)
	}
}

// SetupTest подготовка к каждому тесту
func (suite *UserServiceIntegrationTestSuite) SetupTest() {
	// Очистка базы данных перед каждым тестом
	err := suite.db.DB.Exec("TRUNCATE TABLE users CASCADE").Error
	assert.NoError(suite.T(), err)
}

// TestUserRegistration тестирование регистрации пользователей
func (suite *UserServiceIntegrationTestSuite) TestUserRegistration() {
	tests := []struct {
		name          string
		email         string
		phone         string
		username      string
		password      string
		expectSuccess bool
		expectError   string
	}{
		{
			name:          "successful registration",
			email:         "test@example.com",
			phone:         "+1234567890",
			username:      "testuser",
			password:      "password123",
			expectSuccess: true,
		},
		{
			name:          "duplicate email",
			email:         "test@example.com",
			phone:         "+1234567891",
			username:      "testuser2",
			password:      "password123",
			expectSuccess: false,
			expectError:   "USER_EXISTS",
		},
		{
			name:          "duplicate username",
			email:         "test2@example.com",
			phone:         "+1234567892",
			username:      "testuser",
			password:      "password123",
			expectSuccess: false,
			expectError:   "USER_EXISTS",
		},
		{
			name:          "duplicate phone",
			email:         "test3@example.com",
			phone:         "+1234567890",
			username:      "testuser3",
			password:      "password123",
			expectSuccess: false,
			expectError:   "USER_EXISTS",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			user, err := suite.userService.Register(tt.email, tt.phone, tt.username, tt.password)

			if tt.expectSuccess {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), user)
				assert.Equal(suite.T(), tt.email, user.Email)
				assert.Equal(suite.T(), tt.phone, user.Phone)
				assert.Equal(suite.T(), tt.username, user.Username)
				assert.NotEmpty(suite.T(), user.ID)
				assert.True(suite.T(), user.IsActive)
			} else {
				assert.Error(suite.T(), err)
				if tt.expectError != "" {
					// Проверяем тип ошибки
					appErr, ok := err.(*errors.AppError)
					if ok {
						assert.Equal(suite.T(), tt.expectError, appErr.Code)
					}
				}
			}
		})
	}
}

// TestUserLogin тестирование входа пользователей
func (suite *UserServiceIntegrationTestSuite) TestUserLogin() {
	// Создаем тестового пользователя
	user, err := suite.userService.Register("login@example.com", "+1234567890", "loginuser", "password123")
	assert.NoError(suite.T(), err)

	tests := []struct {
		name          string
		identifier    string
		password      string
		expectSuccess bool
	}{
		{
			name:          "login with email",
			identifier:    "login@example.com",
			password:      "password123",
			expectSuccess: true,
		},
		{
			name:          "login with username",
			identifier:    "loginuser",
			password:      "password123",
			expectSuccess: true,
		},
		{
			name:          "login with phone",
			identifier:    "+1234567890",
			password:      "password123",
			expectSuccess: true,
		},
		{
			name:          "wrong password",
			identifier:    "login@example.com",
			password:      "wrongpassword",
			expectSuccess: false,
		},
		{
			name:          "user not found",
			identifier:    "nonexistent@example.com",
			password:      "password123",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			loginUser, tokens, err := suite.userService.Login(tt.identifier, tt.password)

			if tt.expectSuccess {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), loginUser)
				assert.NotNil(suite.T(), tokens)
				assert.NotEmpty(suite.T(), tokens.AccessToken)
				assert.NotEmpty(suite.T(), tokens.RefreshToken)
				assert.Equal(suite.T(), user.ID, loginUser.ID)
			} else {
				assert.Error(suite.T(), err)
			}
		})
	}
}

// TestUserProfile тестирование работы с профилем пользователя
func (suite *UserServiceIntegrationTestSuite) TestUserProfile() {
	// Создаем тестового пользователя
	user, err := suite.userService.Register("profile@example.com", "+1234567890", "profileuser", "password123")
	assert.NoError(suite.T(), err)

	// Получение профиля
	retrievedUser, err := suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrievedUser.ID)
	assert.Equal(suite.T(), user.Email, retrievedUser.Email)

	// Обновление профиля
	updateData := &models.User{
		ID:       user.ID,
		Username: "updateduser",
		Phone:    "+9876543210",
	}

	updatedUser, err := suite.userService.UpdateUser(updateData)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "updateduser", updatedUser.Username)
	assert.Equal(suite.T(), "+9876543210", updatedUser.Phone)
}

// TestUserBalance тестирование работы с балансом пользователя
func (suite *UserServiceIntegrationTestSuite) TestUserBalance() {
	// Создаем тестового пользователя
	user, err := suite.userService.Register("balance@example.com", "+1234567890", "balanceuser", "password123")
	assert.NoError(suite.T(), err)

	// Проверяем начальный баланс
	assert.Equal(suite.T(), 0.0, user.Balance)

	// Пополнение баланса
	err = suite.userService.UpdateBalance(user.ID, 100.0)
	assert.NoError(suite.T(), err)

	// Проверяем обновленный баланс
	updatedUser, err := suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100.0, updatedUser.Balance)

	// Списание с баланса
	err = suite.userService.UpdateBalance(user.ID, -30.0)
	assert.NoError(suite.T(), err)

	// Проверяем финальный баланс
	finalUser, err := suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 70.0, finalUser.Balance)
}

// TestUserList тестирование списка пользователей
func (suite *UserServiceIntegrationTestSuite) TestUserList() {
	// Создаем несколько тестовых пользователей
	for i := 0; i < 5; i++ {
		_, err := suite.userService.Register(
			fmt.Sprintf("user%d@example.com", i),
			fmt.Sprintf("+123456789%d", i),
			fmt.Sprintf("user%d", i),
			"password123",
		)
		assert.NoError(suite.T(), err)
	}

	// Получение списка пользователей
	users, total, err := suite.userService.GetUsers(0, 10)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(5), total)
	assert.Len(suite.T(), users, 5)

	// Пагинация
	users, total, err = suite.userService.GetUsers(2, 2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(5), total)
	assert.Len(suite.T(), users, 2)
}

// TestUserDeactivation тестирование деактивации пользователей
func (suite *UserServiceIntegrationTestSuite) TestUserDeactivation() {
	// Создаем тестового пользователя
	user, err := suite.userService.Register("deactivate@example.com", "+1234567890", "deactivateuser", "password123")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), user.IsActive)

	// Деактивация пользователя
	err = suite.userService.DeactivateUser(user.ID)
	assert.NoError(suite.T(), err)

	// Проверяем статус
	deactivatedUser, err := suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), deactivatedUser.IsActive)

	// Попытка входа деактивированного пользователя
	_, _, err = suite.userService.Login("deactivate@example.com", "password123")
	assert.Error(suite.T(), err)
}

// Запуск тестов
func TestUserServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceIntegrationTestSuite))
}
