package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Register godoc
// @Summary Register new user
// @Description Membuat user baru
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.AuthRegisterRequest true "Register user input"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /register [post]
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

	utils.RespondWithSuccess(c, gin.H{"message": "User registered successfully"})
}

// Login godoc
// @Summary Login user
// @Description Login dengan email & password, mendapatkan token JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.AuthLoginRequest true "Login user input"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	db := database.GetDB()
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.RespondWithSuccess(c, gin.H{"token": token})
}
