package dto

import (
	"gin-freemarket/models"
	"time"
)

type PurchaseItemInput struct {
	ItemID   uint `json:"item_id" binding:"required,min=1"`
	Quantity uint `json:"quantity" binding:"required,min=1"`
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type ItemResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Price       uint   `json:"price"`
	Description string `json:"description"`
	SoldOut     bool   `json:"sold_out"`
	Quantity    int    `json:"quantity"`
}

type PurchaseResponse struct {
	ID         uint         `json:"id"`
	UserID     uint         `json:"user_id"`
	ItemID     uint         `json:"item_id"`
	Price      int          `json:"price"`
	Quantity   int          `json:"quantity"`
	TotalPrice int          `json:"total_price"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	User       UserResponse `json:"user"`
	Item       ItemResponse `json:"item"`
}

// ToPurchaseResponse returns purchase information including purchaser and item information
func ToPurchaseResponse(purchase *models.Purchase) *PurchaseResponse {
	return &PurchaseResponse{
		ID:         purchase.ID,
		UserID:     purchase.UserID,
		ItemID:     purchase.ItemID,
		Price:      purchase.Price,
		Quantity:   purchase.Quantity,
		TotalPrice: purchase.TotalPrice,
		CreatedAt:  purchase.CreatedAt,
		UpdatedAt:  purchase.UpdatedAt,
		User: UserResponse{
			ID: purchase.User.ID,
		},
		Item: ItemResponse{
			ID:          purchase.Item.ID,
			Name:        purchase.Item.Name,
			Price:       purchase.Item.Price,
			Description: purchase.Item.Description,
			Quantity:    int(purchase.Item.Quantity),
		},
	}
}
