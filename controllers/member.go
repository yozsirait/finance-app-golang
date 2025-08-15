package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetMembers(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	db := database.GetDB()
	var members []models.Member
	if err := db.Where("user_id = ?", userID).Find(&members).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch members")
		return
	}

	utils.RespondWithSuccess(c, members)
}

func CreateMember(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	member := models.Member{
		UserID: userID,
		Name:   input.Name,
	}

	db := database.GetDB()
	if err := db.Create(&member).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create member")
		return
	}

	utils.RespondWithSuccess(c, member)
}

func GetMemberByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid member ID")
		return
	}

	db := database.GetDB()
	var member models.Member
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&member).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Member not found")
		return
	}

	utils.RespondWithSuccess(c, member)
}

func UpdateMember(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid member ID")
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	db := database.GetDB()
	var member models.Member
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&member).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Member not found")
		return
	}

	member.Name = input.Name
	if err := db.Save(&member).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update member")
		return
	}

	utils.RespondWithSuccess(c, member)
}

func DeleteMember(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid member ID")
		return
	}

	db := database.GetDB()
	if err := db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.Member{}).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete member")
		return
	}

	utils.RespondWithSuccess(c, gin.H{"message": "Member deleted successfully"})
}
