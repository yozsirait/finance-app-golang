package controllers

import (
	"net/http"

	"finance-app/database"
	"finance-app/services"
	"finance-app/utils"

	"github.com/gin-gonic/gin"
)

func GetTransactions(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	var q services.TransactionQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	service := services.NewTransactionService(database.GetDB())
	transactions, total, err := service.GetTransactions(userID, q)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithPaginatedData(c, transactions, total, q.Page, q.Limit)
}

func GetTransactionByID(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	service := services.NewTransactionService(database.GetDB())
	trx, err := service.GetTransactionByID(userID, c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Transaction not found")
		return
	}
	utils.RespondWithSuccess(c, trx)
}

func CreateTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, utils.FormatValidationError(err))
		return
	}

	service := services.NewTransactionService(database.GetDB())
	trx, err := service.Create(userID, req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithCreated(c, trx)
}

func UpdateTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, utils.FormatValidationError(err))
		return
	}

	service := services.NewTransactionService(database.GetDB())
	trx, err := service.Update(userID, c.Param("id"), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, trx)
}

func DeleteTransaction(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	service := services.NewTransactionService(database.GetDB())
	if err := service.Delete(userID, c.Param("id")); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithSuccess(c, gin.H{"message": "Transaction deleted successfully"})
}
