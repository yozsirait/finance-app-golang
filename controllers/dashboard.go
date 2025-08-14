package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetDashboardSummary(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	db := database.GetDB()

	// Get total balance across all accounts
	var totalBalance struct {
		Total float64
	}
	if err := db.Table("accounts").
		Joins("JOIN members ON members.id = accounts.member_id").
		Where("members.user_id = ?", userID).
		Select("SUM(balance) as total").
		Scan(&totalBalance).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to calculate total balance")
		return
	}

	// Get total income for current month
	currentMonth := time.Now().Format("2006-01")
	var totalIncome struct {
		Total float64
	}
	if err := db.Table("transactions").
		Where("user_id = ? AND type = 'income' AND date LIKE ?", userID, currentMonth+"%").
		Select("SUM(amount) as total").
		Scan(&totalIncome).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to calculate total income")
		return
	}

	// Get total expenses for current month
	var totalExpense struct {
		Total float64
	}
	if err := db.Table("transactions").
		Where("user_id = ? AND type = 'expense' AND date LIKE ?", userID, currentMonth+"%").
		Select("SUM(amount) as total").
		Scan(&totalExpense).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to calculate total expenses")
		return
	}

	// Get saving targets progress
	var savingTargetsProgress struct {
		TotalTargets  int
		Completed     int
		TotalAmount   float64
		CurrentAmount float64
	}
	if err := db.Table("saving_targets").
		Joins("JOIN members ON members.id = saving_targets.member_id").
		Where("members.user_id = ?", userID).
		Select("COUNT(*) as total_targets, " +
			"SUM(CASE WHEN current_amount >= target_amount THEN 1 ELSE 0 END) as completed, " +
			"SUM(target_amount) as total_amount, " +
			"SUM(current_amount) as current_amount").
		Scan(&savingTargetsProgress).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to calculate saving targets progress")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{
		"total_balance":         totalBalance.Total,
		"current_month_income":  totalIncome.Total,
		"current_month_expense": totalExpense.Total,
		"saving_targets":        savingTargetsProgress,
	})
}

func GetDashboardTransactions(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := c.DefaultQuery("limit", "10")

	db := database.GetDB()
	var transactions []models.Transaction

	// Convert limit string to int
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid limit parameter")
		return
	}

	if err := db.Where("user_id = ?", userID).
		Order("date DESC").
		Limit(limitInt). // Use the converted integer value
		Find(&transactions).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch recent transactions")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, transactions)
}

func GetDashboardCategories(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	transactionType := c.DefaultQuery("type", "expense")
	period := c.DefaultQuery("period", "month")

	currentPeriod := time.Now().Format("2006-01")
	if period == "year" {
		currentPeriod = time.Now().Format("2006")
	}

	db := database.GetDB()
	var categories []struct {
		CategoryID   uint
		CategoryName string
		Total        float64
	}
	if err := db.Table("transactions").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Where("transactions.user_id = ? AND transactions.type = ?", userID, transactionType).
		Where("transactions.date LIKE ?", currentPeriod+"%").
		Group("transactions.category_id").
		Select("transactions.category_id as category_id, " +
			"categories.name as category_name, " +
			"SUM(transactions.amount) as total").
		Scan(&categories).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch category breakdown")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, categories)
}

func GetDashboardBudgets(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	currentPeriod := time.Now().Format("2006-01")

	db := database.GetDB()
	var budgets []struct {
		BudgetCategoryID uint
		CategoryID       uint
		CategoryName     string
		BudgetAmount     float64
		SpentAmount      float64
		Period           string
	}
	if err := db.Table("budget_categories").
		Joins("JOIN categories ON categories.id = budget_categories.category_id").
		Joins("LEFT JOIN transactions ON transactions.category_id = categories.id AND transactions.type = 'expense' AND transactions.date LIKE ?", currentPeriod+"%").
		Where("budget_categories.user_id = ? AND budget_categories.period = 'monthly'", userID).
		Group("budget_categories.id").
		Select("budget_categories.id as budget_category_id, " +
			"categories.id as category_id, " +
			"categories.name as category_name, " +
			"budget_categories.amount as budget_amount, " +
			"COALESCE(SUM(transactions.amount), 0) as spent_amount, " +
			"budget_categories.period as period").
		Scan(&budgets).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch budget status")
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, budgets)
}
