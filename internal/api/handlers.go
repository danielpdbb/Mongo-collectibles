package api //name of the package and how it will be called in other files

import (
	"net/http" // Go's standard library for HTTP stuff (you'll use http.StatusOK, http.StatusBadRequest, etc.)

	"github.com/danielpdbb/Mongo-collectibles/internal/service" // My own package for business logic (like calculating rental prices and handling payments)
	"github.com/gin-gonic/gin"                                  // Gin web framework for handling HTTP requests and responses
)

type QuoteRequest struct { // Struct to hold the incoming JSON data for a quote request
	Size string `json:"size"`
	Days int    `json:"days"`
}

func ShowHome(c *gin.Context) {
	c.File("./web/index.html")
}

func Checkout(c *gin.Context) {
	c.File("./web/checkout.html")
}

func CreateQuote(c *gin.Context) {
	var req QuoteRequest // variable to hold the incoming JSON data

	//BindJSON = error checker for JSON

	if err := c.BindJSON(&req); err != nil { //JSON to Go Data
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request"}) //Convert Go Data to JSON and builds JSON object
	}

	price := service.CalculateRentalPrice(req.Size, req.Days)

	c.JSON(http.StatusOK, gin.H{ //http.StatusOK = 200, gin.H creates a map for JSON
		"total_price": price,
	})
}

// THIS IS STILL TEMPORARY

type PaymentRequest struct {
	Amount int    `json:"amount"`
	Method string `json:"method"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

func CreatePayment(c *gin.Context) {
	var req PaymentRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment request"})
		return
	}

	ref := service.CreatePayMongoPayment(service.PaymentRequest{
		Amount: req.Amount,
		Method: req.Method,
		Name:   req.Name,
		Email:  req.Email,
	})

	c.JSON(http.StatusOK, gin.H{
		"status":    "payment_created",
		"reference": ref,
	})
}
