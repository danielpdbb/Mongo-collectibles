package api

import (
	"net/http"
	"strconv"

	"github.com/danielpdbb/Mongo-collectibles/internal/service"
	"github.com/gin-gonic/gin"
)

// ========================================
// PAYMENT API HANDLERS
// ========================================

// CreatePayment initiates a payment for a rental
// POST /api/payments
// Requires: Authorization header with Bearer token
// Request body: payment details including method and billing info
func CreatePaymentHandler(c *gin.Context) {
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
		RentalID      uint   `json:"rental_id" binding:"required"`
		RentalIDs     []uint `json:"rental_ids"`                        // Optional: for cart checkout with multiple rentals
		PaymentMethod string `json:"payment_method" binding:"required"` // card, gcash, grab_pay, dob_ubp, dob_bpi

		// Billing details
		BillingName    string `json:"billing_name" binding:"required"`
		BillingEmail   string `json:"billing_email" binding:"required,email"`
		BillingPhone   string `json:"billing_phone" binding:"required"`
		BillingAddress string `json:"billing_address"`
		BillingCity    string `json:"billing_city"`
		BillingState   string `json:"billing_state"`
		BillingPostal  string `json:"billing_postal"`

		// Card details (only for card payments)
		CardNumber string `json:"card_number"`
		ExpMonth   int    `json:"exp_month"`
		ExpYear    int    `json:"exp_year"`
		CVC        string `json:"cvc"`

		// Redirect URLs
		SuccessURL string `json:"success_url"`
		FailedURL  string `json:"failed_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate payment method
	validMethods := map[string]bool{
		"card":     true,
		"gcash":    true,
		"grab_pay": true,
		"dob_ubp":  true,
		"dob_bpi":  true,
	}
	if !validMethods[req.PaymentMethod] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid payment method. Supported: card, gcash, grab_pay, dob_ubp, dob_bpi",
		})
		return
	}

	// Determine rental IDs to process
	rentalIDs := req.RentalIDs
	if len(rentalIDs) == 0 {
		rentalIDs = []uint{req.RentalID}
	}

	// Calculate total amount and verify all rentals
	totalAmount := 0
	for _, rentalID := range rentalIDs {
		rental, err := service.GetRentalByID(rentalID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Rental not found: " + strconv.Itoa(int(rentalID)),
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

		// Check if rental is still valid for payment (not expired)
		isValid, err := service.CheckAndExpireRental(rentalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to check rental status",
			})
			return
		}

		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Rental has expired. Please create a new order.",
			})
			return
		}

		// Only allow payment for pending_payment rentals
		if rental.Status != "pending_payment" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Rental cannot be paid. Status: " + rental.Status,
			})
			return
		}

		totalAmount += rental.TotalPrice
	}

	// Set default redirect URLs if not provided
	baseURL := "http://localhost:8080"
	rentalIDsStr := ""
	for i, id := range rentalIDs {
		if i > 0 {
			rentalIDsStr += ","
		}
		rentalIDsStr += strconv.Itoa(int(id))
	}
	if req.SuccessURL == "" {
		req.SuccessURL = baseURL + "/payment/success?rental_ids=" + rentalIDsStr
	}
	if req.FailedURL == "" {
		req.FailedURL = baseURL + "/payment/failed?rental_ids=" + rentalIDsStr
	}

	// Create payment (use first rental as primary)
	result, err := service.CreatePayment(service.PaymentRequest{
		RentalID:       rentalIDs[0], // Primary rental
		RentalIDs:      rentalIDs,    // All rentals for this payment
		Amount:         totalAmount,  // Total of all rentals
		PaymentMethod:  req.PaymentMethod,
		BillingName:    req.BillingName,
		BillingEmail:   req.BillingEmail,
		BillingPhone:   req.BillingPhone,
		BillingAddress: req.BillingAddress,
		BillingCity:    req.BillingCity,
		BillingState:   req.BillingState,
		BillingPostal:  req.BillingPostal,
		CardNumber:     req.CardNumber,
		ExpMonth:       req.ExpMonth,
		ExpYear:        req.ExpYear,
		CVC:            req.CVC,
		SuccessURL:     req.SuccessURL,
		FailedURL:      req.FailedURL,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
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

	// Return payment details
	c.JSON(http.StatusCreated, gin.H{
		"success":      true,
		"message":      result.Message,
		"payment_id":   result.PaymentID,
		"checkout_url": result.CheckoutURL, // For redirect-based payments
		"client_key":   result.ClientKey,   // For card 3DS
		"status":       result.Status,
	})
}

// VerifyPayment checks the status of a payment
// GET /api/payments/:id/verify
// Requires: Authorization header with Bearer token
func VerifyPaymentHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	// Parse payment ID from URL
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid payment ID",
		})
		return
	}

	// Verify payment and get status
	result, err := service.VerifyPayment(uint(paymentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// For extra security, we should verify the payment belongs to the user
	// This is done by checking the rental
	_ = userID // Use in production to verify ownership

	c.JSON(http.StatusOK, gin.H{
		"success":    result.Success,
		"message":    result.Message,
		"payment_id": result.PaymentID,
		"status":     result.Status,
	})
}

// GetPaymentByRental retrieves payment info for a rental
// GET /api/rentals/:id/payment
// Requires: Authorization header with Bearer token
func GetRentalPaymentHandler(c *gin.Context) {
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

	// Verify rental belongs to user
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

	// Get payment
	payment, err := service.GetPaymentByRentalID(uint(rentalID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "No payment found for this rental",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"payment": gin.H{
			"id":             payment.ID,
			"rental_id":      payment.RentalID,
			"amount":         payment.Amount / 100, // Convert centavos to PHP
			"currency":       payment.Currency,
			"status":         payment.Status,
			"payment_method": payment.PaymentMethod,
			"checkout_url":   payment.CheckoutURL,
			"paid_at":        payment.PaidAt,
			"created_at":     payment.CreatedAt,
		},
	})
}

// ========================================
// PAYMENT PAGE HANDLERS
// ========================================

// ShowPayment serves the payment page
func ShowPayment(c *gin.Context) {
	c.File("./web/payment.html")
}

// ShowPaymentSuccess serves the payment success page
func ShowPaymentSuccess(c *gin.Context) {
	c.File("./web/payment-success.html")
}

// ShowPaymentFailed serves the payment failed page
func ShowPaymentFailed(c *gin.Context) {
	c.File("./web/payment-failed.html")
}
