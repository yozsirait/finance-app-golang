package main

import (
	"finance-app/config"
	"finance-app/database"
	"finance-app/routes"
	"fmt"

	_ "finance-app/docs" // docs swagger hasil dari swag init

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Finance App API
// @version 1.0
// @description API documentation for Finance App
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@finance-app.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Format: "Bearer {token}"
func main() {
	// Load config
	cfg := config.LoadConfig()

	// DEBUG: print JWT secret yang dipakai
	fmt.Println("🚀 Using JWT_SECRET:", cfg.JWTSecret)

	// Initialize database
	database.InitDB()
	database.MigrateDB()

	// Setup router
	router := routes.SetupRouter()

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	router.Run(":8080")
}
