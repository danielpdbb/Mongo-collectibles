package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
)

// ========================================
// PAYMONGO SERVICE
// ========================================

const (
	paymongoBaseURL = "https://api.paymongo.com/v1"
)

// formatAmount formats a float amount with commas (e.g., 100000 -> "100,000")
func formatAmount(amount float64) string {
	str := fmt.Sprintf("%.0f", amount)
	n := len(str)
	if n <= 3 {
		return str
	}
	var result strings.Builder
	pre := n % 3
	if pre > 0 {
		result.WriteString(str[:pre])
		if n > pre {
			result.WriteString(",")
		}
	}
	for i := pre; i < n; i += 3 {
		result.WriteString(str[i : i+3])
		if i+3 < n {
			result.WriteString(",")
		}
	}
	return result.String()
}

// PaymentRequest contains all data needed to create a payment
type PaymentRequest struct {
	RentalID      uint   `json:"rental_id"`
	Amount        int    `json:"amount"`         // In PHP (will convert to centavos)
	PaymentMethod string `json:"payment_method"` // card, gcash, grab_pay, dob_ubp, dob_bpi

	// Billing details
	BillingName    string `json:"billing_name"`
	BillingEmail   string `json:"billing_email"`
	BillingPhone   string `json:"billing_phone"`
	BillingAddress string `json:"billing_address"`
	BillingCity    string `json:"billing_city"`
	BillingState   string `json:"billing_state"`
	BillingPostal  string `json:"billing_postal"`

	// For card payments
	CardNumber string `json:"card_number,omitempty"`
	ExpMonth   int    `json:"exp_month,omitempty"`
	ExpYear    int    `json:"exp_year,omitempty"`
	CVC        string `json:"cvc,omitempty"`

	// Redirect URLs
	SuccessURL string `json:"success_url"`
	FailedURL  string `json:"failed_url"`
}

// PaymentResponse contains the result of creating a payment
type PaymentResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	PaymentID   uint   `json:"payment_id,omitempty"`
	CheckoutURL string `json:"checkout_url,omitempty"` // For redirect-based payments
	ClientKey   string `json:"client_key,omitempty"`   // For card 3DS
	Status      string `json:"status,omitempty"`
}

// ========================================
// CREATE PAYMENT (Main Entry Point)
// ========================================

// CreatePayment creates a payment using the appropriate PayMongo method
func CreatePayment(req PaymentRequest) (*PaymentResponse, error) {
	// Validate rental exists and get details
	rental, err := GetRentalByID(req.RentalID)
	if err != nil {
		return nil, errors.New("rental not found")
	}

	// Payment method limits (in PHP)
	methodLimits := map[string]int{
		"card":     10000000, // ₱10,000,000
		"gcash":    100000,   // ₱100,000
		"grab_pay": 100000,   // ₱100,000
		"dob_ubp":  100000,   // ₱100,000
		"dob_bpi":  50000,    // ₱50,000
	}

	// Check payment method limit
	limit, exists := methodLimits[req.PaymentMethod]
	if !exists {
		return nil, errors.New("unsupported payment method")
	}

	if req.Amount > limit {
		return nil, fmt.Errorf("amount exceeds %s limit of ₱%s. Please choose a different payment method",
			req.PaymentMethod, formatAmount(float64(limit)))
	}

	// Convert PHP to centavos (PayMongo requires centavos)
	amountInCentavos := req.Amount * 100

	// Create payment record in our database
	payment := &domain.Payment{
		RentalID:      req.RentalID,
		Amount:        amountInCentavos,
		Currency:      "PHP",
		Status:        "pending",
		PaymentMethod: req.PaymentMethod,
	}

	if err := repository.DB.Create(payment).Error; err != nil {
		return nil, errors.New("failed to create payment record")
	}

	// Save billing details
	billing := &domain.BillingDetails{
		PaymentID:   payment.ID,
		Name:        req.BillingName,
		Email:       req.BillingEmail,
		Phone:       req.BillingPhone,
		AddressLine: req.BillingAddress,
		City:        req.BillingCity,
		State:       req.BillingState,
		PostalCode:  req.BillingPostal,
		Country:     "PH",
	}
	repository.DB.Create(billing)

	// Route to appropriate payment method
	switch req.PaymentMethod {
	case "gcash", "grab_pay", "dob_ubp", "dob_bpi":
		return createSourcePayment(payment, req, amountInCentavos, rental)
	case "card":
		return createCardPayment(payment, req, amountInCentavos, rental)
	default:
		return nil, errors.New("unsupported payment method")
	}
}

// ========================================
// SOURCE-BASED PAYMENTS (GCash, GrabPay, BPI, UBP)
// ========================================

func createSourcePayment(payment *domain.Payment, req PaymentRequest, amount int, rental *domain.Rental) (*PaymentResponse, error) {
	// Build PayMongo source request
	sourceReq := domain.PayMongoSourceRequest{
		Data: domain.PayMongoSourceData{
			Attributes: domain.PayMongoSourceAttributes{
				Amount:   amount,
				Currency: "PHP",
				Type:     req.PaymentMethod,
				Redirect: domain.PayMongoRedirect{
					Success: req.SuccessURL,
					Failed:  req.FailedURL,
				},
				Billing: &domain.PayMongoBilling{
					Name:  req.BillingName,
					Email: req.BillingEmail,
					Phone: req.BillingPhone,
					Address: domain.PayMongoAddress{
						Line1:      req.BillingAddress,
						City:       req.BillingCity,
						State:      req.BillingState,
						PostalCode: req.BillingPostal,
						Country:    "PH",
					},
				},
			},
		},
	}

	// Add description metadata
	if sourceReq.Data.Attributes.Metadata == nil {
		sourceReq.Data.Attributes.Metadata = make(map[string]string)
	}
	sourceReq.Data.Attributes.Metadata["description"] = fmt.Sprintf("Rental #%d - %s", rental.ID, rental.Collectible.Name)
	sourceReq.Data.Attributes.Metadata["rental_id"] = fmt.Sprintf("%d", rental.ID)
	sourceReq.Data.Attributes.Metadata["payment_id"] = fmt.Sprintf("%d", payment.ID)

	// Call PayMongo API
	resp, err := callPayMongoAPI("POST", "/sources", sourceReq)
	if err != nil {
		payment.Status = "failed"
		repository.DB.Save(payment)
		return nil, err
	}

	// Extract checkout URL and source ID
	sourceID := resp.Data.ID
	checkoutURL := resp.Data.Attributes.Redirect.CheckoutURL

	// Update payment record
	payment.SourceID = sourceID
	payment.CheckoutURL = checkoutURL
	payment.PayMongoID = sourceID
	repository.DB.Save(payment)

	return &PaymentResponse{
		Success:     true,
		Message:     "Redirect to complete payment",
		PaymentID:   payment.ID,
		CheckoutURL: checkoutURL,
		Status:      "pending",
	}, nil
}

// ========================================
// CARD PAYMENTS (Payment Intent Flow)
// ========================================

func createCardPayment(payment *domain.Payment, req PaymentRequest, amount int, rental *domain.Rental) (*PaymentResponse, error) {
	// Step 1: Create Payment Intent
	intentReq := domain.PayMongoPaymentIntentRequest{
		Data: domain.PayMongoPaymentIntentData{
			Attributes: domain.PayMongoPaymentIntentAttributes{
				Amount:               amount,
				Currency:             "PHP",
				PaymentMethodAllowed: []string{"card"},
				Description:          fmt.Sprintf("Rental #%d - %s", rental.ID, rental.Collectible.Name),
				StatementDescriptor:  "MONGOCOLLECT",
				Metadata: map[string]string{
					"rental_id":  fmt.Sprintf("%d", rental.ID),
					"payment_id": fmt.Sprintf("%d", payment.ID),
				},
			},
		},
	}

	intentResp, err := callPayMongoAPI("POST", "/payment_intents", intentReq)
	if err != nil {
		payment.Status = "failed"
		repository.DB.Save(payment)
		return nil, err
	}

	paymentIntentID := intentResp.Data.ID
	clientKey := intentResp.Data.Attributes.ClientKey

	// Update payment record with intent ID
	payment.PaymentIntentID = paymentIntentID
	payment.PayMongoID = paymentIntentID
	repository.DB.Save(payment)

	// Step 2: Create Payment Method
	methodReq := domain.PayMongoPaymentMethodRequest{
		Data: domain.PayMongoPaymentMethodData{
			Attributes: domain.PayMongoPaymentMethodAttributes{
				Type: "card",
				Details: domain.PayMongoCardDetails{
					CardNumber: req.CardNumber,
					ExpMonth:   req.ExpMonth,
					ExpYear:    req.ExpYear,
					CVC:        req.CVC,
				},
				Billing: &domain.PayMongoBilling{
					Name:  req.BillingName,
					Email: req.BillingEmail,
					Phone: req.BillingPhone,
					Address: domain.PayMongoAddress{
						Line1:      req.BillingAddress,
						City:       req.BillingCity,
						State:      req.BillingState,
						PostalCode: req.BillingPostal,
						Country:    "PH",
					},
				},
			},
		},
	}

	methodResp, err := callPayMongoAPI("POST", "/payment_methods", methodReq)
	if err != nil {
		return nil, err
	}

	paymentMethodID := methodResp.Data.ID

	// Step 3: Attach Payment Method to Payment Intent
	attachReq := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"payment_method": paymentMethodID,
				"client_key":     clientKey,
				"return_url":     req.SuccessURL,
			},
		},
	}

	attachResp, err := callPayMongoAPI("POST", fmt.Sprintf("/payment_intents/%s/attach", paymentIntentID), attachReq)
	if err != nil {
		return nil, err
	}

	// Check if 3DS is required
	status := attachResp.Data.Attributes.Status
	if status == "awaiting_next_action" {
		// 3DS required - return redirect URL
		redirectURL := attachResp.Data.Attributes.NextAction.Redirect.URL
		payment.CheckoutURL = redirectURL
		repository.DB.Save(payment)

		return &PaymentResponse{
			Success:     true,
			Message:     "3DS verification required",
			PaymentID:   payment.ID,
			CheckoutURL: redirectURL,
			ClientKey:   clientKey,
			Status:      "awaiting_3ds",
		}, nil
	}

	if status == "succeeded" {
		// Payment successful!
		now := time.Now()
		payment.Status = "paid"
		payment.PaidAt = &now
		repository.DB.Save(payment)

		// ACTIVATE THE RENTAL - Non-3DS card payment succeeded
		ActivateRental(payment.RentalID)

		return &PaymentResponse{
			Success:   true,
			Message:   "Payment successful",
			PaymentID: payment.ID,
			Status:    "paid",
		}, nil
	}

	return &PaymentResponse{
		Success:   true,
		Message:   "Payment processing",
		PaymentID: payment.ID,
		Status:    status,
	}, nil
}

// ========================================
// VERIFY PAYMENT STATUS
// ========================================

// VerifyPayment checks the status of a payment and updates accordingly
func VerifyPayment(paymentID uint) (*PaymentResponse, error) {
	var payment domain.Payment
	if err := repository.DB.First(&payment, paymentID).Error; err != nil {
		return nil, errors.New("payment not found")
	}

	if payment.Status == "paid" {
		return &PaymentResponse{
			Success:   true,
			PaymentID: payment.ID,
			Status:    "paid",
			Message:   "Payment already confirmed",
		}, nil
	}

	var resp *domain.PayMongoResponse
	var err error

	// Check based on payment type
	if payment.SourceID != "" {
		// Source-based payment (GCash, GrabPay, etc.)
		resp, err = callPayMongoAPI("GET", fmt.Sprintf("/sources/%s", payment.SourceID), nil)
	} else if payment.PaymentIntentID != "" {
		// Card payment
		resp, err = callPayMongoAPI("GET", fmt.Sprintf("/payment_intents/%s", payment.PaymentIntentID), nil)
	} else {
		return nil, errors.New("invalid payment record")
	}

	if err != nil {
		return nil, err
	}

	status := resp.Data.Attributes.Status

	// Update payment status based on PayMongo status
	if status == "chargeable" || status == "succeeded" || status == "paid" {
		now := time.Now()
		payment.Status = "paid"
		payment.PaidAt = &now
		repository.DB.Save(payment)

		// ACTIVATE THE RENTAL - Payment successful, confirm the rental
		ActivateRental(payment.RentalID)

		// If source is chargeable, we need to create a payment
		if status == "chargeable" && payment.SourceID != "" {
			go createPaymentFromSource(payment)
		}

		return &PaymentResponse{
			Success:   true,
			PaymentID: payment.ID,
			Status:    "paid",
			Message:   "Payment successful",
		}, nil
	}

	if status == "cancelled" || status == "expired" || status == "failed" {
		payment.Status = "failed"
		repository.DB.Save(payment)

		// CANCEL THE RENTAL - Return units to available
		CancelRental(payment.RentalID)

		return &PaymentResponse{
			Success:   false,
			PaymentID: payment.ID,
			Status:    "failed",
			Message:   "Payment failed or cancelled",
		}, nil
	}

	return &PaymentResponse{
		Success:   true,
		PaymentID: payment.ID,
		Status:    status,
		Message:   "Payment pending",
	}, nil
}

// createPaymentFromSource creates a payment from a chargeable source
func createPaymentFromSource(payment domain.Payment) {
	payReq := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"amount":   payment.Amount,
				"currency": "PHP",
				"source": map[string]interface{}{
					"id":   payment.SourceID,
					"type": "source",
				},
			},
		},
	}

	resp, err := callPayMongoAPI("POST", "/payments", payReq)
	if err != nil {
		fmt.Println("Error creating payment from source:", err)
		return
	}

	if resp.Data.Attributes.Status == "paid" {
		now := time.Now()
		payment.Status = "paid"
		payment.PaidAt = &now
		repository.DB.Save(&payment)
	}
}

// ========================================
// GET PAYMENT BY RENTAL
// ========================================

func GetPaymentByRentalID(rentalID uint) (*domain.Payment, error) {
	var payment domain.Payment
	err := repository.DB.Where("rental_id = ?", rentalID).Order("created_at DESC").First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// ========================================
// PAYMONGO API HELPER
// ========================================

func callPayMongoAPI(method, endpoint string, body interface{}) (*domain.PayMongoResponse, error) {
	secretKey := os.Getenv("PAYMONGO_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("PAYMONGO_SECRET_KEY not configured")
	}

	url := paymongoBaseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	// PayMongo uses Basic Auth with secret key as username
	auth := base64.StdEncoding.EncodeToString([]byte(secretKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var paymongoResp domain.PayMongoResponse
	if err := json.Unmarshal(respBody, &paymongoResp); err != nil {
		return nil, err
	}

	// Check for API errors
	if len(paymongoResp.Errors) > 0 {
		return nil, errors.New(paymongoResp.Errors[0].Detail)
	}

	return &paymongoResp, nil
}
