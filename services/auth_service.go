// auth service
// handle user authentication and authorization
// to.takeuchi@raksul.com

package services

import (
	"fmt"
	"gin-freemarket/models"
	"gin-freemarket/repositories"
	"log"
	"os"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// salt is a random string that is used to hash the password
const SALT = "...salt..."

type IAuthService interface {
	Register(email string, password string) error
	Login(email string, password string) (*string, error)
	GetUserFromToken(token string) (*models.User, error)
}

type AuthService struct {
	authRepository repositories.IAuthRepository
	db             *gorm.DB
}

func NewAuthService(authRepository repositories.IAuthRepository, db *gorm.DB) IAuthService {
	return &AuthService{authRepository: authRepository, db: db}
}

func (s *AuthService) Register(email string, password string) error {
	saltedPassword := password + SALT
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.authRepository.CreateUser(models.User{Email: email, Password: string(hashedPassword)})
}

func (s *AuthService) Login(email string, password string) (*string, error) {
	user, err := s.authRepository.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	saltedPassword := password + SALT

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(saltedPassword))
	if err != nil {
		return nil, err
	}

	token, err := s.CreateToken(user.ID, user.Email)
	if err != nil {
		log.Println("Create token failed : ", err)
		return nil, err
	}

	return token, nil
}

// CreateToken creates a JWT token for the user
// check JWT in https://jwt.io/
func (s *AuthService) CreateToken(userId uint, email string) (*string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	// NOTE ------------------------------------------------------------
	// Signature with secret key.
	// To create secret key, use follow command
	// $ openssl rand -hex 64
	// ----------------------------------------------------------------
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Println("Create token failed : ", err)
		return nil, err
	}
	return &tokenString, nil
}

func (s *AuthService) GetUserFromToken(token string) (*models.User, error) {

	// decode token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Println("unexpected method: ", token.Header["alg"])
			return nil, fmt.Errorf("unexpected method: %s", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		log.Println("Get user from token failed : ", err)
		return nil, err
	}

	var user *models.User = nil
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		if float64(time.Now().Unix()) >= claims["exp"].(float64) {
			log.Println("token expired : " + claims["exp"].(string) + " user_id : " + claims["user_id"].(string))
			return nil, fmt.Errorf("token expired")
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			log.Println("user_id is not a float64")
			return nil, fmt.Errorf("user_id is not a float64")
		}

		email, ok := claims["email"].(string)
		if !ok {
			log.Println("email is not a string")
			return nil, fmt.Errorf("email is not a string")
		}

		user = &models.User{
			Model: gorm.Model{ID: uint(userID)},
			Email: email,
		}
	}

	return user, nil
}
