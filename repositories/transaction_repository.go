package repositories

import (
	"finance-app/models"

	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db}
}

func (r *TransactionRepository) QueryBase(userID uint) *gorm.DB {
	return r.db.Model(&models.Transaction{}).Where("user_id = ?", userID)
}

func (r *TransactionRepository) FindByID(userID uint, id string) (*models.Transaction, error) {
	var trx models.Transaction
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Account", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,balance") }).
		Preload("Category", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,type") }).
		Preload("Member", func(db *gorm.DB) *gorm.DB { return db.Select("id,name") }).
		First(&trx).Error
	if err != nil {
		return nil, err
	}
	return &trx, nil
}

func (r *TransactionRepository) Save(trx *models.Transaction) error {
	return r.db.Save(trx).Error
}

func (r *TransactionRepository) Create(trx *models.Transaction) error {
	return r.db.Create(trx).Error
}

func (r *TransactionRepository) Delete(trx *models.Transaction) error {
	return r.db.Delete(trx).Error
}
