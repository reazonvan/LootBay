package repository

import (
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *models.Product) error
	GetByID(id uuid.UUID) (*models.Product, error)
	Update(product *models.Product) error
	Delete(id uuid.UUID) error
	GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Product, error)
	GetAll(limit, offset int) ([]models.Product, error)
	Search(query string, gameID *uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]models.Product, error)
	GetByGameID(gameID uuid.UUID, limit, offset int) ([]models.Product, error)
	GetByCategoryID(categoryID uuid.UUID, limit, offset int) ([]models.Product, error)
	IncrementViews(id uuid.UUID) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Seller").Preload("Game").Preload("Category").First(&product, "id = ?", id).Error
	return &product, err
}

func (r *productRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Product{}, "id = ?", id).Error
}

func (r *productRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Preload("Game").Preload("Category").
		Where("seller_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetAll(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Preload("Seller").Preload("Game").Preload("Category").
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&products).Error
	return products, err
}

func (r *productRepository) Search(query string, gameID *uuid.UUID, categoryID *uuid.UUID, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	db := r.db.Preload("Seller").Preload("Game").Preload("Category").
		Where("is_active = ?", true)

	if query != "" {
		db = db.Where("title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if gameID != nil {
		db = db.Where("game_id = ?", *gameID)
	}

	if categoryID != nil {
		db = db.Where("category_id = ?", *categoryID)
	}

	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

func (r *productRepository) GetByGameID(gameID uuid.UUID, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Preload("Seller").Preload("Category").
		Where("game_id = ? AND is_active = ?", gameID, true).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetByCategoryID(categoryID uuid.UUID, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Preload("Seller").Preload("Game").
		Where("category_id = ? AND is_active = ?", categoryID, true).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&products).Error
	return products, err
}

func (r *productRepository) IncrementViews(id uuid.UUID) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + 1")).Error
}
