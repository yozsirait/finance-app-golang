package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAccounts(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	memberID := c.Query("member_id")

	db := database.GetDB()
	query := db.Joins("JOIN members ON members.id = accounts.member_id").
		Where("members.user_id = ?", userID)

	if memberID != "" {
		query = query.Where("accounts.member_id = ?", memberID)
	}

	var accounts []models.Account
	if err := query.Find(&accounts).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch accounts")
		return
	}

	utils.RespondWithSuccess(c, accounts)
}

func CreateAccount(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		MemberID uint    `json:"member_id" binding:"required"`
		Name     string  `json:"name" binding:"required"`
		Type     string  `json:"type" binding:"required"`
		Balance  float64 `json:"balance"`
		Currency string  `json:"currency"`
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

	account := models.Account{
		MemberID: input.MemberID,
		Name:     input.Name,
		Type:     input.Type,
		Balance:  input.Balance,
		Currency: input.Currency,
	}

	if err := db.Create(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create account")
		return
	}

	utils.RespondWithSuccess(c, account)
}

func GetAccountByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	db := database.GetDB()
	var account models.Account
	if err := db.Joins("JOIN members ON members.id = accounts.member_id").
		Where("members.user_id = ? AND accounts.id = ?", userID, id).
		First(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	utils.RespondWithSuccess(c, account)
}

func UpdateAccount(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	var input struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Currency string `json:"currency"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	db := database.GetDB()
	var account models.Account
	if err := db.Joins("JOIN members ON members.id = accounts.member_id").
		Where("members.user_id = ? AND accounts.id = ?", userID, id).
		First(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	// Update fields if provided
	if input.Name != "" {
		account.Name = input.Name
	}
	if input.Type != "" {
		account.Type = input.Type
	}
	if input.Currency != "" {
		account.Currency = input.Currency
	}

	if err := db.Save(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update account")
		return
	}

	utils.RespondWithSuccess(c, account)
}

func DeleteAccount(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	db := database.GetDB()
	var account models.Account
	if err := db.Joins("JOIN members ON members.id = accounts.member_id").
		Where("members.user_id = ? AND accounts.id = ?", userID, id).
		First(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	if err := db.Delete(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	utils.RespondWithSuccess(c, gin.H{"message": "Account deleted successfully"})
}
