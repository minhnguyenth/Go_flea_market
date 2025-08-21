package controllers

import (
	"gin-freemarket/dto"
	"gin-freemarket/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IAuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type AuthController struct {
	authService services.IAuthService
}

func NewAuthController(authService services.IAuthService) IAuthController {
	return &AuthController{authService: authService}
}

func (c *AuthController) Register(ctx *gin.Context) {
	var request dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Println("Register failed : ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.authService.Register(request.Email, request.Password); err != nil {
		log.Println("Register failed : ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Register success"})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var request dto.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Println("Login failed : ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := c.authService.Login(request.Email, request.Password)
	if err != nil {
		log.Println("Login failed : ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Login success", "token": token})
}
