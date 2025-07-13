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

type GameHandler struct {
	gameService service.GameService
	logger      *logger.Logger
}

func NewGameHandler(gameService service.GameService, logger *logger.Logger) *GameHandler {
	return &GameHandler{
		gameService: gameService,
		logger:      logger,
	}
}

// GetGames получение списка игр
// @Summary Получить список игр
// @Description Получить список активных игр с пагинацией
// @Tags games
// @Produce json
// @Param limit query int false "Лимит игр" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Game
// @Router /games [get]
func (h *GameHandler) GetGames(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	games, err := h.gameService.GetActiveGames(limit, offset)
	if err != nil {
		h.logger.Error("Failed to get games", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, games)
}

// GetGame получение игры по slug
// @Summary Получить игру по slug
// @Description Получить информацию об игре по её slug
// @Tags games
// @Produce json
// @Param slug path string true "Slug игры"
// @Success 200 {object} models.Game
// @Router /games/{slug} [get]
func (h *GameHandler) GetGame(c *gin.Context) {
	slug := c.Param("slug")

	game, err := h.gameService.GetGameBySlug(slug)
	if err != nil {
		h.logger.Error("Failed to get game", "error", err, "slug", slug)
		response.Error(c, errors.ErrNotFound)
		return
	}

	response.Success(c, game)
}

// CreateGame создание игры
// @Summary Создать игру
// @Description Создать новую игру (только для админов)
// @Tags games
// @Accept json
// @Produce json
// @Param game body models.Game true "Данные игры"
// @Success 201 {object} models.Game
// @Router /games [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
	var game models.Game
	if err := c.ShouldBindJSON(&game); err != nil {
		response.ValidationError(c, "Invalid request data")
		return
	}

	if err := h.gameService.CreateGame(&game); err != nil {
		h.logger.Error("Failed to create game", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Created(c, game)
}

// UpdateGame обновление игры
// @Summary Обновить игру
// @Description Обновить информацию об игре (только для админов)
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "ID игры"
// @Param game body models.Game true "Данные игры"
// @Success 200 {object} models.Game
// @Router /games/{id} [put]
func (h *GameHandler) UpdateGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ValidationError(c, "Invalid game ID")
		return
	}

	var game models.Game
	if err := c.ShouldBindJSON(&game); err != nil {
		response.ValidationError(c, "Invalid request data")
		return
	}

	game.ID = id

	if err := h.gameService.UpdateGame(&game); err != nil {
		h.logger.Error("Failed to update game", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.Success(c, game)
}

// DeleteGame удаление игры
// @Summary Удалить игру
// @Description Удалить игру по ID (только для админов)
// @Tags games
// @Produce json
// @Param id path string true "ID игры"
// @Success 204
// @Router /games/{id} [delete]
func (h *GameHandler) DeleteGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ValidationError(c, "Invalid game ID")
		return
	}

	if err := h.gameService.DeleteGame(id); err != nil {
		h.logger.Error("Failed to delete game", "error", err)
		response.Error(c, errors.ErrInternal)
		return
	}

	response.NoContent(c)
}
