package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/errors"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/reazonvan/LootBay/pkg/response"
	"github.com/reazonvan/LootBay/pkg/validation"
)

// NotificationHandler хендлер для уведомлений
type NotificationHandler struct {
	notificationService service.NotificationService
	validator           *validation.Validator
	logger              *logger.Logger
}

// NewNotificationHandler создает новый хендлер уведомлений
func NewNotificationHandler(notificationService service.NotificationService, logger *logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		validator:           validation.NewValidator(),
		logger:              logger,
	}
}

// SendNotificationRequest запрос отправки уведомления
type SendNotificationRequest struct {
	UserID           uuid.UUID `json:"user_id" validate:"required"`
	NotificationType string    `json:"type" validate:"required,oneof=email telegram push"`
	Subject          string    `json:"subject" validate:"required,max=200"`
	Message          string    `json:"message" validate:"required,max=1000"`
}

// NotificationSettingsRequest запрос настроек уведомлений
type NotificationSettingsRequest struct {
	EmailEnabled    bool `json:"email_enabled"`
	TelegramEnabled bool `json:"telegram_enabled"`
	PushEnabled     bool `json:"push_enabled"`
}

// GetNotifications получает уведомления пользователя
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	// Получаем параметры пагинации
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	notifications, err := h.notificationService.GetUserNotifications(userID.(uuid.UUID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get notifications", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	// Получаем количество непрочитанных
	unreadCount, err := h.notificationService.GetUnreadCount(userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get unread count", "error", err)
		unreadCount = 0
	}

	response.Success(c, gin.H{
		"notifications": notifications,
		"unread_count":  unreadCount,
		"limit":         limit,
		"offset":        offset,
	})
}

// MarkAsRead отмечает уведомление как прочитанное
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid notification ID")
		return
	}

	err = h.notificationService.MarkAsRead(notificationID, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to mark notification as read", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.NoContent(c)
}

// DeleteNotification удаляет уведомление
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid notification ID")
		return
	}

	err = h.notificationService.DeleteNotification(notificationID, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to delete notification", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.NoContent(c)
}

// SendNotification отправляет уведомление (только для админов)
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	// TODO: Добавить проверку прав администратора

	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	err := h.notificationService.SendNotification(req.UserID, req.NotificationType, req.Subject, req.Message)
	if err != nil {
		h.logger.Error("Failed to send notification", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, gin.H{
		"message": "Notification sent successfully",
	})
}

// GetSettings получает настройки уведомлений пользователя
func (h *NotificationHandler) GetSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	// TODO: Реализовать получение настроек из базы данных
	// Пока возвращаем значения по умолчанию
	settings := gin.H{
		"email_enabled":    true,
		"telegram_enabled": true,
		"push_enabled":     true,
	}

	h.logger.Info("Getting notification settings", "user_id", userID)
	response.Success(c, settings)
}

// UpdateSettings обновляет настройки уведомлений пользователя
func (h *NotificationHandler) UpdateSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req NotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// TODO: Реализовать сохранение настроек в базе данных
	h.logger.Info("Updating notification settings", "user_id", userID, "settings", req)

	response.Success(c, gin.H{
		"message": "Settings updated successfully",
		"settings": gin.H{
			"email_enabled":    req.EmailEnabled,
			"telegram_enabled": req.TelegramEnabled,
			"push_enabled":     req.PushEnabled,
		},
	})
}
