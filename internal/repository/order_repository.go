package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *models.Order) error
	GetByID(id uuid.UUID) (*models.Order, error)
	Update(order *models.Order) error
	Delete(id uuid.UUID) error
	GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Order, error)
	GetBySellerID(sellerID uuid.UUID, limit, offset int) ([]models.Order, error)
	GetByStatus(status string, limit, offset int) ([]models.Order, error)
	GetExpiredOrders(limit int) ([]models.Order, error)
	UpdateStatus(id uuid.UUID, status string) error
	GetOrdersByProduct(productID uuid.UUID, limit, offset int) ([]models.Order, error)
	GetActiveOrdersCount(userID uuid.UUID) (int64, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Buyer").Preload("Seller").Preload("Product").Preload("Payment").
		First(&order, "id = ?", id).Error
	return &order, err
}

func (r *orderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Order{}, "id = ?", id).Error
}

func (r *orderRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Seller").Preload("Product").Preload("Payment").
		Where("buyer_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetBySellerID(sellerID uuid.UUID, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Buyer").Preload("Product").Preload("Payment").
		Where("seller_id = ?", sellerID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetByStatus(status string, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Buyer").Preload("Seller").Preload("Product").Preload("Payment").
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetExpiredOrders(limit int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Buyer").Preload("Seller").Preload("Product").
		Where("expires_at < ? AND status IN ?", time.Now(), []string{"pending", "paid"}).
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *orderRepository) GetOrdersByProduct(productID uuid.UUID, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Buyer").Preload("Seller").Preload("Payment").
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetActiveOrdersCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).
		Where("buyer_id = ? AND status IN ?", userID, []string{"pending", "paid", "delivered"}).
		Count(&count).Error
	return count, err
}
