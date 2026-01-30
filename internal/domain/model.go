package domain

type Store struct {
	ID   int
	Name string
}

type Warehouse struct {
	ID   int
	Name string
}

type Collectible struct {
	ID   int
	Name string
	Size string // S, M, L
}

type CollectibleUnit struct {
	ID            int
	CollectibleID int
	WarehouseID   int
	IsAvailable   bool
}

type WarehouseDistance struct {
	WarehouseID int
	StoreID     int
	Distance    int
}

type Customer struct {
	ID      int
	Name    string
	Email   string
	Phone   string
	Address string
}

type Rental struct {
	ID                int
	CollectibleUnitID int
	StoreID           int
	CustomerID        int
	RentalDays        int
	TotalPrice        int
	Status            string // PENDING, PAID
}

type Payment struct {
	ID        int
	RentalID  int
	Amount    int
	Method    string
	Status    string
	Reference string
}
