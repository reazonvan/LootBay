package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/reazonvan/LootBay/internal/auth"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/errors"
	"github.com/reazonvan/LootBay/pkg/logger"
)

// UserService интерфейс сервиса пользователей
type UserService interface {
	Register(email, phone, username, password string) (*models.User, error)
	Login(login, password string) (*models.User, *auth.TokenPair, error)
	GetByID(id uuid.UUID) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error) // alias для GetByID
	GetProfile(id uuid.UUID) (*models.User, error)
	UpdateProfile(id uuid.UUID, updates map[string]interface{}) (*models.User, error)
	UpdateUser(user *models.User) (*models.User, error)
	UpdateBalance(userID uuid.UUID, amount float64) error
	GetUsers(offset, limit int) ([]*models.User, int64, error)
	DeactivateUser(userID uuid.UUID) error
	RefreshTokens(refreshToken string) (*auth.TokenPair, error)
	RequestPasswordReset(email string) (string, error)
	ResetPassword(token, newPassword string) error
	VerifyEmail(userID uuid.UUID, token string) error
	CreateSession(userID uuid.UUID, token, userAgent, ip string) error
	DeleteSession(token string) error
}

// userService реализация сервиса пользователей
type userService struct {
	userRepo    repository.UserRepository
	roleService RoleService
	jwtManager  *auth.JWTManager
	redis       *redis.Client
	logger      *logger.Logger
}

// NewUserService создает новый сервис пользователей
func NewUserService(
	userRepo repository.UserRepository,
	roleService RoleService,
	jwtConfig *config.JWTConfig,
	redis *redis.Client,
	logger *logger.Logger,
) UserService {
	return &userService{
		userRepo:    userRepo,
		roleService: roleService,
		jwtManager:  auth.NewJWTManager(jwtConfig, redis),
		redis:       redis,
		logger:      logger,
	}
}

// Register регистрирует нового пользователя
func (s *userService) Register(email, phone, username, password string) (*models.User, error) {
	// Проверяем существование пользователя
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, errors.Wrap(err)
	}
	if existingUser != nil {
		return nil, errors.New("USER_EXISTS", "Пользователь с таким email уже существует", 409, nil)
	}

	existingUser, err = s.userRepo.GetByUsername(username)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, errors.Wrap(err)
	}
	if existingUser != nil {
		return nil, errors.New("USERNAME_EXISTS", "Пользователь с таким именем уже существует", 409, nil)
	}

	// Проверяем телефон
	existingUser, err = s.userRepo.GetByPhone(phone)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, errors.Wrap(err)
	}
	if existingUser != nil {
		return nil, errors.New("PHONE_EXISTS", "Пользователь с таким телефоном уже существует", 409, nil)
	}

	// Хешируем пароль
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// Создаем пользователя
	user := &models.User{
		Email:        email,
		Phone:        phone,
		Username:     username,
		PasswordHash: hashedPassword,
		IsActive:     true,
		LastOnline:   time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.Wrap(err)
	}

	s.logger.Info("User registered", "user_id", user.ID, "email", email)
	return user, nil
}

// Login аутентифицирует пользователя
func (s *userService) Login(login, password string) (*models.User, *auth.TokenPair, error) {
	// Ищем пользователя по email или username
	var user *models.User
	var err error

	user, err = s.userRepo.GetByEmail(login)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, nil, errors.Wrap(err)
	}

	if user == nil {
		user, err = s.userRepo.GetByUsername(login)
		if err != nil && !errors.Is(err, errors.ErrNotFound) {
			return nil, nil, errors.Wrap(err)
		}
	}

	if user == nil {
		return nil, nil, errors.ErrInvalidCredentials
	}

	// Проверяем пароль
	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, nil, errors.ErrInvalidCredentials
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		return nil, nil, errors.New("USER_INACTIVE", "Пользователь заблокирован", 403, nil)
	}

	// Получаем роли пользователя
	roles, err := s.roleService.GetUserRoleNames(user.ID)
	if err != nil {
		s.logger.Error("Failed to get user roles", "error", err, "user_id", user.ID)
		// Если не удалось получить роли, назначаем базовую роль USER
		roles = []string{"USER"}
	}

	// Получаем разрешения пользователя
	permissions, err := s.roleService.GetUserPermissionNames(user.ID)
	if err != nil {
		s.logger.Error("Failed to get user permissions", "error", err, "user_id", user.ID)
		permissions = []string{}
	}

	// Генерируем токены
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Username, user.IsSeller, roles, permissions)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	// Обновляем время последнего визита
	s.userRepo.UpdateLastOnline(user.ID)

	s.logger.Info("User logged in", "user_id", user.ID)
	return user, tokens, nil
}

// GetByID получает пользователя по ID
func (s *userService) GetByID(id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, err
		}
		return nil, errors.Wrap(err)
	}
	return user, nil
}

// GetProfile получает профиль пользователя
func (s *userService) GetProfile(id uuid.UUID) (*models.User, error) {
	return s.GetByID(id)
}

// UpdateProfile обновляет профиль пользователя
func (s *userService) UpdateProfile(id uuid.UUID, updates map[string]interface{}) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if user == nil {
		return nil, errors.ErrNotFound
	}

	// Обновляем разрешенные поля
	allowedFields := []string{"first_name", "last_name", "avatar", "telegram_username"}
	for key, value := range updates {
		isAllowed := false
		for _, field := range allowedFields {
			if field == key {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			continue
		}

		switch key {
		case "first_name":
			if v, ok := value.(string); ok {
				user.FirstName = v
			}
		case "last_name":
			if v, ok := value.(string); ok {
				user.LastName = v
			}
		case "avatar":
			if v, ok := value.(string); ok {
				user.Avatar = v
			}
		case "telegram_username":
			if v, ok := value.(string); ok {
				user.TelegramUsername = &v
			}
		}
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.Wrap(err)
	}

	return user, nil
}

// RefreshTokens обновляет пару токенов
func (s *userService) RefreshTokens(refreshToken string) (*auth.TokenPair, error) {
	tokens, err := s.jwtManager.RefreshTokenPair(refreshToken)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return tokens, nil
}

// RequestPasswordReset создает токен для сброса пароля
func (s *userService) RequestPasswordReset(email string) (string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return "", errors.Wrap(err)
	}
	if user == nil {
		// Не раскрываем информацию о существовании пользователя
		return "", nil
	}

	token, err := s.jwtManager.GeneratePasswordResetToken(user.ID, user.Email)
	if err != nil {
		return "", errors.Wrap(err)
	}

	// Исправление: логируем токен для теста, TODO: отправить email
	s.logger.Info("Password reset token generated", "email", email, "token", token)
	// TODO: emailService.SendPasswordReset(email, token)

	return token, nil
}

// ResetPassword сбрасывает пароль
func (s *userService) ResetPassword(token, newPassword string) error {
	userID, email, err := s.jwtManager.ValidatePasswordResetToken(token)
	if err != nil {
		return errors.Wrap(err)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.Wrap(err)
	}
	if user == nil || user.Email != email {
		return errors.ErrNotFound
	}

	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return errors.Wrap(err)
	}

	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// VerifyEmail подтверждает email пользователя
func (s *userService) VerifyEmail(userID uuid.UUID, token string) error {
	// TODO: Реализовать верификацию email
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.Wrap(err)
	}
	if user == nil {
		return errors.ErrNotFound
	}

	user.IsEmailVerified = true
	if err := s.userRepo.Update(user); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// CreateSession создает сессию пользователя
func (s *userService) CreateSession(userID uuid.UUID, token, userAgent, ip string) error {
	// TODO: Сохранить сессию в Redis
	sessionKey := fmt.Sprintf("session:%s", token)
	sessionData := map[string]interface{}{
		"user_id":    userID.String(),
		"user_agent": userAgent,
		"ip":         ip,
		"created_at": time.Now().Unix(),
	}

	ctx := context.Background()
	return s.redis.HSet(ctx, sessionKey, sessionData).Err()
}

// DeleteSession удаляет сессию
func (s *userService) DeleteSession(token string) error {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", token)
	return s.redis.Del(ctx, sessionKey).Err()
}

// GetUserByID alias для GetByID
func (s *userService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.GetByID(id)
}

// UpdateUser обновляет данные пользователя
func (s *userService) UpdateUser(user *models.User) (*models.User, error) {
	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.Wrap(err)
	}
	return user, nil
}

// UpdateBalance обновляет баланс пользователя
func (s *userService) UpdateBalance(userID uuid.UUID, amount float64) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.Wrap(err)
	}

	user.Balance += amount
	if user.Balance < 0 {
		return errors.New("INSUFFICIENT_BALANCE", "Недостаточно средств", 400, nil)
	}

	return s.userRepo.Update(user)
}

// GetUsers получает список пользователей с пагинацией
func (s *userService) GetUsers(offset, limit int) ([]*models.User, int64, error) {
	users, total, err := s.userRepo.List(offset, limit)
	if err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return users, total, nil
}

// DeactivateUser деактивирует пользователя
func (s *userService) DeactivateUser(userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.Wrap(err)
	}

	user.IsActive = false
	return s.userRepo.Update(user)
}
