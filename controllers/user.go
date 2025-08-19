package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetCurrentUser godoc
// @Summary Get current user
// @Description Mendapatkan detail user yang sedang login (berdasarkan token JWT)
// @Tags User
// @Produce json
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user [get]
// @Security BearerAuth
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

	// Jangan return password hash
	user.Password = ""
	utils.RespondWithSuccess(c, user)
}

// UpdateUser godoc
// @Summary Update current user
// @Description Update data user yang sedang login
// @Tags User
// @Accept json
// @Produce json
// @Param body body models.UserUpdateRequest true "Update user input"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /user [put]
// @Security BearerAuth
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

	// Update fields jika ada input
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
	utils.RespondWithSuccess(c, user)
}

// DeleteUser godoc
// @Summary Delete current user
// @Description Hapus user yang sedang login (berdasarkan token JWT)
// @Tags User
// @Produce json
// @Success 200 {object} models.DeleteResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user [delete]
// @Security BearerAuth
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

	utils.RespondWithSuccess(c, gin.H{"message": "User deleted successfully"})
}
