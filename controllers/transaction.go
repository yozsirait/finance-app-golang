package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateTransaction(c *gin.Context) {
	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	transaction.UserID = userID

	db := database.GetDB()
	if err := db.Create(&transaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create transaction")
		return
	}

	// Update account balance
	var account models.Account
	if err := db.First(&account, transaction.AccountID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	if transaction.Type == "income" {
		account.Balance += transaction.Amount
	} else {
		account.Balance -= transaction.Amount
	}

	if err := db.Save(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update account balance")
		return
	}

	utils.RespondWithJSON(c, http.StatusCreated, transaction)
}

func GetTransactions(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	memberID := c.Query("member_id")
	accountID := c.Query("account_id")
	categoryID := c.Query("category_id")
	transactionType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	db := database.GetDB()
	query := db.Where("user_id = ?", userID)

	if memberID != "" {
		query = query.Where("member_id = ?", memberID)
	}
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}
	if startDate != "" && endDate != "" {
		query = query.Where("date BETWEEN ? AND ?", startDate, endDate)
	}

	var transactions []models.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch transactions")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, transactions)
}

func GetTransactionByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	db := database.GetDB()
	var transaction models.Transaction
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&transaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Transaction not found")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, transaction)
}

func UpdateTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	var input models.Transaction
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	db := database.GetDB()
	var transaction models.Transaction
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&transaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Transaction not found")
		return
	}

	// Revert old transaction amount from account balance
	var oldAccount models.Account
	if err := db.First(&oldAccount, transaction.AccountID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	if transaction.Type == "income" {
		oldAccount.Balance -= transaction.Amount
	} else {
		oldAccount.Balance += transaction.Amount
	}

	if err := db.Save(&oldAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to revert old transaction")
		return
	}

	// Update transaction
	transaction.MemberID = input.MemberID
	transaction.AccountID = input.AccountID
	transaction.CategoryID = input.CategoryID
	transaction.Amount = input.Amount
	transaction.Date = input.Date
	transaction.Description = input.Description
	transaction.Type = input.Type

	if err := db.Save(&transaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update transaction")
		return
	}

	// Apply new transaction amount to account balance
	var newAccount models.Account
	if err := db.First(&newAccount, transaction.AccountID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	if transaction.Type == "income" {
		newAccount.Balance += transaction.Amount
	} else {
		newAccount.Balance -= transaction.Amount
	}

	if err := db.Save(&newAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update account balance")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, transaction)
}

func DeleteTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	db := database.GetDB()
	var transaction models.Transaction
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&transaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Transaction not found")
		return
	}

	// Revert transaction amount from account balance
	var account models.Account
	if err := db.First(&account, transaction.AccountID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	if transaction.Type == "income" {
		account.Balance -= transaction.Amount
	} else {
		account.Balance += transaction.Amount
	}

	if err := db.Save(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to revert transaction")
		return
	}

	if err := db.Delete(&transaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete transaction")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}
