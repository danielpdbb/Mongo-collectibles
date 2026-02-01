package api

import (
	"net/http"

	"github.com/danielpdbb/Mongo-collectibles/internal/service"
	"github.com/gin-gonic/gin"
)

// ========================================
// PAGE HANDLERS (serve HTML files)
// ========================================

// ShowLogin serves the login page
func ShowLogin(c *gin.Context) {
	c.File("./web/login.html")
}

// ShowRegister serves the registration page
func ShowRegister(c *gin.Context) {
	c.File("./web/register.html")
}

// ========================================
// API HANDLERS (return JSON)
// ========================================

// Register creates a new user account
// POST /api/register
// Request body: { "email": "...", "password": "...", "name": "...", "phone": "..." }
func Register(c *gin.Context) {
	// Define request structure
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required"`
		Phone    string `json:"phone"`
	}

	// Parse and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Call auth service to register user
	user, err := service.RegisterUser(req.Email, req.Password, req.Name, req.Phone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Generate token for automatic login after registration
	token, err := service.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Registration successful but failed to generate token",
		})
		return
	}

	// Return success with token and user info
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Registration successful",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone,
		},
	})
}

// Login authenticates a user and returns a JWT token
// POST /api/login
// Request body: { "email": "...", "password": "..." }
func Login(c *gin.Context) {
	// Define request structure
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// Parse and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Call auth service to login
	token, user, err := service.LoginUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Return success with token and user info
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone,
		},
	})
}

// GetCurrentUser returns the currently authenticated user's info
// GET /api/me
// Requires: Authorization header with Bearer token
func GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	// Get user from database
	user, err := service.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Return user info
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone,
		},
	})
}
