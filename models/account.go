package models

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	MemberID uint    `gorm:"not null"`
	Name     string  `gorm:"not null"`
	Type     string  `gorm:"not null"` // e.g., "Bank", "Cash", "E-Wallet"
	Balance  float64 `gorm:"not null;default:0"`
	Currency string  `gorm:"not null;default:'IDR'"`
}
