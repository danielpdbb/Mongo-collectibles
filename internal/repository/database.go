package repository

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase establishes connection to PostgreSQL
func ConnectDatabase() {
	// Database connection string
	// Format: host=HOST user=USER password=PASSWORD dbname=DBNAME port=PORT sslmode=disable
	dsn := "host=localhost user=postgres password=Halaman123! dbname=mongo_collectibles port=5432 sslmode=disable"

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("✅ Database connected successfully")
	DB = database
}

// InitDB is an alias for ConnectDatabase (for compatibility)
func InitDB() {
	ConnectDatabase()
}
