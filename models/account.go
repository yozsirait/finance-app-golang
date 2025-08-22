package models

import "gorm.io/gorm"

const (
	AccountTypeBank    = "Bank"
	AccountTypeEWallet = "e-Wallet"
	AccountTypeCash    = "Cash"
)

var ValidAccountTypes = map[string]bool{
	AccountTypeBank:    true,
	AccountTypeEWallet: true,
	AccountTypeCash:    true,
}

type Account struct {
	gorm.Model
	MemberID uint    `gorm:"not null"`
	Name     string  `gorm:"not null"`
	Type     string  `gorm:"not null;check:type IN ('Bank','e-Wallet','Cash')"`
	Balance  float64 `gorm:"not null;default:0"`
	Currency string  `gorm:"not null;default:'IDR'"`

	// relasi
	Member Member `json:"Member" gorm:"foreignKey:MemberID"`
}
