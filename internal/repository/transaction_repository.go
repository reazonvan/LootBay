package repository

import (
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(transaction *models.Transaction) error
	GetByID(id uuid.UUID) (*models.Transaction, error)
	GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Transaction, error)
	GetByOrderID(orderID uuid.UUID) ([]models.Transaction, error)
	GetByType(userID uuid.UUID, transactionType string, limit, offset int) ([]models.Transaction, error)
	GetUserBalance(userID uuid.UUID) (float64, error)
	CreateWithBalance(transaction *models.Transaction, newBalance float64) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) GetByID(id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Preload("User").Preload("Order").First(&transaction, "id = ?", id).Error
	return &transaction, err
}

func (r *transactionRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Preload("Order").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) GetByOrderID(orderID uuid.UUID) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Preload("User").
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) GetByType(userID uuid.UUID, transactionType string, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Preload("Order").
		Where("user_id = ? AND type = ?", userID, transactionType).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) GetUserBalance(userID uuid.UUID) (float64, error) {
	var user models.User
	err := r.db.Select("balance").First(&user, "id = ?", userID).Error
	return user.Balance, err
}

func (r *transactionRepository) CreateWithBalance(transaction *models.Transaction, newBalance float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Создаем транзакцию
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Обновляем баланс пользователя
		if err := tx.Model(&models.User{}).Where("id = ?", transaction.UserID).Update("balance", newBalance).Error; err != nil {
			return err
		}

		return nil
	})
}
