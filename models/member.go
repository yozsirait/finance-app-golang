package models

import "gorm.io/gorm"

type Member struct {
	gorm.Model
	UserID   uint      `gorm:"not null"`
	Name     string    `gorm:"not null"`
	Accounts []Account `gorm:"foreignKey:MemberID"`
}
