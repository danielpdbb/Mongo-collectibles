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
		// Large Size Collectibles (4)
		{
			Name:     "Iron Man Mark LXXXV Life-Size Statue",
			Size:     "L",
			ImageURL: "https://www.sideshow.com/cdn-cgi/image/quality=90,f=auto/https://www.sideshow.com/storage/product-images/300281/iron-man-mark-vii_marvel_gallery_5f7e14d137814.jpg",
		},
		{
			Name:     "Thanos Infinity Gauntlet Life-Size Replica",
			Size:     "L",
			ImageURL: "https://i.ebayimg.com/images/g/BFgAAOSwLNxcCIvR/s-l1200.jpg",
		},
		{
			Name:     "Captain America Shield Full-Scale Replica",
			Size:     "L",
			ImageURL: "https://www.shutterstock.com/image-photo/captain-america-shield-hyper-realistic-600nw-2674561019.jpg",
		},
		{
			Name:     "Thor Mjolnir & Stormbreaker Set",
			Size:     "L",
			ImageURL: "https://static0.cbrimages.com/wordpress/wp-content/uploads/2022/04/Thor-Mjolnir.jpg?w=1200&h=675&fit=crop",
		},
		// Medium Size Collectibles (4)
		{
			Name:     "Spider-Man Advanced Suit 1:4 Scale",
			Size:     "M",
			ImageURL: "https://www.sideshow.com/wp/wp-content/uploads/2020/01/Spider-Man-Advanced-Suit-Hot-Toys-6.jpg",
		},
		{
			Name:     "Black Panther 1:4 Scale Statue",
			Size:     "M",
			ImageURL: "https://handsomecake.com/cdn/shop/files/301104743_2271086519733765_182266933800230682_n.jpg?v=1731363607",
		},
		{
			Name:     "Hulk Smash 1:4 Scale Statue",
			Size:     "M",
			ImageURL: "https://cdna.artstation.com/p/assets/images/images/050/635/174/large/adam-meah-hulk-smash-2.jpg?1655309388",
		},
		{
			Name:     "Doctor Strange 1:4 Scale Figure",
			Size:     "M",
			ImageURL: "https://www.sideshow.com/cdn-cgi/image/height=850,quality=90,f=auto/https://www.sideshow.com/storage/product-images/300662/doctor-strange_marvel_gallery_5faf3fe00cc14.jpg",
		},
		// Small Size Collectibles (4)
		{
			Name:     "Iron Man Arc Reactor Replica",
			Size:     "S",
			ImageURL: "https://anotoys.com/cdn/shop/products/image_720x_30f385bd-df3c-4c45-8fa7-d505019e4d64_383x@3x.progressive.jpg?v=1691385529",
		},
		{
			Name:     "Infinity Stones Complete Set",
			Size:     "S",
			ImageURL: "https://i.etsystatic.com/19286482/r/il/c6a59d/3293906755/il_fullxfull.3293906755_6h54.jpg",
		},
		{
			Name:     "Baby Groot Dancing Figure",
			Size:     "S",
			ImageURL: "https://m.media-amazon.com/images/I/71nCPIXlotL._AC_UF894,1000_QL80_.jpg",
		},
		{
			Name:     "Loki Scepter Mind Stone Replica",
			Size:     "S",
			ImageURL: "https://static.wikia.nocookie.net/marvelcinematicuniverse/images/1/17/Scepter_Main.jpg/revision/latest?cb=20150806163618",
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
