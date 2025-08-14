package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetRecurringTransactions(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	isActive := c.Query("is_active")

	db := database.GetDB()
	query := db.Where("user_id = ?", userID)

	if isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}

	var recurringTransactions []models.RecurringTransaction
	if err := query.Find(&recurringTransactions).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch recurring transactions")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, recurringTransactions)
}

func CreateRecurringTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		MemberID    uint    `json:"member_id" binding:"required"`
		AccountID   uint    `json:"account_id" binding:"required"`
		CategoryID  uint    `json:"category_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		Description string  `json:"description"`
		Type        string  `json:"type" binding:"required,oneof=income expense"`
		Frequency   string  `json:"frequency" binding:"required,oneof=daily weekly monthly yearly"`
		StartDate   string  `json:"start_date" binding:"required"`
		EndDate     string  `json:"end_date"`
		IsActive    bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Verify member belongs to user
	db := database.GetDB()
	var member models.Member
	if err := db.Where("user_id = ? AND id = ?", userID, input.MemberID).First(&member).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Member not found")
		return
	}

	// Verify account belongs to member
	var account models.Account
	if err := db.Where("member_id = ? AND id = ?", input.MemberID, input.AccountID).First(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	// Verify category belongs to user
	var category models.Category
	if err := db.Where("user_id = ? AND id = ?", userID, input.CategoryID).First(&category).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Category not found")
		return
	}

	recurringTransaction := models.RecurringTransaction{
		UserID:      userID,
		MemberID:    input.MemberID,
		AccountID:   input.AccountID,
		CategoryID:  input.CategoryID,
		Amount:      input.Amount,
		Description: input.Description,
		Type:        input.Type,
		Frequency:   input.Frequency,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		IsActive:    input.IsActive,
	}

	if err := db.Create(&recurringTransaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create recurring transaction")
		return
	}

	utils.RespondWithJSON(c, http.StatusCreated, recurringTransaction)
}

func GetRecurringTransactionByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid recurring transaction ID")
		return
	}

	db := database.GetDB()
	var recurringTransaction models.RecurringTransaction
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&recurringTransaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Recurring transaction not found")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, recurringTransaction)
}

func UpdateRecurringTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid recurring transaction ID")
		return
	}

	var input struct {
		AccountID   uint    `json:"account_id"`
		CategoryID  uint    `json:"category_id"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		Frequency   string  `json:"frequency"`
		EndDate     string  `json:"end_date"`
		IsActive    bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	db := database.GetDB()
	var recurringTransaction models.RecurringTransaction
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&recurringTransaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Recurring transaction not found")
		return
	}

	// Update fields if provided
	if input.AccountID != 0 {
		// Verify account belongs to member
		var account models.Account
		if err := db.Where("member_id = ? AND id = ?", recurringTransaction.MemberID, input.AccountID).First(&account).Error; err != nil {
			utils.RespondWithError(c, http.StatusNotFound, "Account not found")
			return
		}
		recurringTransaction.AccountID = input.AccountID
	}
	if input.CategoryID != 0 {
		// Verify category belongs to user
		var category models.Category
		if err := db.Where("user_id = ? AND id = ?", userID, input.CategoryID).First(&category).Error; err != nil {
			utils.RespondWithError(c, http.StatusNotFound, "Category not found")
			return
		}
		recurringTransaction.CategoryID = input.CategoryID
	}
	if input.Amount != 0 {
		recurringTransaction.Amount = input.Amount
	}
	if input.Description != "" {
		recurringTransaction.Description = input.Description
	}
	if input.Frequency != "" {
		recurringTransaction.Frequency = input.Frequency
	}
	if input.EndDate != "" {
		recurringTransaction.EndDate = input.EndDate
	}
	recurringTransaction.IsActive = input.IsActive

	if err := db.Save(&recurringTransaction).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update recurring transaction")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, recurringTransaction)
}

func DeleteRecurringTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid recurring transaction ID")
		return
	}

	db := database.GetDB()
	if err := db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.RecurringTransaction{}).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete recurring transaction")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{"message": "Recurring transaction deleted successfully"})
}
