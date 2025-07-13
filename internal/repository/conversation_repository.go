package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type ConversationRepository interface {
	Create(conversation *models.Conversation) error
	GetByID(id uuid.UUID) (*models.Conversation, error)
	Update(conversation *models.Conversation) error
	Delete(id uuid.UUID) error
	GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Conversation, error)
	GetByUsers(user1ID, user2ID uuid.UUID) (*models.Conversation, error)
	GetByOrderID(orderID uuid.UUID) (*models.Conversation, error)
	UpdateLastActivity(id uuid.UUID) error
	GetActiveConversations(userID uuid.UUID, limit, offset int) ([]models.Conversation, error)
	CreateOrGet(user1ID, user2ID uuid.UUID, orderID *uuid.UUID) (*models.Conversation, error)
}

type conversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(conversation *models.Conversation) error {
	return r.db.Create(conversation).Error
}

func (r *conversationRepository) GetByID(id uuid.UUID) (*models.Conversation, error) {
	var conversation models.Conversation
	err := r.db.Preload("User1").Preload("User2").Preload("Order").
		First(&conversation, "id = ?", id).Error
	return &conversation, err
}

func (r *conversationRepository) Update(conversation *models.Conversation) error {
	return r.db.Save(conversation).Error
}

func (r *conversationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Conversation{}, "id = ?", id).Error
}

func (r *conversationRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Conversation, error) {
	var conversations []models.Conversation
	err := r.db.Preload("User1").Preload("User2").Preload("Order").
		Where("(user1_id = ? OR user2_id = ?) AND is_active = ?", userID, userID, true).
		Order("last_activity DESC").
		Limit(limit).Offset(offset).
		Find(&conversations).Error
	return conversations, err
}

func (r *conversationRepository) GetByUsers(user1ID, user2ID uuid.UUID) (*models.Conversation, error) {
	var conversation models.Conversation
	err := r.db.Preload("User1").Preload("User2").Preload("Order").
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
			user1ID, user2ID, user2ID, user1ID).
		First(&conversation).Error
	return &conversation, err
}

func (r *conversationRepository) GetByOrderID(orderID uuid.UUID) (*models.Conversation, error) {
	var conversation models.Conversation
	err := r.db.Preload("User1").Preload("User2").Preload("Order").
		Where("order_id = ?", orderID).
		First(&conversation).Error
	return &conversation, err
}

func (r *conversationRepository) UpdateLastActivity(id uuid.UUID) error {
	return r.db.Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("last_activity", time.Now()).Error
}

func (r *conversationRepository) GetActiveConversations(userID uuid.UUID, limit, offset int) ([]models.Conversation, error) {
	var conversations []models.Conversation
	err := r.db.Preload("User1").Preload("User2").Preload("Order").
		Where("(user1_id = ? OR user2_id = ?) AND is_active = ?", userID, userID, true).
		Order("last_activity DESC").
		Limit(limit).Offset(offset).
		Find(&conversations).Error
	return conversations, err
}

func (r *conversationRepository) CreateOrGet(user1ID, user2ID uuid.UUID, orderID *uuid.UUID) (*models.Conversation, error) {
	// Сначала пытаемся найти существующий диалог
	var conversation models.Conversation
	err := r.db.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
		user1ID, user2ID, user2ID, user1ID).First(&conversation).Error

	if err == nil {
		// Диалог найден, обновляем связь с заказом если нужно
		if orderID != nil && conversation.OrderID == nil {
			conversation.OrderID = orderID
			r.db.Save(&conversation)
		}
		return &conversation, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Создаем новый диалог
	conversation = models.Conversation{
		User1ID:      user1ID,
		User2ID:      user2ID,
		OrderID:      orderID,
		LastActivity: time.Now(),
		IsActive:     true,
	}

	err = r.db.Create(&conversation).Error
	return &conversation, err
}
