package domain

import (
	"time"

	"gorm.io/gorm"
)

// ========================================
// PAYMENT MODELS
// ========================================

// Payment represents a payment record for a rental
type Payment struct {
	gorm.Model
	RentalID        uint       `json:"rental_id"`
	Rental          Rental     `json:"rental,omitempty"`
	Amount          int        `json:"amount"`            // Amount in centavos (PHP)
	Currency        string     `json:"currency"`          // Always "PHP"
	Status          string     `json:"status"`            // pending, paid, failed, refunded
	PaymentMethod   string     `json:"payment_method"`    // card, gcash, grab_pay, dob_ubp, dob_bpi
	PayMongoID      string     `json:"paymongo_id"`       // PayMongo payment/source ID
	PaymentIntentID string     `json:"payment_intent_id"` // For card payments
	SourceID        string     `json:"source_id"`         // For e-wallet/bank payments
	CheckoutURL     string     `json:"checkout_url"`      // Redirect URL for e-wallet/bank
	PaidAt          *time.Time `json:"paid_at,omitempty"`
}

// BillingDetails stores customer billing information
type BillingDetails struct {
	gorm.Model
	PaymentID   uint   `json:"payment_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	AddressLine string `json:"address_line"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"` // Default: "PH"
}

// ========================================
// PAYMONGO REQUEST/RESPONSE STRUCTS
// ========================================

// PayMongoSourceRequest is the request body for creating a source
type PayMongoSourceRequest struct {
	Data PayMongoSourceData `json:"data"`
}

type PayMongoSourceData struct {
	Attributes PayMongoSourceAttributes `json:"attributes"`
}

type PayMongoSourceAttributes struct {
	Amount   int               `json:"amount"`   // In centavos
	Currency string            `json:"currency"` // "PHP"
	Type     string            `json:"type"`     // gcash, grab_pay, dob_ubp, dob_bpi
	Redirect PayMongoRedirect  `json:"redirect"`
	Billing  *PayMongoBilling  `json:"billing,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"` // Additional info like description
}

type PayMongoRedirect struct {
	Success string `json:"success"`
	Failed  string `json:"failed"`
}

type PayMongoBilling struct {
	Name    string          `json:"name"`
	Email   string          `json:"email"`
	Phone   string          `json:"phone"`
	Address PayMongoAddress `json:"address"`
}

type PayMongoAddress struct {
	Line1      string `json:"line1"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// PayMongoPaymentIntentRequest for card payments
type PayMongoPaymentIntentRequest struct {
	Data PayMongoPaymentIntentData `json:"data"`
}

type PayMongoPaymentIntentData struct {
	Attributes PayMongoPaymentIntentAttributes `json:"attributes"`
}

type PayMongoPaymentIntentAttributes struct {
	Amount               int               `json:"amount"`
	Currency             string            `json:"currency"`
	PaymentMethodAllowed []string          `json:"payment_method_allowed"`
	Description          string            `json:"description"`
	StatementDescriptor  string            `json:"statement_descriptor"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

// PayMongoPaymentMethodRequest for attaching card to payment intent
type PayMongoPaymentMethodRequest struct {
	Data PayMongoPaymentMethodData `json:"data"`
}

type PayMongoPaymentMethodData struct {
	Attributes PayMongoPaymentMethodAttributes `json:"attributes"`
}

type PayMongoPaymentMethodAttributes struct {
	Type    string              `json:"type"` // "card"
	Details PayMongoCardDetails `json:"details"`
	Billing *PayMongoBilling    `json:"billing,omitempty"`
}

type PayMongoCardDetails struct {
	CardNumber string `json:"card_number"`
	ExpMonth   int    `json:"exp_month"`
	ExpYear    int    `json:"exp_year"`
	CVC        string `json:"cvc"`
}

// PayMongoResponse generic response structure
type PayMongoResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Amount    int    `json:"amount"`
			Currency  string `json:"currency"`
			Status    string `json:"status"`
			Type      string `json:"type"`
			ClientKey string `json:"client_key"`
			Redirect  struct {
				CheckoutURL string `json:"checkout_url"`
				Success     string `json:"success"`
				Failed      string `json:"failed"`
			} `json:"redirect"`
			NextAction struct {
				Type     string `json:"type"`
				Redirect struct {
					URL string `json:"url"`
				} `json:"redirect"`
			} `json:"next_action"`
		} `json:"attributes"`
	} `json:"data"`
	Errors []struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	} `json:"errors,omitempty"`
}
