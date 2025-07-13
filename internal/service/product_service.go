package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

type ProductService interface {
	CreateProduct(product *models.Product) error
	GetProduct(id uuid.UUID) (*models.Product, error)
	UpdateProduct(product *models.Product) error
	DeleteProduct(id uuid.UUID) error
	GetProducts(limit, offset int) ([]models.Product, error)
	GetUserProducts(userID uuid.UUID, limit, offset int) ([]models.Product, error)
	SearchProducts(query string, gameID *uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]models.Product, error)
	GetProductsByGame(gameID uuid.UUID, limit, offset int) ([]models.Product, error)
	GetProductsByCategory(categoryID uuid.UUID, limit, offset int) ([]models.Product, error)
	ViewProduct(id uuid.UUID) (*models.Product, error)
}

type productService struct {
	productRepo  repository.ProductRepository
	gameRepo     repository.GameRepository
	categoryRepo repository.CategoryRepository
	logger       *logger.Logger
}

func NewProductService(
	productRepo repository.ProductRepository,
	gameRepo repository.GameRepository,
	categoryRepo repository.CategoryRepository,
	logger *logger.Logger,
) ProductService {
	return &productService{
		productRepo:  productRepo,
		gameRepo:     gameRepo,
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

func (s *productService) CreateProduct(product *models.Product) error {
	// Валидация данных
	if product.Title == "" {
		return fmt.Errorf("title is required")
	}

	if product.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	if product.SellerID == uuid.Nil {
		return fmt.Errorf("seller_id is required")
	}

	if product.GameID == uuid.Nil {
		return fmt.Errorf("game_id is required")
	}

	if product.CategoryID == uuid.Nil {
		return fmt.Errorf("category_id is required")
	}

	// Проверка существования игры
	if _, err := s.gameRepo.GetByID(product.GameID); err != nil {
		return fmt.Errorf("game not found")
	}

	// Проверка существования категории
	if _, err := s.categoryRepo.GetByID(product.CategoryID); err != nil {
		return fmt.Errorf("category not found")
	}

	// Очистка и валидация типа продукта
	validTypes := []string{"account", "currency", "item", "service"}
	product.Type = strings.ToLower(product.Type)
	found := false
	for _, validType := range validTypes {
		if product.Type == validType {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid product type")
	}

	// Установка значений по умолчанию
	if product.Currency == "" {
		product.Currency = "USD"
	}
	if product.Quantity <= 0 {
		product.Quantity = 1
	}

	s.logger.Info("Creating product", "title", product.Title, "seller_id", product.SellerID)
	return s.productRepo.Create(product)
}

func (s *productService) GetProduct(id uuid.UUID) (*models.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *productService) UpdateProduct(product *models.Product) error {
	// Проверка существования продукта
	existing, err := s.productRepo.GetByID(product.ID)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	// Валидация прав доступа (должен быть владелец)
	if existing.SellerID != product.SellerID {
		return fmt.Errorf("unauthorized")
	}

	// Валидация данных
	if product.Title != "" && product.Title != existing.Title {
		existing.Title = product.Title
	}
	if product.Description != existing.Description {
		existing.Description = product.Description
	}
	if product.Price > 0 && product.Price != existing.Price {
		existing.Price = product.Price
	}
	if product.Quantity > 0 && product.Quantity != existing.Quantity {
		existing.Quantity = product.Quantity
	}
	if product.DeliveryData != existing.DeliveryData {
		existing.DeliveryData = product.DeliveryData
	}

	existing.IsActive = product.IsActive
	existing.IsAutoDelivery = product.IsAutoDelivery

	s.logger.Info("Updating product", "id", product.ID, "seller_id", product.SellerID)
	return s.productRepo.Update(existing)
}

func (s *productService) DeleteProduct(id uuid.UUID) error {
	return s.productRepo.Delete(id)
}

func (s *productService) GetProducts(limit, offset int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.productRepo.GetAll(limit, offset)
}

func (s *productService) GetUserProducts(userID uuid.UUID, limit, offset int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.productRepo.GetByUserID(userID, limit, offset)
}

func (s *productService) SearchProducts(query string, gameID *uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.productRepo.Search(query, gameID, categoryID, limit, offset)
}

func (s *productService) GetProductsByGame(gameID uuid.UUID, limit, offset int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.productRepo.GetByGameID(gameID, limit, offset)
}

func (s *productService) GetProductsByCategory(categoryID uuid.UUID, limit, offset int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.productRepo.GetByCategoryID(categoryID, limit, offset)
}

func (s *productService) ViewProduct(id uuid.UUID) (*models.Product, error) {
	// Увеличиваем счетчик просмотров
	if err := s.productRepo.IncrementViews(id); err != nil {
		s.logger.Error("Failed to increment views", "error", err, "product_id", id)
	}

	return s.productRepo.GetByID(id)
}
