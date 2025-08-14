package models

import "gorm.io/gorm"

type SavingTarget struct {
	gorm.Model
	UserID        uint    `gorm:"not null"`
	MemberID      uint    `gorm:"not null"`
	AccountID     uint    `gorm:"not null"`
	Name          string  `gorm:"not null"`
	TargetAmount  float64 `gorm:"not null"`
	CurrentAmount float64 `gorm:"not null;default:0"`
	TargetDate    string  `gorm:"not null"`
	Description   string
}
