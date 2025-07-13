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
)

type CategoryHandler struct {
	categoryService service.CategoryService
	logger          *logger.Logger
}

func NewCategoryHandler(categoryService service.CategoryService, logger *logger.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

// GetCategories получение списка категорий
// @Summary Получить список категорий
// @Description Получить список активных категорий с пагинацией
// @Tags categories
// @Produce json
// @Param limit query int false "Лимит категорий" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Category
// @Router /categories [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	categories, err := h.categoryService.GetActiveCategories(limit, offset)
	if err != nil {
		h.logger.Error("Failed to get categories", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, categories)
}

// GetCategory получение категории по slug
// @Summary Получить категорию по slug
// @Description Получить информацию о категории по её slug
// @Tags categories
// @Produce json
// @Param slug path string true "Slug категории"
// @Success 200 {object} models.Category
// @Router /categories/{slug} [get]
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.categoryService.GetCategoryBySlug(slug)
	if err != nil {
		h.logger.Error("Failed to get category", "error", err, "slug", slug)
		response.Error(c, errors.ErrNotFound)
		return
	}

	response.Success(c, category)
}

// CreateCategory создание категории
// @Summary Создать категорию
// @Description Создать новую категорию (только для админов)
// @Tags categories
// @Accept json
// @Produce json
// @Param category body models.Category true "Данные категории"
// @Success 201 {object} models.Category
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		response.ValidationError(c, "Invalid request data")
		return
	}

	if err := h.categoryService.CreateCategory(&category); err != nil {
		h.logger.Error("Failed to create category", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Created(c, category)
}

// UpdateCategory обновление категории
// @Summary Обновить категорию
// @Description Обновить информацию о категории (только для админов)
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "ID категории"
// @Param category body models.Category true "Данные категории"
// @Success 200 {object} models.Category
// @Router /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ValidationError(c, "Invalid category ID")
		return
	}

	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		response.ValidationError(c, "Invalid request data")
		return
	}

	category.ID = id

	if err := h.categoryService.UpdateCategory(&category); err != nil {
		h.logger.Error("Failed to update category", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, category)
}

// DeleteCategory удаление категории
// @Summary Удалить категорию
// @Description Удалить категорию по ID (только для админов)
// @Tags categories
// @Produce json
// @Param id path string true "ID категории"
// @Success 204
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ValidationError(c, "Invalid category ID")
		return
	}

	if err := h.categoryService.DeleteCategory(id); err != nil {
		h.logger.Error("Failed to delete category", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.NoContent(c)
}
