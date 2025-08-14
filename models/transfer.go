package models

import "gorm.io/gorm"

type Transfer struct {
	gorm.Model
	UserID        uint    `gorm:"not null"`
	MemberID      uint    `gorm:"not null"`
	FromAccountID uint    `gorm:"not null"`
	ToAccountID   uint    `gorm:"not null"`
	Amount        float64 `gorm:"not null"`
	Date          string  `gorm:"not null"`
	Description   string
	Fee           float64 `gorm:"default:0"`
}
