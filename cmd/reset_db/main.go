package main

import (
	"log"
	"temuin/config"
	"temuin/models"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize DB
	config.ConnectDB()
	db := config.DB

	log.Println("⚠️  RESETTING DATABASE...")

	// Drop all tables in dependency order
	dropTable(db, &models.ItemClaim{})
	dropTable(db, &models.CoinTransaction{})
	dropTable(db, &models.Comment{})
	dropTable(db, &models.ItemReport{})
	dropTable(db, &models.Notification{})
	dropTable(db, &models.LostItemImage{})
	dropTable(db, &models.LostItem{})
	dropTable(db, &models.SubCategory{})
	dropTable(db, &models.Category{})
	dropTable(db, &models.TopUpTransaction{})
	dropTable(db, &models.WithdrawalRequest{})
	dropTable(db, &models.SiteVisit{})
	dropTable(db, &models.User{})

	log.Println("✅ All tables dropped.")

	// Re-migrate
	log.Println("🔄 Migrating schema...")
	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.SubCategory{},
		&models.LostItem{},
		&models.LostItemImage{},
		&models.Comment{},
		&models.CoinTransaction{},
		&models.ItemClaim{},
		&models.ItemReport{},
		&models.Notification{},
		&models.TopUpTransaction{},
		&models.WithdrawalRequest{},
		&models.SiteVisit{},
	)
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	// Seed Initial Data using unified seeder
	config.SeedDB(db)

	log.Println("✨ Database reset complete and seeded successfully!")
}

func dropTable(db *gorm.DB, model interface{}) {
	if err := db.Migrator().DropTable(model); err != nil {
		log.Printf("Warning dropping table: %v", err)
	}
}
