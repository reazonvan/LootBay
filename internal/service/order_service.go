package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

type OrderService interface {
	CreateOrder(order *models.Order) error
	GetOrder(id uuid.UUID) (*models.Order, error)
	UpdateOrder(order *models.Order) error
	CancelOrder(id uuid.UUID, userID uuid.UUID) error
	GetUserOrders(userID uuid.UUID, limit, offset int) ([]models.Order, error)
	GetSellerOrders(sellerID uuid.UUID, limit, offset int) ([]models.Order, error)
	UpdateOrderStatus(id uuid.UUID, status string) error
	GetOrdersByProduct(productID uuid.UUID, limit, offset int) ([]models.Order, error)
	CompleteOrder(id uuid.UUID, userID uuid.UUID) error
}

type orderService struct {
	orderRepo       repository.OrderRepository
	productRepo     repository.ProductRepository
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	logger          *logger.Logger
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	logger *logger.Logger,
) OrderService {
	return &orderService{
		orderRepo:       orderRepo,
		productRepo:     productRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

func (s *orderService) CreateOrder(order *models.Order) error {
	// Валидация данных
	if order.BuyerID == uuid.Nil {
		return fmt.Errorf("buyer_id is required")
	}

	if order.ProductID == uuid.Nil {
		return fmt.Errorf("product_id is required")
	}

	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// Получение товара
	product, err := s.productRepo.GetByID(order.ProductID)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	if !product.IsActive {
		return fmt.Errorf("product is not active")
	}

	// Проверка доступности товара
	if product.Quantity < order.Quantity {
		return fmt.Errorf("insufficient product quantity")
	}

	// Проверка что покупатель не является продавцом
	if order.BuyerID == product.SellerID {
		return fmt.Errorf("buyer cannot be the seller")
	}

	// Установка данных заказа
	order.SellerID = product.SellerID
	order.Amount = product.Price * float64(order.Quantity)
	order.Currency = product.Currency
	order.Status = "pending"
	order.ExpiresAt = &time.Time{}
	*order.ExpiresAt = time.Now().Add(24 * time.Hour) // Заказ истекает через 24 часа

	// Установка значений по умолчанию
	if order.Currency == "" {
		order.Currency = "USD"
	}

	s.logger.Info("Creating order", "buyer_id", order.BuyerID, "product_id", order.ProductID, "amount", order.Amount)
	return s.orderRepo.Create(order)
}

func (s *orderService) GetOrder(id uuid.UUID) (*models.Order, error) {
	return s.orderRepo.GetByID(id)
}

func (s *orderService) UpdateOrder(order *models.Order) error {
	// Проверка существования заказа
	existing, err := s.orderRepo.GetByID(order.ID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка статуса (некоторые поля можно изменить только в определенном статусе)
	if existing.Status == "completed" || existing.Status == "cancelled" {
		return fmt.Errorf("cannot update completed or cancelled order")
	}

	// Обновление доступных полей
	if order.Message != existing.Message {
		existing.Message = order.Message
	}

	if order.DeliveryData != existing.DeliveryData {
		existing.DeliveryData = order.DeliveryData
	}

	s.logger.Info("Updating order", "id", order.ID, "status", existing.Status)
	return s.orderRepo.Update(existing)
}

func (s *orderService) CancelOrder(id uuid.UUID, userID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка прав доступа
	if order.BuyerID != userID && order.SellerID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Проверка возможности отмены
	if order.Status != "pending" && order.Status != "paid" {
		return fmt.Errorf("cannot cancel order with status: %s", order.Status)
	}

	// Если заказ оплачен, нужно вернуть деньги
	if order.Status == "paid" {
		if err := s.refundOrder(order); err != nil {
			return fmt.Errorf("failed to refund order: %w", err)
		}
	}

	s.logger.Info("Cancelling order", "id", id, "user_id", userID, "status", order.Status)
	return s.orderRepo.UpdateStatus(id, "cancelled")
}

func (s *orderService) GetUserOrders(userID uuid.UUID, limit, offset int) ([]models.Order, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.orderRepo.GetByUserID(userID, limit, offset)
}

func (s *orderService) GetSellerOrders(sellerID uuid.UUID, limit, offset int) ([]models.Order, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.orderRepo.GetBySellerID(sellerID, limit, offset)
}

func (s *orderService) UpdateOrderStatus(id uuid.UUID, status string) error {
	// Валидация статуса
	validStatuses := []string{"pending", "paid", "delivered", "completed", "cancelled", "disputed"}
	found := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid status: %s", status)
	}

	s.logger.Info("Updating order status", "id", id, "status", status)
	return s.orderRepo.UpdateStatus(id, status)
}

func (s *orderService) GetOrdersByProduct(productID uuid.UUID, limit, offset int) ([]models.Order, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.orderRepo.GetOrdersByProduct(productID, limit, offset)
}

func (s *orderService) CompleteOrder(id uuid.UUID, userID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка прав доступа (только покупатель может завершить заказ)
	if order.BuyerID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Проверка статуса
	if order.Status != "delivered" {
		return fmt.Errorf("order must be delivered to complete")
	}

	// Перевод денег продавцу
	if err := s.transferToSeller(order); err != nil {
		return fmt.Errorf("failed to transfer funds to seller: %w", err)
	}

	s.logger.Info("Completing order", "id", id, "user_id", userID)
	return s.orderRepo.UpdateStatus(id, "completed")
}

func (s *orderService) refundOrder(order *models.Order) error {
	// Создание транзакции возврата
	transaction := &models.Transaction{
		UserID:      order.BuyerID,
		OrderID:     &order.ID,
		Type:        "refund",
		Amount:      order.Amount,
		Currency:    order.Currency,
		Description: fmt.Sprintf("Refund for order %s", order.ID),
	}

	// Получение текущего баланса
	currentBalance, err := s.transactionRepo.GetUserBalance(order.BuyerID)
	if err != nil {
		return err
	}

	transaction.BalanceBefore = currentBalance
	transaction.BalanceAfter = currentBalance + order.Amount

	return s.transactionRepo.CreateWithBalance(transaction, transaction.BalanceAfter)
}

func (s *orderService) transferToSeller(order *models.Order) error {
	// Создание транзакции для продавца
	transaction := &models.Transaction{
		UserID:      order.SellerID,
		OrderID:     &order.ID,
		Type:        "payment",
		Amount:      order.Amount,
		Currency:    order.Currency,
		Description: fmt.Sprintf("Payment for order %s", order.ID),
	}

	// Получение текущего баланса продавца
	currentBalance, err := s.transactionRepo.GetUserBalance(order.SellerID)
	if err != nil {
		return err
	}

	transaction.BalanceBefore = currentBalance
	transaction.BalanceAfter = currentBalance + order.Amount

	return s.transactionRepo.CreateWithBalance(transaction, transaction.BalanceAfter)
}
