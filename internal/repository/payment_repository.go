package repository

import (
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetByID(id uuid.UUID) (*models.Payment, error)
	Update(payment *models.Payment) error
	Delete(id uuid.UUID) error
	GetByOrderID(orderID uuid.UUID) (*models.Payment, error)
	GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Payment, error)
	GetByStatus(status string, limit, offset int) ([]models.Payment, error)
	GetByGatewayID(gatewayID string) (*models.Payment, error)
	UpdateStatus(id uuid.UUID, status string) error
	GetByPaymentMethod(paymentMethod string, limit, offset int) ([]models.Payment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByID(id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("Order").Preload("Payer").First(&payment, "id = ?", id).Error
	return &payment, err
}

func (r *paymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Payment{}, "id = ?", id).Error
}

func (r *paymentRepository) GetByOrderID(orderID uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("Payer").First(&payment, "order_id = ?", orderID).Error
	return &payment, err
}

func (r *paymentRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Preload("Order").
		Where("payer_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) GetByStatus(status string, limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Preload("Order").Preload("Payer").
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) GetByGatewayID(gatewayID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("Order").Preload("Payer").First(&payment, "gateway_id = ?", gatewayID).Error
	return &payment, err
}

func (r *paymentRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Payment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *paymentRepository) GetByPaymentMethod(paymentMethod string, limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Preload("Order").Preload("Payer").
		Where("payment_method = ?", paymentMethod).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&payments).Error
	return payments, err
}
