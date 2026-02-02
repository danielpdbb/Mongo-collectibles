package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/danielpdbb/Mongo-collectibles/internal/service"
	"github.com/gin-gonic/gin"
)

// ========================================
// RENTAL API HANDLERS
// ========================================

// CreateRental creates a new rental with automatic unit allocation
// POST /api/rentals
// Requires: Authorization header with Bearer token
// Request body: { "collectible_id": 1, "store_id": 1, "quantity": 2, "days": 7 }
func CreateRental(c *gin.Context) {
	// Get user ID from auth middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	// Parse request body
	var req struct {
		CollectibleID uint `json:"collectible_id" binding:"required"`
		StoreID       uint `json:"store_id" binding:"required"`
		Quantity      int  `json:"quantity" binding:"required,min=1"`
		Days          int  `json:"days" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Create the rental
	result, err := service.CreateRental(service.CreateRentalRequest{
		UserID:        userID.(uint),
		CollectibleID: req.CollectibleID,
		StoreID:       req.StoreID,
		Quantity:      req.Quantity,
		Days:          req.Days,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   result.Message,
		})
		return
	}

	// Return success with rental details
	c.JSON(http.StatusCreated, gin.H{
		"success":         true,
		"message":         "Rental created successfully",
		"rental_id":       result.Rental.ID,
		"total_price":     result.Rental.TotalPrice,
		"unit_price":      result.Rental.UnitPrice,
		"quantity":        result.Rental.Quantity,
		"days":            result.Rental.Days,
		"start_date":      result.Rental.StartDate,
		"end_date":        result.Rental.EndDate,
		"allocated_units": result.Allocated,
	})
}

// GetMyRentals retrieves all rentals for the authenticated user
// GET /api/rentals
// Requires: Authorization header with Bearer token
func GetMyRentals(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	rentals, err := service.GetUserRentals(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch rentals",
		})
		return
	}

	// Format response
	var rentalList []gin.H
	for _, r := range rentals {
		// Get warehouse names from rental units
		var warehouses []string
		warehouseMap := make(map[string]bool)
		for _, ru := range r.RentalUnits {
			if !warehouseMap[ru.Warehouse.Name] {
				warehouses = append(warehouses, ru.Warehouse.Name)
				warehouseMap[ru.Warehouse.Name] = true
			}
		}

		// Calculate time remaining for pending_payment rentals
		var timeRemaining *int
		if r.Status == "pending_payment" && r.ExpiresAt != nil {
			remaining := int(time.Until(*r.ExpiresAt).Seconds())
			if remaining < 0 {
				remaining = 0
			}
			timeRemaining = &remaining
		}

		rentalList = append(rentalList, gin.H{
			"id":             r.ID,
			"collectible":    r.Collectible.Name,
			"collectible_id": r.CollectibleID,
			"size":           r.Collectible.Size,
			"image_url":      r.Collectible.ImageURL,
			"store":          r.Store.Name,
			"quantity":       r.Quantity,
			"days":           r.Days,
			"unit_price":     r.UnitPrice,
			"total_price":    r.TotalPrice,
			"status":         r.Status,
			"time_remaining": timeRemaining, // Seconds remaining for payment
			"start_date":     r.StartDate,
			"end_date":       r.EndDate,
			"warehouses":     warehouses,
			"created_at":     r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"rentals": rentalList,
	})
}

// GetRental retrieves a specific rental by ID
// GET /api/rentals/:id
// Requires: Authorization header with Bearer token
func GetRental(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	// Parse rental ID from URL
	rentalID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid rental ID",
		})
		return
	}

	rental, err := service.GetRentalByID(uint(rentalID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Rental not found",
		})
		return
	}

	// Check that rental belongs to user
	if rental.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	// Check and expire if past deadline (for pending_payment rentals)
	service.CheckAndExpireRental(uint(rentalID))

	// Reload rental to get updated status
	rental, _ = service.GetRentalByID(uint(rentalID))

	// Format allocated units
	var allocatedUnits []gin.H
	for _, ru := range rental.RentalUnits {
		allocatedUnits = append(allocatedUnits, gin.H{
			"unit_id":        ru.CollectibleUnitID,
			"warehouse_id":   ru.WarehouseID,
			"warehouse_name": ru.Warehouse.Name,
		})
	}

	// Calculate time remaining for pending_payment rentals
	var expiresAt *string
	var timeRemaining *int
	if rental.Status == "pending_payment" && rental.ExpiresAt != nil {
		formatted := rental.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		expiresAt = &formatted
		remaining := int(time.Until(*rental.ExpiresAt).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		timeRemaining = &remaining
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"rental": gin.H{
			"id":              rental.ID,
			"collectible":     rental.Collectible.Name,
			"size":            rental.Collectible.Size,
			"image_url":       rental.Collectible.ImageURL,
			"store":           rental.Store.Name,
			"quantity":        rental.Quantity,
			"days":            rental.Days,
			"unit_price":      rental.UnitPrice,
			"total_price":     rental.TotalPrice,
			"status":          rental.Status,
			"expires_at":      expiresAt,
			"time_remaining":  timeRemaining, // Seconds remaining for payment
			"start_date":      rental.StartDate,
			"end_date":        rental.EndDate,
			"allocated_units": allocatedUnits,
			"created_at":      rental.CreatedAt,
		},
	})
}

// CancelRental cancels a rental and returns units to inventory
// POST /api/rentals/:id/cancel
// Requires: Authorization header with Bearer token
func CancelRentalHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	// Parse rental ID
	rentalID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid rental ID",
		})
		return
	}

	// Get rental to verify ownership
	rental, err := service.GetRentalByID(uint(rentalID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Rental not found",
		})
		return
	}

	if rental.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	if rental.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Can only cancel active rentals",
		})
		return
	}

	// Cancel the rental
	if err := service.CancelRental(uint(rentalID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cancel rental",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rental cancelled successfully. Units have been returned to inventory.",
	})
}
