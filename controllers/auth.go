package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	db := database.GetDB()
	if err := db.Create(&user).Error; err != nil {
		utils.RespondWithError(c, http.StatusConflict, "Username or email already exists")
		return
	}

	utils.RespondWithJSON(c, http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	db := database.GetDB()
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{"token": token})
}
