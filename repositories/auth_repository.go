package repositories

import (
	"gin-freemarket/models"
	"log"

	"gorm.io/gorm"
)

type IAuthRepository interface {
	CreateUser(user models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
}

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) IAuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(user models.User) error {

	newUser := &models.User{
		Email:    user.Email,
		Password: user.Password,
	}
	return r.db.Create(newUser).Error
}

func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		log.Println("GetUserByEmail failed : ", err)
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		log.Println("GetUserByID failed : ", err)
		return nil, err
	}
	return &user, nil
}
