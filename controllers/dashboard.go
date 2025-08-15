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
		SELECT COALESCE(SUM(a.balance), 0)
		FROM accounts a
		JOIN members m ON m.id = a.member_id
		WHERE m.user_id = ?
	`, userID).Scan(&totalBalance)

	// 2. Income & Expense bulan ini
	var incomeThis, expenseThis float64
	db.Raw(`
		SELECT COALESCE(SUM(amount),0)
		FROM transactions
		WHERE user_id = ? AND type = 'income'
		AND YEAR(date) = ? AND MONTH(date) = ?
	`, userID, currentYear, currentMonth).Scan(&incomeThis)

	db.Raw(`
		SELECT COALESCE(SUM(amount),0)
		FROM transactions
		WHERE user_id = ? AND type = 'expense'
		AND YEAR(date) = ? AND MONTH(date) = ?
	`, userID, currentYear, currentMonth).Scan(&expenseThis)

	// 3. Income & Expense bulan lalu (untuk perbandingan)
	var incomeLast, expenseLast float64
	db.Raw(`
		SELECT COALESCE(SUM(amount),0)
		FROM transactions
		WHERE user_id = ? AND type = 'income'
		AND YEAR(date) = ? AND MONTH(date) = ?
	`, userID, lastYear, lastMonthNum).Scan(&incomeLast)

	db.Raw(`
		SELECT COALESCE(SUM(amount),0)
		FROM transactions
		WHERE user_id = ? AND type = 'expense'
		AND YEAR(date) = ? AND MONTH(date) = ?
	`, userID, lastYear, lastMonthNum).Scan(&expenseLast)

	// 4. Top 3 kategori pengeluaran bulan ini
	var topCategories []struct {
		Name  string  `json:"name"`
		Total float64 `json:"total"`
	}
	db.Raw(`
		SELECT c.name, COALESCE(SUM(t.amount),0) as total
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = ? AND t.type = 'expense'
		AND YEAR(t.date) = ? AND MONTH(t.date) = ?
		GROUP BY c.name
		ORDER BY total DESC
		LIMIT 3
	`, userID, currentYear, currentMonth).Scan(&topCategories)

	// 5. Top 5 transaksi terbesar bulan ini
	var topTransactions []models.Transaction
	db.Where("user_id = ? AND YEAR(date) = ? AND MONTH(date) = ?", userID, currentYear, currentMonth).
		Order("amount DESC").Limit(5).Find(&topTransactions)

	utils.RespondWithSuccess(c, gin.H{
		"total_balance":      totalBalance,
		"income_this_month":  incomeThis,
		"expense_this_month": expenseThis,
		"income_last_month":  incomeLast,
		"expense_last_month": expenseLast,
		"income_change":      incomeThis - incomeLast,
		"expense_change":     expenseThis - expenseLast,
		"net_savings":        incomeThis - expenseThis,
		"top_categories":     topCategories,
		"top_transactions":   topTransactions,
		"current_month":      currentMonth,
		"current_year":       currentYear,
	})
}
