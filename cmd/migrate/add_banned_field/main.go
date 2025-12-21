package main

import (
	"fmt"
	"log"
	"strings"
	"temuin/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to database
	config.ConnectDB()
	db := config.DB

	log.Println("🔄 Starting migration: Adding is_banned field to core_customuser...")

	// Add is_banned column to core_customuser table
	// Using raw SQL for compatibility with existing Django table
	err := db.Exec("ALTER TABLE core_customuser ADD COLUMN is_banned BOOLEAN DEFAULT FALSE").Error
	if err != nil {
		// Check if column already exists
		errStr := err.Error()
		if !strings.Contains(errStr, "column \"is_banned\" of relation \"core_customuser\" already exists") &&
			!strings.Contains(errStr, "Duplicate column name") {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		log.Println("⚠️  Column is_banned already exists, skipping...")
	} else {
		log.Println("✅ Successfully added is_banned column")
	}

	// Create index on is_banned for better query performance
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_core_customuser_is_banned ON core_customuser(is_banned)").Error
	if err != nil {
		log.Printf("⚠️  Warning: Could not create index: %v", err)
	} else {
		log.Println("✅ Successfully created index on is_banned")
	}

	fmt.Println("✨ Migration completed successfully!")
}
