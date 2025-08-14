package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	categoryType := c.Query("type")

	db := database.GetDB()
	query := db.Where("user_id = ?", userID)

	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	var categories []models.Category
	if err := query.Find(&categories).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, categories)
}

func CreateCategory(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
		Type string `json:"type" binding:"required,oneof=income expense"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	category := models.Category{
		UserID: userID,
		Name:   input.Name,
		Type:   input.Type,
	}

	db := database.GetDB()
	if err := db.Create(&category).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create category")
		return
	}

	utils.RespondWithJSON(c, http.StatusCreated, category)
}

func GetCategoryByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	db := database.GetDB()
	var category models.Category
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&category).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Category not found")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, category)
}

func UpdateCategory(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var input struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	db := database.GetDB()
	var category models.Category
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&category).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Category not found")
		return
	}

	// Update fields if provided
	if input.Name != "" {
		category.Name = input.Name
	}
	if input.Type != "" {
		category.Type = input.Type
	}

	if err := db.Save(&category).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update category")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, category)
}

func DeleteCategory(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	db := database.GetDB()
	if err := db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.Category{}).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
