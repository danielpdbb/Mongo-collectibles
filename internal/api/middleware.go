package api

import (
	"net/http"
	"strings"

	"github.com/danielpdbb/Mongo-collectibles/internal/service"
	"github.com/gin-gonic/gin"
)

// ========================================
// AUTHENTICATION MIDDLEWARE
// ========================================

// AuthRequired is middleware that checks for a valid JWT token
// Use this on routes that require the user to be logged in
//
// How it works:
// 1. Looks for "Authorization: Bearer <token>" header
// 2. Validates the token
// 3. Sets userID, userEmail, userName in the context
// 4. If invalid, returns 401 Unauthorized
//
// Usage in routes:
//
//	r.GET("/api/me", api.AuthRequired(), api.GetCurrentUser)
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")

		// Check if header exists and has correct format
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort() // Stop processing further handlers
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate the token
		claims, err := service.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Token is valid! Set user info in context for handlers to use
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userName", claims.Name)

		// Continue to the next handler
		c.Next()
	}
}

// OptionalAuth is middleware that checks for a JWT token but doesn't require it
// Use this on routes that work for both logged-in and anonymous users
// If token is valid, sets user info in context
// If no token or invalid token, continues without setting user info
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")

		// If no header, just continue (user is anonymous)
		if authHeader == "" {
			c.Next()
			return
		}

		// Try to parse and validate the token
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			claims, err := service.ValidateToken(tokenString)
			if err == nil {
				// Token is valid, set user info
				c.Set("userID", claims.UserID)
				c.Set("userEmail", claims.Email)
				c.Set("userName", claims.Name)
			}
		}

		// Continue regardless of token validity
		c.Next()
	}
}
