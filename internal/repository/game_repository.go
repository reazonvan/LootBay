package repository

import (
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type GameRepository interface {
	Create(game *models.Game) error
	GetByID(id uuid.UUID) (*models.Game, error)
	GetBySlug(slug string) (*models.Game, error)
	Update(game *models.Game) error
	Delete(id uuid.UUID) error
	GetAll(limit, offset int) ([]models.Game, error)
	GetActive(limit, offset int) ([]models.Game, error)
	Search(query string, limit, offset int) ([]models.Game, error)
}

type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) Create(game *models.Game) error {
	return r.db.Create(game).Error
}

func (r *gameRepository) GetByID(id uuid.UUID) (*models.Game, error) {
	var game models.Game
	err := r.db.Preload("Categories").First(&game, "id = ?", id).Error
	return &game, err
}

func (r *gameRepository) GetBySlug(slug string) (*models.Game, error) {
	var game models.Game
	err := r.db.Preload("Categories").First(&game, "slug = ?", slug).Error
	return &game, err
}

func (r *gameRepository) Update(game *models.Game) error {
	return r.db.Save(game).Error
}

func (r *gameRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Game{}, "id = ?", id).Error
}

func (r *gameRepository) GetAll(limit, offset int) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Categories").
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) GetActive(limit, offset int) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Categories").
		Where("is_active = ?", true).
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) Search(query string, limit, offset int) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Categories").
		Where("is_active = ? AND (name ILIKE ? OR description ILIKE ?)", true, "%"+query+"%", "%"+query+"%").
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&games).Error
	return games, err
}
