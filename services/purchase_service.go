package services

import (
	"fmt"
	"gin-freemarket/dto"
	"gin-freemarket/models"
	"gin-freemarket/repositories"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IPurchaseService interface {
	Create(userID uint, input dto.PurchaseItemInput) (purchase *dto.PurchaseResponse, err error)
	FindAll(userID uint) ([]*dto.PurchaseResponse, error)
	FindById(userID uint, id uint) (*dto.PurchaseResponse, error)
}

type PurchaseService struct {
	purchaseRepository repositories.IPurchaseRepository
	itemRepository     repositories.IItemRepository
	db                 *gorm.DB
}

func NewPurchaseService(
	purchaseRepository repositories.IPurchaseRepository,
	itemRepository repositories.IItemRepository,
	db *gorm.DB,
) IPurchaseService {
	return &PurchaseService{
		purchaseRepository: purchaseRepository,
		itemRepository:     itemRepository,
		db:                 db,
	}
}

// This implementation handles transactions across multiple tables to ensure consistency between tables, so it's implemented within the Service.
func (s *PurchaseService) Create(userID uint, input dto.PurchaseItemInput) (purchase *dto.PurchaseResponse, err error) {
	// Start transaction
	// Begin gets a new context
	// Lock is held until commit / rollback / transaction timeout
	// To set timeout, use tx.Exec("SET LOCAL statement_timeout = '5s'")
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Verify user
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get item with exclusive lock
	// Postgres / Mysql uses row lock. SQLite uses table lock
	// SELECT * from items where id = ? for update
	var item models.Item
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, input.ItemID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("item not found: %w", err)
	}

	// Check stock
	if item.Quantity <= input.Quantity {
		tx.Rollback()
		return nil, fmt.Errorf("item out of stock")
	}

	// Reduce stock
	item.Quantity -= input.Quantity
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create purchase record
	purchaseModel := &models.Purchase{
		UserID:     userID,
		ItemID:     input.ItemID,
		Price:      int(item.Price),
		Quantity:   int(input.Quantity),
		TotalPrice: int(item.Price * uint(input.Quantity)),
	}

	if err := tx.Create(purchaseModel).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Verify data within transaction as a precaution
	var createdPurchase models.Purchase
	if err := tx.Preload("User").Preload("Item").Where("id = ? AND user_id = ?", purchaseModel.ID, userID).
		First(&createdPurchase).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to verify created data: %w", err)
	}

	// Data validation
	if createdPurchase.ID != purchaseModel.ID ||
		createdPurchase.UserID != purchaseModel.UserID ||
		createdPurchase.ItemID != purchaseModel.ItemID {
		tx.Rollback()
		return nil, fmt.Errorf("data validation failed")
	}

	// Commit if no issues
	if err = tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dto.ToPurchaseResponse(&createdPurchase), nil
}

func (s *PurchaseService) FindAll(userID uint) ([]*dto.PurchaseResponse, error) {
	purchases, err := s.purchaseRepository.FindAll(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.PurchaseResponse, len(purchases))
	for i, purchase := range purchases {
		responses[i] = dto.ToPurchaseResponse(&purchase)
	}

	return responses, nil
}

func (s *PurchaseService) FindById(userID uint, id uint) (*dto.PurchaseResponse, error) {
	purchase, err := s.purchaseRepository.FindById(userID, id)
	if err != nil {
		return nil, err
	}

	return dto.ToPurchaseResponse(purchase), nil
}
