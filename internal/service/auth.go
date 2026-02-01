package service

import (
	"errors"
	"time"

	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ========================================
// CONFIGURATION
// ========================================

// JWT secret key - in production, use environment variable!
var jwtSecret = []byte("your-secret-key-change-in-production")

// Token expiration time
const tokenExpiration = 24 * time.Hour

// ========================================
// JWT CLAIMS
// ========================================

// Claims represents the JWT token claims
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// ========================================
// AUTHENTICATION FUNCTIONS
// ========================================

// RegisterUser creates a new user account
// Returns error if email already exists or validation fails
func RegisterUser(email, password, name, phone string) (*domain.User, error) {
	// Check if email already exists
	var existingUser domain.User
	result := repository.DB.Where("email = ?", email).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("email already registered")
	}

	// Hash the password using bcrypt
	// Cost of 10 is a good balance between security and performance
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create the user
	user := &domain.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
		Phone:    phone,
	}

	// Save to database
	if err := repository.DB.Create(user).Error; err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// LoginUser authenticates a user and returns a JWT token
func LoginUser(email, password string) (string, *domain.User, error) {
	// Find user by email
	var user domain.User
	result := repository.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// Compare password with hash
	// bcrypt.CompareHashAndPassword returns nil if they match
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := GenerateToken(&user)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, &user, nil
}

// ========================================
// JWT TOKEN FUNCTIONS
// ========================================

// GenerateToken creates a new JWT token for a user
func GenerateToken(user *domain.User) (string, error) {
	// Create claims with user info and expiration
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and return the token string
	return token.SignedString(jwtSecret)
}

// ValidateToken parses and validates a JWT token
// Returns the claims if valid, or error if invalid/expired
func ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Extract and return claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := repository.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
