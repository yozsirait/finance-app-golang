package models

import "gorm.io/gorm"

type RecurringTransaction struct {
	gorm.Model
	UserID      uint    `gorm:"not null"`
	MemberID    uint    `gorm:"not null"`
	AccountID   uint    `gorm:"not null"`
	CategoryID  uint    `gorm:"not null"`
	Amount      float64 `gorm:"not null"`
	Description string
	Type        string `gorm:"not null"` // "income" or "expense"
	Frequency   string `gorm:"not null"` // "daily", "weekly", "monthly", "yearly"
	StartDate   string `gorm:"not null"`
	EndDate     string
	IsActive    bool `gorm:"default:true"`
}
