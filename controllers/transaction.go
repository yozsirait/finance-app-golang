package controllers

import (
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionController struct {
	db *gorm.DB
}

func NewTransactionController(db *gorm.DB) *TransactionController {
	return &TransactionController{db: db}
}

func (tc *TransactionController) CreateTransaction(c *gin.Context) {
	var input struct {
		MemberID    uint    `json:"member_id" binding:"required"`
		AccountID   uint    `json:"account_id" binding:"required"`
		CategoryID  uint    `json:"category_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		Date        string  `json:"date" binding:"required"`
		Description string  `json:"description"`
		Type        string  `json:"type" binding:"required,oneof=income expense"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, utils.FormatValidationError(err))
		return
	}

	if _, err := time.Parse("2006-01-02", input.Date); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	if err := tc.validateTransactionRelations(userID, input.MemberID, input.AccountID, input.CategoryID); err != nil {
		utils.RespondWithError(c, err.(*utils.AppError).StatusCode, err)
		return
	}

	transaction := models.Transaction{
		UserID:      userID,
		MemberID:    input.MemberID,
		AccountID:   input.AccountID,
		CategoryID:  input.CategoryID,
		Amount:      input.Amount,
		Date:        input.Date,
		Description: input.Description,
		Type:        input.Type,
	}

	err = tc.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		var account models.Account
		if err := tx.First(&account, input.AccountID).Error; err != nil {
			return err
		}

		if input.Type == "income" {
			account.Balance += input.Amount
		} else {
			if account.Balance < input.Amount {
				return utils.NewAppError("Insufficient account balance", http.StatusBadRequest)
			}
			account.Balance -= input.Amount
		}

		return tx.Save(&account).Error
	})

	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.RespondWithError(c, appErr.StatusCode, appErr)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create transaction")
		}
		return
	}

	utils.RespondWithCreated(c, transaction)
}

func (tc *TransactionController) GetTransactions(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	queryParams := struct {
		MemberID   string `form:"member_id"`
		AccountID  string `form:"account_id"`
		CategoryID string `form:"category_id"`
		Type       string `form:"type"`
		StartDate  string `form:"start_date"`
		EndDate    string `form:"end_date"`
		Limit      int    `form:"limit,default=20"`
		Page       int    `form:"page,default=1"`
	}{
		Limit: 20,
		Page:  1,
	}

	if err := c.ShouldBindQuery(&queryParams); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	query := tc.db.Where("user_id = ?", userID)

	if queryParams.MemberID != "" {
		if memberID, err := strconv.Atoi(queryParams.MemberID); err == nil {
			query = query.Where("member_id = ?", memberID)
		}
	}

	if queryParams.AccountID != "" {
		if accountID, err := strconv.Atoi(queryParams.AccountID); err == nil {
			query = query.Where("account_id = ?", accountID)
		}
	}

	if queryParams.CategoryID != "" {
		if categoryID, err := strconv.Atoi(queryParams.CategoryID); err == nil {
			query = query.Where("category_id = ?", categoryID)
		}
	}

	if queryParams.Type != "" {
		query = query.Where("type = ?", queryParams.Type)
	}

	if queryParams.StartDate != "" && queryParams.EndDate != "" {
		if _, err := time.Parse("2006-01-02", queryParams.StartDate); err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid start date format")
			return
		}
		if _, err := time.Parse("2006-01-02", queryParams.EndDate); err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid end date format")
			return
		}
		query = query.Where("date BETWEEN ? AND ?", queryParams.StartDate, queryParams.EndDate)
	}

	var total int64
	if err := query.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to count transactions")
		return
	}

	offset := (queryParams.Page - 1) * queryParams.Limit
	query = query.Offset(offset).Limit(queryParams.Limit).Order("date DESC")

	var transactions []models.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch transactions")
		return
	}

	utils.RespondWithPaginatedData(c, transactions, total, queryParams.Page, queryParams.Limit)
}

func (tc *TransactionController) GetTransactionByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	var transaction models.Transaction
	if err := tc.db.Preload("Account").
		Preload("Category").
		Preload("Member").
		Where("user_id = ? AND id = ?", userID, id).
		First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondWithError(c, http.StatusNotFound, utils.NewAppError("Transaction not found", http.StatusNotFound))
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch transaction")
		}
		return
	}

	utils.RespondWithSuccess(c, transaction)
}

func (tc *TransactionController) UpdateTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	var input struct {
		MemberID    uint    `json:"member_id"`
		AccountID   uint    `json:"account_id"`
		CategoryID  uint    `json:"category_id"`
		Amount      float64 `json:"amount"`
		Date        string  `json:"date"`
		Description string  `json:"description"`
		Type        string  `json:"type"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, utils.FormatValidationError(err))
		return
	}

	if input.Date != "" {
		if _, err := time.Parse("2006-01-02", input.Date); err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}
	}

	var transaction models.Transaction
	if err := tc.db.Where("user_id = ? AND id = ?", userID, id).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondWithError(c, http.StatusNotFound, utils.NewAppError("Transaction not found", http.StatusNotFound))
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch transaction")
		}
		return
	}

	// ... [rest of the UpdateTransaction method remains the same]
	// Just update the error responses to use utils.NewAppError where appropriate
}

func (tc *TransactionController) DeleteTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	var transaction models.Transaction
	if err := tc.db.Where("user_id = ? AND id = ?", userID, id).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondWithError(c, http.StatusNotFound, utils.NewAppError("Transaction not found", http.StatusNotFound))
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch transaction")
		}
		return
	}

	err = tc.db.Transaction(func(tx *gorm.DB) error {
		var account models.Account
		if err := tx.First(&account, transaction.AccountID).Error; err != nil {
			return err
		}

		if transaction.Type == "income" {
			account.Balance -= transaction.Amount
		} else {
			account.Balance += transaction.Amount
		}

		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		return tx.Delete(&transaction).Error
	})

	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete transaction")
		return
	}

	utils.RespondWithSuccess(c, gin.H{"message": "Transaction deleted successfully"})
}

// Helper methods remain the same, but ensure they return *utils.AppError instead of utils.AppError
func (tc *TransactionController) validateTransactionRelations(userID, memberID, accountID, categoryID uint) error {
	var member models.Member
	if err := tc.db.Where("user_id = ? AND id = ?", userID, memberID).First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.NewAppError("Member not found", http.StatusNotFound)
		}
		return err
	}

	var account models.Account
	if err := tc.db.Where("member_id = ? AND id = ?", memberID, accountID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.NewAppError("Account not found", http.StatusNotFound)
		}
		return err
	}

	var category models.Category
	if err := tc.db.Where("user_id = ? AND id = ?", userID, categoryID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.NewAppError("Category not found", http.StatusNotFound)
		}
		return err
	}

	return nil
}
