package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetCurrentUser(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var user models.User
	db := database.GetDB()
	if err := db.First(&user, userID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	// Don't return password hash
	user.Password = ""
	utils.RespondWithJSON(c, http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	db := database.GetDB()
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	// Update fields if provided
	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := db.Save(&user).Error; err != nil {
		utils.RespondWithError(c, http.StatusConflict, "Username or email already exists")
		return
	}

	user.Password = ""
	utils.RespondWithJSON(c, http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	db := database.GetDB()
	if err := db.Delete(&models.User{}, userID).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{"message": "User deleted successfully"})
}
