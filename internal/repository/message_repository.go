package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *models.Message) error
	GetByID(id uuid.UUID) (*models.Message, error)
	Update(message *models.Message) error
	Delete(id uuid.UUID) error
	GetByConversationID(conversationID uuid.UUID, limit, offset int) ([]models.Message, error)
	GetUnreadMessages(userID uuid.UUID) ([]models.Message, error)
	GetUnreadCount(userID uuid.UUID) (int64, error)
	MarkAsRead(id uuid.UUID, userID uuid.UUID) error
	MarkConversationAsRead(conversationID uuid.UUID, userID uuid.UUID) error
	GetLatestMessage(conversationID uuid.UUID) (*models.Message, error)
	SearchMessages(query string, conversationID uuid.UUID, limit, offset int) ([]models.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) GetByID(id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("Conversation").Preload("Sender").Preload("Receiver").
		First(&message, "id = ?", id).Error
	return &message, err
}

func (r *messageRepository) Update(message *models.Message) error {
	return r.db.Save(message).Error
}

func (r *messageRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Message{}, "id = ?", id).Error
}

func (r *messageRepository) GetByConversationID(conversationID uuid.UUID, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("Sender").Preload("Receiver").
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Limit(limit).Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) GetUnreadMessages(userID uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("Sender").Preload("Conversation").
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *messageRepository) MarkAsRead(id uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Message{}).
		Where("id = ? AND receiver_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (r *messageRepository) MarkConversationAsRead(conversationID uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Message{}).
		Where("conversation_id = ? AND receiver_id = ? AND is_read = ?", conversationID, userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (r *messageRepository) GetLatestMessage(conversationID uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("Sender").
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		First(&message).Error
	return &message, err
}

func (r *messageRepository) SearchMessages(query string, conversationID uuid.UUID, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("Sender").Preload("Receiver").
		Where("conversation_id = ? AND content ILIKE ?", conversationID, "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&messages).Error
	return messages, err
}
