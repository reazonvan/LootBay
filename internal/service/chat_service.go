package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

// ChatService интерфейс для сервиса чата
type ChatService interface {
	// Conversation methods
	CreateConversation(user1ID, user2ID uuid.UUID, orderID *uuid.UUID) (*models.Conversation, error)
	GetConversation(id uuid.UUID, userID uuid.UUID) (*models.Conversation, error)
	GetUserConversations(userID uuid.UUID, limit, offset int) ([]models.Conversation, error)
	DeleteConversation(id uuid.UUID, userID uuid.UUID) error

	// Message methods
	SendMessage(conversationID, senderID, receiverID uuid.UUID, content, messageType string) (*models.Message, error)
	GetMessages(conversationID uuid.UUID, userID uuid.UUID, limit, offset int) ([]models.Message, error)
	MarkMessageAsRead(messageID uuid.UUID, userID uuid.UUID) error
	MarkConversationAsRead(conversationID uuid.UUID, userID uuid.UUID) error
	GetUnreadMessagesCount(userID uuid.UUID) (int64, error)
	SearchMessages(query string, conversationID uuid.UUID, userID uuid.UUID, limit, offset int) ([]models.Message, error)
}

type chatService struct {
	conversationRepo repository.ConversationRepository
	messageRepo      repository.MessageRepository
	userRepo         repository.UserRepository
	logger           *logger.Logger
}

// NewChatService создает новый сервис чата
func NewChatService(
	conversationRepo repository.ConversationRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	logger *logger.Logger,
) ChatService {
	return &chatService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		userRepo:         userRepo,
		logger:           logger,
	}
}

// CreateConversation создает новый диалог
func (s *chatService) CreateConversation(user1ID, user2ID uuid.UUID, orderID *uuid.UUID) (*models.Conversation, error) {
	// Проверяем что пользователи существуют
	user1, err := s.userRepo.GetByID(user1ID)
	if err != nil || user1 == nil {
		return nil, fmt.Errorf("user1 not found")
	}

	user2, err := s.userRepo.GetByID(user2ID)
	if err != nil || user2 == nil {
		return nil, fmt.Errorf("user2 not found")
	}

	// Проверяем что пользователи разные
	if user1ID == user2ID {
		return nil, fmt.Errorf("cannot create conversation with yourself")
	}

	// Пытаемся найти существующий диалог или создать новый
	conversation, err := s.conversationRepo.CreateOrGet(user1ID, user2ID, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	s.logger.Info("Conversation created/found", "id", conversation.ID, "user1_id", user1ID, "user2_id", user2ID)
	return conversation, nil
}

// GetConversation получает диалог по ID
func (s *chatService) GetConversation(id uuid.UUID, userID uuid.UUID) (*models.Conversation, error) {
	conversation, err := s.conversationRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Проверяем права доступа
	if conversation.User1ID != userID && conversation.User2ID != userID {
		return nil, fmt.Errorf("access denied")
	}

	return conversation, nil
}

// GetUserConversations получает диалоги пользователя
func (s *chatService) GetUserConversations(userID uuid.UUID, limit, offset int) ([]models.Conversation, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.conversationRepo.GetByUserID(userID, limit, offset)
}

// DeleteConversation удаляет диалог
func (s *chatService) DeleteConversation(id uuid.UUID, userID uuid.UUID) error {
	conversation, err := s.conversationRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("conversation not found")
	}

	// Проверяем права доступа
	if conversation.User1ID != userID && conversation.User2ID != userID {
		return fmt.Errorf("access denied")
	}

	s.logger.Info("Deleting conversation", "id", id, "user_id", userID)
	return s.conversationRepo.Delete(id)
}

// SendMessage отправляет сообщение
func (s *chatService) SendMessage(conversationID, senderID, receiverID uuid.UUID, content, messageType string) (*models.Message, error) {
	// Проверяем диалог
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Проверяем права доступа
	if conversation.User1ID != senderID && conversation.User2ID != senderID {
		return nil, fmt.Errorf("access denied")
	}

	// Проверяем что получатель участвует в диалоге
	if conversation.User1ID != receiverID && conversation.User2ID != receiverID {
		return nil, fmt.Errorf("receiver is not part of this conversation")
	}

	// Проверяем что отправитель и получатель разные
	if senderID == receiverID {
		return nil, fmt.Errorf("sender and receiver cannot be the same")
	}

	// Валидация контента
	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	if len(content) > 4000 {
		return nil, fmt.Errorf("message content too long")
	}

	// Валидация типа сообщения
	if messageType == "" {
		messageType = "text"
	}

	validTypes := []string{"text", "image", "file", "system"}
	isValidType := false
	for _, validType := range validTypes {
		if messageType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return nil, fmt.Errorf("invalid message type")
	}

	// Создаем сообщение
	message := &models.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Content:        content,
		MessageType:    messageType,
		IsRead:         false,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Обновляем последнюю активность диалога
	conversation.LastMessage = &content
	if err := s.conversationRepo.UpdateLastActivity(conversationID); err != nil {
		s.logger.Error("Failed to update conversation activity", "error", err)
	}

	s.logger.Info("Message sent", "id", message.ID, "conversation_id", conversationID, "sender_id", senderID)
	return message, nil
}

// GetMessages получает сообщения диалога
func (s *chatService) GetMessages(conversationID uuid.UUID, userID uuid.UUID, limit, offset int) ([]models.Message, error) {
	// Проверяем права доступа к диалогу
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if conversation.User1ID != userID && conversation.User2ID != userID {
		return nil, fmt.Errorf("access denied")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.messageRepo.GetByConversationID(conversationID, limit, offset)
}

// MarkMessageAsRead отмечает сообщение как прочитанное
func (s *chatService) MarkMessageAsRead(messageID uuid.UUID, userID uuid.UUID) error {
	// Получаем сообщение
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return fmt.Errorf("message not found")
	}

	// Проверяем что пользователь является получателем
	if message.ReceiverID != userID {
		return fmt.Errorf("access denied")
	}

	if message.IsRead {
		return nil // Уже прочитано
	}

	return s.messageRepo.MarkAsRead(messageID, userID)
}

// MarkConversationAsRead отмечает все сообщения диалога как прочитанные
func (s *chatService) MarkConversationAsRead(conversationID uuid.UUID, userID uuid.UUID) error {
	// Проверяем права доступа
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("conversation not found")
	}

	if conversation.User1ID != userID && conversation.User2ID != userID {
		return fmt.Errorf("access denied")
	}

	return s.messageRepo.MarkConversationAsRead(conversationID, userID)
}

// GetUnreadMessagesCount получает количество непрочитанных сообщений
func (s *chatService) GetUnreadMessagesCount(userID uuid.UUID) (int64, error) {
	return s.messageRepo.GetUnreadCount(userID)
}

// SearchMessages поиск сообщений в диалоге
func (s *chatService) SearchMessages(query string, conversationID uuid.UUID, userID uuid.UUID, limit, offset int) ([]models.Message, error) {
	// Проверяем права доступа
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	if conversation.User1ID != userID && conversation.User2ID != userID {
		return nil, fmt.Errorf("access denied")
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.messageRepo.SearchMessages(query, conversationID, limit, offset)
}
