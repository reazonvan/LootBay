package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User представляет пользователя системы
type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	Username         string         `gorm:"uniqueIndex;not null" json:"username"`
	Phone            string         `gorm:"uniqueIndex" json:"phone"`
	PasswordHash     string         `gorm:"not null" json:"-"`
	FirstName        string         `json:"first_name"`
	LastName         string         `json:"last_name"`
	Avatar           string         `json:"avatar"`
	IsEmailVerified  bool           `gorm:"default:false" json:"is_email_verified"`
	IsSeller         bool           `gorm:"default:false" json:"is_seller"`
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	TelegramID       *int64         `gorm:"index" json:"telegram_id"`
	TelegramUsername *string        `json:"telegram_username"`
	Balance          float64        `gorm:"default:0" json:"balance"`
	Rating           float64        `gorm:"default:0" json:"rating"`
	TotalReviews     int            `gorm:"default:0" json:"total_reviews"`
	PositiveReviews  int            `gorm:"default:0" json:"positive_reviews"`
	LastOnline       time.Time      `json:"last_online"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Products []Product `gorm:"foreignKey:SellerID" json:"products,omitempty"`
	Orders   []Order   `gorm:"foreignKey:BuyerID" json:"orders,omitempty"`
	Reviews  []Review  `gorm:"foreignKey:ReviewerID" json:"reviews,omitempty"`
}

// Game представляет игру
type Game struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Banner      string    `json:"banner"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Categories []Category `gorm:"many2many:game_categories;" json:"categories,omitempty"`
	Products   []Product  `gorm:"foreignKey:GameID" json:"products,omitempty"`
}

// Category представляет категорию товаров
type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Games    []Game    `gorm:"many2many:game_categories;" json:"games,omitempty"`
	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// Product представляет товар или услугу
type Product struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SellerID       uuid.UUID      `gorm:"not null;index" json:"seller_id"`
	GameID         uuid.UUID      `gorm:"not null;index" json:"game_id"`
	CategoryID     uuid.UUID      `gorm:"not null;index" json:"category_id"`
	Title          string         `gorm:"not null" json:"title"`
	Description    string         `json:"description"`
	Price          float64        `gorm:"not null" json:"price"`
	Currency       string         `gorm:"default:'USD'" json:"currency"`
	Type           string         `gorm:"not null" json:"type"` // account, currency, item, service
	Quantity       int            `gorm:"default:1" json:"quantity"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	IsAutoDelivery bool           `gorm:"default:false" json:"is_auto_delivery"`
	DeliveryData   interface{}    `gorm:"type:jsonb" json:"delivery_data"` // JSON data for auto-delivery
	Views          int            `gorm:"default:0" json:"views"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Seller   User     `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	Game     Game     `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Orders   []Order  `gorm:"foreignKey:ProductID" json:"orders,omitempty"`
	Reviews  []Review `gorm:"foreignKey:ProductID" json:"reviews,omitempty"`
}

// Order представляет заказ
type Order struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerID      uuid.UUID   `gorm:"not null;index" json:"buyer_id"`
	SellerID     uuid.UUID   `gorm:"not null;index" json:"seller_id"`
	ProductID    uuid.UUID   `gorm:"not null;index" json:"product_id"`
	PaymentID    *uuid.UUID  `gorm:"index" json:"payment_id"`
	Status       string      `gorm:"not null;default:'pending'" json:"status"` // pending, paid, delivered, completed, cancelled, disputed
	Amount       float64     `gorm:"not null" json:"amount"`
	Currency     string      `gorm:"default:'USD'" json:"currency"`
	Quantity     int         `gorm:"default:1" json:"quantity"`
	Message      string      `json:"message"`
	DeliveryData interface{} `gorm:"type:jsonb" json:"delivery_data"`
	ExpiresAt    *time.Time  `json:"expires_at"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`

	// Relationships
	Buyer   User     `gorm:"foreignKey:BuyerID" json:"buyer,omitempty"`
	Seller  User     `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	Product Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Payment *Payment `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
	Reviews []Review `gorm:"foreignKey:OrderID" json:"reviews,omitempty"`
}

// Payment представляет платеж
type Payment struct {
	ID              uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID         uuid.UUID   `gorm:"not null;index" json:"order_id"`
	PayerID         uuid.UUID   `gorm:"not null;index" json:"payer_id"`
	PaymentMethod   string      `gorm:"not null" json:"payment_method"` // card, paypal, crypto, balance
	Amount          float64     `gorm:"not null" json:"amount"`
	Currency        string      `gorm:"default:'USD'" json:"currency"`
	Fee             float64     `gorm:"default:0" json:"fee"`
	Status          string      `gorm:"not null;default:'pending'" json:"status"` // pending, completed, failed, refunded
	GatewayID       string      `json:"gateway_id"`
	GatewayResponse interface{} `gorm:"type:jsonb" json:"gateway_response"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	// Relationships
	Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Payer User  `gorm:"foreignKey:PayerID" json:"payer,omitempty"`
}

// Review представляет отзыв
type Review struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID    uuid.UUID `gorm:"not null;index" json:"order_id"`
	ProductID  uuid.UUID `gorm:"not null;index" json:"product_id"`
	ReviewerID uuid.UUID `gorm:"not null;index" json:"reviewer_id"`
	RevieweeID uuid.UUID `gorm:"not null;index" json:"reviewee_id"`
	Rating     int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment    string    `json:"comment"`
	IsVisible  bool      `gorm:"default:true" json:"is_visible"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	Order    Order   `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Product  Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Reviewer User    `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	Reviewee User    `gorm:"foreignKey:RevieweeID" json:"reviewee,omitempty"`
}

// Transaction представляет транзакцию баланса
type Transaction struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID  `gorm:"not null;index" json:"user_id"`
	OrderID       *uuid.UUID `gorm:"index" json:"order_id"`
	Type          string     `gorm:"not null" json:"type"` // deposit, withdrawal, payment, refund, fee
	Amount        float64    `gorm:"not null" json:"amount"`
	Currency      string     `gorm:"default:'USD'" json:"currency"`
	BalanceBefore float64    `gorm:"not null" json:"balance_before"`
	BalanceAfter  float64    `gorm:"not null" json:"balance_after"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`

	// Relationships
	User  User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// Notification представляет уведомление
type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"not null;index" json:"user_id"`
	Type      string     `gorm:"not null" json:"type"` // email, telegram, push
	Subject   string     `gorm:"not null" json:"subject"`
	Message   string     `gorm:"not null" json:"message"`
	Status    string     `gorm:"not null;default:'pending'" json:"status"` // pending, sent, failed
	SentAt    *time.Time `json:"sent_at"`
	CreatedAt time.Time  `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// UserSession представляет сессию пользователя
type UserSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AutoResponse представляет автоответ
type AutoResponse struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"not null;index" json:"user_id"`
	Trigger   string    `gorm:"not null" json:"trigger"`
	Response  string    `gorm:"not null" json:"response"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Conversation представляет диалог между пользователями
type Conversation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	User1ID      uuid.UUID  `gorm:"not null;index" json:"user1_id"`
	User2ID      uuid.UUID  `gorm:"not null;index" json:"user2_id"`
	OrderID      *uuid.UUID `gorm:"index" json:"order_id"`
	LastMessage  *string    `json:"last_message"`
	LastActivity time.Time  `gorm:"not null" json:"last_activity"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relationships
	User1    User      `gorm:"foreignKey:User1ID" json:"user1,omitempty"`
	User2    User      `gorm:"foreignKey:User2ID" json:"user2,omitempty"`
	Order    *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// Message представляет сообщение в чате
type Message struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ConversationID  uuid.UUID  `gorm:"not null;index" json:"conversation_id"`
	SenderID        uuid.UUID  `gorm:"not null;index" json:"sender_id"`
	ReceiverID      uuid.UUID  `gorm:"not null;index" json:"receiver_id"`
	Content         string     `gorm:"not null" json:"content"`
	MessageType     string     `gorm:"default:'text'" json:"message_type"` // text, image, file, system
	IsRead          bool       `gorm:"default:false" json:"is_read"`
	ReadAt          *time.Time `json:"read_at"`
	IsSystemMessage bool       `gorm:"default:false" json:"is_system_message"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relationships
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	Sender       User         `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receiver     User         `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
}
