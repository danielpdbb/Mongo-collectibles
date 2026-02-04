package api

import (
	"net/http"
	"strconv"

	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/service"
	"github.com/gin-gonic/gin"
)

// ========================================
// CART PAGE HANDLER
// ========================================

// ShowCart serves the shopping cart page
func ShowCart(c *gin.Context) {
	c.File("./web/cart.html")
}

// ========================================
// CART API HANDLERS
// ========================================

// GetCart returns the user's cart with all items and calculated prices
// GET /api/cart
func GetCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	cart, err := service.GetCartWithItems(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// AddToCart adds an item to the user's cart
// POST /api/cart/items
func AddToCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req domain.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Check available stock first
	availableUnits := service.CountAvailableUnits(req.CollectibleID)
	if availableUnits == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item is out of stock"})
		return
	}

	if int64(req.Quantity) > availableUnits {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "Not enough stock available",
			"available_units": availableUnits,
		})
		return
	}

	item, err := service.AddToCart(userID.(uint), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
		return
	}

	if item == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No stock available"})
		return
	}

	// Return updated cart
	cart, _ := service.GetCartWithItems(userID.(uint))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Item added to cart",
		"item":    item,
		"cart":    cart,
	})
}

// UpdateCartItem updates quantity and days for a cart item
// PUT /api/cart/items/:id
func UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req domain.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	item, err := service.UpdateCartItem(userID.(uint), uint(itemID), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	// Return updated cart
	cart, _ := service.GetCartWithItems(userID.(uint))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cart item updated",
		"item":    item,
		"cart":    cart,
	})
}

// RemoveFromCart removes an item from the cart
// DELETE /api/cart/items/:id
func RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	if err := service.RemoveFromCart(userID.(uint), uint(itemID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	// Return updated cart
	cart, _ := service.GetCartWithItems(userID.(uint))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Item removed from cart",
		"cart":    cart,
	})
}

// ClearCart removes all items from the cart
// DELETE /api/cart
func ClearCartHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	if err := service.ClearCart(userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cart cleared",
	})
}

// GetCartCount returns just the item count (for nav badge)
// GET /api/cart/count
func GetCartCount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	count := service.GetCartItemCount(userID.(uint))

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// CheckoutCart creates rentals for all items in the cart
// POST /api/cart/checkout
func CheckoutCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Validate cart first
	valid, errMsg, err := service.ValidateCartForCheckout(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate cart"})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	// Process checkout
	result, err := service.CheckoutCart(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process checkout",
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   result.Message,
		})
		return
	}

	// Extract rental IDs for payment page
	var rentalIDs []uint
	for _, r := range result.Rentals {
		rentalIDs = append(rentalIDs, r.RentalID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":     true,
		"message":     result.Message,
		"rentals":     result.Rentals,
		"rental_ids":  rentalIDs,
		"total_price": result.TotalPrice,
	})
}
