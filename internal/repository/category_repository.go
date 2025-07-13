package repository

import (
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *models.Category) error
	GetByID(id uuid.UUID) (*models.Category, error)
	GetBySlug(slug string) (*models.Category, error)
	Update(category *models.Category) error
	Delete(id uuid.UUID) error
	GetAll(limit, offset int) ([]models.Category, error)
	GetActive(limit, offset int) ([]models.Category, error)
	GetByGameID(gameID uuid.UUID, limit, offset int) ([]models.Category, error)
	Search(query string, limit, offset int) ([]models.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) GetByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Games").First(&category, "id = ?", id).Error
	return &category, err
}

func (r *categoryRepository) GetBySlug(slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Games").First(&category, "slug = ?", slug).Error
	return &category, err
}

func (r *categoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Category{}, "id = ?", id).Error
}

func (r *categoryRepository) GetAll(limit, offset int) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Preload("Games").
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetActive(limit, offset int) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Preload("Games").
		Where("is_active = ?", true).
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetByGameID(gameID uuid.UUID, limit, offset int) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Joins("JOIN game_categories ON categories.id = game_categories.category_id").
		Where("game_categories.game_id = ? AND categories.is_active = ?", gameID, true).
		Order("categories.name ASC").
		Limit(limit).Offset(offset).
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Search(query string, limit, offset int) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Preload("Games").
		Where("is_active = ? AND (name ILIKE ? OR description ILIKE ?)", true, "%"+query+"%", "%"+query+"%").
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&categories).Error
	return categories, err
}
