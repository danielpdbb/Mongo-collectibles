package main

import (
	"github.com/danielpdbb/Mongo-collectibles/internal/api"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve static UI files
	r.Static("/ui", "./web")

	// Routes
	r.GET("/", api.ShowHome)
	r.POST("/quote", api.CreateQuote)
	r.POST("/checkout", api.Checkout)

	// Start server
	r.Run(":8080")
}
