package models

import "gorm.io/gorm"

type BudgetCategory struct {
	gorm.Model
	UserID     uint    `gorm:"not null"`
	CategoryID uint    `gorm:"not null"`
	Amount     float64 `gorm:"not null"`
	Period     string  `gorm:"not null"` // "monthly", "weekly", "yearly"
}
