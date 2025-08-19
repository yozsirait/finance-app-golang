package routes

import (
	"finance-app/controllers"
	"finance-app/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// =========================
	// Public routes
	// =========================
	r.POST("/register", controllers.Register) // @Tags Auth
	r.POST("/login", controllers.Login)       // @Tags Auth

	// =========================
	// Authenticated routes
	// =========================
	auth := r.Group("/api")
	auth.Use(utils.JWTAuthMiddleware())
	{
		// ========== User ==========
		user := auth.Group("/user")
		{
			user.GET("", controllers.GetCurrentUser) // @Tags User
			user.PUT("", controllers.UpdateUser)     // @Tags User
			user.DELETE("", controllers.DeleteUser)  // @Tags User
		}

		// ========== Members ==========
		members := auth.Group("/members")
		{
			members.GET("", controllers.GetMembers)          // @Tags Member
			members.POST("", controllers.CreateMember)       // @Tags Member
			members.GET("/:id", controllers.GetMemberByID)   // @Tags Member
			members.PUT("/:id", controllers.UpdateMember)    // @Tags Member
			members.DELETE("/:id", controllers.DeleteMember) // @Tags Member
		}

		// ========== Accounts ==========
		accounts := auth.Group("/accounts")
		{
			accounts.GET("", controllers.GetAccounts)           // @Tags Account
			accounts.POST("", controllers.CreateAccount)        // @Tags Account
			accounts.GET("/types", controllers.GetAccountTypes) // @Tags Account
			accounts.GET("/:id", controllers.GetAccountByID)    // @Tags Account
			accounts.PUT("/:id", controllers.UpdateAccount)     // @Tags Account
			accounts.DELETE("/:id", controllers.DeleteAccount)  // @Tags Account
		}

		// ========== Categories ==========
		categories := auth.Group("/categories")
		{
			categories.GET("", controllers.GetCategories)         // @Tags Category
			categories.POST("", controllers.CreateCategory)       // @Tags Category
			categories.GET("/:id", controllers.GetCategoryByID)   // @Tags Category
			categories.PUT("/:id", controllers.UpdateCategory)    // @Tags Category
			categories.DELETE("/:id", controllers.DeleteCategory) // @Tags Category
		}

		// ========== Budgets ==========
		budgets := auth.Group("/budgets")
		{
			budgets.GET("", controllers.GetBudgets)          // @Tags Budget
			budgets.POST("", controllers.CreateBudget)       // @Tags Budget
			budgets.GET("/:id", controllers.GetBudgetByID)   // @Tags Budget
			budgets.PUT("/:id", controllers.UpdateBudget)    // @Tags Budget
			budgets.DELETE("/:id", controllers.DeleteBudget) // @Tags Budget
		}

		// ========== Transactions ==========
		transactions := auth.Group("/transactions")
		{
			transactions.GET("", controllers.GetTransactions)          // @Tags Transaction
			transactions.GET("/:id", controllers.GetTransactionByID)   // @Tags Transaction
			transactions.POST("", controllers.CreateTransaction)       // @Tags Transaction
			transactions.PUT("/:id", controllers.UpdateTransaction)    // @Tags Transaction
			transactions.DELETE("/:id", controllers.DeleteTransaction) // @Tags Transaction
		}

		// ========== Recurring Transactions ==========
		recurring := auth.Group("/recurring-transactions")
		{
			recurring.GET("", controllers.GetRecurringTransactions)          // @Tags RecurringTransaction
			recurring.POST("", controllers.CreateRecurringTransaction)       // @Tags RecurringTransaction
			recurring.GET("/:id", controllers.GetRecurringTransactionByID)   // @Tags RecurringTransaction
			recurring.PUT("/:id", controllers.UpdateRecurringTransaction)    // @Tags RecurringTransaction
			recurring.DELETE("/:id", controllers.DeleteRecurringTransaction) // @Tags RecurringTransaction
		}

		// ========== Transfers ==========
		transfers := auth.Group("/transfers")
		{
			transfers.GET("", controllers.GetTransfers)          // @Tags Transfer
			transfers.POST("", controllers.CreateTransfer)       // @Tags Transfer
			transfers.GET("/:id", controllers.GetTransferByID)   // @Tags Transfer
			transfers.DELETE("/:id", controllers.DeleteTransfer) // @Tags Transfer
		}

		// ========== Saving Targets ==========
		saving := auth.Group("/saving-targets")
		{
			saving.GET("", controllers.GetSavingTargets)          // @Tags SavingTarget
			saving.POST("", controllers.CreateSavingTarget)       // @Tags SavingTarget
			saving.GET("/:id", controllers.GetSavingTargetByID)   // @Tags SavingTarget
			saving.PUT("/:id", controllers.UpdateSavingTarget)    // @Tags SavingTarget
			saving.DELETE("/:id", controllers.DeleteSavingTarget) // @Tags SavingTarget
		}

		// ========== Dashboard ==========
		auth.GET("/dashboard", controllers.GetDashboard) // @Tags Dashboard

		// ========== Reports ==========
		reports := auth.Group("/reports")
		{
			reports.GET("/transactions", controllers.GetReportTransactions)                // @Tags Report
			reports.GET("/summary", controllers.GetReportSummary)                          // @Tags Report
			reports.GET("/budget", controllers.GetBudgetReport)                            // @Tags Report
			reports.GET("/saving", controllers.GetSavingReport)                            // @Tags Report
			reports.GET("/members-comparison", controllers.GetMembersComparisonReport)     // @Tags Report
			reports.GET("/members-comparison-chart", controllers.GetMemberComparisonChart) // @Tags Report

			// Report export
			reports.GET("/export/csv", controllers.ExportTransactionsCSV) // @Tags Report
			reports.GET("/export/pdf", controllers.ExportTransactionsPDF) // @Tags Report
		}
	}

	return r
}
