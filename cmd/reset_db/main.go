package main

import (
	"fmt"
	"log"
	"math/rand"
	"temuin/config"
	"temuin/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// Initialize DB
	config.ConnectDB()
	db := config.DB

	log.Println("⚠️  RESETTING DATABASE...")

	// Drop all tables
	// Order matters for Foreign Keys
	dropTable(db, &models.ItemClaim{})
	dropTable(db, &models.CoinTransaction{})
	dropTable(db, &models.Comment{})
	dropTable(db, &models.ItemReport{})
	dropTable(db, &models.Notification{})
	dropTable(db, &models.LostItem{})
	dropTable(db, &models.SubCategory{})
	dropTable(db, &models.Category{})
	dropTable(db, &models.User{})
	dropTable(db, &models.TopUpTransaction{})
	dropTable(db, &models.TopUpTransaction{})
	dropTable(db, &models.WithdrawalRequest{})
	dropTable(db, &models.LostItemImage{}) // Drop image table

	log.Println("✅ All tables dropped.")

	// Re-migrate
	log.Println("🔄 Migrating schema...")
	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.SubCategory{},
		&models.LostItem{},
		&models.LostItemImage{}, // Migrate image table
		&models.Comment{},
		&models.CoinTransaction{},
		&models.ItemClaim{},
		&models.ItemReport{},
		&models.Notification{},
		&models.TopUpTransaction{},
		&models.WithdrawalRequest{},
	)
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	// Seed Initial Data (Categories)
	seedCategories(db)

	log.Println("✨ Database reset complete and seeded!")
}

func dropTable(db *gorm.DB, model interface{}) {
	if err := db.Migrator().DropTable(model); err != nil {
		log.Printf("Warning dropping table: %v", err)
	}
}

func seedCategories(db *gorm.DB) {
	// 1. Seed User
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	user := models.User{
		Username:    "admin",
		Password:    string(hashedPassword),
		CoinBalance: 1000,
		Email:       "admin@example.com",
		IsSuperuser: true, // Set admin as superuser for moderation
	}
	db.Create(&user)

	// 2. Categories & Subcategories
	catMap := map[string][]string{
		"Elektronik": {"Handphone", "Laptop", "Tablet", "Smartwatch", "Headset"},
		"Dokumen":    {"KTP", "SIM", "STNK", "Paspor", "Kartu ATM"},
		"Aksesoris":  {"Dompet", "Kacamata", "Jam Tangan", "Tas", "Perhiasan"},
		"Pakaian":    {"Jaket", "Topi", "Sepatu", "Baju", "Celana"},
		"Lainnya":    {"Kunci", "Mainan", "Alat Tulis", "Payung", "Botol Minum", "Lainnya"},
	}

	iconMap := map[string]string{
		"Elektronik": "smartphone",
		"Dokumen":    "description",
		"Aksesoris":  "watch",
		"Pakaian":    "checkroom",
		"Lainnya":    "category",
	}

	// Seed generator
	rand.Seed(time.Now().UnixNano())
	locations := []string{"Gedung A", "Gedung B", "Kantin", "Perpustakaan", "Masjid", "Parkiran Motor", "Parkiran Mobil", "Taman", "Lobby Utama", "Ruang Sidang"}

	for name, subs := range catMap {
		cat := models.Category{Name: name, Icon: iconMap[name]}
		db.Create(&cat)

		for _, subName := range subs {
			sub := models.SubCategory{Name: subName, CategoryID: cat.ID}
			db.Create(&sub)

			// 3. Seed Items (5 per subcategory)
			for i := 1; i <= 5; i++ {
				// Status only LOST or FOUND. FOUND means completed.
				statuses := []string{"LOST", "LOST", "LOST", "LOST", "FOUND"}
				status := statuses[rand.Intn(len(statuses))]
				location := locations[rand.Intn(len(locations))]

				// Add some detail to location to match filter query logic test
				location = fmt.Sprintf("%s, Area %d", location, rand.Intn(5)+1)

				item := models.LostItem{
					Title:         fmt.Sprintf("Kehilangan %s #%d", subName, i),
					Description:   fmt.Sprintf("Barang %s di sekitar %s. Mohon info jika menemukan.", subName, location),
					Location:      location,
					BountyCoins:   (rand.Intn(10) + 1) * 10,
					Status:        status,
					UserID:        user.ID,
					CategoryID:    cat.ID,
					SubCategoryID: &sub.ID,
					Image:         "",
				}

				// Logic for FOUND (Completed)
				if status == "FOUND" {
					item.FinderID = &user.ID
					item.FinderConfirmed = true
					item.OwnerConfirmed = true
				} else {
					// Randomly assign a "Finder" who found it but not yet confirmed by owner (Still LOST status)
					if rand.Intn(4) == 0 {
						item.FinderID = &user.ID
						item.FinderConfirmed = true
					}
				}

				db.Create(&item)
			}
		}
	}
}
