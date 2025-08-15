package controllers

import (
	"net/http"
	"time"

	"finance-app/database"
	"finance-app/models"
	"finance-app/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GET /transfers?member_id=&account_id=
func GetTransfers(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}
	memberID := c.Query("member_id")
	accountID := c.Query("account_id")

	var transfers []models.Transfer
	db := database.DB.Preload("Member").Preload("FromAccount").Preload("ToAccount")

	query := db.Joins("JOIN members ON members.id = transfers.member_id").
		Where("members.user_id = ?", userID)

	if memberID != "" {
		query = query.Where("transfers.member_id = ?", memberID)
	}
	if accountID != "" {
		query = query.Where("transfers.from_account_id = ? OR transfers.to_account_id = ?", accountID, accountID)
	}

	if err := query.Find(&transfers).Error; err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get transfers")
		return
	}

	c.JSON(http.StatusOK, transfers)
}

// POST /transfers
func CreateTransfer(c *gin.Context) {
	userID, emsg := utils.GetUserIDFromToken(c)
	if emsg != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, emsg)
		return
	}

	var input models.Transfer
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	if input.FromAccountID == input.ToAccountID {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid transfer. Source and destination account cannot be the same")
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var member models.Member
		if err := tx.Where("id = ? AND user_id = ?", input.MemberID, userID).First(&member).Error; err != nil {
			return err
		}

		var fromAccount, toAccount models.Account
		if err := tx.Where("id = ? AND user_id = ?", input.FromAccountID, userID).First(&fromAccount).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ? AND user_id = ?", input.ToAccountID, userID).First(&toAccount).Error; err != nil {
			return err
		}

		if fromAccount.Balance < input.Amount {
			return gorm.ErrInvalidTransaction
		}

		fromAccount.Balance -= input.Amount
		toAccount.Balance += input.Amount

		if err := tx.Save(&fromAccount).Error; err != nil {
			return err
		}
		if err := tx.Save(&toAccount).Error; err != nil {
			return err
		}

		input.UserID = userID
		if input.Date == "" {
			input.Date = time.Now().Format("2006-01-02")
		}

		if err := tx.Create(&input).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to create transfer")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transfer created successfully"})
}

// GET /transfers/:id
func GetTransferByID(c *gin.Context) {
	userID, emsg := utils.GetUserIDFromToken(c)
	if emsg != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, emsg)
		return
	}
	id := c.Param("id")

	var transfer models.Transfer
	if err := database.DB.Preload("Member").Preload("FromAccount").Preload("ToAccount").
		Joins("JOIN members ON members.id = transfers.member_id").
		Where("transfers.id = ? AND members.user_id = ?", id, userID).
		First(&transfer).Error; err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Transfer not found")
		return
	}

	c.JSON(http.StatusOK, transfer)
}

// DELETE /transfers/:id
func DeleteTransfer(c *gin.Context) {
	userID, emsg := utils.GetUserIDFromToken(c)
	if emsg != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, emsg)
		return
	}
	id := c.Param("id")

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var transfer models.Transfer
		if err := tx.Preload("FromAccount").Preload("ToAccount").
			Joins("JOIN members ON members.id = transfers.member_id").
			Where("transfers.id = ? AND members.user_id = ?", id, userID).
			First(&transfer).Error; err != nil {
			return err
		}

		transfer.FromAccount.Balance += transfer.Amount
		transfer.ToAccount.Balance -= transfer.Amount

		if err := tx.Save(&transfer.FromAccount).Error; err != nil {
			return err
		}
		if err := tx.Save(&transfer.ToAccount).Error; err != nil {
			return err
		}

		if err := tx.Delete(&transfer).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to delete transfer")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer deleted successfully"})
}
