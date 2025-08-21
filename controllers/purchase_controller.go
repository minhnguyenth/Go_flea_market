package controllers

import (
	"gin-freemarket/dto"
	"gin-freemarket/models"
	"gin-freemarket/services"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IPurchaseController interface {
	Create(c *gin.Context)
	FindAll(c *gin.Context)
	FindById(c *gin.Context)
}

type PurchaseController struct {
	purchaseService services.IPurchaseService
}

func NewPurchaseController(purchaseService services.IPurchaseService) IPurchaseController {
	return &PurchaseController{purchaseService: purchaseService}
}

func (c *PurchaseController) Create(ctx *gin.Context) {
	// Get user from context
	userID := c.GetUint("userID")

	var input dto.PurchaseItemInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Println("Failed to bind JSON in PurchaseController.Create", err)
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	purchase, err := c.purchaseService.Create(userID, input)
	if err != nil {
		log.Println("Failed to create purchase in PurchaseController.Create", input.ItemID, userID, err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, purchase)
}

func (c *PurchaseController) FindAll(ctx *gin.Context) {
	user, ok := ctx.Get("user")
	if !ok {
		log.Println("Failed to get user from context in PurchaseController.FindAll")
		ctx.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	userId := user.(*models.User).ID
	purchases, err := c.purchaseService.FindAll(userId)
	if err != nil {
		log.Println("Failed to find all purchases in PurchaseController.FindAll", userId, err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, purchases)
}

func (c *PurchaseController) FindById(ctx *gin.Context) {
	user, ok := ctx.Get("user")
	if !ok {
		log.Println("Failed to get user from context in PurchaseController.FindById")
		ctx.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	userId := user.(*models.User).ID

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid id"})
		return
	}

	purchase, err := c.purchaseService.FindById(userId, uint(id))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, purchase)
}
