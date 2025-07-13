package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/errors"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/reazonvan/LootBay/pkg/response"
	"github.com/reazonvan/LootBay/pkg/validation"
)

// ChatHandler хендлер для чата
type ChatHandler struct {
	chatService      service.ChatService
	websocketService service.WebSocketService
	validator        *validation.Validator
	logger           *logger.Logger
	upgrader         websocket.Upgrader
}

// NewChatHandler создает новый хендлер чата
func NewChatHandler(chatService service.ChatService, websocketService service.WebSocketService, logger *logger.Logger) *ChatHandler {
	return &ChatHandler{
		chatService:      chatService,
		websocketService: websocketService,
		validator:        validation.NewValidator(),
		logger:           logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// В продакшене нужно проверять origin
				return true
			},
		},
	}
}

// CreateConversationRequest запрос создания диалога
type CreateConversationRequest struct {
	User2ID uuid.UUID  `json:"user2_id" validate:"required"`
	OrderID *uuid.UUID `json:"order_id,omitempty"`
}

// SendMessageRequest запрос отправки сообщения
type SendMessageRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
	ReceiverID     uuid.UUID `json:"receiver_id" validate:"required"`
	Content        string    `json:"content" validate:"required,max=4000"`
	MessageType    string    `json:"message_type,omitempty"`
}

// GetConversations получает список диалогов пользователя
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	// Получаем параметры пагинации
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	conversations, err := h.chatService.GetUserConversations(userID.(uuid.UUID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get conversations", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, gin.H{
		"conversations": conversations,
		"limit":         limit,
		"offset":        offset,
	})
}

// CreateConversation создает новый диалог
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	conversation, err := h.chatService.CreateConversation(userID.(uuid.UUID), req.User2ID, req.OrderID)
	if err != nil {
		h.logger.Error("Failed to create conversation", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Created(c, conversation)
}

// GetConversation получает диалог по ID
func (h *ChatHandler) GetConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid conversation ID")
		return
	}

	conversation, err := h.chatService.GetConversation(conversationID, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get conversation", "error", err)
		response.Error(c, errors.ErrNotFound)
		return
	}

	response.Success(c, conversation)
}

// DeleteConversation удаляет диалог
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid conversation ID")
		return
	}

	err = h.chatService.DeleteConversation(conversationID, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to delete conversation", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.NoContent(c)
}

// GetMessages получает сообщения диалога
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid conversation ID")
		return
	}

	// Получаем параметры пагинации
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.chatService.GetMessages(conversationID, userID.(uuid.UUID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get messages", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
	})
}

// SendMessage отправляет сообщение
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	message, err := h.chatService.SendMessage(
		req.ConversationID,
		userID.(uuid.UUID),
		req.ReceiverID,
		req.Content,
		req.MessageType,
	)
	if err != nil {
		h.logger.Error("Failed to send message", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	// Отправляем сообщение через WebSocket если получатель онлайн
	wsMessage := &service.WebSocketMessage{
		Type: "new_message",
		From: userID.(uuid.UUID),
		To:   req.ReceiverID,
		Data: message,
	}

	if err := h.websocketService.SendMessageToUser(req.ReceiverID, wsMessage); err != nil {
		h.logger.Warn("Failed to send WebSocket message", "error", err)
	}

	response.Created(c, message)
}

// MarkMessageAsRead отмечает сообщение как прочитанное
func (h *ChatHandler) MarkMessageAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid message ID")
		return
	}

	err = h.chatService.MarkMessageAsRead(messageID, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to mark message as read", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.NoContent(c)
}

// HandleWebSocket обрабатывает WebSocket соединение
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Апгрейдим соединение до WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return
	}

	// Создаем клиента
	client := &service.Client{
		ID:     uuid.New(),
		UserID: userID.(uuid.UUID),
		Conn:   conn,
		Send:   make(chan *service.WebSocketMessage, 256),
	}

	// Обрабатываем клиента
	h.websocketService.HandleClient(client)
}
