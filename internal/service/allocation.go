package service

import "github.com/danielpdbb/Mongo-collectibles/internal/domain"

func AllocateNearestUnit(
	units []domain.CollectibleUnit,
	distances []domain.WarehouseDistance,
	storeID int,
) *domain.CollectibleUnit {

	minDistance := int(^uint(0) >> 1)
	var selected *domain.CollectibleUnit

	for _, unit := range units {
		if !unit.IsAvailable {
			continue
		}

		for _, d := range distances {
			if d.WarehouseID == unit.WarehouseID && d.StoreID == storeID {
				if d.Distance < minDistance {
					minDistance = d.Distance
					u := unit
					selected = &u
				}
			}
		}
	}

	return selected
}
