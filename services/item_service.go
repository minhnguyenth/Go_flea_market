package services

import (
	"errors"
	"gin-freemarket/dto"
	"gin-freemarket/models"
	"gin-freemarket/repositories"
	"log"

	"gorm.io/gorm"
)

type IItemService interface {
	FindAll() ([]models.Item, error)
	FindById(id uint) (models.Item, error)
	Create(item dto.CreateItemInput, userId uint) (*models.Item, error)
	Update(id uint, item dto.UpdateItemInput, userId uint) (*models.Item, error)
	Delete(id uint, userId uint) error
	DeductItemQuantity(itemID uint, quantity uint) error
}

type ItemService struct {
	itemRepository repositories.IItemRepository
	db             *gorm.DB
}

func NewItemService(itemRepository repositories.IItemRepository, db *gorm.DB) IItemService {
	return &ItemService{itemRepository: itemRepository, db: db}
}

func (s *ItemService) FindAll() ([]models.Item, error) {
	return s.itemRepository.FindAll()
}

func (s *ItemService) FindById(id uint) (models.Item, error) {
	return s.itemRepository.FindById(id)
}

func (s *ItemService) Create(item dto.CreateItemInput, userId uint) (*models.Item, error) {
	newItem := models.Item{
		Name:        item.Name,
		Price:       item.Price,
		Description: item.Description,
		SoldOut:     false,
		Quantity:    item.Quantity,
		UserID:      userId,
	}
	return s.itemRepository.Create(newItem, userId)
}

func (s *ItemService) Update(id uint, item dto.UpdateItemInput, userId uint) (*models.Item, error) {
	targetItem, err := s.itemRepository.FindById(id)
	if err != nil {
		return nil, err
	}

	if targetItem.UserID != userId {
		log.Println("Update failed : Item ID = ", id, ", User ID = ", userId, ", Error = ", "you are not authorized to update this item")
		return nil, errors.New("you are not authorized to update this item")
	}

	if item.Name != nil {
		targetItem.Name = *item.Name
	}
	if item.Price != nil {
		targetItem.Price = *item.Price
	}
	if item.Description != nil {
		targetItem.Description = *item.Description
	}
	if item.SoldOut != nil {
		targetItem.SoldOut = *item.SoldOut
	}
	if item.Quantity != nil {
		targetItem.Quantity = *item.Quantity
	}
	return s.itemRepository.Update(id, targetItem)
}

func (s *ItemService) Delete(id uint, userId uint) error {
	targetItem, err := s.itemRepository.FindById(id)
	if err != nil {
		return err
	}

	if targetItem.UserID != userId {
		log.Println("Delete failed : Item ID = ", id, ", User ID = ", userId, ", Error = ", "you are not authorized to delete this item")
		return errors.New("you are not authorized to delete this item")
	}

	log.Println("Delete success : Item ID = ", id, ", User ID = ", userId)
	return s.itemRepository.Delete(id)
}

func (s *ItemService) DeductItemQuantity(itemID uint, quantity uint) error {
	return s.itemRepository.DeductItemQuantity(itemID, quantity)
}

func (s *ItemService) Purchase(input dto.PurchaseItemInput) error {

	item, err := s.itemRepository.FindById(input.ItemID)
	if err != nil {
		log.Println("Purchase failed : Item ID = ", input.ItemID, ", Quantity = ", input.Quantity, ", Error = ", err)
		log.Println(err)
		return err
	}

	if item.SoldOut {
		log.Println("Item ? is sold out", item.ID)
		return errors.New("item is sold out")
	}

	if item.Quantity < input.Quantity {
		log.Println("Item ID = ", item.ID, " has not enough quantity (order quantity: ", input.Quantity, ", item quantity: ", item.Quantity, ")")
		return errors.New("quantity is not enough")
	}

	s.itemRepository.Lock()

	err = s.itemRepository.Purchase(input.ItemID, input.Quantity)
	if err != nil {
		s.itemRepository.Rollback()
		log.Println("Purchase failed : Item ID = ", input.ItemID, ", Quantity = ", input.Quantity, ", Error = ", err)
		log.Println(err)
		return err
	}

	item, err = s.itemRepository.FindById(input.ItemID)
	if err != nil {
		s.itemRepository.Rollback()
		log.Println("Purchase failed : Item ID = ", input.ItemID, ", Quantity = ", input.Quantity, ", Error = ", err)
		log.Println(err)
		return err
	}

	if item.Quantity == 0 {
		item.SoldOut = true
		_, err = s.itemRepository.Update(input.ItemID, item)
		if err != nil {
			s.itemRepository.Rollback()
			log.Println("Purchase failed : Item ID = ", input.ItemID, ", Quantity = ", input.Quantity, ", Error = ", err)
			log.Println(err)
			return err
		}
	}

	s.itemRepository.Commit()
	log.Println("Purchase success : Item ID = ", input.ItemID, ", Quantity = ", input.Quantity)
	return nil
}
