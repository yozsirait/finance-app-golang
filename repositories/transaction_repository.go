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

/* ===========================
   Helpers
=========================== */

func (r *TransactionRepository) WithRelations(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Account", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, type, balance, currency, member_id")
		}).
		Preload("Category", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, type, user_id")
		}).
		Preload("Member", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, user_id")
		})
}

/* ===========================
   Queries
=========================== */

func (r *TransactionRepository) QueryBase(userID uint) *gorm.DB {
	return r.db.Model(&models.Transaction{}).
		Where("user_id = ?", userID)
}

func (r *TransactionRepository) FindByID(userID uint, id string) (*models.Transaction, error) {
	var trx models.Transaction
	err := r.WithRelations(r.db).
		Where("id = ? AND user_id = ?", id, userID).
		First(&trx).Error
	if err != nil {
		return nil, err
	}
	return &trx, nil
}

/* ===========================
   Mutations
=========================== */

func (r *TransactionRepository) Save(trx *models.Transaction) error {
	return r.db.Save(trx).Error
}

func (r *TransactionRepository) Create(trx *models.Transaction) error {
	return r.db.Create(trx).Error
}

func (r *TransactionRepository) Delete(trx *models.Transaction) error {
	return r.db.Delete(trx).Error
}
