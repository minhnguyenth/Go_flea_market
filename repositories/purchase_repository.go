package repositories

import (
	"gin-freemarket/models"

	"gorm.io/gorm"
)

type IPurchaseRepository interface {
	Create(purchase *models.Purchase) error
	FindAll(userID uint) ([]models.Purchase, error)
	FindById(userID uint, id uint) (*models.Purchase, error)
}

type PurchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) IPurchaseRepository {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) Create(purchase *models.Purchase) error {
	return r.db.Create(purchase).Error
}

func (r *PurchaseRepository) FindAll(userID uint) ([]models.Purchase, error) {
	var purchases []models.Purchase
	err := r.db.Preload("User").Preload("Item").Where("user_id = ?", userID).Find(&purchases).Error
	return purchases, err
}

func (r *PurchaseRepository) FindById(userID uint, id uint) (*models.Purchase, error) {
	var purchase models.Purchase
	err := r.db.Preload("User").Preload("Item").Where("user_id = ? AND id = ?", userID, id).First(&purchase).Error
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}
