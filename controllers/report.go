package controllers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ===============================
// Helper: Format Amount dengan pemisah ribuan
// ===============================
func formatAmount(val interface{}) string {
	if val == nil {
		return "0,00"
	}
	f, err := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
	if err != nil {
		return fmt.Sprintf("%v", val)
	}
	// formatter dengan pemisah ribuan (Indonesia)
	p := message.NewPrinter(language.Indonesian)
	return p.Sprintf("%.2f", f) // contoh: 5.000.000,00
}

// ===============================
// 1. Laporan Transaksi (Detail + Filter)
// ===============================
func GetReportTransactions(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	categoryID := c.Query("category_id")
	accountID := c.Query("account_id")
	tType := c.Query("type")

	db := database.GetDB()
	query := db.Table("transactions").
		Select("transactions.*, categories.name as category_name, accounts.name as account_name").
		Joins("JOIN members ON members.id = transactions.member_id").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Joins("JOIN accounts ON accounts.id = transactions.account_id").
		Where("members.user_id = ?", userID).
		Where("transactions.deleted_at IS NULL") // exclude deleted

	if startDate != "" && endDate != "" {
		query = query.Where("transactions.date BETWEEN ? AND ?", startDate, endDate)
	}
	if categoryID != "" {
		query = query.Where("transactions.category_id = ?", categoryID)
	}
	if accountID != "" {
		query = query.Where("transactions.account_id = ?", accountID)
	}
	if tType != "" {
		query = query.Where("transactions.type = ?", tType)
	}

	var results []map[string]interface{}
	if err := query.Find(&results).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch transactions report")
		return
	}

	utils.RespondWithSuccess(c, results)
}

// ===============================
// 2. Summary Income vs Expense
// ===============================
func GetReportSummary(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	startDate, _ := time.Parse("2006-01-02", month+"-01")
	endDate := startDate.AddDate(0, 1, -1)

	db := database.GetDB()
	var income float64
	var expense float64

	db.Table("transactions").
		Joins("JOIN members ON members.id = transactions.member_id").
		Where("members.user_id = ? AND transactions.type = 'income' AND transactions.date BETWEEN ? AND ? AND transactions.deleted_at IS NULL", userID, startDate, endDate).
		Select("COALESCE(SUM(transactions.amount),0)").Scan(&income)

	db.Table("transactions").
		Joins("JOIN members ON members.id = transactions.member_id").
		Where("members.user_id = ? AND transactions.type = 'expense' AND transactions.date BETWEEN ? AND ? AND transactions.deleted_at IS NULL", userID, startDate, endDate).
		Select("COALESCE(SUM(transactions.amount),0)").Scan(&expense)

	utils.RespondWithSuccess(c, gin.H{
		"month":   month,
		"income":  income,
		"expense": expense,
		"balance": income - expense,
	})
}

// ===============================
// 3. Budget vs Realisasi
// ===============================
func GetBudgetReport(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	period := c.Query("period") // monthly, weekly, yearly
	if period == "" {
		period = "monthly"
	}

	var startDate, endDate time.Time
	now := time.Now()

	switch period {
	case "weekly":
		year, week := now.ISOWeek()
		startDate = time.Date(year, 0, 0, 0, 0, 0, 0, time.UTC)
		for {
			y, w := startDate.ISOWeek()
			if y == year && w == week {
				break
			}
			startDate = startDate.AddDate(0, 0, 1)
		}
		endDate = startDate.AddDate(0, 0, 6)
	case "yearly":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(1, 0, -1)
	default: // monthly
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, -1)
	}

	db := database.GetDB()
	query := db.Table("budget_categories").
		Select(`budget_categories.id as budget_id,
                budget_categories.category_id,
                categories.name as category_name,
                budget_categories.amount as budget_amount,
                COALESCE(SUM(transactions.amount), 0) as actual_amount`).
		Joins("JOIN categories ON categories.id = budget_categories.category_id").
		Joins("LEFT JOIN transactions ON transactions.category_id = budget_categories.category_id AND transactions.type = 'expense' AND transactions.date BETWEEN ? AND ? AND transactions.deleted_at IS NULL", startDate, endDate).
		Where("budget_categories.user_id = ? AND budget_categories.period = ?", userID, period).
		Group("budget_categories.id, budget_categories.category_id, categories.name, budget_categories.amount")

	var reports []struct {
		BudgetID     uint
		CategoryID   uint
		CategoryName string
		BudgetAmount float64
		ActualAmount float64
	}

	if err := query.Find(&reports).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch budget report")
		return
	}

	type Result struct {
		BudgetID     uint    `json:"budget_id"`
		CategoryID   uint    `json:"category_id"`
		CategoryName string  `json:"category_name"`
		BudgetAmount float64 `json:"budget_amount"`
		ActualAmount float64 `json:"actual_amount"`
		Status       string  `json:"status"`
	}

	var results []Result
	for _, r := range reports {
		status := "on budget"
		if r.ActualAmount < r.BudgetAmount {
			status = "under budget"
		} else if r.ActualAmount > r.BudgetAmount {
			status = "over budget"
		}
		results = append(results, Result{
			BudgetID:     r.BudgetID,
			CategoryID:   r.CategoryID,
			CategoryName: r.CategoryName,
			BudgetAmount: r.BudgetAmount,
			ActualAmount: r.ActualAmount,
			Status:       status,
		})
	}

	utils.RespondWithSuccess(c, gin.H{
		"period":  period,
		"reports": results,
	})
}

// ===============================
// 4. Saving Target
// ===============================
func GetSavingReport(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	db := database.GetDB()
	var savings []models.SavingTarget
	if err := db.Where("user_id = ?", userID).Find(&savings).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch saving targets")
		return
	}

	type Result struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Target      float64 `json:"target"`
		Current     float64 `json:"current"`
		ProgressPct float64 `json:"progress_pct"`
		Status      string  `json:"status"`
	}

	var results []Result
	for _, s := range savings {
		progress := (s.CurrentAmount / s.TargetAmount) * 100
		status := "in progress"
		if s.CurrentAmount >= s.TargetAmount {
			status = "achieved"
		}
		results = append(results, Result{
			ID:          s.ID,
			Name:        s.Name,
			Target:      s.TargetAmount,
			Current:     s.CurrentAmount,
			ProgressPct: progress,
			Status:      status,
		})
	}

	utils.RespondWithSuccess(c, results)
}

// ===============================
// Export CSV
// ===============================
func ExportTransactionsCSV(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	db := database.GetDB()
	query := db.Table("transactions").
		Select("transactions.id, transactions.date, transactions.type, transactions.amount, categories.name as category, accounts.name as account").
		Joins("JOIN members ON members.id = transactions.member_id").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Joins("JOIN accounts ON accounts.id = transactions.account_id").
		Where("members.user_id = ?", userID).
		Where("transactions.deleted_at IS NULL")

	if startDate != "" && endDate != "" {
		query = query.Where("transactions.date BETWEEN ? AND ?", startDate, endDate)
	}

	var results []map[string]interface{}
	if err := query.Find(&results).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch data")
		return
	}

	var b bytes.Buffer
	writer := csv.NewWriter(&b)

	writer.Write([]string{"ID", "Date", "Type", "Amount", "Category", "Account"})
	for _, r := range results {
		writer.Write([]string{
			fmt.Sprintf("%v", r["id"]),
			fmt.Sprintf("%v", r["date"]),
			fmt.Sprintf("%v", r["type"]),
			formatAmount(r["amount"]), // pakai formatter
			fmt.Sprintf("%v", r["category"]),
			fmt.Sprintf("%v", r["account"]),
		})
	}
	writer.Flush()

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=transactions_report.csv")
	c.Data(http.StatusOK, "text/csv", b.Bytes())
}

// ===============================
// Export PDF
// ===============================
func ExportTransactionsPDF(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	db := database.GetDB()
	query := db.Table("transactions").
		Select("transactions.id, transactions.date, transactions.type, transactions.amount, categories.name as category, accounts.name as account").
		Joins("JOIN members ON members.id = transactions.member_id").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Joins("JOIN accounts ON accounts.id = transactions.account_id").
		Where("members.user_id = ?", userID).
		Where("transactions.deleted_at IS NULL")

	if startDate != "" && endDate != "" {
		query = query.Where("transactions.date BETWEEN ? AND ?", startDate, endDate)
	}

	var results []map[string]interface{}
	if err := query.Find(&results).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch data")
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Transactions Report")
	pdf.Ln(12)

	// Table Header
	pdf.SetFont("Arial", "B", 10)
	headers := []string{"ID", "Date", "Type", "Amount", "Category", "Account"}
	for _, h := range headers {
		pdf.CellFormat(32, 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Table Rows
	pdf.SetFont("Arial", "", 9)
	for _, r := range results {
		pdf.CellFormat(32, 6, fmt.Sprintf("%v", r["id"]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(32, 6, fmt.Sprintf("%v", r["date"]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(32, 6, fmt.Sprintf("%v", r["type"]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(32, 6, formatAmount(r["amount"]), "1", 0, "R", false, 0, "")
		pdf.CellFormat(32, 6, fmt.Sprintf("%v", r["category"]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(32, 6, fmt.Sprintf("%v", r["account"]), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=transactions_report.pdf")
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// ===============================
// 5. Laporan Perbandingan Antar Member
// ===============================
func GetMemberComparisonChart(c *gin.Context) {
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

	type ChartSet struct {
		Label           string    `json:"label"`
		Data            []float64 `json:"data"`
		BackgroundColor string    `json:"backgroundColor"`
	}

	type ChartData struct {
		Labels   []string   `json:"labels"`
		Datasets []ChartSet `json:"datasets"`
	}

	var labels []string
	var incomeData []float64
	var expenseData []float64
	var balanceData []float64

	for _, m := range members {
		labels = append(labels, m.Name)

		var totalIncome, totalExpense float64
		db.Table("transactions").
			Where("member_id = ? AND type = 'income'", m.ID).
			Select("COALESCE(SUM(amount),0)").Scan(&totalIncome)
		db.Table("transactions").
			Where("member_id = ? AND type = 'expense'", m.ID).
			Select("COALESCE(SUM(amount),0)").Scan(&totalExpense)

		incomeData = append(incomeData, totalIncome)
		expenseData = append(expenseData, totalExpense)
		balanceData = append(balanceData, totalIncome-totalExpense)
	}

	chart := ChartData{
		Labels: labels,
		Datasets: []ChartSet{
			{Label: "Income", Data: incomeData, BackgroundColor: "#4caf50"},
			{Label: "Expense", Data: expenseData, BackgroundColor: "#f44336"},
			{Label: "Balance", Data: balanceData, BackgroundColor: "#2196f3"},
		},
	}

	utils.RespondWithSuccess(c, chart)
}

// ===============================
// 5. Laporan Perbandingan Antar Member
// ===============================
func GetMembersComparisonReport(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	db := database.GetDB()
	query := db.Table("members").
		Select(`
            members.id as member_id,
            members.name as member_name,
            COALESCE(SUM(CASE WHEN transactions.type = 'income' THEN transactions.amount ELSE 0 END), 0) as total_income,
            COALESCE(SUM(CASE WHEN transactions.type = 'expense' THEN transactions.amount ELSE 0 END), 0) as total_expense
        `).
		Joins("LEFT JOIN transactions ON transactions.member_id = members.id")

	if startDate != "" && endDate != "" {
		query = query.Where("transactions.date BETWEEN ? AND ?", startDate, endDate)
	}

	query = query.Where("members.user_id = ?", userID).
		Group("members.id, members.name").
		Order("members.name ASC")

	var results []struct {
		MemberID     uint    `json:"member_id"`
		MemberName   string  `json:"member_name"`
		TotalIncome  float64 `json:"total_income"`
		TotalExpense float64 `json:"total_expense"`
	}

	if err := query.Find(&results).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch members comparison report")
		return
	}

	// Hitung balance per member
	type Result struct {
		MemberID     uint    `json:"member_id"`
		MemberName   string  `json:"member_name"`
		TotalIncome  float64 `json:"total_income"`
		TotalExpense float64 `json:"total_expense"`
		Balance      float64 `json:"balance"`
	}

	var finalResults []Result
	for _, r := range results {
		finalResults = append(finalResults, Result{
			MemberID:     r.MemberID,
			MemberName:   r.MemberName,
			TotalIncome:  r.TotalIncome,
			TotalExpense: r.TotalExpense,
			Balance:      r.TotalIncome - r.TotalExpense,
		})
	}

	utils.RespondWithSuccess(c, gin.H{
		"start_date": startDate,
		"end_date":   endDate,
		"members":    finalResults,
	})
}
