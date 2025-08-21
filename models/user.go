package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Items    []Item `gorm:"constraint:OnDelete:CASCADE;foreignKey:UserID"`
}
