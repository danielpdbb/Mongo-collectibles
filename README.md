# 🎭 MongoCollectibles

A premium MCU collectibles rental platform built with Go, featuring PayMongo payment integration, nearest-warehouse allocation algorithm, and shopping cart functionality.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)
![License](https://img.shields.io/badge/license-MIT-green)

## ✨ Features

### 🛒 Shopping Cart
- Add multiple items to cart with different stores and rental periods
- Adjust quantities and rental days from cart
- Multi-item checkout with single payment
- Real-time price calculation

### 💳 Payment Integration (PayMongo)
- **Credit/Debit Cards** - Up to ₱10,000,000
- **GCash** - Up to ₱100,000
- **GrabPay** - Up to ₱100,000
- **BPI Online Banking** - Up to ₱50,000
- **UnionBank Online** - Up to ₱100,000
- 3DS verification for card payments
- Automatic payment expiration handling

### 📦 Smart Allocation Algorithm
- **Nearest-warehouse-first** allocation
- Minimizes delivery distance to pickup store
- Atomic transactions with rollback on failure
- Real-time stock availability

### 🏪 Multi-Store Support
- 3 pickup stores (Store A, B, C)
- 3 warehouses with distance-based routing
- Configurable warehouse distances

### ⏱️ Payment Timer
- Configurable expiration window (default: 1 minute for testing)
- Automatic unit release on expiration
- Real-time countdown display

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                      PRESENTATION LAYER                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
│  │  index.html │  │  cart.html  │  │ payment.html│  ...      │
│  │  (Catalog)  │  │  (Cart)     │  │ (Checkout)  │           │
│  └─────────────┘  └─────────────┘  └─────────────┘           │
│        ↓                ↓                ↓                    │
│  [Alpine.js + Tailwind CSS + Fetch API]                      │
└──────────────────────────────────────────────────────────────┘
                              ↓ HTTP
┌──────────────────────────────────────────────────────────────┐
│                        API LAYER (Gin)                        │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐              │
│  │ handlers.go│  │cart_handler│  │payment_han │  ...         │
│  └────────────┘  └────────────┘  └────────────┘              │
│        ↓                ↓                ↓                    │
│  [JWT Authentication Middleware]                              │
└──────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────┐
│                      SERVICE LAYER                            │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐              │
│  │ allocation │  │   cart.go  │  │ paymongo.go│  ...         │
│  │    .go     │  │            │  │            │              │
│  └────────────┘  └────────────┘  └────────────┘              │
│        ↓                ↓                ↓                    │
│  [Business Logic + External API Calls]                        │
└──────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────┐
│                    DATA LAYER (GORM + PostgreSQL)             │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ Users | Stores | Warehouses | Collectibles | Units     │  │
│  │ Carts | CartItems | Rentals | RentalUnits | Payments   │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

## 📊 Database Schema

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│      users      │     │     stores      │     │   warehouses    │
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ id              │     │ id              │     │ id              │
│ name            │     │ name            │     │ name            │
│ email           │     └─────────────────┘     └─────────────────┘
│ password_hash   │              │                      │
└─────────────────┘              └──────────┬───────────┘
        │                                   │
        │         ┌─────────────────────────┴───────────────────────┐
        │         │              warehouse_distances                 │
        │         ├─────────────────────────────────────────────────┤
        │         │ warehouse_id │ store_id │ distance (km)         │
        │         └─────────────────────────────────────────────────┘
        │
┌───────┴─────────┐     ┌─────────────────┐     ┌─────────────────┐
│      carts      │     │  collectibles   │     │collectible_units│
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ id              │     │ id              │     │ id              │
│ user_id (FK)    │     │ name            │     │ collectible_id  │
└─────────────────┘     │ size (S/M/L)    │     │ warehouse_id    │
        │               │ image_url       │     │ is_available    │
        │               └─────────────────┘     └─────────────────┘
        │                       │                       │
┌───────┴─────────┐             │                       │
│   cart_items    │             │                       │
├─────────────────┤             │                       │
│ cart_id (FK)    │             │                       │
│ collectible_id  │─────────────┘                       │
│ store_id        │                                     │
│ quantity        │                                     │
│ days            │                                     │
└─────────────────┘                                     │
                                                        │
┌─────────────────┐     ┌─────────────────┐     ┌───────┴─────────┐
│     rentals     │     │     payments    │     │  rental_units   │
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ id              │     │ rental_id       │     │ rental_id       │
│ user_id         │     │ rental_ids      │     │ unit_id (FK)    │
│ collectible_id  │     │ amount          │     │ warehouse_id    │
│ store_id        │     │ status          │     └─────────────────┘
│ quantity        │     │ payment_method  │
│ days            │     │ paymongo_id     │
│ total_price     │     └─────────────────┘
│ status          │
│ expires_at      │
└─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL 15+
- PayMongo Test Account

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/mongo-collectibles.git
cd mongo-collectibles
```

2. **Set up PostgreSQL**
```bash
# Create database
createdb mongocollectibles

# Or via psql
psql -U postgres
CREATE DATABASE mongocollectibles;
```

3. **Configure environment variables**
```bash
cp .env.example .env
# Edit .env with your credentials
```

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=mongocollectibles

# PayMongo API Keys (Test Mode)
PAYMONGO_SECRET_KEY=sk_test_xxxxxxxxxxxx
PAYMONGO_PUBLIC_KEY=pk_test_xxxxxxxxxxxx

# JWT
JWT_SECRET=your-secret-key
```

4. **Run the application**
```bash
go run ./cmd
```

5. **Seed the database** (first run only)
Uncomment line 37 in `cmd/main.go`:
```go
repository.SeedData()
```

6. **Visit the application**
```
http://localhost:8080
```

## 💰 Pricing Structure

| Size | Price/Day | Description |
|------|-----------|-------------|
| **L** (Large) | ₱10,000 | Premium statues (30cm+) |
| **M** (Medium) | ₱5,000 | Standard figures (15-30cm) |
| **S** (Small) | ₱1,000 | Mini figures (<15cm) |

**Special Rate**: 2x daily rate for rentals less than 7 days.

## 🔄 User Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER JOURNEY                             │
└─────────────────────────────────────────────────────────────────┘

  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
  │  Browse  │───▶│Add to    │───▶│  View    │───▶│ Checkout │
  │ Catalog  │    │  Cart    │    │   Cart   │    │ (Pay)    │
  └──────────┘    └──────────┘    └──────────┘    └──────────┘
       │               │               │               │
       ▼               ▼               ▼               ▼
  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
  │ Get Quote│    │ Multiple │    │  Adjust  │    │  Select  │
  │ (price)  │    │  Items   │    │ Qty/Days │    │  Method  │
  └──────────┘    └──────────┘    └──────────┘    └──────────┘
                                       │               │
                                       ▼               ▼
                                  ┌──────────┐    ┌──────────┐
                                  │  Remove  │    │ Complete │
                                  │  Items   │    │ Payment  │
                                  └──────────┘    └──────────┘
                                                       │
                                                       ▼
                                                  ┌──────────┐
                                                  │  Active  │
                                                  │  Rental  │
                                                  └──────────┘
```

## 📡 API Reference

### Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/register` | POST | Create new account |
| `/api/login` | POST | Login and get JWT token |
| `/api/me` | GET | Get current user info |

### Catalog

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/catalogue` | GET | Get all products with availability |
| `/stores` | GET | Get all store locations |
| `/quote` | POST | Calculate rental price quote |

### Cart (Requires Auth)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/cart` | GET | Get user's cart |
| `/api/cart/items` | POST | Add item to cart |
| `/api/cart/items/:id` | PUT | Update cart item |
| `/api/cart/items/:id` | DELETE | Remove from cart |
| `/api/cart` | DELETE | Clear entire cart |
| `/api/cart/checkout` | POST | Checkout all cart items |

### Rentals (Requires Auth)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/rentals` | GET | Get user's rentals |
| `/api/rentals/:id` | GET | Get specific rental |
| `/api/rentals/:id/cancel` | POST | Cancel a rental |

### Payments (Requires Auth)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/payments` | POST | Create payment |
| `/api/payments/:id/verify` | GET | Verify payment status |

## 🧪 Testing

### Test Cards (PayMongo)

| Card Number | Scenario |
|-------------|----------|
| `4343 4343 4343 4345` | Successful payment |
| `4571 7360 0000 0000` | 3DS verification required |
| `4123 4501 3100 5008` | Declined |

**Test card details:**
- Expiry: Any future date (e.g., 12/2028)
- CVC: Any 3 digits (e.g., 123)

### Postman Collection

Import the collection from `docs/postman_collection.json` for API testing.

## 🏗️ Project Structure

```
Mongo-collectibles/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── auth.go          # JWT middleware
│   │   ├── handlers.go      # General handlers
│   │   ├── cart_handlers.go # Cart API
│   │   ├── rental_handlers.go
│   │   └── payment_handlers.go
│   ├── domain/
│   │   ├── model.go         # Core entities
│   │   ├── rental.go        # Rental models
│   │   ├── payment.go       # Payment models
│   │   └── cart.go          # Cart models
│   ├── repository/
│   │   ├── database.go      # DB connection
│   │   └── seed.go          # Seed data
│   └── service/
│       ├── allocation.go    # Allocation algorithm
│       ├── pricing.go       # Pricing logic
│       ├── rental.go        # Rental service
│       ├── cart.go          # Cart service
│       ├── cart_checkout.go # Cart checkout
│       └── paymongo.go      # Payment integration
└── web/
    ├── index.html           # Catalog page
    ├── cart.html            # Cart page
    ├── checkout.html        # Checkout page
    ├── payment.html         # Payment page
    ├── rentals.html         # User rentals
    └── style.css            # Global styles
```

## ⚙️ Configuration

### Payment Timer

Located in `internal/service/rental.go` (line ~78):
```go
// Change from 1 minute (testing) to 5 minutes (production)
expiresAt := time.Now().Add(5 * time.Minute)
```

### PayMongo Mode

For production, update `.env`:
```env
PAYMONGO_SECRET_KEY=sk_live_xxxxxxxxxxxx
PAYMONGO_PUBLIC_KEY=pk_live_xxxxxxxxxxxx
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [PayMongo](https://www.paymongo.com/)
- [Alpine.js](https://alpinejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)
