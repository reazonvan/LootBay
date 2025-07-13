package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

type EscrowService interface {
	HoldFunds(orderID uuid.UUID, userID uuid.UUID) error
	ReleaseFunds(orderID uuid.UUID, userID uuid.UUID) error
	RefundFunds(orderID uuid.UUID, userID uuid.UUID) error
	ProcessExpiredOrders() error
	CreateDispute(orderID uuid.UUID, userID uuid.UUID, reason string) error
	ResolveDispute(orderID uuid.UUID, resolution string, adminID uuid.UUID) error
}

type escrowService struct {
	orderRepo       repository.OrderRepository
	transactionRepo repository.TransactionRepository
	logger          *logger.Logger
}

func NewEscrowService(
	orderRepo repository.OrderRepository,
	transactionRepo repository.TransactionRepository,
	logger *logger.Logger,
) EscrowService {
	return &escrowService{
		orderRepo:       orderRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

func (s *escrowService) HoldFunds(orderID uuid.UUID, userID uuid.UUID) error {
	// Получение заказа
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка прав доступа
	if order.BuyerID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Проверка статуса
	if order.Status != "pending" {
		return fmt.Errorf("order must be pending to hold funds")
	}

	// Проверка баланса покупателя
	buyerBalance, err := s.transactionRepo.GetUserBalance(userID)
	if err != nil {
		return fmt.Errorf("failed to get buyer balance: %w", err)
	}

	if buyerBalance < order.Amount {
		return fmt.Errorf("insufficient balance")
	}

	// Создание транзакции блокировки средств
	transaction := &models.Transaction{
		UserID:        userID,
		OrderID:       &orderID,
		Type:          "payment",
		Amount:        -order.Amount, // Отрицательная сумма для списания
		Currency:      order.Currency,
		BalanceBefore: buyerBalance,
		BalanceAfter:  buyerBalance - order.Amount,
		Description:   fmt.Sprintf("Payment for order %s (held in escrow)", orderID),
	}

	// Списание средств с баланса покупателя
	if err := s.transactionRepo.CreateWithBalance(transaction, transaction.BalanceAfter); err != nil {
		return fmt.Errorf("failed to hold funds: %w", err)
	}

	// Обновление статуса заказа
	if err := s.orderRepo.UpdateStatus(orderID, "paid"); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Info("Funds held in escrow", "order_id", orderID, "user_id", userID, "amount", order.Amount)
	return nil
}

func (s *escrowService) ReleaseFunds(orderID uuid.UUID, userID uuid.UUID) error {
	// Получение заказа
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка прав доступа (только покупатель может освободить средства)
	if order.BuyerID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Проверка статуса
	if order.Status != "delivered" {
		return fmt.Errorf("order must be delivered to release funds")
	}

	// Получение текущего баланса продавца
	sellerBalance, err := s.transactionRepo.GetUserBalance(order.SellerID)
	if err != nil {
		return fmt.Errorf("failed to get seller balance: %w", err)
	}

	// Создание транзакции для перевода средств продавцу
	transaction := &models.Transaction{
		UserID:        order.SellerID,
		OrderID:       &orderID,
		Type:          "payment",
		Amount:        order.Amount,
		Currency:      order.Currency,
		BalanceBefore: sellerBalance,
		BalanceAfter:  sellerBalance + order.Amount,
		Description:   fmt.Sprintf("Payment received for order %s", orderID),
	}

	// Перевод средств продавцу
	if err := s.transactionRepo.CreateWithBalance(transaction, transaction.BalanceAfter); err != nil {
		return fmt.Errorf("failed to release funds: %w", err)
	}

	// Обновление статуса заказа
	if err := s.orderRepo.UpdateStatus(orderID, "completed"); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Info("Funds released from escrow", "order_id", orderID, "seller_id", order.SellerID, "amount", order.Amount)
	return nil
}

func (s *escrowService) RefundFunds(orderID uuid.UUID, userID uuid.UUID) error {
	// Получение заказа
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка прав доступа (покупатель или продавец)
	if order.BuyerID != userID && order.SellerID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Проверка статуса
	if order.Status != "paid" && order.Status != "disputed" {
		return fmt.Errorf("order must be paid or disputed to refund funds")
	}

	// Получение текущего баланса покупателя
	buyerBalance, err := s.transactionRepo.GetUserBalance(order.BuyerID)
	if err != nil {
		return fmt.Errorf("failed to get buyer balance: %w", err)
	}

	// Создание транзакции возврата
	transaction := &models.Transaction{
		UserID:        order.BuyerID,
		OrderID:       &orderID,
		Type:          "refund",
		Amount:        order.Amount,
		Currency:      order.Currency,
		BalanceBefore: buyerBalance,
		BalanceAfter:  buyerBalance + order.Amount,
		Description:   fmt.Sprintf("Refund for order %s", orderID),
	}

	// Возврат средств покупателю
	if err := s.transactionRepo.CreateWithBalance(transaction, transaction.BalanceAfter); err != nil {
		return fmt.Errorf("failed to refund funds: %w", err)
	}

	// Обновление статуса заказа
	if err := s.orderRepo.UpdateStatus(orderID, "cancelled"); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Info("Funds refunded from escrow", "order_id", orderID, "buyer_id", order.BuyerID, "amount", order.Amount)
	return nil
}

func (s *escrowService) ProcessExpiredOrders() error {
	// Получение истекших заказов
	expiredOrders, err := s.orderRepo.GetExpiredOrders(100)
	if err != nil {
		return fmt.Errorf("failed to get expired orders: %w", err)
	}

	s.logger.Info("Processing expired orders", "count", len(expiredOrders))

	for _, order := range expiredOrders {
		// Автоматический возврат средств для просроченных заказов
		if order.Status == "paid" {
			if err := s.RefundFunds(order.ID, order.BuyerID); err != nil {
				s.logger.Error("Failed to refund expired order", "order_id", order.ID, "error", err)
				continue
			}
		} else if order.Status == "pending" {
			// Отмена неоплаченных заказов
			if err := s.orderRepo.UpdateStatus(order.ID, "cancelled"); err != nil {
				s.logger.Error("Failed to cancel expired order", "order_id", order.ID, "error", err)
				continue
			}
		}
	}

	return nil
}

func (s *escrowService) CreateDispute(orderID uuid.UUID, userID uuid.UUID, reason string) error {
	// Получение заказа
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка прав доступа
	if order.BuyerID != userID && order.SellerID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Проверка статуса
	if order.Status != "paid" && order.Status != "delivered" {
		return fmt.Errorf("order must be paid or delivered to create dispute")
	}

	// Обновление статуса заказа на спорный
	if err := s.orderRepo.UpdateStatus(orderID, "disputed"); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Info("Dispute created", "order_id", orderID, "user_id", userID, "reason", reason)
	return nil
}

func (s *escrowService) ResolveDispute(orderID uuid.UUID, resolution string, adminID uuid.UUID) error {
	// Получение заказа
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка статуса
	if order.Status != "disputed" {
		return fmt.Errorf("order must be disputed to resolve")
	}

	// Разрешение спора
	switch resolution {
	case "refund":
		if err := s.RefundFunds(orderID, order.BuyerID); err != nil {
			return fmt.Errorf("failed to refund funds: %w", err)
		}
	case "release":
		if err := s.ReleaseFunds(orderID, order.BuyerID); err != nil {
			return fmt.Errorf("failed to release funds: %w", err)
		}
	default:
		return fmt.Errorf("invalid resolution: %s", resolution)
	}

	s.logger.Info("Dispute resolved", "order_id", orderID, "resolution", resolution, "admin_id", adminID)
	return nil
}
