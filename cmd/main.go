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
		&domain.User{},       // User table for authentication
		&domain.Rental{},     // Rental orders
		&domain.RentalUnit{}, // Allocated units per rental
	)

	// ⚠️ Seed data - Run ONCE to populate database, then comment out
	// repository.SeedData()

	// --------------------
	// PAGE ROUTES (serve HTML files)
	// --------------------
	r.GET("/", api.ShowHome)               // Home page - product catalogue
	r.GET("/checkout", api.ShowCheckout)   // Checkout page
	r.GET("/rentals", api.ShowRentals)     // User's rentals page
	r.GET("/product/:id", api.ShowProduct) // Product detail page
	r.GET("/login", api.ShowLogin)         // Login page
	r.GET("/register", api.ShowRegister)   // Register page

	// --------------------
	// API ROUTES (return JSON)
	// --------------------
	r.GET("/catalogue", api.GetCatalogue) // Get all products with availability
	r.GET("/stores", api.GetStores)       // Get all stores
	r.POST("/quote", api.CreateQuote)     // Calculate rental price quote

	// --------------------
	// AUTH API ROUTES
	// --------------------
	r.POST("/api/register", api.Register)                    // Create new account
	r.POST("/api/login", api.Login)                          // Login and get token
	r.GET("/api/me", api.AuthRequired(), api.GetCurrentUser) // Get current user (requires auth)

	// --------------------
	// RENTAL API ROUTES (all require authentication)
	// --------------------
	r.POST("/api/rentals", api.AuthRequired(), api.CreateRental)                   // Create new rental
	r.GET("/api/rentals", api.AuthRequired(), api.GetMyRentals)                    // Get user's rentals
	r.GET("/api/rentals/:id", api.AuthRequired(), api.GetRental)                   // Get specific rental
	r.POST("/api/rentals/:id/cancel", api.AuthRequired(), api.CancelRentalHandler) // Cancel rental

	// Start server
	log.Println("🚀 Server starting on http://localhost:8080")
	r.Run(":8080")
}
