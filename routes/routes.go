package routes

import (
	"finance-app/controllers"
	"finance-app/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Auth routes
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	// Authenticated routes
	auth := r.Group("/api")
	auth.Use(utils.JWTAuthMiddleware())
	{
		// User routes
		auth.GET("/user", controllers.GetCurrentUser)
		auth.PUT("/user", controllers.UpdateUser)
		auth.DELETE("/user", controllers.DeleteUser)

		// Member routes
		auth.GET("/members", controllers.GetMembers)
		auth.POST("/members", controllers.CreateMember)
		auth.GET("/members/:id", controllers.GetMemberByID)
		auth.PUT("/members/:id", controllers.UpdateMember)
		auth.DELETE("/members/:id", controllers.DeleteMember)

		// Account routes
		auth.GET("/accounts", controllers.GetAccounts) // Bisa filter member_id & type
		auth.POST("/accounts", controllers.CreateAccount)
		auth.GET("/accounts/:id", controllers.GetAccountByID)
		auth.PUT("/accounts/:id", controllers.UpdateAccount)
		auth.DELETE("/accounts/:id", controllers.DeleteAccount)

		// Category routes
		auth.GET("/categories", controllers.GetCategories) // ?type=income&search=food
		auth.POST("/categories", controllers.CreateCategory)
		auth.GET("/categories/:id", controllers.GetCategoryByID)
		auth.PUT("/categories/:id", controllers.UpdateCategory)
		auth.DELETE("/categories/:id", controllers.DeleteCategory)

		// Budget routes
		auth.GET("/budgets", controllers.GetBudgets)
		auth.POST("/budgets", controllers.CreateBudget)
		auth.GET("/budgets/:id", controllers.GetBudgetByID)
		auth.PUT("/budgets/:id", controllers.UpdateBudget)
		auth.DELETE("/budgets/:id", controllers.DeleteBudget)

		// Transaction routes
		auth.GET("/transactions", controllers.GetTransactions)
		auth.GET("/transactions/:id", controllers.GetTransactionByID)
		auth.POST("/transactions", controllers.CreateTransaction)
		auth.PUT("/transactions/:id", controllers.UpdateTransaction)
		auth.DELETE("/transactions/:id", controllers.DeleteTransaction)

		// Recurring Transaction routes
		auth.GET("/recurring-transactions", controllers.GetRecurringTransactions)
		auth.POST("/recurring-transactions", controllers.CreateRecurringTransaction)
		auth.GET("/recurring-transactions/:id", controllers.GetRecurringTransactionByID)
		auth.PUT("/recurring-transactions/:id", controllers.UpdateRecurringTransaction)
		auth.DELETE("/recurring-transactions/:id", controllers.DeleteRecurringTransaction)

		// Transfer routes
		auth.GET("/transfers", controllers.GetTransfers)
		auth.POST("/transfers", controllers.CreateTransfer)
		auth.GET("/transfers/:id", controllers.GetTransferByID)
		auth.DELETE("/transfers/:id", controllers.DeleteTransfer)

		// Saving Target routes
		auth.GET("/saving-targets", controllers.GetSavingTargets)
		auth.POST("/saving-targets", controllers.CreateSavingTarget)
		auth.GET("/saving-targets/:id", controllers.GetSavingTargetByID)
		auth.PUT("/saving-targets/:id", controllers.UpdateSavingTarget)
		auth.DELETE("/saving-targets/:id", controllers.DeleteSavingTarget)

		// Dashboard routes
		auth.GET("/dashboard", controllers.GetDashboard)
	}

	return r
}
