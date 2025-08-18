package controllers

import (
	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetTransactionReport(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// ambil query params
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	memberID := c.Query("member_id")
	accountID := c.Query("account_id")
	categoryID := c.Query("category_id")
	txType := c.Query("type")

	db := database.GetDB()
	query := db.Model(&models.Transaction{}).
		Joins("JOIN members ON members.id = transactions.member_id").
		Where("members.user_id = ?", userID)

	// filter tanggal
	if startDate != "" && endDate != "" {
		start, _ := time.Parse("2006-01-02", startDate)
		end, _ := time.Parse("2006-01-02", endDate)
		query = query.Where("transactions.date BETWEEN ? AND ?", start, end)
	}

	// filter lain
	if memberID != "" {
		query = query.Where("transactions.member_id = ?", memberID)
	}
	if accountID != "" {
		query = query.Where("transactions.account_id = ?", accountID)
	}
	if categoryID != "" {
		query = query.Where("transactions.category_id = ?", categoryID)
	}
	if txType != "" {
		query = query.Where("transactions.type = ?", txType)
	}

	var transactions []models.Transaction
	if err := query.Order("transactions.date desc").Find(&transactions).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch report")
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"count":        len(transactions),
		"transactions": transactions,
	})
}
