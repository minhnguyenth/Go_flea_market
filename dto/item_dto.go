package dto

type CreateItemInput struct {
	Name        string `json:"name" binding:"required,min=3,max=200"`
	Price       uint   `json:"price" binding:"required,min=1,max=100000000"`
	Description string `json:"description" binding:"required,min=3,max=10000"`
	Quantity    uint   `json:"quantity" binding:"required,min=1"`
}

type UpdateItemInput struct {
	Name        *string `json:"name" binding:"omitnil,min=3,max=200"`
	Price       *uint   `json:"price" binding:"omitnil,min=1,max=100000000"`
	Description *string `json:"description" binding:"omitnil,min=3,max=10000"`
	SoldOut     *bool   `json:"sold_out" binding:"omitnil"`
	Quantity    *uint   `json:"quantity" binding:"omitnil,min=1"`
}
