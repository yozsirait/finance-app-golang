package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Helper untuk query join + filter user_id
func getAccountQuery(c *gin.Context) (*gorm.DB, uint, bool) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return nil, 0, false
	}

	db := database.GetDB()
	query := db.Joins("JOIN members ON members.id = accounts.member_id").
		Where("members.user_id = ?", userID)

	return query, userID, true
}

// GET /accounts?member_id=&type=
func GetAccounts(c *gin.Context) {
	query, _, ok := getAccountQuery(c)
	if !ok {
		return
	}

	if memberID := c.Query("member_id"); memberID != "" {
		query = query.Where("accounts.member_id = ?", memberID)
	}
	if accType := c.Query("type"); accType != "" {
		query = query.Where("accounts.type = ?", accType)
	}

	var accounts []models.Account
	if err := query.Find(&accounts).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch accounts")
		return
	}

	utils.RespondWithSuccess(c, accounts)
}

// GET /accounts/member/:member_id
func GetAccountByMemberID(c *gin.Context) {
	query, _, ok := getAccountQuery(c)
	if !ok {
		return
	}

	memberID := c.Param("member_id")
	var accounts []models.Account
	if err := query.Where("accounts.member_id = ?", memberID).Find(&accounts).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch accounts by member")
		return
	}

	utils.RespondWithSuccess(c, accounts)
}

// GET /accounts/type/:type
func GetAccountByType(c *gin.Context) {
	query, _, ok := getAccountQuery(c)
	if !ok {
		return
	}

	accType := c.Param("type")
	var accounts []models.Account
	if err := query.Where("accounts.type = ?", accType).Find(&accounts).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch accounts by type")
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

	// Verifikasi member milik user
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
	query, _, ok := getAccountQuery(c)
	if !ok {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	var account models.Account
	if err := query.Where("accounts.id = ?", id).First(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	utils.RespondWithSuccess(c, account)
}

func UpdateAccount(c *gin.Context) {
	query, _, ok := getAccountQuery(c)
	if !ok {
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

	var account models.Account
	if err := query.Where("accounts.id = ?", id).First(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	if input.Name != "" {
		account.Name = input.Name
	}
	if input.Type != "" {
		account.Type = input.Type
	}
	if input.Currency != "" {
		account.Currency = input.Currency
	}

	if err := database.GetDB().Save(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update account")
		return
	}

	utils.RespondWithSuccess(c, account)
}

func DeleteAccount(c *gin.Context) {
	query, _, ok := getAccountQuery(c)
	if !ok {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid account ID")
		return
	}

	var account models.Account
	if err := query.Where("accounts.id = ?", id).First(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Account not found")
		return
	}

	if err := database.GetDB().Delete(&account).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	utils.RespondWithSuccess(c, gin.H{"message": "Account deleted successfully"})
}
