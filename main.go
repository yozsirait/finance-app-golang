package main

import (
	"finance-app/database"
	"finance-app/routes"
)

func main() {
	// Initialize database
	database.InitDB()
	database.MigrateDB()

	// Setup router
	router := routes.SetupRouter()

	// Start server
	router.Run(":8080")
}
