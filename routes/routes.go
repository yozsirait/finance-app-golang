package routes

import (
	"finance-app/controllers"
	"finance-app/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		// =========================
		// Public routes
		// =========================
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)
		api.POST("/logout", utils.JWTAuthMiddleware(), controllers.Logout)

		// =========================
		// Authenticated routes
		// =========================
		auth := api.Group("")
		auth.Use(utils.JWTAuthMiddleware())
		{
			// ========== User ==========
			user := auth.Group("/user")
			{
				user.GET("", controllers.GetCurrentUser)
				user.PUT("", controllers.UpdateUser)
				user.DELETE("", controllers.DeleteUser)
			}

			// ========== Members ==========
			members := auth.Group("/members")
			{
				members.GET("", controllers.GetMembers)
				members.POST("", controllers.CreateMember)
				members.GET("/:id", controllers.GetMemberByID)
				members.PUT("/:id", controllers.UpdateMember)
				members.DELETE("/:id", controllers.DeleteMember)
			}

			// ========== Accounts ==========
			accounts := auth.Group("/accounts")
			{
				accounts.GET("", controllers.GetAccounts)
				accounts.POST("", controllers.CreateAccount)
				accounts.GET("/types", controllers.GetAccountTypes)
				accounts.GET("/:id", controllers.GetAccountByID)
				accounts.PUT("/:id", controllers.UpdateAccount)
				accounts.DELETE("/:id", controllers.DeleteAccount)
			}

			// ========== Categories ==========
			categories := auth.Group("/categories")
			{
				categories.GET("", controllers.GetCategories)
				categories.POST("", controllers.CreateCategory)
				categories.GET("/:id", controllers.GetCategoryByID)
				categories.PUT("/:id", controllers.UpdateCategory)
				categories.DELETE("/:id", controllers.DeleteCategory)
			}

			// ========== Budgets ==========
			budgets := auth.Group("/budgets")
			{
				budgets.GET("", controllers.GetBudgets)
				budgets.POST("", controllers.CreateBudget)
				budgets.GET("/:id", controllers.GetBudgetByID)
				budgets.PUT("/:id", controllers.UpdateBudget)
				budgets.DELETE("/:id", controllers.DeleteBudget)
			}

			// ========== Transactions ==========
			transactions := auth.Group("/transactions")
			{
				transactions.GET("", controllers.GetTransactions)
				transactions.GET("/:id", controllers.GetTransactionByID)
				transactions.POST("", controllers.CreateTransaction)
				transactions.PUT("/:id", controllers.UpdateTransaction)
				transactions.DELETE("/:id", controllers.DeleteTransaction)
			}

			// ========== Recurring Transactions ==========
			recurring := auth.Group("/recurring-transactions")
			{
				recurring.GET("", controllers.GetRecurringTransactions)
				recurring.POST("", controllers.CreateRecurringTransaction)
				recurring.GET("/:id", controllers.GetRecurringTransactionByID)
				recurring.PUT("/:id", controllers.UpdateRecurringTransaction)
				recurring.DELETE("/:id", controllers.DeleteRecurringTransaction)
			}

			// ========== Transfers ==========
			transfers := auth.Group("/transfers")
			{
				transfers.GET("", controllers.GetTransfers)
				transfers.POST("", controllers.CreateTransfer)
				transfers.GET("/:id", controllers.GetTransferByID)
				transfers.DELETE("/:id", controllers.DeleteTransfer)
			}

			// ========== Saving Targets ==========
			saving := auth.Group("/saving-targets")
			{
				saving.GET("", controllers.GetSavingTargets)
				saving.POST("", controllers.CreateSavingTarget)
				saving.GET("/:id", controllers.GetSavingTargetByID)
				saving.PUT("/:id", controllers.UpdateSavingTarget)
				saving.DELETE("/:id", controllers.DeleteSavingTarget)
			}

			// ========== Dashboard ==========
			auth.GET("/dashboard", controllers.GetDashboard)

			// ========== Reports ==========
			reports := auth.Group("/reports")
			{
				reports.GET("/transactions", controllers.GetReportTransactions)
				reports.GET("/summary", controllers.GetReportSummary)
				reports.GET("/budget", controllers.GetBudgetReport)
				reports.GET("/saving", controllers.GetSavingReport)
				reports.GET("/members-comparison", controllers.GetMembersComparisonReport)
				reports.GET("/members-comparison-chart", controllers.GetMemberComparisonChart)

				// Report export
				reports.GET("/export/csv", controllers.ExportTransactionsCSV)
				reports.GET("/export/pdf", controllers.ExportTransactionsPDF)
			}
		}
	}

	return r
}
