package infra

import (
	"log"

	"github.com/joho/godotenv"
)

func Initialize() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		log.Println("Not using .env file")
	}
}
