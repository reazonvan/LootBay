package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

type CategoryService interface {
	CreateCategory(category *models.Category) error
	GetCategory(id uuid.UUID) (*models.Category, error)
	GetCategoryBySlug(slug string) (*models.Category, error)
	UpdateCategory(category *models.Category) error
	DeleteCategory(id uuid.UUID) error
	GetCategories(limit, offset int) ([]models.Category, error)
	GetActiveCategories(limit, offset int) ([]models.Category, error)
	GetCategoriesByGame(gameID uuid.UUID, limit, offset int) ([]models.Category, error)
	SearchCategories(query string, limit, offset int) ([]models.Category, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
	logger       *logger.Logger
}

func NewCategoryService(categoryRepo repository.CategoryRepository, logger *logger.Logger) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

func (s *categoryService) CreateCategory(category *models.Category) error {
	// Валидация данных
	if category.Name == "" {
		return fmt.Errorf("name is required")
	}

	if category.Slug == "" {
		return fmt.Errorf("slug is required")
	}

	// Нормализация slug
	category.Slug = strings.ToLower(strings.ReplaceAll(category.Slug, " ", "-"))

	s.logger.Info("Creating category", "name", category.Name, "slug", category.Slug)
	return s.categoryRepo.Create(category)
}

func (s *categoryService) GetCategory(id uuid.UUID) (*models.Category, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *categoryService) GetCategoryBySlug(slug string) (*models.Category, error) {
	return s.categoryRepo.GetBySlug(slug)
}

func (s *categoryService) UpdateCategory(category *models.Category) error {
	// Проверка существования категории
	existing, err := s.categoryRepo.GetByID(category.ID)
	if err != nil {
		return fmt.Errorf("category not found")
	}

	// Обновление полей
	if category.Name != "" && category.Name != existing.Name {
		existing.Name = category.Name
	}
	if category.Description != existing.Description {
		existing.Description = category.Description
	}
	if category.Icon != existing.Icon {
		existing.Icon = category.Icon
	}
	existing.IsActive = category.IsActive

	s.logger.Info("Updating category", "id", category.ID, "name", existing.Name)
	return s.categoryRepo.Update(existing)
}

func (s *categoryService) DeleteCategory(id uuid.UUID) error {
	s.logger.Info("Deleting category", "id", id)
	return s.categoryRepo.Delete(id)
}

func (s *categoryService) GetCategories(limit, offset int) ([]models.Category, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.categoryRepo.GetAll(limit, offset)
}

func (s *categoryService) GetActiveCategories(limit, offset int) ([]models.Category, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.categoryRepo.GetActive(limit, offset)
}

func (s *categoryService) GetCategoriesByGame(gameID uuid.UUID, limit, offset int) ([]models.Category, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.categoryRepo.GetByGameID(gameID, limit, offset)
}

func (s *categoryService) SearchCategories(query string, limit, offset int) ([]models.Category, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.categoryRepo.Search(query, limit, offset)
}
