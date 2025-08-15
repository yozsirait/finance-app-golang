package database

import (
	"finance-app/models"
	"log"
)

func MigrateDB() {
	err := DB.AutoMigrate(
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
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed")
}
