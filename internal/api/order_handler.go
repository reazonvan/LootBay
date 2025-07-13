//go:build !userservice
// +build !userservice

package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/errors"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/reazonvan/LootBay/pkg/response"
	"github.com/reazonvan/LootBay/pkg/validation"
)

type OrderHandler struct {
	orderService  service.OrderService
	escrowService service.EscrowService
	validator     *validation.Validator
	logger        *logger.Logger
}

func NewOrderHandler(orderService service.OrderService, escrowService service.EscrowService, logger *logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderService:  orderService,
		escrowService: escrowService,
		validator:     validation.NewValidator(),
		logger:        logger,
	}
}

type CreateOrderRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
	Message   string    `json:"message" validate:"omitempty,max=500"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending paid delivered completed cancelled disputed"`
}

type CreateDisputeRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type ResolveDisputeRequest struct {
	Resolution string `json:"resolution" binding:"required"`
}

// CreateOrder создание заказа
// @Summary Создать заказ
// @Description Создать новый заказ
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Данные заказа"
// @Success 201 {object} models.Order
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	order := &models.Order{
		BuyerID:   userID.(uuid.UUID),
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Message:   req.Message,
	}

	if err := h.orderService.CreateOrder(order); err != nil {
		h.logger.Error("Failed to create order", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Created(c, order)
}

// GetOrders получение списка заказов пользователя
// @Summary Получить заказы пользователя
// @Description Получить список заказов пользователя (покупки и продажи)
// @Tags orders
// @Produce json
// @Param type query string false "Тип заказов: 'purchases' или 'sales'"
// @Param limit query int false "Лимит заказов" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Order
// @Router /orders [get]
func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	orderType := c.DefaultQuery("type", "purchases")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var orders []models.Order
	var err error

	switch orderType {
	case "purchases":
		orders, err = h.orderService.GetUserOrders(userID.(uuid.UUID), limit, offset)
	case "sales":
		orders, err = h.orderService.GetSellerOrders(userID.(uuid.UUID), limit, offset)
	default:
		response.Error(c, errors.ErrValidation)
		return
	}

	if err != nil {
		h.logger.Error("Failed to get orders", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, orders)
}

// GetOrder получение заказа по ID
// @Summary Получить заказ по ID
// @Description Получить информацию о заказе по его ID
// @Tags orders
// @Produce json
// @Param id path string true "ID заказа"
// @Success 200 {object} models.Order
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid order ID")
		return
	}

	order, err := h.orderService.GetOrder(orderID)
	if err != nil {
		h.logger.Error("Failed to get order", "error", err)
		response.Error(c, errors.ErrNotFound)
		return
	}

	response.Success(c, order)
}

// UpdateOrderStatus обновление статуса заказа
// @Summary Обновить статус заказа
// @Description Обновить статус заказа (только для продавца)
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "ID заказа"
// @Param status body UpdateOrderStatusRequest true "Новый статус"
// @Success 200 {object} models.Order
// @Router /orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid order ID")
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.orderService.UpdateOrderStatus(orderID, req.Status); err != nil {
		h.logger.Error("Failed to update order status", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	// Получаем обновлённый заказ для ответа
	order, _ := h.orderService.GetOrder(orderID)
	response.Success(c, order)
}

// CancelOrder отмена заказа
// @Summary Отменить заказ
// @Description Отменить заказ (покупатель или продавец)
// @Tags orders
// @Produce json
// @Param id path string true "ID заказа"
// @Success 200 {object} models.Order
// @Router /orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid order ID")
		return
	}

	err = h.orderService.CancelOrder(orderID, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to cancel order", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.NoContent(c)
}

// CompleteOrder завершение заказа
// @Summary Завершить заказ
// @Description Завершить заказ (только покупатель)
// @Tags orders
// @Produce json
// @Param id path string true "ID заказа"
// @Success 200 {object} models.Order
// @Router /orders/{id}/complete [post]
func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid order ID")
		return
	}

	if err := h.orderService.CompleteOrder(orderID, userID.(uuid.UUID)); err != nil {
		h.logger.Error("Failed to complete order", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	order, _ := h.orderService.GetOrder(orderID)
	response.Success(c, order)
}

// CreateDispute создание спора
// @Summary Создать спор
// @Description Создать спор по заказу
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "ID заказа"
// @Param dispute body CreateDisputeRequest true "Данные спора"
// @Success 200 {object} models.Order
// @Router /orders/{id}/dispute [post]
func (h *OrderHandler) CreateDispute(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid order ID")
		return
	}

	var req CreateDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.escrowService.CreateDispute(orderID, userID.(uuid.UUID), req.Reason); err != nil {
		h.logger.Error("Failed to create dispute", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	updatedOrder, _ := h.orderService.GetOrder(orderID)
	response.Success(c, updatedOrder)
}

// ResolveDispute разрешение спора
// @Summary Разрешить спор
// @Description Разрешить спор по заказу (только для админов)
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "ID заказа"
// @Param resolution body ResolveDisputeRequest true "Решение спора"
// @Success 200 {object} models.Order
// @Router /orders/{id}/resolve [post]
func (h *OrderHandler) ResolveDispute(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.ValidationError(c, "Invalid order ID")
		return
	}

	var req ResolveDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.escrowService.ResolveDispute(orderID, req.Resolution, userID.(uuid.UUID)); err != nil {
		h.logger.Error("Failed to resolve dispute", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	updatedOrder, _ := h.orderService.GetOrder(orderID)
	response.Success(c, updatedOrder)
}
