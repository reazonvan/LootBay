package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/reazonvan/LootBay/pkg/logger"
)

// WebSocketMessage структура сообщения WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	From    uuid.UUID   `json:"from,omitempty"`
	To      uuid.UUID   `json:"to,omitempty"`
	Channel string      `json:"channel,omitempty"`
}

// ChatMessage структура сообщения чата
type ChatMessage struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	ReceiverID     uuid.UUID `json:"receiver_id"`
	Content        string    `json:"content"`
	MessageType    string    `json:"message_type"`
	CreatedAt      time.Time `json:"created_at"`
}

// Client представляет WebSocket клиента
type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Conn   *websocket.Conn
	Send   chan *WebSocketMessage
}

// Hub управляет WebSocket клиентами
type Hub struct {
	clients    map[uuid.UUID]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *WebSocketMessage
	mutex      sync.RWMutex
}

// WebSocketService интерфейс для WebSocket сервиса
type WebSocketService interface {
	CreateHub() *Hub
	HandleClient(client *Client)
	SendMessageToUser(userID uuid.UUID, message *WebSocketMessage) error
	BroadcastToConversation(conversationID uuid.UUID, message *WebSocketMessage) error
	GetOnlineUsers() []uuid.UUID
	IsUserOnline(userID uuid.UUID) bool
}

type webSocketService struct {
	hub    *Hub
	redis  *redis.Client
	logger *logger.Logger
}

// NewWebSocketService создает новый WebSocket сервис
func NewWebSocketService(redis *redis.Client, logger *logger.Logger) WebSocketService { // logger теперь *logger.Logger
	service := &webSocketService{
		redis:  redis,
		logger: logger,
	}
	service.hub = service.CreateHub()
	go service.hub.Run()
	return service
}

// CreateHub создает новый Hub
func (s *webSocketService) CreateHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *WebSocketMessage),
	}
}

// Run запускает hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.UserID] = client
			h.mutex.Unlock()

			// Отправляем подтверждение подключения
			message := &WebSocketMessage{
				Type: "connected",
				Data: map[string]interface{}{
					"user_id": client.UserID,
					"status":  "online",
				},
			}

			select {
			case client.Send <- message:
			default:
				close(client.Send)
				h.mutex.Lock()
				delete(h.clients, client.UserID)
				h.mutex.Unlock()
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for userID, client := range h.clients {
				// Если сообщение адресное, отправляем только получателю
				if message.To != uuid.Nil && message.To != userID {
					continue
				}

				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, userID)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// HandleClient обрабатывает WebSocket клиента
func (s *webSocketService) HandleClient(client *Client) {
	defer func() {
		s.hub.unregister <- client
		client.Conn.Close()
	}()

	// Настройка параметров соединения
	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Регистрируем клиента
	s.hub.register <- client

	// Запускаем горутины для чтения и записи
	go s.writePump(client)
	s.readPump(client)
}

// readPump читает сообщения от WebSocket клиента
func (s *webSocketService) readPump(client *Client) {
	defer func() {
		s.hub.unregister <- client
		client.Conn.Close()
	}()

	for {
		var message WebSocketMessage
		err := client.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket error", "error", err)
			}
			break
		}

		// Устанавливаем отправителя
		message.From = client.UserID

		// Обрабатываем разные типы сообщений
		switch message.Type {
		case "ping":
			response := &WebSocketMessage{
				Type: "pong",
				Data: map[string]interface{}{
					"timestamp": time.Now().Unix(),
				},
			}
			select {
			case client.Send <- response:
			default:
				close(client.Send)
				return
			}

		case "chat_message":
			// Обрабатываем сообщение чата
			s.handleChatMessage(client, &message)

		case "typing":
			// Уведомление о наборе текста
			s.handleTyping(client, &message)

		default:
			s.logger.Warn("Unknown message type", "type", message.Type)
		}
	}
}

// writePump отправляет сообщения WebSocket клиенту
func (s *webSocketService) writePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(message); err != nil {
				s.logger.Error("WebSocket write error", "error", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleChatMessage обрабатывает сообщение чата
func (s *webSocketService) handleChatMessage(client *Client, message *WebSocketMessage) {
	// Сохраняем сообщение в Redis для дальнейшей обработки
	messageData, _ := json.Marshal(message)
	ctx := context.Background()

	key := fmt.Sprintf("chat_message:%s", uuid.New().String())
	s.redis.Set(ctx, key, messageData, 5*time.Minute)

	// Публикуем в Redis для обработки другими сервисами
	s.redis.Publish(ctx, "chat_messages", messageData)

	// Отправляем получателю если он онлайн
	if message.To != uuid.Nil {
		s.SendMessageToUser(message.To, message)
	}
}

// handleTyping обрабатывает уведомление о наборе текста
func (s *webSocketService) handleTyping(client *Client, message *WebSocketMessage) {
	if message.To != uuid.Nil {
		typingMessage := &WebSocketMessage{
			Type: "typing",
			From: client.UserID,
			To:   message.To,
			Data: message.Data,
		}
		s.SendMessageToUser(message.To, typingMessage)
	}
}

// SendMessageToUser отправляет сообщение конкретному пользователю
func (s *webSocketService) SendMessageToUser(userID uuid.UUID, message *WebSocketMessage) error {
	s.hub.mutex.RLock()
	client, exists := s.hub.clients[userID]
	s.hub.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("user %s is not online", userID)
	}

	select {
	case client.Send <- message:
		return nil
	default:
		// Канал заблокирован, отключаем клиента
		s.hub.unregister <- client
		return fmt.Errorf("failed to send message to user %s", userID)
	}
}

// BroadcastToConversation отправляет сообщение всем участникам диалога
func (s *webSocketService) BroadcastToConversation(conversationID uuid.UUID, message *WebSocketMessage) error {
	message.Channel = fmt.Sprintf("conversation:%s", conversationID)

	select {
	case s.hub.broadcast <- message:
		return nil
	default:
		return fmt.Errorf("failed to broadcast message")
	}
}

// GetOnlineUsers возвращает список онлайн пользователей
func (s *webSocketService) GetOnlineUsers() []uuid.UUID {
	s.hub.mutex.RLock()
	defer s.hub.mutex.RUnlock()

	users := make([]uuid.UUID, 0, len(s.hub.clients))
	for userID := range s.hub.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline проверяет, онлайн ли пользователь
func (s *webSocketService) IsUserOnline(userID uuid.UUID) bool {
	s.hub.mutex.RLock()
	defer s.hub.mutex.RUnlock()

	_, exists := s.hub.clients[userID]
	return exists
}
