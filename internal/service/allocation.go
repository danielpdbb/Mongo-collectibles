package service

import (
	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
)

// FindNearestAvailableUnit finds the nearest available unit WITHOUT marking it unavailable
// Use this for quotes/availability checks
func FindNearestAvailableUnit(collectibleID uint, storeID uint) (domain.CollectibleUnit, bool) {
	var units []domain.CollectibleUnit
	repository.DB.Where("collectible_id = ? AND is_available = true", collectibleID).Find(&units)

	if len(units) == 0 {
		return domain.CollectibleUnit{}, false
	}

	var distances []domain.WarehouseDistance
	repository.DB.Where("store_id = ?", storeID).Find(&distances)

	var selected *domain.CollectibleUnit
	min := int(^uint(0) >> 1)

	for _, unit := range units {
		for _, d := range distances {
			if d.WarehouseID == unit.WarehouseID && d.Distance < min {
				min = d.Distance
				u := unit
				selected = &u
			}
		}
	}

	if selected == nil {
		return domain.CollectibleUnit{}, false
	}

	return *selected, true
}

func CountAvailableUnits(collectibleID uint) int64 {
	var count int64
	repository.DB.Model(&domain.CollectibleUnit{}).
		Where("collectible_id = ? AND is_available = ? ", collectibleID, true).
		Count(&count)
	return count
}

// ========================================
// ALLOCATION ALGORITHM
// ========================================

// AllocateUnitsForRental implements the nearest-warehouse-first allocation algorithm
//
// Algorithm:
// 1. Get all warehouses sorted by distance from selected store (ascending)
// 2. For each warehouse (nearest first):
//   - Get available units for this collectible
//   - Allocate as many as needed
//   - Mark them as unavailable
//
// 3. Continue until requested quantity is fulfilled or no more units
// 4. If insufficient units, rollback and return error
func AllocateUnitsForRental(collectibleID, storeID uint, quantity int) domain.AllocationResult {
	var allocatedUnits []domain.AllocatedUnit
	remaining := quantity

	// ========================================
	// Get warehouses sorted by distance from store
	// ========================================
	type WarehouseWithDistance struct {
		WarehouseID   uint
		WarehouseName string
		Distance      int
	}

	var warehouses []WarehouseWithDistance
	repository.DB.
		Table("warehouse_distances").
		Select("warehouse_distances.warehouse_id, warehouses.name as warehouse_name, warehouse_distances.distance").
		Joins("JOIN warehouses ON warehouse_distances.warehouse_id = warehouses.id").
		Where("warehouse_distances.store_id = ?", storeID).
		Where("warehouses.deleted_at IS NULL").
		Order("warehouse_distances.distance ASC"). // Nearest first!
		Scan(&warehouses)

	if len(warehouses) == 0 {
		return domain.AllocationResult{
			Success:      false,
			Message:      "No warehouses configured for this store",
			RequestedQty: quantity,
		}
	}

	// ========================================
	// Allocate from each warehouse in distance order
	// ========================================
	for _, wh := range warehouses {
		if remaining <= 0 {
			break // Fully allocated!
		}

		// Get available units from this warehouse
		var availableUnits []domain.CollectibleUnit
		repository.DB.
			Where("collectible_id = ?", collectibleID).
			Where("warehouse_id = ?", wh.WarehouseID).
			Where("is_available = ?", true).
			Find(&availableUnits)

		// Allocate units from this warehouse
		for _, unit := range availableUnits {
			if remaining <= 0 {
				break
			}

			// Mark unit as unavailable (reserve it)
			repository.DB.Model(&unit).Update("is_available", false)

			// Add to allocated list
			allocatedUnits = append(allocatedUnits, domain.AllocatedUnit{
				UnitID:        unit.ID,
				WarehouseID:   wh.WarehouseID,
				WarehouseName: wh.WarehouseName,
			})

			remaining--
		}
	}

	// ========================================
	// Check if we fully allocated
	// ========================================
	if remaining > 0 {
		// Rollback: Mark all allocated units as available again
		for _, unit := range allocatedUnits {
			repository.DB.Model(&domain.CollectibleUnit{}).
				Where("id = ?", unit.UnitID).
				Update("is_available", true)
		}

		return domain.AllocationResult{
			Success:        false,
			Message:        "Not enough stock available",
			TotalAllocated: len(allocatedUnits),
			RequestedQty:   quantity,
		}
	}

	return domain.AllocationResult{
		Success:        true,
		AllocatedUnits: allocatedUnits,
		TotalAllocated: len(allocatedUnits),
		RequestedQty:   quantity,
	}
}
