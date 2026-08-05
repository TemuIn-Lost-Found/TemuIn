package config

import (
	"fmt"
	"log"
	"os"
	"temuin/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	_ = godotenv.Load()
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "temuin_db"
	}
	dbTls := os.Getenv("DB_TLS")
	if dbTls == "" {
		dbTls = "false"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
		dbUser, dbPass, dbHost, dbPort, dbName, dbTls,
	)

	// Connect to database
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect DB:", err)
	}

	log.Println("✅ Database connected")

	// Auto Migrate
	err = DB.AutoMigrate(
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
		log.Fatal("❌ Failed to migrate database:", err)
	}

	log.Println("✅ Database migrated")

	// Run Seeder
	SeedDB(DB)
}
