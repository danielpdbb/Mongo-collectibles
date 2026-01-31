package api

import (
	"net/http"

	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
	"github.com/danielpdbb/Mongo-collectibles/internal/service"
	"github.com/gin-gonic/gin"
)

// ========================================
// PAGE HANDLERS (serve HTML files)
// ========================================

// ShowHome serves the main catalogue page
func ShowHome(c *gin.Context) {
	c.File("./web/index.html")
}

// ShowCheckout serves the checkout page
func ShowCheckout(c *gin.Context) {
	c.File("./web/checkout.html")
}

// ShowProduct serves the product detail page
func ShowProduct(c *gin.Context) {
	c.File("./web/product-detail.html")
}

// ========================================
// API HANDLERS (return JSON)
// ========================================

// GetCatalogue returns all products with their availability count
// GET /catalogue
func GetCatalogue(c *gin.Context) {
	var collectibles []domain.Collectible
	repository.DB.Find(&collectibles)

	// Build response with availability info
	type ProductResponse struct {
		ID             uint   `json:"id"`
		Name           string `json:"name"`
		Size           string `json:"size"`
		ImageURL       string `json:"imageURL"`
		AvailableUnits int64  `json:"available_units"`
		PricePerDay    int    `json:"price_per_day"`
	}

	var products []ProductResponse
	for _, c := range collectibles {
		available := service.CountAvailableUnits(c.ID)
		pricePerDay := service.CalculateRentalPrice(c.Size, 7) / 7 // Base daily rate

		products = append(products, ProductResponse{
			ID:             c.ID,
			Name:           c.Name,
			Size:           c.Size,
			ImageURL:       c.ImageURL,
			AvailableUnits: available,
			PricePerDay:    pricePerDay,
		})
	}

	c.JSON(http.StatusOK, products)
}

// GetStores returns all available stores
// GET /stores
func GetStores(c *gin.Context) {
	var stores []domain.Store
	repository.DB.Find(&stores)
	c.JSON(http.StatusOK, stores)
}

// CreateQuote calculates a rental price quote
// POST /quote
// Request body: { "collectible_id": 1, "store_id": 1, "days": 7, "quantity": 1 }
func CreateQuote(c *gin.Context) {
	var req struct {
		CollectibleID uint `json:"collectible_id" binding:"required"`
		StoreID       uint `json:"store_id" binding:"required"`
		Days          int  `json:"days" binding:"required,min=1"`
		Quantity      int  `json:"quantity" binding:"required,min=1"`
	}

	// Parse and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Check available stock
	availableCount := service.CountAvailableUnits(req.CollectibleID)
	if int64(req.Quantity) > availableCount {
		c.JSON(http.StatusOK, gin.H{
			"available":       false,
			"message":         "Not enough units available",
			"available_units": availableCount,
		})
		return
	}

	// Find nearest available unit (without reserving it)
	unit, ok := service.FindNearestAvailableUnit(req.CollectibleID, req.StoreID)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"message":   "No units available for this product",
		})
		return
	}

	// Get product info for pricing
	var collectible domain.Collectible
	if err := repository.DB.First(&collectible, req.CollectibleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Get warehouse info
	var warehouse domain.Warehouse
	repository.DB.First(&warehouse, unit.WarehouseID)

	// Calculate price
	unitPrice := service.CalculateRentalPrice(collectible.Size, req.Days)
	totalPrice := unitPrice * req.Quantity

	// Return quote
	c.JSON(http.StatusOK, gin.H{
		"available":       true,
		"collectible":     collectible,
		"warehouse":       warehouse.Name,
		"unit_price":      unitPrice,
		"quantity":        req.Quantity,
		"days":            req.Days,
		"total_price":     totalPrice,
		"available_units": availableCount,
	})
}
