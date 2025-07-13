package service

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/auth"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/pkg/errors"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if u := args.Get(0); u != nil {
		return u.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if u := args.Get(0); u != nil {
		return u.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) GetByPhone(phone string) (*models.User, error) {
	args := m.Called(phone)
	if u := args.Get(0); u != nil {
		return u.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	user.ID = uuid.New()
	return args.Error(0)
}
func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if u := args.Get(0); u != nil {
		return u.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) Update(user *models.User) error {
	return m.Called(user).Error(0)
}
func (m *MockUserRepository) UpdateLastOnline(id uuid.UUID) error {
	return m.Called(id).Error(0)
}
func (m *MockUserRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	args := m.Called(telegramID)
	if u := args.Get(0); u != nil {
		return u.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) List(offset, limit int) ([]*models.User, int64, error) {
	args := m.Called(offset, limit)
	if u := args.Get(0); u != nil {
		return u.([]*models.User), int64(args.Int(1)), args.Error(2)
	}
	return nil, int64(args.Int(1)), args.Error(2)
}
func (m *MockUserRepository) UpdateBalance(id uuid.UUID, amount float64) error {
	return m.Called(id, amount).Error(0)
}

func TestMain(m *testing.M) {
	originalHashPassword := auth.HashPassword
	auth.HashPassword = func(password string) (string, error) {
		return "hashed_" + password, nil
	}
	originalCheckPasswordHash := auth.CheckPasswordHash
	auth.CheckPasswordHash = func(password, hash string) bool {
		return "hashed_"+password == hash
	}

	exitCode := m.Run()

	auth.HashPassword = originalHashPassword
	auth.CheckPasswordHash = originalCheckPasswordHash
	os.Exit(exitCode)
}

func TestUserService_Register(t *testing.T) {
	jwtConfig := &config.JWTConfig{SecretKey: "test", AccessTokenTTL: time.Minute}
	log := logger.NewLogger(true)

	t.Run("successful registration", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userService := NewUserService(mockRepo, jwtConfig, nil, log)

		mockRepo.On("GetByEmail", mock.Anything).Return(nil, nil).Once()
		mockRepo.On("GetByUsername", mock.Anything).Return(nil, nil).Once()
		mockRepo.On("GetByPhone", mock.Anything).Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything).Return(nil).Once()

		_, err := userService.Register("test@test.com", "123", "user", "pass")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("email exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userService := NewUserService(mockRepo, jwtConfig, nil, log)
		mockRepo.On("GetByEmail", mock.Anything).Return(&models.User{}, nil).Once()

		_, err := userService.Register("test@test.com", "123", "user", "pass")
		assert.Error(t, err)
		appErr, ok := err.(*errors.AppError)
		assert.True(t, ok && appErr.Code == "USER_EXISTS")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Login(t *testing.T) {
	password := "password"
	hashedPassword, _ := auth.HashPassword(password)
	jwtConfig := &config.JWTConfig{SecretKey: "test", AccessTokenTTL: time.Minute, RefreshTokenTTL: time.Hour}

	t.Run("successful login with email", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		log := logger.NewLogger(true)
		userService := NewUserService(mockRepo, jwtConfig, nil, log)

		email := "user@example.com"
		user := &models.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: hashedPassword,
			IsActive:     true,
		}

		mockRepo.On("GetByEmail", email).Return(user, nil).Once()
		mockRepo.On("UpdateLastOnline", user.ID).Return(nil).Once()

		_, _, err := userService.Login(email, password)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
