package api

import (
	"net/http"

	"github.com/danielpdbb/Mongo-collectibles/internal/service"
	"github.com/gin-gonic/gin"
)

type QuoteRequest struct {
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
	var req QuoteRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	price := service.CalculateRentalPrice(req.Size, req.Days)

	c.JSON(http.StatusOK, gin.H{
		"total_price": price,
	})
}

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
