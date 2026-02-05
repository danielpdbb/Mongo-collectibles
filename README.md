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

### System Overview

MongoCollectibles follows a **4-tier layered architecture** with clear separation of concerns, ensuring maintainability, testability, and scalability.

```
┌───────────────────────────────────────────────────────────────────────┐
│                        PRESENTATION LAYER                             │
│                         (Static HTML + JS)                            │
│                                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────┐  │
│  │  index.html  │  │  cart.html   │  │ payment.html │  │ rentals │  │
│  │  (Catalog)   │  │  (Shopping)  │  │  (Checkout)  │  │  .html  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └─────────┘  │
│                                                                       │
│  Technology: Alpine.js + Tailwind CSS + Fetch API                    │
│  Responsibilities:                                                    │
│   - User interface rendering                                         │
│   - Client-side state management                                     │
│   - Form validation & user interactions                              │
│   - Real-time price calculations                                     │
└───────────────────────────────────────────────────────────────────────┘
                                    ↓ HTTP/REST
┌───────────────────────────────────────────────────────────────────────┐
│                           API LAYER (Gin)                             │
│                         /internal/api/                                │
│                                                                       │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐   │
│  │  handlers.go     │  │ cart_handlers.go │  │payment_handlers  │   │
│  │  • Catalog       │  │  • Add to cart   │  │  • Create payment│   │
│  │  • Quote API     │  │  • Update items  │  │  • Verify status │   │
│  │  • Stores list   │  │  • Checkout      │  │  • Webhooks      │   │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘   │
│                                                                       │
│  ┌──────────────────┐  ┌──────────────────┐                          │
│  │ auth_handlers.go │  │rental_handlers.go│                          │
│  │  • Register      │  │  • List rentals  │                          │
│  │  • Login         │  │  • Rental detail │                          │
│  │  • Get user      │  │  • Cancel rental │                          │
│  └──────────────────┘  └──────────────────┘                          │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │            middleware.go - JWT Authentication                 │   │
│  │  • Token validation                                           │   │
│  │  • User context injection                                     │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  Responsibilities:                                                    │
│   - Request routing & validation                                     │
│   - Authentication & authorization                                   │
│   - DTO mapping (JSON ↔ Domain models)                               │
│   - Error handling & HTTP responses                                  │
└───────────────────────────────────────────────────────────────────────┘
                                    ↓
┌───────────────────────────────────────────────────────────────────────┐
│                         SERVICE LAYER                                 │
│                        /internal/service/                             │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────────┐    │
│  │  allocation.go - Smart Warehouse Allocation Algorithm        │    │
│  │  • Nearest-warehouse-first algorithm                         │    │
│  │  • Distance-based unit selection                             │    │
│  │  • Availability checking                                     │    │
│  └──────────────────────────────────────────────────────────────┘    │
│                                                                       │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐   │
│  │   pricing.go     │  │    cart.go       │  │cart_checkout.go  │   │
│  │  • Rate calc     │  │  • Cart CRUD     │  │  • Multi-item    │   │
│  │  • Discount      │  │  • Item update   │  │    checkout      │   │
│  │  • Size pricing  │  │  • Validation    │  │  • Atomic ops    │   │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘   │
│                                                                       │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐   │
│  │   rental.go      │  │  paymongo.go     │  │    auth.go       │   │
│  │  • Create rental │  │  • Create intent │  │  • JWT generate  │   │
│  │  • Expire timer  │  │  • Attach source │  │  • Password hash │   │
│  │  • Cancel rental │  │  • Verify status │  │  • User login    │   │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘   │
│                                                                       │
│  Responsibilities:                                                    │
│   - Business logic & rules                                           │
│   - Transaction orchestration                                        │
│   - External API integration (PayMongo)                              │
│   - Domain model operations                                          │
└───────────────────────────────────────────────────────────────────────┘
                                    ↓
┌───────────────────────────────────────────────────────────────────────┐
│                   DATA LAYER (Repository + Domain)                    │
│                   /internal/repository/ + /internal/domain/           │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │  database.go - GORM Connection & Configuration               │   │
│  │  • PostgreSQL connection pooling                             │   │
│  │  • Auto-migration                                            │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  Domain Models (/internal/domain/):                                  │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐   │
│  │   model.go       │  │    user.go       │  │    cart.go       │   │
│  │  • Store         │  │  • User          │  │  • Cart          │   │
│  │  • Warehouse     │  │  • BillingDetail │  │  • CartItem      │   │
│  │  • Collectible   │  └──────────────────┘  └──────────────────┘   │
│  │  • Unit                                                           │
│  │  • Distance      │  ┌──────────────────┐  ┌──────────────────┐   │
│  └──────────────────┘  │   rental.go      │  │  payment.go      │   │
│                        │  • Rental        │  │  • Payment       │   │
│                        │  • RentalUnit    │  └──────────────────┘   │
│                        └──────────────────┘                          │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │  seed.go - Database Initialization                           │   │
│  │  • 3 Stores (A, B, C)                                        │   │
│  │  • 3 Warehouses                                              │   │
│  │  • 30 Collectibles (10 per size)                            │   │
│  │  • 9 Distance mappings                                       │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  Technology: GORM v1.31 + PostgreSQL 15+                             │
│  Responsibilities:                                                    │
│   - Data persistence & retrieval                                     │
│   - Transaction management                                           │
│   - Database schema definitions                                      │
└───────────────────────────────────────────────────────────────────────┘
                                    ↓
┌───────────────────────────────────────────────────────────────────────┐
│                         EXTERNAL SERVICES                             │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────────┐    │
│  │                    PayMongo API                              │    │
│  │  • Payment Intents API                                       │    │
│  │  • Payment Methods (Cards, GCash, GrabPay, Banks)           │    │
│  │  • 3DS Authentication                                        │    │
│  │  • Webhook notifications                                     │    │
│  └──────────────────────────────────────────────────────────────┘    │
└───────────────────────────────────────────────────────────────────────┘
```

### Architectural Patterns

#### 1. **Layered Architecture**
- Clean separation between presentation, API, service, and data layers
- Each layer has well-defined responsibilities
- Dependencies flow downward (no circular dependencies)

#### 2. **Repository Pattern**
- Database access abstracted through domain models
- GORM ORM for type-safe queries
- Centralized in `/internal/repository/`

#### 3. **Service Layer Pattern**
- Business logic isolated from API handlers
- Reusable across different endpoints
- Transaction orchestration and external service integration

#### 4. **JWT Authentication**
- Stateless authentication with middleware
- Token-based authorization for protected routes
- User context propagation

### Data Flow Examples

#### 🛒 Shopping Cart Checkout Flow
```
User (Browser) → [POST /api/cart/checkout]
                        ↓
            cart_handlers.go (API)
            • Validate JWT token
            • Extract user ID
                        ↓
         cart_checkout.go (Service)
         • Begin DB transaction
         • For each cart item:
           ├─→ pricing.go: Calculate price
           ├─→ allocation.go: Find nearest warehouse
           └─→ rental.go: Create rental + allocate units
         • Clear cart
         • Commit transaction
                        ↓
         paymongo.go (Service)
         • Create PayMongo payment intent
         • Attach payment method
         • Return payment URL
                        ↓
            HTTP Response → User redirected to payment
```

#### 📦 Smart Allocation Algorithm
```
Request: Rent item X at Store B

allocation.go:
  1. Query all available units for item X
     SELECT * FROM collectible_units 
     WHERE collectible_id = X AND is_available = true
  
  2. Query distances from Store B to all warehouses
     SELECT warehouse_id, distance FROM warehouse_distances 
     WHERE store_id = B
  
  3. Find unit with minimum distance:
     For each unit:
       - Get warehouse distance to Store B
       - Track minimum distance
       - Select unit from nearest warehouse
  
  4. Return selected unit (or false if none available)
```

#### 💳 Payment Verification Flow
```
PayMongo Webhook → [POST /api/payments/:id/verify]
                        ↓
         payment_handlers.go (API)
         • Verify request authenticity
         • Extract payment status
                        ↓
            paymongo.go (Service)
            • Query PayMongo API
            • Validate payment status
                        ↓
            rental.go (Service)
            IF payment succeeded:
              • Update rental status → "active"
              • Clear expiration timer
            ELSE IF payment failed/expired:
              • Update rental status → "cancelled"
              • Release allocated units
              • Set is_available = true
                        ↓
         Update Payment record in DB
         Set status = "paid" | "failed"
```

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Go + Gin** | High performance, excellent concurrency, simple HTTP routing |
| **PostgreSQL** | ACID compliance for inventory management, relational data integrity |
| **GORM** | Type-safe ORM, auto-migrations, handles complex relationships |
| **JWT Authentication** | Stateless, scalable, works across distributed systems |
| **Alpine.js** | Lightweight reactivity without build step, perfect for server-rendered apps |
| **Nearest-warehouse algorithm** | Minimizes delivery distance, optimizes logistics costs |
| **Atomic transactions** | Ensures inventory consistency during multi-item checkouts |
| **Payment timer** | Prevents inventory locking, automatic unit release on timeout |

### Security Considerations

- 🔐 **Password Hashing**: bcrypt with salt
- 🎫 **JWT Tokens**: Signed with secret key, expiration validation
- 🛡️ **SQL Injection Protection**: GORM parameterized queries
- 🔒 **HTTPS**: TLS encryption for production (configured at reverse proxy)
- 🚫 **CORS**: Configurable origins in Gin middleware
- 💰 **PayMongo 3DS**: 3D Secure authentication for card payments
- ⏱️ **Payment Expiration**: Automatic cancellation after timeout

### Scalability Considerations

#### Vertical Scaling
- PostgreSQL connection pooling (configurable in GORM)
- Gin performance optimizations (JSON serialization, routing)
- Database indexing on foreign keys and frequent queries

#### Horizontal Scaling
- Stateless API design (JWT tokens)
- Database connection pool per instance
- Load balancer compatible (no session state)
- External state stored in PostgreSQL (carts, rentals)

#### Future Optimizations
- Redis caching for catalog data
- Message queue for webhook processing (RabbitMQ/Kafka)
- CDN for static assets (images, CSS, JS)
- Read replicas for PostgreSQL
- Microservices architecture (separate payment/inventory services)

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

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [PayMongo](https://www.paymongo.com/)
- [Alpine.js](https://alpinejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)
