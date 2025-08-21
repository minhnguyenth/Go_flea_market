package controllers

import (
	"gin-freemarket/dto"
	"gin-freemarket/models"
	"gin-freemarket/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IItemController interface {
	FindAll(c *gin.Context)
	FindById(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type ItemController struct {
	itemService services.IItemService
}

func NewItemController(itemService services.IItemService) IItemController {
	return &ItemController{itemService: itemService}
}

func (c *ItemController) FindAll(ctx *gin.Context) {
	items, err := c.itemService.FindAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Println(err)
		return
	}
	ctx.JSON(http.StatusOK, items)
}

func (c *ItemController) FindById(ctx *gin.Context) {
	id := ctx.Param("id")
	itemId, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	item, err := c.itemService.FindById(uint(itemId))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (c *ItemController) Create(ctx *gin.Context) {

	user, exists := ctx.Get("user")
	if !exists {
		log.Println("No user Authenticated ")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId := user.(*models.User).ID

	var input dto.CreateItemInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := c.itemService.Create(input, userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (c *ItemController) Update(ctx *gin.Context) {

	user, exists := ctx.Get("user")
	if !exists {
		log.Println("No user Authenticated ")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId := user.(*models.User).ID

	id := ctx.Param("id")
	itemId, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	var input dto.UpdateItemInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Println(err)
		return
	}
	item, err := c.itemService.Update(uint(itemId), input, userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (c *ItemController) Delete(ctx *gin.Context) {

	user, exists := ctx.Get("user")
	if !exists {
		log.Println("No user Authenticated ")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId := user.(*models.User).ID

	id := ctx.Param("id")
	itemId, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	err = c.itemService.Delete(uint(itemId), userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

// func (c *ItemController) Purchase(ctx *gin.Context) {
// 	var input dto.PurchaseItemInput
// 	if err := ctx.ShouldBindJSON(&input); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	err := c.itemService.Purchase(input)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, gin.H{"message": "Item purchased successfully"})
// }
