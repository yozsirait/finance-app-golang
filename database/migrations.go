package database

import "finance-app/models"

func MigrateDB() {
	DB.AutoMigrate(
		&models.User{},
		&models.Member{},
		&models.Account{},
		&models.Category{},
		&models.BudgetCategory{},
		&models.Transaction{},
		&models.RecurringTransaction{},
		&models.Transfer{},
		&models.SavingTarget{},
	)
}
