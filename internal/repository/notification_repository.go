package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByID(id uuid.UUID) (*models.Notification, error)
	Update(notification *models.Notification) error
	Delete(id uuid.UUID) error
	GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	GetUnsentNotifications(limit int) ([]models.Notification, error)
	GetByStatus(status string, limit, offset int) ([]models.Notification, error)
	MarkAsSent(id uuid.UUID) error
	MarkAsFailed(id uuid.UUID, errorMessage string) error
	GetUnreadCount(userID uuid.UUID) (int64, error)
	DeleteOldNotifications(days int) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByID(id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.Preload("User").First(&notification, "id = ?", id).Error
	return &notification, err
}

func (r *notificationRepository) Update(notification *models.Notification) error {
	return r.db.Save(notification).Error
}

func (r *notificationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Notification{}, "id = ?", id).Error
}

func (r *notificationRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetUnsentNotifications(limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Preload("User").
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetByStatus(status string, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Preload("User").
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkAsSent(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  "sent",
			"sent_at": &now,
		}).Error
}

func (r *notificationRepository) MarkAsFailed(id uuid.UUID, errorMessage string) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  "failed",
			"message": errorMessage,
		}).Error
}

func (r *notificationRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, "sent").
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) DeleteOldNotifications(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return r.db.Where("created_at < ? AND status = ?", cutoffDate, "sent").
		Delete(&models.Notification{}).Error
}
