package models

import "gorm.io/gorm"

type Transaction struct {
	gorm.Model
	UserID      uint    `gorm:"not null"`
	MemberID    uint    `gorm:"not null"`
	AccountID   uint    `gorm:"not null"`
	CategoryID  uint    `gorm:"not null"`
	Amount      float64 `gorm:"not null"`
	Date        string  `gorm:"not null"`
	Description string
	Type        string `gorm:"not null"` // "income" or "expense"

	// Relasi
	Member   Member   `gorm:"foreignKey:MemberID"`
	Account  Account  `gorm:"foreignKey:AccountID"`
	Category Category `gorm:"foreignKey:CategoryID"`
}
