# MongoCollectibles System Documentation

## Table of Contents
1. [System Overview](#system-overview)
2. [Complete Process Flow Diagram](#complete-process-flow-diagram)
3. [Database Schema](#database-schema)
4. [Allocation Algorithm](#allocation-algorithm)
5. [Warehouse Distance Visualization](#warehouse-distance-visualization)
6. [Pricing Logic](#pricing-logic)
7. [PayMongo API Integration](#paymongo-api-integration)
8. [Postman Testing Guide](#postman-testing-guide)
9. [Alternative Solutions](#alternative-solutions)

---

## System Overview

**MongoCollectibles** is a web-based rental platform for collectible items. The system handles:
- Product catalog management
- **Automatic unit allocation** based on nearest warehouse
- **Dynamic pricing** based on size and rental duration
- **Online payments** via PayMongo (Cards, GCash, GrabPay, BPI, UBP)
- Rental lifecycle management

### Tech Stack
- **Backend**: Go (Gin Framework)
- **Database**: PostgreSQL with GORM ORM
- **Frontend**: HTML, Tailwind CSS, Alpine.js
- **Payments**: PayMongo API

---

## Complete Process Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         MONGOCOLLECTIBLES RENTAL FLOW                           │
└─────────────────────────────────────────────────────────────────────────────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   CUSTOMER   │     │   FRONTEND   │     │   BACKEND    │     │   PAYMONGO   │
│              │     │   (Web UI)   │     │  (Go/Gin)    │     │     API      │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │                    │
       │  1. Browse Catalog │                    │                    │
       │ ──────────────────>│                    │                    │
       │                    │  GET /catalogue    │                    │
       │                    │ ──────────────────>│                    │
       │                    │   Return items     │                    │
       │                    │ <──────────────────│                    │
       │   Display Items    │                    │                    │
       │ <──────────────────│                    │                    │
       │                    │                    │                    │
       │  2. Register/Login │                    │                    │
       │ ──────────────────>│ POST /api/register │                    │
       │                    │ POST /api/login    │                    │
       │                    │ ──────────────────>│                    │
       │                    │   JWT Token        │                    │
       │                    │ <──────────────────│                    │
       │   Store Token      │                    │                    │
       │ <──────────────────│                    │                    │
       │                    │                    │                    │
       │  3. Select Item    │                    │                    │
       │     + Store        │                    │                    │
       │     + Quantity     │                    │                    │
       │     + Days         │                    │                    │
       │ ──────────────────>│                    │                    │
       │                    │                    │                    │
       │  4. Checkout       │                    │                    │
       │ ──────────────────>│ POST /api/rentals  │                    │
       │                    │ ──────────────────>│                    │
       │                    │                    │                    │
       │                    │    ┌───────────────────────────────┐    │
       │                    │    │  ALLOCATION ALGORITHM         │    │
       │                    │    │  ─────────────────────────    │    │
       │                    │    │  1. Get warehouses sorted     │    │
       │                    │    │     by distance (ASC)         │    │
       │                    │    │  2. For each warehouse:       │    │
       │                    │    │     - Find available units    │    │
       │                    │    │     - Allocate nearest first  │    │
       │                    │    │     - Mark as unavailable     │    │
       │                    │    │  3. Calculate pricing         │    │
       │                    │    │  4. Create rental record      │    │
       │                    │    │  5. Start payment timer       │    │
       │                    │    └───────────────────────────────┘    │
       │                    │                    │                    │
       │                    │  Rental Created    │                    │
       │                    │  (pending_payment) │                    │
       │                    │ <──────────────────│                    │
       │   Show Payment     │                    │                    │
       │ <──────────────────│                    │                    │
       │                    │                    │                    │
       │  5. Enter Billing  │                    │                    │
       │     + Payment      │                    │                    │
       │       Method       │                    │                    │
       │ ──────────────────>│ POST /api/payments │                    │
       │                    │ ──────────────────>│                    │
       │                    │                    │                    │
       │                    │                    │  [CARD FLOW]       │
       │                    │                    │ ─────────────────> │
       │                    │                    │  1. Create Intent  │
       │                    │                    │  2. Create Method  │
       │                    │                    │  3. Attach Method  │
       │                    │                    │ <───────────────── │
       │                    │                    │  Status/3DS URL    │
       │                    │                    │                    │
       │                    │                    │  [E-WALLET FLOW]   │
       │                    │                    │ ─────────────────> │
       │                    │                    │  Create Source     │
       │                    │                    │ <───────────────── │
       │                    │                    │  Checkout URL      │
       │                    │                    │                    │
       │                    │  Checkout URL /    │                    │
       │                    │  3DS URL / Success │                    │
       │                    │ <──────────────────│                    │
       │                    │                    │                    │
       │  6. Complete       │                    │                    │
       │     Payment        │                    │                    │
       │ ──────────────────>│ ────── REDIRECT ───────────────────────>│
       │                    │                    │                    │
       │                    │<─── Callback to success/failed URL ─────│
       │                    │                    │                    │
       │                    │ GET /payments/:id/ │                    │
       │                    │     verify         │                    │
       │                    │ ──────────────────>│                    │
       │                    │                    │ GET payment status │
       │                    │                    │ ──────────────────>│
       │                    │                    │ <──────────────────│
       │                    │                    │                    │
       │                    │    ┌───────────────────────────────┐    │
       │                    │    │  PAYMENT VERIFICATION         │    │
       │                    │    │  ─────────────────────────    │    │
       │                    │    │  IF paid/chargeable:          │    │
       │                    │    │    - Update payment → paid    │    │
       │                    │    │    - Rental → active          │    │
       │                    │    │  IF failed/expired:           │    │
       │                    │    │    - Payment → failed         │    │
       │                    │    │    - Return units to stock    │    │
       │                    │    │    - Rental → cancelled       │    │
       │                    │    └───────────────────────────────┘    │
       │                    │                    │                    │
       │   Show Success/    │  Payment Result    │                    │
       │   Failed Page      │ <──────────────────│                    │
       │ <──────────────────│                    │                    │
       │                    │                    │                    │
       ▼                    ▼                    ▼                    ▼


┌─────────────────────────────────────────────────────────────────────────────────┐
│                              RENTAL STATUS LIFECYCLE                            │
└─────────────────────────────────────────────────────────────────────────────────┘

                    ┌─────────────────┐
                    │  CREATE RENTAL  │
                    └────────┬────────┘
                             │
                             ▼
                  ┌──────────────────────┐
                  │   pending_payment    │◄───────────────────────┐
                  │   (Timer Started)    │                        │
                  └──────────┬───────────┘                        │
                             │                                    │
            ┌────────────────┼────────────────┐                   │
            │                │                │                   │
            ▼                ▼                ▼                   │
     ┌────────────┐   ┌────────────┐   ┌────────────┐            │
     │  PAYMENT   │   │  PAYMENT   │   │   TIMER    │            │
     │  SUCCESS   │   │  FAILED/   │   │  EXPIRED   │            │
     │            │   │  CANCELLED │   │            │            │
     └─────┬──────┘   └─────┬──────┘   └─────┬──────┘            │
           │                │                │                    │
           ▼                ▼                ▼                    │
     ┌────────────┐   ┌────────────┐   ┌────────────┐            │
     │   active   │   │ cancelled  │   │  expired   │            │
     │  (Rental   │   │  (Units    │   │  (Units    │            │
     │  Confirmed)│   │  Returned) │   │  Returned) │            │
     └─────┬──────┘   └────────────┘   └────────────┘            │
           │                                                      │
           │  (After rental period ends)                          │
           ▼                                                      │
     ┌────────────┐                                               │
     │ completed  │                                               │
     │  (Units    │───────────────────────────────────────────────┘
     │  Returned) │        (Available for next rental)
     └────────────┘
```

---

## Database Schema

### Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              DATABASE SCHEMA                                     │
└─────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────┐          ┌──────────────────┐          ┌──────────────────┐
│      USERS       │          │     RENTALS      │          │   COLLECTIBLES   │
├──────────────────┤          ├──────────────────┤          ├──────────────────┤
│ id (PK)          │──────┐   │ id (PK)          │   ┌──────│ id (PK)          │
│ email (unique)   │      └──>│ user_id (FK)     │   │      │ name             │
│ password         │          │ collectible_id(FK)│<──┘      │ size (S/M/L)     │
│ name             │          │ store_id (FK)    │──┐       │ image_url        │
│ phone            │          │ quantity         │  │       │ created_at       │
│ created_at       │          │ days             │  │       │ updated_at       │
│ updated_at       │          │ unit_price       │  │       └──────────────────┘
└──────────────────┘          │ total_price      │  │               │
                              │ status           │  │               │
                              │ expires_at       │  │               │ 1
                              │ start_date       │  │               │
                              │ end_date         │  │               ▼ *
                              │ created_at       │  │       ┌──────────────────┐
                              │ updated_at       │  │       │COLLECTIBLE_UNITS │
                              └────────┬─────────┘  │       ├──────────────────┤
                                       │            │       │ id (PK)          │
                                       │ 1          │       │ collectible_id(FK)│
                                       │            │       │ warehouse_id (FK)│───┐
                                       ▼ *          │       │ is_available     │   │
                              ┌──────────────────┐  │       │ created_at       │   │
                              │   RENTAL_UNITS   │  │       │ updated_at       │   │
                              ├──────────────────┤  │       └──────────────────┘   │
                              │ id (PK)          │  │                              │
                              │ rental_id (FK)   │  │       ┌──────────────────┐   │
                              │ collectible_unit │  │       │    WAREHOUSES    │   │
                              │   _id (FK)       │  │       ├──────────────────┤   │
                              │ warehouse_id(FK) │──┼──────>│ id (PK)          │<──┘
                              │ created_at       │  │       │ name             │
                              └──────────────────┘  │       │ created_at       │
                                                    │       │ updated_at       │
┌──────────────────┐                                │       └────────┬─────────┘
│      STORES      │                                │                │
├──────────────────┤                                │                │ 1
│ id (PK)          │<───────────────────────────────┘                │
│ name             │                                                 ▼ *
│ created_at       │◄────────────────────────────────┐    ┌──────────────────┐
│ updated_at       │                                 │    │WAREHOUSE_DISTANCE│
└──────────────────┘                                 │    ├──────────────────┤
                                                     │    │ id (PK)          │
                                                     │    │ warehouse_id (FK)│
                                                     └────│ store_id (FK)    │
                                                          │ distance (km)    │
                                                          │ created_at       │
                                                          │ updated_at       │
                                                          └──────────────────┘

┌──────────────────┐          ┌──────────────────┐
│     PAYMENTS     │          │ BILLING_DETAILS  │
├──────────────────┤          ├──────────────────┤
│ id (PK)          │<─────────│ payment_id (FK)  │
│ rental_id (FK)   │          │ name             │
│ amount (centavos)│          │ email            │
│ currency         │          │ phone            │
│ status           │          │ address_line     │
│ payment_method   │          │ city             │
│ paymongo_id      │          │ state            │
│ payment_intent_id│          │ postal_code      │
│ source_id        │          │ country          │
│ checkout_url     │          └──────────────────┘
│ paid_at          │
│ created_at       │
│ updated_at       │
└──────────────────┘
```

### Table Definitions

```sql
-- USERS: Customer accounts
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,  -- bcrypt hashed
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- STORES: Brick-and-mortar pickup locations (3+)
CREATE TABLE stores (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- WAREHOUSES: Storage locations for collectibles
CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- COLLECTIBLES: Product catalog
CREATE TABLE collectibles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    size CHAR(1) CHECK (size IN ('S', 'M', 'L')),
    image_url TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- COLLECTIBLE_UNITS: Physical inventory (each row = 1 physical unit)
CREATE TABLE collectible_units (
    id SERIAL PRIMARY KEY,
    collectible_id INTEGER REFERENCES collectibles(id),
    warehouse_id INTEGER REFERENCES warehouses(id),
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- WAREHOUSE_DISTANCES: Distance matrix for allocation algorithm
CREATE TABLE warehouse_distances (
    id SERIAL PRIMARY KEY,
    warehouse_id INTEGER REFERENCES warehouses(id),
    store_id INTEGER REFERENCES stores(id),
    distance INTEGER NOT NULL,  -- in kilometers
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(warehouse_id, store_id)
);

-- RENTALS: Customer orders
CREATE TABLE rentals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    collectible_id INTEGER REFERENCES collectibles(id),
    store_id INTEGER REFERENCES stores(id),
    quantity INTEGER NOT NULL,
    days INTEGER NOT NULL,
    unit_price INTEGER NOT NULL,      -- in PHP
    total_price INTEGER NOT NULL,     -- in PHP
    status VARCHAR(50) DEFAULT 'pending_payment',
    expires_at TIMESTAMP,             -- payment deadline
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- RENTAL_UNITS: Which specific units are allocated to a rental
CREATE TABLE rental_units (
    id SERIAL PRIMARY KEY,
    rental_id INTEGER REFERENCES rentals(id),
    collectible_unit_id INTEGER REFERENCES collectible_units(id),
    warehouse_id INTEGER REFERENCES warehouses(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- PAYMENTS: PayMongo payment records
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    rental_id INTEGER REFERENCES rentals(id),
    amount INTEGER NOT NULL,          -- in centavos
    currency VARCHAR(3) DEFAULT 'PHP',
    status VARCHAR(50) DEFAULT 'pending',
    payment_method VARCHAR(50),
    paymongo_id VARCHAR(255),
    payment_intent_id VARCHAR(255),
    source_id VARCHAR(255),
    checkout_url TEXT,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- BILLING_DETAILS: Customer billing information
CREATE TABLE billing_details (
    id SERIAL PRIMARY KEY,
    payment_id INTEGER REFERENCES payments(id),
    name VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    address_line TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(2) DEFAULT 'PH',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

---

## Allocation Algorithm

### Overview

The **Nearest-Warehouse-First Allocation Algorithm** ensures efficient logistics by allocating collectible units from the warehouse closest to the customer's selected pickup store.

### Algorithm Pseudocode

```
FUNCTION AllocateUnitsForRental(collectibleID, storeID, quantity):
    
    allocatedUnits = []
    remaining = quantity
    
    // STEP 1: Get warehouses sorted by distance from store (ascending)
    warehouses = SELECT warehouse_id, distance 
                 FROM warehouse_distances 
                 WHERE store_id = storeID 
                 ORDER BY distance ASC
    
    // STEP 2: Allocate from each warehouse in distance order
    FOR EACH warehouse IN warehouses:
        IF remaining <= 0:
            BREAK  // Fully allocated!
        
        // Get available units from this warehouse
        availableUnits = SELECT * FROM collectible_units 
                         WHERE collectible_id = collectibleID 
                         AND warehouse_id = warehouse.id 
                         AND is_available = TRUE
        
        // Allocate units from this warehouse
        FOR EACH unit IN availableUnits:
            IF remaining <= 0:
                BREAK
            
            // Reserve unit
            UPDATE collectible_units SET is_available = FALSE WHERE id = unit.id
            
            // Add to allocation list
            allocatedUnits.APPEND(unit)
            remaining = remaining - 1
    
    // STEP 3: Check if fully allocated
    IF remaining > 0:
        // ROLLBACK: Mark all allocated units as available again
        FOR EACH unit IN allocatedUnits:
            UPDATE collectible_units SET is_available = TRUE WHERE id = unit.id
        
        RETURN AllocationResult{
            Success: FALSE,
            Message: "Not enough stock available"
        }
    
    RETURN AllocationResult{
        Success: TRUE,
        AllocatedUnits: allocatedUnits
    }
```

### Example Allocation Scenario

**Scenario**: Customer wants 3 units of Iron Man statue, picking up at Store B (Makati)

```
Available Units:
┌─────────────────────────────────────────────────────────────────┐
│ Warehouse             │ Distance to Makati │ Available Units   │
├───────────────────────┼────────────────────┼───────────────────┤
│ Warehouse Central     │ 5 km               │ 1 unit            │
│ Warehouse South       │ 8 km               │ 1 unit            │
│ Warehouse North       │ 15 km              │ 2 units           │
└───────────────────────┴────────────────────┴───────────────────┘

Allocation Process:
1. Sort by distance: Central (5km) → South (8km) → North (15km)
2. Allocate from Central: 1 unit (remaining: 2)
3. Allocate from South: 1 unit (remaining: 1)
4. Allocate from North: 1 unit (remaining: 0)
5. ✓ Fully allocated!

Result:
┌─────────────────────────────────────────────────────────────────┐
│ Unit ID │ Allocated From          │ Distance                    │
├─────────┼─────────────────────────┼─────────────────────────────┤
│ Unit 5  │ Warehouse Central       │ 5 km                        │
│ Unit 9  │ Warehouse South         │ 8 km                        │
│ Unit 2  │ Warehouse North         │ 15 km                       │
└─────────┴─────────────────────────┴─────────────────────────────┘
```

### Key Features

1. **Greedy Nearest-First**: Always picks the closest warehouse first
2. **Atomic Allocation**: Either all units are allocated or none (with rollback)
3. **Real-time Availability**: Units are marked unavailable immediately upon allocation
4. **Automatic Return**: Failed payments or expirations return units to available pool

---

## Warehouse Distance Visualization

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        WAREHOUSE-STORE DISTANCE MATRIX                          │
│                              (distances in km)                                  │
└─────────────────────────────────────────────────────────────────────────────────┘

                          STORES (Pickup Locations)
                    ┌─────────────┬─────────────┬─────────────┐
                    │   Store A   │   Store B   │   Store C   │
                    │   Manila    │   Makati    │ Quezon City │
    ┌───────────────┼─────────────┼─────────────┼─────────────┤
    │ Warehouse     │             │             │             │
    │ North         │     5 km    │    15 km    │    10 km    │
W   │ (Caloocan)    │             │             │             │
A   ├───────────────┼─────────────┼─────────────┼─────────────┤
R   │ Warehouse     │             │             │             │
E   │ Central       │     8 km    │     5 km    │    12 km    │
H   │ (Pasig)       │             │             │             │
O   ├───────────────┼─────────────┼─────────────┼─────────────┤
U   │ Warehouse     │             │             │             │
S   │ South         │    12 km    │     8 km    │    20 km    │
E   │ (Paranaque)   │             │             │             │
S   └───────────────┴─────────────┴─────────────┴─────────────┘


┌─────────────────────────────────────────────────────────────────────────────────┐
│                         GEOGRAPHIC VISUALIZATION                                │
└─────────────────────────────────────────────────────────────────────────────────┘

                                    NORTH
                                      ↑
                     ┌────────────────────────────────────┐
                     │                                    │
                     │    ▣ Warehouse North (Caloocan)    │
                     │              ↓ 5km                 │
                     │         ★ Store A (Manila)        │
                     │              ↓ 10km                │
                     │    ▣ Warehouse Central (Pasig)     │←─ 5km →★ Store B (Makati)
                     │              ↓ 8km                 │
                     │    ▣ Warehouse South (Paranaque)   │
                     │                                    │
                     │              ★ Store C (QC)        │← 12km from Central
                     │                                    │
                     └────────────────────────────────────┘
                                      ↓
                                    SOUTH

LEGEND:
  ★ = Store (Customer Pickup Location)
  ▣ = Warehouse (Collectible Storage)


┌─────────────────────────────────────────────────────────────────────────────────┐
│                      NEAREST WAREHOUSE BY STORE                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

Store A - Manila:
  1st choice: Warehouse North (5 km)  ← NEAREST
  2nd choice: Warehouse Central (8 km)
  3rd choice: Warehouse South (12 km)

Store B - Makati:
  1st choice: Warehouse Central (5 km) ← NEAREST
  2nd choice: Warehouse South (8 km)
  3rd choice: Warehouse North (15 km)

Store C - Quezon City:
  1st choice: Warehouse North (10 km)  ← NEAREST
  2nd choice: Warehouse Central (12 km)
  3rd choice: Warehouse South (20 km)


┌─────────────────────────────────────────────────────────────────────────────────┐
│                      DISTANCE DATA (Tuple Format)                               │
└─────────────────────────────────────────────────────────────────────────────────┘

As specified in the project requirements, distances can be represented as tuples:
  [(warehouse_to_A, warehouse_to_B, warehouse_to_C), ...]

Warehouse North:  (5, 15, 10)
Warehouse Central: (8, 5, 12)
Warehouse South:  (12, 8, 20)

Combined: [(5, 15, 10), (8, 5, 12), (12, 8, 20)]
```

---

## Pricing Logic

### Rate Structure

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              PRICING STRUCTURE                                  │
└─────────────────────────────────────────────────────────────────────────────────┘

┌──────────────┬─────────────────────┬─────────────────────┐
│     SIZE     │   NORMAL RATE       │   SPECIAL RATE      │
│              │   (≥ 7 days)        │   (< 7 days)        │
├──────────────┼─────────────────────┼─────────────────────┤
│ Small (S)    │   ₱1,000/day        │   ₱2,000/day        │
│ Medium (M)   │   ₱5,000/day        │   ₱10,000/day       │
│ Large (L)    │   ₱10,000/day       │   ₱20,000/day       │
└──────────────┴─────────────────────┴─────────────────────┘

* Special Rate = 2x Normal Rate (penalty for short-term rentals)
* Minimum rental duration for normal rate: 7 days
```

### Pricing Formula

```go
func CalculateRentalPrice(size string, days int) int {
    // Base daily rate by size
    dailyRate := map[string]int{
        "S": 1000,   // Small: ₱1,000/day
        "M": 5000,   // Medium: ₱5,000/day
        "L": 10000,  // Large: ₱10,000/day
    }[size]
    
    // Apply special rate for short rentals (< 7 days)
    if days < 7 {
        dailyRate *= 2  // Double the rate
    }
    
    return dailyRate * days
}
```

### Pricing Examples

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           PRICING EXAMPLES                                      │
└─────────────────────────────────────────────────────────────────────────────────┘

Example 1: Large item, 7 days (Normal Rate)
  └─ Rate: ₱10,000/day × 7 days = ₱70,000

Example 2: Large item, 5 days (Special Rate - 2x)
  └─ Rate: ₱20,000/day × 5 days = ₱100,000

Example 3: Small item, 14 days (Normal Rate)
  └─ Rate: ₱1,000/day × 14 days = ₱14,000

Example 4: Medium item, 3 days (Special Rate - 2x)
  └─ Rate: ₱10,000/day × 3 days = ₱30,000

Example 5: Multiple units - 2x Medium item, 7 days
  └─ Unit price: ₱5,000/day × 7 days = ₱35,000
  └─ Total: ₱35,000 × 2 units = ₱70,000
```

---

## PayMongo API Integration

### Overview

The system integrates with PayMongo to process payments through multiple channels:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         PAYMONGO INTEGRATION OVERVIEW                           │
└─────────────────────────────────────────────────────────────────────────────────┘

                         SUPPORTED PAYMENT METHODS
┌─────────────────────────────────────────────────────────────────────────────────┐
│  METHOD        │  TYPE           │  LIMIT         │  PAYMONGO FLOW             │
├────────────────┼─────────────────┼────────────────┼────────────────────────────┤
│  💳 Card       │  Credit/Debit   │  ₱10,000,000   │  Payment Intent → Attach   │
│  📱 GCash      │  E-Wallet       │  ₱100,000      │  Source → Redirect         │
│  🟢 GrabPay    │  E-Wallet       │  ₱100,000      │  Source → Redirect         │
│  🏛️ BPI        │  Online Banking │  ₱50,000       │  Source → Redirect         │
│  🏦 UBP        │  Online Banking │  ₱100,000      │  Source → Redirect         │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Card Payment Flow (Payment Intent)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         CARD PAYMENT FLOW                                       │
└─────────────────────────────────────────────────────────────────────────────────┘

    YOUR SYSTEM                          PAYMONGO API
         │                                    │
         │  1. Create Payment Intent          │
         │ ──────────────────────────────────>│
         │    POST /v1/payment_intents        │
         │    {amount, currency, description} │
         │                                    │
         │    Payment Intent ID + Client Key  │
         │ <──────────────────────────────────│
         │                                    │
         │  2. Create Payment Method          │
         │ ──────────────────────────────────>│
         │    POST /v1/payment_methods        │
         │    {card_number, exp, cvc, billing}│
         │                                    │
         │    Payment Method ID               │
         │ <──────────────────────────────────│
         │                                    │
         │  3. Attach Method to Intent        │
         │ ──────────────────────────────────>│
         │    POST /v1/payment_intents/{id}/  │
         │         attach                     │
         │    {payment_method, return_url}    │
         │                                    │
         │    Status: succeeded OR            │
         │    Status: awaiting_next_action    │
         │ <──────────────────────────────────│
         │                                    │
         │  IF awaiting_next_action (3DS):    │
         │  ───────────────────────────────   │
         │  Redirect customer to 3DS URL      │
         │                                    │
         │  Customer completes 3DS            │
         │ <──────────────────────────────────│
         │                                    │
         │  4. Verify Payment                 │
         │ ──────────────────────────────────>│
         │    GET /v1/payment_intents/{id}    │
         │                                    │
         │    Status: succeeded/failed        │
         │ <──────────────────────────────────│
         │                                    │
         ▼                                    ▼
```

### E-Wallet/Bank Payment Flow (Source)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                      E-WALLET / ONLINE BANKING FLOW                             │
└─────────────────────────────────────────────────────────────────────────────────┘

    YOUR SYSTEM                          PAYMONGO API
         │                                    │
         │  1. Create Source                  │
         │ ──────────────────────────────────>│
         │    POST /v1/sources                │
         │    {amount, type: "gcash",         │
         │     redirect: {success, failed}}   │
         │                                    │
         │    Source ID + Checkout URL        │
         │ <──────────────────────────────────│
         │                                    │
         │  2. Redirect to Checkout URL       │
         │ ─────────────> CUSTOMER ──────────>│
         │                                    │
         │                Customer authorizes │
         │                payment in GCash/   │
         │                GrabPay/Bank app    │
         │                                    │
         │  3. Callback to success/failed URL │
         │ <───────────── CUSTOMER <──────────│
         │                                    │
         │  4. Verify Source Status           │
         │ ──────────────────────────────────>│
         │    GET /v1/sources/{id}            │
         │                                    │
         │    Status: chargeable/cancelled/   │
         │            expired                 │
         │ <──────────────────────────────────│
         │                                    │
         │  IF chargeable:                    │
         │  5. Create Payment from Source     │
         │ ──────────────────────────────────>│
         │    POST /v1/payments               │
         │    {source: {id, type}}            │
         │                                    │
         │    Payment successful              │
         │ <──────────────────────────────────│
         │                                    │
         ▼                                    ▼
```

### API Authentication

```go
// PayMongo uses Basic Auth with Secret Key as username
secretKey := os.Getenv("PAYMONGO_SECRET_KEY")  // sk_test_xxx or sk_live_xxx
auth := base64.StdEncoding.EncodeToString([]byte(secretKey + ":"))

req.Header.Set("Authorization", "Basic " + auth)
req.Header.Set("Content-Type", "application/json")
```

### Key PayMongo Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/payment_intents` | POST | Create payment intent for cards |
| `/v1/payment_intents/{id}/attach` | POST | Attach payment method to intent |
| `/v1/payment_intents/{id}` | GET | Verify card payment status |
| `/v1/payment_methods` | POST | Create payment method (card details) |
| `/v1/sources` | POST | Create source for e-wallets/banks |
| `/v1/sources/{id}` | GET | Verify source status |
| `/v1/payments` | POST | Create payment from chargeable source |

### Environment Variables Required

```bash
PAYMONGO_SECRET_KEY=sk_test_xxxxxxxxxxxx  # Test/Live secret key
PAYMONGO_PUBLIC_KEY=pk_test_xxxxxxxxxxxx  # Test/Live public key (for frontend)
```

### Test Cards (PayMongo Sandbox)

| Card Number | Result | 3DS |
|-------------|--------|-----|
| 4343434343434345 | Success | No |
| 4571736000000075 | Success | Yes |
| 4120000000000007 | Decline | No |
| 5100000000000065 | Success (Mastercard) | Yes |

---

## Postman Testing Guide

### Setup

1. **Base URL**: `http://localhost:8080`
2. **Authorization**: Bearer Token (from login response)

### Environment Variables

```json
{
  "base_url": "http://localhost:8080",
  "token": "",
  "rental_id": "",
  "payment_id": ""
}
```

---

### 1. Authentication Tests

#### 1.1 Register User
```
POST {{base_url}}/api/register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123",
  "name": "Test User",
  "phone": "09171234567"
}

Expected Response (201):
{
  "success": true,
  "message": "Registration successful",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User"
  }
}
```

#### 1.2 Login
```
POST {{base_url}}/api/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123"
}

Expected Response (200):
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User"
  }
}

// Save token to environment variable
```

#### 1.3 Get Current User
```
GET {{base_url}}/api/me
Authorization: Bearer {{token}}

Expected Response (200):
{
  "success": true,
  "user": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User"
  }
}
```

---

### 2. Catalog Tests

#### 2.1 Get All Collectibles
```
GET {{base_url}}/catalogue

Expected Response (200):
{
  "success": true,
  "collectibles": [
    {
      "id": 1,
      "name": "Iron Man Mark LXXXV Life-Size Statue",
      "size": "L",
      "imageURL": "https://...",
      "available_units": 4
    },
    ...
  ]
}
```

#### 2.2 Get All Stores
```
GET {{base_url}}/stores

Expected Response (200):
{
  "success": true,
  "stores": [
    {"id": 1, "name": "Store A - Manila"},
    {"id": 2, "name": "Store B - Makati"},
    {"id": 3, "name": "Store C - Quezon City"}
  ]
}
```

#### 2.3 Get Quote (Without Creating Rental)
```
POST {{base_url}}/quote
Content-Type: application/json

{
  "collectible_id": 1,
  "store_id": 1,
  "quantity": 1,
  "days": 7
}

Expected Response (200):
{
  "success": true,
  "collectible": "Iron Man Mark LXXXV Life-Size Statue",
  "size": "L",
  "quantity": 1,
  "days": 7,
  "unit_price": 70000,
  "total_price": 70000,
  "rate_type": "normal",
  "store": "Store A - Manila",
  "available_units": 4,
  "nearest_warehouse": "Warehouse North - Caloocan"
}
```

---

### 3. Rental Tests

#### 3.1 Create Rental (Normal Rate - 7 days)
```
POST {{base_url}}/api/rentals
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "collectible_id": 1,
  "store_id": 1,
  "quantity": 1,
  "days": 7
}

Expected Response (201):
{
  "success": true,
  "message": "Rental created successfully",
  "rental_id": 1,
  "total_price": 70000,
  "unit_price": 70000,
  "quantity": 1,
  "days": 7,
  "allocated_units": [
    {
      "unit_id": 1,
      "warehouse_id": 1,
      "warehouse_name": "Warehouse North - Caloocan"
    }
  ]
}
```

#### 3.2 Create Rental (Special Rate - 5 days)
```
POST {{base_url}}/api/rentals
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "collectible_id": 5,
  "store_id": 2,
  "quantity": 1,
  "days": 5
}

Expected Response (201):
{
  "success": true,
  "rental_id": 2,
  "total_price": 50000,    // M-size: ₱10,000/day × 5 = ₱50,000 (special rate)
  "unit_price": 50000
}
```

#### 3.3 Create Rental (Multiple Units)
```
POST {{base_url}}/api/rentals
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "collectible_id": 9,
  "store_id": 3,
  "quantity": 3,
  "days": 10
}

Expected Response (201):
{
  "success": true,
  "total_price": 30000,    // S-size: ₱1,000/day × 10 × 3 = ₱30,000
  "allocated_units": [
    {"warehouse_name": "Warehouse North - Caloocan"},
    {"warehouse_name": "Warehouse North - Caloocan"},
    {"warehouse_name": "Warehouse Central - Pasig"}
  ]
}
```

#### 3.4 Get User's Rentals
```
GET {{base_url}}/api/rentals
Authorization: Bearer {{token}}

Expected Response (200):
{
  "success": true,
  "rentals": [
    {
      "id": 1,
      "collectible": "Iron Man Mark LXXXV Life-Size Statue",
      "status": "pending_payment",
      "total_price": 70000,
      "time_remaining": 285,
      "expires_at": "2026-02-02T10:05:00Z"
    }
  ]
}
```

#### 3.5 Get Single Rental Details
```
GET {{base_url}}/api/rentals/1
Authorization: Bearer {{token}}

Expected Response (200):
{
  "success": true,
  "rental": {
    "id": 1,
    "status": "pending_payment",
    "total_price": 70000,
    "time_remaining": 250,
    "rental_units": [
      {
        "unit_id": 1,
        "warehouse_name": "Warehouse North - Caloocan"
      }
    ]
  }
}
```

#### 3.6 Test Insufficient Stock
```
POST {{base_url}}/api/rentals
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "collectible_id": 1,
  "store_id": 1,
  "quantity": 100,
  "days": 7
}

Expected Response (200):
{
  "success": false,
  "error": "Not enough stock available. Only 4 units available."
}
```

---

### 4. Payment Tests

#### 4.1 Create Card Payment (Non-3DS)
```
POST {{base_url}}/api/payments
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "rental_id": 1,
  "payment_method": "card",
  "billing_name": "Juan Dela Cruz",
  "billing_email": "juan@example.com",
  "billing_phone": "09171234567",
  "billing_address": "123 Main St",
  "billing_city": "Manila",
  "billing_state": "Metro Manila",
  "billing_postal": "1000",
  "card_number": "4343434343434345",
  "exp_month": 12,
  "exp_year": 2028,
  "cvc": "123",
  "success_url": "http://localhost:8080/payment/success",
  "failed_url": "http://localhost:8080/payment/failed"
}

Expected Response (200):
{
  "success": true,
  "message": "Payment successful",
  "payment_id": 1,
  "status": "paid"
}
```

#### 4.2 Create Card Payment (3DS Required)
```
POST {{base_url}}/api/payments
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "rental_id": 2,
  "payment_method": "card",
  "billing_name": "Juan Dela Cruz",
  "billing_email": "juan@example.com",
  "billing_phone": "09171234567",
  "card_number": "4571736000000075",
  "exp_month": 12,
  "exp_year": 2028,
  "cvc": "123",
  "success_url": "http://localhost:8080/payment/success",
  "failed_url": "http://localhost:8080/payment/failed"
}

Expected Response (200):
{
  "success": true,
  "message": "3DS verification required",
  "payment_id": 2,
  "checkout_url": "https://api.paymongo.com/v1/3ds/...",
  "status": "awaiting_3ds"
}
```

#### 4.3 Create GCash Payment
```
POST {{base_url}}/api/payments
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "rental_id": 3,
  "payment_method": "gcash",
  "billing_name": "Juan Dela Cruz",
  "billing_email": "juan@example.com",
  "billing_phone": "09171234567",
  "success_url": "http://localhost:8080/payment/success",
  "failed_url": "http://localhost:8080/payment/failed"
}

Expected Response (200):
{
  "success": true,
  "message": "Redirect to complete payment",
  "payment_id": 3,
  "checkout_url": "https://pm.link/gcash/test/...",
  "status": "pending"
}
```

#### 4.4 Create GrabPay Payment
```
POST {{base_url}}/api/payments
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "rental_id": 4,
  "payment_method": "grab_pay",
  "billing_name": "Juan Dela Cruz",
  "billing_email": "juan@example.com",
  "billing_phone": "09171234567",
  "success_url": "http://localhost:8080/payment/success",
  "failed_url": "http://localhost:8080/payment/failed"
}
```

#### 4.5 Create BPI Online Banking Payment
```
POST {{base_url}}/api/payments
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "rental_id": 5,
  "payment_method": "dob_bpi",
  "billing_name": "Juan Dela Cruz",
  "billing_email": "juan@example.com",
  "billing_phone": "09171234567",
  "success_url": "http://localhost:8080/payment/success",
  "failed_url": "http://localhost:8080/payment/failed"
}
```

#### 4.6 Create UBP Online Banking Payment
```
POST {{base_url}}/api/payments
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "rental_id": 6,
  "payment_method": "dob_ubp",
  "billing_name": "Juan Dela Cruz",
  "billing_email": "juan@example.com",
  "billing_phone": "09171234567",
  "success_url": "http://localhost:8080/payment/success",
  "failed_url": "http://localhost:8080/payment/failed"
}
```

#### 4.7 Verify Payment Status
```
GET {{base_url}}/api/payments/1/verify
Authorization: Bearer {{token}}

Expected Response (200):
{
  "success": true,
  "payment_id": 1,
  "status": "paid",
  "message": "Payment successful"
}
```

#### 4.8 Test Payment Limit Exceeded
```
POST {{base_url}}/api/payments
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "rental_id": 10,
  "payment_method": "gcash",
  "billing_name": "Test",
  "billing_email": "test@example.com",
  "billing_phone": "09171234567"
}

// If rental total > ₱100,000
Expected Response (400):
{
  "success": false,
  "error": "amount exceeds gcash limit of ₱100,000. Please choose a different payment method"
}
```

---

### 5. Error Handling Tests

#### 5.1 Unauthorized Access
```
GET {{base_url}}/api/rentals
// No Authorization header

Expected Response (401):
{
  "success": false,
  "error": "Not authenticated"
}
```

#### 5.2 Invalid Token
```
GET {{base_url}}/api/rentals
Authorization: Bearer invalid_token_here

Expected Response (401):
{
  "success": false,
  "error": "Invalid token"
}
```

#### 5.3 Invalid Rental ID
```
GET {{base_url}}/api/rentals/99999
Authorization: Bearer {{token}}

Expected Response (404):
{
  "success": false,
  "error": "Rental not found"
}
```

#### 5.4 Missing Required Fields
```
POST {{base_url}}/api/rentals
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "collectible_id": 1
  // Missing: store_id, quantity, days
}

Expected Response (400):
{
  "success": false,
  "error": "Invalid request: ..."
}
```

---

### 6. Integration Flow Test

Run these in sequence to test the complete flow:

```
1. POST /api/register → Create account
2. POST /api/login → Get token
3. GET /catalogue → View products
4. GET /stores → View pickup locations
5. POST /quote → Get price estimate
6. POST /api/rentals → Create rental (starts timer)
7. GET /api/rentals/1 → Verify rental created
8. POST /api/payments → Process payment
9. GET /api/payments/1/verify → Confirm payment
10. GET /api/rentals → Verify rental status is "active"
```

---

## Alternative Solutions

### 1. Allocation Algorithm Alternatives

| Approach | Description | Pros | Cons |
|----------|-------------|------|------|
| **Nearest-First (Current)** | Allocate from closest warehouse | Minimizes delivery time/cost | May deplete nearest warehouse |
| **Round-Robin** | Distribute evenly across warehouses | Balanced inventory | Longer average delivery |
| **Least-Stock-First** | Allocate from warehouse with most stock | Prevents stockouts | Ignores distance |
| **Weighted Scoring** | Score = distance × 0.7 + stock × 0.3 | Balanced optimization | More complex |

### 2. Payment Integration Alternatives

| Provider | Supported Methods | Pros | Cons |
|----------|------------------|------|------|
| **PayMongo (Current)** | Cards, GCash, GrabPay, BPI, UBP | Philippine-focused, easy API | Limited international |
| **Stripe** | Cards, wallets, bank transfers | Global, robust | Higher fees for PH |
| **PayPal** | Cards, PayPal balance | Trusted brand | Higher fees |
| **Dragonpay** | PH banks, e-wallets, OTC | Wide PH coverage | Complex integration |

### 3. Architecture Alternatives

| Pattern | Description | When to Use |
|---------|-------------|-------------|
| **Monolith (Current)** | Single Go application | Small-medium scale, simple deployment |
| **Microservices** | Separate services for auth, inventory, payments | Large scale, team distribution |
| **Serverless** | AWS Lambda / Cloud Functions | Variable traffic, cost optimization |
| **Event-Driven** | Message queues for async processing | High volume, decoupled systems |

### 4. Database Alternatives

| Option | Pros | Cons |
|--------|------|------|
| **PostgreSQL (Current)** | ACID, relations, mature | Scaling complexity |
| **MongoDB** | Flexible schema, horizontal scaling | No ACID transactions |
| **MySQL** | Widely adopted, replication | Less advanced features |
| **CockroachDB** | Distributed PostgreSQL | Newer, learning curve |

### 5. Frontend Alternatives

| Stack | Pros | Cons |
|-------|------|------|
| **HTML + Alpine.js (Current)** | Simple, fast, no build step | Limited for complex UI |
| **React + Next.js** | Component-based, SSR | Heavier, build required |
| **Vue + Nuxt** | Easy learning curve | Smaller ecosystem |
| **HTMX** | Server-side simplicity | Limited client interactivity |

---

## Summary

MongoCollectibles implements a complete rental management system with:

1. ✅ **3+ Brick-and-mortar stores** with configurable warehouse distances
2. ✅ **Automatic unit allocation** using nearest-warehouse-first algorithm
3. ✅ **Size-based pricing** (S: ₱1K, M: ₱5K, L: ₱10K per day)
4. ✅ **Special rate** for rentals < 7 days (2x normal rate)
5. ✅ **Multiple payment methods**: Cards, GCash, GrabPay, BPI, UBP via PayMongo
6. ✅ **Billing details collection** for all payments
7. ✅ **Payment timeout handling** with automatic unit return
8. ✅ **Flexible distance configuration** via database (tuple format supported)

The system is designed for scalability - additional stores, warehouses, collectibles, and payment methods can be added without code changes.
