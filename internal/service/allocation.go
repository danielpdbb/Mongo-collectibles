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

// AllocateNearestAvailableUnit finds AND reserves the nearest available unit
// Use this only when actually creating a rental
func AllocateNearestAvailableUnit(collectibleID uint, storeID uint) (domain.CollectibleUnit, bool) {
	unit, ok := FindNearestAvailableUnit(collectibleID, storeID)
	if !ok {
		return domain.CollectibleUnit{}, false
	}

	// Mark the unit as unavailable immediately
	unit.IsAvailable = false
	repository.DB.Save(&unit)

	return unit, true
}

func CountAvailableUnits(collectibleID uint) int64 {
	var count int64
	repository.DB.Model(&domain.CollectibleUnit{}).
		Where("collectible_id = ? AND is_available = ? ", collectibleID, true).
		Count(&count)
	return count
}
