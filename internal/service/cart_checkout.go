package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
)

// ========================================
// CART CHECKOUT - CREATE RENTALS FROM CART
// ========================================

// CartCheckoutResponse contains the result of checking out a cart
type CartCheckoutResponse struct {
	Success    bool                 `json:"success"`
	Message    string               `json:"message,omitempty"`
	Rentals    []CartCheckoutRental `json:"rentals,omitempty"`
	TotalPrice int                  `json:"total_price"`
}

// CartCheckoutRental contains rental info for a single cart item
type CartCheckoutRental struct {
	RentalID        uint                   `json:"rental_id"`
	CollectibleID   uint                   `json:"collectible_id"`
	CollectibleName string                 `json:"collectible_name"`
	StoreID         uint                   `json:"store_id"`
	StoreName       string                 `json:"store_name"`
	Quantity        int                    `json:"quantity"`
	Days            int                    `json:"days"`
	UnitPrice       int                    `json:"unit_price"`
	TotalPrice      int                    `json:"total_price"`
	AllocatedUnits  []domain.AllocatedUnit `json:"allocated_units"`
}

// CheckoutCart creates rentals for all items in the user's cart
// This is an atomic operation - if any item fails, all are rolled back
func CheckoutCart(userID uint) (*CartCheckoutResponse, error) {
	// Get cart with items
	cart, err := GetCartWithItems(userID)
	if err != nil {
		return &CartCheckoutResponse{
			Success: false,
			Message: "Failed to retrieve cart",
		}, err
	}

	if len(cart.Items) == 0 {
		return &CartCheckoutResponse{
			Success: false,
			Message: "Cart is empty",
		}, nil
	}

	// ========================================
	// STEP 1: Validate all items have sufficient stock
	// ========================================
	for _, item := range cart.Items {
		availableUnits := CountAvailableUnits(item.CollectibleID)
		if int64(item.Quantity) > availableUnits {
			return &CartCheckoutResponse{
				Success: false,
				Message: fmt.Sprintf("Not enough stock for %s. Only %d available.",
					item.Collectible.Name, availableUnits),
			}, nil
		}
	}

	// ========================================
	// STEP 2: Create rentals for each cart item
	// ========================================
	var createdRentals []CartCheckoutRental
	var createdRentalIDs []uint // Track for potential rollback
	totalPrice := 0

	for _, item := range cart.Items {
		// Create rental for this cart item
		result, err := CreateRental(CreateRentalRequest{
			UserID:        userID,
			CollectibleID: item.CollectibleID,
			StoreID:       item.StoreID,
			Quantity:      item.Quantity,
			Days:          item.Days,
		})

		if err != nil || !result.Success {
			// Rollback: Cancel all previously created rentals
			for _, rentalID := range createdRentalIDs {
				CancelRental(rentalID)
			}

			errMsg := "Failed to create rental"
			if result != nil && result.Message != "" {
				errMsg = result.Message
			}

			return &CartCheckoutResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to process %s: %s", item.Collectible.Name, errMsg),
			}, nil
		}

		createdRentalIDs = append(createdRentalIDs, result.Rental.ID)

		// Get store name
		var store domain.Store
		repository.DB.First(&store, item.StoreID)

		createdRentals = append(createdRentals, CartCheckoutRental{
			RentalID:        result.Rental.ID,
			CollectibleID:   item.CollectibleID,
			CollectibleName: item.Collectible.Name,
			StoreID:         item.StoreID,
			StoreName:       store.Name,
			Quantity:        item.Quantity,
			Days:            item.Days,
			UnitPrice:       result.Rental.UnitPrice,
			TotalPrice:      result.Rental.TotalPrice,
			AllocatedUnits:  result.Allocated,
		})

		totalPrice += result.Rental.TotalPrice
	}

	// ========================================
	// STEP 3: Clear the cart after successful checkout
	// ========================================
	ClearCart(userID)

	return &CartCheckoutResponse{
		Success:    true,
		Message:    fmt.Sprintf("Successfully created %d rental(s)", len(createdRentals)),
		Rentals:    createdRentals,
		TotalPrice: totalPrice,
	}, nil
}

// CreateBatchRental creates a single rental that combines multiple cart items
// All items must go to the same store. Uses a "batch rental" approach.
type BatchRentalRequest struct {
	UserID  uint
	StoreID uint
	Items   []BatchRentalItem
}

type BatchRentalItem struct {
	CollectibleID uint
	Quantity      int
	Days          int
}

// BatchRental represents a rental with multiple collectible types
type BatchRental struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null" json:"user_id"`
	StoreID    uint       `gorm:"not null" json:"store_id"`
	TotalPrice int        `gorm:"not null" json:"total_price"`
	Status     string     `gorm:"not null;default:'pending_payment'" json:"status"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ValidateCartForCheckout validates all cart items before checkout
func ValidateCartForCheckout(userID uint) (bool, string, error) {
	cart, err := GetCartWithItems(userID)
	if err != nil {
		return false, "Failed to retrieve cart", err
	}

	if len(cart.Items) == 0 {
		return false, "Cart is empty", nil
	}

	// Check stock for each item
	for _, item := range cart.Items {
		availableUnits := CountAvailableUnits(item.CollectibleID)
		if int64(item.Quantity) > availableUnits {
			return false, fmt.Sprintf("Not enough stock for %s. Requested: %d, Available: %d",
				item.Collectible.Name, item.Quantity, availableUnits), nil
		}
	}

	return true, "", nil
}

// GetCartSummaryForPayment returns cart summary with rental IDs after checkout
// Used to redirect to payment page
func GetCartCheckoutSummary(userID uint, rentalIDs []uint) (*CartCheckoutResponse, error) {
	var rentals []CartCheckoutRental
	totalPrice := 0

	for _, rentalID := range rentalIDs {
		rental, err := GetRentalByID(rentalID)
		if err != nil {
			continue
		}

		// Get warehouse info from rental units
		var allocatedUnits []domain.AllocatedUnit
		for _, ru := range rental.RentalUnits {
			allocatedUnits = append(allocatedUnits, domain.AllocatedUnit{
				UnitID:        ru.CollectibleUnitID,
				WarehouseID:   ru.WarehouseID,
				WarehouseName: ru.Warehouse.Name,
			})
		}

		rentals = append(rentals, CartCheckoutRental{
			RentalID:        rental.ID,
			CollectibleID:   rental.CollectibleID,
			CollectibleName: rental.Collectible.Name,
			StoreID:         rental.StoreID,
			StoreName:       rental.Store.Name,
			Quantity:        rental.Quantity,
			Days:            rental.Days,
			UnitPrice:       rental.UnitPrice,
			TotalPrice:      rental.TotalPrice,
			AllocatedUnits:  allocatedUnits,
		})

		totalPrice += rental.TotalPrice
	}

	return &CartCheckoutResponse{
		Success:    true,
		Rentals:    rentals,
		TotalPrice: totalPrice,
	}, nil
}

// ActivateMultipleRentals activates all rentals in a list
func ActivateMultipleRentals(rentalIDs []uint) error {
	for _, id := range rentalIDs {
		if err := ActivateRental(id); err != nil {
			return errors.New(fmt.Sprintf("failed to activate rental %d: %v", id, err))
		}
	}
	return nil
}

// ExpireMultipleRentals expires all rentals in a list
func ExpireMultipleRentals(rentalIDs []uint) error {
	for _, id := range rentalIDs {
		if err := ExpireRental(id); err != nil {
			return errors.New(fmt.Sprintf("failed to expire rental %d: %v", id, err))
		}
	}
	return nil
}
