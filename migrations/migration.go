package main

import (
	"gin-freemarket/infra"
	"gin-freemarket/models"

	"gorm.io/gorm"
)

func main() {
	db := infra.SetupDB()

	// // Delete existing tables
	// db.Migrator().DropTable(&models.Purchase{})
	// db.Migrator().DropTable(&models.Item{})
	// db.Migrator().DropTable(&models.User{})

	// Control migration order
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. User table
		if err := tx.AutoMigrate(&models.User{}); err != nil {
			return err
		}

		// 2. Item table
		if err := tx.AutoMigrate(&models.Item{}); err != nil {
			return err
		}

		// 3. Purchase table (including foreign key constraints)
		if err := tx.AutoMigrate(&models.Purchase{}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}
}
