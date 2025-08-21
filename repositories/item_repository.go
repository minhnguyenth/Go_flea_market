package repositories

import (
	"errors"
	"gin-freemarket/models"

	"gorm.io/gorm"
)

type IItemRepository interface {
	FindAll() ([]models.Item, error)
	FindById(id uint) (models.Item, error)
	Create(item models.Item, userId uint) (*models.Item, error)
	Update(id uint, item models.Item) (*models.Item, error)
	Delete(id uint) error
	Purchase(itemID uint, quantity uint) error
	DeductItemQuantity(itemID uint, quantity uint) error
	Lock()
	Commit()
	Rollback()
}

// ------------------------------------------------------------------------------------------------
// Item repository works with Postgresql
// -----------------------------------------------------------------------------------------------
type ItemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) IItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) FindAll() ([]models.Item, error) {
	var items []models.Item
	if err := r.db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ItemRepository) FindById(id uint) (models.Item, error) {
	var item models.Item
	if err := r.db.First(&item, id).Error; err != nil {
		return models.Item{}, err
	}
	return item, nil
}

func (r *ItemRepository) Create(item models.Item, userId uint) (*models.Item, error) {
	item.UserID = userId
	if err := r.db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ItemRepository) Update(id uint, updatedItem models.Item) (*models.Item, error) {
	if err := r.db.Model(&models.Item{}).Where("id = ?", id).Updates(&updatedItem).Error; err != nil {
		return nil, err
	}
	return &updatedItem, nil
}

func (r *ItemRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Item{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *ItemRepository) Purchase(itemID uint, quantity uint) error {
	if err := r.db.Model(&models.Item{}).Where("id = ?", itemID).Update("quantity", gorm.Expr("quantity - ?", quantity)).Error; err != nil {
		return err
	}
	return nil
}

func (r *ItemRepository) DeductItemQuantity(itemID uint, quantity uint) error {

	purchaseItem := models.Item{}
	r.db.Model(&purchaseItem).Select("quantity").First(&purchaseItem)

	if purchaseItem.Quantity < quantity {
		return errors.New("quantity is not enough")
	}

	if err := r.db.Model(&models.Item{}).Where("id = ?", itemID).Update("quantity", gorm.Expr("quantity - ?", quantity)).Error; err != nil {
		return err
	}
	return nil
}

func (r *ItemRepository) Lock() {
	r.db.Begin()
}

func (r *ItemRepository) Commit() {
	r.db.Commit()
}

func (r *ItemRepository) Rollback() {
	r.db.Rollback()
}
