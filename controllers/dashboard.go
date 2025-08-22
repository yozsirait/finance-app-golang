package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetDashboard(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	db := database.GetDB()
	now := time.Now()
	currentYear, currentMonth := now.Year(), int(now.Month())
	lastMonth := now.AddDate(0, -1, 0)
	lastYear, lastMonthNum := lastMonth.Year(), int(lastMonth.Month())

	// 1. Total balance semua akun user
	var totalBalance float64
	db.Raw(`
		SELECT COALESCE(SUM(a.balance),0)
		FROM accounts a
		JOIN members m ON m.id = a.member_id
		WHERE m.user_id = ?
	`, userID).Scan(&totalBalance)

	// 2. Income & Expense bulan ini
	var currentSummary struct {
		Income  float64
		Expense float64
	}
	db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END),0) AS income,
			COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END),0) AS expense
		FROM transactions
		WHERE user_id = ? AND YEAR(date)=? AND MONTH(date)=?
	`, userID, currentYear, currentMonth).Scan(&currentSummary)

	// 3. Income & Expense bulan lalu
	var lastSummary struct {
		Income  float64
		Expense float64
	}
	db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END),0) AS income,
			COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END),0) AS expense
		FROM transactions
		WHERE user_id = ? AND YEAR(date)=? AND MONTH(date)=?
	`, userID, lastYear, lastMonthNum).Scan(&lastSummary)

	// Struct reuse untuk kategori chart
	type CategoryChart struct {
		Name  string  `json:"name"`
		Total float64 `json:"total"`
	}
	type BarChart struct {
		Category string  `json:"category"`
		Income   float64 `json:"income"`
		Expense  float64 `json:"expense"`
	}

	// 4. Pie chart (hanya kategori user ini)
	var pieCategories []CategoryChart
	db.Raw(`
		SELECT c.name, COALESCE(SUM(t.amount),0) AS total
		FROM categories c
		LEFT JOIN transactions t 
			ON t.category_id=c.id AND t.user_id=? AND t.type='expense'
			AND YEAR(t.date)=? AND MONTH(t.date)=?
		WHERE c.user_id = ? 
		GROUP BY c.name
		ORDER BY total DESC
	`, userID, currentYear, currentMonth, userID).Scan(&pieCategories)

	// 5. Top transaksi terbesar bulan ini
	var topTransactions []models.Transaction
	db.Where("user_id = ? AND YEAR(date) = ? AND MONTH(date) = ?", userID, currentYear, currentMonth).
		Order("amount DESC").Limit(5).Find(&topTransactions)

	// 6. Bar chart data per kategori (income & expense)
	var barChart []BarChart
	db.Raw(`
		SELECT c.name AS category,
			COALESCE(SUM(CASE WHEN t.type='income' THEN t.amount ELSE 0 END),0) AS income,
			COALESCE(SUM(CASE WHEN t.type='expense' THEN t.amount ELSE 0 END),0) AS expense
		FROM categories c
		LEFT JOIN transactions t 
			ON t.category_id=c.id AND t.user_id=? 
			AND YEAR(t.date)=? AND MONTH(t.date)=?
		WHERE c.user_id = ?
		GROUP BY c.name
		ORDER BY (COALESCE(SUM(t.amount),0)) DESC
	`, userID, currentYear, currentMonth, userID).Scan(&barChart)

	// 7. Top 3 kategori
	var top3Categories []CategoryChart
	db.Raw(`
		SELECT c.name, COALESCE(SUM(t.amount),0) AS total
		FROM categories c
		LEFT JOIN transactions t 
			ON t.category_id=c.id AND t.user_id=? AND t.type='expense'
			AND YEAR(t.date)=? AND MONTH(t.date)=?
		WHERE c.user_id = ?
		GROUP BY c.name
		ORDER BY total DESC
		LIMIT 3
	`, userID, currentYear, currentMonth, userID).Scan(&top3Categories)

	// Response
	utils.RespondWithSuccess(c, gin.H{
		"total_balance":      totalBalance,
		"income_this_month":  currentSummary.Income,
		"expense_this_month": currentSummary.Expense,
		"income_last_month":  lastSummary.Income,
		"expense_last_month": lastSummary.Expense,
		"income_change":      currentSummary.Income - lastSummary.Income,
		"expense_change":     currentSummary.Expense - lastSummary.Expense,
		"net_savings":        currentSummary.Income - currentSummary.Expense,
		"top_categories":     top3Categories,
		"pie_chart_data":     pieCategories,
		"top_transactions":   topTransactions,
		"bar_chart_data":     barChart,
		"current_month":      currentMonth,
		"current_year":       currentYear,
	})
}
