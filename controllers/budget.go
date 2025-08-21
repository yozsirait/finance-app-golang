package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetBudgets(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	period := c.Query("period")
	categoryID := c.Query("category_id")

	db := database.GetDB()
	query := db.Where("user_id = ?", userID)

	if period != "" {
		query = query.Where("period = ?", period)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	var budgets []models.BudgetCategory
	if err := query.Find(&budgets).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch budgets")
		return
	}

	utils.RespondWithSuccess(c, budgets)
}

func CreateBudget(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		CategoryID uint    `json:"category_id" binding:"required"`
		Amount     float64 `json:"amount" binding:"required"`
		Period     string  `json:"period" binding:"required,oneof=monthly weekly yearly"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	db := database.GetDB()
	var category models.Category
	if err := db.Where("user_id = ? AND id = ?", userID, input.CategoryID).First(&category).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Category not found")
		return
	}

	// ✅ Validasi hanya untuk kategori tipe expense
	if category.Type != "expense" {
		utils.RespondWithError(c, http.StatusBadRequest, "Budget hanya dapat dibuat untuk kategori pengeluaran (expense)")
		return
	}

	budget := models.BudgetCategory{
		UserID:     userID,
		CategoryID: input.CategoryID,
		Amount:     input.Amount,
		Period:     input.Period,
	}

	if err := db.Create(&budget).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create budget")
		return
	}

	utils.RespondWithSuccess(c, budget)
}

func UpdateBudget(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid budget ID")
		return
	}

	var input struct {
		Amount float64 `json:"amount"`
		Period string  `json:"period"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	db := database.GetDB()
	var budget models.BudgetCategory
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&budget).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Budget not found")
		return
	}

	// ✅ Pastikan category parent adalah expense
	var category models.Category
	if err := db.Where("id = ?", budget.CategoryID).First(&category).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Category not found")
		return
	}
	if category.Type != "expense" {
		utils.RespondWithError(c, http.StatusBadRequest, "Budget hanya dapat diupdate untuk kategori pengeluaran (expense)")
		return
	}

	// Update fields jika ada
	if input.Amount != 0 {
		budget.Amount = input.Amount
	}
	if input.Period != "" {
		budget.Period = input.Period
	}

	if err := db.Save(&budget).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update budget")
		return
	}

	utils.RespondWithSuccess(c, budget)
}

func GetBudgetByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid budget ID")
		return
	}

	db := database.GetDB()
	var budget models.BudgetCategory
	if err := db.Where("user_id = ? AND id = ?", userID, id).First(&budget).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Budget not found")
		return
	}

	utils.RespondWithSuccess(c, budget)
}

func DeleteBudget(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid budget ID")
		return
	}

	db := database.GetDB()
	if err := db.Where("user_id = ? AND id = ?", userID, id).Delete(&models.BudgetCategory{}).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete budget")
		return
	}

	utils.RespondWithSuccess(c, gin.H{"message": "Budget deleted successfully"})
}
