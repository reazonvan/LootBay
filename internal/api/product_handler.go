//go:build !userservice
// +build !userservice

package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/reazonvan/LootBay/pkg/response"
)

type ProductHandler struct {
	productService service.ProductService
	logger         *logger.Logger
}

func NewProductHandler(productService service.ProductService, logger *logger.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		logger:         logger,
	}
}

// GetProducts получение списка товаров
// @Summary Получить список товаров
// @Description Получить список активных товаров с пагинацией
// @Tags products
// @Produce json
// @Param limit query int false "Лимит товаров" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Product
// @Router /products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	products, err := h.productService.GetProducts(limit, offset)
	if err != nil {
		h.logger.Error("Failed to get products", "error", err)
		response.ErrorMessage(c, http.StatusInternalServerError, "Failed to get products")
		return
	}

	response.SuccessWithCode(c, http.StatusOK, products)
}

// GetProduct получение товара по ID
// @Summary Получить товар по ID
// @Description Получить информацию о товаре по его ID
// @Tags products
// @Produce json
// @Param id path string true "ID товара"
// @Success 200 {object} models.Product
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product, err := h.productService.ViewProduct(id)
	if err != nil {
		h.logger.Error("Failed to get product", "error", err, "id", id)
		response.ErrorMessage(c, http.StatusNotFound, "Product not found")
		return
	}

	response.SuccessWithCode(c, http.StatusOK, product)
}

// SearchProducts поиск товаров
// @Summary Поиск товаров
// @Description Поиск товаров по параметрам
// @Tags products
// @Produce json
// @Param q query string false "Поисковый запрос"
// @Param game_id query string false "ID игры"
// @Param category_id query string false "ID категории"
// @Param limit query int false "Лимит товаров" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Product
// @Router /products/search [get]
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	gameIDStr := c.Query("game_id")
	categoryIDStr := c.Query("category_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var gameID, categoryID *uuid.UUID

	if gameIDStr != "" {
		if parsedGameID, err := uuid.Parse(gameIDStr); err == nil {
			gameID = &parsedGameID
		}
	}

	if categoryIDStr != "" {
		if parsedCategoryID, err := uuid.Parse(categoryIDStr); err == nil {
			categoryID = &parsedCategoryID
		}
	}

	products, err := h.productService.SearchProducts(query, gameID, categoryID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to search products", "error", err)
		response.ErrorMessage(c, http.StatusInternalServerError, "Failed to search products")
		return
	}

	response.SuccessWithCode(c, http.StatusOK, products)
}

// CreateProduct создание товара
// @Summary Создать товар
// @Description Создать новый товар
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.Product true "Данные товара"
// @Success 201 {object} models.Product
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	// Получение ID пользователя из токена
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorMessage(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	product.SellerID = userID.(uuid.UUID)

	if err := h.productService.CreateProduct(&product); err != nil {
		h.logger.Error("Failed to create product", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithCode(c, http.StatusCreated, product)
}

// UpdateProduct обновление товара
// @Summary Обновить товар
// @Description Обновить информацию о товаре
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "ID товара"
// @Param product body models.Product true "Данные товара"
// @Success 200 {object} models.Product
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	// Получение ID пользователя из токена
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorMessage(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	product.ID = id
	product.SellerID = userID.(uuid.UUID)

	if err := h.productService.UpdateProduct(&product); err != nil {
		h.logger.Error("Failed to update product", "error", err)
		response.ErrorMessage(c, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithCode(c, http.StatusOK, product)
}

// DeleteProduct удаление товара
// @Summary Удалить товар
// @Description Удалить товар по ID
// @Tags products
// @Produce json
// @Param id path string true "ID товара"
// @Success 204
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	if err := h.productService.DeleteProduct(id); err != nil {
		h.logger.Error("Failed to delete product", "error", err)
		response.ErrorMessage(c, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	c.Status(http.StatusNoContent)
}
