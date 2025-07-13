package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

type PaymentService interface {
	CreatePayment(payment *models.Payment) error
	GetPayment(id uuid.UUID) (*models.Payment, error)
	UpdatePayment(payment *models.Payment) error
	GetPaymentsByUser(userID uuid.UUID, limit, offset int) ([]models.Payment, error)
	GetPaymentsByOrder(orderID uuid.UUID) (*models.Payment, error)
	ProcessPayment(orderID uuid.UUID, paymentMethod string, userID uuid.UUID) (*models.Payment, error)
	CapturePayment(id uuid.UUID) error
	RefundPayment(id uuid.UUID, amount float64) error
	DepositToBalance(userID uuid.UUID, amount float64, paymentMethod string) error
	WithdrawFromBalance(userID uuid.UUID, amount float64, destination string) error
}

type paymentService struct {
	paymentRepo     repository.PaymentRepository
	orderRepo       repository.OrderRepository
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepository
	logger          *logger.Logger
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	transactionRepo repository.TransactionRepository,
	userRepo repository.UserRepository,
	logger *logger.Logger,
) PaymentService {
	return &paymentService{
		paymentRepo:     paymentRepo,
		orderRepo:       orderRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		logger:          logger,
	}
}

func (s *paymentService) CreatePayment(payment *models.Payment) error {
	// Валидация данных
	if payment.OrderID == uuid.Nil {
		return fmt.Errorf("order_id is required")
	}

	if payment.PayerID == uuid.Nil {
		return fmt.Errorf("payer_id is required")
	}

	if payment.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if payment.PaymentMethod == "" {
		return fmt.Errorf("payment_method is required")
	}

	// Проверка существования заказа
	order, err := s.orderRepo.GetByID(payment.OrderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверка что плательщик является покупателем
	if order.BuyerID != payment.PayerID {
		return fmt.Errorf("unauthorized payer")
	}

	// Проверка статуса заказа
	if order.Status != "pending" {
		return fmt.Errorf("order is not pending")
	}

	// Установка значений по умолчанию
	if payment.Currency == "" {
		payment.Currency = order.Currency
	}

	if payment.Status == "" {
		payment.Status = "pending"
	}

	s.logger.Info("Creating payment", "order_id", payment.OrderID, "payer_id", payment.PayerID, "amount", payment.Amount)
	return s.paymentRepo.Create(payment)
}

func (s *paymentService) GetPayment(id uuid.UUID) (*models.Payment, error) {
	return s.paymentRepo.GetByID(id)
}

func (s *paymentService) UpdatePayment(payment *models.Payment) error {
	// Проверка существования платежа
	existing, err := s.paymentRepo.GetByID(payment.ID)
	if err != nil {
		return fmt.Errorf("payment not found")
	}

	// Обновление полей
	if payment.Status != "" && payment.Status != existing.Status {
		existing.Status = payment.Status
	}

	if payment.GatewayID != "" && payment.GatewayID != existing.GatewayID {
		existing.GatewayID = payment.GatewayID
	}

	if payment.GatewayResponse != "" {
		existing.GatewayResponse = payment.GatewayResponse
	}

	s.logger.Info("Updating payment", "id", payment.ID, "status", existing.Status)
	return s.paymentRepo.Update(existing)
}

func (s *paymentService) GetPaymentsByUser(userID uuid.UUID, limit, offset int) ([]models.Payment, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.paymentRepo.GetByUserID(userID, limit, offset)
}

func (s *paymentService) GetPaymentsByOrder(orderID uuid.UUID) (*models.Payment, error) {
	return s.paymentRepo.GetByOrderID(orderID)
}

func (s *paymentService) ProcessPayment(orderID uuid.UUID, paymentMethod string, userID uuid.UUID) (*models.Payment, error) {
	// Получение заказа
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}

	// Проверка прав доступа
	if order.BuyerID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Создание платежа
	payment := &models.Payment{
		OrderID:       orderID,
		PayerID:       userID,
		PaymentMethod: paymentMethod,
		Amount:        order.Amount,
		Currency:      order.Currency,
		Status:        "pending",
	}

	if err := s.CreatePayment(payment); err != nil {
		return nil, err
	}

	s.logger.Info("Processing payment", "order_id", orderID, "payment_method", paymentMethod, "amount", order.Amount)
	return payment, nil
}

func (s *paymentService) CapturePayment(id uuid.UUID) error {
	payment, err := s.paymentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("payment not found")
	}

	if payment.Status != "pending" {
		return fmt.Errorf("payment is not pending")
	}

	// Обновление статуса платежа
	if err := s.paymentRepo.UpdateStatus(id, "completed"); err != nil {
		return err
	}

	// Обновление статуса заказа
	if err := s.orderRepo.UpdateStatus(payment.OrderID, "paid"); err != nil {
		return err
	}

	s.logger.Info("Payment captured", "id", id, "amount", payment.Amount)
	return nil
}

func (s *paymentService) RefundPayment(id uuid.UUID, amount float64) error {
	payment, err := s.paymentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("payment not found")
	}

	if payment.Status != "completed" {
		return fmt.Errorf("payment is not completed")
	}

	if amount <= 0 || amount > payment.Amount {
		return fmt.Errorf("invalid refund amount")
	}

	// Создание транзакции возврата
	transaction := &models.Transaction{
		UserID:      payment.PayerID,
		OrderID:     &payment.OrderID,
		Type:        "refund",
		Amount:      amount,
		Currency:    payment.Currency,
		Description: fmt.Sprintf("Refund for payment %s", payment.ID),
	}

	// Получение текущего баланса
	currentBalance, err := s.transactionRepo.GetUserBalance(payment.PayerID)
	if err != nil {
		return err
	}

	transaction.BalanceBefore = currentBalance
	transaction.BalanceAfter = currentBalance + amount

	if err := s.transactionRepo.CreateWithBalance(transaction, transaction.BalanceAfter); err != nil {
		return err
	}

	// Обновление статуса платежа
	if err := s.paymentRepo.UpdateStatus(id, "refunded"); err != nil {
		return err
	}

	s.logger.Info("Payment refunded", "id", id, "amount", amount)
	return nil
}

func (s *paymentService) DepositToBalance(userID uuid.UUID, amount float64, paymentMethod string) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	// Получение текущего баланса
	currentBalance, err := s.transactionRepo.GetUserBalance(userID)
	if err != nil {
		return err
	}

	// Создание транзакции пополнения
	transaction := &models.Transaction{
		UserID:        userID,
		Type:          "deposit",
		Amount:        amount,
		Currency:      "USD",
		BalanceBefore: currentBalance,
		BalanceAfter:  currentBalance + amount,
		Description:   fmt.Sprintf("Deposit via %s", paymentMethod),
	}

	if err := s.transactionRepo.CreateWithBalance(transaction, transaction.BalanceAfter); err != nil {
		return err
	}

	s.logger.Info("Balance deposited", "user_id", userID, "amount", amount, "method", paymentMethod)
	return nil
}

func (s *paymentService) WithdrawFromBalance(userID uuid.UUID, amount float64, destination string) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	// Получение текущего баланса
	currentBalance, err := s.transactionRepo.GetUserBalance(userID)
	if err != nil {
		return err
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient balance")
	}

	// Создание транзакции вывода
	transaction := &models.Transaction{
		UserID:        userID,
		Type:          "withdrawal",
		Amount:        -amount,
		Currency:      "USD",
		BalanceBefore: currentBalance,
		BalanceAfter:  currentBalance - amount,
		Description:   fmt.Sprintf("Withdrawal to %s", destination),
	}

	if err := s.transactionRepo.CreateWithBalance(transaction, transaction.BalanceAfter); err != nil {
		return err
	}

	s.logger.Info("Balance withdrawn", "user_id", userID, "amount", amount, "destination", destination)
	return nil
}
