package domain

import "time"

// Rental represents a customer's rental order
type Rental struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	CollectibleID uint      `gorm:"not null" json:"collectible_id"`
	StoreID       uint      `gorm:"not null" json:"store_id"`
	Quantity      int       `gorm:"not null" json:"quantity"`
	Days          int       `gorm:"not null" json:"days"`
	UnitPrice     int       `gorm:"not null" json:"unit_price"`
	TotalPrice    int       `gorm:"not null" json:"total_price"`
	Status        string    `gorm:"not null;default:'active'" json:"status"` // active, completed, cancelled
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships (for eager loading)
	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Collectible Collectible  `gorm:"foreignKey:CollectibleID" json:"collectible,omitempty"`
	Store       Store        `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	RentalUnits []RentalUnit `gorm:"foreignKey:RentalID" json:"rental_units,omitempty"`
}

// RentalUnit tracks which specific collectible units are allocated to a rental
type RentalUnit struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	RentalID          uint      `gorm:"not null" json:"rental_id"`
	CollectibleUnitID uint      `gorm:"not null" json:"collectible_unit_id"`
	WarehouseID       uint      `gorm:"not null" json:"warehouse_id"`
	CreatedAt         time.Time `json:"created_at"`

	// Relationships
	Warehouse       Warehouse       `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	CollectibleUnit CollectibleUnit `gorm:"foreignKey:CollectibleUnitID" json:"collectible_unit,omitempty"`
}

// AllocationResult represents the result of allocating units for a rental
type AllocationResult struct {
	Success        bool            `json:"success"`
	Message        string          `json:"message,omitempty"`
	AllocatedUnits []AllocatedUnit `json:"allocated_units,omitempty"`
	TotalAllocated int             `json:"total_allocated"`
	RequestedQty   int             `json:"requested_quantity"`
}

// AllocatedUnit represents a single unit allocation with its source warehouse
type AllocatedUnit struct {
	UnitID        uint   `json:"unit_id"`
	WarehouseID   uint   `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`
}
