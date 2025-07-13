package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/reazonvan/LootBay/internal/auth"
	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/errors"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/reazonvan/LootBay/pkg/response"
	"github.com/reazonvan/LootBay/pkg/validation"
)

// UserHandler хендлер для работы с пользователями
type UserHandler struct {
	userService service.UserService
	jwtManager  *auth.JWTManager
	validator   *validation.Validator
	logger      *logger.Logger
}

// NewUserHandler создает новый хендлер пользователей
func NewUserHandler(userService service.UserService, jwtManager *auth.JWTManager, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtManager:  jwtManager,
		validator:   validation.NewValidator(),
		logger:      logger,
	}
}

// RegisterRequest запрос регистрации
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,phone"`
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,password,min=8"`
}

// LoginRequest запрос входа
type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateProfileRequest запрос обновления профиля
type UpdateProfileRequest struct {
	FirstName        *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName         *string `json:"last_name" validate:"omitempty,min=1,max=100"`
	Avatar           *string `json:"avatar" validate:"omitempty,url"`
	TelegramUsername *string `json:"telegram_username" validate:"omitempty,min=1,max=32"`
}

// PasswordResetRequest запрос сброса пароля
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// NewPasswordRequest запрос установки нового пароля
type NewPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,password,min=8"`
}

// Register регистрирует нового пользователя
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, err := h.userService.Register(req.Email, req.Phone, req.Username, req.Password)
	if err != nil {
		h.logger.Error("Failed to register user", "error", err)
		response.Error(c, err)
		return
	}

	response.Created(c, gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"phone":    user.Phone,
	})
}

// Login аутентифицирует пользователя
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, tokens, err := h.userService.Login(req.Login, req.Password)
	if err != nil {
		h.logger.Error("Failed to login user", "error", err)
		response.Error(c, err)
		return
	}

	// Сохраняем сессию
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()
	go h.userService.CreateSession(user.ID, tokens.AccessToken, userAgent, ip)

	response.Success(c, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"avatar":     user.Avatar,
			"is_seller":  user.IsSeller,
		},
		"tokens": tokens,
	})
}

// RefreshToken обновляет токены
func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	tokens, err := h.userService.RefreshTokens(refreshToken)
	if err != nil {
		h.logger.Error("Failed to refresh tokens", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"tokens": tokens,
	})
}

// Logout выход из системы
func (h *UserHandler) Logout(c *gin.Context) {
	tokenString, exists := c.Get("token")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	claims, err := h.jwtManager.ValidateToken(tokenString.(string))
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	// Отзываем токен
	err = h.jwtManager.RevokeToken(c.Request.Context(), claims)
	if err != nil {
		h.logger.Error("Failed to revoke token", "error", err)
		// Не возвращаем ошибку клиенту, так как выход все равно должен быть успешным
	}

	// Также можно отозвать и refresh token, если он передается
	// ...

	response.NoContent(c)
}

// GetProfile получает профиль текущего пользователя
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	user, err := h.userService.GetProfile(userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get user profile", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"id":                user.ID,
		"email":             user.Email,
		"username":          user.Username,
		"first_name":        user.FirstName,
		"last_name":         user.LastName,
		"avatar":            user.Avatar,
		"is_email_verified": user.IsEmailVerified,
		"is_seller":         user.IsSeller,
		"balance":           user.Balance,
		"rating":            user.Rating,
		"total_reviews":     user.TotalReviews,
		"telegram_id":       user.TelegramID,
		"telegram_username": user.TelegramUsername,
		"created_at":        user.CreatedAt,
	})
}

// UpdateProfile обновляет профиль пользователя
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Конвертируем в map для обновления
	updates := make(map[string]interface{})
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.TelegramUsername != nil {
		updates["telegram_username"] = *req.TelegramUsername
	}

	user, err := h.userService.UpdateProfile(userID.(uuid.UUID), updates)
	if err != nil {
		h.logger.Error("Failed to update user profile", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"id":                user.ID,
		"email":             user.Email,
		"username":          user.Username,
		"first_name":        user.FirstName,
		"last_name":         user.LastName,
		"avatar":            user.Avatar,
		"telegram_username": user.TelegramUsername,
	})
}

// GetUser получает публичную информацию о пользователе
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"id":               user.ID,
		"username":         user.Username,
		"avatar":           user.Avatar,
		"is_seller":        user.IsSeller,
		"rating":           user.Rating,
		"total_reviews":    user.TotalReviews,
		"positive_reviews": user.PositiveReviews,
		"last_online":      user.LastOnline,
		"created_at":       user.CreatedAt,
	})
}

// ForgotPassword запрос на сброс пароля
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	_, err := h.userService.RequestPasswordReset(req.Email)
	if err != nil {
		h.logger.Error("Failed to request password reset", "error", err)
	}

	// Всегда возвращаем успех для безопасности
	response.Success(c, gin.H{
		"message": "Если пользователь с таким email существует, инструкции по сбросу пароля отправлены на почту",
	})
}

// ResetPassword устанавливает новый пароль
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req NewPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	err := h.userService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		h.logger.Error("Failed to reset password", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Пароль успешно изменен",
	})
}

// VerifyEmail подтверждает email
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.ValidationError(c, "Token is required")
		return
	}

	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid user ID")
		return
	}

	err = h.userService.VerifyEmail(userID, token)
	if err != nil {
		h.logger.Error("Failed to verify email", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Email успешно подтвержден",
	})
}

// PasswordPolicy возвращает политику паролей в machine-readable формате
func (h *UserHandler) PasswordPolicy(c *gin.Context) {
	policy := validation.GetPasswordPolicy()
	response.Success(c, gin.H{
		"password_policy": policy,
	})
}
