package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	UserID uint   `gorm:"not null"`
	Name   string `gorm:"not null"`
	Type   string `gorm:"not null"` // "income" or "expense"
}
