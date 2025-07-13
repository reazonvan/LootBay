//go:build !userservice
// +build !userservice

package api

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/reazonvan/LootBay/pkg/response"
)

type PaymentHandler struct {
	paymentService service.PaymentService
	stripeService  service.StripeService
	paypalService  service.PaypalService
	logger         *logger.Logger
}

func NewPaymentHandler(
	paymentService service.PaymentService,
	stripeService service.StripeService,
	paypalService service.PaypalService,
	logger *logger.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		stripeService:  stripeService,
		paypalService:  paypalService,
		logger:         logger,
	}
}

type CreatePaymentRequest struct {
	OrderID       uuid.UUID `json:"order_id" binding:"required"`
	PaymentMethod string    `json:"payment_method" binding:"required"`
}

type DepositRequest struct {
	Amount        float64 `json:"amount" binding:"required,min=0.01"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
}

type WithdrawRequest struct {
	Amount      float64 `json:"amount" binding:"required,min=0.01"`
	Destination string  `json:"destination" binding:"required"`
}

type RefundRequest struct {
	Amount float64 `json:"amount" binding:"required,min=0.01"`
}

// CreatePayment создание платежа
// @Summary Создать платеж
// @Description Создать новый платеж для заказа
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body CreatePaymentRequest true "Данные платежа"
// @Success 201 {object} models.Payment
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorMessage(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	payment, err := h.paymentService.ProcessPayment(req.OrderID, req.PaymentMethod, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to create payment", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithCode(c, http.StatusCreated, payment)
}

// GetPayment получение платежа по ID
// @Summary Получить платеж по ID
// @Description Получить информацию о платеже по его ID
// @Tags payments
// @Produce json
// @Param id path string true "ID платежа"
// @Success 200 {object} models.Payment
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid payment ID")
		return
	}

	payment, err := h.paymentService.GetPayment(id)
	if err != nil {
		h.logger.Error("Failed to get payment", "error", err, "id", id)
		response.ErrorMessage(c, http.StatusNotFound, "Payment not found")
		return
	}

	response.SuccessWithCode(c, http.StatusOK, payment)
}

// GetPayments получение списка платежей пользователя
// @Summary Получить платежи пользователя
// @Description Получить список платежей пользователя
// @Tags payments
// @Produce json
// @Param limit query int false "Лимит платежей" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Payment
// @Router /payments [get]
func (h *PaymentHandler) GetPayments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorMessage(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	payments, err := h.paymentService.GetPaymentsByUser(userID.(uuid.UUID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get payments", "error", err)
		response.ErrorMessage(c, http.StatusInternalServerError, "Failed to get payments")
		return
	}

	response.SuccessWithCode(c, http.StatusOK, payments)
}

// CapturePayment подтверждение платежа
// @Summary Подтвердить платеж
// @Description Подтвердить и захватить платеж
// @Tags payments
// @Produce json
// @Param id path string true "ID платежа"
// @Success 200 {object} models.Payment
// @Router /payments/{id}/capture [post]
func (h *PaymentHandler) CapturePayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid payment ID")
		return
	}

	if err := h.paymentService.CapturePayment(id); err != nil {
		h.logger.Error("Failed to capture payment", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	payment, _ := h.paymentService.GetPayment(id)
	response.SuccessWithCode(c, http.StatusOK, payment)
}

// RefundPayment возврат платежа
// @Summary Вернуть платеж
// @Description Вернуть средства по платежу
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "ID платежа"
// @Param refund body RefundRequest true "Сумма возврата"
// @Success 200 {object} models.Payment
// @Router /payments/{id}/refund [post]
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid payment ID")
		return
	}

	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	if err := h.paymentService.RefundPayment(id, req.Amount); err != nil {
		h.logger.Error("Failed to refund payment", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	payment, _ := h.paymentService.GetPayment(id)
	response.SuccessWithCode(c, http.StatusOK, payment)
}

// DepositToBalance пополнение баланса
// @Summary Пополнить баланс
// @Description Пополнить баланс пользователя
// @Tags payments
// @Accept json
// @Produce json
// @Param deposit body DepositRequest true "Данные пополнения"
// @Success 200 {object} map[string]interface{}
// @Router /balance/deposit [post]
func (h *PaymentHandler) DepositToBalance(c *gin.Context) {
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorMessage(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.paymentService.DepositToBalance(userID.(uuid.UUID), req.Amount, req.PaymentMethod); err != nil {
		h.logger.Error("Failed to deposit to balance", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithCode(c, http.StatusOK, gin.H{
		"message": "Balance deposited successfully",
		"amount":  req.Amount,
	})
}

// WithdrawFromBalance вывод средств с баланса
// @Summary Вывести средства с баланса
// @Description Вывести средства с баланса пользователя
// @Tags payments
// @Accept json
// @Produce json
// @Param withdraw body WithdrawRequest true "Данные вывода"
// @Success 200 {object} map[string]interface{}
// @Router /balance/withdraw [post]
func (h *PaymentHandler) WithdrawFromBalance(c *gin.Context) {
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorMessage(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.paymentService.WithdrawFromBalance(userID.(uuid.UUID), req.Amount, req.Destination); err != nil {
		h.logger.Error("Failed to withdraw from balance", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithCode(c, http.StatusOK, gin.H{
		"message": "Withdrawal completed successfully",
		"amount":  req.Amount,
	})
}

// GetTransactions получение транзакций пользователя
// @Summary Получить транзакции пользователя
// @Description Получить историю транзакций пользователя
// @Tags payments
// @Produce json
// @Param limit query int false "Лимит транзакций" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Transaction
// @Router /balance/transactions [get]
func (h *PaymentHandler) GetTransactions(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response.ErrorMessage(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	_, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	_, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Здесь нужно добавить метод в TransactionRepository
	// transactions, err := h.transactionService.GetTransactionsByUser(userID.(uuid.UUID), limit, offset)

	// Пока возвращаем пустой массив
	response.SuccessWithCode(c, http.StatusOK, []interface{}{})
}

// StripeWebhook обработка webhook от Stripe
// @Summary Обработать webhook от Stripe
// @Description Обработать уведомление от Stripe
// @Tags payments
// @Accept json
// @Produce json
// @Success 200
// @Router /payments/webhook/stripe [post]
func (h *PaymentHandler) StripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	event, err := h.stripeService.VerifyWebhook(payload, signature)
	if err != nil {
		h.logger.Error("Failed to verify Stripe webhook", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid webhook")
		return
	}

	h.logger.Info("Stripe webhook received", "event", event)
	c.Status(http.StatusOK)
}

// PaypalWebhook обработка webhook от PayPal
// @Summary Обработать webhook от PayPal
// @Description Обработать уведомление от PayPal
// @Tags payments
// @Accept json
// @Produce json
// @Success 200
// @Router /payments/webhook/paypal [post]
func (h *PaymentHandler) PaypalWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	event, err := h.paypalService.VerifyWebhook(payload, headers)
	if err != nil {
		h.logger.Error("Failed to verify PayPal webhook", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid webhook")
		return
	}

	h.logger.Info("PayPal webhook received", "event", event)
	c.Status(http.StatusOK)
}
