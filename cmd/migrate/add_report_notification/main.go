package main

import (
	"fmt"
	"log"
	"temuin/config"
	"temuin/models"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to database
	config.ConnectDB()

	// Auto migrate the new tables
	err := config.DB.AutoMigrate(
		&models.ItemReport{},
		&models.Notification{},
	)

	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("   - Created table: core_itemreport")
	fmt.Println("   - Created table: core_notification")
}
