package main

import (
	"fmt"
	"log"
	"temuin/config"
	"temuin/models"
)

func main() {
	// Initialize DB
	config.ConnectDB()
	db := config.DB

	log.Println("🔄 Adding TopUpTransaction table...")

	// Auto-migrate TopUpTransaction table
	err := db.AutoMigrate(&models.TopUpTransaction{})
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ TopUpTransaction table created successfully!")
	fmt.Println("\nYou can now use the topup feature with Midtrans integration.")
}
