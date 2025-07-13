package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

// NotificationService интерфейс для сервиса уведомлений
type NotificationService interface {
	SendNotification(userID uuid.UUID, notificationType, subject, message string) error
	GetUserNotifications(userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	MarkAsRead(id uuid.UUID, userID uuid.UUID) error
	DeleteNotification(id uuid.UUID, userID uuid.UUID) error
	GetUnreadCount(userID uuid.UUID) (int64, error)
	ProcessPendingNotifications() error

	// Template notifications
	SendOrderNotification(orderID uuid.UUID, userID uuid.UUID, status string) error
	SendPaymentNotification(paymentID uuid.UUID, userID uuid.UUID, status string) error
	SendMessageNotification(messageID uuid.UUID, userID uuid.UUID) error
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
	emailService     EmailService
	telegramService  TelegramService
	pushService      PushService
	logger           logger.Logger
}

// NewNotificationService создает новый сервис уведомлений
func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	emailService EmailService,
	telegramService TelegramService,
	pushService PushService,
	logger logger.Logger,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		emailService:     emailService,
		telegramService:  telegramService,
		pushService:      pushService,
		logger:           logger,
	}
}

// SendNotification отправляет уведомление
func (s *notificationService) SendNotification(userID uuid.UUID, notificationType, subject, message string) error {
	// Проверяем существование пользователя
	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}

	// Валидация данных
	if subject == "" {
		return fmt.Errorf("subject is required")
	}
	if message == "" {
		return fmt.Errorf("message is required")
	}

	validTypes := []string{"email", "telegram", "push"}
	isValidType := false
	for _, validType := range validTypes {
		if notificationType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid notification type")
	}

	// Создаем уведомление
	notification := &models.Notification{
		UserID:  userID,
		Type:    notificationType,
		Subject: subject,
		Message: message,
		Status:  "pending",
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Пытаемся отправить сразу
	go s.sendNotificationAsync(notification, user)

	s.logger.Info("Notification created", "id", notification.ID, "user_id", userID, "type", notificationType)
	return nil
}

// sendNotificationAsync асинхронно отправляет уведомление
func (s *notificationService) sendNotificationAsync(notification *models.Notification, user *models.User) {
	var err error

	switch notification.Type {
	case "email":
		if user.Email != "" && user.IsEmailVerified {
			err = s.emailService.SendEmail(user.Email, notification.Subject, notification.Message)
		} else {
			err = fmt.Errorf("user email not verified or empty")
		}

	case "telegram":
		if user.TelegramID != nil {
			err = s.telegramService.SendMessage(*user.TelegramID, notification.Message)
		} else {
			err = fmt.Errorf("user telegram not connected")
		}

	case "push":
		err = s.pushService.SendPushNotification(user.ID, notification.Subject, notification.Message)

	default:
		err = fmt.Errorf("unknown notification type: %s", notification.Type)
	}

	if err != nil {
		s.logger.Error("Failed to send notification", "id", notification.ID, "error", err)
		s.notificationRepo.MarkAsFailed(notification.ID, err.Error())
	} else {
		s.logger.Info("Notification sent", "id", notification.ID, "type", notification.Type)
		s.notificationRepo.MarkAsSent(notification.ID)
	}
}

// GetUserNotifications получает уведомления пользователя
func (s *notificationService) GetUserNotifications(userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.notificationRepo.GetByUserID(userID, limit, offset)
}

// MarkAsRead отмечает уведомление как прочитанное
func (s *notificationService) MarkAsRead(id uuid.UUID, userID uuid.UUID) error {
	notification, err := s.notificationRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("notification not found")
	}

	if notification.UserID != userID {
		return fmt.Errorf("access denied")
	}

	// Обновляем статус на "read"
	notification.Status = "read"
	return s.notificationRepo.Update(notification)
}

// DeleteNotification удаляет уведомление
func (s *notificationService) DeleteNotification(id uuid.UUID, userID uuid.UUID) error {
	notification, err := s.notificationRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("notification not found")
	}

	if notification.UserID != userID {
		return fmt.Errorf("access denied")
	}

	return s.notificationRepo.Delete(id)
}

// GetUnreadCount получает количество непрочитанных уведомлений
func (s *notificationService) GetUnreadCount(userID uuid.UUID) (int64, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}

// ProcessPendingNotifications обрабатывает неотправленные уведомления
func (s *notificationService) ProcessPendingNotifications() error {
	notifications, err := s.notificationRepo.GetUnsentNotifications(100)
	if err != nil {
		return fmt.Errorf("failed to get unsent notifications: %w", err)
	}

	s.logger.Info("Processing pending notifications", "count", len(notifications))

	for _, notification := range notifications {
		user, err := s.userRepo.GetByID(notification.UserID)
		if err != nil {
			s.logger.Error("Failed to get user for notification", "notification_id", notification.ID, "user_id", notification.UserID)
			continue
		}

		go s.sendNotificationAsync(&notification, user)
	}

	return nil
}

// SendOrderNotification отправляет уведомление о заказе
func (s *notificationService) SendOrderNotification(orderID uuid.UUID, userID uuid.UUID, status string) error {
	var subject, message string

	switch status {
	case "created":
		subject = "Новый заказ создан"
		message = fmt.Sprintf("Ваш заказ %s был создан и ожидает оплаты.", orderID)
	case "paid":
		subject = "Заказ оплачен"
		message = fmt.Sprintf("Ваш заказ %s был успешно оплачен.", orderID)
	case "delivered":
		subject = "Заказ выполнен"
		message = fmt.Sprintf("Ваш заказ %s был выполнен продавцом.", orderID)
	case "completed":
		subject = "Заказ завершен"
		message = fmt.Sprintf("Ваш заказ %s был успешно завершен.", orderID)
	case "cancelled":
		subject = "Заказ отменен"
		message = fmt.Sprintf("Ваш заказ %s был отменен.", orderID)
	default:
		subject = "Обновление заказа"
		message = fmt.Sprintf("Статус вашего заказа %s изменился на: %s", orderID, status)
	}

	// Отправляем email уведомление
	if err := s.SendNotification(userID, "email", subject, message); err != nil {
		s.logger.Error("Failed to send order email notification", "order_id", orderID, "user_id", userID, "error", err)
	}

	// Отправляем telegram уведомление
	if err := s.SendNotification(userID, "telegram", subject, message); err != nil {
		s.logger.Error("Failed to send order telegram notification", "order_id", orderID, "user_id", userID, "error", err)
	}

	return nil
}

// SendPaymentNotification отправляет уведомление о платеже
func (s *notificationService) SendPaymentNotification(paymentID uuid.UUID, userID uuid.UUID, status string) error {
	var subject, message string

	switch status {
	case "completed":
		subject = "Платеж успешно выполнен"
		message = fmt.Sprintf("Ваш платеж %s был успешно обработан.", paymentID)
	case "failed":
		subject = "Ошибка платежа"
		message = fmt.Sprintf("Произошла ошибка при обработке платежа %s.", paymentID)
	case "refunded":
		subject = "Возврат средств"
		message = fmt.Sprintf("Средства по платежу %s были возвращены на ваш счет.", paymentID)
	default:
		subject = "Обновление платежа"
		message = fmt.Sprintf("Статус вашего платежа %s изменился на: %s", paymentID, status)
	}

	return s.SendNotification(userID, "email", subject, message)
}

// SendMessageNotification отправляет уведомление о новом сообщении
func (s *notificationService) SendMessageNotification(messageID uuid.UUID, userID uuid.UUID) error {
	subject := "Новое сообщение"
	message := "Вы получили новое сообщение в чате."

	// Отправляем push уведомление для мгновенного уведомления
	return s.SendNotification(userID, "push", subject, message)
}

// EmailService интерфейс для email сервиса
type EmailService interface {
	SendEmail(to, subject, body string) error
}

// TelegramService интерфейс для telegram сервиса
type TelegramService interface {
	SendMessage(chatID int64, message string) error
}

// PushService интерфейс для push уведомлений
type PushService interface {
	SendPushNotification(userID uuid.UUID, title, body string) error
}

// emailService реализация email сервиса
type emailService struct {
	config *config.EmailConfig
	logger *logger.Logger
}

// NewEmailService создает новый email сервис
func NewEmailService(config *config.EmailConfig, logger *logger.Logger) EmailService {
	return &emailService{
		config: config,
		logger: logger,
	}
}

func (s *emailService) SendEmail(to, subject, body string) error {
	// TODO: Реализовать отправку email через SendGrid или SMTP
	s.logger.Info("Email sent", "to", to, "subject", subject)
	return nil
}

// telegramService реализация telegram сервиса
type telegramService struct {
	config *config.TelegramConfig
	logger *logger.Logger
}

// NewTelegramService создает новый telegram сервис
func NewTelegramService(config *config.TelegramConfig, logger *logger.Logger) TelegramService {
	return &telegramService{
		config: config,
		logger: logger,
	}
}

func (s *telegramService) SendMessage(chatID int64, message string) error {
	// TODO: Реализовать отправку сообщений через Telegram Bot API
	s.logger.Info("Telegram message sent", "chat_id", chatID, "message", message)
	return nil
}

// pushService реализация push сервиса
type pushService struct {
	logger *logger.Logger
}

// NewPushService создает новый push сервис
func NewPushService(logger *logger.Logger) PushService {
	return &pushService{
		logger: logger,
	}
}

func (s *pushService) SendPushNotification(userID uuid.UUID, title, body string) error {
	// TODO: Реализовать отправку push уведомлений
	s.logger.Info("Push notification sent", "user_id", userID, "title", title)
	return nil
}
