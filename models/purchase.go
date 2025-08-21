package models

import (
	"time"
)

type Purchase struct {
	ID         uint `gorm:"primaryKey"`
	UserID     uint `gorm:"not null;index"`
	ItemID     uint `gorm:"not null;index"`
	Price      int  `gorm:"not null"`
	Quantity   int  `gorm:"not null"`
	TotalPrice int  `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	User       User `gorm:"foreignKey:UserID;references:ID"`
	Item       Item `gorm:"foreignKey:ItemID;references:ID"`
}
