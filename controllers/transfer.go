package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetTransfers(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	memberID := c.Query("member_id")
	accountID := c.Query("account_id")

	db := database.GetDB()
	query := db.Joins("JOIN members ON members.id = transfers.member_id").
		Where("members.user_id = ?", userID)

	if memberID != "" {
		query = query.Where("transfers.member_id = ?", memberID)
	}
	if accountID != "" {
		query = query.Where("transfers.from_account_id = ? OR transfers.to_account_id = ?", accountID, accountID)
	}

	var transfers []models.Transfer
	if err := query.Find(&transfers).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch transfers")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, transfers)
}

func CreateTransfer(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		MemberID      uint    `json:"member_id" binding:"required"`
		FromAccountID uint    `json:"from_account_id" binding:"required"`
		ToAccountID   uint    `json:"to_account_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
		Date          string  `json:"date" binding:"required"`
		Description   string  `json:"description"`
		Fee           float64 `json:"fee"`
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

	// Verify from account belongs to member
	var fromAccount models.Account
	if err := db.Where("member_id = ? AND id = ?", input.MemberID, input.FromAccountID).First(&fromAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "From account not found")
		return
	}

	// Verify to account belongs to member
	var toAccount models.Account
	if err := db.Where("member_id = ? AND id = ?", input.MemberID, input.ToAccountID).First(&toAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "To account not found")
		return
	}

	// Check if from account has enough balance
	if fromAccount.Balance < input.Amount+input.Fee {
		utils.RespondWithError(c, http.StatusBadRequest, "Insufficient balance in from account")
		return
	}

	// Create transfer record
	transfer := models.Transfer{
		UserID:        userID,
		MemberID:      input.MemberID,
		FromAccountID: input.FromAccountID,
		ToAccountID:   input.ToAccountID,
		Amount:        input.Amount,
		Date:          input.Date,
		Description:   input.Description,
		Fee:           input.Fee,
	}

	if err := db.Create(&transfer).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create transfer")
		return
	}

	// Update account balances
	fromAccount.Balance -= (input.Amount + input.Fee)
	toAccount.Balance += input.Amount

	if err := db.Save(&fromAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update from account balance")
		return
	}

	if err := db.Save(&toAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update to account balance")
		return
	}

	utils.RespondWithJSON(c, http.StatusCreated, transfer)
}

func GetTransferByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transfer ID")
		return
	}

	db := database.GetDB()
	var transfer models.Transfer
	if err := db.Joins("JOIN members ON members.id = transfers.member_id").
		Where("members.user_id = ? AND transfers.id = ?", userID, id).
		First(&transfer).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Transfer not found")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, transfer)
}

func DeleteTransfer(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transfer ID")
		return
	}

	db := database.GetDB()
	var transfer models.Transfer
	if err := db.Joins("JOIN members ON members.id = transfers.member_id").
		Where("members.user_id = ? AND transfers.id = ?", userID, id).
		First(&transfer).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Transfer not found")
		return
	}

	// Verify from account belongs to member
	var fromAccount models.Account
	if err := db.Where("member_id = ? AND id = ?", transfer.MemberID, transfer.FromAccountID).First(&fromAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "From account not found")
		return
	}

	// Verify to account belongs to member
	var toAccount models.Account
	if err := db.Where("member_id = ? AND id = ?", transfer.MemberID, transfer.ToAccountID).First(&toAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "To account not found")
		return
	}

	// Revert the transfer
	fromAccount.Balance += (transfer.Amount + transfer.Fee)
	toAccount.Balance -= transfer.Amount

	if err := db.Save(&fromAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to revert from account balance")
		return
	}

	if err := db.Save(&toAccount).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to revert to account balance")
		return
	}

	if err := db.Delete(&transfer).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete transfer")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{"message": "Transfer deleted successfully"})
}
