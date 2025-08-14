package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetSavingTargets(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	memberID := c.Query("member_id")
	accountID := c.Query("account_id")
	isCompleted := c.Query("is_completed")

	db := database.GetDB()
	query := db.Joins("JOIN members ON members.id = saving_targets.member_id").
		Where("members.user_id = ?", userID)

	if memberID != "" {
		query = query.Where("saving_targets.member_id = ?", memberID)
	}
	if accountID != "" {
		query = query.Where("saving_targets.account_id = ?", accountID)
	}
	if isCompleted != "" {
		if isCompleted == "true" {
			query = query.Where("saving_targets.current_amount >= saving_targets.target_amount")
		} else {
			query = query.Where("saving_targets.current_amount < saving_targets.target_amount")
		}
	}

	var savingTargets []models.SavingTarget
	if err := query.Find(&savingTargets).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch saving targets")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, savingTargets)
}

func CreateSavingTarget(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		MemberID     uint    `json:"member_id" binding:"required"`
		AccountID    uint    `json:"account_id" binding:"required"`
		Name         string  `json:"name" binding:"required"`
		TargetAmount float64 `json:"target_amount" binding:"required"`
		TargetDate   string  `json:"target_date" binding:"required"`
		Description  string  `json:"description"`
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

	savingTarget := models.SavingTarget{
		UserID:        userID,
		MemberID:      input.MemberID,
		AccountID:     input.AccountID,
		Name:          input.Name,
		TargetAmount:  input.TargetAmount,
		CurrentAmount: 0,
		TargetDate:    input.TargetDate,
		Description:   input.Description,
	}

	if err := db.Create(&savingTarget).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create saving target")
		return
	}

	utils.RespondWithJSON(c, http.StatusCreated, savingTarget)
}

func GetSavingTargetByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid saving target ID")
		return
	}

	db := database.GetDB()
	var savingTarget models.SavingTarget
	if err := db.Joins("JOIN members ON members.id = saving_targets.member_id").
		Where("members.user_id = ? AND saving_targets.id = ?", userID, id).
		First(&savingTarget).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Saving target not found")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, savingTarget)
}

func UpdateSavingTarget(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid saving target ID")
		return
	}

	var input struct {
		Name          string  `json:"name"`
		TargetAmount  float64 `json:"target_amount"`
		CurrentAmount float64 `json:"current_amount"`
		TargetDate    string  `json:"target_date"`
		Description   string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	db := database.GetDB()
	var savingTarget models.SavingTarget
	if err := db.Joins("JOIN members ON members.id = saving_targets.member_id").
		Where("members.user_id = ? AND saving_targets.id = ?", userID, id).
		First(&savingTarget).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Saving target not found")
		return
	}

	// Update fields if provided
	if input.Name != "" {
		savingTarget.Name = input.Name
	}
	if input.TargetAmount != 0 {
		savingTarget.TargetAmount = input.TargetAmount
	}
	if input.CurrentAmount != 0 {
		savingTarget.CurrentAmount = input.CurrentAmount
	}
	if input.TargetDate != "" {
		savingTarget.TargetDate = input.TargetDate
	}
	if input.Description != "" {
		savingTarget.Description = input.Description
	}

	if err := db.Save(&savingTarget).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update saving target")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, savingTarget)
}

func DeleteSavingTarget(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid saving target ID")
		return
	}

	db := database.GetDB()
	if err := db.Joins("JOIN members ON members.id = saving_targets.member_id").
		Where("members.user_id = ? AND saving_targets.id = ?", userID, id).
		Delete(&models.SavingTarget{}).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete saving target")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{"message": "Saving target deleted successfully"})
}
