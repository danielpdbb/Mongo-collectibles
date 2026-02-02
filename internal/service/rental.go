package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
)

// ========================================
// RENTAL CREATION WITH ALLOCATION
// ========================================

// CreateRentalRequest contains all data needed to create a rental
type CreateRentalRequest struct {
	UserID        uint
	CollectibleID uint
	StoreID       uint
	Quantity      int
	Days          int
}

// CreateRentalResponse contains the result of creating a rental
type CreateRentalResponse struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Rental    *domain.Rental         `json:"rental,omitempty"`
	Allocated []domain.AllocatedUnit `json:"allocated_units,omitempty"`
}

// CreateRental creates a new rental with automatic unit allocation
// This is the main function that implements the allocation algorithm
func CreateRental(req CreateRentalRequest) (*CreateRentalResponse, error) {
	// ========================================
	// STEP 1: Validate the collectible exists
	// ========================================
	var collectible domain.Collectible
	if err := repository.DB.First(&collectible, req.CollectibleID).Error; err != nil {
		return &CreateRentalResponse{
			Success: false,
			Message: "Collectible not found",
		}, nil
	}

	// ========================================
	// STEP 2: Check total available units
	// ========================================
	totalAvailable := CountAvailableUnits(req.CollectibleID)
	if int64(req.Quantity) > totalAvailable {
		return &CreateRentalResponse{
			Success: false,
			Message: fmt.Sprintf("Not enough stock available. Only %d units available.", totalAvailable),
		}, nil
	}

	// ========================================
	// STEP 3: Allocate units using nearest-warehouse algorithm
	// ========================================
	allocation := AllocateUnitsForRental(req.CollectibleID, req.StoreID, req.Quantity)

	if !allocation.Success {
		return &CreateRentalResponse{
			Success: false,
			Message: allocation.Message,
		}, nil
	}

	// ========================================
	// STEP 4: Calculate pricing
	// ========================================
	unitPrice := CalculateRentalPrice(collectible.Size, req.Days)
	totalPrice := unitPrice * req.Quantity

	// ========================================
	// STEP 5: Create the rental record
	// ========================================
	rental := &domain.Rental{
		UserID:        req.UserID,
		CollectibleID: req.CollectibleID,
		StoreID:       req.StoreID,
		Quantity:      req.Quantity,
		Days:          req.Days,
		UnitPrice:     unitPrice,
		TotalPrice:    totalPrice,
		Status:        "active",
		StartDate:     time.Now(),
		EndDate:       time.Now().AddDate(0, 0, req.Days),
	}

	if err := repository.DB.Create(rental).Error; err != nil {
		// Rollback: Mark units as available again
		for _, unit := range allocation.AllocatedUnits {
			repository.DB.Model(&domain.CollectibleUnit{}).
				Where("id = ?", unit.UnitID).
				Update("is_available", true)
		}
		return nil, errors.New("failed to create rental")
	}

	// ========================================
	// STEP 6: Create rental unit records
	// ========================================
	for _, unit := range allocation.AllocatedUnits {
		rentalUnit := &domain.RentalUnit{
			RentalID:          rental.ID,
			CollectibleUnitID: unit.UnitID,
			WarehouseID:       unit.WarehouseID,
		}
		repository.DB.Create(rentalUnit)
	}

	return &CreateRentalResponse{
		Success:   true,
		Message:   "Rental created successfully",
		Rental:    rental,
		Allocated: allocation.AllocatedUnits,
	}, nil
}

// ========================================
// RENTAL QUERIES
// ========================================

// GetUserRentals retrieves all rentals for a specific user
func GetUserRentals(userID uint) ([]domain.Rental, error) {
	var rentals []domain.Rental
	err := repository.DB.
		Preload("Collectible").
		Preload("Store").
		Preload("RentalUnits").
		Preload("RentalUnits.Warehouse").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&rentals).Error
	return rentals, err
}

// GetRentalByID retrieves a single rental with all details
func GetRentalByID(rentalID uint) (*domain.Rental, error) {
	var rental domain.Rental
	err := repository.DB.
		Preload("Collectible").
		Preload("Store").
		Preload("User").
		Preload("RentalUnits").
		Preload("RentalUnits.Warehouse").
		First(&rental, rentalID).Error
	if err != nil {
		return nil, err
	}
	return &rental, nil
}

// CompleteRental marks a rental as completed and returns units to available
func CompleteRental(rentalID uint) error {
	var rental domain.Rental
	if err := repository.DB.Preload("RentalUnits").First(&rental, rentalID).Error; err != nil {
		return err
	}

	// Mark all units as available again
	for _, ru := range rental.RentalUnits {
		repository.DB.Model(&domain.CollectibleUnit{}).
			Where("id = ?", ru.CollectibleUnitID).
			Update("is_available", true)
	}

	// Update rental status
	return repository.DB.Model(&rental).Update("status", "completed").Error
}

// CancelRental cancels a rental and returns units to available
func CancelRental(rentalID uint) error {
	var rental domain.Rental
	if err := repository.DB.Preload("RentalUnits").First(&rental, rentalID).Error; err != nil {
		return err
	}

	// Mark all units as available again
	for _, ru := range rental.RentalUnits {
		repository.DB.Model(&domain.CollectibleUnit{}).
			Where("id = ?", ru.CollectibleUnitID).
			Update("is_available", true)
	}

	// Update rental status
	return repository.DB.Model(&rental).Update("status", "cancelled").Error
}
