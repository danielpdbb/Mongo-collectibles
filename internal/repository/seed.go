package repository

import "github.com/danielpdbb/Mongo-collectibles/internal/domain"

func SeedData() {
	// Stores
	stores := []domain.Store{
		{Name: "Store A - Manila"},
		{Name: "Store B - Makati"},
		{Name: "Store C - Quezon City"},
	}
	DB.Create(&stores)

	// Warehouses
	warehouses := []domain.Warehouse{
		{Name: "Warehouse North - Caloocan"},
		{Name: "Warehouse Central - Pasig"},
		{Name: "Warehouse South - Paranaque"},
	}
	DB.Create(&warehouses)

	// MCU Collectibles
	collectibles := []domain.Collectible{
		// Large Size Collectibles
		{
			Name:     "Iron Man Mark LXXXV Life-Size Statue",
			Size:     "L",
			ImageURL: "https://images.unsplash.com/photo-1635863138275-d9b33299680b?w=800",
		},
		{
			Name:     "Thanos Infinity Gauntlet Life-Size Replica",
			Size:     "L",
			ImageURL: "https://images.unsplash.com/photo-1612404730960-5c71577fca11?w=800",
		},
		{
			Name:     "Captain America Shield Full-Scale Replica",
			Size:     "L",
			ImageURL: "https://images.unsplash.com/photo-1569003339405-ea396a5a8a90?w=800",
		},
		{
			Name:     "Thor Stormbreaker Life-Size Replica",
			Size:     "L",
			ImageURL: "https://images.unsplash.com/photo-1655720828018-edd2daec9349?w=800",
		},
		// Medium Size Collectibles
		{
			Name:     "Spider-Man Advanced Suit 1:4 Scale",
			Size:     "M",
			ImageURL: "https://images.unsplash.com/photo-1635805737707-575885ab0820?w=800",
		},
		{
			Name:     "Black Panther 1:4 Scale Statue",
			Size:     "M",
			ImageURL: "https://images.unsplash.com/photo-1559535332-db9971090158?w=800",
		},
		{
			Name:     "Doctor Strange 1:4 Scale Figure",
			Size:     "M",
			ImageURL: "https://images.unsplash.com/photo-1624213111452-35e8d3d5cc18?w=800",
		},
		{
			Name:     "Hulk Smash 1:4 Scale Statue",
			Size:     "M",
			ImageURL: "https://images.unsplash.com/photo-1608889825103-eb5ed706fc64?w=800",
		},
		{
			Name:     "Scarlet Witch 1:4 Scale Figure",
			Size:     "M",
			ImageURL: "https://images.unsplash.com/photo-1611604548018-d56bbd85d681?w=800",
		},
		{
			Name:     "Vision Premium Format Statue",
			Size:     "M",
			ImageURL: "https://images.unsplash.com/photo-1620336655052-b57986f5a26a?w=800",
		},
		// Small Size Collectibles
		{
			Name:     "Iron Man Mini Arc Reactor Replica",
			Size:     "S",
			ImageURL: "https://images.unsplash.com/photo-1509347528160-9a9e33742cdb?w=800",
		},
		{
			Name:     "Loki Scepter Mind Stone Replica",
			Size:     "S",
			ImageURL: "https://images.unsplash.com/photo-1608889175123-8ee362201f81?w=800",
		},
		{
			Name:     "Ant-Man Helmet 1:1 Wearable Replica",
			Size:     "S",
			ImageURL: "https://images.unsplash.com/photo-1608889476561-6242cfdbf622?w=800",
		},
		{
			Name:     "Captain Marvel Photon Blaster Replica",
			Size:     "S",
			ImageURL: "https://images.unsplash.com/photo-1531259683007-016a7b628fc3?w=800",
		},
		{
			Name:     "Infinity Stones Complete Set",
			Size:     "S",
			ImageURL: "https://images.unsplash.com/photo-1614680376408-81e91ffe3db7?w=800",
		},
		{
			Name:     "Baby Groot Dancing Figure",
			Size:     "S",
			ImageURL: "https://images.unsplash.com/photo-1636051028886-0059ad2383c8?w=800",
		},
	}
	DB.Create(&collectibles)

	// Units (stock) - Create multiple units per collectible across warehouses
	for _, c := range collectibles {
		// 2-3 units per warehouse for better availability
		for i := 0; i < 2; i++ {
			DB.Create(&domain.CollectibleUnit{
				CollectibleID: c.ID,
				WarehouseID:   warehouses[0].ID,
				IsAvailable:   true,
			})
		}
		DB.Create(&domain.CollectibleUnit{
			CollectibleID: c.ID,
			WarehouseID:   warehouses[1].ID,
			IsAvailable:   true,
		})
		DB.Create(&domain.CollectibleUnit{
			CollectibleID: c.ID,
			WarehouseID:   warehouses[2].ID,
			IsAvailable:   true,
		})
	}

	// Distances from warehouses to stores (in km)
	// Warehouse North
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[0].ID, StoreID: stores[0].ID, Distance: 5})  // to Manila
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[0].ID, StoreID: stores[1].ID, Distance: 15}) // to Makati
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[0].ID, StoreID: stores[2].ID, Distance: 10}) // to QC

	// Warehouse Central
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[1].ID, StoreID: stores[0].ID, Distance: 8})  // to Manila
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[1].ID, StoreID: stores[1].ID, Distance: 5})  // to Makati
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[1].ID, StoreID: stores[2].ID, Distance: 12}) // to QC

	// Warehouse South
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[2].ID, StoreID: stores[0].ID, Distance: 12}) // to Manila
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[2].ID, StoreID: stores[1].ID, Distance: 8})  // to Makati
	DB.Create(&domain.WarehouseDistance{WarehouseID: warehouses[2].ID, StoreID: stores[2].ID, Distance: 20}) // to QC
}
