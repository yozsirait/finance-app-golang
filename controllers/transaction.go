package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/* ===========================
   Helpers
=========================== */

var ErrInsufficientBalance = errors.New("insufficient account balance")

func adjustAccountBalance(tx *gorm.DB, accountID uint, tType string, amount float64, apply bool) error {
	var acc models.Account
	if err := tx.First(&acc, accountID).Error; err != nil {
		return err
	}
	switch tType {
	case "expense":
		if apply {
			// kurangi saldo
			if acc.Balance < amount {
				return ErrInsufficientBalance
			}
			acc.Balance -= amount
		} else {
			// rollback: kembalikan saldo
			acc.Balance += amount
		}
	case "income":
		if apply {
			acc.Balance += amount
		} else {
			acc.Balance -= amount
			if acc.Balance < 0 {
				// jaga2 jadi negatif saat rollback—boleh negatif? kalau tidak, nolkan/return error
				// Di sini kita biarkan boleh negatif agar konsisten; atur sesuai kebijakanmu.
			}
		}
	default:
		return fmt.Errorf("invalid type: %s", tType)
	}
	return tx.Save(&acc).Error
}

func validateTransactionRelations(db *gorm.DB, userID, memberID, accountID, categoryID uint) error {
	// Member milik user
	var cnt int64
	if err := db.Model(&models.Member{}).
		Where("user_id = ? AND id = ?", userID, memberID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return utils.NewAppError("Member not found", http.StatusNotFound)
	}

	// Account milik member tsb (mengikuti skema kamu)
	cnt = 0
	if err := db.Model(&models.Account{}).
		Where("member_id = ? AND id = ?", memberID, accountID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return utils.NewAppError("Account not found", http.StatusNotFound)
	}

	// Category milik user
	cnt = 0
	if err := db.Model(&models.Category{}).
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
   DTOs
=========================== */

type transactionCreateReq struct {
	MemberID    uint    `json:"member_id" binding:"required"`
	AccountID   uint    `json:"account_id" binding:"required"`
	CategoryID  uint    `json:"category_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Date        string  `json:"date" binding:"required"` // YYYY-MM-DD
	Description string  `json:"description"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
}

type transactionUpdateReq struct {
	MemberID    *uint    `json:"member_id,omitempty"`
	AccountID   *uint    `json:"account_id,omitempty"`
	CategoryID  *uint    `json:"category_id,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	Date        *string  `json:"date,omitempty"` // YYYY-MM-DD
	Description *string  `json:"description,omitempty"`
	Type        *string  `json:"type,omitempty" binding:"omitempty,oneof=income expense"`
}

/* ===========================
   Handlers
=========================== */

// GET /api/transactions
func GetTransactions(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}
	db := database.GetDB()

	// query params
	q := struct {
		MemberID    string  `form:"member_id"`
		AccountID   string  `form:"account_id"`
		CategoryID  string  `form:"category_id"`
		Type        string  `form:"type"`
		StartDate   string  `form:"start_date"`
		EndDate     string  `form:"end_date"`
		MinAmount   float64 `form:"min_amount"`
		MaxAmount   float64 `form:"max_amount"`
		Description string  `form:"description"`
		SortBy      string  `form:"sort_by,default=date"`
		SortOrder   string  `form:"sort_order,default=desc"`
		Limit       int     `form:"limit,default=20"`
		Page        int     `form:"page,default=1"`
	}{}
	if err := c.ShouldBindQuery(&q); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	query := db.Model(&models.Transaction{}).Where("user_id = ?", userID)

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

	// Count first
	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Count error")
		return
	}

	// Sorting safety
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
		utils.RespondWithError(c, http.StatusInternalServerError, "Query error")
		return
	}

	utils.RespondWithPaginatedData(c, transactions, total, q.Page, q.Limit)
}

// GET /api/transactions/:id
func GetTransactionByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}
	db := database.GetDB()

	id := c.Param("id")
	var trx models.Transaction
	if err := db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Account", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,balance") }).
		Preload("Category", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,type") }).
		Preload("Member", func(db *gorm.DB) *gorm.DB { return db.Select("id,name") }).
		First(&trx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(c, http.StatusNotFound, "Transaction not found")
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "DB error")
		}
		return
	}

	utils.RespondWithSuccess(c, trx)
}

// POST /api/transactions
func CreateTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}
	db := database.GetDB()

	var req transactionCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, utils.FormatValidationError(err))
		return
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// Validate FK ownerships
	if err := validateTransactionRelations(db, userID, req.MemberID, req.AccountID, req.CategoryID); err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.RespondWithError(c, appErr.StatusCode, appErr)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Validation error")
		}
		return
	}

	trx := models.Transaction{
		UserID:      userID,
		MemberID:    req.MemberID,
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Date:        req.Date,
		Description: req.Description,
		Type:        req.Type,
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}
		if err := adjustAccountBalance(tx, trx.AccountID, trx.Type, trx.Amount, true); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInsufficientBalance) {
			utils.RespondWithError(c, http.StatusBadRequest, "Insufficient account balance")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create transaction")
		return
	}

	// preload untuk response
	if err := db.Preload("Account", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,balance") }).
		Preload("Category", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,type") }).
		Preload("Member", func(db *gorm.DB) *gorm.DB { return db.Select("id,name") }).
		First(&trx, trx.ID).Error; err != nil {
		// fallback: kirim tanpa preload
	}

	utils.RespondWithCreated(c, trx)
}

// PUT /api/transactions/:id
func UpdateTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}
	db := database.GetDB()

	id := c.Param("id")
	var existing models.Transaction
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(c, http.StatusNotFound, "Transaction not found")
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "DB error")
		}
		return
	}

	var req transactionUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, utils.FormatValidationError(err))
		return
	}
	if req.Date != nil {
		if _, err := time.Parse("2006-01-02", *req.Date); err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}
	}

	// Siapkan nilai baru (fallback ke existing bila nil/zero)
	newMemberID := existing.MemberID
	if req.MemberID != nil {
		newMemberID = *req.MemberID
	}
	newAccountID := existing.AccountID
	if req.AccountID != nil {
		newAccountID = *req.AccountID
	}
	newCategoryID := existing.CategoryID
	if req.CategoryID != nil {
		newCategoryID = *req.CategoryID
	}
	newAmount := existing.Amount
	if req.Amount != nil && *req.Amount > 0 {
		newAmount = *req.Amount
	}
	newDate := existing.Date
	if req.Date != nil && *req.Date != "" {
		newDate = *req.Date
	}
	newDesc := existing.Description
	if req.Description != nil {
		newDesc = *req.Description // boleh kosong kalau mau clear
	}
	newType := existing.Type
	if req.Type != nil && *req.Type != "" {
		newType = *req.Type
	}

	// Validasi FK ownerships untuk nilai baru
	if err := validateTransactionRelations(db, userID, newMemberID, newAccountID, newCategoryID); err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.RespondWithError(c, appErr.StatusCode, appErr)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Validation error")
		}
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// rollback efek lama
		if err := adjustAccountBalance(tx, existing.AccountID, existing.Type, existing.Amount, false); err != nil {
			return err
		}

		// update record
		existing.MemberID = newMemberID
		existing.AccountID = newAccountID
		existing.CategoryID = newCategoryID
		existing.Amount = newAmount
		existing.Date = newDate
		existing.Description = newDesc
		existing.Type = newType

		if err := tx.Save(&existing).Error; err != nil {
			return err
		}

		// apply efek baru
		if err := adjustAccountBalance(tx, existing.AccountID, existing.Type, existing.Amount, true); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInsufficientBalance) {
			utils.RespondWithError(c, http.StatusBadRequest, "Insufficient account balance")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update transaction")
		return
	}

	// preload untuk response
	if err := db.Preload("Account", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,balance") }).
		Preload("Category", func(db *gorm.DB) *gorm.DB { return db.Select("id,name,type") }).
		Preload("Member", func(db *gorm.DB) *gorm.DB { return db.Select("id,name") }).
		First(&existing, existing.ID).Error; err != nil {
	}

	utils.RespondWithSuccess(c, existing)
}

// DELETE /api/transactions/:id
func DeleteTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}
	db := database.GetDB()

	id := c.Param("id")
	var trx models.Transaction
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&trx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(c, http.StatusNotFound, "Transaction not found")
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "DB error")
		}
		return
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// rollback effect
		if err := adjustAccountBalance(tx, trx.AccountID, trx.Type, trx.Amount, false); err != nil {
			return err
		}
		// delete
		return tx.Delete(&trx).Error
	}); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete transaction")
		return
	}

	utils.RespondWithSuccess(c, gin.H{"message": "Transaction deleted successfully"})
}
