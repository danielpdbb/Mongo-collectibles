package main

import (
	"log"

	"github.com/danielpdbb/Mongo-collectibles/internal/api"
	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

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
		&domain.User{},           // User table for authentication
		&domain.Rental{},         // Rental orders
		&domain.RentalUnit{},     // Allocated units per rental
		&domain.Payment{},        // Payment records
		&domain.BillingDetails{}, // Billing information
		&domain.Cart{},           // Shopping carts
		&domain.CartItem{},       // Cart items
	)

	// ⚠️ Seed data - Run ONCE to populate database, then comment out
	repository.SeedData()

	// --------------------
	// PAGE ROUTES (serve HTML files)
	// --------------------
	r.GET("/", api.ShowHome)               // Home page - product catalogue
	r.GET("/checkout", api.ShowCheckout)   // Checkout page
	r.GET("/rentals", api.ShowRentals)     // User's rentals page
	r.GET("/product/:id", api.ShowProduct) // Product detail page
	r.GET("/login", api.ShowLogin)         // Login page
	r.GET("/register", api.ShowRegister)   // Register page
	r.GET("/payment", api.ShowPayment)     // Payment page
	r.GET("/cart", api.ShowCart)           // Shopping cart page

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
	r.POST("/api/rentals", api.AuthRequired(), api.CreateRental)                       // Create new rental
	r.GET("/api/rentals", api.AuthRequired(), api.GetMyRentals)                        // Get user's rentals
	r.GET("/api/rentals/:id", api.AuthRequired(), api.GetRental)                       // Get specific rental
	r.POST("/api/rentals/:id/cancel", api.AuthRequired(), api.CancelRentalHandler)     // Cancel rental
	r.GET("/api/rentals/:id/payment", api.AuthRequired(), api.GetRentalPaymentHandler) // Get payment for rental

	// --------------------
	// CART API ROUTES (all require authentication)
	// --------------------
	r.GET("/api/cart", api.AuthRequired(), api.GetCart)                     // Get user's cart
	r.POST("/api/cart/items", api.AuthRequired(), api.AddToCart)            // Add item to cart
	r.PUT("/api/cart/items/:id", api.AuthRequired(), api.UpdateCartItem)    // Update cart item
	r.DELETE("/api/cart/items/:id", api.AuthRequired(), api.RemoveFromCart) // Remove item from cart
	r.DELETE("/api/cart", api.AuthRequired(), api.ClearCartHandler)         // Clear entire cart
	r.GET("/api/cart/count", api.AuthRequired(), api.GetCartCount)          // Get cart item count
	r.POST("/api/cart/checkout", api.AuthRequired(), api.CheckoutCart)      // Checkout cart (create rentals)

	// --------------------
	// PAYMENT API ROUTES (all require authentication)
	// --------------------
	r.POST("/api/payments", api.AuthRequired(), api.CreatePaymentHandler)           // Create payment
	r.GET("/api/payments/:id/verify", api.AuthRequired(), api.VerifyPaymentHandler) // Verify payment status

	// --------------------
	// PAYMENT RESULT PAGES
	// --------------------
	r.GET("/payment/success", api.ShowPaymentSuccess) // Payment success redirect
	r.GET("/payment/failed", api.ShowPaymentFailed)   // Payment failed redirect

	// Start server
	log.Println("🚀 Server starting on http://localhost:8080")
	r.Run(":8080")
}
