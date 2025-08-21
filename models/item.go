package models

import "gorm.io/gorm"

type Item struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Price       uint   `gorm:"not null"`
	Description string
	SoldOut     bool `gorm:"not null; default:false"`
	Quantity    uint `gorm:"not null; default:1"`
	UserID      uint `gorm:"not null"`
}
