package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

type GameService interface {
	CreateGame(game *models.Game) error
	GetGame(id uuid.UUID) (*models.Game, error)
	GetGameBySlug(slug string) (*models.Game, error)
	UpdateGame(game *models.Game) error
	DeleteGame(id uuid.UUID) error
	GetGames(limit, offset int) ([]models.Game, error)
	GetActiveGames(limit, offset int) ([]models.Game, error)
	SearchGames(query string, limit, offset int) ([]models.Game, error)
}

type gameService struct {
	gameRepo repository.GameRepository
	logger   *logger.Logger
}

func NewGameService(gameRepo repository.GameRepository, logger *logger.Logger) GameService {
	return &gameService{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

func (s *gameService) CreateGame(game *models.Game) error {
	// Валидация данных
	if game.Name == "" {
		return fmt.Errorf("name is required")
	}

	if game.Slug == "" {
		return fmt.Errorf("slug is required")
	}

	// Нормализация slug
	game.Slug = strings.ToLower(strings.ReplaceAll(game.Slug, " ", "-"))

	s.logger.Info("Creating game", "name", game.Name, "slug", game.Slug)
	return s.gameRepo.Create(game)
}

func (s *gameService) GetGame(id uuid.UUID) (*models.Game, error) {
	return s.gameRepo.GetByID(id)
}

func (s *gameService) GetGameBySlug(slug string) (*models.Game, error) {
	return s.gameRepo.GetBySlug(slug)
}

func (s *gameService) UpdateGame(game *models.Game) error {
	// Проверка существования игры
	existing, err := s.gameRepo.GetByID(game.ID)
	if err != nil {
		return fmt.Errorf("game not found")
	}

	// Обновление полей
	if game.Name != "" && game.Name != existing.Name {
		existing.Name = game.Name
	}
	if game.Description != existing.Description {
		existing.Description = game.Description
	}
	if game.Icon != existing.Icon {
		existing.Icon = game.Icon
	}
	if game.Banner != existing.Banner {
		existing.Banner = game.Banner
	}
	existing.IsActive = game.IsActive

	s.logger.Info("Updating game", "id", game.ID, "name", existing.Name)
	return s.gameRepo.Update(existing)
}

func (s *gameService) DeleteGame(id uuid.UUID) error {
	s.logger.Info("Deleting game", "id", id)
	return s.gameRepo.Delete(id)
}

func (s *gameService) GetGames(limit, offset int) ([]models.Game, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.gameRepo.GetAll(limit, offset)
}

func (s *gameService) GetActiveGames(limit, offset int) ([]models.Game, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.gameRepo.GetActive(limit, offset)
}

func (s *gameService) SearchGames(query string, limit, offset int) ([]models.Game, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.gameRepo.Search(query, limit, offset)
}
