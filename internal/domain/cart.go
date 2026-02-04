package domain

import "time"

// Cart represents a user's shopping cart
type Cart struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"` // One cart per user
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User  User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items []CartItem `gorm:"foreignKey:CartID" json:"items,omitempty"`
}

// CartItem represents a single item in the shopping cart
type CartItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CartID        uint      `gorm:"not null" json:"cart_id"`
	CollectibleID uint      `gorm:"not null" json:"collectible_id"`
	StoreID       uint      `gorm:"not null" json:"store_id"`
	Quantity      int       `gorm:"not null;default:1" json:"quantity"`
	Days          int       `gorm:"not null;default:7" json:"days"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships (for eager loading)
	Collectible Collectible `gorm:"foreignKey:CollectibleID" json:"collectible,omitempty"`
	Store       Store       `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

// CartItemResponse includes calculated price info for display
type CartItemResponse struct {
	ID             uint        `json:"id"`
	CollectibleID  uint        `json:"collectible_id"`
	StoreID        uint        `json:"store_id"`
	Quantity       int         `json:"quantity"`
	Days           int         `json:"days"`
	UnitPrice      int         `json:"unit_price"` // Price per unit
	ItemTotal      int         `json:"item_total"` // UnitPrice * Quantity
	Collectible    Collectible `json:"collectible"`
	Store          Store       `json:"store"`
	AvailableUnits int64       `json:"available_units"` // Current stock
}

// CartResponse includes the full cart with calculated totals
type CartResponse struct {
	ID         uint               `json:"id"`
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"` // Sum of all quantities
	TotalPrice int                `json:"total_price"` // Sum of all item totals
	ItemCount  int                `json:"item_count"`  // Number of distinct items
}

// AddToCartRequest is the request body for adding items to cart
type AddToCartRequest struct {
	CollectibleID uint `json:"collectible_id" binding:"required"`
	StoreID       uint `json:"store_id" binding:"required"`
	Quantity      int  `json:"quantity" binding:"required,min=1"`
	Days          int  `json:"days" binding:"required,min=1"`
}

// UpdateCartItemRequest is the request body for updating a cart item
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
	Days     int `json:"days" binding:"required,min=1"`
}
