package services

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"finance-app/models"
	"finance-app/repositories"
	"finance-app/utils"

	"gorm.io/gorm"
)

var ErrInsufficientBalance = errors.New("insufficient account balance")

type TransactionService struct {
	db   *gorm.DB
	repo *repositories.TransactionRepository
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{
		db:   db,
		repo: repositories.NewTransactionRepository(db),
	}
}

/* ===========================
   Helpers
=========================== */

func (s *TransactionService) adjustAccountBalance(tx *gorm.DB, accountID uint, tType string, amount float64, apply bool) error {
	var acc models.Account
	if err := tx.First(&acc, accountID).Error; err != nil {
		return err
	}
	switch tType {
	case "expense":
		if apply {
			if acc.Balance < amount {
				return ErrInsufficientBalance
			}
			acc.Balance -= amount
		} else {
			acc.Balance += amount
		}
	case "income":
		if apply {
			acc.Balance += amount
		} else {
			acc.Balance -= amount
		}
	default:
		return fmt.Errorf("invalid type: %s", tType)
	}
	return tx.Save(&acc).Error
}

func (s *TransactionService) validateRelations(userID, memberID, accountID, categoryID uint) error {
	var cnt int64

	if err := s.db.Model(&models.Member{}).
		Where("user_id = ? AND id = ?", userID, memberID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return utils.NewAppError("Member not found", http.StatusNotFound)
	}

	cnt = 0
	if err := s.db.Model(&models.Account{}).
		Where("member_id = ? AND id = ?", memberID, accountID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return utils.NewAppError("Account not found", http.StatusNotFound)
	}

	cnt = 0
	if err := s.db.Model(&models.Category{}).
		Where("user_id = ? AND id = ?", userID, categoryID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return utils.NewAppError("Category not found", http.StatusNotFound)
	}
	return nil
}

/* ===========================
   Services
=========================== */

func (s *TransactionService) GetTransactions(userID uint, q TransactionQuery) ([]models.Transaction, int64, error) {
	query := s.repo.QueryBase(userID)

	if v := q.MemberID; v != "" {
		if id, e := strconv.Atoi(v); e == nil {
			query = query.Where("member_id = ?", id)
		}
	}
	if v := q.AccountID; v != "" {
		if id, e := strconv.Atoi(v); e == nil {
			query = query.Where("account_id = ?", id)
		}
	}
	if v := q.CategoryID; v != "" {
		if id, e := strconv.Atoi(v); e == nil {
			query = query.Where("category_id = ?", id)
		}
	}
	if v := q.Type; v != "" {
		query = query.Where("type = ?", v)
	}
	if q.StartDate != "" && q.EndDate != "" {
		query = query.Where("date BETWEEN ? AND ?", q.StartDate, q.EndDate)
	}
	if q.MinAmount > 0 {
		query = query.Where("amount >= ?", q.MinAmount)
	}
	if q.MaxAmount > 0 {
		query = query.Where("amount <= ?", q.MaxAmount)
	}
	if q.Description != "" {
		query = query.Where("description LIKE ?", "%"+q.Description+"%")
	}

	// Count dulu
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	allowedSort := map[string]bool{"date": true, "amount": true, "created_at": true, "id": true}
	sortBy := "date"
	if allowedSort[strings.ToLower(q.SortBy)] {
		sortBy = q.SortBy
	}
	sortOrder := "ASC"
	if strings.ToLower(q.SortOrder) == "desc" {
		sortOrder = "DESC"
	}

	offset := (q.Page - 1) * q.Limit
	var transactions []models.Transaction
	if err := query.
		Preload("Account", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,balance") }).
		Preload("Category", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,type") }).
		Preload("Member", func(db *gorm.DB) *gorm.DB { return db.Select("id,name") }).
		Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Offset(offset).
		Limit(q.Limit).
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (s *TransactionService) GetTransactionByID(userID uint, id string) (*models.Transaction, error) {
	return s.repo.FindByID(userID, id)
}

func (s *TransactionService) Create(userID uint, req map[string]interface{}) (*models.Transaction, error) {
	dateStr := req["date"].(string)
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return nil, utils.NewAppError("Invalid date format", http.StatusBadRequest)
	}

	memberID := uint(req["member_id"].(float64))
	accountID := uint(req["account_id"].(float64))
	categoryID := uint(req["category_id"].(float64))

	if err := s.validateRelations(userID, memberID, accountID, categoryID); err != nil {
		return nil, err
	}

	trx := models.Transaction{
		UserID:      userID,
		MemberID:    memberID,
		AccountID:   accountID,
		CategoryID:  categoryID,
		Amount:      req["amount"].(float64),
		Date:        dateStr,
		Description: req["description"].(string),
		Type:        req["type"].(string),
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}
		if err := s.adjustAccountBalance(tx, trx.AccountID, trx.Type, trx.Amount, true); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(userID, fmt.Sprint(trx.ID))
}

func (s *TransactionService) Update(userID uint, id string, req map[string]interface{}) (*models.Transaction, error) {
	existing, err := s.repo.FindByID(userID, id)
	if err != nil {
		return nil, err
	}

	// copy values
	newMemberID := existing.MemberID
	if v, ok := req["member_id"].(float64); ok {
		newMemberID = uint(v)
	}
	newAccountID := existing.AccountID
	if v, ok := req["account_id"].(float64); ok {
		newAccountID = uint(v)
	}
	newCategoryID := existing.CategoryID
	if v, ok := req["category_id"].(float64); ok {
		newCategoryID = uint(v)
	}
	newAmount := existing.Amount
	if v, ok := req["amount"].(float64); ok && v > 0 {
		newAmount = v
	}
	newDate := existing.Date
	if v, ok := req["date"].(string); ok && v != "" {
		if _, err := time.Parse("2006-01-02", v); err == nil {
			newDate = v
		}
	}
	newDesc := existing.Description
	if v, ok := req["description"].(string); ok {
		newDesc = v
	}
	newType := existing.Type
	if v, ok := req["type"].(string); ok && v != "" {
		newType = v
	}

	if err := s.validateRelations(userID, newMemberID, newAccountID, newCategoryID); err != nil {
		return nil, err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// rollback lama
		if err := s.adjustAccountBalance(tx, existing.AccountID, existing.Type, existing.Amount, false); err != nil {
			return err
		}

		// update
		existing.MemberID = newMemberID
		existing.AccountID = newAccountID
		existing.CategoryID = newCategoryID
		existing.Amount = newAmount
		existing.Date = newDate
		existing.Description = newDesc
		existing.Type = newType

		if err := tx.Save(existing).Error; err != nil {
			return err
		}

		// apply baru
		if err := s.adjustAccountBalance(tx, existing.AccountID, existing.Type, existing.Amount, true); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(userID, id)
}

func (s *TransactionService) Delete(userID uint, id string) error {
	trx, err := s.repo.FindByID(userID, id)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.adjustAccountBalance(tx, trx.AccountID, trx.Type, trx.Amount, false); err != nil {
			return err
		}
		return s.repo.Delete(trx)
	})
}
