package models

import "gorm.io/gorm"

type Transfer struct {
	gorm.Model
	UserID        uint
	MemberID      uint
	FromAccountID uint
	ToAccountID   uint
	Amount        float64
	Date          string
	Description   string
	Fee           float64

	Member      Member  `gorm:"foreignKey:MemberID"`
	FromAccount Account `gorm:"foreignKey:FromAccountID"`
	ToAccount   Account `gorm:"foreignKey:ToAccountID"`
}
