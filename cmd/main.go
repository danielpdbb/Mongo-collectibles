package main

import (
	"log"

	"github.com/danielpdbb/Mongo-collectibles/internal/api"
	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// Connect to PostgreSQL database
	repository.ConnectDatabase()

	// Auto-migrate database tables (creates tables if they don't exist)
	repository.DB.AutoMigrate(
		&domain.Store{},
		&domain.Warehouse{},
		&domain.Collectible{},
		&domain.CollectibleUnit{},
		&domain.WarehouseDistance{},
	)

	// ⚠️ Seed data - Run ONCE to populate database, then comment out
	// repository.SeedData()

	// --------------------
	// PAGE ROUTES (serve HTML files)
	// --------------------
	r.GET("/", api.ShowHome)               // Home page - product catalogue
	r.GET("/checkout", api.ShowCheckout)   // Checkout page
	r.GET("/product/:id", api.ShowProduct) // Product detail page

	// --------------------
	// API ROUTES (return JSON)
	// --------------------
	r.GET("/catalogue", api.GetCatalogue) // Get all products with availability
	r.GET("/stores", api.GetStores)       // Get all stores
	r.POST("/quote", api.CreateQuote)     // Calculate rental price quote

	// Start server
	log.Println("🚀 Server starting on http://localhost:8080")
	r.Run(":8080")
}
