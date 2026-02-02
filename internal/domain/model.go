package domain

import "gorm.io/gorm"

type Store struct {
	gorm.Model
	Name string
}

type Warehouse struct {
	gorm.Model
	Name string
}

type Collectible struct {
	gorm.Model
	Name     string `json:"name"`
	Size     string `json:"size"`
	ImageURL string `json:"imageURL"`
}

type CollectibleUnit struct {
	gorm.Model
	CollectibleID uint
	WarehouseID   uint
	IsAvailable   bool
}

type WarehouseDistance struct {
	gorm.Model
	WarehouseID uint
	StoreID     uint
	Distance    int
}
